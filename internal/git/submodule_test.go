package git

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIsSubmodule_NotASubmodule(t *testing.T) {
	tmp := t.TempDir()
	os.MkdirAll(filepath.Join(tmp, ".git"), 0755)
	os.WriteFile(filepath.Join(tmp, ".gitmodules"), []byte(""), 0644)
	if IsSubmodule(tmp, ".bubble") {
		t.Fatal("expected false when .gitmodules is empty")
	}
}

func TestIsSubmodule_IsSubmodule(t *testing.T) {
	tmp := t.TempDir()
	os.MkdirAll(filepath.Join(tmp, ".git"), 0755)
	content := `[submodule ".bubble"]
	path = .bubble
	url = https://example.com/bubble.git
`
	os.WriteFile(filepath.Join(tmp, ".gitmodules"), []byte(content), 0644)
	if !IsSubmodule(tmp, ".bubble") {
		t.Fatal("expected true for .bubble submodule")
	}
}

func TestIsSubmodule_NoGitmodulesFile(t *testing.T) {
	tmp := t.TempDir()
	os.MkdirAll(filepath.Join(tmp, ".git"), 0755)
	if IsSubmodule(tmp, ".bubble") {
		t.Fatal("expected false when no .gitmodules file")
	}
}
