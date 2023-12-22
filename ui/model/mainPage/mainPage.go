package mainpage

import (
	"fmt"
	"net/url"
	"time"
	"yamusic/api"
	"yamusic/config"
	"yamusic/ui/components/playlist"
	"yamusic/ui/components/tracklist"
	"yamusic/ui/model"
	"yamusic/ui/style"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/ebitengine/oto/v3"
)

type Model struct {
	program       *tea.Program
	client        *api.YaMusicClient
	width, height int

	playlistList  list.Model
	trackList     list.Model
	trackProgress progress.Model
	trackerHelp   help.Model

	playerContext *oto.Context
	player        *oto.Player
	trackWrapper  *readWrapper

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
	var err error
	m := &Model{}

	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion())
	m.program = p

	playlistListItems := []list.Item{
		playlist.Item{Name: "my wave", Kind: uint64(playlist.MYWAVE), Active: true, Subitem: false},
		playlist.Item{Name: "likes", Kind: uint64(playlist.LIKES), Active: true, Subitem: false},
		playlist.Item{Name: "playlists:", Kind: 0, Active: false, Subitem: false},
	}

	m.playlistList = list.New(playlistListItems, playlist.ItemDelegate{}, 512, 512)
	m.trackList = list.New([]list.Item{}, tracklist.ItemDelegate{}, 512, 512)
	m.trackProgress = progress.New(progress.WithSolidFill("#FC0"))
	m.trackerHelp = help.New()

	m.trackWrapper = &readWrapper{program: m.program}
	m.likedTracksMap = make(map[string]bool)

	op := &oto.NewContextOptions{}

	op.SampleRate = 44100
	op.ChannelCount = 2
	op.BufferSize = time.Millisecond * time.Duration(config.Current.BufferSize)
	op.Format = oto.FormatSignedInt16LE

	var readyChan chan struct{}
	m.playerContext, readyChan, err = oto.NewContext(op)
	if err != nil {
		model.PrettyExit(err, 12)
	}
	<-readyChan

	controls := config.Current.Controls

	m.playlistList.Title = "Playlists"
	m.playlistList.SetShowStatusBar(false)
	m.playlistList.Styles.Title = m.playlistList.Styles.Title.Foreground(style.AccentColor).UnsetBackground().Padding(0)
	m.playlistList.KeyMap = list.KeyMap{
		CursorUp:   key.NewBinding(controls.PlaylistsUp.Binding(), controls.PlaylistsUp.Help("up")),
		CursorDown: key.NewBinding(controls.PlaylistsDown.Binding(), controls.PlaylistsUp.Help("down")),
	}

	m.trackList.Title = "Tracks"
	m.trackList.Styles.Title = m.trackList.Styles.Title.Foreground(style.NormalTextColor).UnsetBackground().Padding(0)
	m.trackList.KeyMap = list.KeyMap{
		CursorUp:     key.NewBinding(controls.TrackListUp.Binding(), controls.TrackListUp.Help("up")),
		CursorDown:   key.NewBinding(controls.TrackListDown.Binding(), controls.TrackListDown.Help("down")),
		Quit:         key.NewBinding(key.WithKeys(""), controls.TrackListLike.Help("like/unlike")),
		Filter:       key.NewBinding(key.WithKeys(""), controls.TrackListSelect.Help("select")),
		ShowFullHelp: key.NewBinding(key.WithKeys(""), controls.TrackListShare.Help("share")),
	}

	m.trackProgress.ShowPercentage = false
	m.trackProgress.Empty = m.trackProgress.Full
	m.trackProgress.EmptyColor = "#6b6b6b"

	m.initialLoad()
	return m
}

//
// modal.Modal interface implementation
//

func (m *Model) Run() error {
	_, err := m.program.Run()
	return err
}

func (m *Model) Send(msg tea.Msg) {
	go m.program.Send(msg)
}

//
// tea.Modal interface implementation
//

func (m *Model) Init() tea.Cmd {
	return textinput.Blink
}

//
// private methods
//

func (m *Model) resize(width, height int) {
	m.width, m.height = width, height
	m.playlistList.SetSize(32, height-5)
	m.trackList.SetSize(m.width-m.playlistList.Width()-20, height-14)
	m.trackProgress.Width = m.width - m.playlistList.Width() - 13
	m.trackerHelp.Width = m.trackProgress.Width
}

func (m *Model) initialLoad() {
	var err error
	m.client, err = api.NewClient(config.Current.Token)
	if err != nil {
		if _, ok := err.(*url.Error); ok {
			model.PrettyExit(fmt.Errorf("unable to connect to the Yandex server\n\n"), 14)
		} else {
			model.PrettyExit(err, 16)
		}
	}

	playlistListItems := m.playlistList.Items()

	playlists, err := m.client.ListPlaylists()
	if err == nil {
		for _, pl := range playlists {
			playlistListItems = append(playlistListItems, playlist.Item{Name: pl.Title, Kind: pl.Kind, Active: true, Subitem: true})
		}
	}
	m.playlistList.SetItems(playlistListItems)

	tracks, err := m.client.StationTracks(api.MyWaveId, nil)
	if err == nil {
		var playlist []list.Item
		m.playlistTracks = m.playlistTracks[:0]
		for _, t := range tracks.Sequence {
			m.playlistTracks = append(m.playlistTracks, t.Track)
			playlist = append(playlist, tracklist.Item{
				Title:      t.Track.Title,
				Version:    t.Track.Version,
				Artists:    artistList(t.Track.Artists),
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
