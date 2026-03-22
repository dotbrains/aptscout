package db

import (
	"encoding/json"

	"github.com/dotbrains/aptscout/internal/models"
)

// UpsertFloorPlan inserts or updates a floor plan.
func (db *DB) UpsertFloorPlan(fp models.FloorPlan) error {
	features, _ := json.Marshal(fp.Features)
	_, err := db.Exec(`
		INSERT INTO floor_plans (property, code, bedrooms, bathrooms, sqft, deposit, is_renovated, features, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(property, code) DO UPDATE SET
			bedrooms = excluded.bedrooms,
			bathrooms = excluded.bathrooms,
			sqft = excluded.sqft,
			deposit = excluded.deposit,
			is_renovated = excluded.is_renovated,
			features = excluded.features,
			updated_at = excluded.updated_at
	`, fp.Property, fp.Code, fp.Bedrooms, fp.Bathrooms, fp.SqFt, fp.Deposit, fp.IsRenovated, string(features), now())
	return err
}

// GetFloorPlans returns floor plans, optionally filtered by property.
func (db *DB) GetFloorPlans(property *string) ([]models.FloorPlan, error) {
	query := `SELECT property, code, bedrooms, bathrooms, sqft, deposit, is_renovated, features, updated_at FROM floor_plans`
	var args []interface{}
	if property != nil {
		query += ` WHERE property = ?`
		args = append(args, *property)
	}
	query += ` ORDER BY property, code`

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var plans []models.FloorPlan
	for rows.Next() {
		var fp models.FloorPlan
		var features string
		if err := rows.Scan(&fp.Property, &fp.Code, &fp.Bedrooms, &fp.Bathrooms, &fp.SqFt, &fp.Deposit, &fp.IsRenovated, &features, &fp.UpdatedAt); err != nil {
			return nil, err
		}
		_ = json.Unmarshal([]byte(features), &fp.Features)
		if fp.Features == nil {
			fp.Features = []string{}
		}
		plans = append(plans, fp)
	}
	return plans, rows.Err()
}
