package store

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jana-mind/bubbler/internal/model"
	"gopkg.in/yaml.v3"
)

func LoadBoardFile(path string) (model.BoardFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return model.BoardFile{}, fmt.Errorf("read board file: %w", err)
	}

	var bf model.BoardFile
	if err := yaml.Unmarshal(data, &bf); err != nil {
		return model.BoardFile{}, fmt.Errorf("parse board file: %w", err)
	}

	if err := bf.Validate(); err != nil {
		return model.BoardFile{}, fmt.Errorf("validate board file: %w", err)
	}

	return bf, nil
}

func SaveBoardFile(path string, bf model.BoardFile) error {
	data, err := yaml.Marshal(bf)
	if err != nil {
		return fmt.Errorf("marshal board file: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("write board file: %w", err)
	}

	return nil
}

type BoardLock struct {
	filePath string
	lockPath string
}

func LockBoard(boardPath string) (*BoardLock, error) {
	bubblePath := filepath.Dir(boardPath)
	boardName := strings.TrimSuffix(filepath.Base(boardPath), ".yaml")
	lockPath := filepath.Join(bubblePath, "__"+boardName+".lock")

	f, err := os.OpenFile(lockPath, os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		if os.IsExist(err) {
			return nil, fmt.Errorf("board %q is locked by another process", boardName)
		}
		return nil, fmt.Errorf("acquire lock: %w", err)
	}
	f.Close()
	return &BoardLock{filePath: boardPath, lockPath: lockPath}, nil
}

func (l *BoardLock) Unlock() error {
	if err := os.Remove(l.lockPath); err != nil {
		return fmt.Errorf("release lock: %w", err)
	}
	return nil
}

func LoadBoardFileForUpdate(boardPath string) (model.BoardFile, *BoardLock, error) {
	lock, err := LockBoard(boardPath)
	if err != nil {
		return model.BoardFile{}, nil, err
	}
	bf, err := LoadBoardFile(boardPath)
	if err != nil {
		lock.Unlock()
		return model.BoardFile{}, nil, err
	}
	return bf, lock, nil
}

func SaveBoardFileForUpdate(boardPath string, bf model.BoardFile, lock *BoardLock) error {
	if err := SaveBoardFile(boardPath, bf); err != nil {
		return err
	}
	if lock != nil {
		lock.Unlock()
	}
	return nil
}
