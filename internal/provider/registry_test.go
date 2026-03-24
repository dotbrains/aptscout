package provider

import (
	"sort"
	"testing"
)

func TestGet_Exists(t *testing.T) {
	p := Get("desert-club")
	if p == nil {
		t.Fatal("expected desert-club provider, got nil")
	}
	if p.ID() != "desert-club" {
		t.Errorf("expected ID 'desert-club', got %q", p.ID())
	}
}

func TestGet_NotFound(t *testing.T) {
	p := Get("nonexistent")
	if p != nil {
		t.Errorf("expected nil for unknown provider, got %v", p)
	}
}

func TestIDs(t *testing.T) {
	ids := IDs()
	if len(ids) < 2 {
		t.Fatalf("expected at least 2 provider IDs, got %d", len(ids))
	}

	sort.Strings(ids)
	expected := []string{"desert-club", "hideaway"}
	for i, id := range expected {
		if ids[i] != id {
			t.Errorf("expected IDs[%d] = %q, got %q", i, id, ids[i])
		}
	}
}
