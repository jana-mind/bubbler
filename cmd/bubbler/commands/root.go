package commands

import (
	"errors"
	"path/filepath"

	"github.com/jana-mind/bubbler/internal/git"
	"github.com/spf13/cobra"
)

var (
	boardName string
	rootCmd   = &cobra.Command{
		Use:   "bubbler",
		Short: "A CLI kanban board tool",
		Long:  `bubbler is a CLI kanban board tool that stores board state as YAML files in .bubble/`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			_, err := git.FindRepoRoot()
			if err != nil {
				return errors.New("bubbler must be run from inside a git repository")
			}
			return nil
		},
	}
)

func resolveBoardPaths(board string) (boardFile, issuesDir string, err error) {
	repoRoot, err := git.FindRepoRoot()
	if err != nil {
		return "", "", err
	}
	bubblePath := filepath.Join(repoRoot, ".bubble")
	boardFile = filepath.Join(bubblePath, board+".yaml")
	issuesDir = filepath.Join(bubblePath, board)
	return boardFile, issuesDir, nil
}

func boardFlag(cmd *cobra.Command) string {
	if cmd.Flags().Changed("board") {
		f, _ := cmd.Flags().GetString("board")
		return f
	}
	for c := cmd; c.Parent() != nil; c = c.Parent() {
		if c.Flags().Changed("board") {
			f, _ := c.Flags().GetString("board")
			return f
		}
	}
	return "default"
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&boardName, "board", "b", "default", "Board name to operate on")
	rootCmd.AddCommand(newInitCmd())
	rootCmd.AddCommand(newIssueCmd())
	rootCmd.AddCommand(newBoardCmd())
	rootCmd.AddCommand(newCompletionCmd())
	rootCmd.AddCommand(newTuiCmd())
}
