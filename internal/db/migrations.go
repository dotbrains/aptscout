package db

const schema = `
CREATE TABLE IF NOT EXISTS floor_plans (
    property     TEXT NOT NULL DEFAULT 'desert-club',
    code         TEXT NOT NULL,
    bedrooms     INTEGER NOT NULL,
    bathrooms    INTEGER NOT NULL,
    sqft         INTEGER NOT NULL,
    deposit      INTEGER,
    is_renovated BOOLEAN NOT NULL,
    features     TEXT,
    updated_at   TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (property, code)
);

CREATE TABLE IF NOT EXISTS apartments (
    property       TEXT NOT NULL DEFAULT 'desert-club',
    unit_number    TEXT NOT NULL,
    floor_plan     TEXT NOT NULL,
    price          INTEGER,
    available_date TEXT,
    available_now  BOOLEAN NOT NULL DEFAULT 0,
    floor          INTEGER,
    amenities      TEXT,
    is_available   BOOLEAN NOT NULL DEFAULT 1,
    first_seen     TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_seen      TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (property, unit_number)
);

CREATE TABLE IF NOT EXISTS price_history (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    property    TEXT NOT NULL DEFAULT 'desert-club',
    unit_number TEXT NOT NULL,
    price       INTEGER NOT NULL,
    scraped_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_price_history_prop_unit ON price_history(property, unit_number);
CREATE INDEX IF NOT EXISTS idx_price_history_date ON price_history(scraped_at);

CREATE TABLE IF NOT EXISTS scrape_runs (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    property      TEXT NOT NULL DEFAULT 'desert-club',
    started_at    TIMESTAMP NOT NULL,
    completed_at  TIMESTAMP,
    floor_plans   INTEGER NOT NULL DEFAULT 0,
    units_found   INTEGER NOT NULL DEFAULT 0,
    units_new     INTEGER NOT NULL DEFAULT 0,
    units_removed INTEGER NOT NULL DEFAULT 0,
    units_changed INTEGER NOT NULL DEFAULT 0,
    error         TEXT
);
`

func (db *DB) migrate() error {
	_, err := db.Exec(schema)
	return err
}
