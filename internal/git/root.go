package git

import (
	"os"
	"path/filepath"
)

var errNotARepo = &notARepoError{}

type notARepoError struct{}

func (e *notARepoError) Error() string {
	return "not inside a git repository"
}

func FindRepoRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", errNotARepo
		}
		dir = parent
	}
}
