package models

import (
	"context"
	"time"
)

// Fetcher fetches a URL and returns the response body as a string.
type Fetcher func(ctx context.Context, url string) (string, error)

// Provider defines how to scrape a specific apartment complex.
type Provider interface {
	ID() string
	Name() string
	Scrape(ctx context.Context, fetch Fetcher) (*ScrapeData, error)
}

// ScrapeData holds the standardized output from a provider scrape.
type ScrapeData struct {
	FloorPlans []FloorPlan
	Apartments []Apartment
}

// FloorPlan represents a floor plan configuration.
type FloorPlan struct {
	Property    string   `json:"property"`
	Code        string   `json:"code"`
	Bedrooms    int      `json:"bedrooms"`
	Bathrooms   int      `json:"bathrooms"`
	SqFt        int      `json:"sqft"`
	Deposit     int      `json:"deposit"`
	IsRenovated bool     `json:"is_renovated"`
	Features    []string `json:"features"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Apartment represents a single available unit.
type Apartment struct {
	Property      string    `json:"property"`
	UnitNumber    string    `json:"unit_number"`
	FloorPlan     string    `json:"floor_plan"`
	Bedrooms      int       `json:"bedrooms"`
	Bathrooms     int       `json:"bathrooms"`
	SqFt          int       `json:"sqft"`
	Price         int       `json:"price"`
	AvailableDate *string   `json:"available_date"`
	AvailableNow  bool      `json:"available_now"`
	Floor         int       `json:"floor"`
	Amenities     []string  `json:"amenities"`
	IsRenovated   bool      `json:"is_renovated"`
	Deposit       int       `json:"deposit"`
	IsAvailable   bool      `json:"is_available"`
	FirstSeen     time.Time `json:"first_seen"`
	LastSeen      time.Time `json:"last_seen"`
}

// PriceRecord represents a historical price point for a unit.
type PriceRecord struct {
	ID         int       `json:"id"`
	UnitNumber string    `json:"unit_number"`
	Price      int       `json:"price"`
	ScrapedAt  time.Time `json:"scraped_at"`
}

// ScrapeRun represents a single scrape execution.
type ScrapeRun struct {
	ID           int        `json:"id"`
	Property     string     `json:"property"`
	StartedAt    time.Time  `json:"started_at"`
	CompletedAt  *time.Time `json:"completed_at"`
	FloorPlans   int        `json:"floor_plans"`
	UnitsFound   int        `json:"units_found"`
	UnitsNew     int        `json:"units_new"`
	UnitsRemoved int        `json:"units_removed"`
	UnitsChanged int        `json:"units_changed"`
	Error        *string    `json:"error"`
}

// ApartmentFilter holds query parameters for filtering apartments.
type ApartmentFilter struct {
	Property    *string `json:"property"`
	Beds        *int    `json:"beds"`
	Baths       *int    `json:"baths"`
	MinPrice    *int    `json:"min_price"`
	MaxPrice    *int    `json:"max_price"`
	Plan        *string `json:"plan"`
	Renovated   *bool   `json:"renovated"`
	AvailableBy *string `json:"available_by"`
	Sort        string  `json:"sort"`
	Order       string  `json:"order"`
}

// ApartmentDetail is an Apartment with its price history attached.
type ApartmentDetail struct {
	Apartment    Apartment     `json:"apartment"`
	PriceHistory []PriceRecord `json:"price_history"`
}

// Stats holds summary statistics.
type Stats struct {
	FloorPlans    int            `json:"floor_plans"`
	Available     int            `json:"available"`
	ByBedrooms    []BedroomStats `json:"by_bedrooms"`
	LastScrape    *time.Time     `json:"last_scrape"`
	TotalScrapes  int            `json:"total_scrapes"`
}

// BedroomStats holds stats for a specific bedroom count.
type BedroomStats struct {
	Bedrooms int `json:"bedrooms"`
	Count    int `json:"count"`
	MinPrice int `json:"min_price"`
	MaxPrice int `json:"max_price"`
}
