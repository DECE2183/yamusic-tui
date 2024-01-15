package tracker

import (
	"fmt"
	"io"
	"math"
	"time"

	"github.com/dece2183/yamusic-tui/api"
	"github.com/dece2183/yamusic-tui/config"
	"github.com/dece2183/yamusic-tui/ui/helpers"
	"github.com/dece2183/yamusic-tui/ui/model"
	"github.com/dece2183/yamusic-tui/ui/style"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	mp3 "github.com/dece2183/go-stream-mp3"
	"github.com/ebitengine/oto/v3"
)

type PlayerControl uint

const (
	PLAY PlayerControl = iota
	PAUSE
	STOP
	NEXT
	PREV
	LIKE
)

type ProgressControl float64

func (p ProgressControl) Value() float64 {
	return float64(p)
}

type trackerHelpKeyMap struct {
	PlayPause  key.Binding
	PrevTrack  key.Binding
	NextTrack  key.Binding
	LikeUnlike key.Binding
	Forward    key.Binding
	Backward   key.Binding
}

var trackerHelpMap = trackerHelpKeyMap{
	PlayPause: key.NewBinding(
		config.Current.Controls.PlayerPause.Binding(),
		config.Current.Controls.PlayerPause.Help("play/pause"),
	),
	PrevTrack: key.NewBinding(
		config.Current.Controls.PlayerPrevious.Binding(),
		config.Current.Controls.PlayerPrevious.Help("previous track"),
	),
	NextTrack: key.NewBinding(
		config.Current.Controls.PlayerNext.Binding(),
		config.Current.Controls.PlayerNext.Help("next track"),
	),
	LikeUnlike: key.NewBinding(
		config.Current.Controls.PlayerLike.Binding(),
		config.Current.Controls.PlayerLike.Help("like/unlike"),
	),
	Backward: key.NewBinding(
		config.Current.Controls.PlayerRewindBackward.Binding(),
		config.Current.Controls.PlayerRewindBackward.Help(fmt.Sprintf("-%d sec", int(config.Current.RewindDuration))),
	),
	Forward: key.NewBinding(
		config.Current.Controls.PlayerRewindForward.Binding(),
		config.Current.Controls.PlayerRewindForward.Help(fmt.Sprintf("+%d sec", int(config.Current.RewindDuration))),
	),
}

func (k trackerHelpKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.PlayPause, k.NextTrack, k.PrevTrack, k.Forward, k.Backward, k.LikeUnlike}
}

func (k trackerHelpKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.PlayPause, k.NextTrack, k.PrevTrack, k.Forward, k.Backward, k.LikeUnlike},
	}
}

var rewindAmount = time.Duration(config.Current.RewindDuration) * time.Second

type Model struct {
	width    int
	track    *api.Track
	progress progress.Model
	help     help.Model

	volume        float64
	playerContext *oto.Context
	player        *oto.Player
	trackWrapper  *readWrapper

	program  *tea.Program
	likesMap *map[string]bool
}

func New(p *tea.Program, likesMap *map[string]bool) Model {
	m := Model{
		program:  p,
		likesMap: likesMap,
		progress: progress.New(progress.WithSolidFill(string(style.AccentColor))),
		help:     help.New(),
		track:    &api.Track{},
		volume:   config.Current.Volume,
	}

	m.progress.ShowPercentage = false
	m.progress.Empty = m.progress.Full
	m.progress.EmptyColor = string(style.BackgroundColor)

	m.trackWrapper = &readWrapper{program: m.program}

	op := &oto.NewContextOptions{
		SampleRate:   44100,
		ChannelCount: 2,
		BufferSize:   time.Millisecond * time.Duration(config.Current.BufferSize),
		Format:       oto.FormatSignedInt16LE,
	}

	var err error
	var readyChan chan struct{}
	m.playerContext, readyChan, err = oto.NewContext(op)
	if err != nil {
		model.PrettyExit(err, 12)
	}
	<-readyChan

	return m
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) View() string {
	var playButton string
	if m.IsPlaying() {
		playButton = style.ActiveButtonStyle.Padding(0, 1).Margin(0).Render(style.IconStop)
	} else {
		playButton = style.ActiveButtonStyle.Padding(0, 1).Margin(0).Render(style.IconPlay)
	}

	var trackTitle string
	if m.track.Available {
		trackTitle = style.TrackTitleStyle.Render(m.track.Title)
	} else {
		trackTitle = style.TrackTitleStyle.Copy().Strikethrough(true).Render(m.track.Title)
	}

	trackVersion := style.TrackVersionStyle.Render(" " + m.track.Version)
	trackArtist := style.TrackArtistStyle.Render(helpers.ArtistList(m.track.Artists))

	durTotal := time.Millisecond * time.Duration(m.track.DurationMs)
	durEllapsed := time.Millisecond * time.Duration(float64(m.track.DurationMs)*m.progress.Percent())
	trackTime := style.TrackVersionStyle.Render(fmt.Sprintf("%02d:%02d/%02d:%02d",
		int(durEllapsed.Minutes()),
		int(durEllapsed.Seconds())%60,
		int(durTotal.Minutes()),
		int(durTotal.Seconds())%60,
	))

	var trackLike string
	if (*m.likesMap)[m.track.Id] {
		trackLike = style.IconLiked + " "
	} else {
		trackLike = style.IconNotLiked + " "
	}

	trackAddInfo := style.TrackAddInfoStyle.Render(trackLike + trackTime)

	trackTitle = lipgloss.JoinHorizontal(lipgloss.Top, trackTitle, trackVersion)
	trackTitle = lipgloss.JoinVertical(lipgloss.Left, trackTitle, trackArtist, "")
	trackTitle = lipgloss.NewStyle().Width(m.width - 36).Render(trackTitle)
	trackTitle = lipgloss.JoinHorizontal(lipgloss.Top, trackTitle, trackAddInfo)

	tracker := style.TrackProgressStyle.Render(m.progress.View())
	tracker = lipgloss.JoinHorizontal(lipgloss.Top, playButton, tracker)
	tracker = lipgloss.JoinVertical(lipgloss.Left, tracker, trackTitle, m.help.View(trackerHelpMap))

	return style.TrackBoxStyle.Width(m.width - 4).Render(tracker)
}

func (m Model) Update(message tea.Msg) (Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := message.(type) {
	case tea.KeyMsg:
		controls := config.Current.Controls
		keypress := msg.String()

		switch {
		case controls.PlayerPause.Contains(keypress):
			if m.player == nil {
				break
			}
			if m.player.IsPlaying() {
				m.Pause()
			} else {
				m.Play()
			}

		case controls.PlayerRewindForward.Contains(keypress):
			m.rewind(rewindAmount)

		case controls.PlayerRewindBackward.Contains(keypress):
			m.rewind(-rewindAmount)

		case controls.PlayerNext.Contains(keypress):
			cmds = append(cmds, model.Cmd(NEXT))

		case controls.PlayerPrevious.Contains(keypress):
			cmds = append(cmds, model.Cmd(PREV))

		case controls.PlayerLike.Contains(keypress):
			cmds = append(cmds, model.Cmd(LIKE))
		}

	// player control update
	case PlayerControl:
		switch msg {
		case PLAY:
			m.Play()
		case PAUSE:
			m.Pause()
		case STOP:
			m.Stop()
		}

	// track progress update
	case ProgressControl:
		cmd = m.progress.SetPercent(msg.Value())
		cmds = append(cmds, cmd)

	case progress.FrameMsg:
		progressModel, cmd := m.progress.Update(msg)
		m.progress = progressModel.(progress.Model)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m *Model) SetWidth(width int) {
	m.width = width
	m.progress.Width = width - 14
	m.help.Width = width - 8
}

func (m *Model) Width() int {
	return m.width
}

func (m *Model) SetProgress(p float64) tea.Cmd {
	return m.progress.SetPercent(p)
}

func (m *Model) Progress() float64 {
	return m.progress.Percent()
}

func (m *Model) SetVolume(v float64) {
	m.volume = v
	if m.player != nil {
		m.player.SetVolume(v)
	}
}

func (m *Model) Volume() float64 {
	return m.volume
}

func (m *Model) StartTrack(track *api.Track, reader *api.HttpReadSeeker) {
	if m.player != nil {
		m.Stop()
	}

	m.track = track
	decoder, err := mp3.NewDecoder(reader)
	if err != nil {
		return
	}

	m.trackWrapper.trackReader = reader
	m.trackWrapper.decoder = decoder
	m.trackWrapper.trackDurationMs = track.DurationMs
	m.trackWrapper.trackStartTime = time.Now()

	m.player = m.playerContext.NewPlayer(m.trackWrapper)
	m.player.SetVolume(m.volume)
	m.player.Play()
}

func (m *Model) Stop() {
	if m.player == nil {
		return
	}

	if m.player.IsPlaying() {
		m.player.Pause()
	}

	m.player.Close()
	m.player = nil

	if m.trackWrapper.trackReader != nil {
		m.trackWrapper.trackReader.Close()
		m.trackWrapper.trackReader = nil
	}
}

func (m *Model) IsPlaying() bool {
	return m.player != nil && m.trackWrapper.trackReader != nil && m.player.IsPlaying()
}

func (m *Model) CurrentTrack() *api.Track {
	return m.track
}

func (m *Model) Play() {
	if m.player == nil || m.trackWrapper.trackReader == nil {
		return
	}
	if m.player.IsPlaying() {
		return
	}
	m.player.Play()
}

func (m *Model) Pause() {
	if m.player == nil || m.trackWrapper.trackReader == nil {
		return
	}
	if !m.player.IsPlaying() {
		return
	}
	m.player.Pause()
}

func (m *Model) rewind(amount time.Duration) {
	if m.player == nil || m.trackWrapper == nil {
		go m.program.Send(STOP)
		return
	}

	amountMs := amount.Milliseconds()
	currentPos := int64(float64(m.trackWrapper.trackReader.Length()) * m.trackWrapper.trackReader.Progress())
	byteOffset := int64(math.Round((float64(m.trackWrapper.trackReader.Length()) / float64(m.trackWrapper.trackDurationMs)) * float64(amountMs)))

	// align position by 4 bytes
	currentPos += byteOffset
	currentPos -= currentPos % 4

	if currentPos <= 0 {
		m.player.Seek(0, io.SeekStart)
	} else if currentPos >= m.trackWrapper.trackReader.Length() {
		m.player.Seek(0, io.SeekEnd)
	} else {
		m.player.Seek(currentPos, io.SeekStart)
	}
}
