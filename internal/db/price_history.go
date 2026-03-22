package db

import (
	"github.com/dotbrains/aptscout/internal/models"
)

// InsertPriceHistory records a price point for a unit.
func (db *DB) InsertPriceHistory(property, unitNumber string, price int) error {
	_, err := db.Exec(
		`INSERT INTO price_history (property, unit_number, price, scraped_at) VALUES (?, ?, ?, ?)`,
		property, unitNumber, price, now(),
	)
	return err
}

// GetPriceHistory returns price history for a unit, ordered by time.
func (db *DB) GetPriceHistory(property, unitNumber string) ([]models.PriceRecord, error) {
	rows, err := db.Query(
		`SELECT id, unit_number, price, scraped_at FROM price_history WHERE property = ? AND unit_number = ? ORDER BY scraped_at ASC`,
		property, unitNumber,
	)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var records []models.PriceRecord
	for rows.Next() {
		var r models.PriceRecord
		if err := rows.Scan(&r.ID, &r.UnitNumber, &r.Price, &r.ScrapedAt); err != nil {
			return nil, err
		}
		records = append(records, r)
	}
	return records, rows.Err()
}
