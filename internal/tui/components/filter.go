package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	filterFieldStyle   = lipgloss.NewStyle()
	filterHintStyle    = lipgloss.NewStyle()
	filterMatchStyle   = lipgloss.NewStyle().Bold(true)
	filterSelectedStyle = lipgloss.NewStyle()
)

func RenderFilter(m TUIFilterModel, completion CompletionViewModel) string {
	var b strings.Builder

	b.WriteString("Filter by tag: ")
	b.WriteString(filterFieldStyle.Render("["))
	b.WriteString(m.TagFilter)
	b.WriteString(filterFieldStyle.Render("]"))
	b.WriteString("\n")
	b.WriteString(filterHintStyle.Render("(tab to autocomplete, enter to confirm)"))
	b.WriteString("\n")

	if completion.Active && len(completion.Matches) > 0 {
		b.WriteString("\n")
		for i, match := range completion.Matches {
			if i == completion.Index {
				b.WriteString(filterSelectedStyle.Render("> "))
				b.WriteString(filterMatchStyle.Render(match))
			} else {
				b.WriteString("  ")
				b.WriteString(match)
			}
			b.WriteString("\n")
		}
	}

	return b.String()
}
