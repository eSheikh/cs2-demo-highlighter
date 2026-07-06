package tui

import "github.com/charmbracelet/lipgloss"

const (
	accent = lipgloss.Color("205")
	subtle = lipgloss.Color("241")
	good   = lipgloss.Color("42")
	bad    = lipgloss.Color("196")
	light  = lipgloss.Color("231")
)

var (
	headerStyle = lipgloss.NewStyle().Bold(true).Foreground(light).Background(accent).Padding(0, 1)
	stepStyle   = lipgloss.NewStyle().Foreground(light).Background(lipgloss.Color("238")).Padding(0, 1)
	footerStyle = lipgloss.NewStyle().Foreground(subtle).Padding(0, 1)
	bodyStyle   = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(accent).Padding(1, 2)

	titleStyle    = lipgloss.NewStyle().Bold(true).Foreground(accent)
	cursorStyle   = lipgloss.NewStyle().Foreground(accent).Bold(true)
	dimStyle      = lipgloss.NewStyle().Foreground(subtle)
	okStyle       = lipgloss.NewStyle().Foreground(good)
	errStyle      = lipgloss.NewStyle().Foreground(bad)
	selectedStyle = lipgloss.NewStyle().Foreground(accent).Bold(true)
	activeTab     = lipgloss.NewStyle().Foreground(light).Background(accent).Padding(0, 1)
	inactiveTab   = lipgloss.NewStyle().Foreground(subtle).Padding(0, 1)
)

// chrome frames a body between a header bar and a footer/help line, sized to the
// terminal. It keeps every screen visually consistent.
func chrome(width, height int, step, body, footer string) string {
	if width <= 0 {
		return body
	}
	header := lipgloss.JoinHorizontal(
		lipgloss.Top,
		headerStyle.Render("CS2 Demo Highlighter"),
		stepStyle.Render(step),
	)
	header = lipgloss.NewStyle().Width(width).Render(header)
	foot := footerStyle.Width(width).Render(footer)

	bodyHeight := height - lipgloss.Height(header) - lipgloss.Height(foot)
	framed := bodyStyle.Width(width - 2).Height(max(bodyHeight-2, 1)).Render(body)

	return lipgloss.JoinVertical(lipgloss.Left, header, framed, foot)
}
