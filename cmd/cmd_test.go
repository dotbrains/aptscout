package cmd

import (
	"bytes"
	"path/filepath"
	"testing"

	"github.com/dotbrains/aptscout/internal/db"
	"github.com/dotbrains/aptscout/internal/models"
)

func testDBPath(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "test.db")
}

func seedTestDB(t *testing.T, dbPath string) {
	t.Helper()
	database, err := db.Open(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = database.Close() }()

	_ = database.UpsertFloorPlan(models.FloorPlan{
		Property: "test", Code: "A1", Bedrooms: 1, Bathrooms: 1,
		SqFt: 700, Features: []string{},
	})
	_, _, _ = database.UpsertApartment(models.Apartment{
		Property: "test", UnitNumber: "100", FloorPlan: "A1",
		Price: 1500, AvailableNow: true, Amenities: []string{}, IsAvailable: true,
	})
	_ = database.InsertPriceHistory("test", "100", 1500)
	_, _ = database.InsertScrapeRun("test")
}

func runCmd(t *testing.T, args ...string) (string, error) {
	t.Helper()
	root := newRootCmd("test")
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetErr(&buf)
	root.SetArgs(args)
	err := root.Execute()
	return buf.String(), err
}

func TestListCmd_Empty(t *testing.T) {
	dbPath := testDBPath(t)
	// Create empty DB.
	d, err := db.Open(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	_ = d.Close()

	out, err := runCmd(t, "list", "--db", dbPath)
	if err != nil {
		t.Fatalf("unexpected error: %v\noutput: %s", err, out)
	}
}

func TestListCmd_WithData(t *testing.T) {
	dbPath := testDBPath(t)
	seedTestDB(t, dbPath)

	out, err := runCmd(t, "list", "--db", dbPath)
	if err != nil {
		t.Fatalf("unexpected error: %v\noutput: %s", err, out)
	}
}

func TestListCmd_WithFilters(t *testing.T) {
	dbPath := testDBPath(t)
	seedTestDB(t, dbPath)

	out, err := runCmd(t, "list", "--db", dbPath, "--beds", "1", "--max-price", "2000")
	if err != nil {
		t.Fatalf("unexpected error: %v\noutput: %s", err, out)
	}
}

func TestListCmd_JSON(t *testing.T) {
	dbPath := testDBPath(t)
	seedTestDB(t, dbPath)

	out, err := runCmd(t, "list", "--db", dbPath, "--json")
	if err != nil {
		t.Fatalf("unexpected error: %v\noutput: %s", err, out)
	}
}

func TestStatsCmd(t *testing.T) {
	dbPath := testDBPath(t)
	seedTestDB(t, dbPath)

	out, err := runCmd(t, "stats", "--db", dbPath)
	if err != nil {
		t.Fatalf("unexpected error: %v\noutput: %s", err, out)
	}
}

func TestHistoryCmd_Found(t *testing.T) {
	dbPath := testDBPath(t)
	seedTestDB(t, dbPath)

	out, err := runCmd(t, "history", "--db", dbPath, "--property", "test", "100")
	if err != nil {
		t.Fatalf("unexpected error: %v\noutput: %s", err, out)
	}
}

func TestHistoryCmd_NotFound(t *testing.T) {
	dbPath := testDBPath(t)
	seedTestDB(t, dbPath)

	_, err := runCmd(t, "history", "--db", dbPath, "999")
	if err == nil {
		t.Error("expected error for unknown unit")
	}
}

func TestCleanCmd_DryRun(t *testing.T) {
	dbPath := testDBPath(t)
	seedTestDB(t, dbPath)

	out, err := runCmd(t, "clean", "--db", dbPath, "--dry-run")
	if err != nil {
		t.Fatalf("unexpected error: %v\noutput: %s", err, out)
	}
}

func TestCleanCmd(t *testing.T) {
	dbPath := testDBPath(t)
	seedTestDB(t, dbPath)

	out, err := runCmd(t, "clean", "--db", dbPath, "--days", "0")
	if err != nil {
		t.Fatalf("unexpected error: %v\noutput: %s", err, out)
	}
}

func TestPropertiesCmd(t *testing.T) {
	out, err := runCmd(t, "properties")
	if err != nil {
		t.Fatalf("unexpected error: %v\noutput: %s", err, out)
	}
}

func TestServeCmd_InvalidDB(t *testing.T) {
	// Serve with a bad DB path should fail.
	_, err := runCmd(t, "serve", "--db", "/nonexistent/path/db")
	if err == nil {
		t.Error("expected error for invalid DB path")
	}
}
