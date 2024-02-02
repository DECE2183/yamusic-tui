package tracklist

import (
	"github.com/dece2183/yamusic-tui/api"
	"github.com/dece2183/yamusic-tui/ui/helpers"
)

type Item struct {
	Track     *api.Track
	Artists   string
	IsPlaying bool
}

func NewItem(track *api.Track) Item {
	return Item{
		Track:   track,
		Artists: helpers.ArtistList(track.Artists),
	}
}

func (i Item) FilterValue() string {
	return i.Track.Title
}
