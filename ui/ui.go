package ui

import (
	"fmt"
	"io"
	"math"
	"time"
	"yamusic/api"
	"yamusic/config"
	"yamusic/ui/model"
	loginpage "yamusic/ui/model/loginPage"
	mainpage "yamusic/ui/model/mainPage"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	mp3 "github.com/dece2183/go-stream-mp3"
	"golang.design/x/clipboard"
)

var (
	rewindAmount = time.Duration(config.Current.RewindDuration) * time.Second
)

var (
	programm *tea.Program
)

func Run() {
	var err error

	if config.Current.Token == "" {
		err = loginpage.New().Run()
		if err != nil {
			model.PrettyExit(err, 4)
		}
	}

	err = clipboard.Init()
	if err != nil {
		model.PrettyExit(err, 6)
	}

	err = mainpage.New().Run()
}

func (m model) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	controls := config.Current.Controls

	switch msg := message.(type) {
	case tea.WindowSizeMsg:
		m.resize(msg.Width, msg.Height)
		return m, tea.ClearScreen

	case tea.KeyMsg:
		keypress := msg.String()
		if keypress == "esc" || keypress == "ctrl+q" || keypress == "ctrl+c" {
			m.page = _PAGE_QUIT
			return m, tea.Quit
		}

		switch m.page {
		case _PAGE_LOGIN:
			if keypress == "enter" {
				config.Current.Token = m.loginTextInput.Value()
				err := config.Save()
				if err != nil {
					return m, nil
				}
				m.page = _PAGE_MAIN
				m.initialLoad()
				return m, nil
			}
		case _PAGE_MAIN:
			if keypress == controls.TrackListSelect.Key() {
				playlistItem := m.playlistList.SelectedItem().(playlistListItem)
				if !playlistItem.active {
					break
				}
				if playlistItem.kind == uint64(_PLAYLIST_MYWAVE) {
					m.infinitePlaylist = true
					m.currentStationId = api.MyWaveId
				} else {
					m.infinitePlaylist = false
					m.currentStationId = api.StationId{}
				}
				m.playQueue = m.playlistTracks
				m.playCurrentQueue(m.trackList.Index())
				m.currentPlaylist = playlistItem
			} else if keypress == controls.PlayerPause.Key() {
				if m.player == nil {
					break
				}
				if m.player.IsPlaying() {
					m.player.Pause()
				} else {
					m.player.Play()
				}
			} else if keypress == controls.PlayerRewindBackward.Key() {
				m.rewind(-rewindAmount)
			} else if keypress == controls.PlayerRewindForward.Key() {
				m.rewind(rewindAmount)
			} else if keypress == controls.PlayerPrevious.Key() {
				m.prevTrack()
			} else if keypress == controls.PlayerNext.Key() {
				if len(m.playQueue) > 0 {
					currTrack := m.playQueue[m.currentTrackIdx]

					if len(m.currentStationId.Tag) > 0 && len(m.currentStationId.Type) > 0 {
						go m.client.StationFeedback(
							api.ROTOR_SKIP,
							m.currentStationId,
							m.currentStationBatch,
							currTrack.Id,
							int(m.trackWrapper.trackReader.Progress()*float64(currTrack.DurationMs))*1000,
						)
					}
				}

				m.nextTrack()
			} else if keypress == controls.TrackListLike.Key() || keypress == controls.PlayerLike.Key() {
				var track api.Track
				if keypress == controls.TrackListLike.Key() {
					if len(m.playlistTracks) == 0 {
						break
					}

					index := m.trackList.Index()
					track = m.playlistTracks[index]
				} else {
					if len(m.playQueue) == 0 {
						break
					}

					track = m.playQueue[m.currentTrackIdx]
				}

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

				if keypress == controls.TrackListLike.Key() {
					index := m.trackList.Index()

					item := m.trackList.SelectedItem().(trackListItem)
					item.liked = m.likedTracksMap[track.Id]

					cmd = m.trackList.SetItem(index, item)
					cmds = append(cmds, cmd)
				}
			} else if keypress == controls.TrackListShare.Key() {
				if len(m.playlistTracks) == 0 {
					break
				}
				track := m.playlistTracks[m.trackList.Index()]
				link := fmt.Sprintf("https://music.yandex.ru/album/%d/track/%s", track.Albums[0].Id, track.Id)
				clipboard.Write(clipboard.FmtText, []byte(link))
			}
		}

	// player control update
	case playerControl:
		switch msg {
		case _PLAYER_PREV:
			m.prevTrack()
		case _PLAYER_NEXT:
			if len(m.playQueue) > 0 {
				currTrack := m.playQueue[m.currentTrackIdx]

				if len(m.currentStationId.Tag) > 0 && len(m.currentStationId.Type) > 0 {
					go m.client.StationFeedback(
						api.ROTOR_TRACK_FINISHED,
						m.currentStationId,
						m.currentStationBatch,
						currTrack.Id,
						currTrack.DurationMs*1000,
					)
				}
			}
			m.nextTrack()
		case _PLAYER_PAUSE:
			m.pauseTrack()
		case _PLAYER_STOP:
			m.stopTrack()
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
			m.currentStationBatch = tracks.BatchId
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

func (m *model) rewind(amount time.Duration) {
	if m.player == nil || m.trackWrapper == nil {
		go programm.Send(_PLAYER_STOP)
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

func (m *model) playCurrentQueue(trackIndex int) {
	if m.player != nil {
		selectedPlaylis := m.playlistList.SelectedItem().(playlistListItem)
		if m.currentPlaylist.kind == selectedPlaylis.kind && m.currentTrackIdx == trackIndex {
			if m.player.IsPlaying() {
				m.player.Pause()
				return
			} else {
				m.player.Play()
				return
			}
		}
	}

	if len(m.playQueue) == 0 {
		m.stopTrack()
		return
	}

	if len(m.currentStationId.Tag) > 0 && len(m.currentStationId.Type) > 0 {
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

func (m *model) prevTrack() {
	if m.currentTrackIdx == 0 {
		go programm.Send(_PLAYER_STOP)
		return
	}

	m.currentTrackIdx--
	m.playTrack(&m.playQueue[m.currentTrackIdx])

	selectedPlaylis := m.playlistList.SelectedItem().(playlistListItem)
	if m.currentPlaylist.kind == selectedPlaylis.kind {
		go programm.Send(selectedPlaylis)
	}
}

func (m *model) nextTrack() {
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
		go programm.Send(_PLAYER_STOP)
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
	m.stopTrack()

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

	if len(m.currentStationId.Tag) > 0 && len(m.currentStationId.Type) > 0 {
		go m.client.StationFeedback(
			api.ROTOR_TRACK_STARTED,
			m.currentStationId,
			m.currentStationBatch,
			track.Id,
			0,
		)
	}

	m.trackWrapper.trackReader = trackReader
	m.trackWrapper.decoder = decoder
	m.trackWrapper.trackDurationMs = track.DurationMs
	m.trackWrapper.trackStartTime = time.Now()

	m.player = m.playerContext.NewPlayer(m.trackWrapper)
	m.player.SetVolume(config.Current.Volume)
	m.player.Play()

	go m.client.PlayTrack(track, false)
}

func (m model) pauseTrack() {
	if m.player == nil {
		return
	}
	m.player.Pause()
}

func (m *model) stopTrack() {
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
	}
}
