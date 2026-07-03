package app

import (
	"encoding/json"
	"net/http"
)

// Routes wires the API surface in one place. Keeping route registration separate
// from the handlers makes the HTTP contract easier to audit while the heavier
// media, scraping, rename and scan workflows remain in server.go.
func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()

	// System and configuration.
	mux.HandleFunc("/api/health", s.handleHealth)
	mux.HandleFunc("/api/settings", s.handleSettings)

	// Library and scan lifecycle.
	mux.HandleFunc("/api/libraries", s.handleLibraries)
	mux.HandleFunc("/api/scan", s.handleScan)
	mux.HandleFunc("/api/scan/cancel", s.handleScanCancel)
	mux.HandleFunc("/api/tasks", s.handleTasks)
	mux.HandleFunc("/api/items", s.handleItems)
	mux.HandleFunc("/api/browse", s.handleBrowse)

	// Metadata, artwork and rename workflows.
	mux.HandleFunc("/api/artwork", s.handleArtwork)
	mux.HandleFunc("/api/search", s.handleSearch)
	mux.HandleFunc("/api/metadata", s.handleMetadata)
	mux.HandleFunc("/api/scrape", s.handleScrape)
	mux.HandleFunc("/api/rename/preview", s.handleRenamePreview)
	mux.HandleFunc("/api/rename/apply", s.handleRenameApply)
	mux.HandleFunc("/api/local-rename", s.handleLocalRename)

	mux.Handle("/", staticHandler())
	return logMiddleware(mux)
}

func writeJSON(w http.ResponseWriter, value interface{}) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(value)
}

func decodeJSON(w http.ResponseWriter, r *http.Request, out interface{}) bool {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(out); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return false
	}
	return true
}
