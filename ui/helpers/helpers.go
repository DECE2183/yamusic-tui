package helpers

import (
	"strings"

	"github.com/dece2183/yamusic-tui/api"
)

const (
	AI_NOTICE = "ainotice"
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

func CheckTrackAi(client *api.YaMusicClient, track *api.Track) bool {
	credits, err := client.TrackCreadits(track.Id)
	if err != nil {
		return false
	}
	for i := range credits.Credits {
		if strings.Contains(credits.Credits[i].Title, "AI") || strings.Contains(credits.Credits[i].Title, "ИИ") {
			track.Disclaimers = append(track.Disclaimers, AI_NOTICE)
			return true
		}
	}
	return false
}
