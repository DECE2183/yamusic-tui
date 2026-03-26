package mainpage

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/dece2183/yamusic-tui/api"
	"github.com/dece2183/yamusic-tui/log"
	"github.com/dece2183/yamusic-tui/ui/components/playlist"
)

func (m *Model) likePlayingTrack() tea.Cmd {
	var currentPlaylist *playlist.Item
	if m.currentPlaylistIndex >= 0 {
		currentPlaylist = m.playlists.Items()[m.currentPlaylistIndex]
	}

	track := m.tracker.CurrentTrack()
	return m.likeTrack(track, currentPlaylist)
}

func (m *Model) likeSelectedTrack() tea.Cmd {
	if m.currentPlaylistIndex < 0 {
		return nil
	}

	selectedPlaylist := m.playlists.SelectedItem()
	if len(selectedPlaylist.Tracks) == 0 {
		return nil
	}

	track := m.tracklist.SelectedItem().Track
	return m.likeTrack(track, selectedPlaylist)
}

func (m *Model) likeTrack(track *api.Track, pl *playlist.Item) tea.Cmd {
	likedPlaylist, index := m.playlists.GetFirst(playlist.LIKES)

	var evType api.TrackEventType
	if m.likedTracksMap[track.Id] {
		if m.client.UnlikeTrack(track.Id) != nil {
			return nil
		}
		delete(m.likedTracksMap, track.Id)
		likedPlaylist.RemoveTrack(track.Id)
		evType = api.EV_TRACK_UNLIKED
	} else {
		if m.client.LikeTrack(track.Id) != nil {
			return nil
		}
		m.likedTracksMap[track.Id] = true
		likedPlaylist.AddTrack(track)
		evType = api.EV_TRACK_LIKED
	}

	if pl != nil && pl.Rotor {
		ev := api.NewTrackFeedbackEvent(evType, track, 0)
		go m.client.RotorSessionFeedback(pl.SessionId, api.NewFeedback(pl.SessionBatch, ev))
		log.Print(log.LVL_INFO, "feedback event sended: "+ev.Type+" track: "+track.Title)
	}

	cmd := m.playlists.SetItem(index, likedPlaylist)
	if m.playlists.SelectedItem().Kind == playlist.LIKES {
		m.displayPlaylist(likedPlaylist)
	}

	m.indicateCurrentTrackPlaying(m.tracker.IsPlaying())
	return cmd
}
