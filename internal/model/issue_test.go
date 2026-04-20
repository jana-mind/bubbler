package model

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadIssueFile(t *testing.T) {
	t.Run("loads valid issue file", func(t *testing.T) {
		tmp := t.TempDir()
		path := filepath.Join(tmp, "abc123.yaml")
		data := `id: abc123
title: Test issue
column: waiting
tags:
  - bug
description: |
  This is a test issue.
created_at: "2024-11-01T10:22:00Z"
created_by:
  name: Jane Doe
  email: jane@example.com
history:
  - type: created
    at: "2024-11-01T10:22:00Z"
    by:
      name: Jane Doe
      email: jane@example.com
    data:
      title: Test issue
      column: waiting
      tags:
        - bug
`
		if err := os.WriteFile(path, []byte(data), 0644); err != nil {
			t.Fatal(err)
		}

		f, err := LoadIssueFile(path)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if f.ID != "abc123" {
			t.Errorf("expected ID 'abc123', got %q", f.ID)
		}
		if f.Title != "Test issue" {
			t.Errorf("expected title 'Test issue', got %q", f.Title)
		}
		if f.Column != "waiting" {
			t.Errorf("expected column 'waiting', got %q", f.Column)
		}
		if len(f.Tags) != 1 || f.Tags[0] != "bug" {
			t.Errorf("expected tags [bug], got %v", f.Tags)
		}
		if len(f.History) != 1 {
			t.Fatalf("expected 1 history entry, got %d", len(f.History))
		}
		created, ok := f.History[0].Data.(CreatedEntry)
		if !ok {
			t.Fatalf("expected CreatedEntry, got %T", f.History[0].Data)
		}
		if created.Title != "Test issue" {
			t.Errorf("expected created entry title 'Test issue', got %q", created.Title)
		}
	})

	t.Run("missing file", func(t *testing.T) {
		_, err := LoadIssueFile("/nonexistent/abc123.yaml")
		if err == nil {
			t.Error("expected error for missing file")
		}
	})

	t.Run("malformed yaml", func(t *testing.T) {
		tmp := t.TempDir()
		path := filepath.Join(tmp, "abc123.yaml")
		if err := os.WriteFile(path, []byte("not: [yaml"), 0644); err != nil {
			t.Fatal(err)
		}
		_, err := LoadIssueFile(path)
		if err == nil {
			t.Error("expected error for malformed yaml")
		}
	})

	t.Run("loads all history entry types", func(t *testing.T) {
		tmp := t.TempDir()
		path := filepath.Join(tmp, "abc123.yaml")
		data := `id: abc123
title: Test
column: waiting
tags: []
created_at: "2024-11-01T10:22:00Z"
created_by:
  name: Jane Doe
  email: jane@example.com
history:
  - type: created
    at: "2024-11-01T10:22:00Z"
    by:
      name: Jane Doe
      email: jane@example.com
    data:
      title: Test
      column: waiting
      tags: []
  - type: title_changed
    at: "2024-11-02T09:00:00Z"
    by:
      name: Tom Smith
      email: tom@example.com
    data:
      from: Test
      to: Updated Title
  - type: column_changed
    at: "2024-11-02T09:15:00Z"
    by:
      name: Tom Smith
      email: tom@example.com
    data:
      from: waiting
      to: in-progress
  - type: tags_changed
    at: "2024-11-02T09:16:00Z"
    by:
      name: Tom Smith
      email: tom@example.com
    data:
      added:
        - bug
      removed: []
  - type: description_changed
    at: "2024-11-03T10:00:00Z"
    by:
      name: Jane Doe
      email: jane@example.com
    data:
      description: Updated description
  - type: comment
    at: "2024-11-03T14:00:00Z"
    by:
      name: Jane Doe
      email: jane@example.com
    data:
      text: A comment here
`
		if err := os.WriteFile(path, []byte(data), 0644); err != nil {
			t.Fatal(err)
		}

		f, err := LoadIssueFile(path)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if len(f.History) != 6 {
			t.Fatalf("expected 6 history entries, got %d", len(f.History))
		}

		if _, ok := f.History[1].Data.(TitleChangedEntry); !ok {
			t.Errorf("entry 1: expected TitleChangedEntry, got %T", f.History[1].Data)
		}
		if _, ok := f.History[2].Data.(ColumnChangedEntry); !ok {
			t.Errorf("entry 2: expected ColumnChangedEntry, got %T", f.History[2].Data)
		}
		if _, ok := f.History[3].Data.(TagsChangedEntry); !ok {
			t.Errorf("entry 3: expected TagsChangedEntry, got %T", f.History[3].Data)
		}
		if _, ok := f.History[4].Data.(DescriptionChangedEntry); !ok {
			t.Errorf("entry 4: expected DescriptionChangedEntry, got %T", f.History[4].Data)
		}
		if _, ok := f.History[5].Data.(CommentEntry); !ok {
			t.Errorf("entry 5: expected CommentEntry, got %T", f.History[5].Data)
		}
	})

	t.Run("unknown history entry type", func(t *testing.T) {
		tmp := t.TempDir()
		path := filepath.Join(tmp, "abc123.yaml")
		data := `id: abc123
title: Test
column: waiting
tags: []
created_at: "2024-11-01T10:22:00Z"
created_by:
  name: Jane Doe
  email: jane@example.com
history:
  - type: unknown_type
    at: "2024-11-01T10:22:00Z"
    by:
      name: Jane Doe
      email: jane@example.com
    data:
      foo: bar
`
		if err := os.WriteFile(path, []byte(data), 0644); err != nil {
			t.Fatal(err)
		}

		f, err := LoadIssueFile(path)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if len(f.History) != 1 {
			t.Fatalf("expected 1 history entry, got %d", len(f.History))
		}
		unknown, ok := f.History[0].Data.(UnknownEntry)
		if !ok {
			t.Fatalf("expected UnknownEntry, got %T", f.History[0].Data)
		}
		if unknown.Type != "unknown_type" {
			t.Errorf("expected type 'unknown_type', got %q", unknown.Type)
		}
	})
}

func TestSaveIssueFile(t *testing.T) {
	t.Run("saves and reloads issue file", func(t *testing.T) {
		tmp := t.TempDir()
		path := filepath.Join(tmp, "abc123.yaml")

		issue := IssueFile{
			ID:          "abc123",
			Title:       "Test issue",
			Column:      "in-progress",
			Tags:        []string{"bug"},
			Description: "A test issue",
			CreatedAt:   time.Date(2024, 11, 1, 10, 22, 0, 0, time.UTC),
			CreatedBy: Identity{
				Name:  "Jane Doe",
				Email: "jane@example.com",
			},
			History: []HistoryEntry{
				{
					Type: "created",
					At:   time.Date(2024, 11, 1, 10, 22, 0, 0, time.UTC),
					By: Identity{
						Name:  "Jane Doe",
						Email: "jane@example.com",
					},
					Data: CreatedEntry{
						Title:  "Test issue",
						Column: "waiting",
						Tags:   []string{"bug"},
					},
				},
			},
		}

		newEntry := HistoryEntry{
			Type: "column_changed",
			At:   time.Date(2024, 11, 2, 9, 15, 0, 0, time.UTC),
			By: Identity{
				Name:  "Tom Smith",
				Email: "tom@example.com",
			},
			Data: ColumnChangedEntry{
				From: "waiting",
				To:   "in-progress",
			},
		}

		if err := SaveIssueFile(path, issue, []HistoryEntry{newEntry}); err != nil {
			t.Fatalf("expected no error on save, got %v", err)
		}

		reloaded, err := LoadIssueFile(path)
		if err != nil {
			t.Fatalf("expected no error on reload, got %v", err)
		}
		if reloaded.Title != "Test issue" {
			t.Errorf("expected title 'Test issue', got %q", reloaded.Title)
		}
		if reloaded.Column != "in-progress" {
			t.Errorf("expected column 'in-progress', got %q", reloaded.Column)
		}
		if len(reloaded.History) != 2 {
			t.Fatalf("expected 2 history entries, got %d", len(reloaded.History))
		}
		colChange, ok := reloaded.History[1].Data.(ColumnChangedEntry)
		if !ok {
			t.Fatalf("expected ColumnChangedEntry, got %T", reloaded.History[1].Data)
		}
		if colChange.From != "waiting" || colChange.To != "in-progress" {
			t.Errorf("unexpected column change: from=%q to=%q", colChange.From, colChange.To)
		}
	})

	t.Run("write failure", func(t *testing.T) {
		err := SaveIssueFile("/proc/not writable/abc123.yaml", IssueFile{ID: "abc123"}, nil)
		if err == nil {
			t.Error("expected error when writing to unwritable path")
		}
	})
}

func TestIssueFile_Validate(t *testing.T) {
	board := Board{
		Columns: []Column{
			{ID: "waiting", Label: "Waiting"},
			{ID: "in-progress", Label: "In Progress"},
		},
		Tags: []string{"bug", "feature"},
	}

	t.Run("valid issue", func(t *testing.T) {
		f := IssueFile{
			ID:     "abc123",
			Title:  "Test",
			Column: "waiting",
			Tags:   []string{"bug"},
		}
		if err := f.Validate(board); err != nil {
			t.Errorf("expected valid, got %v", err)
		}
	})

	t.Run("invalid column", func(t *testing.T) {
		f := IssueFile{
			ID:     "abc123",
			Title:  "Test",
			Column: "nonexistent",
			Tags:   []string{},
		}
		err := f.Validate(board)
		if err == nil {
			t.Error("expected error for invalid column")
		}
		colErr, ok := err.(*InvalidColumnError)
		if !ok {
			t.Fatalf("expected InvalidColumnError, got %T", err)
		}
		if colErr.IssueID != "abc123" || colErr.Column != "nonexistent" {
			t.Errorf("unexpected error fields: %+v", colErr)
		}
	})

	t.Run("undefined tag", func(t *testing.T) {
		f := IssueFile{
			ID:     "abc123",
			Title:  "Test",
			Column: "waiting",
			Tags:   []string{"chore"},
		}
		err := f.Validate(board)
		if err == nil {
			t.Error("expected error for undefined tag")
		}
		tagErr, ok := err.(*InvalidTagError)
		if !ok {
			t.Fatalf("expected InvalidTagError, got %T", err)
		}
		if tagErr.IssueID != "abc123" || tagErr.Tag != "chore" {
			t.Errorf("unexpected error fields: %+v", tagErr)
		}
	})
}
