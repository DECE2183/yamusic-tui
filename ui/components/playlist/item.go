package playlist

import "github.com/dece2183/yamusic-tui/api"

type Item struct {
	Uid uint64

	Name         string
	Kind         uint64
	Revision     int
	StationId    api.StationId
	StationBatch string
	Active       bool
	Subitem      bool
	Infinite     bool

	Tracks        []api.Track
	CurrentTrack  int
	SelectedTrack int
}

func (i Item) FilterValue() string {
	return i.Name
}

func (i Item) IsSame(other Item) bool {
	return i.Kind == other.Kind && i.Name == other.Name
}
