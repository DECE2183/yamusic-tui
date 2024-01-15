package helpers

import (
	"github.com/dece2183/yamusic-tui/api"
)

func ArtistList(artists []api.Artist) (txt string) {
	for _, a := range artists {
		txt += a.Name + ", "
	}
	if len(txt) > 2 {
		txt = txt[:len(txt)-2]
	}
	return
}
