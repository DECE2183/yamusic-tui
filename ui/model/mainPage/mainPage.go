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

	m.playlist = playlist.New(m.program)
	m.tracklist = tracklist.New(m.program)
	m.tracker = tracker.New(m.program)

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
	case model.PlaylistControl:
		switch msg {
		case model.PLAYLIST_CURSOR_UP, model.PLAYLIST_CURSOR_DOWN:
			item := m.playlist.SelectedItem()

			if len(m.playQueue) > 0 && item.Kind == m.currentPlaylist.Kind {
				m.playlistTracks = m.playQueue
				m.tracklist.Select(m.currentTrackIdx)
			} else if item.Kind == playlist.MYWAVE {
				tracks, err := m.client.StationTracks(api.MyWaveId, nil)
				if err != nil {
					break
				}
				m.currentStationBatch = tracks.BatchId
				m.playlistTracks = m.playlistTracks[:0]
				for _, t := range tracks.Sequence {
					m.playlistTracks = append(m.playlistTracks, t.Track)
				}
			} else if item.Kind == playlist.LIKES {
				tracks, err := m.client.Tracks(m.likedTracksSlice)
				if err != nil {
					break
				}
				m.playlistTracks = tracks
			} else {
				tracks, err := m.client.PlaylistTracks(item.Kind, false)
				if err != nil {
					break
				}
				m.playlistTracks = tracks
			}

			tracks := make([]tracklist.Item, 0, len(m.playlistTracks))
			for _, t := range m.playlistTracks {
				tracks = append(tracks, tracklist.Item{
					Title:      t.Title,
					Version:    t.Version,
					Artists:    helpers.ArtistList(t.Artists),
					Id:         t.Id,
					Liked:      m.likedTracksMap[t.Id],
					Available:  t.Available,
					DurationMs: t.DurationMs,
				})
			}
			m.tracklist.SetItems(tracks)
		}

	// tracklist control update
	case model.TracklistControl:
		switch msg {
		case model.TRACKLIST_PLAY:
			playlistItem := m.playlist.SelectedItem()
			if !playlistItem.Active {
				break
			}
			if playlistItem.Kind == uint64(playlist.MYWAVE) {
				m.infinitePlaylist = true
				m.currentStationId = api.MyWaveId
			} else {
				m.infinitePlaylist = false
				m.currentStationId = api.StationId{}
			}
			m.playQueue = m.playlistTracks
			m.playCurrentQueue(m.tracklist.Index())
			m.currentPlaylist = playlistItem
		case model.TRACKLIST_CURSOR_UP, model.TRACKLIST_CURSOR_DOWN:
		case model.TRACKLIST_LIKE:
			cmd = m.likeSelectedTrack()
			cmds = append(cmds, cmd)
		case model.TRACKLIST_SHARE:
			if len(m.playlistTracks) == 0 {
				break
			}
			track := m.playlistTracks[m.tracklist.Index()]
			link := fmt.Sprintf("https://music.yandex.ru/album/%d/track/%s", track.Albums[0].Id, track.Id)
			clipboard.Write(clipboard.FmtText, []byte(link))
		}

	// player control update
	case model.PlayerControl:
		switch msg {
		case model.PLAYER_NEXT:
			m.nextTrack()
		case model.PLAYER_PREV:
			m.prevTrack()
		case model.PLAYER_LIKE:
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

	playlists, err := m.client.ListPlaylists()
	if err == nil {
		for _, pl := range playlists {
			m.playlist.InsertItem(-1, playlist.Item{Name: pl.Title, Kind: pl.Kind, Active: true, Subitem: true})
		}
	}

	likes, err := m.client.LikedTracks()
	if err == nil {
		m.likedTracksSlice = make([]string, 0, len(likes))
		for _, l := range likes {
			m.likedTracksMap[l.Id] = true
			m.likedTracksSlice = append(m.likedTracksSlice, l.Id)
		}
	}

	tracks, err := m.client.StationTracks(api.MyWaveId, nil)
	if err == nil {
		m.playlistTracks = m.playlistTracks[:0]
		for _, t := range tracks.Sequence {
			m.playlistTracks = append(m.playlistTracks, t.Track)
			m.tracklist.InsertItem(-1, tracklist.Item{
				Title:      t.Track.Title,
				Version:    t.Track.Version,
				Artists:    helpers.ArtistList(t.Track.Artists),
				Id:         t.Track.Id,
				Liked:      m.likedTracksMap[t.Track.Id],
				DurationMs: t.Track.DurationMs,
				Available:  t.Track.Available,
			})
		}
		m.currentStationBatch = tracks.BatchId
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
	if m.currentPlaylist.Kind == selectedPlaylist.Kind && m.tracklist.Index() == m.currentTrackIdx+1 {
		m.tracklist.Select(m.currentTrackIdx)
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
				// automatic append new tracks to the track list if this playlist is selected
				m.playQueue = append(m.playQueue, tr.Track)
				if m.playlist.SelectedItem().Kind == m.currentPlaylist.Kind {
					m.tracklist.InsertItem(-1, tracklist.Item{
						Title:      tr.Track.Title,
						Version:    tr.Track.Version,
						Artists:    helpers.ArtistList(tr.Track.Artists),
						Id:         tr.Track.Id,
						Liked:      m.likedTracksMap[tr.Track.Id],
						DurationMs: tr.Track.DurationMs,
						Available:  tr.Track.Available,
					})
				}
			}
		}
	} else if m.currentTrackIdx+1 >= len(m.playQueue) {
		m.currentTrackIdx = 0
		m.Send(model.PLAYER_STOP)
		return
	}

	m.currentTrackIdx++
	m.playTrack(&m.playQueue[m.currentTrackIdx])

	selectedPlaylist := m.playlist.SelectedItem()
	if m.currentPlaylist.Kind == selectedPlaylist.Kind && m.tracklist.Index() == m.currentTrackIdx-1 {
		m.tracklist.Select(m.currentTrackIdx)
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

	m.currentTrackIdx = trackIndex
	trackToPlay := &m.playQueue[m.currentTrackIdx]

	if m.infinitePlaylist {
		if m.tracker.IsPlaying() {
			currentTrack := m.tracker.CurrentTrack()
			go m.client.StationFeedback(
				api.ROTOR_SKIP,
				m.currentStationId,
				m.currentStationBatch,
				currentTrack.Id,
				int(float64(currentTrack.DurationMs*1000)*m.tracker.Progress()),
			)
			go m.client.StationFeedback(
				api.ROTOR_TRACK_STARTED,
				m.currentStationId,
				m.currentStationBatch,
				trackToPlay.Id,
				0,
			)
		} else {
			go m.client.StationFeedback(
				api.ROTOR_RADIO_STARTED,
				m.currentStationId,
				"",
				"",
				0,
			)
		}
	}

	m.playTrack(trackToPlay)
}

func (m *Model) likePlayingTrack() tea.Cmd {
	track := m.tracker.CurrentTrack()
	m.likeTrack(track)

	m.tracker.Liked = m.likedTracksMap[track.Id]

	selectedPlaylist := m.playlist.SelectedItem()
	if m.currentPlaylist.Kind == selectedPlaylist.Kind && m.currentPlaylist.Name == selectedPlaylist.Name {
		trackItem := m.tracklist.Items()[m.currentTrackIdx]
		trackItem.Liked = m.likedTracksMap[track.Id]
		return m.tracklist.SetItem(m.currentTrackIdx, trackItem)
	}

	return nil
}

func (m *Model) likeSelectedTrack() tea.Cmd {
	if len(m.playlistTracks) == 0 {
		return nil
	}

	index := m.tracklist.Index()
	track := m.playlistTracks[index]

	m.likeTrack(&track)

	selectedPlaylist := m.playlist.SelectedItem()
	if m.currentPlaylist.Kind == selectedPlaylist.Kind && m.currentPlaylist.Name == selectedPlaylist.Name && m.tracklist.Index() == m.currentTrackIdx {
		m.tracker.Liked = m.likedTracksMap[track.Id]
	}

	item := m.tracklist.SelectedItem()
	item.Liked = m.likedTracksMap[track.Id]
	return m.tracklist.SetItem(index, item)
}

func (m *Model) likeTrack(track *api.Track) {
	if m.likedTracksMap[track.Id] {
		if m.client.UnlikeTrack(track.Id) != nil {
			return
		}

		delete(m.likedTracksMap, track.Id)

		for i, id := range m.likedTracksSlice {
			if id == track.Id {
				if i+1 < len(m.likedTracksSlice) {
					m.likedTracksSlice = append(m.likedTracksSlice[:i], m.likedTracksSlice[i+1:]...)
				} else {
					m.likedTracksSlice = m.likedTracksSlice[:i]
				}

				if m.playlist.SelectedItem().Kind == playlist.LIKES {
					m.tracklist.RemoveItem(i)
				}
				break
			}
		}
	} else {
		if m.client.LikeTrack(track.Id) != nil {
			return
		}

		m.likedTracksMap[track.Id] = true
		m.likedTracksSlice = append([]string{track.Id}, m.likedTracksSlice...)

		if m.playlist.SelectedItem().Kind == playlist.LIKES {
			m.tracklist.InsertItem(0, tracklist.Item{
				Title:      track.Title,
				Version:    track.Version,
				Artists:    helpers.ArtistList(track.Artists),
				Id:         track.Id,
				Liked:      true,
				DurationMs: track.DurationMs,
				Available:  track.Available,
			})
		}
	}
}
