package scraper

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/dotbrains/aptscout/internal/db"
	"github.com/dotbrains/aptscout/internal/models"
)

const maxRetries = 3

// Result holds the outcome of a scrape run.
type Result struct {
	Property     string
	FloorPlans   int
	UnitsFound   int
	UnitsNew     int
	UnitsRemoved int
	UnitsChanged int
}

// Scraper fetches and parses apartment data using a Provider.
type Scraper struct {
	client *http.Client
	db     *db.DB
	writer io.Writer
}

// New creates a new Scraper.
func New(database *db.DB, writer io.Writer) *Scraper {
	return &Scraper{
		client: &http.Client{Timeout: 30 * time.Second},
		db:     database,
		writer: writer,
	}
}

// RunProvider executes a scrape for a single provider.
func (s *Scraper) RunProvider(ctx context.Context, prov models.Provider) (*Result, error) {
	property := prov.ID()
	result := &Result{Property: property}

	_, _ = fmt.Fprintf(s.writer, "→ Scraping %s...\n", prov.Name())

	// Record scrape run.
	runID, err := s.db.InsertScrapeRun(property)
	if err != nil {
		return nil, fmt.Errorf("recording scrape run: %w", err)
	}

	data, scrapeErr := prov.Scrape(ctx, s.fetch)
	if data != nil {
		result.FloorPlans = len(data.FloorPlans)
		result.UnitsFound = len(data.Apartments)
	}

	// Persist results even if scrape had partial errors.
	if data != nil {
		if persistErr := s.persist(property, data, result); persistErr != nil && scrapeErr == nil {
			scrapeErr = persistErr
		}
	}

	// Complete the scrape run record.
	run := models.ScrapeRun{
		FloorPlans: result.FloorPlans, UnitsFound: result.UnitsFound,
		UnitsNew: result.UnitsNew, UnitsRemoved: result.UnitsRemoved, UnitsChanged: result.UnitsChanged,
	}
	if scrapeErr != nil {
		errStr := scrapeErr.Error()
		run.Error = &errStr
	}
	_ = s.db.CompleteScrapeRun(runID, run)

	_, _ = fmt.Fprintf(s.writer, "→ %s: %d plans, %d units\n", prov.Name(), result.FloorPlans, result.UnitsFound)
	return result, scrapeErr
}

func (s *Scraper) persist(property string, data *models.ScrapeData, result *Result) error {
	// Upsert floor plans.
	for _, fp := range data.FloorPlans {
		if err := s.db.UpsertFloorPlan(fp); err != nil {
			return fmt.Errorf("upserting floor plan %s: %w", fp.Code, err)
		}
	}

	// Upsert apartments.
	var activeUnits []string
	for _, apt := range data.Apartments {
		activeUnits = append(activeUnits, apt.UnitNumber)
		isNew, priceChanged, err := s.db.UpsertApartment(apt)
		if err != nil {
			return fmt.Errorf("upserting unit %s: %w", apt.UnitNumber, err)
		}
		if isNew || priceChanged {
			if err := s.db.InsertPriceHistory(property, apt.UnitNumber, apt.Price); err != nil {
				return fmt.Errorf("inserting price history for %s: %w", apt.UnitNumber, err)
			}
		}
		if isNew {
			result.UnitsNew++
		}
		if priceChanged {
			result.UnitsChanged++
		}
	}

	// Mark missing units unavailable (scoped to this property).
	removed, err := s.db.MarkUnavailable(property, activeUnits)
	if err != nil {
		return fmt.Errorf("marking unavailable: %w", err)
	}
	result.UnitsRemoved = int(removed)
	return nil
}

func (s *Scraper) fetch(ctx context.Context, url string) (string, error) {
	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			delay := time.Duration(1<<uint(attempt-1)) * time.Second
			time.Sleep(delay)
		}

		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			return "", err
		}
		req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36")
		req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
		req.Header.Set("Accept-Language", "en-US,en;q=0.9")
		req.Header.Set("Cache-Control", "no-cache")

		resp, err := s.client.Do(req)
		if err != nil {
			lastErr = err
			continue
		}

		body, err := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if err != nil {
			lastErr = err
			continue
		}

		if resp.StatusCode == 429 || resp.StatusCode == 403 {
			lastErr = fmt.Errorf("HTTP %d for %s", resp.StatusCode, url)
			continue
		}
		if resp.StatusCode != 200 {
			return "", fmt.Errorf("HTTP %d for %s", resp.StatusCode, url)
		}

		return string(body), nil
	}
	return "", fmt.Errorf("fetch %s failed after %d retries: %w", url, maxRetries, lastErr)
}
