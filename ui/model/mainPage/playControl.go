package mainpage

import (
	"fmt"
	"io"
	"os"

	_ "image/jpeg"
	_ "image/png"

	"github.com/bogem/id3v2/v2"
	"github.com/dece2183/yamusic-tui/api"
	"github.com/dece2183/yamusic-tui/cache"
	"github.com/dece2183/yamusic-tui/log"
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

	track := &currentPlaylist.Tracks[currentPlaylist.CurrentTrack]
	if !track.Available {
		m.Send(tracker.PREV)
		return
	}

	m.playTrack(track)
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
				log.Print(log.LVL_ERROR, "failed to obtain more station tracks: %s", err)
				m.tracker.ShowError("station tracks")
				m.Send(tracker.STOP)
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

	track := &currentPlaylist.Tracks[currentPlaylist.CurrentTrack]
	if !track.Available {
		m.Send(tracker.NEXT)
		return
	}

	m.playTrack(track)
	selectedPlaylist := m.playlists.SelectedItem()
	if currentPlaylist.IsSame(selectedPlaylist) && m.tracklist.Index() == currentPlaylist.CurrentTrack-1 {
		m.tracklist.Select(currentPlaylist.CurrentTrack)
	}
}

func (m *Model) playTrack(track *api.Track) {
	m.tracker.Stop()

	var (
		coverFile  *os.File
		coverStat  os.FileInfo
		coverType  string
		coverBytes []byte
		err        error
	)

	coverPath := m.coverFilePath(track)
	coverFile, err = os.OpenFile(coverPath, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0755)
	if err != nil {
		log.Print(log.LVL_WARNIGN, "unable to open cover file [%s]: %s", coverPath, err)
		goto skipcover
	}

	defer coverFile.Close()

	coverStat, err = coverFile.Stat()
	if err != nil || coverStat.Size() == 0 {
		coverType, err = api.DownloadTrackCover(coverFile, track, 200)
		if err != nil {
			log.Print(log.LVL_WARNIGN, "unable to download track [%s] cover: %s", track.Id, err)
			goto skipcover
		}
		coverFile.Sync()
		stat, _ := coverFile.Stat()
		coverBytes = make([]byte, stat.Size())
		coverFile.Seek(0, io.SeekStart)
		coverFile.Read(coverBytes)
	}

skipcover:
	var trackFromCache bool
	var trackBuffer *stream.BufferedStream
	var trackReader io.ReadCloser
	var trackSize int64
	var lyrics []api.LyricPair
	if track.LyricsInfo.HasAvailableSyncLyrics {
		lyrics, err = m.client.TrackLyricsRequest(track.Id)
		if err != nil {
			log.Print(log.LVL_WARNIGN, "failed to obtain track [%s] lyrics: %s", track.Id, err)
			m.tracker.ShowError("track lyrics")
		}
	}
	trackReader, trackSize, err = cache.Read(track.Id)
	if err == nil {
		trackFromCache = true
	} else {
		dowInfo, err := m.client.TrackDownloadInfo(track.Id)
		if err != nil {
			log.Print(log.LVL_ERROR, "failed to obtain track [%s] info: %s", track.Id, err)
			m.tracker.ShowError("track info")
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

		trackReader, trackSize, err = m.client.DownloadTrack(bestTrackInfo)
		if err != nil {
			log.Print(log.LVL_ERROR, "failed to download track [%s]: %s", track.Id, err)
			m.tracker.ShowError("track download")
			return
		}
	}

	trackBuffer = stream.NewBufferedStream(trackReader, trackSize)
	metadataFile, err := os.OpenFile(m.metadataFilePath(), os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0755)
	if err == nil {
		tag := id3v2.NewEmptyTag()
		if trackFromCache {
			tag.Reset(trackBuffer, id3v2.Options{Parse: true})
		} else {
			tag.SetDefaultEncoding(id3v2.EncodingUTF8)
			tag.SetTitle(track.Title)
			tag.SetAlbum(track.Albums[0].Title)
			tag.SetGenre(track.Albums[0].Genre)
			tag.SetArtist(helpers.ArtistList(track.Artists))
			tag.SetYear(fmt.Sprint(track.Albums[0].Year))
			tag.AddAttachedPicture(id3v2.PictureFrame{
				MimeType:    coverType,
				PictureType: id3v2.PTFrontCover,
				Encoding:    id3v2.EncodingUTF16BE,
				Picture:     coverBytes,
			})
			tag.AddFrame("TLEN", id3v2.TextFrame{
				Encoding: id3v2.EncodingUTF8,
				Text:     fmt.Sprint(track.DurationMs),
			})
		}
		tag.WriteTo(metadataFile)
		io.CopyN(metadataFile, trackBuffer, 32*1024)
		trackBuffer.Seek(0, io.SeekStart)
		metadataFile.Close()
	} else {
		log.Print(log.LVL_WARNIGN, "failed to create metadata file: %s", err)
	}

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

	m.tracker.StartTrack(track, trackBuffer, lyrics)
	m.indicateCurrentTrackPlaying(true)
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
