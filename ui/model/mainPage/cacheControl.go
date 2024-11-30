package mainpage

import (
	"os"

	"github.com/bogem/id3v2"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/dece2183/yamusic-tui/api"
	"github.com/dece2183/yamusic-tui/cache"
	"github.com/dece2183/yamusic-tui/ui/components/playlist"
)

func (m *Model) cacheCurrentTrack() tea.Cmd {
	currentTrack := m.tracker.CurrentTrack()
	if m.tracker.IsStoped() || m.cachedTracksMap[currentTrack.Id] {
		return nil
	}

	metadataFile, err := os.OpenFile(m.metadataFilePath(), os.O_RDONLY, 0755)
	if err != nil {
		return nil
	}

	defer metadataFile.Close()

	cacheFile, err := cache.Write(currentTrack.Id)
	if err != nil {
		return nil
	}

	defer cacheFile.Close()

	tag := id3v2.NewEmptyTag()
	tag.Reset(metadataFile, id3v2.Options{Parse: true})
	tag.WriteTo(cacheFile)
	m.tracker.TrackBuffer().WriteTo(cacheFile)

	m.cachedTracksMap[currentTrack.Id] = true
	cachePlaylist, index := m.playlists.GetFirst(playlist.LOCAL)
	cachePlaylist.AddTrack(currentTrack)
	return m.playlists.SetItem(index, cachePlaylist)
}

func (m *Model) removeCache(track *api.Track) tea.Cmd {
	err := cache.Remove(track.Id)
	if err != nil {
		return nil
	}

	cachePlaylist, index := m.playlists.GetFirst(playlist.LOCAL)
	trackIndex := cachePlaylist.RemoveTrack(track.Id)
	if m.playlists.SelectedItem().Kind == playlist.LOCAL && trackIndex >= 0 {
		m.tracklist.RemoveItem(trackIndex)
		m.tracklist.Select(cachePlaylist.SelectedTrack)
	}

	delete(m.cachedTracksMap, track.Id)
	return m.playlists.SetItem(index, cachePlaylist)
}
