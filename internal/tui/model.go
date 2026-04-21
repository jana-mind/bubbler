package tui

import (
	"errors"
	"strings"

	"charm.land/bubbletea/v2"

	"github.com/jana-mind/bubbler/internal/git"
	"github.com/jana-mind/bubbler/internal/model"
)

type viewState int

const (
	viewBoard viewState = iota
	viewDetail
	viewCreate
	viewMove
	viewEdit
	viewFilter
)

type modalState struct {
	confirmDelete bool
}

type pendingWriteOp interface{}

type pendingIssueSave struct {
	boardName string
	issue     model.IssueFile
	entries   []model.HistoryEntry
}

type pendingIssueDelete struct {
	boardName string
	issueID   string
}

type Model struct {
	boardName string
	board     model.BoardFile
	issues    map[string]model.IssueFile

	focusedColumn int
	focusedIssue  int

	view          viewState
	detailIssueID string

	modal           modalState
	formTitle       string
	formColumn      int
	formTags        []string
	formDescLines   []string
	formDescEditing bool

	tagFilter string

	loading      bool
	writing      bool
	writeErr     error
	pendingWrite pendingWriteOp

	width  int
	height int

	store Store
}

func initialModel(boardName string, store Store) Model {
	return Model{
		boardName:     boardName,
		issues:        make(map[string]model.IssueFile),
		view:          viewBoard,
		focusedColumn: 0,
		focusedIssue:  0,
		store:         store,
	}
}

func (m Model) Init() tea.Cmd {
	repoRoot, err := git.FindRepoRoot()
	if err != nil {
		return func() tea.Msg { return BoardLoadFailed{Err: err} }
	}
	_, err = git.GetIdentity(repoRoot)
	if err != nil {
		return func() tea.Msg {
			return BoardLoadFailed{Err: errors.New("git identity not configured: run `git config --global user.name` and `git config --global user.email`")}
		}
	}
	board, err := m.store.LoadBoard(m.boardName)
	if err != nil {
		return func() tea.Msg { return BoardLoadFailed{Err: err} }
	}
	return func() tea.Msg { return BoardLoaded{Board: board} }
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch t := msg.(type) {
	case BoardLoaded:
		m.board = t.Board
		m.issues = make(map[string]model.IssueFile)
		m.loading = false
		return m, nil

	case BoardLoadFailed:
		m.loading = false
		return m, nil

	case IssueFocused:
		if _, ok := m.issues[t.IssueID]; !ok {
			f, err := m.store.LoadIssue(m.boardName, t.IssueID)
			if err == nil {
				m.issues[t.IssueID] = f
			}
		}
		return m, nil

	case ColumnFocused:
		m.focusedColumn = t.Index
		m.focusedIssue = 0
		return m, nil

	case OpenCreateModal:
		m.view = viewCreate
		m.formTitle = ""
		m.formColumn = 0
		m.formTags = nil
		m.formDescLines = nil
		m.formDescEditing = false
		return m, nil

	case OpenMoveModal:
		m.view = viewMove
		return m, nil

	case OpenEditModal:
		m.view = viewEdit
		if issue, ok := m.issues[t.IssueID]; ok {
			m.formTitle = issue.Title
			m.formDescLines = parseDescription(issue.Description)
			for i, col := range m.board.Board.Columns {
				if col.ID == issue.Column {
					m.formColumn = i
					break
				}
			}
			m.formTags = issue.Tags
		}
		return m, nil

	case FormTitleChanged:
		m.formTitle = t.Text
		return m, nil

	case FormColumnChanged:
		m.formColumn = t.Index
		return m, nil

	case FormTagsChanged:
		m.formTags = t.Tags
		return m, nil

	case FormDescLineAdded:
		m.formDescLines = append(m.formDescLines, t.Line)
		return m, nil

	case FormDescConfirmed:
		m.formDescEditing = false
		return m, nil

	case ConfirmDelete:
		m.modal.confirmDelete = true
		return m, nil

	case ConfirmDeleteConfirmed:
		m.modal.confirmDelete = false
		m.writing = true
		m.pendingWrite = pendingIssueDelete{boardName: m.boardName, issueID: m.detailIssueID}
		return m, cmdDeleteIssue(m.boardName, m.detailIssueID, m.store)

	case ConfirmDeleteCancelled:
		m.modal.confirmDelete = false
		return m, nil

	case TagFilterApplied:
		m.tagFilter = t.Tag
		return m, nil

	case TagFilterCleared:
		m.tagFilter = ""
		return m, nil

	case RefreshRequested:
		board, err := m.store.LoadBoard(m.boardName)
		if err != nil {
			m.writeErr = err
			return m, nil
		}
		m.board = board
		m.issues = make(map[string]model.IssueFile)
		return m, nil

	case WindowResized:
		m.width = t.Width
		m.height = t.Height
		return m, nil

	case WriteCompleted:
		m.writing = false
		switch op := m.pendingWrite.(type) {
		case pendingIssueDelete:
			if t.Err == nil {
				delete(m.issues, op.issueID)
				m.detailIssueID = ""
				m.view = viewBoard
				board, err := m.store.LoadBoard(m.boardName)
				if err == nil {
					m.board = board
				}
			}
		case pendingIssueSave:
			if t.Err == nil {
				board, err := m.store.LoadBoard(m.boardName)
				if err == nil {
					m.board = board
				}
			}
		}
		m.writeErr = t.Err
		m.pendingWrite = nil
		return m, nil

	case WriteRetryRequested:
		if m.pendingWrite == nil {
			return m, nil
		}
		m.writing = true
		switch op := m.pendingWrite.(type) {
		case pendingIssueDelete:
			return m, cmdDeleteIssue(op.boardName, op.issueID, m.store)
		case pendingIssueSave:
			return m, cmdSaveIssue(op.boardName, op.issue, op.entries, m.store)
		}
		return m, nil

	case WriteCancelled:
		m.writing = false
		m.writeErr = nil
		if m.view == viewCreate || m.view == viewMove || m.view == viewEdit {
			m.view = viewBoard
		}
		return m, nil

	case QuitConfirmed:
		return m, tea.Quit

	default:
		return m, nil
	}
}

func parseDescription(desc string) []string {
	if desc == "" {
		return nil
	}
	lines := make([]string, 0)
	start := 0
	for i := 0; i < len(desc); i++ {
		if desc[i] == '\n' {
			lines = append(lines, desc[start:i])
			start = i + 1
		}
	}
	if start < len(desc) {
		lines = append(lines, desc[start:])
	}
	return lines
}

func (m Model) View() tea.View {
	v := tea.NewView(renderView(m))
	v.AltScreen = true
	return v
}

func renderView(m Model) string {
	if m.loading {
		return "Loading...\n"
	}
	if m.boardName == "" {
		return "No board found. Run `bubbler init` first.\nPress any key to exit.\n"
	}
	switch m.view {
	case viewBoard:
		return renderBoard(m)
	case viewDetail:
		return renderDetail(m)
	case viewCreate:
		return renderCreate(m)
	case viewMove:
		return renderMove(m)
	case viewEdit:
		return renderEdit(m)
	case viewFilter:
		return renderFilter(m)
	default:
		return "Unknown view\n"
	}
}

func Run(boardName string) error {
	store, err := newRealStore()
	if err != nil {
		return err
	}
	model := initialModel(boardName, store)
	program := tea.NewProgram(model)
	_, err = program.Run()
	return err
}

func renderBoard(m Model) string {
	if m.width < 60 {
		return "Terminal must be at least 60 columns wide. Resize and try again.\n"
	}

	columns := m.board.Board.Columns
	if len(columns) == 0 {
		return "No columns defined. Run `bubbler board column add` first.\n"
	}

	colWidth := m.width / len(columns)
	if colWidth < 20 {
		colWidth = 20
	}

	var b strings.Builder

	for colIdx, col := range columns {
		b.WriteString("+--[ ")
		b.WriteString(col.Label)
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
	for _, issue := range m.board.Issues {
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
				if colIdx == m.focusedColumn && row == m.focusedIssue {
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

func renderDetail(m Model) string {
	var b strings.Builder

	if m.detailIssueID == "" {
		b.WriteString("No issue selected.\n[q] back")
		return b.String()
	}

	issue, ok := m.issues[m.detailIssueID]
	if !ok {
		f, err := m.store.LoadIssue(m.boardName, m.detailIssueID)
		if err == nil {
			issue = f
			m.issues[m.detailIssueID] = f
		}
	}

	if issue.Title == "" {
		b.WriteString("Issue not found.\n[q] back")
		return b.String()
	}

	b.WriteString("Issue ")
	b.WriteString(m.detailIssueID)
	b.WriteString(" -- ")
	b.WriteString(issue.Title)
	b.WriteString("\n\n")

	b.WriteString("Status:   ")
	for _, col := range m.board.Board.Columns {
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
	b.WriteString(issue.CreatedBy.Name)
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
		b.WriteString(entry.At.Format("2006-01-02 15:04"))
		b.WriteString("  ")
		b.WriteString(entry.By.Name)
		b.WriteString("  ")
		b.WriteString(entry.Type)
		b.WriteString("\n")
	}

	b.WriteString("\n[q] back  [e] edit  [m] move  [d] delete\n")

	return b.String()
}

func renderCreate(m Model) string {
	var b strings.Builder

	b.WriteString("Create new issue\n")
	b.WriteString("------------------------------------------\n")
	b.WriteString("Title:    [")
	b.WriteString(m.formTitle)
	b.WriteString("]\n")

	b.WriteString("Column:   ")
	if len(m.board.Board.Columns) > m.formColumn {
		b.WriteString(m.board.Board.Columns[m.formColumn].Label)
	}
	b.WriteString("  (up/dn to change)\n")

	b.WriteString("Tags:     [")
	b.WriteString(strings.Join(m.formTags, ", "))
	b.WriteString("]\n")

	b.WriteString("\n[enter] create  [esc/c] cancel\n")

	return b.String()
}

func renderMove(m Model) string {
	var b strings.Builder

	b.WriteString("Move issue ")
	b.WriteString(m.detailIssueID)
	b.WriteString("\n")
	b.WriteString("------------------------------------------\n")
	b.WriteString("Target column:\n\n")

	for idx, col := range m.board.Board.Columns {
		arrow := "  "
		if idx == m.formColumn {
			arrow = "-> "
		}
		b.WriteString(arrow)
		b.WriteString(col.Label)
		b.WriteString("\n")
	}

	b.WriteString("\n[enter] confirm  [esc/c] cancel\n")

	return b.String()
}

func renderEdit(m Model) string {
	var b strings.Builder

	b.WriteString("Edit issue ")
	b.WriteString(m.detailIssueID)
	b.WriteString("\n")
	b.WriteString("------------------------------------------\n")
	b.WriteString("Title:    [")
	b.WriteString(m.formTitle)
	b.WriteString("]\n")

	b.WriteString("Description:\n")
	for _, line := range m.formDescLines {
		b.WriteString("[")
		b.WriteString(line)
		b.WriteString("]\n")
	}
	b.WriteString("  (enter for newline, ctrl+d to finish)\n")

	b.WriteString("Tags:     [")
	b.WriteString(strings.Join(m.formTags, ", "))
	b.WriteString("]\n")

	b.WriteString("\n[enter] save  [esc/c] cancel\n")

	return b.String()
}

func renderFilter(m Model) string {
	var b strings.Builder

	b.WriteString("Filter by tag: [")
	b.WriteString(m.tagFilter)
	b.WriteString("]\n")
	b.WriteString("(tab to autocomplete, enter to confirm)\n")

	return b.String()
}
