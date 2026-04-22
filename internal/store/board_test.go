package store

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jana-mind/bubbler/internal/model"
)

func TestLoadBoardFile(t *testing.T) {
	t.Run("loads valid board file", func(t *testing.T) {
		tmp := t.TempDir()
		path := filepath.Join(tmp, "board.yaml")
		data := `board:
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
    title: Test issue
    column: waiting
    tags:
    - bug
`
		if err := os.WriteFile(path, []byte(data), 0644); err != nil {
			t.Fatal(err)
		}

		bf, err := LoadBoardFile(path)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if bf.Board.Name != "default" {
			t.Errorf("expected board name 'default', got %q", bf.Board.Name)
		}
		if len(bf.Board.Columns) != 2 {
			t.Fatalf("expected 2 columns, got %d", len(bf.Board.Columns))
		}
		if len(bf.Issues) != 1 {
			t.Fatalf("expected 1 issue, got %d", len(bf.Issues))
		}
		if bf.Issues[0].ID != "abc123" {
			t.Errorf("expected issue ID 'abc123', got %q", bf.Issues[0].ID)
		}
	})

	t.Run("missing file", func(t *testing.T) {
		_, err := LoadBoardFile("/nonexistent/path/board.yaml")
		if err == nil {
			t.Error("expected error for missing file")
		}
	})

	t.Run("malformed yaml", func(t *testing.T) {
		tmp := t.TempDir()
		path := filepath.Join(tmp, "board.yaml")
		if err := os.WriteFile(path, []byte("not: [yaml"), 0644); err != nil {
			t.Fatal(err)
		}
		_, err := LoadBoardFile(path)
		if err == nil {
			t.Error("expected error for malformed yaml")
		}
	})

	t.Run("invalid column", func(t *testing.T) {
		tmp := t.TempDir()
		path := filepath.Join(tmp, "board.yaml")
		data := `board:
  name: default
  columns:
  - id: waiting
    label: Waiting
  tags: []
issues:
  - id: abc123
    title: Test
    column: nonexistent
    tags: []
`
		if err := os.WriteFile(path, []byte(data), 0644); err != nil {
			t.Fatal(err)
		}
		_, err := LoadBoardFile(path)
		if err == nil {
			t.Error("expected error for invalid column")
		}
	})

	t.Run("undefined tag", func(t *testing.T) {
		tmp := t.TempDir()
		path := filepath.Join(tmp, "board.yaml")
		data := `board:
  name: default
  columns:
  - id: waiting
    label: Waiting
  tags:
  - bug
issues:
  - id: abc123
    title: Test
    column: waiting
    tags:
    - feature
`
		if err := os.WriteFile(path, []byte(data), 0644); err != nil {
			t.Fatal(err)
		}
		_, err := LoadBoardFile(path)
		if err == nil {
			t.Error("expected error for undefined tag")
		}
	})
}

func TestSaveBoardFile(t *testing.T) {
	t.Run("saves and reloads board file", func(t *testing.T) {
		tmp := t.TempDir()
		path := filepath.Join(tmp, "board.yaml")
		bf := model.BoardFile{
			Board: model.Board{
				Name: "default",
				Columns: []model.Column{
					{ID: "waiting", Label: "Waiting"},
					{ID: "in-progress", Label: "In Progress"},
				},
				Tags: []string{"bug", "feature"},
			},
			Issues: []model.IssueSummary{
				{ID: "abc123", Title: "Test issue", Column: "waiting", Tags: []string{"bug"}},
			},
		}

		if err := SaveBoardFile(path, bf); err != nil {
			t.Fatalf("expected no error on save, got %v", err)
		}

		reloaded, err := LoadBoardFile(path)
		if err != nil {
			t.Fatalf("expected no error on reload, got %v", err)
		}
		if reloaded.Board.Name != "default" {
			t.Errorf("expected board name 'default', got %q", reloaded.Board.Name)
		}
		if len(reloaded.Issues) != 1 {
			t.Fatalf("expected 1 issue, got %d", len(reloaded.Issues))
		}
	})

	t.Run("write failure", func(t *testing.T) {
		err := SaveBoardFile("/proc/not writable/board.yaml", model.BoardFile{})
		if err == nil {
			t.Error("expected error when writing to unwritable path")
		}
	})
}

func TestDeleteBoardFileSubmodule(t *testing.T) {
	t.Run("deletes board file and issues directory", func(t *testing.T) {
		tmp := t.TempDir()
		bubblePath := filepath.Join(tmp, ".bubble")
		boardFile := filepath.Join(bubblePath, "myboard.yaml")
		issuesDir := filepath.Join(bubblePath, "myboard")

		bf := model.BoardFile{
			Board: model.Board{
				Name: "myboard",
				Columns: []model.Column{
					{ID: "waiting", Label: "Waiting"},
				},
				Tags: []string{"bug"},
			},
		}

		if err := os.MkdirAll(issuesDir, 0755); err != nil {
			t.Fatal(err)
		}
		if err := SaveBoardFile(boardFile, bf); err != nil {
			t.Fatal(err)
		}

		if err := DeleteBoardFileSubmodule(boardFile, issuesDir, "myboard"); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if _, err := os.Stat(boardFile); !os.IsNotExist(err) {
			t.Errorf("expected board file to be deleted, stat returned: %v", err)
		}
		if _, err := os.Stat(issuesDir); !os.IsNotExist(err) {
			t.Errorf("expected issues directory to be deleted, stat returned: %v", err)
		}
	})

	t.Run("missing board file returns error", func(t *testing.T) {
		tmp := t.TempDir()
		bubblePath := filepath.Join(tmp, ".bubble")
		boardFile := filepath.Join(bubblePath, "nonexistent.yaml")
		issuesDir := filepath.Join(bubblePath, "nonexistent")

		err := DeleteBoardFileSubmodule(boardFile, issuesDir, "nonexistent")
		if err == nil {
			t.Error("expected error when board file does not exist")
		}
	})
}

func TestLockBoard(t *testing.T) {
	t.Run("acquires and releases lock", func(t *testing.T) {
		tmp := t.TempDir()
		bubblePath := filepath.Join(tmp, ".bubble")
		if err := os.MkdirAll(bubblePath, 0755); err != nil {
			t.Fatal(err)
		}
		boardPath := filepath.Join(bubblePath, "testboard.yaml")

		lock, err := LockBoard(boardPath)
		if err != nil {
			t.Fatalf("expected no error acquiring lock, got %v", err)
		}
		defer lock.Unlock()

		lockPath := filepath.Join(bubblePath, "__testboard.lock")
		if _, err := os.Stat(lockPath); err != nil {
			t.Errorf("expected lock file to exist, got %v", err)
		}
	})

	t.Run("fails when already locked", func(t *testing.T) {
		tmp := t.TempDir()
		bubblePath := filepath.Join(tmp, ".bubble")
		if err := os.MkdirAll(bubblePath, 0755); err != nil {
			t.Fatal(err)
		}
		boardPath := filepath.Join(bubblePath, "testboard.yaml")

		lock1, err := LockBoard(boardPath)
		if err != nil {
			t.Fatalf("expected no error acquiring first lock, got %v", err)
		}
		defer lock1.Unlock()

		_, err = LockBoard(boardPath)
		if err == nil {
			t.Error("expected error when board is already locked")
		}
	})

	t.Run("unlock removes lock file", func(t *testing.T) {
		tmp := t.TempDir()
		bubblePath := filepath.Join(tmp, ".bubble")
		if err := os.MkdirAll(bubblePath, 0755); err != nil {
			t.Fatal(err)
		}
		boardPath := filepath.Join(bubblePath, "testboard.yaml")

		lock, err := LockBoard(boardPath)
		if err != nil {
			t.Fatalf("expected no error acquiring lock, got %v", err)
		}

		if err := lock.Unlock(); err != nil {
			t.Fatalf("expected no error releasing lock, got %v", err)
		}

		lockPath := filepath.Join(bubblePath, "__testboard.lock")
		if _, err := os.Stat(lockPath); !os.IsNotExist(err) {
			t.Errorf("expected lock file to be removed, stat returned: %v", err)
		}
	})
}

func TestLoadBoardFileForUpdate(t *testing.T) {
	t.Run("loads board and lock together", func(t *testing.T) {
		tmp := t.TempDir()
		bubblePath := filepath.Join(tmp, ".bubble")
		if err := os.MkdirAll(bubblePath, 0755); err != nil {
			t.Fatal(err)
		}
		boardPath := filepath.Join(bubblePath, "testboard.yaml")
		bf := model.BoardFile{
			Board: model.Board{
				Name: "testboard",
				Columns: []model.Column{
					{ID: "waiting", Label: "Waiting"},
				},
				Tags: []string{"bug"},
			},
		}
		if err := SaveBoardFile(boardPath, bf); err != nil {
			t.Fatal(err)
		}

		loaded, lock, err := LoadBoardFileForUpdate(boardPath)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		defer lock.Unlock()

		if loaded.Board.Name != "testboard" {
			t.Errorf("expected board name 'testboard', got %q", loaded.Board.Name)
		}
	})

	t.Run("returns error when already locked", func(t *testing.T) {
		tmp := t.TempDir()
		bubblePath := filepath.Join(tmp, ".bubble")
		if err := os.MkdirAll(bubblePath, 0755); err != nil {
			t.Fatal(err)
		}
		boardPath := filepath.Join(bubblePath, "testboard.yaml")
		bf := model.BoardFile{
			Board: model.Board{
				Name: "testboard",
				Columns: []model.Column{
					{ID: "waiting", Label: "Waiting"},
				},
				Tags: []string{"bug"},
			},
		}
		if err := SaveBoardFile(boardPath, bf); err != nil {
			t.Fatal(err)
		}

		_, lock1, err := LoadBoardFileForUpdate(boardPath)
		if err != nil {
			t.Fatalf("expected no error for first lock, got %v", err)
		}
		defer lock1.Unlock()

		_, _, err = LoadBoardFileForUpdate(boardPath)
		if err == nil {
			t.Error("expected error when board is already locked")
		}
	})
}

func TestSaveBoardFileForUpdate(t *testing.T) {
	t.Run("saves and unlocks", func(t *testing.T) {
		tmp := t.TempDir()
		bubblePath := filepath.Join(tmp, ".bubble")
		if err := os.MkdirAll(bubblePath, 0755); err != nil {
			t.Fatal(err)
		}
		boardPath := filepath.Join(bubblePath, "testboard.yaml")
		bf := model.BoardFile{
			Board: model.Board{
				Name: "testboard",
				Columns: []model.Column{
					{ID: "waiting", Label: "Waiting"},
				},
				Tags: []string{"bug"},
			},
		}
		if err := SaveBoardFile(boardPath, bf); err != nil {
			t.Fatal(err)
		}

		_, lock, err := LoadBoardFileForUpdate(boardPath)
		if err != nil {
			t.Fatal(err)
		}

		bf.Board.Name = "updated"
		if err := SaveBoardFileForUpdate(boardPath, bf, lock); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		reloaded, err := LoadBoardFile(boardPath)
		if err != nil {
			t.Fatal(err)
		}
		if reloaded.Board.Name != "updated" {
			t.Errorf("expected updated name, got %q", reloaded.Board.Name)
		}

		lockPath := filepath.Join(bubblePath, "__testboard.lock")
		if _, err := os.Stat(lockPath); !os.IsNotExist(err) {
			t.Errorf("expected lock to be released, file still exists")
		}
	})
}
