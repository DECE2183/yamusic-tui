package model

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model interface {
	// Run the program with the Model. Blocking until the program quits.
	Run() error
	// Send the command to the program in a separate goroutine.
	Send(cmd tea.Cmd)
}

type PlaylistControl uint

const (
	PLAYLIST_CURSOR_UP PlaylistControl = iota
	PLAYLIST_CURSOR_DOWN
)

type TracklistControl uint

const (
	TRACKLIST_PLAY TracklistControl = iota
	TRACKLIST_CURSOR_UP
	TRACKLIST_CURSOR_DOWN
	TRACKLIST_SHARE
	TRACKLIST_LIKE
)

type PlayerControl uint

const (
	PLAYER_PLAY PlayerControl = iota
	PLAYER_PAUSE
	PLAYER_STOP
	PLAYER_NEXT
	PLAYER_PREV
	PLAYER_LIKE
)

type ProgressControl float64

func PrettyExit(err error, code int) {
	fmt.Println()

	if err != nil {
		errMsg := lipgloss.NewStyle().Foreground(lipgloss.Color("#F33")).Render("Error:")
		fmt.Println(errMsg, err, "")
	}

	os.Exit(code)
}

func Cmd(msg tea.Msg) func() tea.Msg {
	return func() tea.Msg {
		return msg
	}
}
