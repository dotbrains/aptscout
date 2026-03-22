package hideaway

import (
	"context"
	"testing"

	"github.com/dotbrains/aptscout/internal/models"
)

const fixture = `<html><body>
<h2>Saguaro - Super (1 bed + flex space)</h2>
<div>1 Bedroom / 2 Bath ● 1,175 Sq. Ft.</div>
<div>$2,480 /mo</div>

<h2>Willow - Super</h2>
<div>2 Bedroom / 2 Bath ● 1,170 Sq. Ft.</div>
<div>$2,123 /mo</div>

<h2>Mesquite - Super (2 bed + flex space)</h2>
<div>2 Bedroom / 2 Bath ● 1,364 Sq. Ft.</div>
<div>Call for details</div>

<h2>Filters</h2>
<h2>No matches</h2>
</body></html>`

func TestParseFloorPlans(t *testing.T) {
	plans, err := parseFloorPlans(fixture)
	if err != nil {
		t.Fatal(err)
	}
	if len(plans) != 3 {
		t.Fatalf("expected 3 plans, got %d", len(plans))
	}
	if plans[0].Code != "saguaro-super-1-bed-flex-space" {
		t.Errorf("unexpected code: %s", plans[0].Code)
	}
	if plans[0].Bedrooms != 1 || plans[0].Bathrooms != 2 || plans[0].SqFt != 1175 {
		t.Errorf("unexpected saguaro specs: %+v", plans[0])
	}
	if plans[0].Price != 2480 {
		t.Errorf("expected saguaro price=2480, got %d", plans[0].Price)
	}
	if plans[1].Code != "willow-super" || plans[1].Bedrooms != 2 {
		t.Errorf("unexpected willow: %+v", plans[1])
	}
	// Mesquite has "Call for details" — price should be 0
	if plans[2].Price != 0 {
		t.Errorf("expected mesquite price=0 (call for details), got %d", plans[2].Price)
	}
}

func TestParseFloorPlans_SkipsNonPlan(t *testing.T) {
	plans, _ := parseFloorPlans(fixture)
	for _, p := range plans {
		if p.Code == "filters" || p.Code == "no-matches" {
			t.Errorf("should skip non-plan heading: %s", p.Code)
		}
	}
}

func TestSlugify(t *testing.T) {
	tests := []struct{ in, want string }{
		{"Saguaro - Super", "saguaro-super"},
		{"Mesquite - Deluxe (2 bed + flex space)", "mesquite-deluxe-2-bed-flex-space"},
		{"Simple", "simple"},
	}
	for _, tt := range tests {
		got := slugify(tt.in)
		if got != tt.want {
			t.Errorf("slugify(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestProviderScrape(t *testing.T) {
	p := New()
	if p.ID() != "hideaway" || p.Name() != "Hideaway North Scottsdale" {
		t.Errorf("unexpected ID/Name: %s/%s", p.ID(), p.Name())
	}

	fetch := func(ctx context.Context, url string) (string, error) {
		return fixture, nil
	}

	data, err := p.Scrape(context.Background(), models.Fetcher(fetch))
	if err != nil {
		t.Fatal(err)
	}
	if len(data.FloorPlans) != 3 {
		t.Errorf("expected 3 floor plans, got %d", len(data.FloorPlans))
	}
	// Only plans with price > 0 become apartments
	if len(data.Apartments) != 2 {
		t.Errorf("expected 2 apartments (mesquite has no price), got %d", len(data.Apartments))
	}
	for _, apt := range data.Apartments {
		if apt.Property != "hideaway" {
			t.Errorf("expected property=hideaway, got %s", apt.Property)
		}
	}
}

func TestIsRenovated(t *testing.T) {
	super := parsedPlan{Name: "Willow - Super"}
	deluxe := parsedPlan{Name: "Willow - Deluxe"}

	fpSuper := super.toFloorPlan("test")
	fpDeluxe := deluxe.toFloorPlan("test")

	if !fpSuper.IsRenovated {
		t.Error("expected Super to be renovated")
	}
	if fpDeluxe.IsRenovated {
		t.Error("expected Deluxe to NOT be renovated")
	}
}
