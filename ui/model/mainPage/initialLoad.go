package mainpage

import (
	"errors"
	"net/url"
	"sync"

	"github.com/dece2183/yamusic-tui/api"
	"github.com/dece2183/yamusic-tui/cache"
	"github.com/dece2183/yamusic-tui/config"
	"github.com/dece2183/yamusic-tui/log"
	"github.com/dece2183/yamusic-tui/ui/components/playlist"
)

type LoadingMsg uint

const (
	LOADING_DONE LoadingMsg = iota
)

type menuBlock struct {
	items []*playlist.Item
	err   error
}

func (m *Model) initialLoad() {
	m.tracker.HideError()
	m.playlists.Reset()

	if len(config.Current.Token) == 0 {
		log.Print(log.LVL_ERROR, "missing client token, check the config file at '%s'", config.Path())
		m.tracker.ShowError("missing token")
		m.client = nil
	} else {
		c, err := api.NewClient(config.DirName, config.Current.Token)
		m.client = c
		if err != nil {
			if _, ok := err.(*url.Error); ok {
				log.Print(log.LVL_ERROR, "failed to connect to the Yandex server: %s", err)
				m.tracker.ShowError("unable to connect to the Yandex server")
			} else {
				log.Print(log.LVL_ERROR, "client init error: %s", err)
				m.tracker.ShowError("unable to login: " + err.Error())
			}
		}
	}

	var (
		wg                     sync.WaitGroup
		myWaveMenuBlock        menuBlock
		localTracksMenuBlock   menuBlock
		likedTracksMenuBlock   menuBlock
		likedAlbumsMenuBlock   menuBlock
		pinnedAlbumsMenuBlock  menuBlock
		userPlaylistsMenuBlock menuBlock
	)

	wg.Add(6)
	go m.loadMyWave(&wg, &myWaveMenuBlock)
	go m.loadLocalTracks(&wg, &localTracksMenuBlock)
	go m.loadLikedTracks(&wg, &likedTracksMenuBlock)
	go m.loadLikedAlbums(&wg, &likedAlbumsMenuBlock)
	go m.loadPinnedAlbums(&wg, &pinnedAlbumsMenuBlock)
	go m.loadUserPlaylists(&wg, &userPlaylistsMenuBlock)
	wg.Wait()

	if myWaveMenuBlock.err == nil {
		for _, item := range myWaveMenuBlock.items {
			m.playlists.InsertItem(-1, item)
		}
	} else {
		log.Print(log.LVL_ERROR, "unable to init rotor session: %s", myWaveMenuBlock.err)
		m.tracker.ShowError("unable to init rotor session")
	}

	if localTracksMenuBlock.err == nil {
		for _, item := range localTracksMenuBlock.items {
			m.playlists.InsertItem(-1, item)
			for _, tr := range item.Tracks {
				m.cachedTracksMap[tr.Id] = true
			}
		}
	} else {
		log.Print(log.LVL_ERROR, "failed to list cached tracks: %s", localTracksMenuBlock.err)
		m.tracker.ShowError("cache list")
	}

	m.playlists.InsertItem(-1, playlist.ItemEmpty())
	m.playlists.InsertItem(-1, playlist.ItemCategory("likes:"))

	if likedTracksMenuBlock.err == nil {
		for _, item := range likedTracksMenuBlock.items {
			m.playlists.InsertItem(-1, item)
			for _, tr := range item.Tracks {
				m.likedTracksMap[tr.Id] = true
			}
		}
	} else {
		log.Print(log.LVL_ERROR, "failed to obtain liked tracks: %s", likedTracksMenuBlock.err)
		m.tracker.ShowError("liked tracks")
	}

	if likedAlbumsMenuBlock.err == nil {
		for _, item := range likedAlbumsMenuBlock.items {
			m.playlists.InsertItem(-1, item)
		}
	} else {
		log.Print(log.LVL_ERROR, "failed to obtain liked albums: %s", likedAlbumsMenuBlock.err)
		m.tracker.ShowError("liked albums")
	}

	if pinnedAlbumsMenuBlock.err == nil {
		for _, item := range pinnedAlbumsMenuBlock.items {
			m.playlists.InsertItem(-1, item)
		}
	} else {
		log.Print(log.LVL_ERROR, "failed to obtain pinned albums: %s", pinnedAlbumsMenuBlock.err)
		m.tracker.ShowError("pinned albums")
	}

	if userPlaylistsMenuBlock.err == nil {
		for _, item := range userPlaylistsMenuBlock.items {
			m.playlists.InsertItem(-1, item)
		}
	} else {
		log.Print(log.LVL_ERROR, "failed to obtain user playlists: %s", userPlaylistsMenuBlock.err)
		m.tracker.ShowError("playlists")
	}

	m.currentPlaylistIndex = -1
	m.playlists.Select(0)
	m.Send(LOADING_DONE)
}

func (m *Model) loadMyWave(wg *sync.WaitGroup, block *menuBlock) {
	defer wg.Done()

	if m.client == nil {
		return
	}

	session, err := m.client.RotorNewSession(api.MyWaveId)
	if err != nil {
		block.err = err
		return
	}

	st := &playlist.Item{Name: "my wave", Kind: playlist.MYWAVE, Active: true, Subitem: false, Rotor: true}
	st.SessionId = session.RadioSessionId
	st.SessionBatch = session.BatchId
	if len(session.AcceptedSeeds) > 0 {
		st.StationId = session.AcceptedSeeds[0]
	} else {
		st.StationId = session.Id
	}
	if len(session.Sequence) > 0 {
		st.Tracks = []api.Track{session.Sequence[0].Track}
	} else {
		block.err = errors.New("unable to get session tracks")
	}

	block.items = append(block.items, st)
}

func (m *Model) loadLocalTracks(wg *sync.WaitGroup, block *menuBlock) {
	defer wg.Done()

	localTracks, err := cache.ListTracks()
	if err != nil {
		block.err = err
		return
	}

	st := &playlist.Item{Name: "local", Kind: playlist.LOCAL, Active: true, Subitem: false, Tracks: localTracks}
	block.items = append(block.items, st)
}

func (m *Model) loadLikedTracks(wg *sync.WaitGroup, block *menuBlock) {
	defer wg.Done()

	if m.client == nil {
		return
	}

	likes, err := m.client.LikedTracks()
	if err != nil {
		block.err = err
		return
	}

	ids := make([]string, len(likes))
	for i, tr := range likes {
		ids[i] = tr.Id
	}

	tracks, err := m.client.Tracks(ids)
	if err != nil {
		block.err = err
		return
	}

	st := &playlist.Item{Name: "tracks", Kind: playlist.LIKES, Active: len(tracks) > 0, Subitem: true, Tracks: tracks}
	block.items = append(block.items, st)
}

func (m *Model) loadLikedAlbums(wg *sync.WaitGroup, block *menuBlock) {
	defer wg.Done()

	if m.client == nil {
		return
	}

	likedAlbums, err := m.client.LikedAlbums()
	if err != nil {
		block.err = err
		return
	}

	albums := make([]api.Album, 0, len(likedAlbums))
	for _, albumInfo := range likedAlbums {
		album, err := m.client.Album(albumInfo.Id, true)
		if err != nil {
			log.Print(log.LVL_ERROR, "failed to obtain album [%d] info: %s", albumInfo.Id, err)
			continue
		}
		albums = append(albums, album)
	}

	st := &playlist.Item{Name: "albums", Kind: playlist.ALBUMS, Active: len(albums) > 0, Subitem: true, Albums: albums, SelectedAlbum: -1}
	block.items = append(block.items, st)
}

func (m *Model) loadPinnedAlbums(wg *sync.WaitGroup, block *menuBlock) {
	defer wg.Done()

	if m.client == nil {
		return
	}

	pinnedAlbums, err := m.client.PinnedAlbums()
	if err != nil {
		block.err = err
		return
	}

	albums := make([]api.Album, 0, len(pinnedAlbums))
	for _, albumInfo := range pinnedAlbums {
		album, err := m.client.Album(albumInfo.Data.Id, true)
		if err != nil {
			log.Print(log.LVL_ERROR, "failed to obtain pinned album [%d] info: %s", albumInfo.Data.Id, err)
			continue
		}
		albums = append(albums, album)
	}

	if len(albums) > 0 {
		block.items = append(block.items, playlist.ItemEmpty())
		block.items = append(block.items, playlist.ItemCategory("pins:"))

		var albumTracks []api.Track
		for _, album := range albums {
			for _, volume := range album.Volumes {
				albumTracks = append(albumTracks, volume...)
			}
			if len(albumTracks) == 0 {
				continue
			}
			block.items = append(block.items, &playlist.Item{
				Name:    album.Title,
				Kind:    playlist.ALBUMS,
				Active:  true,
				Subitem: true,
				Tracks:  albumTracks,
			})
		}
	}
}

func (m *Model) loadUserPlaylists(wg *sync.WaitGroup, block *menuBlock) {
	defer wg.Done()

	if m.client == nil {
		return
	}

	playlists, err := m.client.ListPlaylists()
	if err != nil {
		block.err = err
		return
	}

	playlistTracks := make([][]api.Track, len(playlists))
	var innerWg sync.WaitGroup
	for i, pl := range playlists {
		innerWg.Add(1)
		go func(i int, pl api.Playlist) {
			defer innerWg.Done()
			tracks, terr := m.client.PlaylistTracks(pl.Kind, pl.Owner.Uid, false)
			if terr != nil {
				log.Print(log.LVL_ERROR, "failed to obtain user playlist [%s] tracks: %s", pl.Title, terr)
				return
			}
			playlistTracks[i] = tracks
		}(i, pl)
	}
	innerWg.Wait()

	if len(playlists) > 0 {
		block.items = append(block.items, playlist.ItemEmpty())
		block.items = append(block.items, playlist.ItemCategory("playlists:"))

		for i, pl := range playlists {
			tracks := playlistTracks[i]
			if len(tracks) == 0 {
				m.tracker.ShowError("playlist tracks")
				continue
			}
			block.items = append(block.items, &playlist.Item{
				Name:     pl.Title,
				Kind:     pl.Kind,
				Revision: pl.Revision,
				Active:   true,
				Subitem:  true,
				Tracks:   tracks,
			})
		}
	}
}
