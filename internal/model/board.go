package model

import (
	"strconv"
	"time"
)

type Board struct {
	Name    string   `yaml:"name"`
	Columns []Column `yaml:"columns"`
	Tags    []string `yaml:"tags"`
	NextID  int      `yaml:"next_id"`
}

func (b *Board) TakeNextID() string {
	id := b.NextID
	b.NextID++
	return strconv.Itoa(id)
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
	Board  Board          `yaml:"board"`
	Issues []IssueSummary `yaml:"issues"`
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

type Identity struct {
	Name  string `yaml:"name"`
	Email string `yaml:"email"`
}

type HistoryEntry struct {
	Type string    `yaml:"type"`
	At   time.Time `yaml:"at"`
	By   Identity  `yaml:"by"`
	Data HistoryData
}

type HistoryData interface {
	HistoryType() string
}

type CreatedEntry struct {
	Title  string   `yaml:"title"`
	Column string   `yaml:"column"`
	Tags   []string `yaml:"tags"`
}

func (e CreatedEntry) HistoryType() string { return "created" }

type TitleChangedEntry struct {
	From string `yaml:"from"`
	To   string `yaml:"to"`
}

func (e TitleChangedEntry) HistoryType() string { return "title_changed" }

type ColumnChangedEntry struct {
	From string `yaml:"from"`
	To   string `yaml:"to"`
}

func (e ColumnChangedEntry) HistoryType() string { return "column_changed" }

type TagsChangedEntry struct {
	Added   []string `yaml:"added"`
	Removed []string `yaml:"removed"`
}

func (e TagsChangedEntry) HistoryType() string { return "tags_changed" }

type DescriptionChangedEntry struct {
	Description string `yaml:"description"`
}

func (e DescriptionChangedEntry) HistoryType() string { return "description_changed" }

type CommentEntry struct {
	Text string `yaml:"text"`
}

func (e CommentEntry) HistoryType() string { return "comment" }

type UnknownEntry struct {
	Type string         `yaml:"type"`
	Data map[string]any `yaml:"data"`
}

func (e UnknownEntry) HistoryType() string { return e.Type }
