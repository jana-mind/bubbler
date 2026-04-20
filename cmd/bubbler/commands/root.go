package commands

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "bubbler",
	Short: "A CLI kanban board tool",
	Long:  `bubbler is a CLI kanban board tool that stores board state as YAML files in .bubble/`,
}

func Execute() error {
	return rootCmd.Execute()
}
