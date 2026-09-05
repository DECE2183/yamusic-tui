package mainpage

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"sync"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dece2183/yamusic-tui/api"
	"github.com/dece2183/yamusic-tui/config"
	"github.com/dece2183/yamusic-tui/log"
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
			log.Print(log.LVL_ERROR, "failed to search [%s]: %s", req, err)
			m.tracker.ShowError("search")
			return nil
		}

		m.isLoading = true
		go m.displaySearchResults(searchRes, nil)
		return m.spinner.Tick
	case search.CANCEL:
		m.isSearchActive = false
	case search.UPDATE_SUGGESTIONS:
		suggestions, err := m.client.SearchSuggest(m.searchDialog.InputValue())
		if err != nil {
			log.Print(log.LVL_ERROR, "failed to obtain search [%s] suggestions: %s", m.searchDialog.InputValue(), err)
			m.tracker.ShowError("search seggestion")
			return nil
		}
		m.searchDialog.SetSuggestions(suggestions.Suggestions)
	}

	return cmd
}

func (m *Model) displaySearchResults(res api.SearchResult, searchConfig *config.Search) {
	defer m.Send(LOADING_DONE)

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

	var (
		wg             sync.WaitGroup
		artistBlocks   []menuBlock
		albumBlocks    []menuBlock
		playlistBlocks []menuBlock
	)

	if searchConfig == nil {
		searchConfig = config.Current.Search
	}

	if searchConfig.Artists && len(res.Artists.Results) > 0 {
		artistBlocks = make([]menuBlock, len(res.Artists.Results))
		for i := range res.Artists.Results {
			artist := &res.Artists.Results[i]
			if artist.Id == 0 || !strings.Contains(strings.ToLower(artist.Name), strings.ToLower(res.Text)) {
				continue
			}
			wg.Add(1)
			go m.loadSearchArtist(&wg, artist, searchConfig.Popular, &artistBlocks[i])
		}
	}

	if searchConfig.Albums && len(res.Albums.Results) > 0 {
		albumBlocks = make([]menuBlock, len(res.Albums.Results))
		for i := range res.Albums.Results {
			album := &res.Albums.Results[i]
			if album.Id == 0 || !strings.Contains(strings.ToLower(album.Title), strings.ToLower(res.Text)) {
				continue
			}
			wg.Add(1)
			go m.loadSearchAlbums(&wg, album, &albumBlocks[i])
		}
	}

	if searchConfig.Playlists && len(res.Playlists.Results) > 0 {
		playlistBlocks = make([]menuBlock, len(res.Playlists.Results))
		for i := range res.Playlists.Results {
			pl := &res.Playlists.Results[i]
			if pl.Kind == 0 || pl.Owner.Uid == 0 || !strings.Contains(strings.ToLower(pl.Title), strings.ToLower(res.Text)) {
				continue
			}
			wg.Add(1)
			go m.loadSearchPlaylists(&wg, pl, &albumBlocks[i])
		}
	}

	wg.Wait()
	for i := range artistBlocks {
		if artistBlocks[i].err != nil || len(artistBlocks[i].items) == 0 {
			continue
		}
		playlists = append(playlists, artistBlocks[i].items...)
	}
	for i := range albumBlocks {
		if albumBlocks[i].err != nil || len(albumBlocks[i].items) == 0 {
			continue
		}
		playlists = append(playlists, albumBlocks[i].items...)
	}
	for i := range playlistBlocks {
		if playlistBlocks[i].err != nil || len(playlistBlocks[i].items) == 0 {
			continue
		}
		playlists = append(playlists, playlistBlocks[i].items...)
	}

	if !playlists[len(playlists)-1].Active {
		return
	}

	m.playlists.SetItems(playlists)
	m.playlists.Select(searchResIndex)
	m.Send(playlist.CURSOR_DOWN)
}

func (m *Model) loadSearchArtist(wg *sync.WaitGroup, artist *api.Artist, withPopular bool, block *menuBlock) {
	defer wg.Done()

	if m.client == nil {
		return
	}

	if withPopular {
		artistTracks, err := m.client.ArtistPopularTracks(artist.Id)
		if err != nil {
			sval, _ := m.searchDialog.SuggestionValue()
			log.Print(log.LVL_ERROR, "failed to obtain search [%s] artist [%s] tracks: %s", sval, artist.Name, err)
			block.err = errors.New("search artist tracks")
			return
		}

		tracks, err := m.client.Tracks(artistTracks.Tracks)
		if err != nil {
			sval, _ := m.searchDialog.SuggestionValue()
			log.Print(log.LVL_ERROR, "failed to obtain search [%s] artist [%s] tracks full info: %s", sval, artist.Name, err)
			block.err = errors.New("search artist tracks info")
			return
		}

		block.items = append(block.items, &playlist.Item{
			Name:    artist.Name + " popular",
			Active:  true,
			Subitem: true,
			Tracks:  tracks,
		})
	}

	artistAlbums, err := m.client.ArtistAlbums(artist.Id)
	if err != nil {
		sval, _ := m.searchDialog.SuggestionValue()
		log.Print(log.LVL_ERROR, "failed to obtain search [%s] artist [%s] albums: %s", sval, artist.Name, err)
		block.err = errors.New("search artist albums")
		return
	}

	albums := make([]api.Album, 0, len(artistAlbums))
	for i := range artistAlbums {
		album, err := m.client.Album(artistAlbums[i].Id, true)
		if err != nil {
			log.Print(log.LVL_ERROR, "failed to obtain album [%d] info: %s", album.Id, err)
			continue
		}
		albums = append(albums, album)
	}

	if len(albums) > 0 {
		slices.SortFunc(albums, func(a, b api.Album) int {
			return b.Year - a.Year
		})
		block.items = append(block.items, &playlist.Item{
			Name:          artist.Name + " albums",
			Kind:          playlist.ALBUMS,
			Active:        true,
			Subitem:       true,
			Albums:        albums,
			SelectedAlbum: -1,
		})
	}
}

func (m *Model) loadSearchAlbums(wg *sync.WaitGroup, album *api.Album, block *menuBlock) {
	defer wg.Done()

	if m.client == nil {
		return
	}

	albumWithTracks, err := m.client.Album(album.Id, true)
	if err != nil {
		sval, _ := m.searchDialog.SuggestionValue()
		log.Print(log.LVL_ERROR, "failed to obtain search [%s] album [%s] tracks: %s", sval, album.Title, err)
		block.err = errors.New("search album tracks")
		return
	}

	albumArtists := helpers.ArtistList(albumWithTracks.Artists)
	albumTracks := make([]api.Track, 0, len(albumWithTracks.Volumes[0]))
	for _, volume := range albumWithTracks.Volumes {
		albumTracks = append(albumTracks, volume...)
	}

	block.items = append(block.items, &playlist.Item{
		Name:    fmt.Sprintf("%s (%s)", albumWithTracks.Title, albumArtists),
		Active:  true,
		Subitem: true,
		Tracks:  albumTracks,
	})
}

func (m *Model) loadSearchPlaylists(wg *sync.WaitGroup, pl *api.Playlist, block *menuBlock) {
	defer wg.Done()

	if m.client == nil {
		return
	}

	playlistTracks, err := m.client.PlaylistTracks(pl.Kind, pl.Owner.Uid, false)
	if err != nil {
		sval, _ := m.searchDialog.SuggestionValue()
		log.Print(log.LVL_ERROR, "failed to obtain search [%s] playlist [%s] tracks: %s", sval, pl.Title, err)
		block.err = errors.New("search playlist tracks")
		return
	}

	block.items = append(block.items, &playlist.Item{
		Name:    pl.Title + " by " + pl.Owner.Name,
		Active:  true,
		Subitem: true,
		Tracks:  playlistTracks,
	})
}
