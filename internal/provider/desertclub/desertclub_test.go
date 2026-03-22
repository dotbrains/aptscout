package desertclub

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/dotbrains/aptscout/internal/models"
)

const fixtureIndex = `<html><body>
<h2>A1R</h2><div>1 Bed</div><div>1 Bath</div><div>715 Sq. Ft.</div><div>Deposit: $600</div>
<h2>B2R</h2><div>2 Bed</div><div>2 Bath</div><div>1,142 Sq. Ft.</div><div>Deposit: $700</div>
<h2>B2R</h2><!-- duplicate -->
</body></html>`

const fixtureDetail = `<html><body>
<h3>Apartment: # 2146</h3>
<div>Available Now</div>
<div>Starting at: $2,095</div>
<div>Garage Unattached, 2nd Floor</div>

<h3>Apartment: # 3169</h3>
<div>Date Available: 4/17/2026</div>
<div>Starting at: $2,010</div>
<div>3rd Floor</div>
</body></html>`

const fixtureEmpty = `<html><body>
<h2>C1R</h2><div>3 Bed</div><div>2 Bath</div>
<p>No apartments available</p>
</body></html>`

func TestParseFloorPlanIndex(t *testing.T) {
	plans, err := parseFloorPlanIndex(fixtureIndex)
	if err != nil {
		t.Fatal(err)
	}
	if len(plans) != 2 {
		t.Fatalf("expected 2 plans (deduped), got %d", len(plans))
	}
	if plans[0].Code != "A1R" || plans[1].Code != "B2R" {
		t.Errorf("unexpected codes: %s, %s", plans[0].Code, plans[1].Code)
	}
	if plans[1].SqFt != 1142 {
		t.Errorf("expected B2R sqft=1142, got %d", plans[1].SqFt)
	}
}

func TestParseFloorPlanIndex_Empty(t *testing.T) {
	plans, err := parseFloorPlanIndex("<html><body><h1>No plans</h1></body></html>")
	if err != nil {
		t.Fatal(err)
	}
	if len(plans) != 0 {
		t.Errorf("expected 0 plans, got %d", len(plans))
	}
}

func TestParseFloorPlanDetail(t *testing.T) {
	apts, err := parseFloorPlanDetail(fixtureDetail)
	if err != nil {
		t.Fatal(err)
	}
	if len(apts) != 2 {
		t.Fatalf("expected 2 apartments, got %d", len(apts))
	}
	if apts[0].UnitNumber != "2146" || apts[0].Price != 2095 || !apts[0].AvailableNow {
		t.Errorf("unexpected apt[0]: %+v", apts[0])
	}
	if apts[1].UnitNumber != "3169" || apts[1].Price != 2010 || apts[1].AvailableDate != "4/17/2026" {
		t.Errorf("unexpected apt[1]: %+v", apts[1])
	}
}

func TestParseFloorPlanDetail_Empty(t *testing.T) {
	apts, _ := parseFloorPlanDetail(fixtureEmpty)
	if len(apts) != 0 {
		t.Errorf("expected 0 apartments, got %d", len(apts))
	}
}

func TestToModel_BedBathFallback(t *testing.T) {
	p := parsedFloorPlan{Code: "B3R", SqFt: 1153}
	fp := p.toModel("test")
	if fp.Bedrooms != 2 || fp.Bathrooms != 2 {
		t.Errorf("expected 2/2 from B code fallback, got %d/%d", fp.Bedrooms, fp.Bathrooms)
	}
	if !fp.IsRenovated {
		t.Error("expected IsRenovated=true for R suffix")
	}
}

func TestToModel_DateConversion(t *testing.T) {
	pa := parsedApartment{UnitNumber: "100", Price: 2000, AvailableDate: "4/17/2026"}
	apt := pa.toModel("test", "B2R")
	if apt.AvailableDate == nil || *apt.AvailableDate != "2026-04-17" {
		t.Errorf("expected 2026-04-17, got %v", apt.AvailableDate)
	}
}

func TestProviderScrape(t *testing.T) {
	p := New()
	if p.ID() != "desert-club" || p.Name() != "Desert Club Apartments" {
		t.Errorf("unexpected ID/Name: %s/%s", p.ID(), p.Name())
	}

	// Mock fetcher that returns canned HTML
	fetch := func(ctx context.Context, url string) (string, error) {
		if strings.Contains(url, "/floorplans/") {
			return fixtureDetail, nil
		}
		if strings.HasSuffix(url, "/floorplans") {
			return fixtureIndex, nil
		}
		return "", fmt.Errorf("unexpected URL: %s", url)
	}

	data, err := p.Scrape(context.Background(), models.Fetcher(fetch))
	if err != nil {
		t.Fatal(err)
	}
	if len(data.FloorPlans) != 2 {
		t.Errorf("expected 2 floor plans, got %d", len(data.FloorPlans))
	}
	// Each plan fetches the detail page, which has 2 apartments
	if len(data.Apartments) != 4 {
		t.Errorf("expected 4 apartments (2 plans × 2 units), got %d", len(data.Apartments))
	}
	for _, apt := range data.Apartments {
		if apt.Property != "desert-club" {
			t.Errorf("expected property=desert-club, got %s", apt.Property)
		}
	}
}
