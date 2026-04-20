package git

import (
	"errors"
	"os"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
)

var ErrNoIdentity = errors.New("git identity not configured: user.name and user.email are required")

type Identity struct {
	Name  string
	Email string
}

func GetIdentity(repoPath string) (Identity, error) {
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return Identity{}, ErrNoIdentity
	}
	cfg, err := repo.ConfigScoped(config.GlobalScope)
	if err != nil {
		return Identity{}, ErrNoIdentity
	}
	if cfg.User.Name == "" || cfg.User.Email == "" {
		return Identity{}, ErrNoIdentity
	}
	return Identity{Name: cfg.User.Name, Email: cfg.User.Email}, nil
}

func GetIdentityFromEnv() (Identity, error) {
	name := os.Getenv("GIT_AUTHOR_NAME")
	email := os.Getenv("GIT_AUTHOR_EMAIL")
	if name != "" && email != "" {
		return Identity{Name: name, Email: email}, nil
	}
	return Identity{}, ErrNoIdentity
}
