package playlist

import "github.com/dece2183/yamusic-tui/api"

type Item struct {
	Name         string
	Kind         uint64
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
