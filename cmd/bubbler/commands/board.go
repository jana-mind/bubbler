package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jana-mind/bubbler/internal/git"
	"github.com/jana-mind/bubbler/internal/model"
	"github.com/jana-mind/bubbler/internal/store"
	"github.com/spf13/cobra"
)

func newBoardCmd() *cobra.Command {
	boardCmd := &cobra.Command{
		Use:   "board",
		Short: "Manage board columns and tags",
	}
	boardCmd.AddCommand(newBoardListCmd())
	boardCmd.AddCommand(newBoardCreateCmd())
	boardCmd.AddCommand(newBoardDeleteCmd())
	boardCmd.AddCommand(newColumnCmd())
	boardCmd.AddCommand(newTagCmd())
	return boardCmd
}

func newBoardListCmd() *cobra.Command {
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List all boards",
		RunE:  runBoardList,
	}
	return listCmd
}

func runBoardList(cmd *cobra.Command, args []string) error {
	repoRoot, err := git.FindRepoRoot()
	if err != nil {
		return err
	}
	bubblePath := filepath.Join(repoRoot, ".bubble")
	entries, err := os.ReadDir(bubblePath)
	if err != nil {
		return fmt.Errorf("read bubble directory: %w", err)
	}
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".yaml" {
			continue
		}
		name := strings.TrimSuffix(entry.Name(), ".yaml")
		fmt.Println(name)
	}
	return nil
}

func newBoardCreateCmd() *cobra.Command {
	createCmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create a new board",
		Args:  cobra.ExactArgs(1),
		RunE:  runBoardCreate,
	}
	return createCmd
}

func runBoardCreate(cmd *cobra.Command, args []string) error {
	name := args[0]
	if name == "" {
		return fmt.Errorf("board name cannot be empty")
	}
	if strings.ContainsAny(name, "/\\.") {
		return fmt.Errorf("board name must not contain /, \\, or .")
	}

	repoRoot, err := git.FindRepoRoot()
	if err != nil {
		return err
	}
	bubblePath := filepath.Join(repoRoot, ".bubble")
	boardFile := filepath.Join(bubblePath, name+".yaml")
	if _, err := os.Stat(boardFile); err == nil {
		return fmt.Errorf("board %q already exists", name)
	}

	issuesDir := filepath.Join(bubblePath, name)
	if err := os.MkdirAll(issuesDir, 0755); err != nil {
		return fmt.Errorf("create issues directory: %w", err)
	}

	defaultColumns := []column{
		{ID: "waiting", Label: "Waiting"},
		{ID: "in-progress", Label: "In Progress"},
		{ID: "completed", Label: "Completed"},
	}
	b := board{
		Board: boardMeta{
			Name:    name,
			Columns: defaultColumns,
			Tags:    []string{"bug", "feature", "docs", "chore"},
			NextID:  1,
		},
	}

	boardPath := filepath.Join(bubblePath, name+".yaml")
	if err := store.SaveBoardFile(boardPath, modelBoardToBoardFile(b)); err != nil {
		os.RemoveAll(issuesDir)
		return fmt.Errorf("save board file: %w", err)
	}

	fmt.Println(name)
	return nil
}

func newBoardDeleteCmd() *cobra.Command {
	deleteCmd := &cobra.Command{
		Use:   "delete <name>",
		Short: "Delete a board",
		Args:  cobra.ExactArgs(1),
		RunE:  runBoardDelete,
	}
	return deleteCmd
}

func runBoardDelete(cmd *cobra.Command, args []string) error {
	name := args[0]
	if name == "default" {
		return fmt.Errorf("the default board may never be deleted")
	}

	repoRoot, err := git.FindRepoRoot()
	if err != nil {
		return err
	}
	bubblePath := filepath.Join(repoRoot, ".bubble")
	boardFile := filepath.Join(bubblePath, name+".yaml")
	if _, err := os.Stat(boardFile); os.IsNotExist(err) {
		return fmt.Errorf("board %q does not exist", name)
	}

	issuesDir := filepath.Join(bubblePath, name)
	entries, err := os.ReadDir(issuesDir)
	if err != nil {
		return fmt.Errorf("read issues directory: %w", err)
	}
	var issues []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".yaml") {
			issues = append(issues, e.Name())
		}
	}
	if len(issues) > 0 {
		return fmt.Errorf("board %q has %d issue(s) and cannot be deleted", name, len(issues))
	}

	if err := os.Remove(boardFile); err != nil {
		return fmt.Errorf("remove board file: %w", err)
	}
	if err := os.RemoveAll(issuesDir); err != nil {
		return fmt.Errorf("remove issues directory: %w", err)
	}

	return nil
}

func modelBoardToBoardFile(b board) model.BoardFile {
	cols := make([]model.Column, len(b.Board.Columns))
	for i, c := range b.Board.Columns {
		cols[i] = model.Column{ID: c.ID, Label: c.Label}
	}
	return model.BoardFile{
		Board: model.Board{
			Name:    b.Board.Name,
			Columns: cols,
			Tags:    b.Board.Tags,
			NextID:  b.Board.NextID,
		},
		Issues: nil,
	}
}

func newColumnCmd() *cobra.Command {
	columnCmd := &cobra.Command{
		Use:   "column",
		Short: "Manage columns",
	}
	columnCmd.AddCommand(newColumnListCmd())
	columnCmd.AddCommand(newColumnAddCmd())
	columnCmd.AddCommand(newColumnRemoveCmd())
	return columnCmd
}

func newColumnListCmd() *cobra.Command {
	columnListCmd := &cobra.Command{
		Use:   "list",
		Short: "List all columns",
		RunE:  runColumnList,
	}
	return columnListCmd
}

func runColumnList(cmd *cobra.Command, args []string) error {
	board := boardFlag(cmd)
	boardPath, _, err := resolveBoardPaths(board)
	if err != nil {
		return err
	}
	bf, err := store.LoadBoardFileSubmodule(boardPath)
	if err != nil {
		return fmt.Errorf("load board: %w", err)
	}
	for _, col := range bf.Board.Columns {
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
	board := boardFlag(cmd)
	boardPath, _, err := resolveBoardPaths(board)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	bf, err := store.LoadBoardFileSubmodule(boardPath)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	ids := make([]string, len(bf.Board.Columns))
	for i, col := range bf.Board.Columns {
		ids[i] = col.ID
	}
	return ids, cobra.ShellCompDirectiveNoFileComp
}

func tagNameCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) != 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	board := boardFlag(cmd)
	boardPath, _, err := resolveBoardPaths(board)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	bf, err := store.LoadBoardFileSubmodule(boardPath)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	return bf.Board.Tags, cobra.ShellCompDirectiveNoFileComp
}

func runColumnAdd(cmd *cobra.Command, args []string) error {
	board := boardFlag(cmd)
	boardPath, _, err := resolveBoardPaths(board)
	if err != nil {
		return err
	}
	bf, err := store.LoadBoardFileSubmodule(boardPath)
	if err != nil {
		return fmt.Errorf("load board: %w", err)
	}
	name := args[0]
	id := slugify(name)
	for _, col := range bf.Board.Columns {
		if col.ID == id {
			return fmt.Errorf("column %q already exists", id)
		}
	}
	bf.Board.Columns = append(bf.Board.Columns, struct {
		ID    string `yaml:"id"`
		Label string `yaml:"label"`
	}{ID: id, Label: name})
	if err := store.SaveBoardFileSubmodule(boardPath, bf); err != nil {
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
	board := boardFlag(cmd)
	boardPath, _, err := resolveBoardPaths(board)
	if err != nil {
		return err
	}
	bf, err := store.LoadBoardFileSubmodule(boardPath)
	if err != nil {
		return fmt.Errorf("load board: %w", err)
	}
	name := args[0]
	var foundIdx int = -1
	for i, col := range bf.Board.Columns {
		if col.ID == name || col.Label == name {
			foundIdx = i
			break
		}
	}
	if foundIdx == -1 {
		return fmt.Errorf("column %q not found", name)
	}
	for _, iss := range bf.Issues {
		if iss.Column == bf.Board.Columns[foundIdx].ID {
			return fmt.Errorf("column %q has issues and cannot be removed", name)
		}
	}
	bf.Board.Columns = append(bf.Board.Columns[:foundIdx], bf.Board.Columns[foundIdx+1:]...)
	if err := store.SaveBoardFileSubmodule(boardPath, bf); err != nil {
		return fmt.Errorf("save board file: %w", err)
	}
	return nil
}

func newTagCmd() *cobra.Command {
	tagCmd := &cobra.Command{
		Use:   "tag",
		Short: "Manage tags",
	}
	tagCmd.AddCommand(newTagListCmd())
	tagCmd.AddCommand(newTagAddCmd())
	tagCmd.AddCommand(newTagRemoveCmd())
	return tagCmd
}

func newTagListCmd() *cobra.Command {
	tagListCmd := &cobra.Command{
		Use:   "list",
		Short: "List all tags",
		RunE:  runTagList,
	}
	return tagListCmd
}

func runTagList(cmd *cobra.Command, args []string) error {
	board := boardFlag(cmd)
	boardPath, _, err := resolveBoardPaths(board)
	if err != nil {
		return err
	}
	bf, err := store.LoadBoardFileSubmodule(boardPath)
	if err != nil {
		return fmt.Errorf("load board: %w", err)
	}
	for _, tag := range bf.Board.Tags {
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
	board := boardFlag(cmd)
	boardPath, _, err := resolveBoardPaths(board)
	if err != nil {
		return err
	}
	bf, err := store.LoadBoardFileSubmodule(boardPath)
	if err != nil {
		return fmt.Errorf("load board: %w", err)
	}
	name := args[0]
	for _, tag := range bf.Board.Tags {
		if tag == name {
			return fmt.Errorf("tag %q already exists", name)
		}
	}
	bf.Board.Tags = append(bf.Board.Tags, name)
	if err := store.SaveBoardFileSubmodule(boardPath, bf); err != nil {
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
	board := boardFlag(cmd)
	boardPath, _, err := resolveBoardPaths(board)
	if err != nil {
		return err
	}
	bf, err := store.LoadBoardFileSubmodule(boardPath)
	if err != nil {
		return fmt.Errorf("load board: %w", err)
	}
	name := args[0]
	found := false
	for _, tag := range bf.Board.Tags {
		if tag == name {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("tag %q not found", name)
	}
	for _, iss := range bf.Issues {
		for _, t := range iss.Tags {
			if t == name {
				return fmt.Errorf("tag %q is used by issue %s and cannot be removed", name, iss.ID)
			}
		}
	}
	var newTags []string
	for _, tag := range bf.Board.Tags {
		if tag != name {
			newTags = append(newTags, tag)
		}
	}
	bf.Board.Tags = newTags
	if err := store.SaveBoardFileSubmodule(boardPath, bf); err != nil {
		return fmt.Errorf("save board file: %w", err)
	}
	return nil
}
