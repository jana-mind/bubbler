package model

import (
	"testing"
)

func TestBoard_ColumnIDs(t *testing.T) {
	b := Board{
		Columns: []Column{
			{ID: "waiting", Label: "Waiting"},
			{ID: "in-progress", Label: "In Progress"},
			{ID: "completed", Label: "Completed"},
		},
	}
	ids := b.ColumnIDs()
	if len(ids) != 3 {
		t.Fatalf("expected 3 columns, got %d", len(ids))
	}
	if ids[0] != "waiting" || ids[1] != "in-progress" || ids[2] != "completed" {
		t.Fatalf("unexpected column IDs: %v", ids)
	}
}

func TestBoard_HasColumn(t *testing.T) {
	b := Board{
		Columns: []Column{
			{ID: "waiting", Label: "Waiting"},
			{ID: "in-progress", Label: "In Progress"},
		},
	}
	if !b.HasColumn("waiting") {
		t.Error("expected HasColumn(waiting) to be true")
	}
	if !b.HasColumn("in-progress") {
		t.Error("expected HasColumn(in-progress) to be true")
	}
	if b.HasColumn("completed") {
		t.Error("expected HasColumn(completed) to be false")
	}
}

func TestBoard_TagSet(t *testing.T) {
	b := Board{
		Tags: []string{"bug", "feature", "docs"},
	}
	s := b.TagSet()
	if _, ok := s["bug"]; !ok {
		t.Error("expected bug in tag set")
	}
	if _, ok := s["feature"]; !ok {
		t.Error("expected feature in tag set")
	}
	if _, ok := s["chore"]; ok {
		t.Error("did not expect chore in tag set")
	}
}

func TestBoardFile_Validate(t *testing.T) {
	t.Run("valid board", func(t *testing.T) {
		bf := BoardFile{
			Board: Board{
				Columns: []Column{{ID: "col1", Label: "Col 1"}},
				Tags:    []string{"bug"},
			},
			Issues: []IssueSummary{
				{ID: "abc123", Title: "Test", Column: "col1", Tags: []string{"bug"}},
			},
		}
		if err := bf.Validate(); err != nil {
			t.Errorf("expected valid, got %v", err)
		}
	})

	t.Run("issue with invalid column", func(t *testing.T) {
		bf := BoardFile{
			Board: Board{
				Columns: []Column{{ID: "col1", Label: "Col 1"}},
				Tags:    []string{},
			},
			Issues: []IssueSummary{
				{ID: "abc123", Title: "Test", Column: "col2", Tags: []string{}},
			},
		}
		err := bf.Validate()
		if err == nil {
			t.Error("expected error for invalid column")
		}
		colErr, ok := err.(*InvalidColumnError)
		if !ok {
			t.Fatalf("expected InvalidColumnError, got %T", err)
		}
		if colErr.IssueID != "abc123" || colErr.Column != "col2" {
			t.Errorf("unexpected error fields: %+v", colErr)
		}
	})

	t.Run("issue with undefined tag", func(t *testing.T) {
		bf := BoardFile{
			Board: Board{
				Columns: []Column{{ID: "col1", Label: "Col 1"}},
				Tags:    []string{"bug"},
			},
			Issues: []IssueSummary{
				{ID: "abc123", Title: "Test", Column: "col1", Tags: []string{"feature"}},
			},
		}
		err := bf.Validate()
		if err == nil {
			t.Error("expected error for undefined tag")
		}
		tagErr, ok := err.(*InvalidTagError)
		if !ok {
			t.Fatalf("expected InvalidTagError, got %T", err)
		}
		if tagErr.IssueID != "abc123" || tagErr.Tag != "feature" {
			t.Errorf("unexpected error fields: %+v", tagErr)
		}
	})
}