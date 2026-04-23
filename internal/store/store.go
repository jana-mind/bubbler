package store

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jana-mind/bubbler/internal/git"
	"github.com/jana-mind/bubbler/internal/model"
)

func bubbleAndRepoRoot(filePath string) (bubblePath, repoRoot string, err error) {
	ext := filepath.Ext(filePath)
	if ext == ".yaml" {
		bubblePath = filepath.Dir(filePath)
		repoRoot = filepath.Dir(bubblePath)
	} else {
		bubblePath = filepath.Dir(filepath.Dir(filePath))
		repoRoot = filepath.Dir(bubblePath)
	}
	if bubblePath == "" || repoRoot == "" {
		return "", "", fmt.Errorf("could not derive bubble/repo paths from %q", filePath)
	}
	return bubblePath, repoRoot, nil
}

func LoadBoardFileSubmodule(path string) (model.BoardFile, error) {
	bubblePath, repoRoot, err := bubbleAndRepoRoot(path)
	if err != nil {
		return model.BoardFile{}, err
	}
	if git.IsSubmodule(repoRoot, bubblePath) {
		if err := git.Pull(repoRoot, bubblePath); err != nil {
			return model.BoardFile{}, fmt.Errorf("submodule pull: %w", err)
		}
	}
	return LoadBoardFile(path)
}

func SaveBoardFileSubmodule(path string, bf model.BoardFile) error {
	bubblePath, repoRoot, err := bubbleAndRepoRoot(path)
	if err != nil {
		return err
	}
	if git.IsSubmodule(repoRoot, bubblePath) {
		if err := git.Pull(repoRoot, bubblePath); err != nil {
			return fmt.Errorf("submodule pull: %w", err)
		}
	}
	if err := SaveBoardFile(path, bf); err != nil {
		return err
	}
	if git.IsSubmodule(repoRoot, bubblePath) {
		if err := git.CommitAndPush(repoRoot, bubblePath, commitMessageForBoard()); err != nil {
			return fmt.Errorf("submodule commit/push: %w", err)
		}
	}
	return nil
}

func LoadIssueFileSubmodule(path string) (model.IssueFile, error) {
	bubblePath, repoRoot, err := bubbleAndRepoRoot(path)
	if err != nil {
		return model.IssueFile{}, err
	}
	if git.IsSubmodule(repoRoot, bubblePath) {
		if err := git.Pull(repoRoot, bubblePath); err != nil {
			return model.IssueFile{}, fmt.Errorf("submodule pull: %w", err)
		}
	}
	return model.LoadIssueFile(path)
}

func SaveIssueFileSubmodule(path string, issue model.IssueFile, entries []model.HistoryEntry) error {
	bubblePath, repoRoot, err := bubbleAndRepoRoot(path)
	if err != nil {
		return err
	}
	if git.IsSubmodule(repoRoot, bubblePath) {
		if err := git.Pull(repoRoot, bubblePath); err != nil {
			return fmt.Errorf("submodule pull: %w", err)
		}
	}
	if err := model.SaveIssueFile(path, issue, entries); err != nil {
		return err
	}
	if git.IsSubmodule(repoRoot, bubblePath) {
		if err := git.CommitAndPush(repoRoot, bubblePath, commitMessageForIssue(issue.ID)); err != nil {
			return fmt.Errorf("submodule commit/push: %w", err)
		}
	}
	return nil
}

func DeleteIssueFile(issuePath, boardPath string, issueID string) error {
	bf, err := LoadBoardFile(boardPath)
	if err != nil {
		return fmt.Errorf("load board file: %w", err)
	}
	found := false
	for i, issue := range bf.Issues {
		if issue.ID == issueID {
			bf.Issues = append(bf.Issues[:i], bf.Issues[i+1:]...)
			found = true
			if err := SaveBoardFile(boardPath, bf); err != nil {
				return fmt.Errorf("save board file: %w", err)
			}
			break
		}
	}
	if !found {
		return fmt.Errorf("issue %q not found in board state", issueID)
	}
	if _, err := os.Stat(issuePath); os.IsNotExist(err) {
		return nil
	}
	if err := os.Remove(issuePath); err != nil {
		return fmt.Errorf("remove issue file: %w", err)
	}
	return nil
}

func DeleteBoardFileSubmodule(boardFile, issuesDir, boardName string) error {
	bubblePath := filepath.Dir(boardFile)
	repoRoot := filepath.Dir(bubblePath)
	if git.IsSubmodule(repoRoot, bubblePath) {
		if err := git.Pull(repoRoot, bubblePath); err != nil {
			return fmt.Errorf("submodule pull: %w", err)
		}
	}
	if err := os.Remove(boardFile); err != nil {
		return fmt.Errorf("remove board file: %w", err)
	}
	if err := os.RemoveAll(issuesDir); err != nil {
		return fmt.Errorf("remove issues directory: %w", err)
	}
	if git.IsSubmodule(repoRoot, bubblePath) {
		msg := fmt.Sprintf("bubbler: delete board %s", boardName)
		if err := git.CommitAndPush(repoRoot, bubblePath, msg); err != nil {
			return fmt.Errorf("submodule commit/push: %w", err)
		}
	}
	return nil
}

func commitMessageForBoard() string {
	return "bubbler: update board"
}

func commitMessageForIssue(issueID string) string {
	return fmt.Sprintf("bubbler: update issue %s", issueID)
}
