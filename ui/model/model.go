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
