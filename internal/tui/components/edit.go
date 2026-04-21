package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	editLabelStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	editFieldStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("13"))
	editHintStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
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

	b.WriteString("\n")
	b.WriteString(editHintStyle.Render("[enter] save  [esc/c] cancel"))

	return b.String()
}
