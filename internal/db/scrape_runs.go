package db

import (
	"database/sql"
	"time"

	"github.com/dotbrains/aptscout/internal/models"
)

// InsertScrapeRun starts a new scrape run and returns its ID.
func (db *DB) InsertScrapeRun(property string) (int64, error) {
	res, err := db.Exec(`INSERT INTO scrape_runs (property, started_at) VALUES (?, ?)`, property, now())
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// CompleteScrapeRun updates a scrape run with results.
func (db *DB) CompleteScrapeRun(id int64, run models.ScrapeRun) error {
	_, err := db.Exec(`
		UPDATE scrape_runs SET
			completed_at = ?, floor_plans = ?, units_found = ?,
			units_new = ?, units_removed = ?, units_changed = ?, error = ?
		WHERE id = ?
	`, now(), run.FloorPlans, run.UnitsFound, run.UnitsNew, run.UnitsRemoved, run.UnitsChanged, run.Error, id)
	return err
}

// GetScrapeRuns returns all scrape runs, most recent first.
func (db *DB) GetScrapeRuns() ([]models.ScrapeRun, error) {
	rows, err := db.Query(`
		SELECT id, property, started_at, completed_at, floor_plans, units_found, units_new, units_removed, units_changed, error
		FROM scrape_runs ORDER BY started_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var runs []models.ScrapeRun
	for rows.Next() {
		var r models.ScrapeRun
		var completedAt sql.NullTime
		var errStr sql.NullString
		if err := rows.Scan(&r.ID, &r.Property, &r.StartedAt, &completedAt, &r.FloorPlans, &r.UnitsFound, &r.UnitsNew, &r.UnitsRemoved, &r.UnitsChanged, &errStr); err != nil {
			return nil, err
		}
		if completedAt.Valid {
			r.CompletedAt = &completedAt.Time
		}
		if errStr.Valid {
			r.Error = &errStr.String
		}
		runs = append(runs, r)
	}
	return runs, rows.Err()
}

// GetLastScrapeTime returns the time of the most recent completed scrape.
func (db *DB) GetLastScrapeTime() (*time.Time, error) {
	var s sql.NullString
	err := db.QueryRow(`SELECT MAX(completed_at) FROM scrape_runs WHERE completed_at IS NOT NULL`).Scan(&s)
	if err != nil || !s.Valid {
		return nil, err
	}
	t, err := time.Parse(time.RFC3339, s.String)
	if err != nil {
		return nil, nil
	}
	return &t, nil
}

// GetStats returns summary statistics, optionally filtered by property.
func (db *DB) GetStats(property *string) (*models.Stats, error) {
	stats := &models.Stats{}

	// Floor plan count
	if property != nil {
		_ = db.QueryRow(`SELECT COUNT(*) FROM floor_plans WHERE property = ?`, *property).Scan(&stats.FloorPlans)
		_ = db.QueryRow(`SELECT COUNT(*) FROM apartments WHERE is_available = 1 AND property = ?`, *property).Scan(&stats.Available)
	} else {
		_ = db.QueryRow(`SELECT COUNT(*) FROM floor_plans`).Scan(&stats.FloorPlans)
		_ = db.QueryRow(`SELECT COUNT(*) FROM apartments WHERE is_available = 1`).Scan(&stats.Available)
	}

	// By bedrooms
	bedroomQuery := `
		SELECT fp.bedrooms, COUNT(*) as cnt, MIN(a.price), MAX(a.price)
		FROM apartments a
		JOIN floor_plans fp ON a.property = fp.property AND a.floor_plan = fp.code
		WHERE a.is_available = 1 AND a.price > 0
	`
	var bedroomArgs []interface{}
	if property != nil {
		bedroomQuery += ` AND a.property = ?`
		bedroomArgs = append(bedroomArgs, *property)
	}
	bedroomQuery += ` GROUP BY fp.bedrooms ORDER BY fp.bedrooms`
	rows, err := db.Query(bedroomQuery, bedroomArgs...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	for rows.Next() {
		var bs models.BedroomStats
		if err := rows.Scan(&bs.Bedrooms, &bs.Count, &bs.MinPrice, &bs.MaxPrice); err != nil {
			return nil, err
		}
		stats.ByBedrooms = append(stats.ByBedrooms, bs)
	}

	// Last scrape
	stats.LastScrape, _ = db.GetLastScrapeTime()

	// Total scrapes
	_ = db.QueryRow(`SELECT COUNT(*) FROM scrape_runs`).Scan(&stats.TotalScrapes)

	return stats, nil
}

// CleanStale removes apartments not seen in the given number of days.
func (db *DB) CleanStale(days int) (int64, error) {
	cutoff := time.Now().UTC().AddDate(0, 0, -days).Format(time.RFC3339)
	res, err := db.Exec(`DELETE FROM apartments WHERE is_available = 0 AND last_seen < ?`, cutoff)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}
