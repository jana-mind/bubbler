package commands

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/jana-mind/bubbler/internal/git"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var errBubbleAlreadyExists = errors.New(".bubble/ already exists")
var errNotInRepo = errors.New("not inside a git repository")

type board struct {
	Board boardMeta `yaml:"board"`
}

type boardMeta struct {
	Name    string    `yaml:"name"`
	Columns []column  `yaml:"columns"`
	Tags    []string  `yaml:"tags"`
}

type column struct {
	ID    string `yaml:"id"`
	Label string `yaml:"label"`
}

func newInitCmd() *cobra.Command {
	initCmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize the .bubble/ directory in the current repo",
		Long:  `Creates .bubble/default.yaml and .bubble/default/ directory. Fails if .bubble/ already exists.`,
		RunE:  runInit,
	}
	return initCmd
}

func runInit(cmd *cobra.Command, args []string) error {
	repoRoot, err := git.FindRepoRoot()
	if err != nil {
		return errNotInRepo
	}

	bubblePath := filepath.Join(repoRoot, ".bubble")
	if _, err := os.Stat(bubblePath); err == nil {
		return errBubbleAlreadyExists
	}

	defaultDir := filepath.Join(bubblePath, "default")
	if err := os.MkdirAll(defaultDir, 0755); err != nil {
		return err
	}

	b := board{
		Board: boardMeta{
			Name: "default",
			Columns: []column{
				{ID: "waiting", Label: "Waiting"},
				{ID: "in-progress", Label: "In Progress"},
				{ID: "completed", Label: "Completed"},
			},
			Tags: []string{"bug", "feature", "docs", "chore"},
		},
	}

	data, err := yaml.Marshal(b)
	if err != nil {
		return err
	}

	boardFile := filepath.Join(bubblePath, "default.yaml")
	if err := os.WriteFile(boardFile, data, 0644); err != nil {
		return err
	}

	return nil
}
