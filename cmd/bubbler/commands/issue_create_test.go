package commands

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestCreateInteractiveTabCompletionBug(t *testing.T) {
	tmpDir := t.TempDir()
	tmpDir = filepath.Join(tmpDir, "repo")
	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		t.Fatal(err)
	}

	gitDir := filepath.Join(tmpDir, ".git")
	if err := os.MkdirAll(gitDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(gitDir, "config"), []byte(`[user]
	name = Test
	email = test@test.com
`), 0644); err != nil {
		t.Fatal(err)
	}

	bubbleDir := filepath.Join(tmpDir, ".bubble")
	if err := os.MkdirAll(filepath.Join(bubbleDir, "default"), 0755); err != nil {
		t.Fatal(err)
	}
	defaultYaml := `board:
  name: default
  columns:
  - id: waiting
    label: Waiting
  - id: in-progress
    label: In Progress
  tags:
  - bug
  - feature
issues: []
`
	if err := os.WriteFile(filepath.Join(bubbleDir, "default.yaml"), []byte(defaultYaml), 0644); err != nil {
		t.Fatal(err)
	}

	// Build bubbler binary to a temp location so tests are portable
	bubblerBin := filepath.Join(tmpDir, "bubbler_test_bin")
	buildCmd := exec.Command("go", "build", "-o", bubblerBin, ".")
	buildCmd.Dir = filepath.Dir(filepath.Dir(filepath.Dir(filepath.Dir(filepath.FromSlash("../../.."))))) // repo root
	if out, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to build bubbler: %v\n%s", err, out)
	}

	t.Run("create with --title flag, Enter does not execute partial shell command", func(t *testing.T) {
		cmd := exec.Command(bubblerBin, "issue", "create",
			"--title", "Test Issue")
		cmd.Dir = tmpDir
		cmd.Env = append(os.Environ(), "EDITOR=true", "PATH="+os.Getenv("PATH"))

		var stdin bytes.Buffer
		cmd.Stdin = &stdin

		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Logf("stdout+stderr: %s", string(out))
		}

		for _, line := range strings.Split(string(out), "\n") {
			if strings.Contains(line, "sh:") || strings.Contains(line, "bash:") || strings.Contains(line, ": command not found") {
				t.Errorf("shell command execution detected: %s", line)
			}
		}
	})
}
