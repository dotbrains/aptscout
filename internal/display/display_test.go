package display

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/dotbrains/aptscout/internal/models"
)

func TestOrdinal(t *testing.T) {
	tests := []struct {
		n    int
		want string
	}{
		{0, "—"},
		{1, "1st"},
		{2, "2nd"},
		{3, "3rd"},
		{4, "4th"},
		{11, "11th"},
		{12, "12th"},
		{13, "13th"},
		{21, "21st"},
		{22, "22nd"},
		{23, "23rd"},
		{100, "100th"},
		{101, "101st"},
	}
	for _, tt := range tests {
		got := ordinal(tt.n)
		if got != tt.want {
			t.Errorf("ordinal(%d) = %q, want %q", tt.n, got, tt.want)
		}
	}
}

func TestFormatNumber(t *testing.T) {
	tests := []struct {
		n    int
		want string
	}{
		{0, "0"},
		{42, "42"},
		{999, "999"},
		{1000, "1,000"},
		{1142, "1,142"},
		{12345, "12,345"},
		{1000000, "1,000,000"},
	}
	for _, tt := range tests {
		got := formatNumber(tt.n)
		if got != tt.want {
			t.Errorf("formatNumber(%d) = %q, want %q", tt.n, got, tt.want)
		}
	}
}

func TestPlural(t *testing.T) {
	if plural(1) != "" {
		t.Error("plural(1) should be empty")
	}
	if plural(0) != "s" {
		t.Error("plural(0) should be 's'")
	}
	if plural(5) != "s" {
		t.Error("plural(5) should be 's'")
	}
}

func TestApartmentTable(t *testing.T) {
	date := "2026-04-17"
	apts := []models.Apartment{
		{UnitNumber: "2146", FloorPlan: "B2R", Bedrooms: 2, Bathrooms: 2, SqFt: 1142, Price: 2095, AvailableNow: true, Floor: 2, Amenities: []string{"Garage"}, Deposit: 700},
		{UnitNumber: "3169", FloorPlan: "B2R", Bedrooms: 2, Bathrooms: 2, SqFt: 1142, Price: 2010, AvailableDate: &date, Floor: 3, Amenities: []string{}},
	}

	var buf bytes.Buffer
	ApartmentTable(&buf, apts)
	out := buf.String()

	if !strings.Contains(out, "#2146") {
		t.Error("expected #2146 in output")
	}
	if !strings.Contains(out, "2026-04-17") {
		t.Error("expected date in output")
	}
	if !strings.Contains(out, "Garage") {
		t.Error("expected amenity in output")
	}
	if !strings.Contains(out, "2 apartments found") {
		t.Error("expected count in output")
	}
}

func TestApartmentTable_Empty(t *testing.T) {
	var buf bytes.Buffer
	ApartmentTable(&buf, []models.Apartment{})
	if !strings.Contains(buf.String(), "0 apartments found") {
		t.Error("expected 0 count")
	}
}

func TestPriceHistoryTable(t *testing.T) {
	apt := &models.Apartment{UnitNumber: "2146", FloorPlan: "B2R", Bedrooms: 2, Bathrooms: 2, SqFt: 1142}
	records := []models.PriceRecord{
		{Price: 2195, ScrapedAt: time.Date(2026, 3, 15, 10, 0, 0, 0, time.UTC)},
		{Price: 2095, ScrapedAt: time.Date(2026, 3, 18, 10, 0, 0, 0, time.UTC)},
		{Price: 2095, ScrapedAt: time.Date(2026, 3, 22, 18, 0, 0, 0, time.UTC)},
	}

	var buf bytes.Buffer
	PriceHistoryTable(&buf, apt, records)
	out := buf.String()

	if !strings.Contains(out, "UNIT #2146") {
		t.Error("expected unit header")
	}
	if !strings.Contains(out, "-$100") {
		t.Error("expected -$100 change")
	}
	if !strings.Contains(out, "(no change)") {
		t.Error("expected (no change)")
	}
}

func TestStatsDisplay(t *testing.T) {
	stats := &models.Stats{
		FloorPlans: 14,
		Available:  24,
		ByBedrooms: []models.BedroomStats{
			{Bedrooms: 1, Count: 11, MinPrice: 1635, MaxPrice: 1975},
			{Bedrooms: 2, Count: 8, MinPrice: 2010, MaxPrice: 2220},
		},
		TotalScrapes: 5,
	}

	var buf bytes.Buffer
	StatsDisplay(&buf, stats)
	out := buf.String()

	if !strings.Contains(out, "Floor Plans:    14") {
		t.Error("expected floor plan count")
	}
	if !strings.Contains(out, "Available Now:  24") {
		t.Error("expected available count")
	}
	if !strings.Contains(out, "1 bed:  11 units") {
		t.Error("expected bedroom stats")
	}
	if !strings.Contains(out, "Last Scrape:    never") {
		t.Error("expected 'never' when LastScrape is nil")
	}
}
