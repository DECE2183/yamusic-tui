package mainpage

import (
	"fmt"
	"io"
	"os"
	"sync"

	_ "image/jpeg"
	_ "image/png"

	"github.com/bogem/id3v2/v2"
	"github.com/dece2183/yamusic-tui/api"
	"github.com/dece2183/yamusic-tui/cache"
	"github.com/dece2183/yamusic-tui/log"
	"github.com/dece2183/yamusic-tui/stream"
	"github.com/dece2183/yamusic-tui/ui/components/playlist"
	"github.com/dece2183/yamusic-tui/ui/components/tracker"
	"github.com/dece2183/yamusic-tui/ui/components/tracklist"
	"github.com/dece2183/yamusic-tui/ui/helpers"
)

const (
	_TRACK_DOWNLOAD_TRIES     = 3
	_TRACK_FINISHED_THRESHOLD = 0.8
)

func (m *Model) feedbackOnTrack(batch string) *api.RotorFeedback {
	currTrack := m.tracker.CurrentTrack()
	if currTrack == nil {
		return nil
	}
	var evType api.TrackEventType
	if m.tracker.Progress() > _TRACK_FINISHED_THRESHOLD {
		evType = api.EV_TRACK_FINISHED
	} else {
		evType = api.EV_TRACK_SKIPED
	}
	ev := api.NewTrackFeedbackEvent(evType, currTrack, m.tracker.Playtime().Seconds())
	fb := api.NewFeedback(batch, ev)
	log.Print(log.LVL_INFO, "feedback event sended: "+ev.Type+" track: "+currTrack.Title)
	return fb
}

func (m *Model) rotateTracks(currentPlaylist *playlist.Item) {
	if !currentPlaylist.Rotor {
		return
	}

	suggestedTracks, err := m.client.RotorSessionTracks(currentPlaylist.SessionId, []*api.RotorFeedback{}, currentPlaylist.Tracks)
	if err != nil {
		log.Print(log.LVL_ERROR, "failed to obtain more rotor tracks: %s", err)
		m.tracker.ShowError("next track obtain failure")
		m.Send(tracker.STOP)
		return
	}

	currentPlaylist.SessionBatch = suggestedTracks.BatchId

	existing := make(map[string]bool, len(currentPlaylist.Tracks))
	for _, t := range currentPlaylist.Tracks {
		existing[t.Id] = true
	}
	addedStart := len(currentPlaylist.Tracks)
	for _, item := range suggestedTracks.Sequence {
		if existing[item.Track.Id] {
			continue
		}
		existing[item.Track.Id] = true
		currentPlaylist.Tracks = append(currentPlaylist.Tracks, item.Track)
	}

	if len(currentPlaylist.Tracks) == addedStart {
		return
	}

	if m.playlists.SelectedItem().IsSame(currentPlaylist) {
		tackItems := m.tracklist.Items()
		if len(tackItems) > 0 {
			lastTrack := tackItems[len(tackItems)-1]
			lastTrack.IsSuggestion = false
			m.tracklist.SetItem(len(tackItems)-1, lastTrack)
		}
		for i := addedStart; i < len(currentPlaylist.Tracks); i++ {
			item := tracklist.Item{
				Track:   &currentPlaylist.Tracks[i],
				Artists: helpers.ArtistList(currentPlaylist.Tracks[i].Artists),
			}
			if i == len(currentPlaylist.Tracks)-1 {
				item.IsSuggestion = true
			}
			m.tracklist.InsertItem(-1, item)
		}
	}
}

func (m *Model) prevTrack() {
	if m.currentPlaylistIndex < 0 {
		return
	}

	currentPlaylist := m.playlists.Items()[m.currentPlaylistIndex]

	if currentPlaylist.Rotor && m.tracker.IsPlaying() {
		go m.client.RotorSessionFeedback(currentPlaylist.SessionId, m.feedbackOnTrack(currentPlaylist.SessionBatch))
	}

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

	if currentPlaylist.Rotor && m.tracker.IsPlaying() {
		go m.client.RotorSessionFeedback(currentPlaylist.SessionId, m.feedbackOnTrack(currentPlaylist.SessionBatch))
	}

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

	m.playlists.SetItem(m.currentPlaylistIndex, currentPlaylist)
	track := &currentPlaylist.Tracks[currentPlaylist.CurrentTrack]
	if !track.Available {
		m.Send(tracker.STOP)
		return
	}

	if currentPlaylist.CurrentTrack == len(currentPlaylist.Tracks)-1 {
		m.rotateTracks(currentPlaylist)
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
		wg sync.WaitGroup

		coverType  string
		coverBytes []byte

		lyrics []api.LyricPair

		trackReader    io.ReadCloser
		trackSize      int64
		trackFromCache bool
		downloadErr    error
	)

	wg.Add(1)
	go func() {
		defer wg.Done()
		coverPath := m.coverFilePath(track)
		coverFile, ferr := os.OpenFile(coverPath, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0755)
		if ferr != nil {
			log.Print(log.LVL_WARNIGN, "unable to open cover file [%s]: %s", coverPath, ferr)
			return
		}
		defer coverFile.Close()
		stat, serr := coverFile.Stat()
		if serr != nil || stat.Size() == 0 {
			cType, derr := api.DownloadTrackCover(coverFile, track, 200)
			if derr != nil {
				log.Print(log.LVL_WARNIGN, "unable to download track [%s] cover: %s", track.Id, derr)
				return
			}
			coverType = cType
			coverFile.Sync()
			s, _ := coverFile.Stat()
			buf := make([]byte, s.Size())
			coverFile.Seek(0, io.SeekStart)
			coverFile.Read(buf)
			coverBytes = buf
		}
	}()

	if track.LyricsInfo.HasAvailableSyncLyrics {
		wg.Add(1)
		go func() {
			defer wg.Done()
			lyr, lerr := m.client.TrackLyricsRequest(track.Id)
			if lerr != nil {
				log.Print(log.LVL_WARNIGN, "failed to obtain track [%s] lyrics: %s", track.Id, lerr)
				m.tracker.ShowError("track lyrics")
				return
			}
			lyrics = lyr
		}()
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		tr, ts, cerr := cache.Read(track.Id)
		if cerr == nil {
			trackReader = tr
			trackSize = ts
			trackFromCache = true
			return
		}
		var lastErr error
		for i := 0; i < _TRACK_DOWNLOAD_TRIES; i++ {
			trackInfos, ierr := m.client.TrackDownloadInfo(track.Id)
			if ierr != nil {
				log.Print(log.LVL_ERROR, "failed to obtain track [%s] info: %s", track.Id, ierr)
				lastErr = ierr
				continue
			}
			var bestBitrate int
			var bestTrackInfo api.TrackDownloadInfo
			for _, ti := range trackInfos {
				if ti.BbitrateInKbps > bestBitrate {
					bestBitrate = ti.BbitrateInKbps
					bestTrackInfo = ti
				}
			}
			tr2, ts2, derr := m.client.DownloadTrack(bestTrackInfo)
			if derr != nil {
				log.Print(log.LVL_ERROR, "failed to download track [%s]: %s", track.Id, derr)
				lastErr = derr
				continue
			}
			trackReader = tr2
			trackSize = ts2
			lastErr = nil
			break
		}
		downloadErr = lastErr
	}()

	wg.Wait()

	if downloadErr != nil && trackReader == nil {
		m.tracker.ShowError("track download")
		return
	}

	trackBuffer := stream.NewBufferedStream(trackReader, trackSize)
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
			ev := api.NewTrackFeedbackEvent(api.EV_TRACK_STARTED, track, 0)
			go m.client.RotorSessionFeedback(currentPlaylist.SessionId, api.NewFeedback(currentPlaylist.SessionBatch, ev))
			log.Print(log.LVL_INFO, "feedback event sended: "+ev.Type+" track: "+track.Title)
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
		if currentPlaylist.Rotor {
			if m.tracker.IsPlaying() {
				go m.client.RotorSessionFeedback(currentPlaylist.SessionId, m.feedbackOnTrack(currentPlaylist.SessionBatch))
			}
			if !currentPlaylist.IsSame(selectedPlaylist) {
				ev := api.NewRadioFeedbackEvent(api.EV_RADIO_FINISHED)
				go m.client.RotorSessionFeedback(currentPlaylist.SessionId, api.NewFeedback(currentPlaylist.SessionBatch, ev))
				log.Print(log.LVL_INFO, "feedback event sended: "+ev.Type)
			}
		}
	}

	m.indicateCurrentTrackPlaying(false)

	if selectedPlaylist.Rotor {
		if trackIndex == len(selectedPlaylist.Tracks)-1 {
			m.rotateTracks(selectedPlaylist)
		}
		if m.currentPlaylistIndex != m.playlists.Index() {
			ev := api.NewRadioFeedbackEvent(api.EV_RADIO_STARTED)
			go m.client.RotorSessionFeedback(selectedPlaylist.SessionId, api.NewFeedback("", ev))
			log.Print(log.LVL_INFO, "feedback event sended: "+ev.Type)
		}
	}

	selectedPlaylist.CurrentTrack = trackIndex
	m.currentPlaylistIndex = m.playlists.Index()
	m.playlists.SetItem(m.currentPlaylistIndex, selectedPlaylist)
	m.playTrack(trackToPlay)
}
