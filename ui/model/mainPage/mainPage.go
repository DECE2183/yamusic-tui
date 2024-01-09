package mainpage

import (
	"fmt"
	"net/url"
	"yamusic/api"
	"yamusic/config"
	"yamusic/ui/components/playlist"
	"yamusic/ui/components/tracker"
	"yamusic/ui/components/tracklist"
	"yamusic/ui/helpers"
	"yamusic/ui/model"
	"yamusic/ui/style"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	program       *tea.Program
	client        *api.YaMusicClient
	width, height int

	playlist  playlist.Model
	trackList list.Model
	tracker   tracker.Model

	infinitePlaylist    bool
	currentStationId    api.StationId
	currentStationBatch string

	playQueue       []api.Track
	currentTrackIdx int
	playlistTracks  []api.Track
	currentPlaylist playlist.Item

	likedTracksMap   map[string]bool
	likedTracksSlice []string
}

// mainpage.Model constructor.
func New() *Model {
	m := &Model{}

	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion())
	m.program = p
	m.likedTracksMap = make(map[string]bool)

	m.trackList = list.New([]list.Item{}, tracklist.ItemDelegate{}, 512, 512)
	m.tracker = tracker.New(m.program)

	controls := config.Current.Controls

	m.trackList.Title = "Tracks"
	m.trackList.Styles.Title = m.trackList.Styles.Title.Foreground(style.NormalTextColor).UnsetBackground().Padding(0)
	m.trackList.KeyMap = list.KeyMap{
		CursorUp:     key.NewBinding(controls.TrackListUp.Binding(), controls.TrackListUp.Help("up")),
		CursorDown:   key.NewBinding(controls.TrackListDown.Binding(), controls.TrackListDown.Help("down")),
		Quit:         key.NewBinding(key.WithKeys(""), controls.TrackListLike.Help("like/unlike")),
		Filter:       key.NewBinding(key.WithKeys(""), controls.TrackListSelect.Help("select")),
		ShowFullHelp: key.NewBinding(key.WithKeys(""), controls.TrackListShare.Help("share")),
	}

	m.initialLoad()
	return m
}

//
// model.Model interface implementation
//

func (m *Model) Run() error {
	_, err := m.program.Run()
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
			m.tracker, cmd = m.tracker.Update(message)
			cmds = append(cmds, cmd)
		}

	// playlist control update
	case model.PlaylistControl:

	// tracklist control update
	case model.TracklistControl:

	// player control update
	case model.PlayerControl:
		switch msg {
		case model.PLAYER_NEXT:
			m.nextTrack()
		case model.PLAYER_PREV:
			m.prevTrack()
		}

		m.tracker, cmd = m.tracker.Update(message)
		cmds = append(cmds, cmd)

	// player progress update
	case model.ProgressControl:
		m.tracker, cmd = m.tracker.Update(message)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m *Model) View() string {
	sidePanel := style.SideBoxStyle.Render(m.playlist.View())
	midPanel := lipgloss.JoinVertical(lipgloss.Left, style.TrackBoxStyle.Render(m.trackList.View()), m.tracker.View())
	return lipgloss.JoinHorizontal(lipgloss.Bottom, sidePanel, midPanel)
}

//
// private methods
//

func (m *Model) resize(width, height int) {
	m.width, m.height = width, height
	m.playlist.SetSize(32, height-5)
	m.trackList.SetSize(m.width-m.playlist.Width()-20, height-14)
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

	playlists, err := m.client.ListPlaylists()
	if err == nil {
		for _, pl := range playlists {
			m.playlist.AddItem(playlist.Item{Name: pl.Title, Kind: pl.Kind, Active: true, Subitem: true})
		}
	}

	tracks, err := m.client.StationTracks(api.MyWaveId, nil)
	if err == nil {
		var playlist []list.Item
		m.playlistTracks = m.playlistTracks[:0]
		for _, t := range tracks.Sequence {
			m.playlistTracks = append(m.playlistTracks, t.Track)
			playlist = append(playlist, tracklist.Item{
				Title:      t.Track.Title,
				Version:    t.Track.Version,
				Artists:    helpers.ArtistList(t.Track.Artists),
				Id:         t.Track.Id,
				Liked:      false,
				DurationMs: t.Track.DurationMs,
				Available:  t.Track.Available,
			})
		}
		m.trackList.SetItems(playlist)
		m.currentStationBatch = tracks.BatchId
	}

	likes, err := m.client.LikedTracks()
	if err == nil {
		m.likedTracksSlice = make([]string, 0, len(likes))
		for _, l := range likes {
			m.likedTracksMap[l.Id] = true
			m.likedTracksSlice = append(m.likedTracksSlice, l.Id)
		}
	}
}

func (m *Model) prevTrack() {
	if m.currentTrackIdx == 0 {
		m.Send(model.PLAYER_STOP)
		return
	}

	m.currentTrackIdx--
	m.playTrack(&m.playQueue[m.currentTrackIdx])

	selectedPlaylist := m.playlist.SelectedItem()
	if m.currentPlaylist.Kind == selectedPlaylist.Kind {
		m.Send(selectedPlaylist)
	}
}

func (m *Model) nextTrack() {
	if len(m.playQueue) == 0 {
		return
	}

	if m.infinitePlaylist {
		currTrack := m.playQueue[m.currentTrackIdx]

		if m.tracker.Progress() == 1 {
			go m.client.StationFeedback(
				api.ROTOR_TRACK_FINISHED,
				m.currentStationId,
				m.currentStationBatch,
				currTrack.Id,
				currTrack.DurationMs*1000,
			)
		} else {
			go m.client.StationFeedback(
				api.ROTOR_SKIP,
				m.currentStationId,
				m.currentStationBatch,
				currTrack.Id,
				int(float64(currTrack.DurationMs*1000)*m.tracker.Progress()),
			)
		}

		if m.currentTrackIdx+2 >= len(m.playQueue) {
			tracks, err := m.client.StationTracks(api.MyWaveId, &currTrack)
			if err != nil {
				return
			}

			for _, tr := range tracks.Sequence {
				m.playQueue = append(m.playQueue, tr.Track)
			}
		}
	} else if m.currentTrackIdx+1 >= len(m.playQueue) {
		m.currentTrackIdx = 0
		m.Send(model.PLAYER_STOP)
		return
	}

	m.currentTrackIdx++
	m.playTrack(&m.playQueue[m.currentTrackIdx])

	selectedPlaylis := m.playlist.SelectedItem()
	if m.currentPlaylist.Kind == selectedPlaylis.Kind {
		m.Send(selectedPlaylis)
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

	m.tracker.StartTrack(track, trackReader)

	if m.infinitePlaylist {
		go m.client.StationFeedback(
			api.ROTOR_TRACK_STARTED,
			m.currentStationId,
			m.currentStationBatch,
			track.Id,
			0,
		)
	}

	go m.client.PlayTrack(track, false)
}

func (m *Model) playCurrentQueue(trackIndex int) {
	if len(m.playQueue) == 0 {
		m.Send(model.PLAYER_STOP)
		return
	}

	selectedPlaylist := m.playlist.SelectedItem()
	if m.currentPlaylist.Kind == selectedPlaylist.Kind && m.currentTrackIdx == trackIndex {
		if m.tracker.IsPlaying() {
			m.tracker.Pause()
			return
		} else {
			m.tracker.Play()
			return
		}
	}

	if m.infinitePlaylist {
		go m.client.StationFeedback(
			api.ROTOR_RADIO_STARTED,
			m.currentStationId,
			"",
			"",
			0,
		)
	}

	m.currentTrackIdx = trackIndex
	m.playTrack(&m.playQueue[m.currentTrackIdx])
}
