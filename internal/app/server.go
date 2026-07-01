package app

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"tmmweb/internal/media"
	"tmmweb/internal/nfo"
	"tmmweb/internal/tmdb"
)

type Config struct {
	DataDir string
	TMDBKey string
	Client  *http.Client
}

type Server struct {
	config    Config
	mu        sync.Mutex
	libraries []media.Library
	items     map[string]media.Item
	tmdb      tmdb.Client
}

func NewServer(config Config) (*Server, error) {
	if err := os.MkdirAll(config.DataDir, 0755); err != nil {
		return nil, err
	}
	server := &Server{
		config: config,
		items:  map[string]media.Item{},
		tmdb:   tmdb.Client{Key: config.TMDBKey, HTTP: config.Client, Lang: "zh-CN"},
	}
	_ = server.loadLibraries()
	return server, nil
}

func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/health", s.handleHealth)
	mux.HandleFunc("/api/libraries", s.handleLibraries)
	mux.HandleFunc("/api/scan", s.handleScan)
	mux.HandleFunc("/api/search", s.handleSearch)
	mux.HandleFunc("/api/scrape", s.handleScrape)
	mux.HandleFunc("/api/rename/preview", s.handleRenamePreview)
	mux.HandleFunc("/api/rename/apply", s.handleRenameApply)
	mux.Handle("/", staticHandler())
	return logMiddleware(mux)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]interface{}{"ok": true, "tmdbEnabled": s.tmdb.Enabled()})
}

func (s *Server) handleLibraries(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.mu.Lock()
		defer s.mu.Unlock()
		writeJSON(w, s.libraries)
	case http.MethodPost:
		var input media.Library
		if !decodeJSON(w, r, &input) {
			return
		}
		if input.Path == "" {
			http.Error(w, "path is required", http.StatusBadRequest)
			return
		}
		if input.Name == "" {
			input.Name = filepath.Base(input.Path)
		}
		if input.Type == "" {
			input.Type = "movie"
		}
		input.ID = randomID()
		s.mu.Lock()
		s.libraries = append(s.libraries, input)
		err := s.saveLibraries()
		s.mu.Unlock()
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		writeJSON(w, input)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleScan(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", 405)
		return
	}
	var input struct {
		LibraryID string `json:"libraryId"`
	}
	if !decodeJSON(w, r, &input) {
		return
	}
	library, ok := s.findLibrary(input.LibraryID)
	if !ok {
		http.Error(w, "library not found", 404)
		return
	}
	items, err := media.ScanLibrary(library)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	s.mu.Lock()
	for _, item := range items {
		s.items[item.ID] = item
	}
	s.mu.Unlock()
	writeJSON(w, map[string]interface{}{"items": items, "count": len(items)})
}

func (s *Server) handleSearch(w http.ResponseWriter, r *http.Request) {
	itemID := r.URL.Query().Get("itemId")
	query := r.URL.Query().Get("q")
	year := r.URL.Query().Get("year")
	if query == "" && itemID != "" {
		if item, ok := s.findItem(itemID); ok {
			query = item.TitleGuess
			year = item.YearGuess
		}
	}
	results, err := s.tmdb.SearchMovie(query, year)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	writeJSON(w, results)
}

func (s *Server) handleScrape(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", 405)
		return
	}
	var input struct {
		ItemID      int    `json:"-"`
		Item        string `json:"itemId"`
		TMDBID      int    `json:"tmdbId"`
		WriteNFO    bool   `json:"writeNfo"`
		WriteImages bool   `json:"writeImages"`
	}
	if !decodeJSON(w, r, &input) {
		return
	}
	item, ok := s.findItem(input.Item)
	if !ok {
		http.Error(w, "item not found; run scan first", 404)
		return
	}
	movie, err := s.tmdb.Movie(input.TMDBID)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	if input.WriteNFO {
		if err := nfo.WriteMovie(item.Dir, movie); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
	}
	if input.WriteImages {
		if movie.PosterPath != "" {
			if data, err := s.tmdb.DownloadImage(movie.PosterPath); err == nil {
				_ = nfo.WriteImage(item.Dir, "poster.jpg", data)
			}
		}
		if movie.BackdropPath != "" {
			if data, err := s.tmdb.DownloadImage(movie.BackdropPath); err == nil {
				_ = nfo.WriteImage(item.Dir, "fanart.jpg", data)
			}
		}
	}
	item.MatchedID = movie.ID
	item.MatchedName = movie.Title
	s.mu.Lock()
	s.items[item.ID] = item
	s.mu.Unlock()
	writeJSON(w, map[string]interface{}{"item": item, "movie": movie})
}

func (s *Server) handleRenamePreview(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ItemID  string `json:"itemId"`
		Title   string `json:"title"`
		Year    string `json:"year"`
		TMDBID  int    `json:"tmdbId"`
		Pattern string `json:"pattern"`
	}
	if !decodeJSON(w, r, &input) {
		return
	}
	item, ok := s.findItem(input.ItemID)
	if !ok {
		http.Error(w, "item not found", 404)
		return
	}
	preview := media.BuildMovieRename(item, input.Title, input.Year, input.TMDBID, input.Pattern)
	writeJSON(w, preview)
}

func (s *Server) handleRenameApply(w http.ResponseWriter, r *http.Request) {
	var preview media.RenamePreview
	if !decodeJSON(w, r, &preview) {
		return
	}
	if err := media.ApplyRename(preview); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	writeJSON(w, map[string]interface{}{"ok": true})
}

func (s *Server) findLibrary(id string) (media.Library, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, library := range s.libraries {
		if library.ID == id {
			return library, true
		}
	}
	return media.Library{}, false
}

func (s *Server) findItem(id string) (media.Item, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	item, ok := s.items[id]
	return item, ok
}

func (s *Server) loadLibraries() error {
	path := filepath.Join(s.config.DataDir, "libraries.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, &s.libraries)
}

func (s *Server) saveLibraries() error {
	data, err := json.MarshalIndent(s.libraries, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(s.config.DataDir, "libraries.json"), data, 0644)
}

func writeJSON(w http.ResponseWriter, value interface{}) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(value)
}

func decodeJSON(w http.ResponseWriter, r *http.Request, out interface{}) bool {
	if err := json.NewDecoder(r.Body).Decode(out); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return false
	}
	return true
}

func randomID() string {
	var buf [8]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return strconv.FormatInt(int64(os.Getpid()), 16)
	}
	return hex.EncodeToString(buf[:])
}

func staticHandler() http.Handler {
	sub, err := fs.Sub(staticFS, "static")
	if err != nil {
		panic(err)
	}
	fileServer := http.FileServer(http.FS(sub))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/")
		if path == "" {
			r.URL.Path = "/"
			fileServer.ServeHTTP(w, r)
			return
		}
		if _, err := fs.Stat(sub, path); err != nil {
			r.URL.Path = "/"
		}
		fileServer.ServeHTTP(w, r)
	})
}

func logMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}

var _ = fmt.Sprintf
