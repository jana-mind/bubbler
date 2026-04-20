package model

import "time"

type Board struct {
	Name    string   `yaml:"name"`
	Columns []Column `yaml:"columns"`
	Tags    []string `yaml:"tags"`
}

type Column struct {
	ID    string `yaml:"id"`
	Label string `yaml:"label"`
}

type IssueSummary struct {
	ID     string   `yaml:"id"`
	Title  string   `yaml:"title"`
	Column string   `yaml:"column"`
	Tags   []string `yaml:"tags"`
}

type BoardFile struct {
	Board  Board         `yaml:"board"`
	Issues []IssueSummary `yaml:"issues"`
}

type Identity struct {
	Name  string `yaml:"name"`
	Email string `yaml:"email"`
}

type HistoryEntry struct {
	Type string         `yaml:"type"`
	At   time.Time       `yaml:"at"`
	By   Identity       `yaml:"by"`
	Data map[string]any `yaml:"data"`
}

type IssueFile struct {
	ID          string        `yaml:"id"`
	Title       string        `yaml:"title"`
	Column      string        `yaml:"column"`
	Tags        []string      `yaml:"tags"`
	Description string        `yaml:"description"`
	CreatedAt   time.Time     `yaml:"created_at"`
	CreatedBy   Identity      `yaml:"created_by"`
	History     []HistoryEntry `yaml:"history"`
}

func (b Board) ColumnIDs() []string {
	ids := make([]string, len(b.Columns))
	for i, c := range b.Columns {
		ids[i] = c.ID
	}
	return ids
}

func (b Board) HasColumn(id string) bool {
	for _, c := range b.Columns {
		if c.ID == id {
			return true
		}
	}
	return false
}

func (b Board) TagSet() map[string]struct{} {
	s := make(map[string]struct{}, len(b.Tags))
	for _, t := range b.Tags {
		s[t] = struct{}{}
	}
	return s
}

func (f BoardFile) Validate() error {
	colSet := f.Board.HasColumn
	for _, issue := range f.Issues {
		if !colSet(issue.Column) {
			return &InvalidColumnError{IssueID: issue.ID, Column: issue.Column}
		}
		for _, tag := range issue.Tags {
			if _, ok := f.Board.TagSet()[tag]; !ok {
				return &InvalidTagError{IssueID: issue.ID, Tag: tag}
			}
		}
	}
	return nil
}

type InvalidColumnError struct {
	IssueID string
	Column  string
}

func (e *InvalidColumnError) Error() string {
	return "issue " + e.IssueID + " has invalid column: " + e.Column
}

type InvalidTagError struct {
	IssueID string
	Tag     string
}

func (e *InvalidTagError) Error() string {
	return "issue " + e.IssueID + " has undefined tag: " + e.Tag
}