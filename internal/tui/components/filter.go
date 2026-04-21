package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	filterFieldStyle = lipgloss.NewStyle()
	filterHintStyle  = lipgloss.NewStyle()
)

func RenderFilter(m TUIFilterModel) string {
	var b strings.Builder

	b.WriteString("Filter by tag: ")
	b.WriteString(filterFieldStyle.Render("["))
	b.WriteString(m.TagFilter)
	b.WriteString(filterFieldStyle.Render("]"))
	b.WriteString("\n")
	b.WriteString(filterHintStyle.Render("(tab to autocomplete, enter to confirm)"))

	return b.String()
}
