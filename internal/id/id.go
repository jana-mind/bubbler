package id

import (
	"crypto/rand"
	"fmt"
)

const length = 6
const maxAttempts = 8

func Generate(existingIDs []string) (string, error) {
	used := make(map[string]struct{}, len(existingIDs))
	for _, id := range existingIDs {
		used[id] = struct{}{}
	}

	for i := 0; i < maxAttempts; i++ {
		candidate, err := generateOne()
		if err != nil {
			return "", err
		}
		if _, ok := used[candidate]; !ok {
			return candidate, nil
		}
	}

	return "", fmt.Errorf("failed to generate unique ID after %d attempts", maxAttempts)
}

func generateOne() (string, error) {
	bytes := make([]byte, 3)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("rand read: %w", err)
	}
	return fmt.Sprintf("%06x", uint32(bytes[0])<<16|uint32(bytes[1])<<8|uint32(bytes[2]))[:length], nil
}
