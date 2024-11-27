package mainpage

import (
	"fmt"
	"io"
	"os"

	"github.com/bogem/id3v2"
	"github.com/dece2183/yamusic-tui/api"
	"github.com/dece2183/yamusic-tui/cache"
	"github.com/dece2183/yamusic-tui/config"
	"github.com/dece2183/yamusic-tui/stream"
	"github.com/dece2183/yamusic-tui/ui/components/tracker"
	"github.com/dece2183/yamusic-tui/ui/components/tracklist"
	"github.com/dece2183/yamusic-tui/ui/helpers"
)

func (m *Model) prevTrack() {
	if m.currentPlaylistIndex < 0 {
		return
	}

	currentPlaylist := m.playlists.Items()[m.currentPlaylistIndex]
	if len(currentPlaylist.Tracks) == 0 || currentPlaylist.CurrentTrack == 0 {
		m.Send(tracker.STOP)
		return
	}

	m.indicateCurrentTrackPlaying(false)

	currentPlaylist.CurrentTrack--
	m.playlists.SetItem(m.currentPlaylistIndex, currentPlaylist)
	m.playTrack(&currentPlaylist.Tracks[currentPlaylist.CurrentTrack])

	selectedPlaylist := m.playlists.SelectedItem()
	if currentPlaylist.IsSame(selectedPlaylist) && m.tracklist.Index() == currentPlaylist.CurrentTrack+1 {
		m.tracklist.Select(currentPlaylist.CurrentTrack)
	}
}

func (m *Model) nextTrack() {
	if m.currentPlaylistIndex < 0 {
		return
	}

	currentPlaylist := m.playlists.Items()[m.currentPlaylistIndex]
	if len(currentPlaylist.Tracks) == 0 {
		m.Send(tracker.STOP)
		return
	}

	m.indicateCurrentTrackPlaying(false)

	if currentPlaylist.Infinite {
		currTrack := currentPlaylist.Tracks[currentPlaylist.CurrentTrack]

		if m.tracker.Progress() == 1 {
			go m.client.StationFeedback(
				api.ROTOR_TRACK_FINISHED,
				currentPlaylist.StationId,
				currentPlaylist.StationBatch,
				currTrack.Id,
				currTrack.DurationMs*1000,
			)
		} else {
			go m.client.StationFeedback(
				api.ROTOR_SKIP,
				currentPlaylist.StationId,
				currentPlaylist.StationBatch,
				currTrack.Id,
				int(float64(currTrack.DurationMs*1000)*m.tracker.Progress()),
			)
		}

		if currentPlaylist.CurrentTrack+2 >= len(currentPlaylist.Tracks) {
			tracks, err := m.client.StationTracks(api.MyWaveId, &currTrack)
			if err != nil {
				return
			}

			for _, tr := range tracks.Sequence {
				// automatic append new tracks to the track list if this playlist is selected
				currentPlaylist.Tracks = append(currentPlaylist.Tracks, tr.Track)
				if m.playlists.SelectedItem().IsSame(currentPlaylist) {
					newTrack := &currentPlaylist.Tracks[len(currentPlaylist.Tracks)-1]
					m.tracklist.InsertItem(-1, tracklist.NewItem(newTrack))
				}
			}
		}
	} else if currentPlaylist.CurrentTrack+1 >= len(currentPlaylist.Tracks) {
		currentPlaylist.CurrentTrack = 0
		m.playlists.SetItem(m.currentPlaylistIndex, currentPlaylist)
		m.Send(tracker.STOP)
		return
	}

	currentPlaylist.CurrentTrack++
	m.playlists.SetItem(m.currentPlaylistIndex, currentPlaylist)
	m.playTrack(&currentPlaylist.Tracks[currentPlaylist.CurrentTrack])

	selectedPlaylist := m.playlists.SelectedItem()
	if currentPlaylist.IsSame(selectedPlaylist) && m.tracklist.Index() == currentPlaylist.CurrentTrack-1 {
		m.tracklist.Select(currentPlaylist.CurrentTrack)
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

	var (
		coverFile *os.File
		coverStat os.FileInfo
	)

	var coverBytes []byte
	coverPath := m.coverFilePath(track)
	coverFile, err = os.OpenFile(coverPath, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0755)
	if err != nil {
		goto skipcover
	}

	defer coverFile.Close()

	coverStat, err = coverFile.Stat()
	if err != nil || coverStat.Size() == 0 {
		err = api.DownloadTrackCover(coverFile, track, 200)
		if err != nil {
			goto skipcover
		}
		coverFile.Sync()
		stat, _ := coverFile.Stat()
		coverBytes = make([]byte, stat.Size())
		coverFile.Seek(0, io.SeekStart)
		coverFile.Read(coverBytes)
	}

skipcover:
	tag := id3v2.NewEmptyTag()
	tag.SetTitle(track.Title)
	tag.SetAlbum(track.Albums[0].Title)
	tag.SetGenre(track.Albums[0].Genre)
	tag.SetArtist(helpers.ArtistList(track.Artists))
	tag.SetYear(fmt.Sprint(track.Albums[0].Year))
	tag.AddAttachedPicture(id3v2.PictureFrame{
		MimeType:    "image/jpeg",
		PictureType: id3v2.PTFrontCover,
		Picture:     coverBytes,
	})

	var cacheFile *os.File
	var cacheWriters = make([]io.WriteCloser, 0, 1)

	cacheFile, err = os.OpenFile(m.currentTrackFilePath(), os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0755)
	if err == nil {
		cacheWriters = append(cacheWriters, cacheFile)
		tag.WriteTo(cacheFile)
		cacheFile.Sync()
	}

	var trackReader io.ReadCloser
	var trackSize int64

	trackReader, trackSize, err = cache.Read(track.Id)
	if err != nil {
		trackReader, trackSize, err = m.client.DownloadTrack(bestTrackInfo)
		if err != nil {
			return
		}
		cacheMode := config.Current.CacheTracks
		if cacheMode == config.CACHE_ALL || (cacheMode == config.CACHE_LIKED_ONLY && m.likedTracksMap[track.Id]) {
			cacheFile, err = cache.Write(track.Id)
			if err == nil {
				cacheWriters = append(cacheWriters, cacheFile)
				tag.WriteTo(cacheFile)
				cacheFile.Sync()
			}
		}
	}

	tag.Close()

	m.indicateCurrentTrackPlaying(true)
	m.tracker.StartTrack(track, stream.NewBufferedStream(trackReader, trackSize, cacheWriters...))

	if m.currentPlaylistIndex >= 0 {
		currentPlaylist := m.playlists.Items()[m.currentPlaylistIndex]
		if currentPlaylist.Infinite {
			go m.client.StationFeedback(
				api.ROTOR_TRACK_STARTED,
				currentPlaylist.StationId,
				currentPlaylist.StationBatch,
				track.Id,
				0,
			)
		}
	}

	m.mediaHandler.OnPlayback()
	go m.client.PlayTrack(track, false)
}

func (m *Model) playSelectedPlaylist(trackIndex int) {
	selectedPlaylist := m.playlists.SelectedItem()
	if len(selectedPlaylist.Tracks) == 0 {
		m.Send(tracker.STOP)
		return
	}

	trackToPlay := &selectedPlaylist.Tracks[selectedPlaylist.SelectedTrack]

	if m.currentPlaylistIndex >= 0 {
		currentPlaylist := m.playlists.Items()[m.currentPlaylistIndex]
		if currentPlaylist.IsSame(selectedPlaylist) && m.tracker.CurrentTrack().Id == trackToPlay.Id {
			if m.tracker.IsPlaying() {
				m.tracker.Pause()
				return
			} else {
				m.tracker.Play()
				return
			}
		}
	}

	m.indicateCurrentTrackPlaying(false)
	selectedPlaylist.CurrentTrack = trackIndex

	if selectedPlaylist.Infinite {
		if m.tracker.IsPlaying() {
			currentTrack := m.tracker.CurrentTrack()
			go m.client.StationFeedback(
				api.ROTOR_SKIP,
				selectedPlaylist.StationId,
				selectedPlaylist.StationBatch,
				currentTrack.Id,
				int(float64(currentTrack.DurationMs*1000)*m.tracker.Progress()),
			)
			go m.client.StationFeedback(
				api.ROTOR_TRACK_STARTED,
				selectedPlaylist.StationId,
				selectedPlaylist.StationBatch,
				trackToPlay.Id,
				0,
			)
		} else {
			go m.client.StationFeedback(
				api.ROTOR_RADIO_STARTED,
				selectedPlaylist.StationId,
				"",
				"",
				0,
			)
		}
	}

	m.currentPlaylistIndex = m.playlists.Index()
	m.playlists.SetItem(m.currentPlaylistIndex, selectedPlaylist)
	m.playTrack(trackToPlay)
}
