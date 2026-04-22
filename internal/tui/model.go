package tui

import (
	"errors"
	"path/filepath"
	"strconv"
	"strings"
	"time"

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

type pendingIssueCreate struct {
	boardName string
	issue     model.IssueFile
	entries   []model.HistoryEntry
}

type pendingIssueEdit struct {
	boardName string
	issueID   string
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
	tagInput  string

	loading      bool
	writing      bool
	writeErr     error
	pendingWrite pendingWriteOp

	width  int
	height int

	repoRoot    string
	gitIdentity model.Identity

	store Store
}

func initialModel(boardName string, store Store) Model {
	return Model{
		boardName:     boardName,
		issues:        make(map[string]model.IssueFile),
		view:          viewBoard,
		focusedColumn: 0,
		focusedIssue:  0,
		repoRoot:      store.RepoRoot(),
		store:         store,
	}
}

func (m Model) Init() tea.Cmd {
	repoRoot, err := git.FindRepoRoot()
	if err != nil {
		return func() tea.Msg { return BoardLoadFailed{Err: err} }
	}
	identity, err := git.GetIdentity(repoRoot)
	if err != nil {
		return func() tea.Msg {
			return BoardLoadFailed{Err: errors.New("git identity not configured: run `git config --global user.name` and `git config --global user.email`")}
		}
	}
	m.repoRoot = repoRoot
	m.gitIdentity = model.Identity{Name: identity.Name, Email: identity.Email}
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
		if m.board.Board.NextID == 0 {
			maxID := 0
			for _, iss := range m.board.Issues {
				if n, err := strconv.Atoi(iss.ID); err == nil && n > maxID {
					maxID = n
				}
			}
			m.board.Board.NextID = maxID + 1
		}
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

	case TagInputChanged:
		m.tagInput = t.Text
		m.completion.active = true
		m.completion.matches = matchTags(t.Text, m.board.Board.Tags)
		m.completion.index = 0
		m.completion.input = t.Text
		m.completion.target = "tag"
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
		if !m.writing {
			return m, nil
		}
		m.writing = false
		var commitMsg string
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
				commitMsg = "bubbler: delete issue " + op.issueID
			}
		case pendingIssueCreate:
			if t.Err == nil {
				m.view = viewBoard
				board, err := m.store.LoadBoard(m.boardName)
				if err == nil {
					m.board = board
				}
				commitMsg = "bubbler: create issue " + op.issue.ID
			}
		case pendingIssueEdit:
			if t.Err == nil {
				m.issues[op.issueID] = op.issue
				m.view = viewDetail
				board, err := m.store.LoadBoard(m.boardName)
				if err == nil {
					m.board = board
				}
				commitMsg = "bubbler: edit issue " + op.issueID
			}
		case pendingIssueSave:
			if t.Err == nil {
				m.view = viewDetail
				board, err := m.store.LoadBoard(m.boardName)
				if err == nil {
					m.board = board
				}
				commitMsg = "bubbler: move " + op.issue.ID + " to " + op.issue.Column
			}
		}
		m.writeErr = t.Err
		m.pendingWrite = nil
		if commitMsg != "" && t.Err == nil {
			bubblePath := filepath.Join(m.repoRoot, ".bubble")
			return m, cmdCommitAndPush(m.repoRoot, bubblePath, commitMsg)
		}
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
		case pendingIssueCreate:
			return m, cmdSaveIssue(op.boardName, op.issue, op.entries, m.store)
		case pendingIssueEdit:
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
		} else if m.view == viewCreate || m.view == viewEdit {
			m.completion.active = true
			m.completion.matches = matchTags("", m.board.Board.Tags)
			m.completion.index = 0
			m.completion.input = ""
			m.completion.target = "tag"
		}
		return m, nil

	case TabCompletionResult:
		m.completion.active = false
		return m, nil

	case CreateSubmit:
		if m.formTitle == "" {
			m.writeErr = errors.New("title cannot be empty")
			return m, nil
		}
		newID := m.board.Board.TakeNextID()
		if err := m.store.SaveBoard(m.boardName, m.board); err != nil {
			m.writeErr = err
			return m, nil
		}
		now := time.Now()
		col := m.board.Board.Columns[m.formColumn]
		entry := model.HistoryEntry{
			Type: "created",
			At:   now,
			By:   m.gitIdentity,
			Data: model.CreatedEntry{
				Title:  m.formTitle,
				Column: col.ID,
				Tags:   m.formTags,
			},
		}
		issue := model.IssueFile{
			ID:          newID,
			Title:       m.formTitle,
			Column:      col.ID,
			Tags:        m.formTags,
			Description: strings.Join(m.formDescLines, "\n"),
			CreatedAt:   now,
			CreatedBy:   m.gitIdentity,
			History:     nil,
		}
		m.writing = true
		m.pendingWrite = pendingIssueCreate{boardName: m.boardName, issue: issue, entries: []model.HistoryEntry{entry}}
		return m, cmdSaveIssue(m.boardName, issue, []model.HistoryEntry{entry}, m.store)

	case MoveSubmit:
		issue, ok := m.issues[m.detailIssueID]
		if !ok {
			f, err := m.store.LoadIssue(m.boardName, m.detailIssueID)
			if err != nil {
				m.writeErr = err
				return m, nil
			}
			issue = f
		}
		oldColumn := issue.Column
		newColumn := m.board.Board.Columns[m.formColumn].ID
		now := time.Now()
		entry := model.HistoryEntry{
			Type: "column_changed",
			At:   now,
			By:   m.gitIdentity,
			Data: model.ColumnChangedEntry{From: oldColumn, To: newColumn},
		}
		issue.Column = newColumn
		m.writing = true
		m.pendingWrite = pendingIssueSave{boardName: m.boardName, issue: issue, entries: []model.HistoryEntry{entry}}
		return m, cmdSaveIssue(m.boardName, issue, []model.HistoryEntry{entry}, m.store)

	case EditSave:
		if m.detailIssueID == "" {
			return m, nil
		}
		issue, ok := m.issues[m.detailIssueID]
		if !ok {
			f, err := m.store.LoadIssue(m.boardName, m.detailIssueID)
			if err != nil {
				m.writeErr = err
				return m, nil
			}
			issue = f
		}
		var entries []model.HistoryEntry
		newDesc := strings.Join(m.formDescLines, "\n")
		if m.formTitle != issue.Title {
			entries = append(entries, model.HistoryEntry{
				Type: "title_changed",
				At:   time.Now(),
				By:   m.gitIdentity,
				Data: model.TitleChangedEntry{From: issue.Title, To: m.formTitle},
			})
		}
		if newDesc != issue.Description {
			entries = append(entries, model.HistoryEntry{
				Type: "description_changed",
				At:   time.Now(),
				By:   m.gitIdentity,
				Data: model.DescriptionChangedEntry{Description: newDesc},
			})
		}
		if !tagSetsEqual(m.formTags, issue.Tags) {
			added, removed := tagDiff(issue.Tags, m.formTags)
			entries = append(entries, model.HistoryEntry{
				Type: "tags_changed",
				At:   time.Now(),
				By:   m.gitIdentity,
				Data: model.TagsChangedEntry{Added: added, Removed: removed},
			})
		}
		issue.Title = m.formTitle
		issue.Description = newDesc
		issue.Tags = m.formTags
		if len(entries) == 0 {
			m.view = viewDetail
			return m, nil
		}
		m.writing = true
		m.pendingWrite = pendingIssueEdit{boardName: m.boardName, issueID: m.detailIssueID, issue: issue, entries: entries}
		return m, cmdSaveIssue(m.boardName, issue, entries, m.store)

	case tea.KeyMsg:
		if m.writing {
			return m, nil
		}

		if m.modal.confirmDelete {
			switch t.String() {
			case "y", "Y":
				m.modal.confirmDelete = false
				m.writing = true
				m.pendingWrite = pendingIssueDelete{boardName: m.boardName, issueID: m.detailIssueID}
				return m, cmdDeleteIssue(m.boardName, m.detailIssueID, m.store)
			case "n", "N", "c", "C", "esc":
				m.modal.confirmDelete = false
				return m, nil
			}
			return m, nil
		}

		if m.completion.active {
			switch t.String() {
			case "enter":
				if len(m.completion.matches) > 0 {
					selected := m.completion.matches[m.completion.index]
					if m.completion.target == "filter" {
						m.tagFilter = selected
					} else {
						m.formTags = append(m.formTags, selected)
					}
				} else if m.completion.target == "tag" && m.tagInput != "" {
					if !contains(m.formTags, m.tagInput) {
						m.formTags = append(m.formTags, m.tagInput)
					}
				}
				m.completion.active = false
				m.tagInput = ""
				return m, nil
			case "tab", "down":
				if len(m.completion.matches) > 0 {
					m.completion.index = (m.completion.index + 1) % len(m.completion.matches)
				}
				return m, nil
			case "up":
				if len(m.completion.matches) > 0 {
					m.completion.index--
					if m.completion.index < 0 {
						m.completion.index = len(m.completion.matches) - 1
					}
				}
				return m, nil
			case "esc", "c", "C":
				m.completion.active = false
				m.tagInput = ""
				return m, nil
			}
		}

		switch m.view {
		case viewBoard:
			switch t.String() {
			case "right", "l", "L":
				if m.focusedColumn < len(m.board.Board.Columns)-1 {
					m.focusedColumn++
					m.focusedIssue = 0
				}
				return m, nil
			case "left", "h", "H":
				if m.focusedColumn > 0 {
					m.focusedColumn--
					m.focusedIssue = 0
				}
				return m, nil
			case "down", "j", "J":
				m.focusedIssue++
				return m, nil
			case "up", "k", "K":
				if m.focusedIssue > 0 {
					m.focusedIssue--
				}
				return m, nil
			case "enter":
				colIssues := issuesInColumn(m.board.Board.Columns[m.focusedColumn].ID, m.board.Issues)
				if m.focusedIssue < len(colIssues) {
					issueID := colIssues[m.focusedIssue].ID
					m.detailIssueID = issueID
					if _, ok := m.issues[issueID]; !ok {
						f, err := m.store.LoadIssue(m.boardName, issueID)
						if err == nil {
							m.issues[issueID] = f
						}
					}
					m.view = viewDetail
				}
				return m, nil
			case "n", "N":
				m.view = viewCreate
				m.formTitle = ""
				m.formColumn = m.focusedColumn
				if m.formColumn >= len(m.board.Board.Columns) {
					m.formColumn = 0
				}
				m.formTags = nil
				m.formDescLines = nil
				m.formDescEditing = false
				m.tagInput = ""
				m.completion.active = false
				return m, nil
			case "t", "T":
				m.view = viewFilter
				m.tagFilter = ""
				m.completion.active = false
				return m, nil
			case "q", "Q":
				return m, tea.Quit
			}

		case viewDetail:
			switch t.String() {
			case "q", "Q", "esc":
				m.view = viewBoard
				m.detailIssueID = ""
				return m, nil
			case "e", "E":
				if m.detailIssueID == "" {
					return m, nil
				}
				issue, ok := m.issues[m.detailIssueID]
				if ok {
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
				m.view = viewEdit
				return m, nil
			case "m", "M":
				if m.detailIssueID == "" {
					return m, nil
				}
				issue, ok := m.issues[m.detailIssueID]
				if ok {
					for i, col := range m.board.Board.Columns {
						if col.ID == issue.Column {
							m.formColumn = i
							break
						}
					}
				}
				m.view = viewMove
				return m, nil
			case "d", "D":
				m.modal.confirmDelete = true
				return m, nil
			}

		case viewCreate:
			switch t.String() {
			case "enter":
				return m, func() tea.Msg { return CreateSubmit{} }
			case "esc", "c", "C":
				m.view = viewBoard
				return m, nil
			case "up":
				m.formColumn--
				if m.formColumn < 0 {
					m.formColumn = len(m.board.Board.Columns) - 1
				}
				return m, nil
			case "down":
				m.formColumn = (m.formColumn + 1) % len(m.board.Board.Columns)
				return m, nil
			}

		case viewMove:
			switch t.String() {
			case "enter":
				return m, func() tea.Msg { return MoveSubmit{} }
			case "esc", "c", "C":
				m.view = viewDetail
				return m, nil
			case "up":
				m.formColumn--
				if m.formColumn < 0 {
					m.formColumn = len(m.board.Board.Columns) - 1
				}
				return m, nil
			case "down":
				m.formColumn = (m.formColumn + 1) % len(m.board.Board.Columns)
				return m, nil
			}

		case viewEdit:
			switch t.String() {
			case "enter":
				return m, func() tea.Msg { return EditSave{} }
			case "esc", "c", "C":
				m.view = viewDetail
				return m, nil
			}

		case viewFilter:
			switch t.String() {
			case "enter":
				m.view = viewBoard
				return m, nil
			case "esc", "c", "C":
				m.tagFilter = ""
				m.view = viewBoard
				return m, nil
			case "tab":
				m.completion.active = true
				m.completion.matches = matchTags(m.tagFilter, m.board.Board.Tags)
				m.completion.index = 0
				m.completion.input = m.tagFilter
				m.completion.target = "filter"
				return m, nil
			}
		}

		return m, nil

	default:
		return m, nil
	}
}

func issuesInColumn(colID string, issues []model.IssueSummary) []model.IssueSummary {
	var result []model.IssueSummary
	for _, iss := range issues {
		if iss.Column == colID {
			result = append(result, iss)
		}
	}
	return result
}

func contains(haystack []string, needle string) bool {
	for _, s := range haystack {
		if s == needle {
			return true
		}
	}
	return false
}

func tagSetsEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	set := make(map[string]struct{}, len(a))
	for _, t := range a {
		set[t] = struct{}{}
	}
	for _, t := range b {
		if _, ok := set[t]; !ok {
			return false
		}
	}
	return true
}

func tagDiff(old, new []string) (added, removed []string) {
	oldSet := make(map[string]struct{}, len(old))
	for _, t := range old {
		oldSet[t] = struct{}{}
	}
	newSet := make(map[string]struct{}, len(new))
	for _, t := range new {
		newSet[t] = struct{}{}
		if _, ok := oldSet[t]; !ok {
			added = append(added, t)
		}
	}
	for _, t := range old {
		if _, ok := newSet[t]; !ok {
			removed = append(removed, t)
		}
	}
	return
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
		TagInput:      m.tagInput,
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
