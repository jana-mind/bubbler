package commands

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/jana-mind/bubbler/internal/git"
	"github.com/jana-mind/bubbler/internal/id"
	"github.com/jana-mind/bubbler/internal/model"
	"github.com/jana-mind/bubbler/internal/store"
	"github.com/spf13/cobra"
)

var errNoTitle = errors.New("title is required")
var errNoSuchColumn = errors.New("no such column")
var errInvalidTag = errors.New("invalid tag")
var errEditorFailed = errors.New("editor failed")

type createOptions struct {
	title       string
	column      string
	tags        []string
	description string
}

func newIssueCmd() *cobra.Command {
	issueCmd := &cobra.Command{
		Use:   "issue",
		Short: "Manage issues",
	}
	issueCmd.AddCommand(newCreateCmd())
	issueCmd.AddCommand(newListCmd())
	issueCmd.AddCommand(newShowCmd())
	return issueCmd
}

func newCreateCmd() *cobra.Command {
	var opts createOptions
	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new issue",
		RunE:  func(cmd *cobra.Command, args []string) error { return runCreate(cmd, &opts) },
	}
	createCmd.Flags().StringVarP(&opts.title, "title", "t", "", "Issue title")
	createCmd.Flags().StringVarP(&opts.column, "column", "c", "", "Starting column (default: first column)")
	createCmd.Flags().StringArrayVar(&opts.tags, "tag", nil, "Tag to apply (repeatable)")
	createCmd.Flags().StringVarP(&opts.description, "description", "d", "", "Issue description")
	return createCmd
}

type listOptions struct {
	all    bool
	column string
	tags   []string
}

func newListCmd() *cobra.Command {
	var opts listOptions
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List issues on the board",
		RunE:  func(cmd *cobra.Command, args []string) error { return runList(cmd, &opts) },
	}
	listCmd.Flags().BoolVar(&opts.all, "all", false, "Include completed issues")
	listCmd.Flags().StringVarP(&opts.column, "column", "c", "", "Filter by column")
	listCmd.Flags().StringArrayVarP(&opts.tags, "tag", "t", nil, "Filter by tag (repeatable, AND logic)")
	return listCmd
}

func runList(cmd *cobra.Command, opts *listOptions) error {
	repoRoot, err := git.FindRepoRoot()
	if err != nil {
		return err
	}

	boardPath := filepath.Join(repoRoot, ".bubble", "default.yaml")
	board, err := store.LoadBoardFileSubmodule(boardPath)
	if err != nil {
		return fmt.Errorf("load board: %w", err)
	}

	issues := board.Issues
	if !opts.all {
		var active []model.IssueSummary
		for _, iss := range issues {
			if iss.Column != "completed" {
				active = append(active, iss)
			}
		}
		issues = active
	}

	if opts.column != "" {
		var filtered []model.IssueSummary
		for _, iss := range issues {
			if iss.Column == opts.column {
				filtered = append(filtered, iss)
			}
		}
		issues = filtered
	}

	if len(opts.tags) > 0 {
		var filtered []model.IssueSummary
	outer:
		for _, iss := range issues {
			tagSet := make(map[string]struct{}, len(iss.Tags))
			for _, t := range iss.Tags {
				tagSet[t] = struct{}{}
			}
			for _, want := range opts.tags {
				if _, ok := tagSet[want]; !ok {
					continue outer
				}
			}
			filtered = append(filtered, iss)
		}
		issues = filtered
	}

	if len(issues) == 0 {
		return nil
	}

	byColumn := make(map[string][]model.IssueSummary)
	for _, iss := range issues {
		byColumn[iss.Column] = append(byColumn[iss.Column], iss)
	}

	for _, col := range board.Board.Columns {
		if issList, ok := byColumn[col.ID]; ok {
			fmt.Printf("%s:\n", col.Label)
			for _, iss := range issList {
				tagStr := ""
				if len(iss.Tags) > 0 {
					tagStr = " [" + joinStrings(iss.Tags, ", ") + "]"
				}
				fmt.Printf("  %s  %s%s\n", iss.ID, iss.Title, tagStr)
			}
			fmt.Println()
		}
	}

	return nil
}

func joinStrings(ss []string, sep string) string {
	if len(ss) == 0 {
		return ""
	}
	res := ss[0]
	for i := 1; i < len(ss); i++ {
		res += sep + ss[i]
	}
	return res
}

func newShowCmd() *cobra.Command {
	showCmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show full issue detail and history",
		Args:  cobra.ExactArgs(1),
		RunE:  func(cmd *cobra.Command, args []string) error { return runShow(cmd, args[0]) },
	}
	return showCmd
}

func runShow(cmd *cobra.Command, issueID string) error {
	repoRoot, err := git.FindRepoRoot()
	if err != nil {
		return err
	}

	issuePath := filepath.Join(repoRoot, ".bubble", "default", issueID+".yaml")
	issue, err := store.LoadIssueFileSubmodule(issuePath)
	if err != nil {
		return fmt.Errorf("load issue %q: %w", issueID, err)
	}

	fmt.Printf("Title:    %s\n", issue.Title)
	fmt.Printf("ID:       %s\n", issue.ID)
	fmt.Printf("Column:   %s\n", issue.Column)
	if len(issue.Tags) > 0 {
		fmt.Printf("Tags:     %s\n", joinStrings(issue.Tags, ", "))
	}
	fmt.Printf("Created:  %s by %s <%s>\n", issue.CreatedAt.Format(time.RFC3339), issue.CreatedBy.Name, issue.CreatedBy.Email)
	fmt.Println("Description:")
	if issue.Description != "" {
		fmt.Println(issue.Description)
	} else {
		fmt.Println("(none)")
	}
	fmt.Println("History:")
	if len(issue.History) == 0 {
		fmt.Println("  (no history)")
	} else {
		for _, entry := range issue.History {
			fmt.Printf("  [%s] %s by %s <%s>\n", entry.At.Format(time.RFC3339), entry.Type, entry.By.Name, entry.By.Email)
			formatHistoryData(entry.Data)
		}
	}

	return nil
}

func formatHistoryData(data model.HistoryData) {
	switch d := data.(type) {
	case model.CreatedEntry:
		fmt.Printf("    title=%q column=%q tags=%v\n", d.Title, d.Column, d.Tags)
	case model.TitleChangedEntry:
		fmt.Printf("    from=%q to=%q\n", d.From, d.To)
	case model.ColumnChangedEntry:
		fmt.Printf("    from=%q to=%q\n", d.From, d.To)
	case model.TagsChangedEntry:
		if len(d.Added) > 0 {
			fmt.Printf("    added: %v\n", d.Added)
		}
		if len(d.Removed) > 0 {
			fmt.Printf("    removed: %v\n", d.Removed)
		}
	case model.DescriptionChangedEntry:
		fmt.Printf("    %q\n", d.Description)
	case model.CommentEntry:
		fmt.Printf("    %s\n", d.Text)
	default:
		fmt.Printf("    (unrecognized entry type)\n")
	}
}

func runCreate(cmd *cobra.Command, opts *createOptions) error {
	repoRoot, err := git.FindRepoRoot()
	if err != nil {
		return err
	}

	boardPath := filepath.Join(repoRoot, ".bubble", "default.yaml")
	board, err := store.LoadBoardFileSubmodule(boardPath)
	if err != nil {
		return fmt.Errorf("load board: %w", err)
	}

	title := opts.title
	if title == "" {
		title, err = promptTitle()
		if err != nil {
			return err
		}
		if title == "" {
			return errNoTitle
		}
	}

	column := opts.column
	if column == "" {
		if len(board.Board.Columns) == 0 {
			return errors.New("board has no columns")
		}
		column = board.Board.Columns[0].ID
	}
	if !board.Board.HasColumn(column) {
		return fmt.Errorf("%w: %q", errNoSuchColumn, column)
	}

	tagSet := board.Board.TagSet()
	for _, tag := range opts.tags {
		if _, ok := tagSet[tag]; !ok {
			return fmt.Errorf("%w: %q", errInvalidTag, tag)
		}
	}

	description := opts.description
	if !cmd.Flags().Changed("description") {
		description, err = editWithEditor("")
		if err != nil {
			if errors.Is(err, errEditorFailed) {
				return err
			}
		}
	}

	identity, err := git.GetIdentity(repoRoot)
	if err != nil {
		return err
	}

	existingIDs := make([]string, len(board.Issues))
	for i, issue := range board.Issues {
		existingIDs[i] = issue.ID
	}

	issueID, err := id.Generate(existingIDs)
	if err != nil {
		return fmt.Errorf("generate ID: %w", err)
	}

	now := time.Now().UTC()

	issueFile := model.IssueFile{
		ID:          issueID,
		Title:       title,
		Column:      column,
		Tags:        opts.tags,
		Description: description,
		CreatedAt:   now,
		CreatedBy: model.Identity{
			Name:  identity.Name,
			Email: identity.Email,
		},
		History: nil,
	}

	createdEntry := model.HistoryEntry{
		Type: "created",
		At:   now,
		By: model.Identity{
			Name:  identity.Name,
			Email: identity.Email,
		},
		Data: model.CreatedEntry{
			Title:  title,
			Column: column,
			Tags:   opts.tags,
		},
	}

	issuePath := filepath.Join(repoRoot, ".bubble", "default", issueID+".yaml")
	if err := store.SaveIssueFileSubmodule(issuePath, issueFile, []model.HistoryEntry{createdEntry}); err != nil {
		return fmt.Errorf("save issue file: %w", err)
	}

	board.Issues = append(board.Issues, model.IssueSummary{
		ID:     issueID,
		Title:  title,
		Column: column,
		Tags:   opts.tags,
	})

	if err := store.SaveBoardFileSubmodule(boardPath, board); err != nil {
		return fmt.Errorf("save board file: %w", err)
	}

	fmt.Println(issueID)
	return nil
}

func promptTitle() (string, error) {
	fmt.Print("Title: ")
	var title string
	fmt.Scanln(&title)
	return title, nil
}

func editWithEditor(content string) (string, error) {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		return "", nil
	}

	tmp, err := os.CreateTemp("", "bubbler-edit-*")
	if err != nil {
		return "", fmt.Errorf("create temp file: %w", err)
	}
	tmpPath := tmp.Name()
	tmp.Close()

	if content != "" {
		if err := os.WriteFile(tmpPath, []byte(content), 0644); err != nil {
			os.Remove(tmpPath)
			return "", fmt.Errorf("write temp file: %w", err)
		}
	}

	err = exec.Command(editor, tmpPath).Run()
	if err != nil {
		os.Remove(tmpPath)
		return "", fmt.Errorf("%w: %q: %v", errEditorFailed, editor, err)
	}

	data, err := os.ReadFile(tmpPath)
	os.Remove(tmpPath)
	if err != nil {
		return "", fmt.Errorf("read edited file: %w", err)
	}

	return string(data), nil
}