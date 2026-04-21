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
	loadBoardErr  error
	loadIssueErr  error
	saveBoardErr  error
	saveIssueErr  error
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