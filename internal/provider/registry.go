package provider

import (
	"github.com/dotbrains/aptscout/internal/models"
	"github.com/dotbrains/aptscout/internal/provider/desertclub"
	"github.com/dotbrains/aptscout/internal/provider/hideaway"
)

// All is the registry of all available providers.
var All = map[string]models.Provider{
	"desert-club": desertclub.New(),
	"hideaway":    hideaway.New(),
}

// Get returns a provider by ID, or nil if not found.
func Get(id string) models.Provider {
	return All[id]
}

// IDs returns a sorted list of all provider IDs.
func IDs() []string {
	ids := make([]string, 0, len(All))
	for id := range All {
		ids = append(ids, id)
	}
	return ids
}
