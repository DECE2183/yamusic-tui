package cache

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/bogem/id3v2/v2"
	"github.com/dece2183/yamusic-tui/api"
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

		artistNames := strings.Split(tag.Artist(), ",")
		artists := make([]api.Artist, len(artistNames))
		for i := range artistNames {
			artists[i].Name = artistNames[i]
		}

		stat, _ := entry.Info()
		year, _ := strconv.Atoi(tag.Year())
		durationMs, _ := strconv.Atoi(tag.GetTextFrame("TLEN").Text)

		if durationMs > 0 {
			tracks = append(tracks, api.Track{
				Id:         name[:len(name)-len(ext)],
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
