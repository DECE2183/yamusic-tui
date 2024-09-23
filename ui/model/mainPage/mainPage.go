package mainpage

import (
	"fmt"
	"math/rand"
	"net/url"
	"slices"
	"strings"

	"github.com/dece2183/yamusic-tui/api"
	"github.com/dece2183/yamusic-tui/config"
	"github.com/dece2183/yamusic-tui/ui/components/playlist"
	"github.com/dece2183/yamusic-tui/ui/components/search"
	"github.com/dece2183/yamusic-tui/ui/components/tracker"
	"github.com/dece2183/yamusic-tui/ui/components/tracklist"
	"github.com/dece2183/yamusic-tui/ui/style"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"golang.design/x/clipboard"
)

type Model struct {
	program       *tea.Program
	client        *api.YaMusicClient
	width, height int

	playlists   *playlist.Model
	tracklist   *tracklist.Model
	tracker     *tracker.Model
	search      *search.Model
	addPlaylist *search.Model

	isSearchActive      bool
	isAddPlaylistActive bool

	currentPlaylistIndex int
	likedTracksMap       map[string]bool
}

// mainpage.Model constructor.
func New() *Model {
	m := &Model{}

	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion())
	m.program = p
	m.likedTracksMap = make(map[string]bool)

	m.playlists = playlist.New(m.program, "YaMusic")
	m.tracklist = tracklist.New(m.program, &m.likedTracksMap)
	m.tracker = tracker.New(m.program, &m.likedTracksMap)
	m.search = search.New("Search", "search")
	m.addPlaylist = search.New("Add to playlist", "add")

	return m
}

//
// model.Model interface implementation
//

func (m *Model) Run() error {
	err := clipboard.Init()
	if err != nil {
		return err
	}

	err = m.initialLoad()
	if err != nil {
		return err
	}

	_, err = m.program.Run()
	return err
}

func (m *Model) Send(msg tea.Msg) {
	go m.program.Send(msg)
}

//
// tea.Model interface implementation
//

func (m *Model) Init() tea.Cmd {
	return textinput.Blink
}

func (m *Model) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := message.(type) {
	case tea.WindowSizeMsg:
		m.resize(msg.Width, msg.Height)
		return m, tea.ClearScreen

	case tea.KeyMsg:
		controls := config.Current.Controls
		keypress := msg.String()

		switch {
		case controls.Quit.Contains(keypress):
			return m, tea.Quit
		case m.isSearchActive:
			m.search, cmd = m.search.Update(message)
			cmds = append(cmds, cmd)
		case m.isAddPlaylistActive:
			m.addPlaylist, cmd = m.addPlaylist.Update(message)
			cmds = append(cmds, cmd)
		default:
			m.playlists, cmd = m.playlists.Update(message)
			cmds = append(cmds, cmd)
			m.tracklist, cmd = m.tracklist.Update(message)
			cmds = append(cmds, cmd)
			m.tracker, cmd = m.tracker.Update(message)
			cmds = append(cmds, cmd)
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

			if m.tracker.IsPlaying() {
				m.indicateCurrentTrackPlaying(true)
			}

			m.tracklist.Shufflable = (selectedPlaylist.Kind != playlist.NONE && selectedPlaylist.Kind != playlist.MYWAVE && len(selectedPlaylist.Tracks) > 0)
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
			m.playlists.SetItem(m.playlists.Index(), currentPlaylist)
		case tracklist.LIKE:
			cmd = m.likeSelectedTrack()
			cmds = append(cmds, cmd)
		case tracklist.ADD_TO_PLAYLIST:
			m.isAddPlaylistActive = true
			m.Send(search.UPDATE_SUGGESTIONS)
		case tracklist.REMOVE_FROM_PLAYLIST:
			selectedPlaylist := m.playlists.SelectedItem()
			if selectedPlaylist.Kind == playlist.NONE || selectedPlaylist.Kind == playlist.MYWAVE {
				break
			}

			selectedTrackIndex := selectedPlaylist.SelectedTrack
			if selectedPlaylist.Kind == playlist.LIKES {
				selectedTrack := selectedPlaylist.Tracks[selectedTrackIndex]
				cmd = m.likeTrack(&selectedTrack)
				cmds = append(cmds, cmd)
				break
			}

			if len(selectedPlaylist.Tracks) < 2 {
				err := m.client.RemovePlaylist(selectedPlaylist.Kind)
				if err != nil {
					break
				}
				playlists := m.playlists.Items()
				if m.currentPlaylistIndex >= 0 {
					currentPlaylist := playlists[m.currentPlaylistIndex]
					if selectedPlaylist.IsSame(currentPlaylist) && m.tracker.IsPlaying() {
						m.currentPlaylistIndex = -1
					}
				}
				m.playlists.RemoveItem(m.playlists.Index())
				if len(playlists) <= 1 {
					m.playlists.Select(0)
				}
				m.displayPlaylist(m.playlists.SelectedItem())
				break
			}

			_, err := m.client.RemoveFromPlaylist(selectedPlaylist.Kind, selectedPlaylist.Revision, selectedTrackIndex)
			if err != nil {
				break
			}

			selectedPlaylist.Revision++
			selectedPlaylist.Tracks = slices.Delete(selectedPlaylist.Tracks, selectedTrackIndex, selectedTrackIndex+1)
			selectedPlaylist.SelectedTrack = selectedTrackIndex
			m.playlists.SetItem(m.playlists.Index(), selectedPlaylist)
			m.displayPlaylist(selectedPlaylist)

			if m.currentPlaylistIndex >= 0 {
				currentPlaylist := m.playlists.Items()[m.currentPlaylistIndex]
				if selectedPlaylist.IsSame(currentPlaylist) && m.tracker.IsPlaying() {
					m.indicateCurrentTrackPlaying(true)
				}
			}
		case tracklist.SEARCH:
			m.isSearchActive = true
			m.Send(search.UPDATE_SUGGESTIONS)
		case tracklist.SHUFFLE:
			selectedPlaylist := m.playlists.SelectedItem()
			if selectedPlaylist.Kind == playlist.NONE || selectedPlaylist.Kind == playlist.MYWAVE || len(selectedPlaylist.Tracks) == 0 {
				break
			}

			currentTrackIndex := selectedPlaylist.CurrentTrack
			selectedTrackIndex := selectedPlaylist.SelectedTrack
			currentTrack := selectedPlaylist.Tracks[currentTrackIndex]
			selectedTrack := selectedPlaylist.Tracks[selectedTrackIndex]

			tracks := make([]api.Track, len(selectedPlaylist.Tracks))
			trackList := make([]tracklist.Item, len(selectedPlaylist.Tracks))
			perm := rand.Perm(len(tracks))

			for i, v := range perm {
				tracks[v] = selectedPlaylist.Tracks[i]
				trackList[v] = tracklist.NewItem(&tracks[v])
				if currentTrack.Id == tracks[v].Id {
					currentTrackIndex = v
				}
				if selectedTrackIndex > 0 && selectedTrack.Id == tracks[v].Id {
					selectedTrackIndex = v
				}
			}

			selectedPlaylist.Tracks = tracks
			selectedPlaylist.SelectedTrack = selectedTrackIndex
			selectedPlaylist.CurrentTrack = currentTrackIndex
			m.playlists.SetItem(m.playlists.Index(), selectedPlaylist)
			m.tracklist.SetItems(trackList)
			m.tracklist.Select(selectedTrackIndex)

			if m.currentPlaylistIndex >= 0 {
				currentPlaylist := m.playlists.Items()[m.currentPlaylistIndex]
				if selectedPlaylist.IsSame(currentPlaylist) && m.tracker.IsPlaying() {
					m.indicateCurrentTrackPlaying(true)
				}
			}
		case tracklist.SHARE:
			track := m.tracklist.SelectedItem().Track
			link := fmt.Sprintf("https://music.yandex.ru/album/%d/track/%s", track.Albums[0].Id, track.Id)
			clipboard.Write(clipboard.FmtText, []byte(link))
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

	default:
		if m.isSearchActive {
			m.search, cmd = m.search.Update(message)
			cmds = append(cmds, cmd)
		} else if m.isAddPlaylistActive {
			m.addPlaylist, cmd = m.addPlaylist.Update(message)
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
	if m.isSearchActive {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, m.search.View())
	} else if m.isAddPlaylistActive {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, m.addPlaylist.View())
	}

	var sidePanel string
	if m.playlists.Width() > 0 {
		sidePanel = style.SideBoxStyle.Render(m.playlists.View())
	}
	midPanel := lipgloss.JoinVertical(lipgloss.Left, style.TrackBoxStyle.Render(m.tracklist.View()), m.tracker.View())
	return lipgloss.JoinHorizontal(lipgloss.Bottom, sidePanel, midPanel)
}

//
// private methods
//

func (m *Model) resize(width, height int) {
	m.width, m.height = width, height
	if m.width > style.PlaylistsSidePanelWidth*3 {
		m.playlists.SetSize(style.PlaylistsSidePanelWidth, height-4)
	} else {
		m.playlists.SetSize(0, height-4)
	}
	m.tracklist.SetSize(m.width-m.playlists.Width()-4, height-14)
	m.tracker.SetWidth(m.width - m.playlists.Width() - 4)

	searchWidth := style.SearchModalWidth
	if searchWidth > width {
		searchWidth = width - 2
	}

	m.search.SetSize(searchWidth, height-4)
	m.addPlaylist.SetSize(searchWidth, height-4)
}

func (m *Model) initialLoad() error {
	var err error
	if len(config.Current.Token) == 0 {
		return fmt.Errorf("wrong token")
	}

	m.client, err = api.NewClient(config.Current.Token)
	if err != nil {
		if _, ok := err.(*url.Error); ok {
			return fmt.Errorf("unable to connect to the Yandex server")
		} else {
			return err
		}
	}

	for i, station := range m.playlists.Items() {
		switch station.Kind {
		case playlist.MYWAVE:
			tracks, err := m.client.StationTracks(api.MyWaveId, nil)
			if err != nil {
				continue
			}

			station.StationId = tracks.Id
			station.StationBatch = tracks.BatchId
			for _, t := range tracks.Sequence {
				station.Tracks = append(station.Tracks, t.Track)
			}
			m.playlists.SetItem(i, station)
		case playlist.LIKES:
			likes, err := m.client.LikedTracks()
			if err != nil {
				continue
			}

			likedTracksId := make([]string, len(likes))
			for l, track := range likes {
				m.likedTracksMap[track.Id] = true
				likedTracksId[l] = track.Id
			}

			likedTracks, err := m.client.Tracks(likedTracksId)
			if err != nil {
				continue
			}

			station.Tracks = likedTracks
			m.playlists.SetItem(i, station)
		default:
		}
	}

	playlists, err := m.client.ListPlaylists()
	if err == nil {
		for _, pl := range playlists {
			playlistTracks, err := m.client.PlaylistTracks(pl.Kind, false)
			if err != nil {
				continue
			}

			m.playlists.InsertItem(-1, playlist.Item{
				Name:     pl.Title,
				Kind:     pl.Kind,
				Revision: pl.Revision,
				Active:   true,
				Subitem:  true,
				Tracks:   playlistTracks,
			})
		}
	}

	m.playlists.Select(0)
	m.Send(playlist.CURSOR_UP)

	return nil
}

func (m *Model) displayPlaylist(pl playlist.Item) {
	trackList := make([]tracklist.Item, len(pl.Tracks))
	for i := range pl.Tracks {
		trackList[i] = tracklist.NewItem(&pl.Tracks[i])
	}
	m.tracklist.SetItems(trackList)
	m.tracklist.Select(pl.SelectedTrack)
	switch pl.Kind {
	case playlist.MYWAVE:
		m.tracklist.Title = "My wave"
	case playlist.LIKES:
		m.tracklist.Title = "Liked tracks"
	default:
		m.tracklist.Title = "Tracks from " + pl.Name
	}
}

func (m *Model) searchControl(msg search.Control) tea.Cmd {
	var cmd tea.Cmd

	switch msg {
	case search.SELECT:
		m.isSearchActive = false

		req, ok := m.search.SuggestionValue()
		if !ok {
			return nil
		}

		searchRes, err := m.client.Search(req, api.SEARCH_ALL)
		if err != nil {
			return nil
		}

		cmd = m.displaySearchResults(searchRes)
	case search.CANCEL:
		m.isSearchActive = false
	case search.UPDATE_SUGGESTIONS:
		suggestions, err := m.client.SearchSuggest(m.search.InputValue())
		if err != nil {
			return nil
		}
		m.search.SetSuggestions(suggestions.Suggestions)
	}

	return cmd
}

func (m *Model) addPlaylistControl(msg search.Control) tea.Cmd {
	var cmd tea.Cmd

	switch msg {
	case search.SELECT:
		m.isAddPlaylistActive = false

		selectedPlaylist := m.playlists.SelectedItem()
		if len(selectedPlaylist.Tracks) == 0 {
			return nil
		}

		playlists := m.playlists.Items()
		inputVal, ok := m.addPlaylist.SuggestionValue()
		if !ok {
			return nil
		}

		foundPlaylistIndex := -1
		var foundPlaylist *playlist.Item
		for i := range playlists {
			if playlists[i].Active && playlists[i].Kind >= playlist.USER {
				if strings.EqualFold(playlists[i].Name, inputVal) {
					foundPlaylist = &playlists[i]
					foundPlaylistIndex = i
					break
				} else if foundPlaylistIndex < 0 {
					foundPlaylistIndex = i
				}
			}
		}

		if foundPlaylist == nil {
			pl, err := m.client.CreatePlaylist(inputVal, true)
			if err != nil {
				return nil
			}

			foundPlaylist = &playlist.Item{
				Name:     pl.Title,
				Kind:     pl.Kind,
				Revision: pl.Revision,
				Active:   true,
				Subitem:  true,
			}

			m.playlists.InsertItem(foundPlaylistIndex, *foundPlaylist)
			if foundPlaylistIndex < m.playlists.Index() {
				m.playlists.Select(m.playlists.Index() + 1)
			}
		}

		if selectedPlaylist.Kind == foundPlaylist.Kind {
			return nil
		}

		selectedTrack := &selectedPlaylist.Tracks[selectedPlaylist.SelectedTrack]
		_, err := m.client.AddToPlaylist(foundPlaylist.Kind, foundPlaylist.Revision, len(foundPlaylist.Tracks), selectedTrack.Id)
		if err != nil {
			return nil
		}

		foundPlaylist.Revision++
		foundPlaylist.Tracks = append(foundPlaylist.Tracks, *selectedTrack)
		cmd = m.playlists.SetItem(foundPlaylistIndex, *foundPlaylist)

		m.isAddPlaylistActive = false
	case search.CANCEL:
		m.isAddPlaylistActive = false
	case search.UPDATE_SUGGESTIONS:
		inputVal := strings.ToLower(m.addPlaylist.InputValue())
		playlists := m.playlists.Items()
		suggestions := make([]string, 0, len(playlists))
		for _, pl := range playlists {
			if !pl.Active || pl.Kind < playlist.USER || (len(inputVal) > 0 && !strings.Contains(strings.ToLower(pl.Name), inputVal)) {
				continue
			}
			suggestions = append(suggestions, pl.Name)
		}
		m.addPlaylist.SetSuggestions(suggestions)
	}

	return cmd
}

func (m *Model) prevTrack() {
	if m.currentPlaylistIndex < 0 {
		return
	}

	currentPlaylist := m.playlists.Items()[m.currentPlaylistIndex]
	if len(currentPlaylist.Tracks) == 0 || currentPlaylist.CurrentTrack == 0 {
		m.Send(tracker.STOP)
		return
	}

	m.indicateCurrentTrackPlaying(false)

	currentPlaylist.CurrentTrack--
	m.playlists.SetItem(m.currentPlaylistIndex, currentPlaylist)
	m.playTrack(&currentPlaylist.Tracks[currentPlaylist.CurrentTrack])

	selectedPlaylist := m.playlists.SelectedItem()
	if currentPlaylist.IsSame(selectedPlaylist) && m.tracklist.Index() == currentPlaylist.CurrentTrack+1 {
		m.tracklist.Select(currentPlaylist.CurrentTrack)
	}
}

func (m *Model) nextTrack() {
	if m.currentPlaylistIndex < 0 {
		return
	}

	currentPlaylist := m.playlists.Items()[m.currentPlaylistIndex]
	if len(currentPlaylist.Tracks) == 0 {
		m.Send(tracker.STOP)
		return
	}

	m.indicateCurrentTrackPlaying(false)

	if currentPlaylist.Infinite {
		currTrack := currentPlaylist.Tracks[currentPlaylist.CurrentTrack]

		if m.tracker.Progress() == 1 {
			go m.client.StationFeedback(
				api.ROTOR_TRACK_FINISHED,
				currentPlaylist.StationId,
				currentPlaylist.StationBatch,
				currTrack.Id,
				currTrack.DurationMs*1000,
			)
		} else {
			go m.client.StationFeedback(
				api.ROTOR_SKIP,
				currentPlaylist.StationId,
				currentPlaylist.StationBatch,
				currTrack.Id,
				int(float64(currTrack.DurationMs*1000)*m.tracker.Progress()),
			)
		}

		if currentPlaylist.CurrentTrack+2 >= len(currentPlaylist.Tracks) {
			tracks, err := m.client.StationTracks(api.MyWaveId, &currTrack)
			if err != nil {
				return
			}

			for _, tr := range tracks.Sequence {
				// automatic append new tracks to the track list if this playlist is selected
				currentPlaylist.Tracks = append(currentPlaylist.Tracks, tr.Track)
				if m.playlists.SelectedItem().IsSame(currentPlaylist) {
					newTrack := &currentPlaylist.Tracks[len(currentPlaylist.Tracks)-1]
					m.tracklist.InsertItem(-1, tracklist.NewItem(newTrack))
				}
			}
		}
	} else if currentPlaylist.CurrentTrack+1 >= len(currentPlaylist.Tracks) {
		currentPlaylist.CurrentTrack = 0
		m.playlists.SetItem(m.currentPlaylistIndex, currentPlaylist)
		m.Send(tracker.STOP)
		return
	}

	currentPlaylist.CurrentTrack++
	m.playlists.SetItem(m.currentPlaylistIndex, currentPlaylist)
	m.playTrack(&currentPlaylist.Tracks[currentPlaylist.CurrentTrack])

	selectedPlaylist := m.playlists.SelectedItem()
	if currentPlaylist.IsSame(selectedPlaylist) && m.tracklist.Index() == currentPlaylist.CurrentTrack-1 {
		m.tracklist.Select(currentPlaylist.CurrentTrack)
	}
}

func (m *Model) playTrack(track *api.Track) {
	m.tracker.Stop()

	dowInfo, err := m.client.TrackDownloadInfo(track.Id)
	if err != nil {
		return
	}

	var bestBitrate int
	var bestTrackInfo api.TrackDownloadInfo
	for _, t := range dowInfo {
		if t.BbitrateInKbps > bestBitrate {
			bestBitrate = t.BbitrateInKbps
			bestTrackInfo = t
		}
	}

	trackReader, _, err := m.client.DownloadTrack(bestTrackInfo)
	if err != nil {
		return
	}

	m.indicateCurrentTrackPlaying(true)
	m.tracker.StartTrack(track, trackReader)

	if m.currentPlaylistIndex >= 0 {
		currentPlaylist := m.playlists.Items()[m.currentPlaylistIndex]
		if currentPlaylist.Infinite {
			go m.client.StationFeedback(
				api.ROTOR_TRACK_STARTED,
				currentPlaylist.StationId,
				currentPlaylist.StationBatch,
				track.Id,
				0,
			)
		}
	}

	go m.client.PlayTrack(track, false)
}

func (m *Model) playSelectedPlaylist(trackIndex int) {
	selectedPlaylist := m.playlists.SelectedItem()
	if len(selectedPlaylist.Tracks) == 0 {
		m.Send(tracker.STOP)
		return
	}

	trackToPlay := &selectedPlaylist.Tracks[selectedPlaylist.SelectedTrack]

	if m.currentPlaylistIndex >= 0 {
		currentPlaylist := m.playlists.Items()[m.currentPlaylistIndex]
		if currentPlaylist.IsSame(selectedPlaylist) && m.tracker.CurrentTrack().Id == trackToPlay.Id {
			if m.tracker.IsPlaying() {
				m.tracker.Pause()
				return
			} else {
				m.tracker.Play()
				return
			}
		}
	}

	m.indicateCurrentTrackPlaying(false)
	selectedPlaylist.CurrentTrack = trackIndex

	if selectedPlaylist.Infinite {
		if m.tracker.IsPlaying() {
			currentTrack := m.tracker.CurrentTrack()
			go m.client.StationFeedback(
				api.ROTOR_SKIP,
				selectedPlaylist.StationId,
				selectedPlaylist.StationBatch,
				currentTrack.Id,
				int(float64(currentTrack.DurationMs*1000)*m.tracker.Progress()),
			)
			go m.client.StationFeedback(
				api.ROTOR_TRACK_STARTED,
				selectedPlaylist.StationId,
				selectedPlaylist.StationBatch,
				trackToPlay.Id,
				0,
			)
		} else {
			go m.client.StationFeedback(
				api.ROTOR_RADIO_STARTED,
				selectedPlaylist.StationId,
				"",
				"",
				0,
			)
		}
	}

	m.currentPlaylistIndex = m.playlists.Index()
	m.playlists.SetItem(m.currentPlaylistIndex, selectedPlaylist)
	m.playTrack(trackToPlay)
}

func (m *Model) likePlayingTrack() tea.Cmd {
	track := m.tracker.CurrentTrack()
	return m.likeTrack(track)
}

func (m *Model) likeSelectedTrack() tea.Cmd {
	if m.currentPlaylistIndex < 0 {
		return nil
	}

	currentPlaylist := m.playlists.Items()[m.currentPlaylistIndex]
	if len(currentPlaylist.Tracks) == 0 {
		return nil
	}

	track := m.tracklist.SelectedItem().Track
	return m.likeTrack(track)
}

func (m *Model) likeTrack(track *api.Track) tea.Cmd {
	if m.likedTracksMap[track.Id] {
		if m.client.UnlikeTrack(track.Id) != nil {
			return nil
		}

		delete(m.likedTracksMap, track.Id)

		likedPlaylist := m.playlists.Items()[1]
		for i, ltrack := range likedPlaylist.Tracks {
			if ltrack.Id == track.Id {
				if i+1 < len(likedPlaylist.Tracks) {
					likedPlaylist.Tracks = append(likedPlaylist.Tracks[:i], likedPlaylist.Tracks[i+1:]...)
				} else {
					likedPlaylist.Tracks = likedPlaylist.Tracks[:i]
				}

				if m.playlists.SelectedItem().Kind == playlist.LIKES {
					m.tracklist.RemoveItem(i)
				}
				break
			}
		}

		return m.playlists.SetItem(1, likedPlaylist)
	} else {
		if m.client.LikeTrack(track.Id) != nil {
			return nil
		}

		m.likedTracksMap[track.Id] = true
		likedPlaylist := m.playlists.Items()[1]
		likedPlaylist.Tracks = append([]api.Track{*track}, likedPlaylist.Tracks...)
		return m.playlists.SetItem(1, likedPlaylist)
	}
}

func (m *Model) indicateCurrentTrackPlaying(playing bool) {
	if m.currentPlaylistIndex < 0 {
		return
	}
	currentPlaylist := m.playlists.Items()[m.currentPlaylistIndex]
	if currentPlaylist.IsSame(m.playlists.SelectedItem()) && currentPlaylist.CurrentTrack < len(m.tracklist.Items()) {
		track := m.tracklist.Items()[currentPlaylist.CurrentTrack]
		track.IsPlaying = playing
		m.tracklist.SetItem(currentPlaylist.CurrentTrack, track)
	}
}

func (m *Model) displaySearchResults(res api.SearchResult) tea.Cmd {
	playlists := m.playlists.Items()
	searchResIndex := len(playlists) + 2
	for i, pl := range playlists {
		if !pl.Active && !pl.Subitem && pl.Name == "search results:" {
			playlists = playlists[:i-1]
			searchResIndex = i + 1
			break
		}
	}

	playlists = append(playlists,
		playlist.Item{Name: "", Kind: playlist.NONE, Active: false, Subitem: false},
		playlist.Item{Name: "search results:", Kind: playlist.NONE, Active: false, Subitem: false},
	)

	if len(res.Tracks.Results) > 0 {
		playlists = append(playlists, playlist.Item{
			Name:    "tracks",
			Active:  true,
			Subitem: true,
			Tracks:  res.Tracks.Results,
		})
	}

	if len(res.Artists.Results) > 0 {
		for _, artist := range res.Artists.Results {
			if !strings.Contains(strings.ToLower(artist.Name), strings.ToLower(res.Text)) {
				continue
			}

			artistTracks, err := m.client.ArtistPopularTracks(artist.Id)
			if err != nil {
				continue
			}

			tracks, err := m.client.Tracks(artistTracks.Tracks)
			if err != nil {
				continue
			}

			playlists = append(playlists, playlist.Item{
				Name:    artist.Name,
				Active:  true,
				Subitem: true,
				Tracks:  tracks,
			})
		}
	}

	cmd := m.playlists.SetItems(playlists)
	m.playlists.Select(searchResIndex)
	m.Send(playlist.CURSOR_DOWN)

	return cmd
}
