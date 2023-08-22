package ui

func (m model) renderMainPage() string {
	sidePanel := m.sideList.View()

	return sideBoxStyle.Render(sidePanel)
}
