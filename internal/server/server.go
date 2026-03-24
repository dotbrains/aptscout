package server

import (
	"context"
	"embed"
	"encoding/json"
	"io/fs"
	"net"
	"net/http"
	"strconv"
	"strings"

	"github.com/dotbrains/aptscout/internal/db"
	"github.com/dotbrains/aptscout/internal/models"
	"github.com/dotbrains/aptscout/internal/provider"
	"github.com/dotbrains/aptscout/internal/scraper"
)

//go:embed static/*
var staticFS embed.FS

// Server serves the web UI and API.
type Server struct {
	db  *db.DB
	mux *http.ServeMux
}

// New creates a new Server.
func New(database *db.DB) *Server {
	s := &Server{db: database, mux: http.NewServeMux()}
	s.routes()
	return s
}

// ListenAndServe starts the HTTP server on the given address.
func (s *Server) ListenAndServe(addr string) error {
	return http.ListenAndServe(addr, s.mux)
}

// Serve starts the HTTP server on an existing listener.
func (s *Server) Serve(ln net.Listener) error {
	return http.Serve(ln, s.mux)
}

func (s *Server) routes() {
	// API routes
	s.mux.HandleFunc("/api/apartments", s.handleApartments)
	s.mux.HandleFunc("/api/apartments/", s.handleApartmentDetail)
	s.mux.HandleFunc("/api/floor-plans", s.handleFloorPlans)
	s.mux.HandleFunc("/api/stats", s.handleStats)
	s.mux.HandleFunc("/api/scrape-runs", s.handleScrapeRuns)
	s.mux.HandleFunc("/api/scrape", s.handleScrape)

	// Static files
	sub, _ := fs.Sub(staticFS, "static")
	s.mux.Handle("/", http.FileServer(http.FS(sub)))
}

func (s *Server) handleApartments(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	f := models.ApartmentFilter{Sort: "price", Order: "asc"}
	q := r.URL.Query()

	if v := q.Get("beds"); v != "" {
		n, _ := strconv.Atoi(v)
		f.Beds = &n
	}
	if v := q.Get("baths"); v != "" {
		n, _ := strconv.Atoi(v)
		f.Baths = &n
	}
	if v := q.Get("min_price"); v != "" {
		n, _ := strconv.Atoi(v)
		f.MinPrice = &n
	}
	if v := q.Get("max_price"); v != "" {
		n, _ := strconv.Atoi(v)
		f.MaxPrice = &n
	}
	if v := q.Get("property"); v != "" {
		f.Property = &v
	}
	if v := q.Get("plan"); v != "" {
		f.Plan = &v
	}
	if v := q.Get("renovated"); v == "true" {
		b := true
		f.Renovated = &b
	}
	if v := q.Get("available_by"); v != "" {
		f.AvailableBy = &v
	}
	if v := q.Get("sort"); v != "" {
		f.Sort = v
	}
	if v := q.Get("order"); v != "" {
		f.Order = v
	}

	apts, err := s.db.ListApartments(f)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if apts == nil {
		apts = []models.Apartment{}
	}

	jsonResponse(w, apts)
}

func (s *Server) handleApartmentDetail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Path: /api/apartments/{property}/{unit}
	path := strings.TrimPrefix(r.URL.Path, "/api/apartments/")
	parts := strings.SplitN(path, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		jsonError(w, "path must be /api/apartments/{property}/{unit}", http.StatusBadRequest)
		return
	}
	property, unitNumber := parts[0], parts[1]

	apt, err := s.db.GetApartment(property, unitNumber)
	if err != nil {
		jsonError(w, "unit not found", http.StatusNotFound)
		return
	}

	history, err := s.db.GetPriceHistory(property, unitNumber)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if history == nil {
		history = []models.PriceRecord{}
	}

	jsonResponse(w, models.ApartmentDetail{Apartment: *apt, PriceHistory: history})
}

func (s *Server) handleFloorPlans(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var propFilter *string
	if v := r.URL.Query().Get("property"); v != "" {
		propFilter = &v
	}
	plans, err := s.db.GetFloorPlans(propFilter)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if plans == nil {
		plans = []models.FloorPlan{}
	}

	jsonResponse(w, plans)
}

func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var propFilter *string
	if v := r.URL.Query().Get("property"); v != "" {
		propFilter = &v
	}
	stats, err := s.db.GetStats(propFilter)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	jsonResponse(w, stats)
}

func (s *Server) handleScrapeRuns(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	runs, err := s.db.GetScrapeRuns()
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if runs == nil {
		runs = []models.ScrapeRun{}
	}

	jsonResponse(w, runs)
}

func (s *Server) handleScrape(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	sc := scraper.New(s.db, discard{})
	var totalResult scraper.Result
	for _, prov := range provider.All {
		r, err := sc.RunProvider(context.Background(), prov)
		if err != nil {
			continue
		}
		totalResult.UnitsFound += r.UnitsFound
		totalResult.UnitsNew += r.UnitsNew
		totalResult.UnitsChanged += r.UnitsChanged
		totalResult.UnitsRemoved += r.UnitsRemoved
		totalResult.FloorPlans += r.FloorPlans
	}
	result := &totalResult
	var err error
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	jsonResponse(w, result)
}

func jsonResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache")
	_ = json.NewEncoder(w).Encode(data)
}

func jsonError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

// discard implements io.Writer and discards all output.
type discard struct{}

func (discard) Write(p []byte) (int, error) { return len(p), nil }
