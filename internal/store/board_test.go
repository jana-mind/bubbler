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
