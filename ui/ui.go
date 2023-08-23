package ui

import (
	"io"
	"time"
	"yamusic/api"
	"yamusic/config"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ebitengine/oto/v3"
	"github.com/hajimehoshi/go-mp3"
)

type page uint

const (
	_PAGE_LOGIN page = iota
	_PAGE_MAIN  page = iota
	_PAGE_QUIT  page = iota
)

type trackReaderWrapper struct {
	decoder         *mp3.Decoder
	trackReader     io.ReadCloser
	trackDurationMs int
	lastUpdateTime  time.Time
	trackStartTime  time.Time
}

type model struct {
	client *api.YaMusicClient

	width, height int
	page          page

	loginTextInput textinput.Model
	playlistList   list.Model
	trackList      list.Model
	trackProgress  progress.Model

	playerContext   *oto.Context
	player          *oto.Player
	playQueue       []api.Track
	infiniteQueue   bool
	currentTrackIdx int
	trackWrapper    *trackReaderWrapper
}

type playerControl uint

type progressControl float64

const (
	_PLAYER_PLAY  playerControl = iota
	_PLAYER_PAUSE playerControl = iota
	_PLAYER_NEXT  playerControl = iota
	_PLAYER_PREV  playerControl = iota
)

var (
	programm *tea.Program
)

func Run(client *api.YaMusicClient) {
	var err error

	m := model{
		client: client,
		page:   _PAGE_MAIN,

		loginTextInput: textinput.New(),
		playlistList:   list.New([]list.Item{}, playlistListItemDelegate{}, 512, 512),
		trackList:      list.New([]list.Item{}, trackListItemDelegate{}, 512, 512),
		trackProgress:  progress.New(progress.WithSolidFill("#FC0")),

		trackWrapper: &trackReaderWrapper{},
	}

	op := &oto.NewContextOptions{}

	op.SampleRate = 44100
	op.ChannelCount = 2
	op.Format = oto.FormatSignedInt16LE

	var readyChan chan struct{}
	m.playerContext, readyChan, err = oto.NewContext(op)
	if err != nil {
		panic("oto.NewContext failed: " + err.Error())
	}
	<-readyChan

	m.loginTextInput.Width = 64
	m.loginTextInput.CharLimit = 60

	m.playlistList.SetShowHelp(false)
	m.playlistList.Styles.StatusBar = m.playlistList.Styles.StatusBar.Background(lipgloss.Color("#121212"))
	m.playlistList.Styles.TitleBar = m.playlistList.Styles.TitleBar.Background(lipgloss.Color("#121212"))
	m.playlistList.KeyMap = list.KeyMap{
		CursorUp:   key.NewBinding(key.WithKeys("ctrl+up")),
		CursorDown: key.NewBinding(key.WithKeys("ctrl+down")),
	}

	m.trackList.SetShowHelp(false)
	m.trackList.Styles.StatusBar = m.trackList.Styles.StatusBar.Background(lipgloss.Color("#181818"))
	m.trackList.Styles.TitleBar = m.trackList.Styles.TitleBar.Background(lipgloss.Color("#181818"))
	m.trackList.KeyMap = list.KeyMap{
		CursorUp:   key.NewBinding(key.WithKeys("up")),
		CursorDown: key.NewBinding(key.WithKeys("down")),
	}

	m.trackProgress.ShowPercentage = false
	m.trackProgress.Empty = m.trackProgress.Full
	m.trackProgress.EmptyColor = "#6b6b6b"

	playlistListItems := []list.Item{
		playlistListItem{"wave", 0, true, false},
		playlistListItem{"likes", 0, true, false},
		playlistListItem{"dislikes", 0, true, false},
		playlistListItem{"playlists:", 0, false, false},
	}

	if config.GetToken() == "" {
		m.playlistList.SetItems(playlistListItems)
		m.page = _PAGE_LOGIN
		m.loginTextInput.Focus()
	} else {
		playlists, err := m.client.ListPlaylists()
		if err == nil {
			for _, playlist := range playlists {
				playlistListItems = append(playlistListItems, playlistListItem{playlist.Title, playlist.Uid, true, true})
			}
		}

		m.playlistList.SetItems(playlistListItems)
	}

	programm = tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion())
	programm.Run()
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := message.(type) {
	case tea.WindowSizeMsg:
		m.resize(msg.Width, msg.Height)
		return m, nil

	case tea.KeyMsg:
		keypress := msg.String()
		if keypress == "esc" || keypress == "ctrl+q" || keypress == "ctrl+c" {
			m.page = _PAGE_QUIT
			return m, tea.Quit
		}

		switch m.page {
		case _PAGE_LOGIN:
			if keypress == "enter" {
				err := config.SaveToken(m.loginTextInput.Value())
				if err != nil {
					return m, nil
				}
				m.page = _PAGE_MAIN
				return m, nil
			}
		case _PAGE_MAIN:
			if keypress == "enter" {
				playlistItem := m.playlistList.SelectedItem().(playlistListItem)
				if playlistItem.id == 0 {
					switch playlistItem.name {
					case "wave":
						m.playWave()
					}
				}
			}
		}

	case playerControl:
		switch msg {
		case _PLAYER_NEXT:
			m.nextTrack()
		}

	case progressControl:
		cmd = m.trackProgress.SetPercent(float64(msg))
		cmds = append(cmds, cmd)

	case progress.FrameMsg:
		progressModel, cmd := m.trackProgress.Update(msg)
		m.trackProgress = progressModel.(progress.Model)
		cmds = append(cmds, cmd)
	}

	switch m.page {
	case _PAGE_LOGIN:
		m.loginTextInput, cmd = m.loginTextInput.Update(message)
		cmds = append(cmds, cmd)
	case _PAGE_MAIN:
		m.playlistList, cmd = m.playlistList.Update(message)
		cmds = append(cmds, cmd)
		m.trackList, cmd = m.trackList.Update(message)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	switch m.page {
	case _PAGE_LOGIN:
		return m.renderLoginPage()
	case _PAGE_MAIN:
		return m.renderMainPage()
	}

	return ""
}

func (m *model) resize(w, h int) {
	m.width, m.height = w, h
	m.playlistList.SetSize(32, h)
	m.trackList.SetSize(m.width-m.playlistList.Width(), h-5)
	m.trackProgress.Width = m.width - m.playlistList.Width()
}

func (m *model) playWave() {
	if m.player != nil {
		if m.player.IsPlaying() {
			m.player.Pause()
			return
		} else {
			m.player.Play()
			return
		}
	}

	tracks, err := m.client.StationTracks(api.MyWaveId, nil)
	if err != nil {
		return
	}

	m.currentTrackIdx = 0
	m.playQueue = m.playQueue[:0]
	m.infiniteQueue = true

	for _, tr := range tracks.Sequence {
		m.playQueue = append(m.playQueue, tr.Track)
	}

	if len(m.playQueue) == 0 {
		return
	}

	m.playTrack(&m.playQueue[m.currentTrackIdx])
}

func (m *model) nextTrack() {
	m.player.Close()
	m.player = nil

	if m.currentTrackIdx+1 >= len(m.playQueue) {
		if m.infiniteQueue {
			tracks, err := m.client.StationTracks(api.MyWaveId, &m.playQueue[m.currentTrackIdx])
			if err != nil {
				return
			}

			for _, tr := range tracks.Sequence {
				m.playQueue = append(m.playQueue, tr.Track)
			}
		} else {
			m.currentTrackIdx = 0
			go programm.Send(_PLAYER_PAUSE)
			return
		}
	}

	m.currentTrackIdx++
	m.playTrack(&m.playQueue[m.currentTrackIdx])
}

func (m *model) playTrack(track *api.Track) {
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

	decoder, err := mp3.NewDecoder(trackReader)
	if err != nil {
		return
	}

	if m.trackWrapper.trackReader != nil {
		m.trackWrapper.trackReader.Close()
	}
	m.trackWrapper.trackReader = trackReader
	m.trackWrapper.decoder = decoder
	m.trackWrapper.trackDurationMs = track.DurationMs
	m.trackWrapper.trackStartTime = time.Now()

	m.player = m.playerContext.NewPlayer(m.trackWrapper)
	m.player.Play()
}

func (w *trackReaderWrapper) Read(dest []byte) (n int, err error) {
	n, err = w.decoder.Read(dest)
	if err == io.EOF {
		w.trackReader.Close()
		w.trackReader = nil
		go func() {
			programm.Send(1)
			programm.Send(_PLAYER_NEXT)
		}()
	} else if time.Since(w.lastUpdateTime) > time.Millisecond*33 {
		w.lastUpdateTime = time.Now()
		fraction := progressControl(float64(time.Since(w.trackStartTime).Milliseconds()) / float64(w.trackDurationMs))
		go programm.Send(fraction)
	}
	return
}
