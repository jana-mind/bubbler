package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/jana-mind/bubbler/internal/model"
)

var (
	colHeadStyle    = lipgloss.NewStyle().Bold(true)
	narrowTermStyle = lipgloss.NewStyle()
)

func RenderBoard(m TUIViewModel, width int) string {
	if width < 60 {
		return narrowTermStyle.Render("Terminal must be at least 60 columns wide. Resize and try again.")
	}

	columns := m.Board.Board.Columns
	if len(columns) == 0 {
		return "No columns defined. Run `bubbler board column add` first.\n"
	}

	colWidth := width / len(columns)
	if colWidth < 20 {
		colWidth = 20
	}

	var b strings.Builder

	for colIdx, col := range columns {
		b.WriteString("+--[ ")
		b.WriteString(colHeadStyle.Render(col.Label))
		b.WriteString(" ]")
		remaining := colWidth - len(col.Label) - 7
		if remaining > 0 {
			b.WriteString(strings.Repeat("-", remaining))
		}
		if colIdx < len(columns)-1 {
			b.WriteString("-+--")
		}
	}
	b.WriteString("+\n")

	colIssues := make([][]model.IssueSummary, len(columns))
	for _, issue := range m.Board.Issues {
		for idx, col := range columns {
			if issue.Column == col.ID {
				if idx < len(colIssues) {
					colIssues[idx] = append(colIssues[idx], issue)
				}
				break
			}
		}
	}

	maxRows := 0
	for _, issues := range colIssues {
		if len(issues) > maxRows {
			maxRows = len(issues)
		}
	}

	for row := 0; row < maxRows; row++ {
		for colIdx, issues := range colIssues {
			b.WriteString("| ")
			if row < len(issues) {
				issue := issues[row]
				prefix := " "
				if colIdx == m.FocusedColumn && row == m.FocusedIssue {
					prefix = ">"
				}
				id := issue.ID
				if len(id) > 6 {
					id = id[:6]
				}
				title := issue.Title
				if len(title) > colWidth-10 {
					title = title[:colWidth-13] + "..."
				}
				b.WriteString(prefix)
				b.WriteString(id)
				b.WriteString(" ")
				b.WriteString(title)
				remaining := colWidth - len(prefix) - len(id) - 1 - len(title) - 1
				if remaining > 0 {
					b.WriteString(strings.Repeat(" ", remaining))
				}
			} else {
				b.WriteString(strings.Repeat(" ", colWidth-2))
			}
			if colIdx < len(colIssues)-1 {
				b.WriteString(" |")
			}
		}
		b.WriteString(" |\n")
	}

	for colIdx := range columns {
		b.WriteString("+")
		b.WriteString(strings.Repeat("-", colWidth))
		if colIdx < len(columns)-1 {
			b.WriteString("-+")
		}
	}
	b.WriteString("+\n")

	b.WriteString("[up/dn] col  [lf/rt] issue  [enter] view  [n] new  [t] filter  [q] quit\n")

	return b.String()
}
