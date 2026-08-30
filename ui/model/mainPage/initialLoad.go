package mainpage

import (
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
	items []playlist.Item
	err   error
}

func (m *Model) initialLoad() {
	m.tracker.HideError()

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

	myWaveIdx, likesIdx, localIdx, albumsIdx, pinsIdx := -1, -1, -1, -1, -1
	for i, st := range m.playlists.Items() {
		switch st.Kind {
		case playlist.MYWAVE:
			myWaveIdx = i
		case playlist.LIKES:
			likesIdx = i
		case playlist.LOCAL:
			localIdx = i
		default:
			switch st.Name {
			case "albums":
				albumsIdx = i
			case "pins:":
				pinsIdx = i
			}
		}
	}

	var (
		wg sync.WaitGroup

		myWaveSession api.StationTracks
		myWaveErr     error

		likedAlbums    []api.Album
		likedAlbumsErr error

		likedTracksFull []api.Track
		likedTracksIds  []string
		likedErr        error

		pinnedAlbums    []api.Album
		pinnedAlbumsErr error

		localTracks []api.Track
		localErr    error

		userPlaylists      []api.Playlist
		userPlaylistsErr   error
		userPlaylistTracks [][]api.Track
	)

	if m.client != nil && myWaveIdx >= 0 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			session, err := m.client.RotorNewSession(api.MyWaveId)
			myWaveSession = session
			myWaveErr = err
		}()
	}

	if m.client != nil && likesIdx >= 0 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			likes, err := m.client.LikedTracks()
			if err != nil {
				likedErr = err
				return
			}
			ids := make([]string, len(likes))
			for i, tr := range likes {
				ids[i] = tr.Id
			}
			likedTracksIds = ids
			full, err := m.client.Tracks(ids)
			likedTracksFull = full
			likedErr = err
		}()
	}

	if m.client != nil && albumsIdx >= 0 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			la, err := m.client.LikedAlbums()
			if err != nil {
				likedAlbumsErr = err
				return
			}
			for _, likedAlbum := range la {
				album, err := m.client.Album(likedAlbum.Id, true)
				if err != nil {
					log.Print(log.LVL_ERROR, "failed to obtain album [%d] info: %s", likedAlbum.Id, err)
					continue
				}
				likedAlbums = append(likedAlbums, album)
			}
			likedAlbumsErr = err
		}()
	}

	if m.client != nil && pinsIdx >= 0 {
		pa, err := m.client.PinnedAlbums()
		if err != nil {
			pinnedAlbumsErr = err
			return
		}
		for _, pinnedAlbum := range pa {
			album, err := m.client.Album(pinnedAlbum.Data.Id, true)
			if err != nil {
				log.Print(log.LVL_ERROR, "failed to obtain album [%d] info: %s", pinnedAlbum.Data.Id, err)
				continue
			}
			pinnedAlbums = append(pinnedAlbums, album)
		}
		pinnedAlbumsErr = err
	}

	if localIdx >= 0 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			tracks, err := cache.ListTracks()
			localTracks = tracks
			localErr = err
		}()
	}

	if m.client != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			pls, err := m.client.ListPlaylists()
			if err != nil {
				userPlaylistsErr = err
				return
			}
			userPlaylists = pls
			userPlaylistTracks = make([][]api.Track, len(pls))

			var innerWg sync.WaitGroup
			for i, pl := range pls {
				innerWg.Add(1)
				go func(i int, pl api.Playlist) {
					defer innerWg.Done()
					tracks, terr := m.client.PlaylistTracks(pl.Kind, pl.Owner.Uid, false)
					if terr != nil {
						log.Print(log.LVL_ERROR, "failed to obtain playlist [%s] tracks: %s", pl.Title, terr)
						return
					}
					userPlaylistTracks[i] = tracks
				}(i, pl)
			}
			innerWg.Wait()
		}()
	}

	wg.Wait()

	if myWaveIdx >= 0 && myWaveErr == nil && m.client != nil {
		st := m.playlists.Items()[myWaveIdx]
		st.StationId = myWaveSession.Id
		st.SessionId = myWaveSession.RadioSessionId
		st.SessionBatch = myWaveSession.BatchId
		if len(myWaveSession.Sequence) > 0 {
			st.Tracks = []api.Track{myWaveSession.Sequence[0].Track}
		}
		m.playlists.SetItem(myWaveIdx, st)
	} else if myWaveErr != nil {
		log.Print(log.LVL_ERROR, "unable to init rotor session: %s", myWaveErr)
		m.tracker.ShowError("unable to init rotor session")
		return
	}

	if likesIdx >= 0 && likedErr == nil && m.client != nil {
		for _, id := range likedTracksIds {
			m.likedTracksMap[id] = true
		}
		st := m.playlists.Items()[likesIdx]
		st.Tracks = likedTracksFull
		m.playlists.SetItem(likesIdx, st)
	} else if likedErr != nil {
		log.Print(log.LVL_ERROR, "failed to obtain liked tracks: %s", likedErr)
		m.tracker.ShowError("liked tracks")
	}

	if albumsIdx >= 0 && likedAlbumsErr == nil && m.client != nil {
		st := m.playlists.Items()[albumsIdx]
		st.Kind = playlist.ALBUMS
		st.Active = len(likedAlbums) > 0
		st.Albums = likedAlbums
		st.SelectedAlbum = -1
		m.playlists.SetItem(albumsIdx, st)
	} else if likedAlbumsErr != nil {
		log.Print(log.LVL_ERROR, "failed to obtain liked albums: %s", likedAlbumsErr)
		m.tracker.ShowError("liked albums")
	}

	if pinsIdx >= 0 && pinnedAlbumsErr == nil && m.client != nil {
		for _, album := range pinnedAlbums {
			var albumTracks []api.Track
			for _, volume := range album.Volumes {
				albumTracks = append(albumTracks, volume...)
			}
			if len(albumTracks) == 0 {
				continue
			}
			m.playlists.InsertItem(pinsIdx+1, &playlist.Item{
				Name:    album.Title,
				Kind:    playlist.ALBUMS,
				Active:  true,
				Subitem: true,
				Tracks:  albumTracks,
			})
			pinsIdx++
		}
	} else if pinnedAlbumsErr != nil {
		log.Print(log.LVL_ERROR, "failed to obtain pinned albums: %s", pinnedAlbumsErr)
		m.tracker.ShowError("pinned albums")
	}

	if localIdx >= 0 && localErr == nil {
		st := m.playlists.Items()[localIdx]
		st.Tracks = localTracks
		for _, tr := range localTracks {
			m.cachedTracksMap[tr.Id] = true
		}
		m.playlists.SetItem(localIdx, st)
	} else if localErr != nil {
		log.Print(log.LVL_ERROR, "failed to list cached tracks: %s", localErr)
		m.tracker.ShowError("cache list")
	}

	if m.client != nil && userPlaylistsErr == nil {
		for i, pl := range userPlaylists {
			tracks := userPlaylistTracks[i]
			if tracks == nil {
				m.tracker.ShowError("playlist tracks")
				continue
			}
			m.playlists.InsertItem(-1, &playlist.Item{
				Name:     pl.Title,
				Kind:     pl.Kind,
				Revision: pl.Revision,
				Active:   true,
				Subitem:  true,
				Tracks:   tracks,
			})
		}
	} else if userPlaylistsErr != nil {
		log.Print(log.LVL_ERROR, "failed to obtain user playlists: %s", userPlaylistsErr)
		m.tracker.ShowError("playlists")
	}

	m.currentPlaylistIndex = -1
	m.playlists.Select(0)
	m.Send(LOADING_DONE)
}

func (m *Model) loadMyWave(wg *sync.WaitGroup, block *menuBlock) {
	defer wg.Done()

	session, err := m.client.RotorNewSession(api.MyWaveId)
	if err != nil {
		block.err = err
		return
	}

	st := &playlist.Item{Name: "my wave", Kind: playlist.MYWAVE, Active: true, Subitem: false, Rotor: true}
	st.StationId = session.Id
	st.SessionId = session.RadioSessionId
	st.SessionBatch = session.BatchId
	if len(session.Sequence) > 0 {
		st.Tracks = []api.Track{session.Sequence[0].Track}
	}

	block.items = append(block.items, st)
}

func (m *Model) loadLocal(wg *sync.WaitGroup) {
	defer wg.Done()
}

func (m *Model) loadLikes(wg *sync.WaitGroup) {
	defer wg.Done()
}

func (m *Model) loadLikedAlbums(wg *sync.WaitGroup) {
	defer wg.Done()
}

func (m *Model) loadPins(wg *sync.WaitGroup) {
	defer wg.Done()
}

func (m *Model) loadUserPlaylists(wg *sync.WaitGroup) {
	defer wg.Done()
}
