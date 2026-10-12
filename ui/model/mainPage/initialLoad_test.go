package mainpage

import (
	"testing"

	"github.com/dece2183/yamusic-tui/api"
)

func TestPinnedAlbumItemsDoNotShareTracks(t *testing.T) {
	albums := []api.Album{
		{Title: "first", Volumes: [][]api.Track{{{Id: "a1"}, {Id: "a2"}}}},
		{Title: "second", Volumes: [][]api.Track{{{Id: "b1"}}}},
	}

	items := pinnedAlbumItems(albums)

	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
	if got := len(items[0].Tracks); got != 2 {
		t.Errorf("first album: expected 2 tracks, got %d", got)
	}
	if got := len(items[1].Tracks); got != 1 {
		t.Errorf("second album: expected 1 track, got %d (tracks leaked from the previous album)", got)
	}
	if got := items[1].Tracks[0].Id; got != "b1" {
		t.Errorf("second album: expected first track b1, got %s", got)
	}
}

func TestPinnedAlbumItemsSkipEmptyAlbums(t *testing.T) {
	albums := []api.Album{
		{Title: "empty"},
		{Title: "second", Volumes: [][]api.Track{{{Id: "b1"}}}},
	}

	items := pinnedAlbumItems(albums)

	if len(items) != 1 {
		t.Fatalf("expected only the non-empty album, got %d items", len(items))
	}
	if items[0].Name != "second" {
		t.Errorf("expected item %q, got %q", "second", items[0].Name)
	}
}
