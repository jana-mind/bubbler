package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	moveArrowStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("13")).Bold(true)
	moveLabelStyle = lipgloss.NewStyle()
	moveHintStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
)

func RenderMove(m TUIMoveModel) string {
	var b strings.Builder

	b.WriteString("Move issue ")
	b.WriteString(m.DetailIssueID)
	b.WriteString("\n")
	b.WriteString("------------------------------------------\n")
	b.WriteString("Target column:\n\n")

	for idx, col := range m.Board.Board.Columns {
		if idx == m.FormColumn {
			b.WriteString(moveArrowStyle.Render("-> "))
		} else {
			b.WriteString("   ")
		}
		b.WriteString(moveLabelStyle.Render(col.Label))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(moveHintStyle.Render("[enter] confirm  [esc/c] cancel"))

	return b.String()
}
