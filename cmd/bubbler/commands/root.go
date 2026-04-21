package commands

import (
	"errors"

	"github.com/jana-mind/bubbler/internal/git"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
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

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(newInitCmd())
	rootCmd.AddCommand(newIssueCmd())
	rootCmd.AddCommand(newBoardCmd())
	rootCmd.AddCommand(newCompletionCmd())
	rootCmd.AddCommand(newTuiCmd())
}
