package cmd

import (
	"testing"
)

func TestPlural(t *testing.T) {
	tests := []struct {
		n    int
		want string
	}{
		{0, "s"},
		{1, ""},
		{2, "s"},
		{10, "s"},
	}
	for _, tt := range tests {
		got := plural(tt.n)
		if got != tt.want {
			t.Errorf("plural(%d) = %q, want %q", tt.n, got, tt.want)
		}
	}
}

func TestJoinParts(t *testing.T) {
	tests := []struct {
		parts []string
		want  string
	}{
		{nil, ""},
		{[]string{"a"}, "a"},
		{[]string{"a", "b"}, "a, b"},
		{[]string{"1 new", "2 changed", "3 removed"}, "1 new, 2 changed, 3 removed"},
	}
	for _, tt := range tests {
		got := joinParts(tt.parts)
		if got != tt.want {
			t.Errorf("joinParts(%v) = %q, want %q", tt.parts, got, tt.want)
		}
	}
}

func TestResolveProviders_All(t *testing.T) {
	flagProperty = ""
	providers := resolveProviders()
	if len(providers) < 2 {
		t.Errorf("expected at least 2 providers, got %d", len(providers))
	}
}

func TestResolveProviders_Specific(t *testing.T) {
	flagProperty = "desert-club"
	defer func() { flagProperty = "" }()

	providers := resolveProviders()
	if len(providers) != 1 {
		t.Fatalf("expected 1 provider, got %d", len(providers))
	}
	if providers[0].ID() != "desert-club" {
		t.Errorf("expected desert-club, got %s", providers[0].ID())
	}
}

func TestResolveProviders_Unknown(t *testing.T) {
	flagProperty = "nonexistent"
	defer func() { flagProperty = "" }()

	providers := resolveProviders()
	if providers != nil {
		t.Errorf("expected nil for unknown property, got %v", providers)
	}
}

func TestPropertyFilter(t *testing.T) {
	flagProperty = ""
	if propertyFilter() != nil {
		t.Error("expected nil when flagProperty is empty")
	}

	flagProperty = "test"
	defer func() { flagProperty = "" }()
	pf := propertyFilter()
	if pf == nil || *pf != "test" {
		t.Errorf("expected pointer to 'test', got %v", pf)
	}
}

func TestExecute(t *testing.T) {
	// Just verify Execute doesn't panic with --version.
	err := Execute("test")
	// Execute with no args prints help and returns nil.
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
