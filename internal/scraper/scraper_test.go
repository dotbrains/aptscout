package scraper

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dotbrains/aptscout/internal/db"
	"github.com/dotbrains/aptscout/internal/models"
)

// mockProvider is a fake provider for testing.
type mockProvider struct {
	id   string
	name string
	data *models.ScrapeData
	err  error
}

func (m *mockProvider) ID() string   { return m.id }
func (m *mockProvider) Name() string { return m.name }
func (m *mockProvider) Scrape(ctx context.Context, fetch models.Fetcher) (*models.ScrapeData, error) {
	return m.data, m.err
}

func mustDB(t *testing.T) *db.DB {
	t.Helper()
	d, err := db.OpenMemory()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = d.Close() })
	return d
}

func TestRunProvider_NewUnits(t *testing.T) {
	database := mustDB(t)
	var buf bytes.Buffer
	s := New(database, &buf)

	prov := &mockProvider{
		id: "test", name: "Test Property",
		data: &models.ScrapeData{
			FloorPlans: []models.FloorPlan{
				{Property: "test", Code: "A1", Bedrooms: 1, Bathrooms: 1, SqFt: 700, Features: []string{}},
			},
			Apartments: []models.Apartment{
				{Property: "test", UnitNumber: "100", FloorPlan: "A1", Price: 1500, AvailableNow: true, Amenities: []string{}},
				{Property: "test", UnitNumber: "101", FloorPlan: "A1", Price: 1600, AvailableNow: true, Amenities: []string{}},
			},
		},
	}

	result, err := s.RunProvider(context.Background(), prov)
	if err != nil {
		t.Fatal(err)
	}
	if result.FloorPlans != 1 {
		t.Errorf("expected 1 floor plan, got %d", result.FloorPlans)
	}
	if result.UnitsFound != 2 {
		t.Errorf("expected 2 units, got %d", result.UnitsFound)
	}
	if result.UnitsNew != 2 {
		t.Errorf("expected 2 new units, got %d", result.UnitsNew)
	}

	// Verify price history was recorded for new units
	history, _ := database.GetPriceHistory("test", "100")
	if len(history) != 1 || history[0].Price != 1500 {
		t.Errorf("expected 1 price record for unit 100, got %+v", history)
	}
}

func TestRunProvider_PriceChange(t *testing.T) {
	database := mustDB(t)
	var buf bytes.Buffer
	s := New(database, &buf)

	prov := &mockProvider{
		id: "test", name: "Test",
		data: &models.ScrapeData{
			FloorPlans: []models.FloorPlan{
				{Property: "test", Code: "A1", Bedrooms: 1, Bathrooms: 1, SqFt: 700, Features: []string{}},
			},
			Apartments: []models.Apartment{
				{Property: "test", UnitNumber: "100", FloorPlan: "A1", Price: 1500, AvailableNow: true, Amenities: []string{}},
			},
		},
	}

	// First scrape
	_, _ = s.RunProvider(context.Background(), prov)

	// Second scrape with price change
	prov.data.Apartments[0].Price = 1400
	result, _ := s.RunProvider(context.Background(), prov)

	if result.UnitsChanged != 1 {
		t.Errorf("expected 1 changed, got %d", result.UnitsChanged)
	}
	if result.UnitsNew != 0 {
		t.Errorf("expected 0 new on second scrape, got %d", result.UnitsNew)
	}

	history, _ := database.GetPriceHistory("test", "100")
	if len(history) != 2 {
		t.Errorf("expected 2 price records, got %d", len(history))
	}
}

func TestRunProvider_MarksUnavailable(t *testing.T) {
	database := mustDB(t)
	var buf bytes.Buffer
	s := New(database, &buf)

	prov := &mockProvider{
		id: "test", name: "Test",
		data: &models.ScrapeData{
			FloorPlans: []models.FloorPlan{
				{Property: "test", Code: "A1", Bedrooms: 1, Bathrooms: 1, SqFt: 700, Features: []string{}},
			},
			Apartments: []models.Apartment{
				{Property: "test", UnitNumber: "100", FloorPlan: "A1", Price: 1500, AvailableNow: true, Amenities: []string{}},
				{Property: "test", UnitNumber: "101", FloorPlan: "A1", Price: 1600, AvailableNow: true, Amenities: []string{}},
			},
		},
	}

	// First scrape: 2 units
	_, _ = s.RunProvider(context.Background(), prov)

	// Second scrape: only unit 100
	prov.data.Apartments = prov.data.Apartments[:1]
	result, _ := s.RunProvider(context.Background(), prov)

	if result.UnitsRemoved != 1 {
		t.Errorf("expected 1 removed, got %d", result.UnitsRemoved)
	}

	apts, _ := database.ListApartments(models.ApartmentFilter{Sort: "price"})
	if len(apts) != 1 || apts[0].UnitNumber != "100" {
		t.Errorf("expected only unit 100 available, got %+v", apts)
	}
}

func TestRunProvider_ScrapeRunRecorded(t *testing.T) {
	database := mustDB(t)
	var buf bytes.Buffer
	s := New(database, &buf)

	prov := &mockProvider{
		id: "test", name: "Test",
		data: &models.ScrapeData{
			FloorPlans: []models.FloorPlan{
				{Property: "test", Code: "A1", Bedrooms: 1, Bathrooms: 1, SqFt: 700, Features: []string{}},
			},
			Apartments: []models.Apartment{},
		},
	}

	_, _ = s.RunProvider(context.Background(), prov)

	runs, _ := database.GetScrapeRuns()
	if len(runs) != 1 || runs[0].Property != "test" {
		t.Errorf("expected 1 scrape run for 'test', got %+v", runs)
	}
}

func TestRunProvider_ScrapeError(t *testing.T) {
	database := mustDB(t)
	var buf bytes.Buffer
	s := New(database, &buf)

	prov := &mockProvider{
		id: "test", name: "Test",
		data: &models.ScrapeData{},
		err:  context.DeadlineExceeded,
	}

	_, err := s.RunProvider(context.Background(), prov)
	if err == nil {
		t.Error("expected error from provider")
	}

	// Scrape run should still be recorded with error
	runs, _ := database.GetScrapeRuns()
	if len(runs) != 1 || runs[0].Error == nil {
		t.Error("expected scrape run with error recorded")
	}
}

// --- fetch tests ---

func TestFetch_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, "<html>ok</html>")
	}))
	defer ts.Close()

	database := mustDB(t)
	var buf bytes.Buffer
	s := New(database, &buf)

	body, err := s.fetch(context.Background(), ts.URL+"/page")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if body != "<html>ok</html>" {
		t.Errorf("unexpected body: %q", body)
	}
}

func TestFetch_SendsBrowserHeaders(t *testing.T) {
	var gotHeaders http.Header
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotHeaders = r.Header
		_, _ = fmt.Fprint(w, "ok")
	}))
	defer ts.Close()

	database := mustDB(t)
	var buf bytes.Buffer
	s := New(database, &buf)

	_, _ = s.fetch(context.Background(), ts.URL+"/page")

	for _, h := range []string{"User-Agent", "Sec-Fetch-Dest", "Sec-Fetch-Mode", "Referer"} {
		if gotHeaders.Get(h) == "" {
			t.Errorf("expected %s header to be set", h)
		}
	}
}

func TestFetch_403_BailsAfterTwoConsecutive(t *testing.T) {
	var attempts int
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.WriteHeader(403)
	}))
	defer ts.Close()

	database := mustDB(t)
	var buf bytes.Buffer
	s := New(database, &buf)
	// Override the client to skip real backoff delays.
	s.client = ts.Client()
	s.client.Timeout = 0

	_, err := s.fetch(context.Background(), ts.URL+"/page")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "403") {
		t.Errorf("expected 403 in error, got: %v", err)
	}
	// Should stop after 2 consecutive 403s (attempts 0 and 1), not exhaust all retries.
	if attempts > 3 {
		t.Errorf("expected at most 3 attempts, got %d", attempts)
	}
}

func TestFetch_429_RetriesAllAttempts(t *testing.T) {
	var attempts int
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts <= 2 {
			w.WriteHeader(429)
			return
		}
		_, _ = fmt.Fprint(w, "ok")
	}))
	defer ts.Close()

	database := mustDB(t)
	var buf bytes.Buffer
	s := New(database, &buf)
	s.client = ts.Client()

	body, err := s.fetch(context.Background(), ts.URL+"/page")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if body != "ok" {
		t.Errorf("unexpected body: %q", body)
	}
	if attempts != 3 {
		t.Errorf("expected 3 attempts, got %d", attempts)
	}
}

func TestFetch_NonRetryableStatus(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
	}))
	defer ts.Close()

	database := mustDB(t)
	var buf bytes.Buffer
	s := New(database, &buf)

	_, err := s.fetch(context.Background(), ts.URL+"/page")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "500") {
		t.Errorf("expected 500 in error, got: %v", err)
	}
}

// --- originOf tests ---

func TestOriginOf(t *testing.T) {
	tests := []struct {
		url  string
		want string
	}{
		{"https://example.com/path/page", "https://example.com"},
		{"http://localhost:8080/api/test", "http://localhost:8080"},
		{"https://sub.domain.com:443/", "https://sub.domain.com:443"},
	}
	for _, tt := range tests {
		got := originOf(tt.url)
		if got != tt.want {
			t.Errorf("originOf(%q) = %q, want %q", tt.url, got, tt.want)
		}
	}
}
