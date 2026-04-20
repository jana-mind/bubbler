package commands

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/jana-mind/bubbler/internal/git"
	"github.com/jana-mind/bubbler/internal/store"
	"github.com/spf13/cobra"
)

func newBoardCmd() *cobra.Command {
	boardCmd := &cobra.Command{
		Use:   "board",
		Short: "Manage board columns and tags",
	}
	boardCmd.AddCommand(newColumnsCmd())
	boardCmd.AddCommand(newColumnAddCmd())
	boardCmd.AddCommand(newColumnRemoveCmd())
	boardCmd.AddCommand(newTagsCmd())
	boardCmd.AddCommand(newTagAddCmd())
	boardCmd.AddCommand(newTagRemoveCmd())
	return boardCmd
}

func newColumnsCmd() *cobra.Command {
	columnsCmd := &cobra.Command{
		Use:   "columns",
		Short: "List all columns",
		RunE:  runColumns,
	}
	return columnsCmd
}

func runColumns(cmd *cobra.Command, args []string) error {
	repoRoot, err := git.FindRepoRoot()
	if err != nil {
		return err
	}
	boardPath := filepath.Join(repoRoot, ".bubble", "default.yaml")
	board, err := store.LoadBoardFileSubmodule(boardPath)
	if err != nil {
		return fmt.Errorf("load board: %w", err)
	}
	for _, col := range board.Board.Columns {
		fmt.Printf("%s  %s\n", col.ID, col.Label)
	}
	return nil
}

func newColumnAddCmd() *cobra.Command {
	columnAddCmd := &cobra.Command{
		Use:   "add <name>",
		Short: "Add a new column",
		Args:  cobra.ExactArgs(1),
		RunE:  runColumnAdd,
	}
	return columnAddCmd
}

func slugify(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, " ", "-")
	return s
}

func columnNameCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) != 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	repoRoot, err := git.FindRepoRoot()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	boardPath := filepath.Join(repoRoot, ".bubble", "default.yaml")
	board, err := store.LoadBoardFileSubmodule(boardPath)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	ids := make([]string, len(board.Board.Columns))
	for i, col := range board.Board.Columns {
		ids[i] = col.ID
	}
	return ids, cobra.ShellCompDirectiveNoFileComp
}

func tagNameCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) != 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	repoRoot, err := git.FindRepoRoot()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	boardPath := filepath.Join(repoRoot, ".bubble", "default.yaml")
	board, err := store.LoadBoardFileSubmodule(boardPath)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	return board.Board.Tags, cobra.ShellCompDirectiveNoFileComp
}

func runColumnAdd(cmd *cobra.Command, args []string) error {
	repoRoot, err := git.FindRepoRoot()
	if err != nil {
		return err
	}
	boardPath := filepath.Join(repoRoot, ".bubble", "default.yaml")
	board, err := store.LoadBoardFileSubmodule(boardPath)
	if err != nil {
		return fmt.Errorf("load board: %w", err)
	}
	name := args[0]
	id := slugify(name)
	for _, col := range board.Board.Columns {
		if col.ID == id {
			return fmt.Errorf("column %q already exists", id)
		}
	}
	board.Board.Columns = append(board.Board.Columns, struct {
		ID    string `yaml:"id"`
		Label string `yaml:"label"`
	}{ID: id, Label: name})
	if err := store.SaveBoardFileSubmodule(boardPath, board); err != nil {
		return fmt.Errorf("save board file: %w", err)
	}
	fmt.Println(id)
	return nil
}

func newColumnRemoveCmd() *cobra.Command {
	columnRemoveCmd := &cobra.Command{
		Use:               "remove <name>",
		Short:             "Remove a column",
		Args:              cobra.ExactArgs(1),
		RunE:              runColumnRemove,
		ValidArgsFunction: columnNameCompletion,
	}
	return columnRemoveCmd
}

func runColumnRemove(cmd *cobra.Command, args []string) error {
	repoRoot, err := git.FindRepoRoot()
	if err != nil {
		return err
	}
	boardPath := filepath.Join(repoRoot, ".bubble", "default.yaml")
	board, err := store.LoadBoardFileSubmodule(boardPath)
	if err != nil {
		return fmt.Errorf("load board: %w", err)
	}
	name := args[0]
	var foundIdx int = -1
	for i, col := range board.Board.Columns {
		if col.ID == name || col.Label == name {
			foundIdx = i
			break
		}
	}
	if foundIdx == -1 {
		return fmt.Errorf("column %q not found", name)
	}
	for _, iss := range board.Issues {
		if iss.Column == board.Board.Columns[foundIdx].ID {
			return fmt.Errorf("column %q has issues and cannot be removed", name)
		}
	}
	board.Board.Columns = append(board.Board.Columns[:foundIdx], board.Board.Columns[foundIdx+1:]...)
	if err := store.SaveBoardFileSubmodule(boardPath, board); err != nil {
		return fmt.Errorf("save board file: %w", err)
	}
	return nil
}

func newTagsCmd() *cobra.Command {
	tagsCmd := &cobra.Command{
		Use:   "tags",
		Short: "List all tags",
		RunE:  runTags,
	}
	return tagsCmd
}

func runTags(cmd *cobra.Command, args []string) error {
	repoRoot, err := git.FindRepoRoot()
	if err != nil {
		return err
	}
	boardPath := filepath.Join(repoRoot, ".bubble", "default.yaml")
	board, err := store.LoadBoardFileSubmodule(boardPath)
	if err != nil {
		return fmt.Errorf("load board: %w", err)
	}
	for _, tag := range board.Board.Tags {
		fmt.Println(tag)
	}
	return nil
}

func newTagAddCmd() *cobra.Command {
	tagAddCmd := &cobra.Command{
		Use:   "add <name>",
		Short: "Add a new tag",
		Args:  cobra.ExactArgs(1),
		RunE:  runTagAdd,
	}
	return tagAddCmd
}

func runTagAdd(cmd *cobra.Command, args []string) error {
	repoRoot, err := git.FindRepoRoot()
	if err != nil {
		return err
	}
	boardPath := filepath.Join(repoRoot, ".bubble", "default.yaml")
	board, err := store.LoadBoardFileSubmodule(boardPath)
	if err != nil {
		return fmt.Errorf("load board: %w", err)
	}
	name := args[0]
	for _, tag := range board.Board.Tags {
		if tag == name {
			return fmt.Errorf("tag %q already exists", name)
		}
	}
	board.Board.Tags = append(board.Board.Tags, name)
	if err := store.SaveBoardFileSubmodule(boardPath, board); err != nil {
		return fmt.Errorf("save board file: %w", err)
	}
	return nil
}

func newTagRemoveCmd() *cobra.Command {
	tagRemoveCmd := &cobra.Command{
		Use:               "remove <name>",
		Short:             "Remove a tag",
		Args:              cobra.ExactArgs(1),
		RunE:              runTagRemove,
		ValidArgsFunction: tagNameCompletion,
	}
	return tagRemoveCmd
}

func runTagRemove(cmd *cobra.Command, args []string) error {
	repoRoot, err := git.FindRepoRoot()
	if err != nil {
		return err
	}
	boardPath := filepath.Join(repoRoot, ".bubble", "default.yaml")
	board, err := store.LoadBoardFileSubmodule(boardPath)
	if err != nil {
		return fmt.Errorf("load board: %w", err)
	}
	name := args[0]
	found := false
	for _, tag := range board.Board.Tags {
		if tag == name {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("tag %q not found", name)
	}
	for _, iss := range board.Issues {
		for _, t := range iss.Tags {
			if t == name {
				return fmt.Errorf("tag %q is used by issue %s and cannot be removed", name, iss.ID)
			}
		}
	}
	var newTags []string
	for _, tag := range board.Board.Tags {
		if tag != name {
			newTags = append(newTags, tag)
		}
	}
	board.Board.Tags = newTags
	if err := store.SaveBoardFileSubmodule(boardPath, board); err != nil {
		return fmt.Errorf("save board file: %w", err)
	}
	return nil
}
