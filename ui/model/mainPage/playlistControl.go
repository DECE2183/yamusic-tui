package mainpage

import (
	"math/rand"
	"slices"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dece2183/yamusic-tui/api"
	"github.com/dece2183/yamusic-tui/ui/components/input"
	"github.com/dece2183/yamusic-tui/ui/components/playlist"
	"github.com/dece2183/yamusic-tui/ui/components/search"
	"github.com/dece2183/yamusic-tui/ui/components/tracklist"
)

func (m *Model) addPlaylistControl(msg search.Control) tea.Cmd {
	var cmd tea.Cmd

	switch msg {
	case search.SELECT:
		m.isAddPlaylistActive = false

		selectedPlaylist := m.playlists.SelectedItem()
		if len(selectedPlaylist.Tracks) == 0 {
			return nil
		}

		playlists := m.playlists.Items()
		inputVal, ok := m.searchDialog.SuggestionValue()
		if !ok {
			return nil
		}

		foundPlaylistIndex := -1
		var foundPlaylist *playlist.Item
		for i := range playlists {
			if playlists[i].Active && playlists[i].Kind >= playlist.USER {
				if strings.EqualFold(playlists[i].Name, inputVal) {
					foundPlaylist = playlists[i]
					foundPlaylistIndex = i
					break
				} else if foundPlaylistIndex < 0 {
					foundPlaylistIndex = i
				}
			}
		}

		if foundPlaylist == nil {
			pl, err := m.client.CreatePlaylist(inputVal, true)
			if err != nil {
				return nil
			}

			foundPlaylist = &playlist.Item{
				Name:     pl.Title,
				Kind:     pl.Kind,
				Revision: pl.Revision,
				Active:   true,
				Subitem:  true,
			}

			m.playlists.InsertItem(foundPlaylistIndex, foundPlaylist)
			if foundPlaylistIndex < m.playlists.Index() {
				m.playlists.Select(m.playlists.Index() + 1)
			}
		}

		if selectedPlaylist.Kind == foundPlaylist.Kind {
			return nil
		}

		selectedTrack := &selectedPlaylist.Tracks[m.tracklist.Index()]
		pl, err := m.client.AddToPlaylist(foundPlaylist.Kind, foundPlaylist.Revision, len(foundPlaylist.Tracks), selectedTrack.Id)
		if err != nil {
			return nil
		}

		foundPlaylist.Revision = pl.Revision
		foundPlaylist.Tracks = append(foundPlaylist.Tracks, *selectedTrack)
		cmd = m.playlists.SetItem(foundPlaylistIndex, foundPlaylist)

		m.isAddPlaylistActive = false
	case search.CANCEL:
		m.isAddPlaylistActive = false
	case search.UPDATE_SUGGESTIONS:
		inputVal := strings.ToLower(m.searchDialog.InputValue())
		playlists := m.playlists.Items()
		suggestions := make([]string, 0, len(playlists))
		for _, pl := range playlists {
			if !pl.Active || pl.Kind < playlist.USER || (len(inputVal) > 0 && !strings.Contains(strings.ToLower(pl.Name), inputVal)) {
				continue
			}
			suggestions = append(suggestions, pl.Name)
		}
		m.searchDialog.SetSuggestions(suggestions)
	}

	return cmd
}

func (m *Model) renamePlaylistControl(msg input.Control) tea.Cmd {
	var cmd tea.Cmd

	if msg != input.APPLY {
		return nil
	}

	newName := m.inputDialog.Value()
	if len(strings.ReplaceAll(newName, " ", "")) == 0 {
		return nil
	}

	selectedPlaylist := m.playlists.SelectedItem()
	pl, err := m.client.RenamePlaylist(selectedPlaylist.Kind, newName)
	if err != nil {
		return nil
	}

	selectedPlaylist.Name = pl.Title
	selectedPlaylist.Revision = pl.Revision
	m.playlists.SetItem(m.playlists.Index(), selectedPlaylist)

	return cmd
}

func (m *Model) removeFromPlaylist(pl *playlist.Item, index int) tea.Cmd {
	if index >= len(pl.Tracks) {
		return nil
	}

	switch pl.Kind {
	case playlist.NONE, playlist.MYWAVE:
		return nil
	case playlist.LIKES:
		selectedTrack := pl.Tracks[index]
		return m.likeTrack(&selectedTrack)
	case playlist.LOCAL:
		selectedTrack := pl.Tracks[index]
		return m.removeCache(&selectedTrack)
	default:
		var cmd tea.Cmd

		if len(pl.Tracks) < 2 {
			err := m.client.RemovePlaylist(pl.Kind)
			if err != nil {
				return nil
			}
			playlists := m.playlists.Items()
			if m.currentPlaylistIndex >= 0 {
				currentPlaylist := playlists[m.currentPlaylistIndex]
				if pl.IsSame(currentPlaylist) && m.tracker.IsPlaying() {
					m.currentPlaylistIndex = -1
				}
			}
			m.playlists.RemoveItem(m.playlists.Index())
			if len(playlists) <= 1 {
				m.playlists.Select(0)
			}
			m.displayPlaylist(m.playlists.SelectedItem())
			return nil
		}

		newpl, err := m.client.RemoveFromPlaylist(pl.Kind, pl.Revision, index)
		if err != nil {
			return nil
		}

		pl.Revision = newpl.Revision
		pl.Tracks = slices.Delete(pl.Tracks, index, index+1)
		if index >= len(pl.Tracks) {
			pl.SelectedTrack = len(pl.Tracks) - 1
		} else {
			pl.SelectedTrack = index
		}
		deleteCurrentTrack := index == pl.CurrentTrack
		if deleteCurrentTrack {
			pl.CurrentTrack = len(pl.Tracks)
		} else if pl.CurrentTrack > index {
			pl.CurrentTrack--
		}
		cmd = m.playlists.SetItem(m.playlists.Index(), pl)
		m.displayPlaylist(pl)

		if m.currentPlaylistIndex >= 0 {
			currentPlaylist := m.playlists.Items()[m.currentPlaylistIndex]
			if pl.IsSame(currentPlaylist) && m.tracker.IsPlaying() {
				m.indicateCurrentTrackPlaying(!deleteCurrentTrack)
			}
		}

		return cmd
	}
}

func (m *Model) shufflePlaylist(pl *playlist.Item) tea.Cmd {
	var cmds []tea.Cmd
	if pl.Kind == playlist.NONE || pl.Kind == playlist.MYWAVE || len(pl.Tracks) == 0 {
		return nil
	}

	currentTrackIndex := pl.CurrentTrack
	selectedTrackIndex := pl.SelectedTrack
	currentTrack := pl.Tracks[currentTrackIndex]
	selectedTrack := pl.Tracks[selectedTrackIndex]

	tracks := make([]api.Track, len(pl.Tracks))
	trackList := make([]tracklist.Item, len(pl.Tracks))
	perm := rand.Perm(len(tracks))

	for i, v := range perm {
		tracks[v] = pl.Tracks[i]
		trackList[v] = tracklist.NewItem(&tracks[v])
		if currentTrack.Id == tracks[v].Id {
			currentTrackIndex = v
		}
		if selectedTrackIndex > 0 && selectedTrack.Id == tracks[v].Id {
			selectedTrackIndex = v
		}
	}

	pl.Tracks = tracks
	pl.SelectedTrack = selectedTrackIndex
	pl.CurrentTrack = currentTrackIndex
	cmds = append(cmds, m.playlists.SetItem(m.playlists.Index(), pl))
	cmds = append(cmds, m.tracklist.SetItems(trackList))
	m.tracklist.Select(selectedTrackIndex)

	if m.currentPlaylistIndex >= 0 {
		currentPlaylist := m.playlists.Items()[m.currentPlaylistIndex]
		if pl.IsSame(currentPlaylist) && m.tracker.IsPlaying() {
			m.indicateCurrentTrackPlaying(true)
		}
	}

	return tea.Batch(cmds...)
}

func (m *Model) displayPlaylist(pl *playlist.Item) {
	trackList := make([]tracklist.Item, len(pl.Tracks))
	for i := range pl.Tracks {
		trackList[i] = tracklist.NewItem(&pl.Tracks[i])
	}
	m.tracklist.SetItems(trackList)
	m.tracklist.Select(pl.SelectedTrack)
	switch pl.Kind {
	case playlist.MYWAVE:
		m.tracklist.Title = "My wave"
	case playlist.LIKES:
		m.tracklist.Title = "Liked tracks"
	case playlist.LOCAL:
		m.tracklist.Title = "Cached tracks"
	default:
		m.tracklist.Title = "Tracks from " + pl.Name
	}
}

func (m *Model) indicateCurrentTrackPlaying(playing bool) {
	if m.currentPlaylistIndex < 0 {
		return
	}
	currentPlaylist := m.playlists.Items()[m.currentPlaylistIndex]
	if currentPlaylist.IsSame(m.playlists.SelectedItem()) && currentPlaylist.CurrentTrack < len(m.tracklist.Items()) {
		track := m.tracklist.Items()[currentPlaylist.CurrentTrack]
		track.IsPlaying = playing
		m.tracklist.SetItem(currentPlaylist.CurrentTrack, track)
	}
}
