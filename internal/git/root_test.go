package git

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindRepoRoot_NotARepo(t *testing.T) {
	tmp := t.TempDir()
	os.Chdir(tmp)
	_, err := FindRepoRoot()
	if err == nil {
		t.Fatal("expected error outside repo")
	}
	if err != errNotARepo {
		t.Fatalf("expected errNotARepo, got: %T %v", err, err)
	}
}

func TestFindRepoRoot_InsideRepo(t *testing.T) {
	tmp := t.TempDir()
	os.MkdirAll(filepath.Join(tmp, ".git"), 0755)
	os.Chdir(tmp)
	root, err := FindRepoRoot()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if root != tmp {
		t.Fatalf("expected %q, got %q", tmp, root)
	}
}

func TestFindRepoRoot_Subdirectory(t *testing.T) {
	tmp := t.TempDir()
	os.MkdirAll(filepath.Join(tmp, ".git"), 0755)
	subdir := filepath.Join(tmp, "a", "b", "c")
	os.MkdirAll(subdir, 0755)
	os.Chdir(subdir)
	root, err := FindRepoRoot()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if root != tmp {
		t.Fatalf("expected %q, got %q", tmp, root)
	}
}
