package ui

import "github.com/charmbracelet/lipgloss"

func (m model) renderLoginPage() string {
	title := dialogTitleStyle.Render("Enter your token")
	buttons := lipgloss.Place(42, 1, lipgloss.Right, lipgloss.Center, activeButtonStyle.Render("Ok"))
	content := lipgloss.JoinVertical(lipgloss.Left, title, m.loginTextInput.View(), buttons)

	dialog := lipgloss.Place(m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		dialogBoxStyle.Render(content),
	)

	return dialog
}
