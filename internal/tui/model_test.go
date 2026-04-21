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
	t.Run("RefreshRequested reloads board from store", func(t *testing.T) {
		store := &mockStore{}
		m := initialModel("default", store)
		m.board = model.BoardFile{Board: model.Board{Name: "old"}}

		result, _ := m.Update(RefreshRequested{})
		m = result.(Model)
		if m.board.Board.Name != "" {
			t.Errorf("expected board to be reloaded, got name %q", m.board.Board.Name)
		}
		if len(m.issues) != 0 {
			t.Error("expected issues map to be cleared")
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

var _ = errTest // suppress unused warning