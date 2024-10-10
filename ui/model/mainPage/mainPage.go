package mainpage

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/dece2183/yamusic-tui/api"
	"github.com/dece2183/yamusic-tui/config"
	"github.com/dece2183/yamusic-tui/media"
	"github.com/dece2183/yamusic-tui/media/handler"
	"github.com/dece2183/yamusic-tui/ui/components/input"
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
	mediaHandler  handler.MediaHandler
	width, height int

	playlists    *playlist.Model
	tracklist    *tracklist.Model
	tracker      *tracker.Model
	searchDialog *search.Model
	inputDialog  *input.Model

	isSearchActive         bool
	isAddPlaylistActive    bool
	isRenamePlaylistActive bool

	currentPlaylistIndex int
	likedTracksMap       map[string]bool
}

// mainpage.Model constructor.
func New() *Model {
	m := &Model{}

	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion())
	m.program = p
	m.mediaHandler = media.NewHandler(config.ConfigPath, "Yandex music terminal client")
	m.likedTracksMap = make(map[string]bool)

	m.playlists = playlist.New(m.program, "YaMusic")
	m.tracklist = tracklist.New(m.program, &m.likedTracksMap)
	m.tracker = tracker.New(m.program, &m.likedTracksMap)
	m.searchDialog = search.New()
	m.inputDialog = input.New()

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

	m.mediaHandler.Enable()
	defer m.mediaHandler.Disable()
	go m.mediaHandle()

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
		case m.isSearchActive || m.isAddPlaylistActive:
			m.searchDialog, cmd = m.searchDialog.Update(message)
			cmds = append(cmds, cmd)
		case m.isRenamePlaylistActive:
			m.inputDialog, cmd = m.inputDialog.Update(message)
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
		case playlist.RENAME:
			selectedPlaylist := m.playlists.SelectedItem()
			if selectedPlaylist.Kind < playlist.USER {
				break
			}
			m.inputDialog.Title = "Rename playlist " + selectedPlaylist.Name
			m.inputDialog.SetValue(selectedPlaylist.Name)
			m.isRenamePlaylistActive = true
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
		case tracker.PLAY, tracker.PAUSE:
			m.mediaHandler.OnPlayPause()
		case tracker.STOP:
			m.mediaHandler.OnEnded()
		case tracker.REWIND:
			m.mediaHandler.OnSeek(m.tracker.Position())
		case tracker.VOLUME:
			m.mediaHandler.OnVolume()
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
		if m.isSearchActive || m.isAddPlaylistActive {
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
	if m.isSearchActive || m.isAddPlaylistActive {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, m.searchDialog.View())
	} else if m.isRenamePlaylistActive {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, m.inputDialog.View())
	}

	var sidePanel string
	if m.playlists.Width() > 0 {
		sidePanel = m.playlists.View()
	}
	midPanel := lipgloss.JoinVertical(lipgloss.Left, m.tracklist.View(), m.tracker.View())
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

	m.searchDialog.SetSize(searchWidth, height-4)
	m.inputDialog.SetWidth(searchWidth)
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
			playlistTracks, err := m.client.PlaylistTracks(pl.Kind, pl.Owner.Uid, false)
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
			albumArtists := make([]string, 0, len(track.Albums[0].Artists))
			md := handler.TrackMetadata{
				TrackId:      track.Id,
				Length:       time.Duration(track.DurationMs) * time.Millisecond,
				CoverUrl:     m.coverFilePath(track),
				AlbumName:    track.Albums[0].Title,
				AlbumArtists: albumArtists,
				Artists:      artists,
				Genre:        []string{track.Albums[0].Genre},
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
	tempDir := filepath.Join(os.TempDir(), config.ConfigPath)
	if os.MkdirAll(tempDir, 0755) != nil {
		return ""
	}
	return filepath.Join(tempDir, track.Id+".jpg")
}
