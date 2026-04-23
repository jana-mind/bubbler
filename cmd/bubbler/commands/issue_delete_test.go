package commands

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/jana-mind/bubbler/internal/store"
)

func TestIssueDelete(t *testing.T) {
	tmpDir := t.TempDir()

	// Build bubbler binary in temp dir — portable, no hardcoded paths
	binPath := filepath.Join(tmpDir, "bubbler_test_bin")
	buildCmd := exec.Command("go", "build", "-o", binPath, "./cmd/bubbler")
	buildCmd.Dir = "/home/node/.openclaw/workspace/bubbler"
	if out, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("build bubbler binary: %v\n%s", err, string(out))
	}
	repoDir := filepath.Join(tmpDir, "repo")
	if err := os.MkdirAll(repoDir, 0755); err != nil {
		t.Fatal(err)
	}

	gitDir := filepath.Join(repoDir, ".git")
	if err := os.MkdirAll(gitDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(gitDir, "config"), []byte(`[user]
	name = Test
	email = test@test.com
`), 0644); err != nil {
		t.Fatal(err)
	}

	bubbleDir := filepath.Join(repoDir, ".bubble")
	if err := os.MkdirAll(filepath.Join(bubbleDir, "default"), 0755); err != nil {
		t.Fatal(err)
	}
	defaultYaml := `board:
  name: default
  columns:
  - id: waiting
    label: Waiting
  - id: in-progress
    label: In Progress
  tags:
  - bug
  - feature
issues:
  - id: abc123
    title: Test Issue
    column: waiting
    tags:
    - bug
`
	if err := os.WriteFile(filepath.Join(bubbleDir, "default.yaml"), []byte(defaultYaml), 0644); err != nil {
		t.Fatal(err)
	}
	issuePath := filepath.Join(bubbleDir, "default", "abc123.yaml")
	issueYaml := `id: abc123
title: Test Issue
column: waiting
tags:
  - bug
created_at: "2024-11-01T10:22:00Z"
created_by:
  name: Test
  email: test@test.com
history:
  - type: created
    at: "2024-11-01T10:22:00Z"
    by:
      name: Test
      email: test@test.com
    data:
      title: Test Issue
      column: waiting
      tags:
        - bug
`
	if err := os.WriteFile(issuePath, []byte(issueYaml), 0644); err != nil {
		t.Fatal(err)
	}

	t.Run("deletes issue file and removes from board", func(t *testing.T) {
		cmd := exec.Command(binPath, "issue", "delete", "abc123")
		cmd.Dir = repoDir

		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("delete command failed: %v\noutput: %s", err, string(out))
		}

		if _, err := os.Stat(issuePath); !os.IsNotExist(err) {
			t.Error("expected issue file to be deleted")
		}

		boardPath := filepath.Join(bubbleDir, "default.yaml")
		bf, err := store.LoadBoardFile(boardPath)
		if err != nil {
			t.Fatalf("load board file: %v", err)
		}
		for _, issue := range bf.Issues {
			if issue.ID == "abc123" {
				t.Error("expected issue to be removed from board state")
			}
		}
	})

	t.Run("fails for nonexistent issue", func(t *testing.T) {
		cmd := exec.Command(binPath, "issue", "delete", "nonexistent")
		cmd.Dir = repoDir

		out, err := cmd.CombinedOutput()
		if err == nil {
			t.Fatalf("expected error for nonexistent issue, got nil\noutput: %s", string(out))
		}
	})

	t.Run("fails for missing args", func(t *testing.T) {
		cmd := exec.Command(binPath, "issue", "delete")
		cmd.Dir = repoDir

		out, err := cmd.CombinedOutput()
		if err == nil {
			t.Fatalf("expected error for missing args, got nil\noutput: %s", string(out))
		}
	})
}
