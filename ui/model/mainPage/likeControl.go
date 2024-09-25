package mainpage

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/dece2183/yamusic-tui/api"
	"github.com/dece2183/yamusic-tui/ui/components/playlist"
)

func (m *Model) likePlayingTrack() tea.Cmd {
	track := m.tracker.CurrentTrack()
	return m.likeTrack(track)
}

func (m *Model) likeSelectedTrack() tea.Cmd {
	if m.currentPlaylistIndex < 0 {
		return nil
	}

	currentPlaylist := m.playlists.Items()[m.currentPlaylistIndex]
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

		likedPlaylist := m.playlists.Items()[1]
		for i, ltrack := range likedPlaylist.Tracks {
			if ltrack.Id == track.Id {
				if i+1 < len(likedPlaylist.Tracks) {
					likedPlaylist.Tracks = append(likedPlaylist.Tracks[:i], likedPlaylist.Tracks[i+1:]...)
				} else {
					likedPlaylist.Tracks = likedPlaylist.Tracks[:i]
				}

				if m.playlists.SelectedItem().Kind == playlist.LIKES {
					m.tracklist.RemoveItem(i)
				}
				break
			}
		}

		return m.playlists.SetItem(1, likedPlaylist)
	} else {
		if m.client.LikeTrack(track.Id) != nil {
			return nil
		}

		m.likedTracksMap[track.Id] = true
		likedPlaylist := m.playlists.Items()[1]
		likedPlaylist.Tracks = append([]api.Track{*track}, likedPlaylist.Tracks...)
		return m.playlists.SetItem(1, likedPlaylist)
	}
}
