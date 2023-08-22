package main

import (
	"fmt"
	"time"
	"yamusic/api"
	"yamusic/config"
	"yamusic/ui"

	"github.com/ebitengine/oto/v3"
	"github.com/hajimehoshi/go-mp3"
)

func main() {
	cl, err := api.NewClient(config.GetToken())
	if err != nil {
		panic(err)
	}

	ui.Run(cl)

	playlists, err := cl.ListPlaylists()
	if err != nil {
		panic(err)
	}

	tracks, err := cl.PlaylistTracks(playlists[0].Kind, false)
	if err != nil {
		panic(err)
	}

	dowInfo, err := cl.TrackDownloadInfo(tracks[0].Id)
	if err != nil {
		panic(err)
	}

	var bestBitrate int
	var bestTrackInfo api.TrackDownloadInfo
	for _, t := range dowInfo {
		if t.BbitrateInKbps > bestBitrate {
			bestBitrate = t.BbitrateInKbps
			bestTrackInfo = t
		}
	}

	track, err := cl.DownloadTrack(bestTrackInfo)
	if err != nil {
		panic(err)
	}
	defer track.Close()

	// Decode file. This process is done as the file plays so it won't
	// load the whole thing into memory.
	decodedMp3, err := mp3.NewDecoder(track)
	if err != nil {
		panic("mp3.NewDecoder failed: " + err.Error())
	}

	// Prepare an Oto context (this will use your default audio device) that will
	// play all our sounds. Its configuration can't be changed later.

	op := &oto.NewContextOptions{}

	op.SampleRate = 44100
	op.ChannelCount = 2
	op.Format = oto.FormatSignedInt16LE

	otoCtx, readyChan, err := oto.NewContext(op)
	if err != nil {
		panic("oto.NewContext failed: " + err.Error())
	}
	<-readyChan

	player := otoCtx.NewPlayer(decodedMp3)
	defer player.Close()

	fmt.Printf("plying track id: %d...\n\n", tracks[0].Id)
	player.Play()

	for player.IsPlaying() {
		time.Sleep(time.Millisecond)
	}
}
