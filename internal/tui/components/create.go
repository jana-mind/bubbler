package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	createLabelStyle = lipgloss.NewStyle()
	createFieldStyle = lipgloss.NewStyle()
	createHintStyle  = lipgloss.NewStyle()
)

func RenderCreate(m TUICreateModel) string {
	var b strings.Builder

	b.WriteString("Create new issue\n")
	b.WriteString("------------------------------------------\n")
	b.WriteString("Title:    ")
	b.WriteString(createFieldStyle.Render("["))
	b.WriteString(m.FormTitle)
	b.WriteString(createFieldStyle.Render("]"))
	b.WriteString("\n")

	b.WriteString("Column:   ")
	if len(m.Board.Board.Columns) > m.FormColumn {
		b.WriteString(m.Board.Board.Columns[m.FormColumn].Label)
	}
	b.WriteString("  (up/dn to change)\n")

	b.WriteString("Tags:     ")
	b.WriteString(createFieldStyle.Render("["))
	b.WriteString(strings.Join(m.FormTags, ", "))
	b.WriteString(createFieldStyle.Render("]"))
	b.WriteString("\n")

	b.WriteString("\n")
	b.WriteString(createHintStyle.Render("[enter] create  [esc/c] cancel"))

	return b.String()
}
