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

const (
	_TRACK_DOWNLOAD_TRIES = 3
)

func (m *Model) rotateTracks() {
	currentPlaylist := m.playlists.Items()[m.currentPlaylistIndex]
	if !currentPlaylist.Rotor {
		return
	}

	currTrack := &currentPlaylist.Tracks[currentPlaylist.CurrentTrack]

	var (
		nextTracks api.StationTracks
		err        error
	)

	if m.tracker.TrackBuffer().IsBuffered() {
		nextTracks, err = m.client.RotorSessionNextTrack(
			currentPlaylist.SessionId, currentPlaylist.SessionBatch,
			currTrack, m.tracker.Playtime().Seconds(),
			currentPlaylist.Tracks,
		)
	} else {
		nextTracks, err = m.client.RotorSessionSkipTrack(
			currentPlaylist.SessionId, currentPlaylist.SessionBatch,
			currTrack, m.tracker.Playtime().Seconds(),
			currentPlaylist.Tracks,
		)
	}

	if err != nil {
		log.Print(log.LVL_ERROR, "failed to obtain more rotor tracks: %s", err)
		m.tracker.ShowError("next track obtain failure")
		m.Send(tracker.STOP)
		return
	}

	currentPlaylist.SessionBatch = nextTracks.BatchId
	currentPlaylist.Tracks = currentPlaylist.Tracks[:currentPlaylist.CurrentTrack+1]
	for _, tr := range nextTracks.Sequence {
		currentPlaylist.Tracks = append(currentPlaylist.Tracks, tr.Track)
	}

	if m.playlists.SelectedItem().IsSame(currentPlaylist) {
		listItems := make([]tracklist.Item, len(currentPlaylist.Tracks))
		for i := range currentPlaylist.Tracks {
			listItems[i] = tracklist.Item{
				Track:        &currentPlaylist.Tracks[i],
				Artists:      helpers.ArtistList(currentPlaylist.Tracks[i].Artists),
				IsSuggestion: i > currentPlaylist.CurrentTrack,
			}
		}
		m.tracklist.SetItems(listItems)
	}
}

func (m *Model) prevTrack() {
	if m.currentPlaylistIndex < 0 {
		return
	}

	currentPlaylist := m.playlists.Items()[m.currentPlaylistIndex]
	if len(currentPlaylist.Tracks) == 0 || currentPlaylist.CurrentTrack == 0 {
		m.Send(tracker.STOP)
		return
	}

	selectedPlaylist := m.playlists.SelectedItem()
	shouldFollow := currentPlaylist.IsSame(selectedPlaylist) && m.tracklist.Index() == currentPlaylist.CurrentTrack

	m.indicateCurrentTrackPlaying(false)

	currentPlaylist.CurrentTrack--
	for currentPlaylist.CurrentTrack > 0 && !currentPlaylist.Tracks[currentPlaylist.CurrentTrack].Available {
		currentPlaylist.CurrentTrack--
	}

	m.rotateTracks()
	m.playlists.SetItem(m.currentPlaylistIndex, currentPlaylist)

	track := &currentPlaylist.Tracks[currentPlaylist.CurrentTrack]
	if !track.Available {
		m.Send(tracker.STOP)
		return
	}

	m.playTrack(track)
	if shouldFollow {
		m.tracklist.Select(currentPlaylist.CurrentTrack)
		currentPlaylist.SelectedTrack = currentPlaylist.CurrentTrack
		m.playlists.SetItem(m.currentPlaylistIndex, currentPlaylist)
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

	if currentPlaylist.CurrentTrack+1 >= len(currentPlaylist.Tracks) {
		currentPlaylist.CurrentTrack = 0
		m.playlists.SetItem(m.currentPlaylistIndex, currentPlaylist)
		m.Send(tracker.STOP)
		return
	}

	selectedPlaylist := m.playlists.SelectedItem()
	shouldFollow := currentPlaylist.IsSame(selectedPlaylist) && m.tracklist.Index() == currentPlaylist.CurrentTrack

	currentPlaylist.CurrentTrack++
	for currentPlaylist.CurrentTrack < len(currentPlaylist.Tracks)-1 && !currentPlaylist.Tracks[currentPlaylist.CurrentTrack].Available {
		currentPlaylist.CurrentTrack++
	}

	m.rotateTracks()
	m.playlists.SetItem(m.currentPlaylistIndex, currentPlaylist)

	track := &currentPlaylist.Tracks[currentPlaylist.CurrentTrack]
	if !track.Available {
		m.Send(tracker.STOP)
		return
	}

	m.playTrack(track)
	if shouldFollow {
		m.tracklist.Select(currentPlaylist.CurrentTrack)
		currentPlaylist.SelectedTrack = currentPlaylist.CurrentTrack
		m.playlists.SetItem(m.currentPlaylistIndex, currentPlaylist)
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
		var trackInfos []api.TrackDownloadInfo
		var bestTrackInfo api.TrackDownloadInfo

		for i := 0; i < _TRACK_DOWNLOAD_TRIES; i++ {
			trackInfos, err = m.client.TrackDownloadInfo(track.Id)
			if err != nil {
				log.Print(log.LVL_ERROR, "failed to obtain track [%s] info: %s", track.Id, err)
				continue
			}

			var bestBitrate int
			for _, t := range trackInfos {
				if t.BbitrateInKbps > bestBitrate {
					bestBitrate = t.BbitrateInKbps
					bestTrackInfo = t
				}
			}

			trackReader, trackSize, err = m.client.DownloadTrack(bestTrackInfo)
			if err != nil {
				log.Print(log.LVL_ERROR, "failed to download track [%s]: %s", track.Id, err)
				continue
			}

			break
		}

		if err != nil {
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
			if len(track.Albums) != 0 {
				tag.SetAlbum(track.Albums[0].Title)
				tag.SetGenre(track.Albums[0].Genre)
				tag.SetYear(fmt.Sprint(track.Albums[0].Year))
			}
			tag.SetArtist(helpers.ArtistList(track.Artists))
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
		if currentPlaylist.Rotor {
			go m.client.RotorSessionTrackStarted(currentPlaylist.SessionId, track)
		}
	}

	m.tracker.StartTrack(track, trackBuffer, lyrics)
	m.indicateCurrentTrackPlaying(true)
	m.mediaHandler.OnPlayback()

	if m.client != nil {
		go m.client.PlayTrack(track, trackFromCache)
	}
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
		if currentPlaylist.IsSame(selectedPlaylist) && selectedPlaylist.CurrentTrack == trackIndex && m.tracker.CurrentTrack().Id == trackToPlay.Id {
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

	if m.currentPlaylistIndex >= 0 {
		currentPlaylist := m.playlists.Items()[m.currentPlaylistIndex]
		if currentPlaylist.Rotor && !currentPlaylist.IsSame(selectedPlaylist) {
			go m.client.RotorSessionRadioFinished(selectedPlaylist.SessionId)
		}
	}

	if selectedPlaylist.Rotor {
		if m.currentPlaylistIndex != m.playlists.Index() {
			go m.client.RotorSessionRadioStarted(selectedPlaylist.SessionId)
		}
		go m.client.RotorSessionTrackStarted(selectedPlaylist.SessionId, trackToPlay)
	}

	m.currentPlaylistIndex = m.playlists.Index()
	m.rotateTracks()

	m.playlists.SetItem(m.currentPlaylistIndex, selectedPlaylist)
	m.playTrack(trackToPlay)
}
