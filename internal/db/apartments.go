package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/dotbrains/aptscout/internal/models"
)

// UpsertApartment inserts or updates an apartment. Returns whether the unit is new and whether the price changed.
func (db *DB) UpsertApartment(apt models.Apartment) (isNew bool, priceChanged bool, err error) {
	amenities, _ := json.Marshal(apt.Amenities)
	timestamp := now()

	// Check if it already exists.
	var existingPrice sql.NullInt64
	var existingAvailable bool
	err = db.QueryRow(`SELECT price, is_available FROM apartments WHERE property = ? AND unit_number = ?`, apt.Property, apt.UnitNumber).
		Scan(&existingPrice, &existingAvailable)

	if err == sql.ErrNoRows {
		// New unit.
		_, err = db.Exec(`
			INSERT INTO apartments (property, unit_number, floor_plan, price, available_date, available_now, floor, amenities, is_available, first_seen, last_seen)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, 1, ?, ?)
		`, apt.Property, apt.UnitNumber, apt.FloorPlan, apt.Price, apt.AvailableDate, apt.AvailableNow, apt.Floor, string(amenities), timestamp, timestamp)
		return true, false, err
	}
	if err != nil {
		return false, false, err
	}

	// Existing unit — update it.
	priceChanged = existingPrice.Valid && int(existingPrice.Int64) != apt.Price
	_, err = db.Exec(`
		UPDATE apartments SET
			price = ?, available_date = ?, available_now = ?, floor = ?, amenities = ?,
			is_available = 1, last_seen = ?
		WHERE property = ? AND unit_number = ?
	`, apt.Price, apt.AvailableDate, apt.AvailableNow, apt.Floor, string(amenities), timestamp, apt.Property, apt.UnitNumber)
	return false, priceChanged, err
}

// MarkUnavailable marks all units for a property NOT in the given set as unavailable.
func (db *DB) MarkUnavailable(property string, activeUnits []string) (int64, error) {
	if len(activeUnits) == 0 {
		res, err := db.Exec(`UPDATE apartments SET is_available = 0, last_seen = ? WHERE property = ? AND is_available = 1`, now(), property)
		if err != nil {
			return 0, err
		}
		return res.RowsAffected()
	}

	placeholders := make([]string, len(activeUnits))
	args := make([]interface{}, 0, len(activeUnits)+2)
	args = append(args, now(), property)
	for i, u := range activeUnits {
		placeholders[i] = "?"
		args = append(args, u)
	}

	query := fmt.Sprintf(
		`UPDATE apartments SET is_available = 0, last_seen = ? WHERE property = ? AND is_available = 1 AND unit_number NOT IN (%s)`,
		strings.Join(placeholders, ","),
	)
	res, err := db.Exec(query, args...)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// ListApartments returns apartments matching the given filter.
func (db *DB) ListApartments(f models.ApartmentFilter) ([]models.Apartment, error) {
	query := `
		SELECT a.property, a.unit_number, a.floor_plan, fp.bedrooms, fp.bathrooms, fp.sqft,
		       a.price, a.available_date, a.available_now, a.floor, a.amenities,
		       fp.is_renovated, fp.deposit, a.is_available, a.first_seen, a.last_seen
		FROM apartments a
		JOIN floor_plans fp ON a.property = fp.property AND a.floor_plan = fp.code
		WHERE a.is_available = 1
	`
	var args []interface{}

	if f.Property != nil {
		query += " AND a.property = ?"
		args = append(args, *f.Property)
	}
	if f.Beds != nil {
		query += " AND fp.bedrooms = ?"
		args = append(args, *f.Beds)
	}
	if f.Baths != nil {
		query += " AND fp.bathrooms = ?"
		args = append(args, *f.Baths)
	}
	if f.MinPrice != nil {
		query += " AND a.price >= ?"
		args = append(args, *f.MinPrice)
	}
	if f.MaxPrice != nil {
		query += " AND a.price <= ?"
		args = append(args, *f.MaxPrice)
	}
	if f.Plan != nil {
		query += " AND UPPER(a.floor_plan) = UPPER(?)"
		args = append(args, *f.Plan)
	}
	if f.Renovated != nil && *f.Renovated {
		query += " AND fp.is_renovated = 1"
	}
	if f.AvailableBy != nil {
		query += " AND (a.available_now = 1 OR a.available_date <= ?)"
		args = append(args, *f.AvailableBy)
	}

	// Sort
	sortCol := "a.price"
	switch f.Sort {
	case "date":
		sortCol = "COALESCE(a.available_date, '0000-00-00')"
	case "sqft":
		sortCol = "fp.sqft"
	case "unit":
		sortCol = "a.unit_number"
	case "price":
		sortCol = "a.price"
	}
	order := "ASC"
	if f.Order == "desc" {
		order = "DESC"
	}
	query += fmt.Sprintf(" ORDER BY %s %s", sortCol, order)

	return db.scanApartments(query, args...)
}

// GetApartment returns a single apartment by property + unit number.
func (db *DB) GetApartment(property, unitNumber string) (*models.Apartment, error) {
	query := `
		SELECT a.property, a.unit_number, a.floor_plan, fp.bedrooms, fp.bathrooms, fp.sqft,
		       a.price, a.available_date, a.available_now, a.floor, a.amenities,
		       fp.is_renovated, fp.deposit, a.is_available, a.first_seen, a.last_seen
		FROM apartments a
		JOIN floor_plans fp ON a.property = fp.property AND a.floor_plan = fp.code
		WHERE a.property = ? AND a.unit_number = ?
	`
	apts, err := db.scanApartments(query, property, unitNumber)
	if err != nil {
		return nil, err
	}
	if len(apts) == 0 {
		return nil, sql.ErrNoRows
	}
	return &apts[0], nil
}

func (db *DB) scanApartments(query string, args ...interface{}) ([]models.Apartment, error) {
	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var apts []models.Apartment
	for rows.Next() {
		var apt models.Apartment
		var amenities string
		var availDate sql.NullString
		if err := rows.Scan(
			&apt.Property, &apt.UnitNumber, &apt.FloorPlan, &apt.Bedrooms, &apt.Bathrooms, &apt.SqFt,
			&apt.Price, &availDate, &apt.AvailableNow, &apt.Floor, &amenities,
			&apt.IsRenovated, &apt.Deposit, &apt.IsAvailable, &apt.FirstSeen, &apt.LastSeen,
		); err != nil {
			return nil, err
		}
		if availDate.Valid {
			apt.AvailableDate = &availDate.String
		}
		_ = json.Unmarshal([]byte(amenities), &apt.Amenities)
		if apt.Amenities == nil {
			apt.Amenities = []string{}
		}
		apts = append(apts, apt)
	}
	return apts, rows.Err()
}
