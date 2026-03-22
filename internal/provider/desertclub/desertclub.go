package desertclub

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/html"

	"github.com/dotbrains/aptscout/internal/models"
)

const (
	baseURL        = "https://arizona.weidner.com/apartments/az/phoenix/desert-club0"
	maxConcurrency = 2
	requestDelay   = 1 * time.Second
)

var (
	floorPlanCodeRe = regexp.MustCompile(`^[A-Z]\d[RP]$`)
	unitNumberRe    = regexp.MustCompile(`#\s*(\d+)`)
	priceRe         = regexp.MustCompile(`\$\s*([\d,]+)`)
	dateAvailRe     = regexp.MustCompile(`Date Available:\s*(\d{1,2}/\d{1,2}/\d{4})`)
	floorRe         = regexp.MustCompile(`(\d+)(?:st|nd|rd|th)\s+Floor`)
	sqftRe          = regexp.MustCompile(`([\d,]+)\s*Sq\.\s*Ft\.`)
	depositRe       = regexp.MustCompile(`Deposit:\s*\$\s*([\d,]+)`)
	bedRe           = regexp.MustCompile(`(\d+)\s*Bed`)
	bathRe          = regexp.MustCompile(`(\d+)\s*Bath`)
)

// Provider scrapes Desert Club Apartments (Weidner).
type Provider struct{}

// New returns a new Desert Club provider.
func New() *Provider { return &Provider{} }

func (p *Provider) ID() string   { return "desert-club" }
func (p *Provider) Name() string { return "Desert Club Apartments" }

func (p *Provider) Scrape(ctx context.Context, fetch models.Fetcher) (*models.ScrapeData, error) {
	data := &models.ScrapeData{}

	// Step 1: Fetch and parse the floor plans index.
	indexBody, err := fetch(ctx, baseURL+"/floorplans")
	if err != nil {
		return data, fmt.Errorf("fetching floor plans index: %w", err)
	}

	parsedPlans, err := parseFloorPlanIndex(indexBody)
	if err != nil {
		return data, fmt.Errorf("parsing floor plans index: %w", err)
	}

	for _, pp := range parsedPlans {
		data.FloorPlans = append(data.FloorPlans, pp.toModel(p.ID()))
	}

	// Step 2: Fetch each floor plan detail page concurrently.
	type planResult struct {
		code string
		apts []parsedApartment
		err  error
	}
	results := make(chan planResult, len(parsedPlans))
	sem := make(chan struct{}, maxConcurrency)
	var wg sync.WaitGroup

	for _, pp := range parsedPlans {
		wg.Add(1)
		go func(code string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			time.Sleep(requestDelay)

			url := fmt.Sprintf("%s/floorplans/%s", baseURL, strings.ToLower(code))
			body, err := fetch(ctx, url)
			if err != nil {
				results <- planResult{code: code, err: err}
				return
			}
			apts, err := parseFloorPlanDetail(body)
			results <- planResult{code: code, apts: apts, err: err}
		}(pp.Code)
	}
	go func() { wg.Wait(); close(results) }()

	var errors []string
	for r := range results {
		if r.err != nil {
			errors = append(errors, fmt.Sprintf("%s: %v", r.code, r.err))
			continue
		}
		for _, pa := range r.apts {
			data.Apartments = append(data.Apartments, pa.toModel(p.ID(), r.code))
		}
	}

	if len(errors) > 0 {
		return data, fmt.Errorf("scrape errors: %s", strings.Join(errors, "; "))
	}
	return data, nil
}

// --- internal parsed types ---

type parsedFloorPlan struct {
	Code      string
	Bedrooms  int
	Bathrooms int
	SqFt      int
	Deposit   int
}

func (pp parsedFloorPlan) toModel(property string) models.FloorPlan {
	beds, baths := pp.Bedrooms, pp.Bathrooms
	if beds == 0 && len(pp.Code) >= 2 {
		switch pp.Code[0] {
		case 'A', 'S':
			beds, baths = 1, 1
		case 'B':
			beds, baths = 2, 2
		case 'C':
			beds, baths = 3, 2
		}
	}
	return models.FloorPlan{
		Property:    property,
		Code:        pp.Code,
		Bedrooms:    beds,
		Bathrooms:   baths,
		SqFt:        pp.SqFt,
		Deposit:     pp.Deposit,
		IsRenovated: strings.HasSuffix(pp.Code, "R"),
		Features:    []string{},
	}
}

type parsedApartment struct {
	UnitNumber    string
	Price         int
	AvailableNow  bool
	AvailableDate string
	Floor         int
	Amenities     []string
}

func (pa parsedApartment) toModel(property, floorPlanCode string) models.Apartment {
	apt := models.Apartment{
		Property:     property,
		UnitNumber:   pa.UnitNumber,
		FloorPlan:    floorPlanCode,
		Price:        pa.Price,
		AvailableNow: pa.AvailableNow,
		Floor:        pa.Floor,
		Amenities:    pa.Amenities,
		IsAvailable:  true,
	}
	if pa.AvailableDate != "" {
		parts := strings.Split(pa.AvailableDate, "/")
		if len(parts) == 3 {
			iso := fmt.Sprintf("%s-%02s-%02s", parts[2], zeroPad(parts[0]), zeroPad(parts[1]))
			apt.AvailableDate = &iso
		}
	}
	if apt.Amenities == nil {
		apt.Amenities = []string{}
	}
	return apt
}

// --- HTML parsers ---

func parseFloorPlanIndex(body string) ([]parsedFloorPlan, error) {
	doc, err := html.Parse(strings.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("parsing HTML: %w", err)
	}

	var plans []parsedFloorPlan
	var current *parsedFloorPlan

	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "h2" {
			text := strings.TrimSpace(extractText(n))
			if floorPlanCodeRe.MatchString(strings.ToUpper(text)) {
				if current != nil {
					plans = append(plans, *current)
				}
				current = &parsedFloorPlan{Code: strings.ToUpper(text)}
			}
		}
		if current != nil && n.Type == html.TextNode {
			text := n.Data
			if m := bedRe.FindStringSubmatch(text); m != nil {
				current.Bedrooms, _ = strconv.Atoi(m[1])
			}
			if m := bathRe.FindStringSubmatch(text); m != nil {
				current.Bathrooms, _ = strconv.Atoi(m[1])
			}
			if m := sqftRe.FindStringSubmatch(text); m != nil {
				current.SqFt = parseNumber(m[1])
			}
			if m := depositRe.FindStringSubmatch(text); m != nil {
				current.Deposit = parseNumber(m[1])
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(doc)
	if current != nil {
		plans = append(plans, *current)
	}

	// Deduplicate.
	seen := make(map[string]bool)
	var deduped []parsedFloorPlan
	for _, p := range plans {
		if !seen[p.Code] {
			seen[p.Code] = true
			deduped = append(deduped, p)
		}
	}
	return deduped, nil
}

func parseFloorPlanDetail(body string) ([]parsedApartment, error) {
	doc, err := html.Parse(strings.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("parsing HTML: %w", err)
	}

	var apartments []parsedApartment
	var current *parsedApartment

	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "h3" {
			text := extractText(n)
			if m := unitNumberRe.FindStringSubmatch(text); m != nil {
				if current != nil {
					apartments = append(apartments, *current)
				}
				current = &parsedApartment{UnitNumber: m[1]}
			}
		}
		if current != nil && n.Type == html.TextNode {
			text := strings.TrimSpace(n.Data)
			if strings.Contains(text, "Available Now") {
				current.AvailableNow = true
			}
			if m := dateAvailRe.FindStringSubmatch(text); m != nil {
				current.AvailableDate = m[1]
			}
			if (strings.Contains(text, "Starting at") || strings.Contains(text, "$")) && current.Price == 0 {
				if m := priceRe.FindStringSubmatch(text); m != nil {
					current.Price = parseNumber(m[1])
				}
			}
			if m := floorRe.FindStringSubmatch(text); m != nil {
				current.Floor, _ = strconv.Atoi(m[1])
			}
			// Amenities.
			if len(text) > 2 && len(text) < 200 &&
				!strings.Contains(text, "function") && !strings.Contains(text, "{") &&
				!strings.Contains(text, "(") && !strings.Contains(text, "$") &&
				!strings.Contains(text, "Apartment") && !strings.Contains(text, "Starting") &&
				!strings.Contains(text, "Apply") && !strings.Contains(text, "Available") &&
				!strings.Contains(text, "Floor Plan") && !strings.Contains(text, "question") {
				if strings.Contains(text, "Floor") || strings.Contains(text, "Garage") ||
					strings.Contains(text, "Wheelchair") {
					for _, p := range strings.Split(text, ",") {
						p = strings.TrimSpace(p)
						if p != "" && len(p) < 50 && !floorRe.MatchString(p) &&
							!strings.Contains(p, "Floor Plan") &&
							p != "Floorplan" && p != "Plank Flooring" &&
							!strings.HasPrefix(p, "Platinum") {
							current.Amenities = append(current.Amenities, p)
						}
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(doc)
	if current != nil {
		apartments = append(apartments, *current)
	}
	return apartments, nil
}

// --- helpers ---

func extractText(n *html.Node) string {
	var sb strings.Builder
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.TextNode {
			sb.WriteString(n.Data)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(n)
	return sb.String()
}

func parseNumber(s string) int {
	s = strings.ReplaceAll(s, ",", "")
	n, _ := strconv.Atoi(s)
	return n
}

func zeroPad(s string) string {
	if len(s) == 1 {
		return "0" + s
	}
	return s
}
