package mainpage

import (
	"os"

	"github.com/bogem/id3v2/v2"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/dece2183/yamusic-tui/api"
	"github.com/dece2183/yamusic-tui/cache"
	"github.com/dece2183/yamusic-tui/log"
	"github.com/dece2183/yamusic-tui/ui/components/playlist"
)

func (m *Model) cacheCurrentTrack() tea.Cmd {
	currentTrack := m.tracker.CurrentTrack()
	if m.tracker.IsStoped() || m.cachedTracksMap[currentTrack.Id] {
		return nil
	}

	metadataFile, err := os.OpenFile(m.metadataFilePath(), os.O_RDONLY, 0755)
	if err != nil {
		log.Print(log.LVL_ERROR, "failed to open cache file: %s", err)
		m.tracker.ShowError("cache open")
		return nil
	}

	defer metadataFile.Close()

	cacheFile, err := cache.Write(currentTrack.Id)
	if err != nil {
		log.Print(log.LVL_ERROR, "failed to write cache file: %s", err)
		m.tracker.ShowError("cache write")
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
	cmd := m.playlists.SetItem(index, cachePlaylist)

	if m.playlists.SelectedItem().Kind == playlist.LOCAL {
		m.displayPlaylist(cachePlaylist)
	}

	m.indicateCurrentTrackPlaying(m.tracker.IsPlaying())
	return cmd
}

func (m *Model) removeCache(track *api.Track) tea.Cmd {
	if m.tracker.CurrentTrack().Id == track.Id && len(m.tracker.CurrentTrack().RealId) == 0 {
		m.tracker.ShowError("can't remove currently playing track")
		return nil
	}

	err := cache.Remove(track.Id)
	if err != nil {
		log.Print(log.LVL_ERROR, "failed to remove cached file: %s", err)
		m.tracker.ShowError("cache remove")
		return nil
	}

	cachePlaylist, index := m.playlists.GetFirst(playlist.LOCAL)
	cachePlaylist.RemoveTrack(track.Id)

	delete(m.cachedTracksMap, track.Id)
	cmd := m.playlists.SetItem(index, cachePlaylist)

	if m.playlists.SelectedItem().Kind == playlist.LOCAL {
		m.displayPlaylist(cachePlaylist)
	}

	m.indicateCurrentTrackPlaying(m.tracker.IsPlaying())
	return cmd
}
