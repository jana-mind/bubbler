package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	detailTitleStyle = lipgloss.NewStyle().Bold(true)
	detailMetaStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	detailHistStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	detailHintStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
)

func RenderDetail(m TUIDetailModel) string {
	var b strings.Builder

	if m.DetailIssueID == "" {
		b.WriteString("No issue selected.\n[q] back")
		return b.String()
	}

	issue, ok := m.Issues[m.DetailIssueID]
	if !ok {
		b.WriteString("Issue not found.\n[q] back")
		return b.String()
	}

	if issue.Title == "" {
		b.WriteString("Issue not found.\n[q] back")
		return b.String()
	}

	b.WriteString("Issue ")
	b.WriteString(m.DetailIssueID)
	b.WriteString(" -- ")
	b.WriteString(detailTitleStyle.Render(issue.Title))
	b.WriteString("\n\n")

	b.WriteString("Status:   ")
	for _, col := range m.Board.Board.Columns {
		if col.ID == issue.Column {
			b.WriteString(col.Label)
			break
		}
	}
	b.WriteString("\n")

	b.WriteString("Tags:     ")
	b.WriteString(strings.Join(issue.Tags, ", "))
	if len(issue.Tags) == 0 {
		b.WriteString("none")
	}
	b.WriteString("\n")

	b.WriteString("Created:  ")
	b.WriteString(detailMetaStyle.Render(issue.CreatedBy.Name))
	b.WriteString("\n\n")

	b.WriteString("Description:\n")
	if issue.Description != "" {
		for _, line := range strings.Split(issue.Description, "\n") {
			b.WriteString("  ")
			b.WriteString(line)
			b.WriteString("\n")
		}
	} else {
		b.WriteString("  (no description)\n")
	}

	b.WriteString("\nHistory:\n")
	for i := len(issue.History) - 1; i >= 0; i-- {
		entry := issue.History[i]
		b.WriteString("  ")
		b.WriteString(detailHistStyle.Render(entry.At.Format("2006-01-02 15:04")))
		b.WriteString("  ")
		b.WriteString(entry.By.Name)
		b.WriteString("  ")
		b.WriteString(entry.Type)
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(detailHintStyle.Render("[q] back  [e] edit  [m] move  [d] delete"))

	return b.String()
}
