package mainpage

import (
	"net/url"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/dece2183/yamusic-tui/api"
	"github.com/dece2183/yamusic-tui/cache"
	"github.com/dece2183/yamusic-tui/config"
	"github.com/dece2183/yamusic-tui/log"
	"github.com/dece2183/yamusic-tui/media/handler"
	"github.com/dece2183/yamusic-tui/ui/components/input"
	"github.com/dece2183/yamusic-tui/ui/components/playlist"
	"github.com/dece2183/yamusic-tui/ui/components/search"
	"github.com/dece2183/yamusic-tui/ui/components/tracker"
	"github.com/dece2183/yamusic-tui/ui/components/tracklist"
	"github.com/dece2183/yamusic-tui/ui/model"
	"github.com/dece2183/yamusic-tui/ui/style"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dece2183/go-clipboard"
)

type LoadingMsg uint

const (
	LOADING_DONE LoadingMsg = iota
)

type LikedTracksRefreshedMsg struct {
	Ids    []string
	Tracks []api.Track
}

type MyWaveRefreshedMsg struct {
	Session api.StationTracks
}

type PlaylistsRefreshedMsg struct {
	Entries []cache.PlaylistEntry
}

type Model struct {
	program       *tea.Program
	client        *api.YaMusicClient
	clipboard     *clipboard.Clipboard
	mediaHandler  handler.MediaHandler
	width, height int

	spinner   spinner.Model
	playlists *playlist.Model
	tracklist *tracklist.Model
	tracker   *tracker.Model

	searchDialog           *search.Model
	inputDialog            *input.Model
	isLoading              bool
	isSearchActive         bool
	isAddPlaylistActive    bool
	isRenamePlaylistActive bool
	isPlaylistHideOverride bool

	currentPlaylistIndex int
	likedTracksMap       map[string]bool
	cachedTracksMap      map[string]bool
}

// mainpage.Model constructor.
func New(mediaHandler handler.MediaHandler) *Model {
	m := &Model{}

	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion())
	m.program = p
	m.clipboard = clipboard.New()
	m.mediaHandler = mediaHandler
	m.likedTracksMap = make(map[string]bool)
	m.cachedTracksMap = make(map[string]bool)
	m.spinner = spinner.New(spinner.WithSpinner(spinner.Points))
	m.playlists = playlist.New(m.program, "YaMusic")
	m.tracklist = tracklist.New(m.program, &m.likedTracksMap, &m.cachedTracksMap)
	m.tracker = tracker.New(m.program, &m.likedTracksMap)
	m.searchDialog = search.New()
	m.inputDialog = input.New()

	return m
}

//
// model.Model interface implementation
//

func (m *Model) Run() error {
	go m.mediaHandle()
	_, err := m.program.Run()
	m.tracker.Stop()
	return err
}

func (m *Model) Send(msg tea.Msg) {
	go m.program.Send(msg)
}

//
// tea.Model interface implementation
//

func (m *Model) Init() tea.Cmd {
	m.isLoading = true
	go m.initialLoad()
	return m.spinner.Tick
}

func (m *Model) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := message.(type) {
	case LoadingMsg:
		m.isLoading = false
		return m, model.Cmd(playlist.CURSOR_UP)

	case LikedTracksRefreshedMsg:
		for k := range m.likedTracksMap {
			delete(m.likedTracksMap, k)
		}
		for _, id := range msg.Ids {
			m.likedTracksMap[id] = true
		}
		for i, st := range m.playlists.Items() {
			if st.Kind == playlist.LIKES {
				st.Tracks = msg.Tracks
				m.playlists.SetItem(i, st)
				if m.playlists.SelectedItem().Kind == playlist.LIKES {
					m.displayPlaylist(st)
				}
				break
			}
		}
		return m, nil

	case MyWaveRefreshedMsg:
		for i, st := range m.playlists.Items() {
			if st.Kind == playlist.MYWAVE {
				st.StationId = msg.Session.Id
				st.SessionId = msg.Session.RadioSessionId
				st.SessionBatch = msg.Session.BatchId
				existing := make(map[string]bool, len(st.Tracks))
				for _, t := range st.Tracks {
					existing[t.Id] = true
				}
				for _, item := range msg.Session.Sequence {
					if !existing[item.Track.Id] {
						st.Tracks = append(st.Tracks, item.Track)
						existing[item.Track.Id] = true
					}
				}
				m.playlists.SetItem(i, st)
				if m.playlists.SelectedItem().Kind == playlist.MYWAVE {
					m.displayPlaylist(st)
				}
				break
			}
		}
		return m, nil

	case PlaylistsRefreshedMsg:
		entryByKind := make(map[uint64]cache.PlaylistEntry, len(msg.Entries))
		for _, e := range msg.Entries {
			entryByKind[e.Playlist.Kind] = e
		}
		for i, st := range m.playlists.Items() {
			if !st.Subitem {
				continue
			}
			if e, ok := entryByKind[st.Kind]; ok {
				st.Tracks = e.Tracks
				st.Revision = e.Playlist.Revision
				m.playlists.SetItem(i, st)
				if m.playlists.SelectedItem().Kind == st.Kind {
					m.displayPlaylist(st)
				}
			}
		}
		return m, nil

	case tea.WindowSizeMsg:
		m.resize(msg.Width, msg.Height)
		return m, tea.ClearScreen

	case tea.KeyMsg:
		controls := config.Current.Controls
		keypress := msg.String()

		switch {
		case controls.Quit.Contains(keypress):
			return m, tea.Quit
		case m.isSearchActive || m.isAddPlaylistActive:
			m.searchDialog, cmd = m.searchDialog.Update(message)
			cmds = append(cmds, cmd)
		case m.isRenamePlaylistActive:
			m.inputDialog, cmd = m.inputDialog.Update(message)
			cmds = append(cmds, cmd)
		case controls.Reload.Contains(keypress):
			m.isLoading = true
			cmd = m.playlists.Reset()
			cmds = append(cmds, cmd)
			cmds = append(cmds, m.spinner.Tick)
			go m.initialLoad()
		default:
			if m.isLoading {
				m.spinner, cmd = m.spinner.Update(message)
				cmds = append(cmds, cmd)
			} else {
				m.playlists, cmd = m.playlists.Update(message)
				cmds = append(cmds, cmd)
				m.tracklist, cmd = m.tracklist.Update(message)
				cmds = append(cmds, cmd)
				m.tracker, cmd = m.tracker.Update(message)
				cmds = append(cmds, cmd)
			}
		}

	// playlist control update
	case playlist.Control:
		switch msg {
		case playlist.CURSOR_UP, playlist.CURSOR_DOWN:
			selectedPlaylist := m.playlists.SelectedItem()

			if m.currentPlaylistIndex >= 0 {
				currentPlaylist := m.playlists.Items()[m.currentPlaylistIndex]
				if selectedPlaylist.IsSame(currentPlaylist) && len(selectedPlaylist.Tracks) > 0 {
					selectedPlaylist.SelectedTrack = selectedPlaylist.CurrentTrack
					m.playlists.SetItem(m.playlists.Index(), selectedPlaylist)
				}
			}

			m.displayPlaylist(selectedPlaylist)
			m.indicateCurrentTrackPlaying(m.tracker.IsPlaying())

			m.tracklist.Shufflable = (selectedPlaylist.Kind != playlist.NONE && selectedPlaylist.Kind != playlist.MYWAVE && len(selectedPlaylist.Tracks) > 0)
		case playlist.RENAME:
			selectedPlaylist := m.playlists.SelectedItem()
			if selectedPlaylist.Kind < playlist.USER {
				break
			}
			m.inputDialog.Title = "Rename playlist " + selectedPlaylist.Name
			m.inputDialog.SetValue(selectedPlaylist.Name)
			m.isRenamePlaylistActive = true
		case playlist.TOGGLE_VIEW:
			m.isPlaylistHideOverride = !m.isPlaylistHideOverride
		}

	// tracklist control update
	case tracklist.Control:
		switch msg {
		case tracklist.PLAY:
			playlistItem := m.playlists.SelectedItem()
			if !playlistItem.Active {
				break
			}
			m.playSelectedPlaylist(m.tracklist.Index())
		case tracklist.CURSOR_UP, tracklist.CURSOR_DOWN:
			currentPlaylist := m.playlists.SelectedItem()
			cursorIndex := m.tracklist.Index()
			currentPlaylist.SelectedTrack = cursorIndex
			cmd = m.playlists.SetItem(m.playlists.Index(), currentPlaylist)
			cmds = append(cmds, cmd)
		case tracklist.LIKE:
			cmd = m.likeSelectedTrack()
			cmds = append(cmds, cmd)
		case tracklist.ADD_TO_PLAYLIST:
			selectedTrack := m.tracklist.SelectedItem()
			m.searchDialog.Title = "Add " + selectedTrack.Track.Title + " to"
			m.searchDialog.Action = "add"
			m.isAddPlaylistActive = true
			m.Send(search.UPDATE_SUGGESTIONS)
		case tracklist.REMOVE_FROM_PLAYLIST:
			selectedPlaylist := m.playlists.SelectedItem()
			cmd = m.removeFromPlaylist(selectedPlaylist, m.tracklist.Index())
			cmds = append(cmds, cmd)
		case tracklist.SEARCH:
			m.searchDialog.Title = "Search"
			m.searchDialog.Action = "search"
			m.isSearchActive = true
			m.Send(search.UPDATE_SUGGESTIONS)
		case tracklist.SHUFFLE:
			cmd = m.shufflePlaylist(m.playlists.SelectedItem())
			cmds = append(cmds, cmd)
		case tracklist.SHARE:
			link := api.ShareTrackLink(m.tracklist.SelectedItem().Track)
			if link != "" {
				m.clipboard.CopyText(link)
			}
		}

	// player control update
	case tracker.Control:
		switch msg {
		case tracker.NEXT:
			m.nextTrack()
		case tracker.PREV:
			m.prevTrack()
		case tracker.LIKE:
			cmd = m.likePlayingTrack()
			cmds = append(cmds, cmd)
		case tracker.PLAY, tracker.PAUSE:
			m.mediaHandler.OnPlayPause()
		case tracker.STOP:
			m.mediaHandler.OnEnded()
		case tracker.REWIND:
			m.mediaHandler.OnSeek(m.tracker.Position())
		case tracker.VOLUME:
			m.mediaHandler.OnVolume()
		case tracker.CACHE_TRACK:
			cmd = m.cacheCurrentTrack()
			cmds = append(cmds, cmd)
		case tracker.BUFFERING_COMPLETE:
			cacheMode := config.Current.CacheTracks
			if cacheMode == config.CACHE_ALL || (cacheMode == config.CACHE_LIKED_ONLY && m.likedTracksMap[m.tracker.CurrentTrack().Id]) {
				cmd = m.cacheCurrentTrack()
				cmds = append(cmds, cmd)
			}
		}

		m.tracker, cmd = m.tracker.Update(message)
		cmds = append(cmds, cmd)

	// search control update
	case search.Control:
		if m.isSearchActive {
			cmd = m.searchControl(msg)
			cmds = append(cmds, cmd)
		} else if m.isAddPlaylistActive {
			cmd = m.addPlaylistControl(msg)
			cmds = append(cmds, cmd)
		}

	// input dialog control update
	case input.Control:
		m.isRenamePlaylistActive = false
		cmd = m.renamePlaylistControl(msg)
		cmds = append(cmds, cmd)

	default:
		if m.isLoading {
			m.spinner, cmd = m.spinner.Update(message)
			cmds = append(cmds, cmd)
		} else if m.isSearchActive || m.isAddPlaylistActive {
			m.searchDialog, cmd = m.searchDialog.Update(message)
			cmds = append(cmds, cmd)
		} else if m.isRenamePlaylistActive {
			m.inputDialog, cmd = m.inputDialog.Update(message)
			cmds = append(cmds, cmd)
		} else {
			m.playlists, cmd = m.playlists.Update(message)
			cmds = append(cmds, cmd)
			m.tracklist, cmd = m.tracklist.Update(message)
			cmds = append(cmds, cmd)
			m.tracker, cmd = m.tracker.Update(message)
			cmds = append(cmds, cmd)
		}
	}

	return m, tea.Batch(cmds...)
}

func (m *Model) View() string {
	if m.isLoading {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, m.spinner.View())
	}

	if m.isSearchActive || m.isAddPlaylistActive {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, m.searchDialog.View())
	} else if m.isRenamePlaylistActive {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, m.inputDialog.View())
	}

	playlistView := m.playlists.View()
	playlistWidth := lipgloss.Width(playlistView)

	m.tracker.SetWidth(m.width - playlistWidth - 2)
	m.tracklist.SetWidth(m.width - playlistWidth - 2)

	trackerView := m.tracker.View()
	trackerHeight := lipgloss.Height(trackerView)
	m.tracklist.SetHeight(m.height - trackerHeight - 2)

	tracklistView := m.tracklist.View()

	var midPanel string
	if m.tracklist.Hidden {
		midPanel = trackerView
	} else if m.tracker.Hidden {
		midPanel = tracklistView
	} else {
		midPanel = lipgloss.JoinVertical(lipgloss.Left, tracklistView, trackerView)
	}

	return lipgloss.JoinHorizontal(lipgloss.Bottom, playlistView, midPanel)
}

//
// private methods
//

func (m *Model) resize(width, height int) {
	m.width, m.height = width, height

	m.playlists.SetSize(style.SidePanelWidth, height-4)
	if !m.isPlaylistHideOverride {
		m.playlists.Hidden = m.width < style.SidePanelAutohide
	}

	searchWidth := style.SearchModalWidth
	if searchWidth > m.width {
		searchWidth = m.width - 2
	}

	m.searchDialog.SetSize(searchWidth, m.height-4)
	m.inputDialog.SetWidth(searchWidth)
}

func (m *Model) initialLoad() {
	m.tracker.HideError()

	if len(config.Current.Token) == 0 {
		log.Print(log.LVL_ERROR, "missing client token, check the config file at '%s'", config.Path())
		m.tracker.ShowError("missing token")
		m.client = nil
	} else {
		var usedCache bool
		if config.MetaCacheEnabled() {
			if acc, accErr := cache.ReadAccount(); accErr == nil && acc != nil && acc.Uid != 0 {
				m.client = api.NewClientWithUid(config.DirName, config.Current.Token, acc.Uid)
				usedCache = true
			}
		}
		if !usedCache {
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
			} else if c != nil && config.MetaCacheEnabled() {
				if werr := cache.WriteAccount(&cache.AccountData{Uid: c.UserId()}); werr != nil {
					log.Print(log.LVL_WARNIGN, "failed to write account cache: %s", werr)
				}
			}
		}
	}

	myWaveIdx, likesIdx, localIdx := -1, -1, -1
	for i, st := range m.playlists.Items() {
		switch st.Kind {
		case playlist.MYWAVE:
			myWaveIdx = i
		case playlist.LIKES:
			likesIdx = i
		case playlist.LOCAL:
			localIdx = i
		}
	}

	var (
		wg sync.WaitGroup

		myWaveSession     api.StationTracks
		myWaveErr         error
		myWaveCachedTrack *api.Track

		likedTracksFull []api.Track
		likedTracksIds  []string
		likedErr        error

		localTracks []api.Track
		localErr    error

		userPlaylists      []api.Playlist
		userPlaylistsErr   error
		userPlaylistTracks [][]api.Track
	)

	if m.client != nil && myWaveIdx >= 0 {
		var cached *cache.MyWaveData
		if config.MetaCacheEnabled() {
			cached, _ = cache.ReadMyWave()
		}
		if cached != nil {
			myWaveSession = api.StationTracks{
				Id:             cached.StationId,
				RadioSessionId: cached.SessionId,
				BatchId:        cached.SessionBatch,
			}
			cachedTrack := cached.Track
			myWaveCachedTrack = &cachedTrack
			go func() {
				session, err := m.client.RotorNewSession(api.MyWaveId)
				if err != nil {
					log.Print(log.LVL_ERROR, "refresh rotor session: %s", err)
					return
				}
				m.Send(MyWaveRefreshedMsg{Session: session})
			}()
		} else {
			wg.Add(1)
			go func() {
				defer wg.Done()
				session, err := m.client.RotorNewSession(api.MyWaveId)
				myWaveSession = session
				myWaveErr = err
			}()
		}
	}

	if m.client != nil && likesIdx >= 0 {
		var cached *cache.LikedTracksData
		if config.MetaCacheEnabled() {
			cached, _ = cache.ReadLikedTracks()
		}
		if cached != nil && len(cached.Tracks) > 0 {
			likedTracksIds = cached.Ids
			likedTracksFull = cached.Tracks
			go m.refreshLikedTracks(cached.Ids)
		} else {
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
				if err == nil && config.MetaCacheEnabled() {
					if werr := cache.WriteLikedTracks(&cache.LikedTracksData{Ids: ids, Tracks: full}); werr != nil {
						log.Print(log.LVL_WARNIGN, "failed to write liked tracks cache: %s", werr)
					}
				}
			}()
		}
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
		var cachedPls *cache.PlaylistsData
		if config.MetaCacheEnabled() {
			cachedPls, _ = cache.ReadPlaylists()
		}
		hit := cachedPls != nil && len(cachedPls.Entries) > 0
		if hit {
			userPlaylists = make([]api.Playlist, len(cachedPls.Entries))
			userPlaylistTracks = make([][]api.Track, len(cachedPls.Entries))
			for i, e := range cachedPls.Entries {
				userPlaylists[i] = e.Playlist
				userPlaylistTracks[i] = e.Tracks
			}
			go m.refreshPlaylists(cachedPls)
		} else {
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

				if config.MetaCacheEnabled() {
					entries := make([]cache.PlaylistEntry, 0, len(pls))
					for i, pl := range pls {
						if userPlaylistTracks[i] == nil {
							continue
						}
						entries = append(entries, cache.PlaylistEntry{Playlist: pl, Tracks: userPlaylistTracks[i]})
					}
					if werr := cache.WritePlaylists(&cache.PlaylistsData{Entries: entries}); werr != nil {
						log.Print(log.LVL_WARNIGN, "failed to write playlists cache: %s", werr)
					}
				}
			}()
		}
	}

	wg.Wait()

	if myWaveIdx >= 0 && myWaveErr == nil && m.client != nil {
		st := m.playlists.Items()[myWaveIdx]
		st.StationId = myWaveSession.Id
		st.SessionId = myWaveSession.RadioSessionId
		st.SessionBatch = myWaveSession.BatchId
		if myWaveCachedTrack != nil {
			st.Tracks = []api.Track{*myWaveCachedTrack}
		} else if len(myWaveSession.Sequence) > 0 {
			st.Tracks = make([]api.Track, 0, len(myWaveSession.Sequence))
			for _, item := range myWaveSession.Sequence {
				st.Tracks = append(st.Tracks, item.Track)
			}
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

func (m *Model) refreshLikedTracks(cachedIds []string) {
	if m.client == nil {
		return
	}
	likes, err := m.client.LikedTracks()
	if err != nil {
		log.Print(log.LVL_ERROR, "refresh liked tracks ids: %s", err)
		return
	}

	newIds := make([]string, len(likes))
	for i, tr := range likes {
		newIds[i] = tr.Id
	}

	if sameStringSet(cachedIds, newIds) {
		return
	}

	full, err := m.client.Tracks(newIds)
	if err != nil {
		log.Print(log.LVL_ERROR, "refresh liked tracks full info: %s", err)
		return
	}

	if config.MetaCacheEnabled() {
		if werr := cache.WriteLikedTracks(&cache.LikedTracksData{Ids: newIds, Tracks: full}); werr != nil {
			log.Print(log.LVL_WARNIGN, "failed to write liked tracks cache: %s", werr)
		}
	}
	m.Send(LikedTracksRefreshedMsg{Ids: newIds, Tracks: full})
}

func (m *Model) refreshPlaylists(cached *cache.PlaylistsData) {
	if m.client == nil {
		return
	}
	pls, err := m.client.ListPlaylists()
	if err != nil {
		log.Print(log.LVL_ERROR, "refresh list playlists: %s", err)
		return
	}

	cachedMap := make(map[uint64]int, len(cached.Entries))
	for _, e := range cached.Entries {
		cachedMap[e.Playlist.Kind] = e.Playlist.Revision
	}

	changed := len(pls) != len(cached.Entries)
	if !changed {
		for _, pl := range pls {
			if rev, ok := cachedMap[pl.Kind]; !ok || rev != pl.Revision {
				changed = true
				break
			}
		}
	}
	if !changed {
		return
	}

	results := make([][]api.Track, len(pls))
	var innerWg sync.WaitGroup
	for i, pl := range pls {
		innerWg.Add(1)
		go func(i int, pl api.Playlist) {
			defer innerWg.Done()
			tracks, terr := m.client.PlaylistTracks(pl.Kind, pl.Owner.Uid, false)
			if terr != nil {
				log.Print(log.LVL_ERROR, "refresh playlist [%s] tracks: %s", pl.Title, terr)
				return
			}
			results[i] = tracks
		}(i, pl)
	}
	innerWg.Wait()

	entries := make([]cache.PlaylistEntry, 0, len(pls))
	for i, pl := range pls {
		if results[i] == nil {
			continue
		}
		entries = append(entries, cache.PlaylistEntry{Playlist: pl, Tracks: results[i]})
	}
	if config.MetaCacheEnabled() {
		if werr := cache.WritePlaylists(&cache.PlaylistsData{Entries: entries}); werr != nil {
			log.Print(log.LVL_WARNIGN, "failed to write playlists cache: %s", werr)
		}
	}
	m.Send(PlaylistsRefreshedMsg{Entries: entries})
}

func sameStringSet(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	set := make(map[string]struct{}, len(a))
	for _, s := range a {
		set[s] = struct{}{}
	}
	for _, s := range b {
		if _, ok := set[s]; !ok {
			return false
		}
	}
	return true
}

func (m *Model) mediaHandle() {
	for msg := range m.mediaHandler.Message() {
		switch msg.Type {
		case handler.MSG_NEXT:
			m.Send(tracker.NEXT)
		case handler.MSG_PREVIOUS:
			m.Send(tracker.PREV)
		case handler.MSG_PLAY:
			m.tracker.Play()
			m.Send(tracker.PLAY)
		case handler.MSG_PAUSE:
			m.tracker.Pause()
			m.Send(tracker.PAUSE)
		case handler.MSG_PLAYPAUSE:
			if m.tracker.IsPlaying() {
				m.tracker.Pause()
				m.Send(tracker.PAUSE)
			} else {
				m.tracker.Play()
				m.Send(tracker.PLAY)
			}
		case handler.MSG_STOP:
			m.Send(tracker.STOP)
		case handler.MSG_SEEK:
			offset, ok := msg.Arg.(time.Duration)
			if ok {
				m.tracker.Rewind(offset)
			}
		case handler.MSG_SETPOS:
			pos, ok := msg.Arg.(time.Duration)
			if ok {
				m.tracker.SetPos(pos)
			}

		case handler.MSG_SET_SHUFFLE:
			val, ok := msg.Arg.(bool)
			if !ok || !val {
				break
			}
			currentPlaylist := m.playlists.Items()[m.currentPlaylistIndex]
			if len(currentPlaylist.Tracks) == 0 {
				break
			}
			if currentPlaylist.Kind >= playlist.LIKES {
				cmd := m.shufflePlaylist(currentPlaylist)
				m.Send(func() tea.Cmd {
					return cmd
				})
			}
		case handler.MSG_SET_VOLUME:
			vol, ok := msg.Arg.(float64)
			if ok {
				m.tracker.SetVolume(vol)
			}

		case handler.MSG_GET_PLAYBACKSTATUS:
			var state handler.PlaybackState
			if m.tracker.IsPlaying() {
				state = handler.STATE_PLAYING
			} else {
				if m.tracker.IsStoped() {
					state = handler.STATE_STOPED
				} else {
					state = handler.STATE_PAUSED
				}
			}
			m.mediaHandler.SendAnswer(state)
		case handler.MSG_GET_SHUFFLE:
			m.mediaHandler.SendAnswer(false)
		case handler.MSG_GET_METADATA:
			if m.tracker.IsStoped() {
				m.mediaHandler.SendAnswer(handler.TrackMetadata{})
				break
			}
			track := m.tracker.CurrentTrack()
			artists := make([]string, 0, len(track.Artists))
			for i := range track.Artists {
				artists = append(artists, track.Artists[i].Name)
			}
			albumArtists := make([]string, 0)
			var albumName string
			genre := make([]string, 0)
			if len(track.Albums) != 0 {
				for i := range track.Albums[0].Artists {
					albumArtists = append(albumArtists, track.Albums[0].Artists[i].Name)
				}
				albumName = track.Albums[0].Title
				genre = append(genre, track.Albums[0].Genre)
			}

			md := handler.TrackMetadata{
				TrackId:      track.Id,
				Length:       time.Duration(track.DurationMs) * time.Millisecond,
				CoverUrl:     m.coverFilePath(track),
				AlbumName:    albumName,
				AlbumArtists: albumArtists,
				Artists:      artists,
				Genre:        genre,
				Title:        track.Title,
				Url:          api.ShareTrackLink(track),
			}
			m.mediaHandler.SendAnswer(md)
		case handler.MSG_GET_VOLUME:
			m.mediaHandler.SendAnswer(m.tracker.Volume())
		case handler.MSG_GET_POSITION:
			m.mediaHandler.SendAnswer(m.tracker.Position())
		}
	}
}

func (m *Model) coverFilePath(track *api.Track) string {
	if m.shouldCacheTrackMeta(track.Id) {
		if p, err := cache.CoverPath(track.Id); err == nil {
			return p
		}
	}
	tempDir := filepath.Join(os.TempDir(), config.DirName)
	if os.MkdirAll(tempDir, 0755) != nil {
		return ""
	}
	return filepath.Join(tempDir, track.Id+".jpg")
}

// shouldCacheTrackMeta reports whether per-track metadata (cover, lyrics)
// should be persisted to disk for the given track, based on cache-tracks scope.
func (m *Model) shouldCacheTrackMeta(trackId string) bool {
	switch config.Current.CacheTracks {
	case config.CACHE_NONE:
		return false
	case config.CACHE_LIKED_ONLY:
		return m.likedTracksMap[trackId]
	case config.CACHE_ALL:
		return true
	}
	return false
}

func (m *Model) metadataFilePath() string {
	tempDir := filepath.Join(os.TempDir(), config.DirName)
	if os.MkdirAll(tempDir, 0755) != nil {
		return ""
	}
	return filepath.Join(tempDir, "metadata.mp3")
}
