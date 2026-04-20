package model

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

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

func (f IssueFile) Validate(board Board) error {
	if !board.HasColumn(f.Column) {
		return &InvalidColumnError{IssueID: f.ID, Column: f.Column}
	}
	tagSet := board.TagSet()
	for _, tag := range f.Tags {
		if _, ok := tagSet[tag]; !ok {
			return &InvalidTagError{IssueID: f.ID, Tag: tag}
		}
	}
	return nil
}

type issueHistoryEntry struct {
	Type string        `yaml:"type"`
	At   time.Time     `yaml:"at"`
	By   Identity      `yaml:"by"`
	Data map[string]any `yaml:"data"`
}

type issueFileRaw struct {
	ID          string              `yaml:"id"`
	Title       string              `yaml:"title"`
	Column      string              `yaml:"column"`
	Tags        []string            `yaml:"tags"`
	Description string              `yaml:"description"`
	CreatedAt   time.Time           `yaml:"created_at"`
	CreatedBy   Identity            `yaml:"created_by"`
	History     []issueHistoryEntry `yaml:"history"`
}

func makeHistoryEntry(r issueHistoryEntry) (HistoryEntry, error) {
	switch r.Type {
	case "created":
		var data CreatedEntry
		if err := decodeDataMap(r.Data, &data); err != nil {
			return HistoryEntry{}, fmt.Errorf("created entry: %w", err)
		}
		return HistoryEntry{Type: r.Type, At: r.At, By: r.By, Data: data}, nil
	case "title_changed":
		var data TitleChangedEntry
		if err := decodeDataMap(r.Data, &data); err != nil {
			return HistoryEntry{}, fmt.Errorf("title_changed entry: %w", err)
		}
		return HistoryEntry{Type: r.Type, At: r.At, By: r.By, Data: data}, nil
	case "column_changed":
		var data ColumnChangedEntry
		if err := decodeDataMap(r.Data, &data); err != nil {
			return HistoryEntry{}, fmt.Errorf("column_changed entry: %w", err)
		}
		return HistoryEntry{Type: r.Type, At: r.At, By: r.By, Data: data}, nil
	case "tags_changed":
		var data TagsChangedEntry
		if err := decodeDataMap(r.Data, &data); err != nil {
			return HistoryEntry{}, fmt.Errorf("tags_changed entry: %w", err)
		}
		return HistoryEntry{Type: r.Type, At: r.At, By: r.By, Data: data}, nil
	case "description_changed":
		var data DescriptionChangedEntry
		if err := decodeDataMap(r.Data, &data); err != nil {
			return HistoryEntry{}, fmt.Errorf("description_changed entry: %w", err)
		}
		return HistoryEntry{Type: r.Type, At: r.At, By: r.By, Data: data}, nil
	case "comment":
		var data CommentEntry
		if err := decodeDataMap(r.Data, &data); err != nil {
			return HistoryEntry{}, fmt.Errorf("comment entry: %w", err)
		}
		return HistoryEntry{Type: r.Type, At: r.At, By: r.By, Data: data}, nil
	default:
		return HistoryEntry{Type: r.Type, At: r.At, By: r.By, Data: UnknownEntry{Type: r.Type, Data: r.Data}}, nil
	}
}

func decodeDataMap(data map[string]any, v interface{}) error {
	if data == nil {
		return nil
	}
	bytes, err := yaml.Marshal(data)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(bytes, v)
}

func toMap(v interface{}) map[string]any {
	b, _ := yaml.Marshal(v)
	var m map[string]any
	yaml.Unmarshal(b, &m)
	return m
}

func (h HistoryEntry) toIssueHistoryEntry() issueHistoryEntry {
	var data map[string]any
	switch d := h.Data.(type) {
	case CreatedEntry:
		data = toMap(d)
	case TitleChangedEntry:
		data = toMap(d)
	case ColumnChangedEntry:
		data = toMap(d)
	case TagsChangedEntry:
		data = toMap(d)
	case DescriptionChangedEntry:
		data = toMap(d)
	case CommentEntry:
		data = toMap(d)
	case UnknownEntry:
		data = d.Data
	default:
		data = nil
	}
	return issueHistoryEntry{Type: h.Type, At: h.At, By: h.By, Data: data}
}

func LoadIssueFile(path string) (IssueFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return IssueFile{}, fmt.Errorf("read issue file: %w", err)
	}

	var raw issueFileRaw
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return IssueFile{}, fmt.Errorf("parse issue file: %w", err)
	}

	history := make([]HistoryEntry, len(raw.History))
	for i, r := range raw.History {
		entry, err := makeHistoryEntry(r)
		if err != nil {
			return IssueFile{}, fmt.Errorf("entry %d: %w", i, err)
		}
		history[i] = entry
	}

	return IssueFile{
		ID:          raw.ID,
		Title:       raw.Title,
		Column:      raw.Column,
		Tags:        raw.Tags,
		Description: raw.Description,
		CreatedAt:   raw.CreatedAt,
		CreatedBy:   raw.CreatedBy,
		History:     history,
	}, nil
}

func SaveIssueFile(path string, issue IssueFile, newEntries []HistoryEntry) error {
	allEntries := make([]HistoryEntry, len(issue.History)+len(newEntries))
	copy(allEntries, issue.History)
	copy(allEntries[len(issue.History):], newEntries)

	entries := make([]issueHistoryEntry, len(allEntries))
	for i, h := range allEntries {
		entries[i] = h.toIssueHistoryEntry()
	}

	raw := issueFileRaw{
		ID:          issue.ID,
		Title:       issue.Title,
		Column:      issue.Column,
		Tags:        issue.Tags,
		Description: issue.Description,
		CreatedAt:   issue.CreatedAt,
		CreatedBy:   issue.CreatedBy,
		History:     entries,
	}

	data, err := yaml.Marshal(raw)
	if err != nil {
		return fmt.Errorf("marshal issue file: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("write issue file: %w", err)
	}

	return nil
}