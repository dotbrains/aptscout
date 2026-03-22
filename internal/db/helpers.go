package db

import "time"

// now returns the current time formatted for SQLite.
func now() string {
	return time.Now().UTC().Format(time.RFC3339)
}
