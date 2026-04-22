package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/go-git/go-git/v5"
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

func TestPull_Success(t *testing.T) {
	tmp := t.TempDir()

	parentDir := filepath.Join(tmp, "parent")
	childDir := filepath.Join(tmp, "child")
	bareDir := filepath.Join(tmp, "bare.git")
	os.MkdirAll(parentDir, 0755)
	os.MkdirAll(childDir, 0755)
	os.MkdirAll(bareDir, 0755)

	gitSetup := func(dir, email, name string) {
		execCmd(t, dir, "git", "init")
		execCmd(t, dir, "git", "config", "user.email", email)
		execCmd(t, dir, "git", "config", "user.name", name)
	}

	gitSetup(parentDir, "parent@test.com", "Parent User")
	gitSetup(childDir, "child@test.com", "Child User")

	execCmd(t, bareDir, "git", "init", "--bare")
	execCmd(t, childDir, "git", "remote", "add", "origin", bareDir)

	os.WriteFile(filepath.Join(childDir, "README"), []byte("initial"), 0644)
	execCmd(t, childDir, "git", "add", ".")
	execCmd(t, childDir, "git", "commit", "-m", "initial")
	execCmd(t, childDir, "git", "push", "-u", "origin", "master")

	execCmd(t, parentDir, "git", "submodule", "add", bareDir, ".bubble")
	execCmd(t, parentDir, "git", "config", "user.email", "parent@test.com")
	execCmd(t, parentDir, "git", "config", "user.name", "Parent User")
	execCmd(t, parentDir, "git", "commit", "-m", "add submodule")

	execCmd(t, childDir, "git", "pull", "origin", "master")
	os.WriteFile(filepath.Join(childDir, "file.txt"), []byte("hello"), 0644)
	execCmd(t, childDir, "git", "add", ".")
	execCmd(t, childDir, "git", "commit", "-m", "add file")
	execCmd(t, childDir, "git", "push", "origin", "master")

	err := Pull(parentDir, ".bubble")
	if err != nil {
		t.Fatalf("Pull failed: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(parentDir, ".bubble", "file.txt"))
	if err != nil {
		t.Fatalf("read file: %v", err)
	}
	if string(data) != "hello" {
		t.Errorf("expected 'hello', got %q", string(data))
	}
}

func TestStageAndCommit_Success(t *testing.T) {
	tmp := t.TempDir()

	parentDir := filepath.Join(tmp, "parent")
	childDir := filepath.Join(tmp, "child")
	bareDir := filepath.Join(tmp, "bare.git")
	os.MkdirAll(parentDir, 0755)
	os.MkdirAll(childDir, 0755)
	os.MkdirAll(bareDir, 0755)

	gitSetup := func(dir, email, name string) {
		execCmd(t, dir, "git", "init")
		execCmd(t, dir, "git", "config", "user.email", email)
		execCmd(t, dir, "git", "config", "user.name", name)
	}

	gitSetup(parentDir, "parent@test.com", "Parent User")
	gitSetup(childDir, "child@test.com", "Child User")

	execCmd(t, bareDir, "git", "init", "--bare")
	execCmd(t, childDir, "git", "remote", "add", "origin", bareDir)

	os.WriteFile(filepath.Join(childDir, "README"), []byte("initial"), 0644)
	execCmd(t, childDir, "git", "add", ".")
	execCmd(t, childDir, "git", "commit", "-m", "initial")
	execCmd(t, childDir, "git", "push", "-u", "origin", "master")

	execCmd(t, parentDir, "git", "submodule", "add", bareDir, ".bubble")
	execCmd(t, parentDir, "git", "config", "user.email", "parent@test.com")
	gitSetup(filepath.Join(parentDir, ".bubble"), "child@test.com", "Child User")
	execCmd(t, parentDir, "git", "commit", "-m", "add submodule")

	os.WriteFile(filepath.Join(parentDir, ".bubble", "newfile.txt"), []byte("content"), 0644)

	err := StageAndCommit(parentDir, ".bubble", "add newfile")
	if err != nil {
		t.Fatalf("StageAndCommit failed: %v", err)
	}

	subRepo, err := git.PlainOpen(filepath.Join(parentDir, ".bubble"))
	if err != nil {
		t.Fatalf("open submodule: %v", err)
	}
	hash, err := subRepo.ResolveRevision("HEAD")
	if err != nil {
		t.Fatalf("resolve HEAD: %v", err)
	}
	if hash.IsZero() {
		t.Error("HEAD is zero, commit was not created")
	}
}

func TestStageAndCommit_NoChanges(t *testing.T) {
	tmp := t.TempDir()

	parentDir := filepath.Join(tmp, "parent")
	childDir := filepath.Join(tmp, "child")
	bareDir := filepath.Join(tmp, "bare.git")
	os.MkdirAll(parentDir, 0755)
	os.MkdirAll(childDir, 0755)
	os.MkdirAll(bareDir, 0755)

	gitSetup := func(dir, email, name string) {
		execCmd(t, dir, "git", "init")
		execCmd(t, dir, "git", "config", "user.email", email)
		execCmd(t, dir, "git", "config", "user.name", name)
	}

	gitSetup(parentDir, "parent@test.com", "Parent User")
	gitSetup(childDir, "child@test.com", "Child User")

	execCmd(t, bareDir, "git", "init", "--bare")
	execCmd(t, childDir, "git", "remote", "add", "origin", bareDir)

	os.WriteFile(filepath.Join(childDir, "README"), []byte("initial"), 0644)
	execCmd(t, childDir, "git", "add", ".")
	execCmd(t, childDir, "git", "commit", "-m", "initial")
	execCmd(t, childDir, "git", "push", "-u", "origin", "master")

	execCmd(t, parentDir, "git", "submodule", "add", bareDir, ".bubble")
	execCmd(t, parentDir, "git", "config", "user.email", "parent@test.com")
	gitSetup(filepath.Join(parentDir, ".bubble"), "child@test.com", "Child User")
	execCmd(t, parentDir, "git", "commit", "-m", "add submodule")

	err := StageAndCommit(parentDir, ".bubble", "no changes")
	if !errorsIs(err, ErrNoChanges) {
		t.Errorf("expected ErrNoChanges, got: %v", err)
	}
}

func TestCommitAndPush_Success(t *testing.T) {
	tmp := t.TempDir()

	parentDir := filepath.Join(tmp, "parent")
	childDir := filepath.Join(tmp, "child")
	bareDir := filepath.Join(tmp, "bare.git")
	os.MkdirAll(parentDir, 0755)
	os.MkdirAll(childDir, 0755)
	os.MkdirAll(bareDir, 0755)

	gitSetup := func(dir, email, name string) {
		execCmd(t, dir, "git", "init")
		execCmd(t, dir, "git", "config", "user.email", email)
		execCmd(t, dir, "git", "config", "user.name", name)
	}

	gitSetup(parentDir, "parent@test.com", "Parent User")
	gitSetup(childDir, "child@test.com", "Child User")

	execCmd(t, bareDir, "git", "init", "--bare")
	execCmd(t, childDir, "git", "remote", "add", "origin", bareDir)

	os.WriteFile(filepath.Join(childDir, "README"), []byte("initial"), 0644)
	execCmd(t, childDir, "git", "add", ".")
	execCmd(t, childDir, "git", "commit", "-m", "initial")
	execCmd(t, childDir, "git", "push", "-u", "origin", "master")

	execCmd(t, parentDir, "git", "submodule", "add", bareDir, ".bubble")
	execCmd(t, parentDir, "git", "config", "user.email", "parent@test.com")
	gitSetup(filepath.Join(parentDir, ".bubble"), "child@test.com", "Child User")
	execCmd(t, parentDir, "git", "commit", "-m", "add submodule")

	os.WriteFile(filepath.Join(parentDir, ".bubble", "pushed.txt"), []byte("pushed content"), 0644)

	err := CommitAndPush(parentDir, ".bubble", "add pushed file via bubbler")
	if err != nil {
		t.Fatalf("CommitAndPush failed: %v", err)
	}

	bareRepo, err := git.PlainOpen(bareDir)
	if err != nil {
		t.Fatalf("open bare repo: %v", err)
	}
	hash, err := bareRepo.ResolveRevision("HEAD")
	if err != nil {
		t.Fatalf("resolve HEAD in bare: %v", err)
	}
	if hash.IsZero() {
		t.Error("bare repo HEAD is zero, push failed")
	}
}

func execCmd(t *testing.T, dir, name string, args ...string) {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GIT_ALLOW_PROTOCOL=file")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("exec %s %v in %s failed: %v\n%s", name, args, dir, err, output)
	}
}

func TestIsSubmodule_BubbleManualSkipsSubmoduleCheck(t *testing.T) {
	tmp := t.TempDir()
	parentDir := filepath.Join(tmp, "parent")
	childDir := filepath.Join(tmp, "child")
	bareDir := filepath.Join(tmp, "bare.git")
	os.MkdirAll(parentDir, 0755)
	os.MkdirAll(childDir, 0755)
	os.MkdirAll(bareDir, 0755)

	execCmd(t, parentDir, "git", "init")
	execCmd(t, parentDir, "git", "config", "user.email", "test@test.com")
	execCmd(t, parentDir, "git", "config", "user.name", "Test")
	execCmd(t, childDir, "git", "init")
	execCmd(t, childDir, "git", "config", "user.email", "test@test.com")
	execCmd(t, childDir, "git", "config", "user.name", "Test")
	execCmd(t, bareDir, "git", "init", "--bare")
	execCmd(t, childDir, "git", "remote", "add", "origin", bareDir)
	os.WriteFile(filepath.Join(childDir, "README"), []byte("init"), 0644)
	execCmd(t, childDir, "git", "add", ".")
	execCmd(t, childDir, "git", "commit", "-m", "init")
	execCmd(t, childDir, "git", "push", "-u", "origin", "master")
	execCmd(t, parentDir, "git", "submodule", "add", bareDir, ".bubble")
	execCmd(t, parentDir, "git", "config", "user.email", "test@test.com")
	execCmd(t, parentDir, "git", "commit", "-m", "add submodule")

	bubblePath := filepath.Join(parentDir, ".bubble")
	if !IsSubmodule(parentDir, bubblePath) {
		t.Fatal("expected true before adding .bubble-manual")
	}

	os.WriteFile(filepath.Join(bubblePath, ".bubble-manual"), []byte(""), 0644)
	if IsSubmodule(parentDir, bubblePath) {
		t.Fatal("expected false after adding .bubble-manual")
	}
}

func errorsIs(err, target error) bool {
	return err == target
}
