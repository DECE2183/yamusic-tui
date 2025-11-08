package tracklist

import (
	"github.com/bircoder432/yamusic-tui-enhanced/api"
	"github.com/bircoder432/yamusic-tui-enhanced/ui/helpers"
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
