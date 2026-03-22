package hideaway

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/net/html"

	"github.com/dotbrains/aptscout/internal/models"
)

const baseURL = "https://www.hideawaynorthscottsdale.com"

var (
	priceRe = regexp.MustCompile(`\$\s*([\d,]+)`)
	sqftRe  = regexp.MustCompile(`([\d,]+(?:-[\d,]+)?)\s*Sq\.\s*Ft\.`)
	bedRe   = regexp.MustCompile(`(\d+)\s*Bedroom`)
	bathRe  = regexp.MustCompile(`(\d+)\s*Bath`)
)

// Provider scrapes Hideaway North Scottsdale apartments.
type Provider struct{}

// New returns a new Hideaway provider.
func New() *Provider { return &Provider{} }

func (p *Provider) ID() string   { return "hideaway" }
func (p *Provider) Name() string { return "Hideaway North Scottsdale" }

func (p *Provider) Scrape(ctx context.Context, fetch models.Fetcher) (*models.ScrapeData, error) {
	data := &models.ScrapeData{}

	body, err := fetch(ctx, baseURL+"/floorplans")
	if err != nil {
		return data, fmt.Errorf("fetching floor plans: %w", err)
	}

	plans, err := parseFloorPlans(body)
	if err != nil {
		return data, fmt.Errorf("parsing floor plans: %w", err)
	}

	for _, pp := range plans {
		fp := pp.toFloorPlan(p.ID())
		data.FloorPlans = append(data.FloorPlans, fp)

		// Each floor plan on Hideaway's index is also a "unit" since
		// individual unit numbers aren't exposed. Use the plan code as
		// a synthetic unit identifier with the starting price.
		if pp.Price > 0 {
			data.Apartments = append(data.Apartments, models.Apartment{
				Property:     p.ID(),
				UnitNumber:   pp.Code, // synthetic: plan code as unit
				FloorPlan:    pp.Code,
				Price:        pp.Price,
				AvailableNow: true,
				Amenities:    []string{},
				IsAvailable:  true,
			})
		}
	}

	return data, nil
}

type parsedPlan struct {
	Code      string // slugified name, e.g. "saguaro-super"
	Name      string // original name, e.g. "Saguaro - Super"
	Bedrooms  int
	Bathrooms int
	SqFt      int
	Price     int // 0 = "Call for details"
}

func (pp parsedPlan) toFloorPlan(property string) models.FloorPlan {
	isRenovated := strings.Contains(strings.ToLower(pp.Name), "super")
	return models.FloorPlan{
		Property:    property,
		Code:        pp.Code,
		Bedrooms:    pp.Bedrooms,
		Bathrooms:   pp.Bathrooms,
		SqFt:        pp.SqFt,
		Deposit:     0,
		IsRenovated: isRenovated,
		Features:    []string{},
	}
}

func parseFloorPlans(body string) ([]parsedPlan, error) {
	doc, err := html.Parse(strings.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("parsing HTML: %w", err)
	}

	var plans []parsedPlan
	var current *parsedPlan

	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "h2" {
			text := strings.TrimSpace(extractText(n))
			// Skip non-plan headings.
			if text != "" && text != "Filters" && text != "No matches" &&
				text != "More Filters" && !strings.Contains(text, "More Filters") &&
				!strings.Contains(text, "Apartment") && len(text) > 3 {
				if current != nil {
					plans = append(plans, *current)
				}
				current = &parsedPlan{
					Name: text,
					Code: slugify(text),
				}
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
				// Handle ranges like "1,170-1,187"
				s := m[1]
				if idx := strings.Index(s, "-"); idx >= 0 {
					s = s[idx+1:] // take the max
				}
				current.SqFt = parseNumber(s)
			}
			if strings.Contains(text, "/mo") || strings.Contains(text, "$") {
				if m := priceRe.FindStringSubmatch(text); m != nil && current.Price == 0 {
					current.Price = parseNumber(m[1])
				}
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

	// Deduplicate by code.
	seen := make(map[string]bool)
	var deduped []parsedPlan
	for _, p := range plans {
		if !seen[p.Code] {
			seen[p.Code] = true
			deduped = append(deduped, p)
		}
	}
	return deduped, nil
}

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

func slugify(s string) string {
	s = strings.ToLower(s)
	s = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			return r
		}
		return '-'
	}, s)
	// Collapse multiple dashes.
	for strings.Contains(s, "--") {
		s = strings.ReplaceAll(s, "--", "-")
	}
	return strings.Trim(s, "-")
}
