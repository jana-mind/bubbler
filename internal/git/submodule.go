package git

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
)

var (
	ErrNoChanges     = errors.New("no changes to commit")
	ErrNothingToPush = errors.New("nothing to push")
)

func IsSubmodule(repoRoot, bubblePath string) bool {
	manualPath := filepath.Join(bubblePath, ".bubble-manual")
	if _, err := os.Stat(manualPath); err == nil {
		return false
	}

	modulesFile := filepath.Join(repoRoot, ".gitmodules")
	data, err := os.ReadFile(modulesFile)
	if err != nil {
		return false
	}
	relPath, err := filepath.Rel(repoRoot, bubblePath)
	comparePath := bubblePath
	if err == nil {
		comparePath = relPath
	}
	lines := strings.Split(string(data), "\n")
	inSubmodule := false
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "[submodule ") {
			content := strings.Trim(line[11:], "\"]")
			inSubmodule = content == comparePath
		} else if inSubmodule && strings.HasPrefix(line, "path = ") {
			return strings.Trim(line[7:], " \"") == comparePath
		}
	}
	return false
}

func Pull(repoRoot, bubblePath string) error {
	parentRepo, err := git.PlainOpen(repoRoot)
	if err != nil {
		return fmt.Errorf("open parent repo: %w", err)
	}
	wt, err := parentRepo.Worktree()
	if err != nil {
		return fmt.Errorf("get parent worktree: %w", err)
	}
	sm, err := wt.Submodule(bubblePath)
	if err != nil {
		return fmt.Errorf("get submodule %q: %w", bubblePath, err)
	}
	subRepo, err := sm.Repository()
	if err != nil {
		return fmt.Errorf("open submodule repository: %w", err)
	}
	subWt, err := subRepo.Worktree()
	if err != nil {
		return fmt.Errorf("get submodule worktree: %w", err)
	}
	err = subWt.Pull(&git.PullOptions{})
	if err != nil && !errors.Is(err, git.NoErrAlreadyUpToDate) {
		return fmt.Errorf("pull: %w", err)
	}
	return nil
}

func StageAndCommit(repoRoot, bubblePath, message string) error {
	parentRepo, err := git.PlainOpen(repoRoot)
	if err != nil {
		return fmt.Errorf("open parent repo: %w", err)
	}
	wt, err := parentRepo.Worktree()
	if err != nil {
		return fmt.Errorf("get parent worktree: %w", err)
	}
	sm, err := wt.Submodule(bubblePath)
	if err != nil {
		return fmt.Errorf("get submodule %q: %w", bubblePath, err)
	}
	subRepo, err := sm.Repository()
	if err != nil {
		return fmt.Errorf("open submodule repository: %w", err)
	}
	subWt, err := subRepo.Worktree()
	if err != nil {
		return fmt.Errorf("get submodule worktree: %w", err)
	}
	_, err = subWt.Add(".")
	if err != nil {
		return fmt.Errorf("stage changes: %w", err)
	}
	_, err = subWt.Commit(message, &git.CommitOptions{})
	if err != nil {
		if errors.Is(err, git.ErrEmptyCommit) {
			return ErrNoChanges
		}
		return fmt.Errorf("commit: %w", err)
	}
	return nil
}

func Push(repoRoot, bubblePath string) error {
	parentRepo, err := git.PlainOpen(repoRoot)
	if err != nil {
		return fmt.Errorf("open parent repo: %w", err)
	}
	wt, err := parentRepo.Worktree()
	if err != nil {
		return fmt.Errorf("get parent worktree: %w", err)
	}
	sm, err := wt.Submodule(bubblePath)
	if err != nil {
		return fmt.Errorf("get submodule %q: %w", bubblePath, err)
	}
	subRepo, err := sm.Repository()
	if err != nil {
		return fmt.Errorf("open submodule repository: %w", err)
	}
	err = subRepo.Push(&git.PushOptions{})
	if err != nil {
		if errors.Is(err, git.NoErrAlreadyUpToDate) {
			return nil
		}
		return fmt.Errorf("push: %w", err)
	}
	return nil
}

func CommitAndPush(repoRoot, bubblePath, message string) error {
	if err := StageAndCommit(repoRoot, bubblePath, message); err != nil {
		return err
	}
	if err := Push(repoRoot, bubblePath); err != nil {
		return err
	}
	return nil
}
