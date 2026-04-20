package id

import (
	"strings"
	"testing"
)

func TestGenerate(t *testing.T) {
	id, err := Generate(nil)
	if err != nil {
		t.Fatalf("Generate(nil) = %v", err)
	}
	if len(id) != length {
		t.Fatalf("Generate(nil) len = %d, want %d", len(id), length)
	}
	for _, c := range id {
		if !strings.ContainsRune("0123456789abcdef", c) {
			t.Fatalf("Generate(nil) = %q, contains non-hex char %c", id, c)
		}
	}
}

func TestGenerateNoCollisionWithExisting(t *testing.T) {
	existing := []string{"aabbcc", "ddeeff", "001122"}
	id, err := Generate(existing)
	if err != nil {
		t.Fatalf("Generate(existing) = %v", err)
	}
	for _, e := range existing {
		if id == e {
			t.Fatalf("Generate(existing) = %q, collides with %q", id, e)
		}
	}
}

func TestGenerateUniqueness(t *testing.T) {
	ids := make(map[string]struct{}, 100)
	for i := 0; i < 100; i++ {
		id, err := Generate(nil)
		if err != nil {
			t.Fatalf("Generate(nil) iteration %d: %v", i, err)
		}
		if len(id) != length {
			t.Fatalf("Generate(nil) iteration %d: len=%d, want %d", i, len(id), length)
		}
		if _, ok := ids[id]; ok {
			t.Fatalf("Generate(nil) produced duplicate %q", id)
		}
		ids[id] = struct{}{}
	}
}
