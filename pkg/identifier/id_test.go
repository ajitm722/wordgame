package identifier

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// UUID v4 format: 8-4-4-4-12 hex digits separated by hyphens (36 chars total).
const (
	uuidStringLen = 36
	hyph1         = 8  // position of first hyphen
	hyph2         = 13 // position of second hyphen
	hyph3         = 18 // position of third hyphen
	hyph4         = 23 // position of fourth hyphen
)

// TestGenerateIdentifier verifies GenerateIdentifier returns non-empty, unique identifiers.
func TestGenerateIdentifier(t *testing.T) {
	id1, err := GenerateIdentifier()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id1 == "" {
		t.Error("expected non-empty identifier")
	}

	id2, err := GenerateIdentifier()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id2 == "" {
		t.Error("expected non-empty identifier")
	}

	if id1 == id2 {
		t.Error("expected unique identifiers")
	}
}

// TestGenerateIdentifier_Format verifies the generated identifier matches UUID v4 format.
func TestGenerateIdentifier_Format(t *testing.T) {
	id, err := GenerateIdentifier()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(id) != uuidStringLen {
		t.Errorf("expected length %d, got %d", uuidStringLen, len(id))
	}
	if id[hyph1] != '-' || id[hyph2] != '-' || id[hyph3] != '-' || id[hyph4] != '-' {
		t.Errorf("expected UUID format with dashes, got %s", id)
	}
}

// TestGenerateIdentifier_Error verifies GenerateIdentifier propagates UUID generator errors.
func TestGenerateIdentifier_Error(t *testing.T) {
	// Swap in a failing generator to cover the error path
	orig := newUUID
	defer func() { newUUID = orig }()

	newUUID = func() (uuid.UUID, error) {
		return uuid.Nil, errors.New("crypto/rand failure")
	}

	id, err := GenerateIdentifier()
	if err == nil {
		t.Fatal("expected error from failing UUID generator")
	}
	if id != "" {
		t.Errorf("expected empty id on error, got %q", id)
	}
}
