package mainpage

import (
	"testing"

	"github.com/dece2183/yamusic-tui/api"
	"github.com/dece2183/yamusic-tui/config"
	"github.com/dece2183/yamusic-tui/ui/components/playlist"
	"github.com/dece2183/yamusic-tui/ui/components/tracklist"
)

// A pinned album is stored as a playlist.Item with Kind == ALBUMS but with
// Tracks filled and Albums left empty, and thus SelectedAlbum keeps its zero
// value (0). displayPlaylist must not index into the empty Albums slice.
func TestDisplayPlaylistPinnedAlbumDoesNotPanic(t *testing.T) {
	if err := config.InitialLoad(); err != nil {
		t.Fatalf("config init: %v", err)
	}

	likes := make(map[string]bool)
	cache := make(map[string]bool)

	m := &Model{
		tracklist: tracklist.New(nil, &likes, &cache),
	}

	pl := &playlist.Item{
		Name:    "Pinned Album",
		Kind:    playlist.ALBUMS,
		Active:  true,
		Subitem: true,
		Tracks:  []api.Track{{Id: "1", Title: "track"}},
	}

	m.displayPlaylist(pl)
}
