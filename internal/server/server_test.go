package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dotbrains/aptscout/internal/db"
	"github.com/dotbrains/aptscout/internal/models"
)

func mustDB(t *testing.T) *db.DB {
	t.Helper()
	d, err := db.OpenMemory()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = d.Close() })
	return d
}

func seedData(t *testing.T, database *db.DB) {
	t.Helper()
	fp := models.FloorPlan{
		Property: "test", Code: "A1", Bedrooms: 1, Bathrooms: 1,
		SqFt: 700, Features: []string{},
	}
	if err := database.UpsertFloorPlan(fp); err != nil {
		t.Fatal(err)
	}
	apt := models.Apartment{
		Property: "test", UnitNumber: "100", FloorPlan: "A1",
		Price: 1500, AvailableNow: true, Amenities: []string{}, IsAvailable: true,
	}
	if _, _, err := database.UpsertApartment(apt); err != nil {
		t.Fatal(err)
	}
	if err := database.InsertPriceHistory("test", "100", 1500); err != nil {
		t.Fatal(err)
	}
	_, _ = database.InsertScrapeRun("test")
}

func TestHandleApartments_GET(t *testing.T) {
	database := mustDB(t)
	seedData(t, database)
	srv := New(database)

	req := httptest.NewRequest(http.MethodGet, "/api/apartments", nil)
	rec := httptest.NewRecorder()
	srv.mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("expected application/json, got %s", ct)
	}

	var apts []models.Apartment
	if err := json.NewDecoder(rec.Body).Decode(&apts); err != nil {
		t.Fatal(err)
	}
	if len(apts) != 1 {
		t.Errorf("expected 1 apartment, got %d", len(apts))
	}
}

func TestHandleApartments_WithFilters(t *testing.T) {
	database := mustDB(t)
	seedData(t, database)
	srv := New(database)

	req := httptest.NewRequest(http.MethodGet, "/api/apartments?beds=1&baths=1&min_price=1000&max_price=2000&property=test&sort=price&order=asc", nil)
	rec := httptest.NewRecorder()
	srv.mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var apts []models.Apartment
	if err := json.NewDecoder(rec.Body).Decode(&apts); err != nil {
		t.Fatal(err)
	}
	if len(apts) != 1 {
		t.Errorf("expected 1 apartment, got %d", len(apts))
	}
}

func TestHandleApartments_MethodNotAllowed(t *testing.T) {
	srv := New(mustDB(t))

	req := httptest.NewRequest(http.MethodPost, "/api/apartments", nil)
	rec := httptest.NewRecorder()
	srv.mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rec.Code)
	}
}

func TestHandleApartments_EmptyResult(t *testing.T) {
	srv := New(mustDB(t))

	req := httptest.NewRequest(http.MethodGet, "/api/apartments", nil)
	rec := httptest.NewRecorder()
	srv.mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var apts []models.Apartment
	if err := json.NewDecoder(rec.Body).Decode(&apts); err != nil {
		t.Fatal(err)
	}
	if len(apts) != 0 {
		t.Errorf("expected empty array, got %d", len(apts))
	}
}

func TestHandleApartmentDetail_Found(t *testing.T) {
	database := mustDB(t)
	seedData(t, database)
	srv := New(database)

	req := httptest.NewRequest(http.MethodGet, "/api/apartments/test/100", nil)
	rec := httptest.NewRecorder()
	srv.mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var detail models.ApartmentDetail
	if err := json.NewDecoder(rec.Body).Decode(&detail); err != nil {
		t.Fatal(err)
	}
	if detail.Apartment.UnitNumber != "100" {
		t.Errorf("expected unit 100, got %s", detail.Apartment.UnitNumber)
	}
	if len(detail.PriceHistory) != 1 {
		t.Errorf("expected 1 price record, got %d", len(detail.PriceHistory))
	}
}

func TestHandleApartmentDetail_NotFound(t *testing.T) {
	srv := New(mustDB(t))

	req := httptest.NewRequest(http.MethodGet, "/api/apartments/test/999", nil)
	rec := httptest.NewRecorder()
	srv.mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

func TestHandleApartmentDetail_BadPath(t *testing.T) {
	srv := New(mustDB(t))

	req := httptest.NewRequest(http.MethodGet, "/api/apartments/invalid", nil)
	rec := httptest.NewRecorder()
	srv.mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestHandleApartmentDetail_MethodNotAllowed(t *testing.T) {
	srv := New(mustDB(t))

	req := httptest.NewRequest(http.MethodPost, "/api/apartments/test/100", nil)
	rec := httptest.NewRecorder()
	srv.mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rec.Code)
	}
}

func TestHandleFloorPlans_GET(t *testing.T) {
	database := mustDB(t)
	seedData(t, database)
	srv := New(database)

	req := httptest.NewRequest(http.MethodGet, "/api/floor-plans", nil)
	rec := httptest.NewRecorder()
	srv.mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var plans []models.FloorPlan
	if err := json.NewDecoder(rec.Body).Decode(&plans); err != nil {
		t.Fatal(err)
	}
	if len(plans) != 1 {
		t.Errorf("expected 1 floor plan, got %d", len(plans))
	}
}

func TestHandleFloorPlans_WithPropertyFilter(t *testing.T) {
	database := mustDB(t)
	seedData(t, database)
	srv := New(database)

	req := httptest.NewRequest(http.MethodGet, "/api/floor-plans?property=test", nil)
	rec := httptest.NewRecorder()
	srv.mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestHandleFloorPlans_MethodNotAllowed(t *testing.T) {
	srv := New(mustDB(t))

	req := httptest.NewRequest(http.MethodDelete, "/api/floor-plans", nil)
	rec := httptest.NewRecorder()
	srv.mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rec.Code)
	}
}

func TestHandleStats_GET(t *testing.T) {
	database := mustDB(t)
	seedData(t, database)
	srv := New(database)

	req := httptest.NewRequest(http.MethodGet, "/api/stats", nil)
	rec := httptest.NewRecorder()
	srv.mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var stats models.Stats
	if err := json.NewDecoder(rec.Body).Decode(&stats); err != nil {
		t.Fatal(err)
	}
	if stats.Available != 1 {
		t.Errorf("expected 1 available, got %d", stats.Available)
	}
}

func TestHandleStats_WithPropertyFilter(t *testing.T) {
	database := mustDB(t)
	seedData(t, database)
	srv := New(database)

	req := httptest.NewRequest(http.MethodGet, "/api/stats?property=test", nil)
	rec := httptest.NewRecorder()
	srv.mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestHandleStats_MethodNotAllowed(t *testing.T) {
	srv := New(mustDB(t))

	req := httptest.NewRequest(http.MethodPost, "/api/stats", nil)
	rec := httptest.NewRecorder()
	srv.mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rec.Code)
	}
}

func TestHandleScrapeRuns_GET(t *testing.T) {
	database := mustDB(t)
	seedData(t, database)
	srv := New(database)

	req := httptest.NewRequest(http.MethodGet, "/api/scrape-runs", nil)
	rec := httptest.NewRecorder()
	srv.mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var runs []models.ScrapeRun
	if err := json.NewDecoder(rec.Body).Decode(&runs); err != nil {
		t.Fatal(err)
	}
	if len(runs) != 1 {
		t.Errorf("expected 1 scrape run, got %d", len(runs))
	}
}

func TestHandleScrapeRuns_MethodNotAllowed(t *testing.T) {
	srv := New(mustDB(t))

	req := httptest.NewRequest(http.MethodPost, "/api/scrape-runs", nil)
	rec := httptest.NewRecorder()
	srv.mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rec.Code)
	}
}

func TestHandleScrape_MethodNotAllowed(t *testing.T) {
	srv := New(mustDB(t))

	req := httptest.NewRequest(http.MethodGet, "/api/scrape", nil)
	rec := httptest.NewRecorder()
	srv.mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rec.Code)
	}
}

func TestHandleApartments_RenovatedFilter(t *testing.T) {
	database := mustDB(t)
	seedData(t, database)
	srv := New(database)

	req := httptest.NewRequest(http.MethodGet, "/api/apartments?renovated=true", nil)
	rec := httptest.NewRecorder()
	srv.mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestHandleApartments_PlanFilter(t *testing.T) {
	database := mustDB(t)
	seedData(t, database)
	srv := New(database)

	req := httptest.NewRequest(http.MethodGet, "/api/apartments?plan=A1", nil)
	rec := httptest.NewRecorder()
	srv.mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestHandleApartments_AvailableByFilter(t *testing.T) {
	database := mustDB(t)
	seedData(t, database)
	srv := New(database)

	req := httptest.NewRequest(http.MethodGet, "/api/apartments?available_by=2026-12-31", nil)
	rec := httptest.NewRecorder()
	srv.mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestHandleScrape_POST(t *testing.T) {
	database := mustDB(t)
	srv := New(database)

	req := httptest.NewRequest(http.MethodPost, "/api/scrape", nil)
	rec := httptest.NewRecorder()
	srv.mux.ServeHTTP(rec, req)

	// This triggers real provider scraping which may fail on network, but
	// should still return 200 with a result (errors are swallowed per-provider).
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestDiscard_Write(t *testing.T) {
	d := discard{}
	n, err := d.Write([]byte("test"))
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if n != 4 {
		t.Errorf("expected 4, got %d", n)
	}
}

func TestStaticFiles(t *testing.T) {
	srv := New(mustDB(t))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	srv.mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200 for static index, got %d", rec.Code)
	}
}
