package tracklist

import (
	"github.com/dece2183/yamusic-tui/api"
	"github.com/dece2183/yamusic-tui/ui/helpers"
)

type Item struct {
	Track        *api.Track
	Album        *api.Album
	Artists      string
	IsPlaying    bool
	IsSuggestion bool
}

func NewItem(track *api.Track) Item {
	return Item{
		Track:   track,
		Artists: helpers.ArtistList(track.Artists),
	}
}

func NewAlbumItem(album *api.Album) Item {
	return Item{
		Album:   album,
		Artists: helpers.ArtistList(album.Artists),
	}
}

func (i Item) FilterValue() string {
	if i.Track != nil {
		return i.Track.Title
	}
	if i.Album != nil {
		return i.Album.Title
	}
	return ""
}
