package identifier

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

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

func TestGenerateIdentifier_Format(t *testing.T) {
	id, err := GenerateIdentifier()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(id) != 36 {
		t.Errorf("expected length 36, got %d", len(id))
	}
	if id[8] != '-' || id[13] != '-' || id[18] != '-' || id[23] != '-' {
		t.Errorf("expected UUID format with dashes, got %s", id)
	}
}

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
