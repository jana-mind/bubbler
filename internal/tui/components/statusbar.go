package components

import (
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	statusbarStyle = lipgloss.NewStyle().MarginTop(1)
)

func RenderStatusbar(boardName string, issueCount int, tagFilter string) string {
	var b strings.Builder

	b.WriteString(statusbarStyle.Render(boardName))
	b.WriteString(" | ")
	b.WriteString(statusbarStyle.Render(formatIssueCount(issueCount)))
	if tagFilter != "" {
		b.WriteString(" | ")
		b.WriteString(statusbarStyle.Render("filter: "))
		b.WriteString(statusbarStyle.Render(tagFilter))
	}

	return b.String()
}

func formatIssueCount(n int) string {
	if n == 1 {
		return "1 issue"
	}
	return strings.TrimSpace(strconv.Itoa(n)) + " issues"
}
