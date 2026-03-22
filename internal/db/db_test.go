package db

import (
	"testing"

	"github.com/dotbrains/aptscout/internal/models"
)

func mustOpen(t *testing.T) *DB {
	t.Helper()
	db, err := OpenMemory()
	if err != nil {
		t.Fatalf("OpenMemory: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func seedFloorPlan(t *testing.T, db *DB, property, code string, beds, baths, sqft int) {
	t.Helper()
	err := db.UpsertFloorPlan(models.FloorPlan{
		Property: property, Code: code, Bedrooms: beds, Bathrooms: baths,
		SqFt: sqft, Deposit: 700, IsRenovated: true, Features: []string{},
	})
	if err != nil {
		t.Fatalf("UpsertFloorPlan: %v", err)
	}
}

func seedApartment(t *testing.T, db *DB, property, unit, plan string, price int) {
	t.Helper()
	isNew, _, err := db.UpsertApartment(models.Apartment{
		Property: property, UnitNumber: unit, FloorPlan: plan,
		Price: price, AvailableNow: true, Amenities: []string{},
	})
	if err != nil {
		t.Fatalf("UpsertApartment: %v", err)
	}
	if !isNew {
		t.Fatalf("expected new unit %s", unit)
	}
}

// --- Floor Plans ---

func TestUpsertFloorPlan(t *testing.T) {
	db := mustOpen(t)
	seedFloorPlan(t, db, "test", "A1R", 1, 1, 715)

	plans, err := db.GetFloorPlans(nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(plans) != 1 {
		t.Fatalf("expected 1 plan, got %d", len(plans))
	}
	if plans[0].Code != "A1R" || plans[0].Bedrooms != 1 || plans[0].SqFt != 715 {
		t.Errorf("unexpected plan: %+v", plans[0])
	}
}

func TestUpsertFloorPlan_Update(t *testing.T) {
	db := mustOpen(t)
	seedFloorPlan(t, db, "test", "A1R", 1, 1, 715)
	// Update sqft
	seedFloorPlan(t, db, "test", "A1R", 1, 1, 800)

	plans, _ := db.GetFloorPlans(nil)
	if len(plans) != 1 || plans[0].SqFt != 800 {
		t.Errorf("expected updated sqft=800, got %+v", plans)
	}
}

func TestGetFloorPlans_PropertyFilter(t *testing.T) {
	db := mustOpen(t)
	seedFloorPlan(t, db, "prop-a", "X1", 1, 1, 500)
	seedFloorPlan(t, db, "prop-b", "Y1", 2, 2, 900)

	propA := "prop-a"
	plans, _ := db.GetFloorPlans(&propA)
	if len(plans) != 1 || plans[0].Code != "X1" {
		t.Errorf("expected 1 plan for prop-a, got %d", len(plans))
	}

	all, _ := db.GetFloorPlans(nil)
	if len(all) != 2 {
		t.Errorf("expected 2 total plans, got %d", len(all))
	}
}

// --- Apartments ---

func TestUpsertApartment_New(t *testing.T) {
	db := mustOpen(t)
	seedFloorPlan(t, db, "test", "B2R", 2, 2, 1142)
	seedApartment(t, db, "test", "2146", "B2R", 2095)

	apt, err := db.GetApartment("test", "2146")
	if err != nil {
		t.Fatal(err)
	}
	if apt.Price != 2095 || apt.FloorPlan != "B2R" {
		t.Errorf("unexpected apartment: %+v", apt)
	}
}

func TestUpsertApartment_PriceChange(t *testing.T) {
	db := mustOpen(t)
	seedFloorPlan(t, db, "test", "B2R", 2, 2, 1142)
	seedApartment(t, db, "test", "2146", "B2R", 2095)

	// Update price
	_, changed, err := db.UpsertApartment(models.Apartment{
		Property: "test", UnitNumber: "2146", FloorPlan: "B2R",
		Price: 2010, AvailableNow: true, Amenities: []string{},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !changed {
		t.Error("expected priceChanged=true")
	}

	apt, _ := db.GetApartment("test", "2146")
	if apt.Price != 2010 {
		t.Errorf("expected price 2010, got %d", apt.Price)
	}
}

func TestUpsertApartment_NoPriceChange(t *testing.T) {
	db := mustOpen(t)
	seedFloorPlan(t, db, "test", "B2R", 2, 2, 1142)
	seedApartment(t, db, "test", "2146", "B2R", 2095)

	_, changed, _ := db.UpsertApartment(models.Apartment{
		Property: "test", UnitNumber: "2146", FloorPlan: "B2R",
		Price: 2095, AvailableNow: true, Amenities: []string{},
	})
	if changed {
		t.Error("expected priceChanged=false for same price")
	}
}

func TestMarkUnavailable(t *testing.T) {
	db := mustOpen(t)
	seedFloorPlan(t, db, "test", "A1R", 1, 1, 715)
	seedApartment(t, db, "test", "100", "A1R", 1600)
	seedApartment(t, db, "test", "101", "A1R", 1700)
	seedApartment(t, db, "test", "102", "A1R", 1800)

	removed, err := db.MarkUnavailable("test", []string{"100"})
	if err != nil {
		t.Fatal(err)
	}
	if removed != 2 {
		t.Errorf("expected 2 removed, got %d", removed)
	}

	apts, _ := db.ListApartments(models.ApartmentFilter{Sort: "price"})
	if len(apts) != 1 || apts[0].UnitNumber != "100" {
		t.Errorf("expected only unit 100 available, got %+v", apts)
	}
}

func TestMarkUnavailable_Empty(t *testing.T) {
	db := mustOpen(t)
	seedFloorPlan(t, db, "test", "A1R", 1, 1, 715)
	seedApartment(t, db, "test", "100", "A1R", 1600)

	removed, _ := db.MarkUnavailable("test", []string{})
	if removed != 1 {
		t.Errorf("expected 1 removed, got %d", removed)
	}
}

func TestMarkUnavailable_ScopedToProperty(t *testing.T) {
	db := mustOpen(t)
	seedFloorPlan(t, db, "prop-a", "A1R", 1, 1, 715)
	seedFloorPlan(t, db, "prop-b", "A1R", 1, 1, 715)
	seedApartment(t, db, "prop-a", "100", "A1R", 1600)
	seedApartment(t, db, "prop-b", "200", "A1R", 1700)

	// Mark prop-a unavailable — prop-b should be untouched
	_, _ = db.MarkUnavailable("prop-a", []string{})

	apts, _ := db.ListApartments(models.ApartmentFilter{Sort: "price"})
	if len(apts) != 1 || apts[0].Property != "prop-b" {
		t.Errorf("expected only prop-b unit, got %+v", apts)
	}
}

func TestListApartments_Filters(t *testing.T) {
	db := mustOpen(t)
	seedFloorPlan(t, db, "test", "A1R", 1, 1, 715)
	seedFloorPlan(t, db, "test", "B2R", 2, 2, 1142)
	seedApartment(t, db, "test", "100", "A1R", 1600)
	seedApartment(t, db, "test", "200", "B2R", 2095)
	seedApartment(t, db, "test", "201", "B2R", 2200)

	// Filter by beds
	beds := 2
	apts, _ := db.ListApartments(models.ApartmentFilter{Beds: &beds, Sort: "price"})
	if len(apts) != 2 {
		t.Errorf("expected 2 two-bed units, got %d", len(apts))
	}

	// Filter by max price
	maxP := 2000
	apts, _ = db.ListApartments(models.ApartmentFilter{MaxPrice: &maxP, Sort: "price"})
	if len(apts) != 1 || apts[0].UnitNumber != "100" {
		t.Errorf("expected 1 unit under $2000, got %+v", apts)
	}

	// Sort desc
	apts, _ = db.ListApartments(models.ApartmentFilter{Sort: "price", Order: "desc"})
	if len(apts) != 3 || apts[0].Price != 2200 {
		t.Errorf("expected desc sort, first price=2200, got %+v", apts)
	}
}

func TestListApartments_PropertyFilter(t *testing.T) {
	db := mustOpen(t)
	seedFloorPlan(t, db, "prop-a", "A1R", 1, 1, 715)
	seedFloorPlan(t, db, "prop-b", "B1R", 2, 2, 1090)
	seedApartment(t, db, "prop-a", "100", "A1R", 1600)
	seedApartment(t, db, "prop-b", "200", "B1R", 2000)

	propA := "prop-a"
	apts, _ := db.ListApartments(models.ApartmentFilter{Property: &propA, Sort: "price"})
	if len(apts) != 1 || apts[0].Property != "prop-a" {
		t.Errorf("expected 1 prop-a unit, got %d", len(apts))
	}
}

// --- Price History ---

func TestPriceHistory(t *testing.T) {
	db := mustOpen(t)
	seedFloorPlan(t, db, "test", "B2R", 2, 2, 1142)
	seedApartment(t, db, "test", "2146", "B2R", 2095)

	_ = db.InsertPriceHistory("test", "2146", 2095)
	_ = db.InsertPriceHistory("test", "2146", 2010)

	records, err := db.GetPriceHistory("test", "2146")
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != 2 {
		t.Fatalf("expected 2 records, got %d", len(records))
	}
	if records[0].Price != 2095 || records[1].Price != 2010 {
		t.Errorf("unexpected prices: %d, %d", records[0].Price, records[1].Price)
	}
}

// --- Scrape Runs ---

func TestScrapeRuns(t *testing.T) {
	db := mustOpen(t)

	id, err := db.InsertScrapeRun("test-prop")
	if err != nil || id == 0 {
		t.Fatalf("InsertScrapeRun: id=%d err=%v", id, err)
	}

	err = db.CompleteScrapeRun(id, models.ScrapeRun{FloorPlans: 5, UnitsFound: 10, UnitsNew: 3})
	if err != nil {
		t.Fatal(err)
	}

	runs, _ := db.GetScrapeRuns()
	if len(runs) != 1 {
		t.Fatalf("expected 1 run, got %d", len(runs))
	}
	if runs[0].Property != "test-prop" || runs[0].UnitsFound != 10 {
		t.Errorf("unexpected run: %+v", runs[0])
	}
}

// --- Stats ---

func TestGetStats(t *testing.T) {
	db := mustOpen(t)
	seedFloorPlan(t, db, "test", "A1R", 1, 1, 715)
	seedFloorPlan(t, db, "test", "B2R", 2, 2, 1142)
	seedApartment(t, db, "test", "100", "A1R", 1600)
	seedApartment(t, db, "test", "200", "B2R", 2095)
	seedApartment(t, db, "test", "201", "B2R", 2200)

	stats, err := db.GetStats(nil)
	if err != nil {
		t.Fatal(err)
	}
	if stats.FloorPlans != 2 {
		t.Errorf("expected 2 floor plans, got %d", stats.FloorPlans)
	}
	if stats.Available != 3 {
		t.Errorf("expected 3 available, got %d", stats.Available)
	}
	if len(stats.ByBedrooms) != 2 {
		t.Errorf("expected 2 bedroom groups, got %d", len(stats.ByBedrooms))
	}
}

func TestGetStats_PropertyFilter(t *testing.T) {
	db := mustOpen(t)
	seedFloorPlan(t, db, "prop-a", "A1R", 1, 1, 715)
	seedFloorPlan(t, db, "prop-b", "B1R", 2, 2, 1090)
	seedApartment(t, db, "prop-a", "100", "A1R", 1600)
	seedApartment(t, db, "prop-b", "200", "B1R", 2000)

	propA := "prop-a"
	stats, _ := db.GetStats(&propA)
	if stats.FloorPlans != 1 || stats.Available != 1 {
		t.Errorf("expected 1/1 for prop-a, got %d/%d", stats.FloorPlans, stats.Available)
	}
}

func TestCleanStale(t *testing.T) {
	db := mustOpen(t)
	seedFloorPlan(t, db, "test", "A1R", 1, 1, 715)
	seedApartment(t, db, "test", "100", "A1R", 1600)
	// Mark unavailable
	_, _ = db.MarkUnavailable("test", []string{})
	// Backdate last_seen so CleanStale picks it up
	_, _ = db.Exec(`UPDATE apartments SET last_seen = '2020-01-01T00:00:00Z' WHERE unit_number = '100'`)

	removed, err := db.CleanStale(1)
	if err != nil {
		t.Fatal(err)
	}
	if removed != 1 {
		t.Errorf("expected 1 cleaned, got %d", removed)
	}
}
