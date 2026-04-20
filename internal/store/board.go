package store

import (
	"fmt"
	"os"

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