package ui

import (
	"io"
	"time"
	"yamusic/api"
	"yamusic/config"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
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
	trackerHelp    help.Model

	playerContext *oto.Context
	player        *oto.Player
	trackWrapper  *trackReaderWrapper

	infinitePlaylist bool

	playQueue       []api.Track
	currentTrackIdx int
	playlistTracks  []api.Track
	currentPlaylist playlistListItem

	likedTracksMap   map[string]bool
	likedTracksSlice []string
}

type playerControl uint

const (
	_PLAYER_PLAY  playerControl = iota
	_PLAYER_PAUSE playerControl = iota
	_PLAYER_NEXT  playerControl = iota
	_PLAYER_PREV  playerControl = iota
)

type progressControl float64

type viewPlaylistControl uint64

const (
	_PLAYLIST_MYWAVE viewPlaylistControl = iota
	_PLAYLIST_LIKES  viewPlaylistControl = iota
	_PLAYLIST_PREDEF viewPlaylistControl = iota
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

		trackWrapper:   &trackReaderWrapper{},
		likedTracksMap: make(map[string]bool),
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

	m.playlistList.Title = "Playlists"
	m.playlistList.SetShowStatusBar(false)
	m.playlistList.Styles.Title = m.playlistList.Styles.Title.Foreground(accentColor).UnsetBackground().Padding(0)
	m.playlistList.KeyMap = list.KeyMap{
		CursorUp:   key.NewBinding(key.WithKeys("ctrl+up"), key.WithHelp("ctrl+↑", "up")),
		CursorDown: key.NewBinding(key.WithKeys("ctrl+down"), key.WithHelp("ctrl+↓", "down")),
	}

	m.trackList.Title = "Tracks"
	m.trackList.Styles.Title = m.trackList.Styles.Title.Foreground(normalTextColor).UnsetBackground().Padding(0)
	m.trackList.KeyMap = list.KeyMap{
		CursorUp:     key.NewBinding(key.WithKeys("up"), key.WithHelp("↑", "up")),
		CursorDown:   key.NewBinding(key.WithKeys("down"), key.WithHelp("↓", "down")),
		GoToEnd:      key.NewBinding(key.WithKeys(""), key.WithHelp("→", "next track")),
		Filter:       key.NewBinding(key.WithKeys(""), key.WithHelp("space", "play/pause")),
		Quit:         key.NewBinding(key.WithKeys(""), key.WithHelp("enter", "select")),
		ShowFullHelp: key.NewBinding(key.WithKeys(""), key.WithHelp("l", "like/unlike")),
	}

	m.trackProgress.ShowPercentage = false
	m.trackProgress.Empty = m.trackProgress.Full
	m.trackProgress.EmptyColor = "#6b6b6b"

	playlistListItems := []list.Item{
		playlistListItem{"my wave", uint64(_PLAYLIST_MYWAVE), true, false},
		playlistListItem{"likes", uint64(_PLAYLIST_LIKES), true, false},
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
				playlistListItems = append(playlistListItems, playlistListItem{playlist.Title, playlist.Kind, true, true})
			}
		}
		m.playlistList.SetItems(playlistListItems)

		tracks, err := m.client.StationTracks(api.MyWaveId, nil)
		if err == nil {
			var playlist []list.Item
			m.playlistTracks = m.playlistTracks[:0]
			for _, t := range tracks.Sequence {
				m.playlistTracks = append(m.playlistTracks, t.Track)
				playlist = append(playlist, trackListItem{
					title:      t.Track.Title,
					version:    t.Track.Version,
					artists:    artistList(t.Track.Artists),
					id:         t.Track.Id,
					liked:      false,
					durationMs: t.Track.DurationMs,
				})
			}
			m.trackList.SetItems(playlist)
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
				if !playlistItem.active {
					break
				}
				m.currentPlaylist = playlistItem
				m.infinitePlaylist = playlistItem.kind == uint64(_PLAYLIST_MYWAVE)
				m.playQueue = m.playlistTracks
				m.playCurrentQueue(m.trackList.Index())
			} else if keypress == " " {
				if m.player == nil {
					break
				}
				if m.player.IsPlaying() {
					m.player.Pause()
				} else {
					m.player.Play()
				}
			} else if keypress == "right" {
				m.nextTrack()
			} else if keypress == "l" {
				if len(m.playlistTracks) == 0 {
					break
				}
				index := m.trackList.Index()
				track := m.playlistTracks[index]
				if m.likedTracksMap[track.Id] {
					if m.client.UnlikeTrack(track.Id) != nil {
						break
					}
					delete(m.likedTracksMap, track.Id)
					for i, id := range m.likedTracksSlice {
						if id == track.Id {
							if i+1 < len(m.likedTracksSlice) {
								m.likedTracksSlice = append(m.likedTracksSlice[:i], m.likedTracksSlice[i+1:]...)
							} else {
								m.likedTracksSlice = m.likedTracksSlice[:i]
							}
							break
						}
					}
				} else {
					if m.client.LikeTrack(track.Id) != nil {
						break
					}
					m.likedTracksMap[track.Id] = true
					m.likedTracksSlice = append(m.likedTracksSlice, track.Id)
				}
				item := m.trackList.SelectedItem().(trackListItem)
				item.liked = m.likedTracksMap[track.Id]
				cmd = m.trackList.SetItem(index, item)
				cmds = append(cmds, cmd)
			}
		}

	// player control update
	case playerControl:
		switch msg {
		case _PLAYER_NEXT:
			m.nextTrack()
		}

	// track progress update
	case progressControl:
		cmd = m.trackProgress.SetPercent(float64(msg))
		cmds = append(cmds, cmd)

	// selected playlist
	case playlistListItem:
		var playlist []list.Item

		if len(m.playQueue) > 0 && msg.kind == m.currentPlaylist.kind {
			m.playlistTracks = m.playQueue
			m.trackList.Select(m.currentTrackIdx)
		} else if viewPlaylistControl(msg.kind) == _PLAYLIST_MYWAVE {
			tracks, err := m.client.StationTracks(api.MyWaveId, nil)
			if err != nil {
				break
			}
			m.playlistTracks = m.playlistTracks[:0]
			for _, t := range tracks.Sequence {
				m.playlistTracks = append(m.playlistTracks, t.Track)
			}
		} else if viewPlaylistControl(msg.kind) == _PLAYLIST_LIKES {
			tracks, err := m.client.Tracks(m.likedTracksSlice)
			if err != nil {
				break
			}
			m.playlistTracks = tracks
		} else {
			tracks, err := m.client.PlaylistTracks(msg.kind, false)
			if err != nil {
				break
			}
			m.playlistTracks = tracks
		}

		for _, t := range m.playlistTracks {
			playlist = append(playlist, trackListItem{
				title:      t.Title,
				version:    t.Version,
				artists:    artistList(t.Artists),
				id:         t.Id,
				liked:      m.likedTracksMap[t.Id],
				available:  t.Available,
				durationMs: t.DurationMs,
			})
		}

		m.trackList.Title = "Tracks from " + msg.name
		cmd = m.trackList.SetItems(playlist)
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
	m.playlistList.SetSize(32, h-5)
	m.trackList.SetSize(m.width-m.playlistList.Width()-20, h-12)
	m.trackProgress.Width = m.width - m.playlistList.Width() - 13
}

func (m *model) playCurrentQueue(trackIndex int) {
	if m.player != nil {
		if m.currentTrackIdx == trackIndex {
			if m.player.IsPlaying() {
				m.player.Pause()
				return
			} else {
				m.player.Play()
				return
			}
		} else {
			m.player.Close()
			m.player = nil
		}
	}

	if len(m.playQueue) == 0 {
		return
	}

	m.currentTrackIdx = trackIndex
	m.playTrack(&m.playQueue[m.currentTrackIdx])
}

func (m *model) nextTrack() {
	m.player.Close()
	m.player = nil

	if m.infinitePlaylist && m.currentTrackIdx+2 >= len(m.playQueue) {
		tracks, err := m.client.StationTracks(api.MyWaveId, &m.playQueue[m.currentTrackIdx])
		if err != nil {
			return
		}

		for _, tr := range tracks.Sequence {
			m.playQueue = append(m.playQueue, tr.Track)
		}
	} else if m.currentTrackIdx+1 >= len(m.playQueue) {
		m.currentTrackIdx = 0
		go programm.Send(_PLAYER_PAUSE)
		return
	}

	m.currentTrackIdx++
	m.playTrack(&m.playQueue[m.currentTrackIdx])

	selectedPlaylis := m.playlistList.SelectedItem().(playlistListItem)
	if m.currentPlaylist.kind == selectedPlaylis.kind {
		go programm.Send(selectedPlaylis)
	}
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
