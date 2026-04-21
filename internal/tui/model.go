package tui

import (
	"errors"
	"strings"

	"charm.land/bubbletea/v2"

	"github.com/jana-mind/bubbler/internal/git"
	"github.com/jana-mind/bubbler/internal/model"
	"github.com/jana-mind/bubbler/internal/tui/components"
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

type completionState struct {
	active  bool
	matches []string
	index   int
	input   string
	target  string
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
	completion      completionState
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

	case tea.WindowSizeMsg:
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

	case TabPressed:
		if m.view == viewFilter {
			m.completion.active = true
			m.completion.matches = matchTags(m.tagFilter, m.board.Board.Tags)
			m.completion.index = 0
			m.completion.input = m.tagFilter
			m.completion.target = "filter"
		}
		return m, nil

	case TabCompletionResult:
		m.completion.active = false
		return m, nil

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

func matchTags(input string, available []string) []string {
	if input == "" {
		return available
	}
	var matches []string
	for _, tag := range available {
		if strings.HasPrefix(tag, input) {
			matches = append(matches, tag)
		}
	}
	return matches
}

func (m Model) View() tea.View {
	v := tea.NewView(components.RenderView(components.MainViewModel{
		ViewState:     int(m.view),
		Board:         m.board,
		Issues:        m.issues,
		FocusedColumn: m.focusedColumn,
		FocusedIssue:  m.focusedIssue,
		DetailIssueID: m.detailIssueID,
		ConfirmDelete: m.modal.confirmDelete,
		FormTitle:     m.formTitle,
		FormColumn:    m.formColumn,
		FormTags:      m.formTags,
		FormDescLines: m.formDescLines,
		TagFilter:     m.tagFilter,
		WriteErr:      m.writeErr,
		Loading:       m.loading,
		BoardName:     m.boardName,
		Width:         m.width,
		Height:        m.height,
		Completion: components.CompletionViewModel{
			Active:  m.completion.active,
			Matches: m.completion.matches,
			Index:   m.completion.index,
		},
	}))
	v.AltScreen = true
	return v
}

func Run(boardName string) error {
	store, err := newRealStore()
	if err != nil {
		return err
	}
	model := initialModel(boardName, store)
	program := tea.NewProgram(model,
		tea.WithWindowSize(80, 24),
	)
	_, err = program.Run()
	return err
}
