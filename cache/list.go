package cache

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/bogem/id3v2/v2"
	"github.com/dece2183/yamusic-tui/api"
)

const (
	TAG_TRACK_ID  = "TRID"
	TAG_ARTIST_ID = "ARID"
	TAG_DURATION  = "TLEN"
)

func ListTracks() ([]api.Track, error) {
	dir, err := getCacheDir()
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	tracks := make([]api.Track, 0, len(entries))
	for _, entry := range entries {
		name := entry.Name()
		ext := strings.ToLower(filepath.Ext(name))
		if entry.IsDir() || ext != ".mp3" {
			continue
		}

		tag, err := id3v2.Open(filepath.Join(dir, name), id3v2.Options{Parse: true})
		if err != nil {
			continue
		}

		idFrame, ok := tag.GetLastFrame(TAG_TRACK_ID).(id3v2.TextFrame)
		var trackId string
		if ok && len(idFrame.Text) > 0 {
			trackId = idFrame.Text
		} else {
			trackId = name[:len(name)-len(ext)]
		}

		artistNames := strings.Split(tag.Artist(), ",")
		artists := make([]api.Artist, len(artistNames))
		artistIdsFrames := tag.GetFrames(TAG_ARTIST_ID)
		for i := range artistNames {
			artists[i].Name = artistNames[i]
			if i < len(artistIdsFrames) {
				artistId, ok := artistIdsFrames[i].(id3v2.UnknownFrame)
				if !ok {
					continue
				}
				_, str, _ := strings.Cut(string(artistId.Body), string([]byte{0}))
				artists[i].Id, _ = strconv.ParseUint(str, 10, 64)
			}
		}

		stat, _ := entry.Info()
		year, _ := strconv.Atoi(tag.Year())

		durationFrame, ok := tag.GetLastFrame(TAG_DURATION).(id3v2.TextFrame)
		if !ok {
			continue
		}

		durationMs, _ := strconv.Atoi(durationFrame.Text)
		if durationMs > 0 {
			tracks = append(tracks, api.Track{
				Id:         trackId,
				Title:      tag.Title(),
				Available:  true,
				FileSize:   int(stat.Size()),
				DurationMs: int(durationMs),
				Artists:    artists,
				Albums: []api.Album{
					{
						Title: tag.Album(),
						Genre: tag.Genre(),
						Year:  year,
					},
				},
			})
		}

		tag.Close()
	}

	return tracks, nil
}
