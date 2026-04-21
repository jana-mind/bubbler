package commands

import (
	"fmt"
	"os"

	"github.com/jana-mind/bubbler/internal/git"
	"github.com/jana-mind/bubbler/internal/tui"
	"github.com/spf13/cobra"
)

func newTuiCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tui",
		Short: "Open the interactive kanban board",
		RunE: func(cmd *cobra.Command, args []string) error {
			boardName, _ := cmd.Flags().GetString("board")

			repoRoot, err := git.FindRepoRoot()
			if err != nil {
				return fmt.Errorf("not in a git repository: %w", err)
			}

			bubblePath := repoRoot + "/.bubble"
			if _, err := os.Stat(bubblePath); os.IsNotExist(err) {
				return fmt.Errorf("no board found. run `bubbler init` first")
			}

			if err := tui.Run(boardName); err != nil {
				return fmt.Errorf("tui error: %w", err)
			}
			return nil
		},
	}
	cmd.Flags().StringP("board", "b", "default", "board name to open")
	return cmd
}
