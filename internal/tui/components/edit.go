package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	editLabelStyle = lipgloss.NewStyle()
	editFieldStyle = lipgloss.NewStyle()
	editHintStyle  = lipgloss.NewStyle()
)

func RenderEdit(m TUIEditModel) string {
	var b strings.Builder

	b.WriteString("Edit issue ")
	b.WriteString(m.DetailIssueID)
	b.WriteString("\n")
	b.WriteString("------------------------------------------\n")
	b.WriteString("Title:    ")
	b.WriteString(editFieldStyle.Render("["))
	b.WriteString(m.FormTitle)
	b.WriteString(editFieldStyle.Render("]"))
	b.WriteString("\n")

	b.WriteString("Description:\n")
	for _, line := range m.FormDescLines {
		b.WriteString(editFieldStyle.Render("["))
		b.WriteString(line)
		b.WriteString(editFieldStyle.Render("]"))
		b.WriteString("\n")
	}
	b.WriteString(editHintStyle.Render("  (enter for newline, ctrl+d to finish)\n"))

	b.WriteString("Tags:     ")
	b.WriteString(editFieldStyle.Render("["))
	b.WriteString(strings.Join(m.FormTags, ", "))
	b.WriteString(editFieldStyle.Render("]"))
	b.WriteString("\n")

	if m.Completion.Active && len(m.Completion.Matches) > 0 {
		b.WriteString("\n")
		for i, match := range m.Completion.Matches {
			if i == m.Completion.Index {
				b.WriteString(editFieldStyle.Render("> "))
				b.WriteString(match)
			} else {
				b.WriteString("  ")
				b.WriteString(match)
			}
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(editHintStyle.Render("[enter] save  [esc/c] cancel"))

	return b.String()
}
