package tui

import (
	"testing"

	"github.com/jana-mind/bubbler/internal/model"
)

func TestMockStore(t *testing.T) {
	s := &mockStore{}

	t.Run("LoadBoard returns empty board", func(t *testing.T) {
		bf, err := s.LoadBoard("default")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if bf.Board.Name != "" {
			t.Errorf("expected empty board, got name %q", bf.Board.Name)
		}
	})

	t.Run("LoadIssue returns empty issue", func(t *testing.T) {
		f, err := s.LoadIssue("default", "abc123")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if f.ID != "" {
			t.Errorf("expected empty issue, got ID %q", f.ID)
		}
	})

	t.Run("SaveBoard returns no error", func(t *testing.T) {
		err := s.SaveBoard("default", model.BoardFile{})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("SaveIssue returns no error", func(t *testing.T) {
		err := s.SaveIssue("default", model.IssueFile{}, nil)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("DeleteIssue returns no error", func(t *testing.T) {
		err := s.DeleteIssue("default", "abc123")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})
}

type errorStore struct {
	loadBoardErr   error
	loadIssueErr   error
	saveBoardErr   error
	saveIssueErr   error
	deleteIssueErr error
}

func (s *errorStore) LoadBoard(boardName string) (model.BoardFile, error) {
	return model.BoardFile{}, s.loadBoardErr
}

func (s *errorStore) SaveBoard(boardName string, bf model.BoardFile) error {
	return s.saveBoardErr
}

func (s *errorStore) LoadIssue(boardName, issueID string) (model.IssueFile, error) {
	return model.IssueFile{}, s.loadIssueErr
}

func (s *errorStore) SaveIssue(boardName string, issue model.IssueFile, entries []model.HistoryEntry) error {
	return s.saveIssueErr
}

func (s *errorStore) DeleteIssue(boardName, issueID string) error {
	return s.deleteIssueErr
}

func (s *errorStore) RepoRoot() string {
	return ""
}

func TestStoreErrors(t *testing.T) {
	t.Run("LoadBoard error propagates", func(t *testing.T) {
		s := &errorStore{loadBoardErr: errTest}
		_, err := s.LoadBoard("default")
		if err != errTest {
			t.Errorf("expected errTest, got %v", err)
		}
	})

	t.Run("LoadIssue error propagates", func(t *testing.T) {
		s := &errorStore{loadIssueErr: errTest}
		_, err := s.LoadIssue("default", "abc123")
		if err != errTest {
			t.Errorf("expected errTest, got %v", err)
		}
	})

	t.Run("SaveIssue error propagates", func(t *testing.T) {
		s := &errorStore{saveIssueErr: errTest}
		err := s.SaveIssue("default", model.IssueFile{}, nil)
		if err != errTest {
			t.Errorf("expected errTest, got %v", err)
		}
	})

	t.Run("DeleteIssue error propagates", func(t *testing.T) {
		s := &errorStore{deleteIssueErr: errTest}
		err := s.DeleteIssue("default", "abc123")
		if err != errTest {
			t.Errorf("expected errTest, got %v", err)
		}
	})
}

var errTest = errorStoreErr{}

type errorStoreErr struct{}

func (e errorStoreErr) Error() string { return "test error" }

func TestErrorStoreWithModel(t *testing.T) {
	t.Run("RefreshRequested sets loading and returns cmd for async load", func(t *testing.T) {
		s := &errorStore{loadBoardErr: errTest}
		m := initialModel("default", s)
		m.loading = false

		result, cmd := m.Update(RefreshRequested{})
		m = result.(Model)
		if !m.loading {
			t.Error("expected loading to be true after RefreshRequested")
		}
		if cmd == nil {
			t.Error("expected cmd to be returned for board reload")
		}
	})

	t.Run("BoardLoadFailed clears loading and sets writeErr", func(t *testing.T) {
		s := &errorStore{}
		m := initialModel("default", s)
		m.loading = true

		result, _ := m.Update(BoardLoadFailed{Err: errTest})
		m = result.(Model)
		if m.loading {
			t.Error("expected loading to be false after BoardLoadFailed")
		}
		if m.writeErr == nil {
			t.Error("expected writeErr to be set")
		}
	})

	t.Run("LoadIssue error does not panic and issue remains unloaded", func(t *testing.T) {
		s := &errorStore{loadIssueErr: errTest}
		m := initialModel("default", s)
		m.view = viewDetail
		m.detailIssueID = "issue1"
		m.board = model.BoardFile{
			Board: model.Board{
				Columns: []model.Column{
						{ID: "col1", Label: "Col 1"},
					},
					Tags: []string{},
				},
			}

		m.issues["issue1"] = model.IssueFile{ID: "issue1", Title: "Test", Column: "col1"}
		result, _ := m.Update(IssueFocused{IssueID: "issue1"})
		m = result.(Model)
		_, ok := m.issues["issue1"]
		if !ok {
			t.Error("expected issue1 to remain in issues map")
		}
	})
}

func TestMockStoreWithModel(t *testing.T) {
	t.Run("RefreshRequested clears issues and sets loading", func(t *testing.T) {
		s := &mockStore{}
		m := initialModel("default", s)
		m.board = model.BoardFile{Board: model.Board{Name: "old"}}

		result, cmd := m.Update(RefreshRequested{})
		m = result.(Model)
		if cmd == nil {
			t.Error("expected cmd for board reload")
		}
		if len(m.issues) != 0 {
			t.Error("expected issues map to be cleared")
		}
		if !m.loading {
			t.Error("expected loading to be true")
		}

		newBoard := model.BoardFile{Board: model.Board{Name: "new-name"}}
		result, _ = m.Update(BoardLoaded{Board: newBoard})
		m = result.(Model)
		if m.board.Board.Name != "new-name" {
			t.Errorf("expected board name 'new-name', got %q", m.board.Board.Name)
		}
		if m.loading {
			t.Error("expected loading to be false")
		}
	})
}
