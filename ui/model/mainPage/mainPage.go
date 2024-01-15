package mainpage

import (
	"fmt"
	"net/url"
	"yamusic/api"
	"yamusic/config"
	"yamusic/ui/components/playlist"
	"yamusic/ui/components/tracker"
	"yamusic/ui/components/tracklist"
	"yamusic/ui/model"
	"yamusic/ui/style"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"golang.design/x/clipboard"
)

type Model struct {
	program       *tea.Program
	client        *api.YaMusicClient
	width, height int

	playlist  playlist.Model
	tracklist tracklist.Model
	tracker   tracker.Model

	currentPlaylistIndex int
	likedTracksMap       map[string]bool
}

// mainpage.Model constructor.
func New() *Model {
	m := &Model{}

	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion())
	m.program = p
	m.likedTracksMap = make(map[string]bool)

	m.playlist = playlist.New(m.program)
	m.tracklist = tracklist.New(m.program, &m.likedTracksMap)
	m.tracker = tracker.New(m.program, &m.likedTracksMap)

	m.initialLoad()
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
		default:
			m.playlist, cmd = m.playlist.Update(message)
			cmds = append(cmds, cmd)
			m.tracklist, cmd = m.tracklist.Update(message)
			cmds = append(cmds, cmd)
			m.tracker, cmd = m.tracker.Update(message)
			cmds = append(cmds, cmd)
		}

	// playlist control update
	case playlist.PlaylistControl:
		switch msg {
		case playlist.CURSOR_UP, playlist.CURSOR_DOWN:
			selectedPlaylist := m.playlist.SelectedItem()
			currentPlaylist := m.playlist.Items()[m.currentPlaylistIndex]

			if selectedPlaylist.Kind == currentPlaylist.Kind && len(selectedPlaylist.Tracks) > 0 {
				selectedPlaylist.SelectedTrack = selectedPlaylist.CurrentTrack
				m.playlist.SetItem(m.playlist.Index(), selectedPlaylist)
			}

			tracks := make([]tracklist.Item, 0, len(selectedPlaylist.Tracks))
			for i := range selectedPlaylist.Tracks {
				track := &selectedPlaylist.Tracks[i]
				tracks = append(tracks, tracklist.NewItem(track))
			}

			m.tracklist.SetItems(tracks)
			m.tracklist.Select(selectedPlaylist.SelectedTrack)
			if m.tracker.IsPlaying() {
				m.indicateCurrentTrackPlaying(true)
			}
		}

	// tracklist control update
	case tracklist.TracklistControl:
		switch msg {
		case tracklist.PLAY:
			playlistItem := m.playlist.SelectedItem()
			if !playlistItem.Active {
				break
			}
			m.playCurrentQueue(m.tracklist.Index())
			m.currentPlaylistIndex = m.playlist.Index()
		case tracklist.CURSOR_UP, tracklist.CURSOR_DOWN:
			currentPlaylist := m.playlist.SelectedItem()
			cursorIndex := m.tracklist.Index()
			currentPlaylist.SelectedTrack = cursorIndex
			m.playlist.SetItem(m.playlist.Index(), currentPlaylist)
		case tracklist.LIKE:
			cmd = m.likeSelectedTrack()
			cmds = append(cmds, cmd)
		case tracklist.SHARE:
			track := m.tracklist.SelectedItem().Track
			link := fmt.Sprintf("https://music.yandex.ru/album/%d/track/%s", track.Albums[0].Id, track.Id)
			clipboard.Write(clipboard.FmtText, []byte(link))
		}

	// player control update
	case tracker.PlayerControl:
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

	default:
		m.playlist, cmd = m.playlist.Update(message)
		cmds = append(cmds, cmd)
		m.tracklist, cmd = m.tracklist.Update(message)
		cmds = append(cmds, cmd)
		m.tracker, cmd = m.tracker.Update(message)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m *Model) View() string {
	sidePanel := style.SideBoxStyle.Render(m.playlist.View())
	midPanel := lipgloss.JoinVertical(lipgloss.Left, style.TrackBoxStyle.Render(m.tracklist.View()), m.tracker.View())
	return lipgloss.JoinHorizontal(lipgloss.Bottom, sidePanel, midPanel)
}

//
// private methods
//

func (m *Model) resize(width, height int) {
	m.width, m.height = width, height
	m.playlist.SetSize(32, height-5)
	m.tracklist.SetSize(m.width-m.playlist.Width()-20, height-14)
	m.tracker.SetWidth(m.width - m.playlist.Width())
}

func (m *Model) initialLoad() {
	var err error
	m.client, err = api.NewClient(config.Current.Token)
	if err != nil {
		if _, ok := err.(*url.Error); ok {
			model.PrettyExit(fmt.Errorf("unable to connect to the Yandex server"), 14)
		} else {
			model.PrettyExit(err, 16)
		}
	}

	for i, station := range m.playlist.Items() {
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
			m.playlist.SetItem(i, station)
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
			m.playlist.SetItem(i, station)
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

			m.playlist.InsertItem(-1, playlist.Item{
				Name:    pl.Title,
				Kind:    pl.Kind,
				Active:  true,
				Subitem: true,
				Tracks:  playlistTracks,
			})
		}
	}

	m.playlist.Select(0)
	m.Send(playlist.CURSOR_UP)
}

func (m *Model) prevTrack() {
	currentPlaylist := m.playlist.Items()[m.currentPlaylistIndex]

	if currentPlaylist.CurrentTrack == 0 {
		m.Send(tracker.STOP)
		return
	}

	m.indicateCurrentTrackPlaying(false)

	currentPlaylist.CurrentTrack--
	m.playlist.SetItem(m.currentPlaylistIndex, currentPlaylist)
	m.playTrack(&currentPlaylist.Tracks[currentPlaylist.CurrentTrack])

	selectedPlaylist := m.playlist.SelectedItem()
	if currentPlaylist.Kind == selectedPlaylist.Kind && m.tracklist.Index() == currentPlaylist.CurrentTrack+1 {
		m.tracklist.Select(currentPlaylist.CurrentTrack)
	}
}

func (m *Model) nextTrack() {
	currentPlaylist := m.playlist.Items()[m.currentPlaylistIndex]

	if len(currentPlaylist.Tracks) == 0 {
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
				if m.playlist.SelectedItem().Kind == currentPlaylist.Kind {
					newTrack := &currentPlaylist.Tracks[len(currentPlaylist.Tracks)-1]
					m.tracklist.InsertItem(-1, tracklist.NewItem(newTrack))
				}
			}
		}
	} else if currentPlaylist.CurrentTrack+1 >= len(currentPlaylist.Tracks) {
		currentPlaylist.CurrentTrack = 0
		m.playlist.SetItem(m.currentPlaylistIndex, currentPlaylist)
		m.Send(tracker.STOP)
		return
	}

	currentPlaylist.CurrentTrack++
	m.playlist.SetItem(m.currentPlaylistIndex, currentPlaylist)
	m.playTrack(&currentPlaylist.Tracks[currentPlaylist.CurrentTrack])

	selectedPlaylist := m.playlist.SelectedItem()
	if currentPlaylist.Kind == selectedPlaylist.Kind && m.tracklist.Index() == currentPlaylist.CurrentTrack-1 {
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

	currentPlaylist := m.playlist.Items()[m.currentPlaylistIndex]
	if currentPlaylist.Infinite {
		go m.client.StationFeedback(
			api.ROTOR_TRACK_STARTED,
			currentPlaylist.StationId,
			currentPlaylist.StationBatch,
			track.Id,
			0,
		)
	}

	go m.client.PlayTrack(track, false)
}

func (m *Model) playCurrentQueue(trackIndex int) {
	currentPlaylist := m.playlist.Items()[m.currentPlaylistIndex]

	if len(currentPlaylist.Tracks) == 0 {
		m.Send(tracker.STOP)
		return
	}

	m.indicateCurrentTrackPlaying(false)
	selectedPlaylist := m.playlist.SelectedItem()
	if currentPlaylist.Kind == selectedPlaylist.Kind && currentPlaylist.CurrentTrack == trackIndex {
		if m.tracker.IsPlaying() {
			m.tracker.Pause()
			return
		} else {
			m.tracker.Play()
			return
		}
	}

	currentPlaylist.CurrentTrack = trackIndex
	trackToPlay := &currentPlaylist.Tracks[currentPlaylist.CurrentTrack]

	if currentPlaylist.Infinite {
		if m.tracker.IsPlaying() {
			currentTrack := m.tracker.CurrentTrack()
			go m.client.StationFeedback(
				api.ROTOR_SKIP,
				currentPlaylist.StationId,
				currentPlaylist.StationBatch,
				currentTrack.Id,
				int(float64(currentTrack.DurationMs*1000)*m.tracker.Progress()),
			)
			go m.client.StationFeedback(
				api.ROTOR_TRACK_STARTED,
				currentPlaylist.StationId,
				currentPlaylist.StationBatch,
				trackToPlay.Id,
				0,
			)
		} else {
			go m.client.StationFeedback(
				api.ROTOR_RADIO_STARTED,
				currentPlaylist.StationId,
				"",
				"",
				0,
			)
		}
	}

	m.playlist.SetItem(m.currentPlaylistIndex, currentPlaylist)
	m.playTrack(trackToPlay)
}

func (m *Model) likePlayingTrack() tea.Cmd {
	track := m.tracker.CurrentTrack()
	return m.likeTrack(track)
}

func (m *Model) likeSelectedTrack() tea.Cmd {
	currentPlaylist := m.playlist.Items()[m.currentPlaylistIndex]
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

		likedPlaylist := m.playlist.Items()[1]
		for i, ltrack := range likedPlaylist.Tracks {
			if ltrack.Id == track.Id {
				if i+1 < len(likedPlaylist.Tracks) {
					likedPlaylist.Tracks = append(likedPlaylist.Tracks[:i], likedPlaylist.Tracks[i+1:]...)
				} else {
					likedPlaylist.Tracks = likedPlaylist.Tracks[:i]
				}

				if m.playlist.SelectedItem().Kind == playlist.LIKES {
					m.tracklist.RemoveItem(i)
				}
				break
			}
		}

		return m.playlist.SetItem(1, likedPlaylist)
	} else {
		if m.client.LikeTrack(track.Id) != nil {
			return nil
		}

		m.likedTracksMap[track.Id] = true
		likedPlaylist := m.playlist.Items()[1]
		likedPlaylist.Tracks = append([]api.Track{*track}, likedPlaylist.Tracks...)
		return m.playlist.SetItem(1, likedPlaylist)
	}
}

func (m *Model) indicateCurrentTrackPlaying(playing bool) {
	currentPlaylist := m.playlist.Items()[m.currentPlaylistIndex]
	if currentPlaylist.Kind == m.playlist.SelectedItem().Kind && currentPlaylist.CurrentTrack < len(m.tracklist.Items()) {
		track := m.tracklist.Items()[currentPlaylist.CurrentTrack]
		track.IsPlaying = playing
		m.tracklist.SetItem(currentPlaylist.CurrentTrack, track)
	}
}
