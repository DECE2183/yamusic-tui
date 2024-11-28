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

func (i *Item) FilterValue() string {
	return i.Name
}

func (i *Item) IsSame(other *Item) bool {
	return i.Kind == other.Kind && i.Name == other.Name
}

func (pl *Item) AddTrack(track *api.Track) {
	pl.Tracks = append([]api.Track{*track}, pl.Tracks...)
}

func (pl *Item) AddTrackToEnd(track *api.Track) {
	pl.Tracks = append(pl.Tracks, *track)
}

func (pl *Item) RemoveTrack(trackId string) int {
	for i, ltrack := range pl.Tracks {
		if ltrack.Id == trackId {
			if i+1 < len(pl.Tracks) {
				pl.Tracks = append(pl.Tracks[:i], pl.Tracks[i+1:]...)
			} else {
				pl.Tracks = pl.Tracks[:i]
			}
			if pl.SelectedTrack == len(pl.Tracks) && len(pl.Tracks) > 0 {
				pl.SelectedTrack--
			}
			return i
		}
	}
	return -1
}
