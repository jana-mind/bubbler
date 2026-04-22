package tui

import (
	"testing"

	"charm.land/bubbletea/v2"

	"github.com/jana-mind/bubbler/internal/model"
)

func TestViewTransitions(t *testing.T) {
	t.Run("board to create modal via OpenCreateModal", func(t *testing.T) {
		store := &mockStore{}
		m := initialModel("default", store)
		m.view = viewBoard

		result, _ := m.Update(OpenCreateModal{})
		m = result.(Model)
		if m.view != viewCreate {
			t.Errorf("expected viewCreate, got %v", m.view)
		}
		if m.formTitle != "" {
			t.Errorf("expected empty formTitle, got %q", m.formTitle)
		}
		if m.formColumn != 0 {
			t.Errorf("expected formColumn 0, got %d", m.formColumn)
		}
	})

	t.Run("board to move modal via OpenMoveModal", func(t *testing.T) {
		store := &mockStore{}
		m := initialModel("default", store)
		m.view = viewBoard

		result, _ := m.Update(OpenMoveModal{IssueID: "issue1"})
		m = result.(Model)
		if m.view != viewMove {
			t.Errorf("expected viewMove, got %v", m.view)
		}
	})

	t.Run("board to edit modal via OpenEditModal", func(t *testing.T) {
		store := &mockStore{}
		issue := model.IssueFile{
			ID:     "issue1",
			Title:  "Test Issue",
			Column: "col1",
		}
		m := initialModel("default", store)
		m.view = viewBoard
		m.issues["issue1"] = issue
		m.board = model.BoardFile{
			Board: model.Board{
				Columns: []model.Column{
					{ID: "col1", Label: "Col 1"},
				},
				Tags: []string{},
			},
		}

		result, _ := m.Update(OpenEditModal{IssueID: "issue1"})
		m = result.(Model)
		if m.view != viewEdit {
			t.Errorf("expected viewEdit, got %v", m.view)
		}
		if m.formTitle != "Test Issue" {
			t.Errorf("expected formTitle 'Test Issue', got %q", m.formTitle)
		}
	})

	t.Run("TagFilterApplied sets tagFilter", func(t *testing.T) {
		store := &mockStore{}
		m := initialModel("default", store)
		m.view = viewBoard

		result, _ := m.Update(TagFilterApplied{Tag: "bug"})
		m = result.(Model)
		if m.tagFilter != "bug" {
			t.Errorf("expected tagFilter 'bug', got %q", m.tagFilter)
		}
	})

	t.Run("TagFilterCleared clears tagFilter", func(t *testing.T) {
		store := &mockStore{}
		m := initialModel("default", store)
		m.view = viewBoard
		m.tagFilter = "bug"

		result, _ := m.Update(TagFilterCleared{})
		m = result.(Model)
		if m.tagFilter != "" {
			t.Errorf("expected empty tagFilter, got %q", m.tagFilter)
		}
	})

	t.Run("confirm delete sets modal.confirmDelete", func(t *testing.T) {
		store := &mockStore{}
		m := initialModel("default", store)
		m.modal.confirmDelete = false

		result, _ := m.Update(ConfirmDelete{IssueID: "issue1"})
		m = result.(Model)
		if !m.modal.confirmDelete {
			t.Error("expected confirmDelete to be true")
		}
	})

	t.Run("confirm delete cancelled clears modal", func(t *testing.T) {
		store := &mockStore{}
		m := initialModel("default", store)
		m.modal.confirmDelete = true

		result, _ := m.Update(ConfirmDeleteCancelled{})
		m = result.(Model)
		if m.modal.confirmDelete {
			t.Error("expected confirmDelete to be false")
		}
	})

	t.Run("confirm delete confirmed triggers write", func(t *testing.T) {
		store := &mockStore{}
		m := initialModel("default", store)
		m.modal.confirmDelete = true
		m.detailIssueID = "issue1"

		result, _ := m.Update(ConfirmDeleteConfirmed{})
		m = result.(Model)
		if !m.writing {
			t.Error("expected writing to be true")
		}
		if m.pendingWrite == nil {
			t.Error("expected pendingWrite to be set")
		}
	})
}

func TestLazyIssueLoading(t *testing.T) {
	store := &mockStore{}

	m := initialModel("default", store)

	if _, ok := m.issues["issue1"]; ok {
		t.Error("issue should not be loaded until IssueFocused")
	}

	result, _ := m.Update(IssueFocused{IssueID: "issue1"})
	m = result.(Model)
	if _, ok := m.issues["issue1"]; !ok {
		t.Error("issue should be loaded after IssueFocused")
	}
}

func TestWriteCompletion(t *testing.T) {
	t.Run("WriteCompleted clears writing and updates board on pendingIssueSave", func(t *testing.T) {
		store := &mockStore{}
		m := initialModel("default", store)
		m.writing = true
		m.pendingWrite = pendingIssueSave{
			boardName: "default",
			issue:     model.IssueFile{ID: "issue1"},
			entries:   nil,
		}

		result, _ := m.Update(WriteCompleted{Err: nil})
		m = result.(Model)
		if m.writing {
			t.Error("expected writing to be false")
		}
		if m.pendingWrite != nil {
			t.Error("expected pendingWrite to be cleared")
		}
	})

	t.Run("WriteCompleted sets writeErr on error", func(t *testing.T) {
		store := &mockStore{}
		m := initialModel("default", store)
		m.writing = true
		m.writeErr = nil

		result, _ := m.Update(WriteCompleted{Err: errTest})
		m = result.(Model)
		if m.writeErr == nil {
			t.Error("expected writeErr to be set")
		}
		if m.writing {
			t.Error("expected writing to be false after WriteCompleted")
		}
	})

	t.Run("WriteCancelled clears writing and returns to board", func(t *testing.T) {
		store := &mockStore{}
		m := initialModel("default", store)
		m.writing = true
		m.view = viewCreate
		m.writeErr = errTest

		result, _ := m.Update(WriteCancelled{})
		m = result.(Model)
		if m.writing {
			t.Error("expected writing to be false")
		}
		if m.view != viewBoard {
			t.Errorf("expected viewBoard, got %v", m.view)
		}
	})

	t.Run("WriteRetryRequested retries pending write", func(t *testing.T) {
		store := &mockStore{}
		m := initialModel("default", store)
		m.writing = false
		m.pendingWrite = pendingIssueSave{
			boardName: "default",
			issue:     model.IssueFile{ID: "issue1"},
			entries:   nil,
		}

		_, cmd := m.Update(WriteRetryRequested{})
		if cmd == nil {
			t.Error("expected cmd to be returned for retry")
		}
	})
}

func TestRefresh(t *testing.T) {
	t.Run("RefreshRequested clears issues and returns cmd", func(t *testing.T) {
		store := &mockStore{}
		m := initialModel("default", store)
		m.board = model.BoardFile{Board: model.Board{Name: "old"}}

		result, cmd := m.Update(RefreshRequested{})
		m = result.(Model)
		if cmd == nil {
			t.Error("expected cmd to be returned")
		}
		if len(m.issues) != 0 {
			t.Error("expected issues map to be cleared")
		}
		if !m.loading {
			t.Error("expected loading to be true")
		}
	})

	t.Run("RefreshRequested on BoardLoaded restores new board", func(t *testing.T) {
		store := &mockStore{}
		m := initialModel("default", store)
		m.loading = true

		newBoard := model.BoardFile{Board: model.Board{Name: "new-name"}}
		result, _ := m.Update(BoardLoaded{Board: newBoard})
		m = result.(Model)
		if m.board.Board.Name != "new-name" {
			t.Errorf("expected board name 'new-name', got %q", m.board.Board.Name)
		}
		if m.loading {
			t.Error("expected loading to be false")
		}
	})
}

func TestWindowResize(t *testing.T) {
	t.Run("tea.WindowSizeMsg updates dimensions", func(t *testing.T) {
		store := &mockStore{}
		m := initialModel("default", store)

		result, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
		m = result.(Model)
		if m.width != 120 {
			t.Errorf("expected width 120, got %d", m.width)
		}
		if m.height != 40 {
			t.Errorf("expected height 40, got %d", m.height)
		}
	})
}

func TestColumnFocus(t *testing.T) {
	t.Run("ColumnFocused updates focusedColumn and resets focusedIssue", func(t *testing.T) {
		store := &mockStore{}
		m := initialModel("default", store)
		m.focusedColumn = 0
		m.focusedIssue = 3

		result, _ := m.Update(ColumnFocused{Index: 2})
		m = result.(Model)
		if m.focusedColumn != 2 {
			t.Errorf("expected focusedColumn 2, got %d", m.focusedColumn)
		}
		if m.focusedIssue != 0 {
			t.Errorf("expected focusedIssue 0, got %d", m.focusedIssue)
		}
	})
}

func TestBoardLoaded(t *testing.T) {
	t.Run("BoardLoaded sets board state", func(t *testing.T) {
		store := &mockStore{}
		m := initialModel("default", store)

		board := model.BoardFile{Board: model.Board{Name: "test"}}
		result, _ := m.Update(BoardLoaded{Board: board})
		m = result.(Model)
		if m.board.Board.Name != "test" {
			t.Errorf("expected board name 'test', got %q", m.board.Board.Name)
		}
		if m.loading {
			t.Error("expected loading to be false")
		}
	})

	t.Run("BoardLoadFailed clears loading", func(t *testing.T) {
		store := &mockStore{}
		m := initialModel("default", store)
		m.loading = true

		result, _ := m.Update(BoardLoadFailed{Err: errTest})
		m = result.(Model)
		if m.loading {
			t.Error("expected loading to be false")
		}
	})
}

func TestCreateSubmit(t *testing.T) {
	t.Run("empty title sets writeErr and does not create issue", func(t *testing.T) {
		store := &mockStore{}
		m := initialModel("default", store)
		m.formTitle = ""
		m.view = viewCreate

		result, cmd := m.Update(CreateSubmit{})
		m = result.(Model)
		if m.writeErr == nil {
			t.Error("expected writeErr to be set")
		}
		if m.writing {
			t.Error("expected writing to be false")
		}
		if cmd != nil {
			t.Error("expected no cmd")
		}
	})

	t.Run("non-empty title triggers write operation", func(t *testing.T) {
		store := &mockStore{}
		m := initialModel("default", store)
		m.formTitle = "Test Issue"
		m.formColumn = 0
		m.formTags = []string{"bug"}
		m.board = model.BoardFile{
			Board: model.Board{
				Columns: []model.Column{
					{ID: "col1", Label: "Col 1"},
				},
				Tags: []string{"bug"},
			},
		}
		m.writing = false

		result, cmd := m.Update(CreateSubmit{})
		m = result.(Model)
		if !m.writing {
			t.Error("expected writing to be true")
		}
		if m.pendingWrite == nil {
			t.Error("expected pendingWrite to be set")
		}
		if cmd == nil {
			t.Error("expected cmd to be returned")
		}
	})
}

func TestMoveSubmit(t *testing.T) {
	t.Run("non-empty detailIssueID triggers write operation", func(t *testing.T) {
		store := &mockStore{}
		m := initialModel("default", store)
		m.detailIssueID = "issue1"
		m.issues["issue1"] = model.IssueFile{
			ID:     "issue1",
			Title:  "Test",
			Column: "col1",
		}
		m.formColumn = 1
		m.board = model.BoardFile{
			Board: model.Board{
				Columns: []model.Column{
					{ID: "col1", Label: "Col 1"},
					{ID: "col2", Label: "Col 2"},
				},
			},
		}

		result, cmd := m.Update(MoveSubmit{})
		m = result.(Model)
		if !m.writing {
			t.Error("expected writing to be true")
		}
		if cmd == nil {
			t.Error("expected cmd to be returned")
		}
	})
}

func TestEditSave(t *testing.T) {
	t.Run("no detailIssueID does nothing", func(t *testing.T) {
		store := &mockStore{}
		m := initialModel("default", store)
		m.detailIssueID = ""

		result, cmd := m.Update(EditSave{})
		m = result.(Model)
		if cmd != nil {
			t.Error("expected no cmd")
		}
	})

	t.Run("no changes returns to detail view without writing", func(t *testing.T) {
		store := &mockStore{}
		m := initialModel("default", store)
		m.detailIssueID = "issue1"
		m.formTitle = "Test Issue"
		m.formDescLines = nil
		m.formTags = nil
		m.issues["issue1"] = model.IssueFile{
			ID:          "issue1",
			Title:       "Test Issue",
			Column:      "col1",
			Description: "",
			Tags:        nil,
		}
		m.board = model.BoardFile{
			Board: model.Board{
				Columns: []model.Column{
					{ID: "col1", Label: "Col 1"},
				},
				Tags: []string{},
			},
		}

		result, cmd := m.Update(EditSave{})
		m = result.(Model)
		if m.view != viewDetail {
			t.Errorf("expected viewDetail, got %v", m.view)
		}
		if cmd != nil {
			t.Error("expected no cmd")
		}
	})

	t.Run("title change triggers write operation", func(t *testing.T) {
		store := &mockStore{}
		m := initialModel("default", store)
		m.detailIssueID = "issue1"
		m.formTitle = "New Title"
		m.formDescLines = nil
		m.formTags = nil
		m.issues["issue1"] = model.IssueFile{
			ID:          "issue1",
			Title:       "Old Title",
			Column:      "col1",
			Description: "",
			Tags:        nil,
		}
		m.board = model.BoardFile{
			Board: model.Board{
				Columns: []model.Column{
					{ID: "col1", Label: "Col 1"},
				},
				Tags: []string{},
			},
		}

		result, cmd := m.Update(EditSave{})
		m = result.(Model)
		if !m.writing {
			t.Error("expected writing to be true")
		}
		if m.pendingWrite == nil {
			t.Error("expected pendingWrite to be set")
		}
		if cmd == nil {
			t.Error("expected cmd to be returned")
		}
	})

	t.Run("description change triggers write operation", func(t *testing.T) {
		store := &mockStore{}
		m := initialModel("default", store)
		m.detailIssueID = "issue1"
		m.formTitle = "Test Issue"
		m.formDescLines = []string{"new description"}
		m.formTags = nil
		m.issues["issue1"] = model.IssueFile{
			ID:          "issue1",
			Title:       "Test Issue",
			Column:      "col1",
			Description: "",
			Tags:        nil,
		}
		m.board = model.BoardFile{
			Board: model.Board{
				Columns: []model.Column{
					{ID: "col1", Label: "Col 1"},
				},
				Tags: []string{},
			},
		}

		result, cmd := m.Update(EditSave{})
		m = result.(Model)
		if !m.writing {
			t.Error("expected writing to be true")
		}
		if cmd == nil {
			t.Error("expected cmd to be returned")
		}
	})

	t.Run("tag change triggers write operation", func(t *testing.T) {
		store := &mockStore{}
		m := initialModel("default", store)
		m.detailIssueID = "issue1"
		m.formTitle = "Test Issue"
		m.formDescLines = nil
		m.formTags = []string{"bug", "urgent"}
		m.issues["issue1"] = model.IssueFile{
			ID:          "issue1",
			Title:       "Test Issue",
			Column:      "col1",
			Description: "",
			Tags:        []string{"bug"},
		}
		m.board = model.BoardFile{
			Board: model.Board{
				Columns: []model.Column{
					{ID: "col1", Label: "Col 1"},
				},
				Tags: []string{"bug", "urgent"},
			},
		}

		result, cmd := m.Update(EditSave{})
		m = result.(Model)
		if !m.writing {
			t.Error("expected writing to be true")
		}
		if cmd == nil {
			t.Error("expected cmd to be returned")
		}
	})
}

func TestFormChanges(t *testing.T) {
	t.Run("FormTitleChanged updates formTitle", func(t *testing.T) {
		store := &mockStore{}
		m := initialModel("default", store)

		result, _ := m.Update(FormTitleChanged{Text: "New Title"})
		m = result.(Model)
		if m.formTitle != "New Title" {
			t.Errorf("expected formTitle 'New Title', got %q", m.formTitle)
		}
	})

	t.Run("FormColumnChanged updates formColumn", func(t *testing.T) {
		store := &mockStore{}
		m := initialModel("default", store)

		result, _ := m.Update(FormColumnChanged{Index: 3})
		m = result.(Model)
		if m.formColumn != 3 {
			t.Errorf("expected formColumn 3, got %d", m.formColumn)
		}
	})

	t.Run("FormTagsChanged updates formTags", func(t *testing.T) {
		store := &mockStore{}
		m := initialModel("default", store)

		result, _ := m.Update(FormTagsChanged{Tags: []string{"bug", "urgent"}})
		m = result.(Model)
		if len(m.formTags) != 2 {
			t.Errorf("expected 2 tags, got %d", len(m.formTags))
		}
	})

	t.Run("FormDescLineAdded appends line", func(t *testing.T) {
		store := &mockStore{}
		m := initialModel("default", store)
		m.formDescLines = []string{"line1"}

		result, _ := m.Update(FormDescLineAdded{Line: "line2"})
		m = result.(Model)
		if len(m.formDescLines) != 2 {
			t.Errorf("expected 2 lines, got %d", len(m.formDescLines))
		}
		if m.formDescLines[1] != "line2" {
			t.Errorf("expected line2 at index 1, got %q", m.formDescLines[1])
		}
	})

	t.Run("FormDescConfirmed exits multiline mode", func(t *testing.T) {
		store := &mockStore{}
		m := initialModel("default", store)
		m.formDescEditing = true

		result, _ := m.Update(FormDescConfirmed{})
		m = result.(Model)
		if m.formDescEditing {
			t.Error("expected formDescEditing to be false")
		}
	})
}

func TestInitialModel(t *testing.T) {
	m := initialModel("myboard", &mockStore{})

	if m.boardName != "myboard" {
		t.Errorf("expected boardName 'myboard', got %q", m.boardName)
	}
	if m.view != viewBoard {
		t.Errorf("expected viewBoard, got %v", m.view)
	}
	if m.focusedColumn != 0 {
		t.Errorf("expected focusedColumn 0, got %d", m.focusedColumn)
	}
	if m.focusedIssue != 0 {
		t.Errorf("expected focusedIssue 0, got %d", m.focusedIssue)
	}
	if m.store == nil {
		t.Error("expected store to be set")
	}
}

func TestParseDescription(t *testing.T) {
	t.Run("empty string returns nil", func(t *testing.T) {
		lines := parseDescription("")
		if lines != nil {
			t.Errorf("expected nil, got %v", lines)
		}
	})

	t.Run("single line without newline", func(t *testing.T) {
		lines := parseDescription("hello")
		if len(lines) != 1 || lines[0] != "hello" {
			t.Errorf("expected ['hello'], got %v", lines)
		}
	})

	t.Run("multiline splits correctly", func(t *testing.T) {
		lines := parseDescription("line1\nline2\nline3")
		if len(lines) != 3 {
			t.Fatalf("expected 3 lines, got %d", len(lines))
		}
		if lines[0] != "line1" || lines[1] != "line2" || lines[2] != "line3" {
			t.Errorf("unexpected lines: %v", lines)
		}
	})
}

func TestMatchTags(t *testing.T) {
	tags := []string{"bug", "feature", "docs", "chore", "urgent"}

	t.Run("empty input returns all tags", func(t *testing.T) {
		matches := matchTags("", tags)
		if len(matches) != len(tags) {
			t.Errorf("expected %d matches, got %d", len(tags), len(matches))
		}
	})

	t.Run("prefix match", func(t *testing.T) {
		matches := matchTags("bu", tags)
		if len(matches) != 1 || matches[0] != "bug" {
			t.Errorf("expected [bug], got %v", matches)
		}
	})

	t.Run("no match", func(t *testing.T) {
		matches := matchTags("xyz", tags)
		if len(matches) != 0 {
			t.Errorf("expected [], got %v", matches)
		}
	})

	t.Run("case sensitive", func(t *testing.T) {
		matches := matchTags("BU", tags)
		if len(matches) != 0 {
			t.Errorf("expected [], got %v", matches)
		}
	})
}

func TestTabCompletion(t *testing.T) {
	t.Run("TabPressed in filter view activates completion", func(t *testing.T) {
		store := &mockStore{}
		m := initialModel("default", store)
		m.view = viewFilter
		m.board = model.BoardFile{
			Board: model.Board{
				Tags: []string{"bug", "feature", "chore"},
			},
		}

		result, _ := m.Update(TabPressed{})
		m = result.(Model)
		if !m.completion.active {
			t.Error("expected completion.active to be true")
		}
		if len(m.completion.matches) != 3 {
			t.Errorf("expected 3 matches, got %d", len(m.completion.matches))
		}
		if m.completion.index != 0 {
			t.Errorf("expected index 0, got %d", m.completion.index)
		}
	})

	t.Run("TabPressed in create view with no tagInput shows all tags", func(t *testing.T) {
		store := &mockStore{}
		m := initialModel("default", store)
		m.view = viewCreate
		m.tagInput = ""
		m.board = model.BoardFile{
			Board: model.Board{
				Tags: []string{"bug", "feature", "chore"},
				Columns: []model.Column{
					{ID: "col1", Label: "Col 1"},
				},
			},
		}

		result, _ := m.Update(TabPressed{})
		m = result.(Model)
		if !m.completion.active {
			t.Error("expected completion.active to be true")
		}
		if len(m.completion.matches) != 3 {
			t.Errorf("expected 3 matches, got %d", len(m.completion.matches))
		}
		if m.completion.target != "tag" {
			t.Errorf("expected target 'tag', got %q", m.completion.target)
		}
	})

	t.Run("TabPressed in create view with tagInput filters matches", func(t *testing.T) {
		store := &mockStore{}
		m := initialModel("default", store)
		m.view = viewCreate
		m.tagInput = "bu"
		m.board = model.BoardFile{
			Board: model.Board{
				Tags: []string{"bug", "feature", "chore"},
				Columns: []model.Column{
					{ID: "col1", Label: "Col 1"},
				},
			},
		}

		result, _ := m.Update(TabPressed{})
		m = result.(Model)
		if !m.completion.active {
			t.Error("expected completion.active to be true")
		}
		if len(m.completion.matches) != 1 {
			t.Errorf("expected 1 match, got %d", len(m.completion.matches))
		}
		if m.completion.matches[0] != "bug" {
			t.Errorf("expected match 'bug', got %q", m.completion.matches[0])
		}
		if m.completion.input != "bu" {
			t.Errorf("expected input 'bu', got %q", m.completion.input)
		}
		if m.completion.target != "tag" {
			t.Errorf("expected target 'tag', got %q", m.completion.target)
		}
	})

	t.Run("TabPressed in edit view with tagInput filters matches", func(t *testing.T) {
		store := &mockStore{}
		m := initialModel("default", store)
		m.view = viewEdit
		m.tagInput = "fea"
		m.board = model.BoardFile{
			Board: model.Board{
				Tags: []string{"bug", "feature", "chore"},
				Columns: []model.Column{
					{ID: "col1", Label: "Col 1"},
				},
			},
		}

		result, _ := m.Update(TabPressed{})
		m = result.(Model)
		if !m.completion.active {
			t.Error("expected completion.active to be true")
		}
		if len(m.completion.matches) != 1 {
			t.Errorf("expected 1 match, got %d", len(m.completion.matches))
		}
		if m.completion.matches[0] != "feature" {
			t.Errorf("expected match 'feature', got %q", m.completion.matches[0])
		}
	})

	t.Run("TabPressed outside filter view does nothing", func(t *testing.T) {
		store := &mockStore{}
		m := initialModel("default", store)
		m.view = viewBoard

		result, _ := m.Update(TabPressed{})
		m = result.(Model)
		if m.completion.active {
			t.Error("expected completion.active to be false")
		}
	})

	t.Run("TabCompletionResult deactivates completion", func(t *testing.T) {
		store := &mockStore{}
		m := initialModel("default", store)
		m.completion.active = true

		result, _ := m.Update(TabCompletionResult{})
		m = result.(Model)
		if m.completion.active {
			t.Error("expected completion.active to be false")
		}
	})
}

func TestTagInputChanged(t *testing.T) {
	t.Run("TagInputChanged activates completion and sets tagInput", func(t *testing.T) {
		store := &mockStore{}
		m := initialModel("default", store)
		m.view = viewCreate
		m.board = model.BoardFile{
			Board: model.Board{
				Tags: []string{"bug", "feature", "chore"},
				Columns: []model.Column{
					{ID: "col1", Label: "Col 1"},
				},
			},
		}

		result, _ := m.Update(TagInputChanged{Text: "bu"})
		m = result.(Model)
		if !m.completion.active {
			t.Error("expected completion.active to be true")
		}
		if m.tagInput != "bu" {
			t.Errorf("expected tagInput 'bu', got %q", m.tagInput)
		}
		if len(m.completion.matches) != 1 || m.completion.matches[0] != "bug" {
			t.Errorf("expected matches [bug], got %v", m.completion.matches)
		}
		if m.completion.target != "tag" {
			t.Errorf("expected target 'tag', got %q", m.completion.target)
		}
	})

	t.Run("TagInputChanged with no matching input shows empty matches", func(t *testing.T) {
		store := &mockStore{}
		m := initialModel("default", store)
		m.view = viewCreate
		m.board = model.BoardFile{
			Board: model.Board{
				Tags: []string{"bug", "feature", "chore"},
			},
		}

		result, _ := m.Update(TagInputChanged{Text: "xyz"})
		m = result.(Model)
		if !m.completion.active {
			t.Error("expected completion.active to be true")
		}
		if len(m.completion.matches) != 0 {
			t.Errorf("expected no matches, got %d", len(m.completion.matches))
		}
	})
}

func TestContains(t *testing.T) {
	t.Run("returns true when needle present", func(t *testing.T) {
		if !contains([]string{"a", "b", "c"}, "b") {
			t.Error("expected true")
		}
	})

	t.Run("returns false when needle absent", func(t *testing.T) {
		if contains([]string{"a", "b", "c"}, "d") {
			t.Error("expected false")
		}
	})

	t.Run("returns false for empty haystack", func(t *testing.T) {
		if contains([]string{}, "a") {
			t.Error("expected false")
		}
	})
}

func makeKey(name string) tea.KeyMsg {
	switch name {
	case "up":
		return tea.KeyPressMsg{Code: tea.KeyUp}
	case "down":
		return tea.KeyPressMsg{Code: tea.KeyDown}
	case "enter":
		return tea.KeyPressMsg{Code: tea.KeyEnter}
	case "tab":
		return tea.KeyPressMsg{Code: tea.KeyTab}
	case "esc":
		return tea.KeyPressMsg{Code: tea.KeyEsc}
	case "left":
		return tea.KeyPressMsg{Code: tea.KeyLeft}
	case "right":
		return tea.KeyPressMsg{Code: tea.KeyRight}
	default:
		r := rune(name[0])
		return tea.KeyPressMsg{Code: r}
	}
}

func TestCreateViewCycling(t *testing.T) {
	t.Run("up cycles formColumn down", func(t *testing.T) {
		store := &mockStore{}
		m := initialModel("default", store)
		m.view = viewCreate
		m.formColumn = 0
		m.board = model.BoardFile{
			Board: model.Board{
				Columns: []model.Column{
					{ID: "col1", Label: "Col 1"},
					{ID: "col2", Label: "Col 2"},
				},
			},
		}

		result, _ := m.Update(makeKey("up"))
		m = result.(Model)
		if m.formColumn != 1 {
			t.Errorf("expected formColumn 1, got %d", m.formColumn)
		}
	})

	t.Run("down cycles formColumn up", func(t *testing.T) {
		store := &mockStore{}
		m := initialModel("default", store)
		m.view = viewCreate
		m.formColumn = 0
		m.board = model.BoardFile{
			Board: model.Board{
				Columns: []model.Column{
					{ID: "col1", Label: "Col 1"},
					{ID: "col2", Label: "Col 2"},
				},
			},
		}

		result, _ := m.Update(makeKey("down"))
		m = result.(Model)
		if m.formColumn != 1 {
			t.Errorf("expected formColumn 1, got %d", m.formColumn)
		}
	})

	t.Run("up from 0 wraps to last column", func(t *testing.T) {
		store := &mockStore{}
		m := initialModel("default", store)
		m.view = viewCreate
		m.formColumn = 0
		m.board = model.BoardFile{
			Board: model.Board{
				Columns: []model.Column{
					{ID: "col1", Label: "Col 1"},
					{ID: "col2", Label: "Col 2"},
				},
			},
		}

		result, _ := m.Update(makeKey("up"))
		m = result.(Model)
		if m.formColumn != 1 {
			t.Errorf("expected formColumn 1, got %d", m.formColumn)
		}
	})

	t.Run("down from last wraps to 0", func(t *testing.T) {
		store := &mockStore{}
		m := initialModel("default", store)
		m.view = viewCreate
		m.formColumn = 1
		m.board = model.BoardFile{
			Board: model.Board{
				Columns: []model.Column{
					{ID: "col1", Label: "Col 1"},
					{ID: "col2", Label: "Col 2"},
				},
			},
		}

		result, _ := m.Update(makeKey("down"))
		m = result.(Model)
		if m.formColumn != 0 {
			t.Errorf("expected formColumn 0, got %d", m.formColumn)
		}
	})
}

func TestEscapeKeyDiscardsFormState(t *testing.T) {
	t.Run("esc from create returns to board", func(t *testing.T) {
		store := &mockStore{}
		m := initialModel("default", store)
		m.view = viewCreate
		m.formTitle = "some title"
		m.formColumn = 2
		m.formTags = []string{"bug"}
		m.formDescLines = []string{"desc line"}
		m.tagInput = "feat"
		m.completion.active = false

		result, _ := m.Update(makeKey("esc"))
		m = result.(Model)
		if m.view != viewBoard {
			t.Errorf("expected viewBoard, got %v", m.view)
		}
	})

	t.Run("esc from move returns to detail", func(t *testing.T) {
		store := &mockStore{}
		m := initialModel("default", store)
		m.view = viewMove

		result, _ := m.Update(makeKey("esc"))
		m = result.(Model)
		if m.view != viewDetail {
			t.Errorf("expected viewDetail, got %v", m.view)
		}
	})

	t.Run("esc from edit returns to detail", func(t *testing.T) {
		store := &mockStore{}
		m := initialModel("default", store)
		m.view = viewEdit
		m.formTitle = "modified"

		result, _ := m.Update(makeKey("esc"))
		m = result.(Model)
		if m.view != viewDetail {
			t.Errorf("expected viewDetail, got %v", m.view)
		}
	})
}

func TestCompletionEscapeClearsState(t *testing.T) {
	t.Run("esc during active completion clears completion and tagInput", func(t *testing.T) {
		store := &mockStore{}
		m := initialModel("default", store)
		m.view = viewCreate
		m.tagInput = "bu"
		m.completion.active = true
		m.completion.target = "tag"

		result, _ := m.Update(makeKey("esc"))
		m = result.(Model)
		if m.completion.active {
			t.Error("expected completion.active to be false")
		}
		if m.tagInput != "" {
			t.Errorf("expected tagInput to be cleared, got %q", m.tagInput)
		}
	})

	t.Run("c during active completion clears completion and tagInput", func(t *testing.T) {
		store := &mockStore{}
		m := initialModel("default", store)
		m.view = viewFilter
		m.tagFilter = "bu"
		m.completion.active = true
		m.completion.target = "filter"

		result, _ := m.Update(makeKey("c"))
		m = result.(Model)
		if m.completion.active {
			t.Error("expected completion.active to be false")
		}
	})
}

func TestCompletionNavigateCycles(t *testing.T) {
	t.Run("tab cycles completion index forward", func(t *testing.T) {
		store := &mockStore{}
		m := initialModel("default", store)
		m.view = viewCreate
		m.tagInput = ""
		m.completion.active = true
		m.completion.matches = []string{"bug", "feature", "chore"}
		m.completion.index = 0

		result, _ := m.Update(makeKey("tab"))
		m = result.(Model)
		if m.completion.index != 1 {
			t.Errorf("expected index 1, got %d", m.completion.index)
		}
	})

	t.Run("down cycles completion index forward", func(t *testing.T) {
		store := &mockStore{}
		m := initialModel("default", store)
		m.view = viewCreate
		m.completion.active = true
		m.completion.matches = []string{"bug", "feature", "chore"}
		m.completion.index = 1

		result, _ := m.Update(makeKey("down"))
		m = result.(Model)
		if m.completion.index != 2 {
			t.Errorf("expected index 2, got %d", m.completion.index)
		}
	})

	t.Run("up cycles completion index backward", func(t *testing.T) {
		store := &mockStore{}
		m := initialModel("default", store)
		m.view = viewCreate
		m.completion.active = true
		m.completion.matches = []string{"bug", "feature", "chore"}
		m.completion.index = 1

		result, _ := m.Update(makeKey("up"))
		m = result.(Model)
		if m.completion.index != 0 {
			t.Errorf("expected index 0, got %d", m.completion.index)
		}
	})

	t.Run("up from index 0 wraps to last", func(t *testing.T) {
		store := &mockStore{}
		m := initialModel("default", store)
		m.view = viewCreate
		m.completion.active = true
		m.completion.matches = []string{"bug", "feature", "chore"}
		m.completion.index = 0

		result, _ := m.Update(makeKey("up"))
		m = result.(Model)
		if m.completion.index != 2 {
			t.Errorf("expected index 2, got %d", m.completion.index)
		}
	})
}

func TestCompletionEnterConfirmsSelection(t *testing.T) {
	t.Run("enter with matches appends selected tag to formTags", func(t *testing.T) {
		store := &mockStore{}
		m := initialModel("default", store)
		m.view = viewCreate
		m.tagInput = "bu"
		m.formTags = []string{}
		m.completion.active = true
		m.completion.matches = []string{"bug", "build"}
		m.completion.index = 0
		m.completion.target = "tag"
		m.completion.input = "bu"

		result, _ := m.Update(makeKey("enter"))
		m = result.(Model)
		if !contains(m.formTags, "bug") {
			t.Error("expected bug to be added to formTags")
		}
		if m.completion.active {
			t.Error("expected completion to be deactivated")
		}
		if m.tagInput != "" {
			t.Errorf("expected tagInput to be cleared, got %q", m.tagInput)
		}
	})

	t.Run("enter with no matches but non-empty tagInput adds tagInput as tag", func(t *testing.T) {
		store := &mockStore{}
		m := initialModel("default", store)
		m.view = viewCreate
		m.tagInput = "newtag"
		m.formTags = []string{}
		m.completion.active = true
		m.completion.matches = []string{}
		m.completion.target = "tag"

		result, _ := m.Update(makeKey("enter"))
		m = result.(Model)
		if !contains(m.formTags, "newtag") {
			t.Error("expected newtag to be added to formTags")
		}
	})

	t.Run("enter in filter view sets tagFilter", func(t *testing.T) {
		store := &mockStore{}
		m := initialModel("default", store)
		m.view = viewFilter
		m.tagFilter = "bu"
		m.completion.active = true
		m.completion.matches = []string{"bug"}
		m.completion.index = 0
		m.completion.target = "filter"

		result, _ := m.Update(makeKey("enter"))
		m = result.(Model)
		if m.tagFilter != "bug" {
			t.Errorf("expected tagFilter 'bug', got %q", m.tagFilter)
		}
	})
}

func TestEnterInDetailDoesNotTriggerCreate(t *testing.T) {
	t.Run("enter in detail view does nothing", func(t *testing.T) {
		store := &mockStore{}
		m := initialModel("default", store)
		m.view = viewDetail
		m.detailIssueID = "issue1"
		m.issues["issue1"] = model.IssueFile{ID: "issue1", Title: "Test", Column: "col1"}

		_, cmd := m.Update(makeKey("enter"))
		if cmd != nil {
			t.Error("expected no cmd")
		}
	})
}

func TestWriteErrorStateSet(t *testing.T) {
	t.Run("writeErr is set on WriteCompleted error", func(t *testing.T) {
		store := &mockStore{}
		m := initialModel("default", store)
		m.writing = true
		m.pendingWrite = pendingIssueSave{
			boardName: "default",
			issue:     model.IssueFile{ID: "issue1"},
			entries:   nil,
		}

		result, _ := m.Update(WriteCompleted{Err: errTest})
		m = result.(Model)
		if m.writeErr == nil {
			t.Error("expected writeErr to be set")
		}
		if m.writing {
			t.Error("expected writing to be false after WriteCompleted")
		}
	})
}

var _ = errTest // suppress unused warning
