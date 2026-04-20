package git

import (
	"os"
	"path/filepath"
	"strings"
)

func IsSubmodule(repoRoot, bubblePath string) bool {
	modulesFile := filepath.Join(repoRoot, ".gitmodules")
	data, err := os.ReadFile(modulesFile)
	if err != nil {
		return false
	}
	lines := strings.Split(string(data), "\n")
	inSubmodule := false
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "[submodule ") {
			content := strings.Trim(line[11:], "\"]")
			inSubmodule = content == bubblePath
		} else if inSubmodule && strings.HasPrefix(line, "path = ") {
			return strings.Trim(line[7:], " \"") == bubblePath
		}
	}
	return false
}
