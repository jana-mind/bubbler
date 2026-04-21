package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	confirmWarnStyle = lipgloss.NewStyle().Bold(true)
	confirmHintStyle = lipgloss.NewStyle()
)

func RenderConfirm(detailIssueID string) string {
	var b strings.Builder

	b.WriteString("Delete issue ")
	b.WriteString(confirmWarnStyle.Render(detailIssueID))
	b.WriteString("? This cannot be undone.\n\n")
	b.WriteString(confirmWarnStyle.Render("[enter] confirm  [c] cancel"))

	return b.String()
}
