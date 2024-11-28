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

		likedPlaylist, index := m.playlists.GetFirst(playlist.LIKES)
		trackIndex := likedPlaylist.RemoveTrack(track.Id)
		if m.playlists.SelectedItem().Kind == playlist.LIKES && trackIndex >= 0 {
			m.tracklist.RemoveItem(trackIndex)
			m.tracklist.Select(likedPlaylist.SelectedTrack)
		}

		return m.playlists.SetItem(index, likedPlaylist)
	} else {
		if m.client.LikeTrack(track.Id) != nil {
			return nil
		}

		m.likedTracksMap[track.Id] = true
		likedPlaylist, index := m.playlists.GetFirst(playlist.LIKES)
		likedPlaylist.AddTrack(track)

		return m.playlists.SetItem(index, likedPlaylist)
	}
}
