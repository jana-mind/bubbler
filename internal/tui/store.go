package tui

import (
	"fmt"
	"path/filepath"

	"github.com/jana-mind/bubbler/internal/git"
	"github.com/jana-mind/bubbler/internal/model"
	"github.com/jana-mind/bubbler/internal/store"
)

type Store interface {
	LoadBoard(boardName string) (model.BoardFile, error)
	SaveBoard(boardName string, bf model.BoardFile) error
	LoadIssue(boardName, issueID string) (model.IssueFile, error)
	SaveIssue(boardName string, issue model.IssueFile, entries []model.HistoryEntry) error
	DeleteIssue(boardName, issueID string) error
	RepoRoot() string
}

type realStore struct {
	repoRoot string
}

func newRealStore() (*realStore, error) {
	root, err := git.FindRepoRoot()
	if err != nil {
		return nil, err
	}
	return &realStore{repoRoot: root}, nil
}

func (s *realStore) RepoRoot() string {
	return s.repoRoot
}

func (s *realStore) boardPath(boardName string) string {
	return filepath.Join(s.repoRoot, ".bubble", boardName+".yaml")
}

func (s *realStore) issuePath(boardName, issueID string) string {
	return filepath.Join(s.repoRoot, ".bubble", boardName, issueID+".yaml")
}

func (s *realStore) LoadBoard(boardName string) (model.BoardFile, error) {
	return store.LoadBoardFileSubmodule(s.boardPath(boardName))
}

func (s *realStore) SaveBoard(boardName string, bf model.BoardFile) error {
	return store.SaveBoardFileSubmodule(s.boardPath(boardName), bf)
}

func (s *realStore) LoadIssue(boardName, issueID string) (model.IssueFile, error) {
	return store.LoadIssueFileSubmodule(s.issuePath(boardName, issueID))
}

func (s *realStore) SaveIssue(boardName string, issue model.IssueFile, entries []model.HistoryEntry) error {
	return store.SaveIssueFileSubmodule(s.issuePath(boardName, issue.ID), issue, entries)
}

func (s *realStore) DeleteIssue(boardName, issueID string) error {
	issuePath := s.issuePath(boardName, issueID)
	boardPath := s.boardPath(boardName)
	if git.IsSubmodule(s.repoRoot, filepath.Dir(boardPath)) {
		if err := git.Pull(s.repoRoot, filepath.Dir(boardPath)); err != nil {
			return err
		}
	}
	if err := store.DeleteIssueFile(issuePath, boardPath, issueID); err != nil {
		return err
	}
	if git.IsSubmodule(s.repoRoot, filepath.Dir(boardPath)) {
		if err := git.CommitAndPush(s.repoRoot, filepath.Dir(boardPath), fmt.Sprintf("bubbler: delete issue %s", issueID)); err != nil {
			return err
		}
	}
	return nil
}

type mockStore struct{}

func (m *mockStore) RepoRoot() string {
	return ""
}

func (m *mockStore) LoadBoard(boardName string) (model.BoardFile, error) {
	return model.BoardFile{}, nil
}

func (m *mockStore) SaveBoard(boardName string, bf model.BoardFile) error {
	return nil
}

func (m *mockStore) LoadIssue(boardName, issueID string) (model.IssueFile, error) {
	return model.IssueFile{}, nil
}

func (m *mockStore) SaveIssue(boardName string, issue model.IssueFile, entries []model.HistoryEntry) error {
	return nil
}

func (m *mockStore) DeleteIssue(boardName, issueID string) error {
	return nil
}
