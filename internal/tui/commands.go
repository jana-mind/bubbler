package tui

import (
	"charm.land/bubbletea/v2"

	"github.com/jana-mind/bubbler/internal/git"
	"github.com/jana-mind/bubbler/internal/model"
)

func cmdLoadBoard(boardName string, store Store) tea.Cmd {
	return func() tea.Msg {
		board, err := store.LoadBoard(boardName)
		if err != nil {
			return BoardLoadFailed{Err: err}
		}
		return BoardLoaded{Board: board}
	}
}

func cmdSaveIssue(boardName string, issue model.IssueFile, entries []model.HistoryEntry, store Store) tea.Cmd {
	return func() tea.Msg {
		err := store.SaveIssue(boardName, issue, entries)
		return WriteCompleted{Err: err}
	}
}

func cmdDeleteIssue(boardName, issueID string, store Store) tea.Cmd {
	return func() tea.Msg {
		err := store.DeleteIssue(boardName, issueID)
		return WriteCompleted{Err: err}
	}
}

func cmdGitPull(repoRoot, bubblePath string) tea.Cmd {
	return func() tea.Msg {
		err := git.Pull(repoRoot, bubblePath)
		return WriteCompleted{Err: err}
	}
}

func cmdCommitAndPush(repoRoot, bubblePath, message string) tea.Cmd {
	return func() tea.Msg {
		if err := git.StageAndCommit(repoRoot, bubblePath, message); err != nil {
			return WriteCompleted{Err: err}
		}
		if err := git.Push(repoRoot, bubblePath); err != nil {
			return WriteCompleted{Err: err}
		}
		return WriteCompleted{Err: nil}
	}
}
