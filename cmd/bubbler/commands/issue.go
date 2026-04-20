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
