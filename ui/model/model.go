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

type PlayerControl uint

const (
	PLAYER_PLAY  PlayerControl = iota
	PLAYER_PAUSE PlayerControl = iota
	PLAYER_STOP  PlayerControl = iota
	PLAYER_NEXT  PlayerControl = iota
	PLAYER_PREV  PlayerControl = iota
)

type ProgressControl float64

func PrettyExit(err error, code int) {
	fmt.Println()
	errMsg := lipgloss.NewStyle().Foreground(lipgloss.Color("#F33")).Render("Error:")
	fmt.Println(errMsg, err)
	os.Exit(code)
}
