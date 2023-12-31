package ui

import "yamusic/api"

func artistList(artists []api.Artist) (txt string) {
	for _, a := range artists {
		txt += a.Name + ", "
	}
	if len(txt) > 2 {
		txt = txt[:len(txt)-2]
	}
	return
}
