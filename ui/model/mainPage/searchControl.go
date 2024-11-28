package mainpage

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dece2183/yamusic-tui/api"
	"github.com/dece2183/yamusic-tui/config"
	"github.com/dece2183/yamusic-tui/ui/components/playlist"
	"github.com/dece2183/yamusic-tui/ui/components/search"
	"github.com/dece2183/yamusic-tui/ui/helpers"
)

func (m *Model) searchControl(msg search.Control) tea.Cmd {
	var cmd tea.Cmd

	switch msg {
	case search.SELECT:
		m.isSearchActive = false

		req, ok := m.searchDialog.SuggestionValue()
		if !ok {
			return nil
		}

		searchRes, err := m.client.Search(req, api.SEARCH_ALL)
		if err != nil {
			return nil
		}

		cmd = m.displaySearchResults(searchRes)
	case search.CANCEL:
		m.isSearchActive = false
	case search.UPDATE_SUGGESTIONS:
		suggestions, err := m.client.SearchSuggest(m.searchDialog.InputValue())
		if err != nil {
			return nil
		}
		m.searchDialog.SetSuggestions(suggestions.Suggestions)
	}

	return cmd
}

func (m *Model) displaySearchResults(res api.SearchResult) tea.Cmd {
	playlists := m.playlists.Items()
	searchResIndex := len(playlists) + 2
	for i, pl := range playlists {
		if !pl.Active && !pl.Subitem && pl.Name == "search results:" {
			playlists = playlists[:i-1]
			searchResIndex = i + 1
			break
		}
	}

	playlists = append(playlists,
		&playlist.Item{Name: "", Kind: playlist.NONE, Active: false, Subitem: false},
		&playlist.Item{Name: "search results:", Kind: playlist.NONE, Active: false, Subitem: false},
	)

	if len(res.Tracks.Results) > 0 {
		playlists = append(playlists, &playlist.Item{
			Name:    "search \"" + res.Text + "\"",
			Active:  true,
			Subitem: true,
			Tracks:  res.Tracks.Results,
		})
	}

	if config.Current.Search.Artists && len(res.Artists.Results) > 0 {
		// playlists = append(playlists, playlist.Item{Name: "", Kind: playlist.NONE, Active: false, Subitem: false})
		for _, artist := range res.Artists.Results {
			if !strings.Contains(strings.ToLower(artist.Name), strings.ToLower(res.Text)) {
				continue
			}

			artistTracks, err := m.client.ArtistPopularTracks(artist.Id)
			if err != nil {
				continue
			}

			tracks, err := m.client.Tracks(artistTracks.Tracks)
			if err != nil {
				continue
			}

			playlists = append(playlists, &playlist.Item{
				Name:    artist.Name,
				Active:  true,
				Subitem: true,
				Tracks:  tracks,
			})
		}
	}

	if config.Current.Search.Albums && len(res.Albums.Results) > 0 {
		// playlists = append(playlists, playlist.Item{Name: "", Kind: playlist.NONE, Active: false, Subitem: false})
		for _, album := range res.Albums.Results {
			if !strings.Contains(strings.ToLower(album.Title), strings.ToLower(res.Text)) {
				continue
			}

			albumWithTracks, err := m.client.Album(album.Id, true)
			if err != nil {
				continue
			}

			albumArtists := helpers.ArtistList(albumWithTracks.Artists)
			if len(albumWithTracks.Volumes) > 1 {
				for i := range albumWithTracks.Volumes {
					playlists = append(playlists, &playlist.Item{
						Name:    fmt.Sprintf("%s vol.%d (%s)", albumWithTracks.Title, i, albumArtists),
						Active:  true,
						Subitem: true,
						Tracks:  albumWithTracks.Volumes[i],
					})
				}
			} else {
				playlists = append(playlists, &playlist.Item{
					Name:    fmt.Sprintf("%s (%s)", albumWithTracks.Title, albumArtists),
					Active:  true,
					Subitem: true,
					Tracks:  albumWithTracks.Volumes[0],
				})
			}
		}
	}

	if config.Current.Search.Playlists && len(res.Playlists.Results) > 0 {
		// playlists = append(playlists, playlist.Item{Name: "", Kind: playlist.NONE, Active: false, Subitem: false})
		for _, pl := range res.Playlists.Results {
			if !strings.Contains(strings.ToLower(pl.Title), strings.ToLower(res.Text)) {
				continue
			}

			playlistTracks, err := m.client.PlaylistTracks(pl.Kind, pl.Owner.Uid, false)
			if err != nil {
				continue
			}

			playlists = append(playlists, &playlist.Item{
				Name:    pl.Title + " by " + pl.Owner.Name,
				Active:  true,
				Subitem: true,
				Tracks:  playlistTracks,
			})
		}
	}

	cmd := m.playlists.SetItems(playlists)
	m.playlists.Select(searchResIndex)
	m.Send(playlist.CURSOR_DOWN)

	return cmd
}
