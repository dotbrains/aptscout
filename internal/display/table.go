//nolint:errcheck // CLI output writes — errors are not actionable
package display

import (
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	"github.com/dotbrains/aptscout/internal/models"
)

// ApartmentTable prints apartments in a formatted table.
func ApartmentTable(w io.Writer, apts []models.Apartment) {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintf(tw, "UNIT\tPLAN\tBEDS\tBATHS\tSQFT\tPRICE\tAVAILABLE\tFLOOR\tAMENITIES\n")
	for _, a := range apts {
		avail := "Now"
		if a.AvailableDate != nil {
			avail = *a.AvailableDate
		}
		floor := ordinal(a.Floor)
		amenities := "—"
		if len(a.Amenities) > 0 {
			amenities = strings.Join(a.Amenities, ", ")
		}
		fmt.Fprintf(tw, "#%s\t%s\t%d\t%d\t%s\t$%s\t%s\t%s\t%s\n",
			a.UnitNumber, a.FloorPlan, a.Bedrooms, a.Bathrooms,
			formatNumber(a.SqFt), formatNumber(a.Price),
			avail, floor, amenities,
		)
	}
	tw.Flush()
	fmt.Fprintf(w, "\n%d apartment%s found.\n", len(apts), plural(len(apts)))
}

// PriceHistoryTable prints price history for a unit.
func PriceHistoryTable(w io.Writer, apt *models.Apartment, records []models.PriceRecord) {
	fmt.Fprintf(w, "UNIT #%s — %s (%d bed / %d bath, %s sq ft)\n\n",
		apt.UnitNumber, apt.FloorPlan, apt.Bedrooms, apt.Bathrooms, formatNumber(apt.SqFt))

	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintf(tw, "DATE\tPRICE\tCHANGE\n")
	var prevPrice int
	for _, r := range records {
		change := "—"
		if prevPrice > 0 {
			diff := r.Price - prevPrice
			if diff == 0 {
				change = "(no change)"
			} else if diff > 0 {
				change = fmt.Sprintf("+$%s", formatNumber(diff))
			} else {
				change = fmt.Sprintf("-$%s", formatNumber(-diff))
			}
		}
		fmt.Fprintf(tw, "%s\t$%s\t%s\n",
			r.ScrapedAt.Format("2006-01-02 15:04"),
			formatNumber(r.Price), change,
		)
		prevPrice = r.Price
	}
	tw.Flush()
}

// StatsDisplay prints summary statistics.
func StatsDisplay(w io.Writer, stats *models.Stats) {
	fmt.Fprintf(w, "Desert Club Apartments — 6901 E Chauncey Lane, Phoenix, AZ 85054\n\n")
	fmt.Fprintf(w, "Floor Plans:    %d\n", stats.FloorPlans)
	fmt.Fprintf(w, "Available Now:  %d units\n\n", stats.Available)

	if len(stats.ByBedrooms) > 0 {
		fmt.Fprintf(w, "By Bedrooms:\n")
		for _, bs := range stats.ByBedrooms {
			fmt.Fprintf(w, "  %d bed:  %d unit%s ($%s – $%s)\n",
				bs.Bedrooms, bs.Count, plural(bs.Count),
				formatNumber(bs.MinPrice), formatNumber(bs.MaxPrice))
		}
		fmt.Fprintln(w)
	}

	if stats.LastScrape != nil {
		fmt.Fprintf(w, "Last Scrape:    %s\n", stats.LastScrape.Format("2006-01-02 15:04:05"))
	} else {
		fmt.Fprintf(w, "Last Scrape:    never\n")
	}
	fmt.Fprintf(w, "Total Scrapes:  %d\n", stats.TotalScrapes)
}

func ordinal(n int) string {
	if n == 0 {
		return "—"
	}
	suffix := "th"
	switch n % 10 {
	case 1:
		if n%100 != 11 {
			suffix = "st"
		}
	case 2:
		if n%100 != 12 {
			suffix = "nd"
		}
	case 3:
		if n%100 != 13 {
			suffix = "rd"
		}
	}
	return fmt.Sprintf("%d%s", n, suffix)
}

func formatNumber(n int) string {
	s := fmt.Sprintf("%d", n)
	if len(s) <= 3 {
		return s
	}
	var result []byte
	for i, c := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			result = append(result, ',')
		}
		result = append(result, byte(c))
	}
	return string(result)
}

func plural(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}
