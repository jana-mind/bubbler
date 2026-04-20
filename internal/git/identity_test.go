package git

import (
	"os"
	"testing"

	"github.com/go-git/go-git/v5"
)

func TestGetIdentity(t *testing.T) {
	tmp := t.TempDir()
	repo, err := git.PlainInit(tmp, false)
	if err != nil {
		t.Fatalf("failed to init git repo: %v", err)
	}

	_, err = GetIdentity(tmp)
	if err != ErrNoIdentity {
		t.Fatalf("expected ErrNoIdentity when no identity set, got: %v", err)
	}

	cfg, _ := repo.Config()
	cfg.User.Name = "Test User"
	cfg.User.Email = "test@example.com"
	repo.Storer.SetConfig(cfg)

	ident, err := GetIdentity(tmp)
	if err != nil {
		t.Fatalf("expected no error with identity set, got: %v", err)
	}
	if ident.Name != "Test User" || ident.Email != "test@example.com" {
		t.Fatalf("expected identity to match, got: %+v", ident)
	}
}

func TestGetIdentityFromEnv(t *testing.T) {
	os.Unsetenv("GIT_AUTHOR_NAME")
	os.Unsetenv("GIT_AUTHOR_EMAIL")
	defer func() {
		os.Unsetenv("GIT_AUTHOR_NAME")
		os.Unsetenv("GIT_AUTHOR_EMAIL")
	}()

	_, err := GetIdentityFromEnv()
	if err != ErrNoIdentity {
		t.Fatalf("expected ErrNoIdentity with no env vars set, got: %v", err)
	}

	os.Setenv("GIT_AUTHOR_NAME", "Env User")
	os.Setenv("GIT_AUTHOR_EMAIL", "env@example.com")
	ident, err := GetIdentityFromEnv()
	if err != nil {
		t.Fatalf("expected no error with env vars set, got: %v", err)
	}
	if ident.Name != "Env User" || ident.Email != "env@example.com" {
		t.Fatalf("expected identity from env, got: %+v", ident)
	}
}
