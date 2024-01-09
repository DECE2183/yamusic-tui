package ui

import (
	"fmt"
	"yamusic/api"
	"yamusic/config"
	"yamusic/ui/model"
	loginpage "yamusic/ui/model/loginPage"
	mainpage "yamusic/ui/model/mainPage"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"golang.design/x/clipboard"
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
		switch m.page {
		case _PAGE_MAIN:
			if controls.TrackListSelect.Contains(keypress) {
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
			} else if controls.TrackListLike.Contains(keypress) || controls.PlayerLike.Contains(keypress) {
				var track api.Track
				if controls.TrackListLike.Contains(keypress) {
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

				if controls.TrackListLike.Contains(keypress) {
					index := m.trackList.Index()

					item := m.trackList.SelectedItem().(trackListItem)
					item.liked = m.likedTracksMap[track.Id]

					cmd = m.trackList.SetItem(index, item)
					cmds = append(cmds, cmd)
				}
			} else if controls.TrackListShare.Contains(keypress) {
				if len(m.playlistTracks) == 0 {
					break
				}
				track := m.playlistTracks[m.trackList.Index()]
				link := fmt.Sprintf("https://music.yandex.ru/album/%d/track/%s", track.Albums[0].Id, track.Id)
				clipboard.Write(clipboard.FmtText, []byte(link))
			}
		}

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
