package commands

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/jana-mind/bubbler/internal/git"
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
	issueCmd.AddCommand(newEditCmd())
	issueCmd.AddCommand(newMoveCmd())
	issueCmd.AddCommand(newIssueTagCmd())
	issueCmd.AddCommand(newUntagCmd())
	issueCmd.AddCommand(newCommentCmd())
	issueCmd.AddCommand(newHistoryCmd())
	issueCmd.AddCommand(newIssueDeleteCmd())
	return issueCmd
}

func getIssueIDs(cmd *cobra.Command) ([]string, error) {
	board := boardFlag(cmd)
	boardPath, _, err := resolveBoardPaths(board)
	if err != nil {
		return nil, err
	}
	bf, err := store.LoadBoardFileSubmodule(boardPath)
	if err != nil {
		return nil, err
	}
	ids := make([]string, len(bf.Issues))
	for i, iss := range bf.Issues {
		ids[i] = iss.ID
	}
	return ids, nil
}

func getColumnIDs(cmd *cobra.Command) ([]string, error) {
	board := boardFlag(cmd)
	boardPath, _, err := resolveBoardPaths(board)
	if err != nil {
		return nil, err
	}
	bf, err := store.LoadBoardFileSubmodule(boardPath)
	if err != nil {
		return nil, err
	}
	return bf.Board.ColumnIDs(), nil
}

func getTagNames(cmd *cobra.Command) ([]string, error) {
	board := boardFlag(cmd)
	boardPath, _, err := resolveBoardPaths(board)
	if err != nil {
		return nil, err
	}
	bf, err := store.LoadBoardFileSubmodule(boardPath)
	if err != nil {
		return nil, err
	}
	return bf.Board.Tags, nil
}

func issueIDCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) != 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	ids, err := getIssueIDs(cmd)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	return ids, cobra.ShellCompDirectiveNoFileComp
}

func columnFlagCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	ids, err := getColumnIDs(cmd)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	return ids, cobra.ShellCompDirectiveNoFileComp
}

func tagFlagCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	tags, err := getTagNames(cmd)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	return tags, cobra.ShellCompDirectiveNoFileComp
}

func newCreateCmd() *cobra.Command {
	var opts createOptions
	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new issue",
		Args:  cobra.NoArgs,
		RunE:  func(cmd *cobra.Command, args []string) error { return runCreate(cmd, &opts) },
	}
	createCmd.Flags().StringVarP(&opts.title, "title", "t", "", "Issue title")
	createCmd.Flags().StringVarP(&opts.column, "column", "c", "", "Starting column (default: first column)")
	createCmd.Flags().StringArrayVar(&opts.tags, "tag", nil, "Tag to apply (repeatable)")
	createCmd.Flags().StringVarP(&opts.description, "description", "d", "", "Issue description")
	createCmd.RegisterFlagCompletionFunc("column", columnFlagCompletion)
	createCmd.RegisterFlagCompletionFunc("tag", tagFlagCompletion)
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
	listCmd.RegisterFlagCompletionFunc("column", columnFlagCompletion)
	listCmd.RegisterFlagCompletionFunc("tag", tagFlagCompletion)
	return listCmd
}

func runList(cmd *cobra.Command, opts *listOptions) error {
	board := boardFlag(cmd)
	boardPath, _, err := resolveBoardPaths(board)
	if err != nil {
		return err
	}

	bf, err := store.LoadBoardFileSubmodule(boardPath)
	if err != nil {
		return fmt.Errorf("load board: %w", err)
	}

	issues := bf.Issues
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

	for _, col := range bf.Board.Columns {
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
		Use:               "show <id>",
		Short:             "Show full issue detail and history",
		Args:              cobra.ExactArgs(1),
		RunE:              func(cmd *cobra.Command, args []string) error { return runShow(cmd, args[0]) },
		ValidArgsFunction: issueIDCompletion,
	}
	return showCmd
}

func runShow(cmd *cobra.Command, issueID string) error {
	board := boardFlag(cmd)
	_, issuesDir, err := resolveBoardPaths(board)
	if err != nil {
		return err
	}

	issuePath := filepath.Join(issuesDir, issueID+".yaml")
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
	board := boardFlag(cmd)
	boardPath, issuesDir, err := resolveBoardPaths(board)
	if err != nil {
		return err
	}

	bf, err := store.LoadBoardFileSubmodule(boardPath)
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
		if len(bf.Board.Columns) == 0 {
			return errors.New("board has no columns")
		}
		column = bf.Board.Columns[0].ID
	}
	if !bf.Board.HasColumn(column) {
		return fmt.Errorf("%w: %q", errNoSuchColumn, column)
	}

	tagSet := bf.Board.TagSet()
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

	identity, err := git.GetIdentityFromRepoRoot()
	if err != nil {
		return err
	}

	if bf.Board.NextID == 0 {
		maxID := 0
		for _, iss := range bf.Issues {
			if n, err := strconv.Atoi(iss.ID); err == nil && n > maxID {
				maxID = n
			}
		}
		bf.Board.NextID = maxID + 1
	}

	issueID := bf.Board.TakeNextID()

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

	issuePath := filepath.Join(issuesDir, issueID+".yaml")
	if err := store.SaveIssueFileSubmodule(issuePath, issueFile, []model.HistoryEntry{createdEntry}); err != nil {
		return fmt.Errorf("save issue file: %w", err)
	}

	bf.Issues = append(bf.Issues, model.IssueSummary{
		ID:     issueID,
		Title:  title,
		Column: column,
		Tags:   opts.tags,
	})

	if err := store.SaveBoardFileSubmodule(boardPath, bf); err != nil {
		return fmt.Errorf("save board file: %w", err)
	}

	fmt.Println(issueID)
	return nil
}

func promptTitle() (string, error) {
	fmt.Print("Title: ")
	reader := bufio.NewReader(os.Stdin)
	title, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimRight(title, "\r\n"), nil
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

func newEditCmd() *cobra.Command {
	var title, description string
	editCmd := &cobra.Command{
		Use:   "edit <id>",
		Short: "Edit an issue's title or description",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runEdit(cmd, args[0], title, description)
		},
		ValidArgsFunction: issueIDCompletion,
	}
	editCmd.Flags().StringVarP(&title, "title", "t", "", "New title")
	editCmd.Flags().StringVarP(&description, "description", "d", "", "New description")
	return editCmd
}

func runEdit(cmd *cobra.Command, issueID, newTitle, newDesc string) error {
	board := boardFlag(cmd)
	boardPath, issuesDir, err := resolveBoardPaths(board)
	if err != nil {
		return err
	}

	issuePath := filepath.Join(issuesDir, issueID+".yaml")
	issue, err := store.LoadIssueFileSubmodule(issuePath)
	if err != nil {
		return fmt.Errorf("load issue %q: %w", issueID, err)
	}

	bf, err := store.LoadBoardFileSubmodule(boardPath)
	if err != nil {
		return fmt.Errorf("load board: %w", err)
	}

	identity, err := git.GetIdentityFromRepoRoot()
	if err != nil {
		return err
	}
	now := time.Now().UTC()

	hasTitle := cmd.Flags().Changed("title")
	hasDesc := cmd.Flags().Changed("description")

	if !hasTitle && !hasDesc {
		prompt := issue.Title + "\n\n" + issue.Description
		content, err := editWithEditor(prompt)
		if err != nil {
			return err
		}
		content = strings.TrimRight(content, "\n")
		lines := strings.SplitN(content, "\n", 2)
		newTitle = lines[0]
		newDesc = ""
		if len(lines) > 1 {
			newDesc = strings.TrimLeft(lines[1], "\n")
		}
		hasTitle = newTitle != issue.Title
		hasDesc = newDesc != issue.Description
		if !hasTitle && !hasDesc {
			return nil
		}
	}

	var newEntries []model.HistoryEntry

	if hasTitle {
		oldTitle := issue.Title
		issue.Title = newTitle
		for i := range bf.Issues {
			if bf.Issues[i].ID == issueID {
				bf.Issues[i].Title = newTitle
				break
			}
		}
		newEntries = append(newEntries, model.HistoryEntry{
			Type: "title_changed",
			At:   now,
			By:   model.Identity{Name: identity.Name, Email: identity.Email},
			Data: model.TitleChangedEntry{From: oldTitle, To: newTitle},
		})
	}

	if hasDesc {
		issue.Description = newDesc
		newEntries = append(newEntries, model.HistoryEntry{
			Type: "description_changed",
			At:   now,
			By:   model.Identity{Name: identity.Name, Email: identity.Email},
			Data: model.DescriptionChangedEntry{Description: newDesc},
		})
	}

	if err := store.SaveIssueFileSubmodule(issuePath, issue, newEntries); err != nil {
		return fmt.Errorf("save issue file: %w", err)
	}
	if hasTitle {
		if err := store.SaveBoardFileSubmodule(boardPath, bf); err != nil {
			return fmt.Errorf("save board file: %w", err)
		}
	}
	return nil
}

func newMoveCmd() *cobra.Command {
	moveCmd := &cobra.Command{
		Use:               "move <id> [column]",
		Short:             "Move an issue to a different column",
		Args:              cobra.RangeArgs(1, 2),
		RunE:              runMove,
		ValidArgsFunction: issueIDCompletion,
	}
	return moveCmd
}

func runMove(cmd *cobra.Command, args []string) error {
	board := boardFlag(cmd)
	boardPath, issuesDir, err := resolveBoardPaths(board)
	if err != nil {
		return err
	}

	issueID := args[0]

	bf, err := store.LoadBoardFileSubmodule(boardPath)
	if err != nil {
		return fmt.Errorf("load board: %w", err)
	}

	if len(args) == 1 {
		var cols []string
		for _, c := range bf.Board.Columns {
			cols = append(cols, fmt.Sprintf("%s (%s)", c.ID, c.Label))
		}
		return fmt.Errorf("Error: you need to define a column. Available columns are: %s", strings.Join(cols, ", "))
	}

	targetCol := args[1]

	if !bf.Board.HasColumn(targetCol) {
		return fmt.Errorf("%w: %q", errNoSuchColumn, targetCol)
	}

	issuePath := filepath.Join(issuesDir, issueID+".yaml")
	issue, err := store.LoadIssueFileSubmodule(issuePath)
	if err != nil {
		return fmt.Errorf("load issue %q: %w", issueID, err)
	}

	fromCol := issue.Column
	if fromCol == targetCol {
		return nil
	}

	identity, err := git.GetIdentityFromRepoRoot()
	if err != nil {
		return err
	}
	now := time.Now().UTC()

	issue.Column = targetCol
	newEntry := model.HistoryEntry{
		Type: "column_changed",
		At:   now,
		By:   model.Identity{Name: identity.Name, Email: identity.Email},
		Data: model.ColumnChangedEntry{From: fromCol, To: targetCol},
	}

	if err := store.SaveIssueFileSubmodule(issuePath, issue, []model.HistoryEntry{newEntry}); err != nil {
		return fmt.Errorf("save issue file: %w", err)
	}

	for i := range bf.Issues {
		if bf.Issues[i].ID == issueID {
			bf.Issues[i].Column = targetCol
			break
		}
	}

	if err := store.SaveBoardFileSubmodule(boardPath, bf); err != nil {
		return fmt.Errorf("save board file: %w", err)
	}

	return nil
}

func newIssueTagCmd() *cobra.Command {
	tagCmd := &cobra.Command{
		Use:               "tag <id> <tag>",
		Short:             "Add a tag to an issue",
		Args:              cobra.ExactArgs(2),
		RunE:              runTag,
		ValidArgsFunction: issueIDCompletion,
	}
	return tagCmd
}

func runTag(cmd *cobra.Command, args []string) error {
	board := boardFlag(cmd)
	boardPath, issuesDir, err := resolveBoardPaths(board)
	if err != nil {
		return err
	}

	issueID := args[0]
	tagName := args[1]

	bf, err := store.LoadBoardFileSubmodule(boardPath)
	if err != nil {
		return fmt.Errorf("load board: %w", err)
	}

	tagSet := bf.Board.TagSet()
	if _, ok := tagSet[tagName]; !ok {
		return fmt.Errorf("%w: %q", errInvalidTag, tagName)
	}

	issuePath := filepath.Join(issuesDir, issueID+".yaml")
	issue, err := store.LoadIssueFileSubmodule(issuePath)
	if err != nil {
		return fmt.Errorf("load issue %q: %w", issueID, err)
	}

	for _, t := range issue.Tags {
		if t == tagName {
			return nil
		}
	}

	identity, err := git.GetIdentityFromRepoRoot()
	if err != nil {
		return err
	}
	now := time.Now().UTC()

	issue.Tags = append(issue.Tags, tagName)
	newEntry := model.HistoryEntry{
		Type: "tags_changed",
		At:   now,
		By:   model.Identity{Name: identity.Name, Email: identity.Email},
		Data: model.TagsChangedEntry{Added: []string{tagName}, Removed: nil},
	}

	if err := store.SaveIssueFileSubmodule(issuePath, issue, []model.HistoryEntry{newEntry}); err != nil {
		return fmt.Errorf("save issue file: %w", err)
	}

	for i := range bf.Issues {
		if bf.Issues[i].ID == issueID {
			bf.Issues[i].Tags = append(bf.Issues[i].Tags, tagName)
			break
		}
	}

	if err := store.SaveBoardFileSubmodule(boardPath, bf); err != nil {
		return fmt.Errorf("save board file: %w", err)
	}

	return nil
}

func newUntagCmd() *cobra.Command {
	untagCmd := &cobra.Command{
		Use:               "untag <id> <tag>",
		Short:             "Remove a tag from an issue",
		Args:              cobra.ExactArgs(2),
		RunE:              runUntag,
		ValidArgsFunction: issueIDCompletion,
	}
	return untagCmd
}

func runUntag(cmd *cobra.Command, args []string) error {
	board := boardFlag(cmd)
	boardPath, issuesDir, err := resolveBoardPaths(board)
	if err != nil {
		return err
	}

	issueID := args[0]
	tagName := args[1]

	bf, err := store.LoadBoardFileSubmodule(boardPath)
	if err != nil {
		return fmt.Errorf("load board: %w", err)
	}

	issuePath := filepath.Join(issuesDir, issueID+".yaml")
	issue, err := store.LoadIssueFileSubmodule(issuePath)
	if err != nil {
		return fmt.Errorf("load issue %q: %w", issueID, err)
	}

	found := false
	for _, t := range issue.Tags {
		if t == tagName {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("issue %q does not have tag %q", issueID, tagName)
	}

	identity, err := git.GetIdentityFromRepoRoot()
	if err != nil {
		return err
	}
	now := time.Now().UTC()

	var newTags []string
	for _, t := range issue.Tags {
		if t != tagName {
			newTags = append(newTags, t)
		}
	}
	issue.Tags = newTags

	newEntry := model.HistoryEntry{
		Type: "tags_changed",
		At:   now,
		By:   model.Identity{Name: identity.Name, Email: identity.Email},
		Data: model.TagsChangedEntry{Added: nil, Removed: []string{tagName}},
	}

	if err := store.SaveIssueFileSubmodule(issuePath, issue, []model.HistoryEntry{newEntry}); err != nil {
		return fmt.Errorf("save issue file: %w", err)
	}

	for i := range bf.Issues {
		if bf.Issues[i].ID == issueID {
			var newBoardTags []string
			for _, t := range bf.Issues[i].Tags {
				if t != tagName {
					newBoardTags = append(newBoardTags, t)
				}
			}
			bf.Issues[i].Tags = newBoardTags
			break
		}
	}

	if err := store.SaveBoardFileSubmodule(boardPath, bf); err != nil {
		return fmt.Errorf("save board file: %w", err)
	}

	return nil
}

func newCommentCmd() *cobra.Command {
	var message string
	commentCmd := &cobra.Command{
		Use:               "comment <id>",
		Short:             "Add a comment to an issue",
		Args:              cobra.ExactArgs(1),
		RunE:              func(cmd *cobra.Command, args []string) error { return runComment(cmd, args[0], message) },
		ValidArgsFunction: issueIDCompletion,
	}
	commentCmd.Flags().StringVarP(&message, "message", "m", "", "Comment text")
	return commentCmd
}

func runComment(cmd *cobra.Command, issueID, message string) error {
	board := boardFlag(cmd)
	_, issuesDir, err := resolveBoardPaths(board)
	if err != nil {
		return err
	}

	issuePath := filepath.Join(issuesDir, issueID+".yaml")
	issue, err := store.LoadIssueFileSubmodule(issuePath)
	if err != nil {
		return fmt.Errorf("load issue %q: %w", issueID, err)
	}

	text := message
	if !cmd.Flags().Changed("message") {
		text, err = editWithEditor("")
		if err != nil {
			return err
		}
		text = strings.TrimRight(text, "\n")
	}

	if text == "" {
		fmt.Println("no comment added")
		return nil
	}

	identity, err := git.GetIdentityFromRepoRoot()
	if err != nil {
		return err
	}
	now := time.Now().UTC()

	newEntry := model.HistoryEntry{
		Type: "comment",
		At:   now,
		By:   model.Identity{Name: identity.Name, Email: identity.Email},
		Data: model.CommentEntry{Text: text},
	}

	if err := store.SaveIssueFileSubmodule(issuePath, issue, []model.HistoryEntry{newEntry}); err != nil {
		return fmt.Errorf("save issue file: %w", err)
	}

	return nil
}

func newIssueDeleteCmd() *cobra.Command {
	deleteCmd := &cobra.Command{
		Use:               "delete <id>",
		Short:             "Delete an issue",
		Args:              cobra.ExactArgs(1),
		RunE:              runIssueDelete,
		ValidArgsFunction: issueIDCompletion,
	}
	return deleteCmd
}

func runIssueDelete(cmd *cobra.Command, args []string) error {
	board := boardFlag(cmd)
	boardPath, issuesDir, err := resolveBoardPaths(board)
	if err != nil {
		return err
	}
	issueID := args[0]
	issuePath := filepath.Join(issuesDir, issueID+".yaml")
	if _, err := store.LoadIssueFileSubmodule(issuePath); err != nil {
		return fmt.Errorf("load issue %q: %w", issueID, err)
	}
	if err := store.DeleteIssueFileSubmodule(issuePath, boardPath, issueID); err != nil {
		return fmt.Errorf("delete issue: %w", err)
	}
	return nil
}

func newHistoryCmd() *cobra.Command {
	historyCmd := &cobra.Command{
		Use:               "history <id>",
		Short:             "Show full history log for an issue",
		Args:              cobra.ExactArgs(1),
		RunE:              func(cmd *cobra.Command, args []string) error { return runHistory(cmd, args[0]) },
		ValidArgsFunction: issueIDCompletion,
	}
	return historyCmd
}

func runHistory(cmd *cobra.Command, issueID string) error {
	board := boardFlag(cmd)
	_, issuesDir, err := resolveBoardPaths(board)
	if err != nil {
		return err
	}

	issuePath := filepath.Join(issuesDir, issueID+".yaml")
	issue, err := store.LoadIssueFileSubmodule(issuePath)
	if err != nil {
		return fmt.Errorf("load issue %q: %w", issueID, err)
	}

	if len(issue.History) == 0 {
		return nil
	}

	for _, entry := range issue.History {
		fmt.Printf("[%s] %s by %s <%s>\n", entry.At.Format(time.RFC3339), entry.Type, entry.By.Name, entry.By.Email)
		formatHistoryData(entry.Data)
	}

	return nil
}
