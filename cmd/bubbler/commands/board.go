package commands

import (
	"fmt"
	"strings"

	"github.com/jana-mind/bubbler/internal/store"
	"github.com/spf13/cobra"
)

func newBoardCmd() *cobra.Command {
	boardCmd := &cobra.Command{
		Use:   "board",
		Short: "Manage board columns and tags",
	}
	boardCmd.AddCommand(newColumnCmd())
	boardCmd.AddCommand(newTagCmd())
	return boardCmd
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
