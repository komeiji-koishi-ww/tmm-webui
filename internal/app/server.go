package app

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"tmmweb/internal/media"
	"tmmweb/internal/nfo"
	"tmmweb/internal/store"
	"tmmweb/internal/tmdb"
)

const (
	scanPersistBatchSize     = 100
	scanPersistFlushInterval = 2 * time.Second
)

type Config struct {
	DataDir string
	TMDBKey string
	Client  *http.Client
}

type AppSettings struct {
	TMDBAPIKey              string `json:"tmdbApiKey"`
	ProxyEnabled            bool   `json:"proxyEnabled"`
	ProxyHost               string `json:"proxyHost"`
	ProxyPort               int    `json:"proxyPort"`
	ProxyUsername           string `json:"proxyUsername"`
	ProxyPassword           string `json:"proxyPassword"`
	MovieScrapeMetadata     *bool  `json:"movieScrapeMetadata,omitempty"`
	MovieScrapeNFO          *bool  `json:"movieScrapeNfo,omitempty"`
	MovieScrapeImages       *bool  `json:"movieScrapeImages,omitempty"`
	MovieScrapeOverwrite    *bool  `json:"movieScrapeOverwrite,omitempty"`
	TVShowScrapeMetadata    *bool  `json:"tvShowScrapeMetadata,omitempty"`
	TVShowEpisodeMetadata   *bool  `json:"tvShowEpisodeMetadata,omitempty"`
	TVShowScrapeNFO         *bool  `json:"tvShowScrapeNfo,omitempty"`
	TVShowScrapeImages      *bool  `json:"tvShowScrapeImages,omitempty"`
	TVShowScrapeOverwrite   *bool  `json:"tvShowScrapeOverwrite,omitempty"`
	MovieRenamerPathname    string `json:"movieRenamerPathname"`
	MovieRenamerFilename    string `json:"movieRenamerFilename"`
	TVShowRenamerShowFolder string `json:"tvShowRenamerShowFolder"`
	TVShowRenamerSeason     string `json:"tvShowRenamerSeason"`
	TVShowRenamerFilename   string `json:"tvShowRenamerFilename"`
	MoviePosterName         string `json:"moviePosterName"`
	MovieFanartName         string `json:"movieFanartName"`
	MoviePosterNames        string `json:"moviePosterNames"`
	MovieFanartNames        string `json:"movieFanartNames"`
	TVShowPosterName        string `json:"tvShowPosterName"`
	TVShowFanartName        string `json:"tvShowFanartName"`
	TVShowPosterNames       string `json:"tvShowPosterNames"`
	TVShowFanartNames       string `json:"tvShowFanartNames"`
}

type SettingsResponse struct {
	TMDBConfigured          bool   `json:"tmdbConfigured"`
	TMDBEnabled             bool   `json:"tmdbEnabled"`
	TMDBKeySource           string `json:"tmdbKeySource"`
	ProxyEnabled            bool   `json:"proxyEnabled"`
	ProxyHost               string `json:"proxyHost"`
	ProxyPort               int    `json:"proxyPort"`
	ProxyUsername           string `json:"proxyUsername"`
	ProxyPassword           bool   `json:"proxyPassword"`
	MovieScrapeMetadata     bool   `json:"movieScrapeMetadata"`
	MovieScrapeNFO          bool   `json:"movieScrapeNfo"`
	MovieScrapeImages       bool   `json:"movieScrapeImages"`
	MovieScrapeOverwrite    bool   `json:"movieScrapeOverwrite"`
	TVShowScrapeMetadata    bool   `json:"tvShowScrapeMetadata"`
	TVShowEpisodeMetadata   bool   `json:"tvShowEpisodeMetadata"`
	TVShowScrapeNFO         bool   `json:"tvShowScrapeNfo"`
	TVShowScrapeImages      bool   `json:"tvShowScrapeImages"`
	TVShowScrapeOverwrite   bool   `json:"tvShowScrapeOverwrite"`
	MovieRenamerPathname    string `json:"movieRenamerPathname"`
	MovieRenamerFilename    string `json:"movieRenamerFilename"`
	TVShowRenamerShowFolder string `json:"tvShowRenamerShowFolder"`
	TVShowRenamerSeason     string `json:"tvShowRenamerSeason"`
	TVShowRenamerFilename   string `json:"tvShowRenamerFilename"`
	MoviePosterName         string `json:"moviePosterName"`
	MovieFanartName         string `json:"movieFanartName"`
	MoviePosterNames        string `json:"moviePosterNames"`
	MovieFanartNames        string `json:"movieFanartNames"`
	TVShowPosterName        string `json:"tvShowPosterName"`
	TVShowFanartName        string `json:"tvShowFanartName"`
	TVShowPosterNames       string `json:"tvShowPosterNames"`
	TVShowFanartNames       string `json:"tvShowFanartNames"`
}

type SettingsUpdate struct {
	TMDBAPIKey              *string `json:"tmdbApiKey"`
	ClearTMDBKey            bool    `json:"clearTmdbKey"`
	ProxyEnabled            bool    `json:"proxyEnabled"`
	ProxyHost               string  `json:"proxyHost"`
	ProxyPort               int     `json:"proxyPort"`
	ProxyUsername           string  `json:"proxyUsername"`
	ProxyPassword           *string `json:"proxyPassword"`
	ClearProxyPassword      bool    `json:"clearProxyPassword"`
	MovieScrapeMetadata     *bool   `json:"movieScrapeMetadata"`
	MovieScrapeNFO          *bool   `json:"movieScrapeNfo"`
	MovieScrapeImages       *bool   `json:"movieScrapeImages"`
	MovieScrapeOverwrite    *bool   `json:"movieScrapeOverwrite"`
	TVShowScrapeMetadata    *bool   `json:"tvShowScrapeMetadata"`
	TVShowEpisodeMetadata   *bool   `json:"tvShowEpisodeMetadata"`
	TVShowScrapeNFO         *bool   `json:"tvShowScrapeNfo"`
	TVShowScrapeImages      *bool   `json:"tvShowScrapeImages"`
	TVShowScrapeOverwrite   *bool   `json:"tvShowScrapeOverwrite"`
	MovieRenamerPathname    string  `json:"movieRenamerPathname"`
	MovieRenamerFilename    string  `json:"movieRenamerFilename"`
	TVShowRenamerShowFolder string  `json:"tvShowRenamerShowFolder"`
	TVShowRenamerSeason     string  `json:"tvShowRenamerSeason"`
	TVShowRenamerFilename   string  `json:"tvShowRenamerFilename"`
	MoviePosterName         string  `json:"moviePosterName"`
	MovieFanartName         string  `json:"movieFanartName"`
	MoviePosterNames        string  `json:"moviePosterNames"`
	MovieFanartNames        string  `json:"movieFanartNames"`
	TVShowPosterName        string  `json:"tvShowPosterName"`
	TVShowFanartName        string  `json:"tvShowFanartName"`
	TVShowPosterNames       string  `json:"tvShowPosterNames"`
	TVShowFanartNames       string  `json:"tvShowFanartNames"`
}

type Server struct {
	config    Config
	mu        sync.Mutex
	libraries []media.Library
	items     map[string]media.Item
	tasks     map[string]*Task
	settings  AppSettings
	store     *store.Store
	tmdb      tmdb.Client
}

type Task struct {
	ID           string `json:"id"`
	Type         string `json:"type"`
	LibraryID    string `json:"libraryId"`
	LibraryName  string `json:"libraryName"`
	State        string `json:"state"`
	SourcePath   string `json:"sourcePath,omitempty"`
	CurrentPath  string `json:"currentPath,omitempty"`
	VisitedFiles int    `json:"visitedFiles"`
	FoundItems   int    `json:"foundItems"`
	ResultCount  int    `json:"resultCount"`
	Error        string `json:"error,omitempty"`
	StartedAt    string `json:"startedAt"`
	FinishedAt   string `json:"finishedAt,omitempty"`
}

func NewServer(config Config) (*Server, error) {
	if err := os.MkdirAll(config.DataDir, 0755); err != nil {
		return nil, err
	}
	db, err := store.Open(config.DataDir)
	if err != nil {
		return nil, err
	}
	settings, err := loadAppSettings(config.DataDir)
	if err != nil {
		_ = db.Close()
		return nil, err
	}
	tmdbClient, err := newTMDBClient(config, settings)
	if err != nil {
		_ = db.Close()
		return nil, err
	}
	server := &Server{
		config:   config,
		items:    map[string]media.Item{},
		tasks:    map[string]*Task{},
		settings: settings,
		store:    db,
		tmdb:     tmdbClient,
	}
	if err := server.loadLibraries(); err != nil {
		_ = db.Close()
		return nil, err
	}
	if err := server.loadItems(); err != nil {
		_ = db.Close()
		return nil, err
	}
	_ = server.loadTasks()
	_ = server.migrateJSON()
	return server, nil
}

func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/health", s.handleHealth)
	mux.HandleFunc("/api/settings", s.handleSettings)
	mux.HandleFunc("/api/libraries", s.handleLibraries)
	mux.HandleFunc("/api/scan", s.handleScan)
	mux.HandleFunc("/api/scan/cancel", s.handleScanCancel)
	mux.HandleFunc("/api/tasks", s.handleTasks)
	mux.HandleFunc("/api/items", s.handleItems)
	mux.HandleFunc("/api/browse", s.handleBrowse)
	mux.HandleFunc("/api/search", s.handleSearch)
	mux.HandleFunc("/api/metadata", s.handleMetadata)
	mux.HandleFunc("/api/scrape", s.handleScrape)
	mux.HandleFunc("/api/rename/preview", s.handleRenamePreview)
	mux.HandleFunc("/api/rename/apply", s.handleRenameApply)
	mux.Handle("/", staticHandler())
	return logMiddleware(mux)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	client := s.tmdbClient()
	writeJSON(w, map[string]interface{}{"ok": true, "tmdbEnabled": client.Enabled()})
}

func (s *Server) handleSettings(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.mu.Lock()
		defer s.mu.Unlock()
		writeJSON(w, s.settingsResponseLocked())
	case http.MethodPut:
		var input SettingsUpdate
		if !decodeJSON(w, r, &input) {
			return
		}
		s.mu.Lock()
		settings := s.settings
		s.mu.Unlock()
		if input.ClearTMDBKey {
			settings.TMDBAPIKey = ""
		} else if input.TMDBAPIKey != nil && strings.TrimSpace(*input.TMDBAPIKey) != "" {
			settings.TMDBAPIKey = strings.TrimSpace(*input.TMDBAPIKey)
		}
		settings.ProxyEnabled = input.ProxyEnabled
		settings.ProxyHost = strings.TrimSpace(input.ProxyHost)
		settings.ProxyPort = input.ProxyPort
		settings.ProxyUsername = strings.TrimSpace(input.ProxyUsername)
		if input.ClearProxyPassword {
			settings.ProxyPassword = ""
		} else if input.ProxyPassword != nil && *input.ProxyPassword != "" {
			settings.ProxyPassword = *input.ProxyPassword
		}
		if input.MovieScrapeMetadata != nil {
			settings.MovieScrapeMetadata = input.MovieScrapeMetadata
		}
		if input.MovieScrapeNFO != nil {
			settings.MovieScrapeNFO = input.MovieScrapeNFO
		}
		if input.MovieScrapeImages != nil {
			settings.MovieScrapeImages = input.MovieScrapeImages
		}
		if input.MovieScrapeOverwrite != nil {
			settings.MovieScrapeOverwrite = input.MovieScrapeOverwrite
		}
		if input.TVShowScrapeMetadata != nil {
			settings.TVShowScrapeMetadata = input.TVShowScrapeMetadata
		}
		if input.TVShowEpisodeMetadata != nil {
			settings.TVShowEpisodeMetadata = input.TVShowEpisodeMetadata
		}
		if input.TVShowScrapeNFO != nil {
			settings.TVShowScrapeNFO = input.TVShowScrapeNFO
		}
		if input.TVShowScrapeImages != nil {
			settings.TVShowScrapeImages = input.TVShowScrapeImages
		}
		if input.TVShowScrapeOverwrite != nil {
			settings.TVShowScrapeOverwrite = input.TVShowScrapeOverwrite
		}
		settings.MovieRenamerPathname = strings.TrimSpace(input.MovieRenamerPathname)
		settings.MovieRenamerFilename = strings.TrimSpace(input.MovieRenamerFilename)
		settings.TVShowRenamerShowFolder = strings.TrimSpace(input.TVShowRenamerShowFolder)
		settings.TVShowRenamerSeason = strings.TrimSpace(input.TVShowRenamerSeason)
		settings.TVShowRenamerFilename = strings.TrimSpace(input.TVShowRenamerFilename)
		settings.MoviePosterName = strings.TrimSpace(input.MoviePosterName)
		settings.MovieFanartName = strings.TrimSpace(input.MovieFanartName)
		settings.MoviePosterNames = strings.TrimSpace(input.MoviePosterNames)
		settings.MovieFanartNames = strings.TrimSpace(input.MovieFanartNames)
		settings.TVShowPosterName = strings.TrimSpace(input.TVShowPosterName)
		settings.TVShowFanartName = strings.TrimSpace(input.TVShowFanartName)
		settings.TVShowPosterNames = strings.TrimSpace(input.TVShowPosterNames)
		settings.TVShowFanartNames = strings.TrimSpace(input.TVShowFanartNames)
		tmdbClient, err := newTMDBClient(s.config, settings)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := saveAppSettings(s.config.DataDir, settings); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		s.mu.Lock()
		s.settings = settings
		s.tmdb = tmdbClient
		response := s.settingsResponseLocked()
		s.mu.Unlock()
		writeJSON(w, response)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
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
		input.Paths = media.NormalizePaths(input)
		if len(input.Paths) == 0 {
			http.Error(w, "path is required", http.StatusBadRequest)
			return
		}
		input.Path = input.Paths[0]
		if input.Name == "" {
			input.Name = filepath.Base(input.Paths[0])
		}
		if input.Type == "" {
			input.Type = "movie"
		}
		input.ID = randomID()
		s.mu.Lock()
		s.libraries = append(s.libraries, input)
		err := s.store.SaveLibrary(input)
		s.mu.Unlock()
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		writeJSON(w, input)
	case http.MethodDelete:
		id := r.URL.Query().Get("id")
		if id == "" {
			http.Error(w, "id is required", http.StatusBadRequest)
			return
		}
		s.mu.Lock()
		found := false
		libraries := make([]media.Library, 0, len(s.libraries))
		for _, library := range s.libraries {
			if library.ID == id {
				found = true
				continue
			}
			libraries = append(libraries, library)
		}
		if !found {
			s.mu.Unlock()
			http.Error(w, "library not found", http.StatusNotFound)
			return
		}
		s.libraries = libraries
		for itemID, item := range s.items {
			if item.LibraryID == id {
				delete(s.items, itemID)
			}
		}
		err := s.store.DeleteLibrary(id)
		if err == nil {
			err = s.store.DeleteLibraryItems(id)
		}
		s.mu.Unlock()
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		writeJSON(w, map[string]interface{}{"ok": true})
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
	s.mu.Lock()
	if task := s.runningScanTaskLocked(library.ID); task != nil {
		s.mu.Unlock()
		writeJSON(w, map[string]interface{}{"task": task, "started": false})
		return
	}
	task := &Task{
		ID:          randomID(),
		Type:        "scan",
		LibraryID:   library.ID,
		LibraryName: library.Name,
		State:       "running",
		StartedAt:   time.Now().Format(time.RFC3339),
	}
	s.tasks[task.ID] = task
	_ = s.store.SaveTask(task.toRecord())
	s.mu.Unlock()
	go s.runScanTask(task.ID, library)
	writeJSON(w, map[string]interface{}{"task": task, "started": true})
}

func (s *Server) handleScanCancel(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", 405)
		return
	}
	var input struct {
		LibraryID string `json:"libraryId"`
		TaskID    string `json:"taskId"`
	}
	if !decodeJSON(w, r, &input) {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	var task *Task
	if input.TaskID != "" {
		task = s.tasks[input.TaskID]
	} else if input.LibraryID != "" {
		task = s.runningScanTaskLocked(input.LibraryID)
	}
	if task == nil {
		http.Error(w, "running scan task not found", 404)
		return
	}
	if task.State == "running" {
		task.State = "canceling"
		task.Error = "正在停止扫描"
		_ = s.store.SaveTask(task.toRecord())
	}
	writeJSON(w, map[string]interface{}{"task": task, "canceled": task.State == "canceling" || task.State == "canceled"})
}

func (s *Server) handleTasks(w http.ResponseWriter, r *http.Request) {
	libraryID := r.URL.Query().Get("libraryId")
	taskID := r.URL.Query().Get("taskId")
	s.mu.Lock()
	defer s.mu.Unlock()
	if taskID != "" {
		if task, ok := s.tasks[taskID]; ok {
			writeJSON(w, task)
			return
		}
		http.Error(w, "task not found", 404)
		return
	}
	tasks := make([]*Task, 0, len(s.tasks))
	for _, task := range s.tasks {
		if libraryID == "" || task.LibraryID == libraryID {
			tasks = append(tasks, task)
		}
	}
	writeJSON(w, map[string]interface{}{"tasks": tasks})
}

func (s *Server) handleItems(w http.ResponseWriter, r *http.Request) {
	libraryID := r.URL.Query().Get("libraryId")
	s.mu.Lock()
	defer s.mu.Unlock()
	items := make([]media.Item, 0, len(s.items))
	for _, item := range s.items {
		if libraryID == "" || item.LibraryID == libraryID {
			items = append(items, item)
		}
	}
	writeJSON(w, map[string]interface{}{"items": items, "count": len(items)})
}

func (s *Server) handleBrowse(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	if path == "" {
		path = "/"
	}
	path = filepath.Clean(path)
	entries, err := os.ReadDir(path)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	type entry struct {
		Name string `json:"name"`
		Path string `json:"path"`
		Dir  bool   `json:"dir"`
	}
	result := struct {
		Path    string  `json:"path"`
		Parent  string  `json:"parent"`
		Entries []entry `json:"entries"`
	}{Path: path, Parent: filepath.Dir(path)}
	for _, file := range entries {
		if !file.IsDir() {
			continue
		}
		name := file.Name()
		if strings.HasPrefix(name, ".") && path != "/" && path != "/Volumes" {
			continue
		}
		result.Entries = append(result.Entries, entry{
			Name: name,
			Path: filepath.Join(path, name),
			Dir:  true,
		})
	}
	writeJSON(w, result)
}

func (s *Server) handleSearch(w http.ResponseWriter, r *http.Request) {
	itemID := r.URL.Query().Get("itemId")
	query := r.URL.Query().Get("q")
	year := r.URL.Query().Get("year")
	mediaType := r.URL.Query().Get("type")
	if query == "" && itemID != "" {
		if item, ok := s.findItem(itemID); ok {
			query = item.TitleGuess
			year = item.YearGuess
			mediaType = item.Kind
		}
	}
	client := s.tmdbClient()
	var (
		results []tmdb.SearchResult
		err     error
	)
	if mediaType == "tvshow" {
		results, err = client.SearchTV(query, year)
	} else {
		results, err = client.SearchMovie(query, year)
	}
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	writeJSON(w, results)
}

func (s *Server) handleMetadata(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil || id <= 0 {
		http.Error(w, "id is required", http.StatusBadRequest)
		return
	}
	mediaType := r.URL.Query().Get("type")
	client := s.tmdbClient()
	if mediaType == "tvshow" {
		show, err := client.TVShow(id)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		writeJSON(w, show)
		return
	}
	movie, err := client.Movie(id)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	writeJSON(w, movie)
}

func (s *Server) handleScrape(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", 405)
		return
	}
	var input struct {
		ItemID      int    `json:"-"`
		Item        string `json:"itemId"`
		Scope       string `json:"scope"`
		LibraryID   string `json:"libraryId"`
		ShowName    string `json:"showName"`
		Season      int    `json:"season"`
		TMDBID      int    `json:"tmdbId"`
		MediaType   string `json:"mediaType"`
		WriteNFO    bool   `json:"writeNfo"`
		WriteImages bool   `json:"writeImages"`
		WriteMeta   bool   `json:"writeMeta"`
		Overwrite   bool   `json:"overwrite"`
	}
	if !decodeJSON(w, r, &input) {
		return
	}
	item, ok := s.findItem(input.Item)
	if !ok {
		http.Error(w, "item not found; run scan first", 404)
		return
	}
	settings := s.appSettings()
	client := s.tmdbClient()
	if item.Kind == "tvshow" || input.MediaType == "tvshow" {
		show, err := client.TVShow(input.TMDBID)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		scope := strings.TrimSpace(input.Scope)
		if scope == "" {
			scope = "episode"
		}
		libraryID := input.LibraryID
		if libraryID == "" {
			libraryID = item.LibraryID
		}
		showName := strings.TrimSpace(input.ShowName)
		if showName == "" {
			showName = item.ShowGuess
		}
		season := input.Season
		if season == 0 {
			season = item.Season
		}
		showDir := item.Dir
		if item.Season > 0 || item.Episode > 0 {
			showDir = filepath.Dir(item.Dir)
		}
		if input.WriteNFO {
			if err := writeTVShowNFO(showDir, show, input.Overwrite); err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
		}
		if input.WriteImages {
			if show.PosterPath != "" {
				if data, err := client.DownloadImage(show.PosterPath); err == nil {
					_ = writeImages(showDir, imageNames(settings.TVShowPosterNames, settings.TVShowPosterName, defaultTVShowPosterNames(), item), data, input.Overwrite)
				}
			}
			if show.BackdropPath != "" {
				if data, err := client.DownloadImage(show.BackdropPath); err == nil {
					_ = writeImages(showDir, imageNames(settings.TVShowFanartNames, settings.TVShowFanartName, defaultTVShowFanartNames(), item), data, input.Overwrite)
				}
			}
		}
		updated := []media.Item{item}
		if scope == "show" || scope == "season" {
			updated = s.matchTVItems(libraryID, showName, season, scope, show)
		}
		if len(updated) == 0 {
			updated = append(updated, item)
		}
		for i := range updated {
			if input.WriteMeta {
				updated[i].MatchedID = show.ID
				updated[i].MatchedName = show.Title
			}
			if input.WriteImages {
				updated[i].HasPoster = updated[i].HasPoster || show.PosterPath != ""
				updated[i].HasFanart = updated[i].HasFanart || show.BackdropPath != ""
			}
		}
		s.mu.Lock()
		for _, changed := range updated {
			s.items[changed.ID] = changed
		}
		s.mu.Unlock()
		if err := s.store.SaveItems(updated); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		writeJSON(w, map[string]interface{}{"item": updated[0], "items": updated, "show": show, "scope": scope})
		return
	}
	movie, err := client.Movie(input.TMDBID)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	if input.WriteNFO {
		if err := writeMovieNFO(item.Dir, movie, input.Overwrite); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
	}
	if input.WriteImages {
		if movie.PosterPath != "" {
			if data, err := client.DownloadImage(movie.PosterPath); err == nil {
				_ = writeImages(item.Dir, imageNames(settings.MoviePosterNames, settings.MoviePosterName, defaultMoviePosterNames(), item), data, input.Overwrite)
			}
		}
		if movie.BackdropPath != "" {
			if data, err := client.DownloadImage(movie.BackdropPath); err == nil {
				_ = writeImages(item.Dir, imageNames(settings.MovieFanartNames, settings.MovieFanartName, defaultMovieFanartNames(), item), data, input.Overwrite)
			}
		}
	}
	if input.WriteMeta {
		item.MatchedID = movie.ID
		item.MatchedName = movie.Title
	}
	if input.WriteNFO {
		item.HasNFO = true
	}
	if input.WriteImages {
		item.HasPoster = item.HasPoster || movie.PosterPath != ""
		item.HasFanart = item.HasFanart || movie.BackdropPath != ""
	}
	s.mu.Lock()
	s.items[item.ID] = item
	s.mu.Unlock()
	if err := s.store.SaveItem(item); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	writeJSON(w, map[string]interface{}{"item": item, "movie": movie})
}

func (s *Server) handleRenamePreview(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ItemID                  string `json:"itemId"`
		Title                   string `json:"title"`
		Year                    string `json:"year"`
		TMDBID                  int    `json:"tmdbId"`
		Pattern                 string `json:"pattern"`
		MovieRenamerPathname    string `json:"movieRenamerPathname"`
		MovieRenamerFilename    string `json:"movieRenamerFilename"`
		TVShowRenamerShowFolder string `json:"tvShowRenamerShowFolder"`
		TVShowRenamerSeason     string `json:"tvShowRenamerSeason"`
		TVShowRenamerFilename   string `json:"tvShowRenamerFilename"`
	}
	if !decodeJSON(w, r, &input) {
		return
	}
	item, ok := s.findItem(input.ItemID)
	if !ok {
		http.Error(w, "item not found", 404)
		return
	}
	settings := s.appSettings()
	var preview media.RenamePreview
	if item.Kind == "tvshow" {
		preview = media.BuildTVShowRename(
			item,
			input.Title,
			"",
			input.Year,
			input.TMDBID,
			defaultString(input.TVShowRenamerShowFolder, defaultString(settings.TVShowRenamerShowFolder, "{showTitle}")),
			defaultString(input.TVShowRenamerSeason, defaultString(settings.TVShowRenamerSeason, "Season {seasonNr2}")),
			defaultString(input.TVShowRenamerFilename, defaultString(settings.TVShowRenamerFilename, "{showTitle} - S{seasonNr2}E{episodeNr2} - {title}")),
		)
	} else {
		folderPattern := input.MovieRenamerPathname
		filePattern := input.MovieRenamerFilename
		if strings.TrimSpace(folderPattern) == "" && strings.TrimSpace(filePattern) == "" {
			folderPattern = input.Pattern
			filePattern = input.Pattern
		}
		preview = media.BuildMovieRenameWithPatterns(
			item,
			input.Title,
			input.Year,
			input.TMDBID,
			defaultString(folderPattern, defaultString(settings.MovieRenamerPathname, "{title} ({year})")),
			defaultString(filePattern, defaultString(settings.MovieRenamerFilename, "{title} ({year})")),
		)
	}
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

func (s *Server) matchTVItems(libraryID string, showName string, season int, scope string, show tmdb.TVShow) []media.Item {
	s.mu.Lock()
	defer s.mu.Unlock()
	var updated []media.Item
	for _, item := range s.items {
		if item.LibraryID != libraryID || item.Kind != "tvshow" {
			continue
		}
		if showName != "" && item.ShowGuess != showName {
			continue
		}
		if scope == "season" && season > 0 && item.Season != season {
			continue
		}
		item.MatchedID = show.ID
		item.MatchedName = show.Title
		item.HasPoster = item.HasPoster || show.PosterPath != ""
		item.HasFanart = item.HasFanart || show.BackdropPath != ""
		updated = append(updated, item)
	}
	return updated
}

func writeMovieNFO(dir string, movie tmdb.Movie, overwrite bool) error {
	path := filepath.Join(dir, "movie.nfo")
	if !overwrite {
		if _, err := os.Stat(path); err == nil {
			return nil
		}
	}
	return nfo.WriteMovie(dir, movie)
}

func writeTVShowNFO(dir string, show tmdb.TVShow, overwrite bool) error {
	path := filepath.Join(dir, "tvshow.nfo")
	if !overwrite {
		if _, err := os.Stat(path); err == nil {
			return nil
		}
	}
	return nfo.WriteTVShow(dir, show)
}

func writeImage(dir string, name string, data []byte, overwrite bool) error {
	if !overwrite {
		if _, err := os.Stat(filepath.Join(dir, name)); err == nil {
			return nil
		}
	}
	return nfo.WriteImage(dir, name, data)
}

func writeImages(dir string, names []string, data []byte, overwrite bool) error {
	var firstErr error
	for _, name := range names {
		if err := writeImage(dir, name, data, overwrite); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

func imageNames(configured string, legacy string, defaults []string, item media.Item) []string {
	values := splitConfiguredNames(configured)
	if len(values) == 0 && strings.TrimSpace(legacy) != "" {
		values = []string{legacy}
	}
	if len(values) == 0 {
		values = defaults
	}
	fileBase := strings.TrimSuffix(item.FileName, filepath.Ext(item.FileName))
	seen := map[string]bool{}
	var names []string
	for _, value := range values {
		name := strings.TrimSpace(value)
		if name == "" {
			continue
		}
		name = strings.ReplaceAll(name, "{filename}", fileBase)
		name = strings.ReplaceAll(name, "{moviefilename}", fileBase)
		name = strings.ReplaceAll(name, "{movieFilename}", fileBase)
		if filepath.Ext(name) == "" {
			name += ".jpg"
		}
		name = filepath.Base(name)
		if name == "." || seen[name] {
			continue
		}
		seen[name] = true
		names = append(names, name)
	}
	if len(names) == 0 {
		return defaults
	}
	return names
}

func splitConfiguredNames(value string) []string {
	return strings.FieldsFunc(value, func(r rune) bool {
		return r == '\n' || r == '\r' || r == ',' || r == ';'
	})
}

func defaultMoviePosterNames() []string {
	return []string{"poster.jpg", "folder.jpg", "{filename}-poster.jpg"}
}

func defaultMovieFanartNames() []string {
	return []string{"fanart.jpg", "{filename}-fanart.jpg"}
}

func defaultTVShowPosterNames() []string {
	return []string{"poster.jpg", "folder.jpg"}
}

func defaultTVShowFanartNames() []string {
	return []string{"fanart.jpg", "backdrop.jpg"}
}

func (s *Server) runningScanTaskLocked(libraryID string) *Task {
	for _, task := range s.tasks {
		if task.Type == "scan" && task.LibraryID == libraryID && task.State == "running" {
			return task
		}
	}
	return nil
}

func (s *Server) runScanTask(taskID string, library media.Library) {
	var pending []media.Item
	var persistErr error
	lastFlush := time.Now()
	flushPending := func(force bool) error {
		if len(pending) == 0 {
			return nil
		}
		if !force && len(pending) < scanPersistBatchSize && time.Since(lastFlush) < scanPersistFlushInterval {
			return nil
		}
		batch := append([]media.Item(nil), pending...)
		if err := s.store.SaveItems(batch); err != nil {
			pending = nil
			return err
		}
		pending = pending[:0]
		lastFlush = time.Now()
		return nil
	}

	shouldCancel := func() bool {
		s.mu.Lock()
		defer s.mu.Unlock()
		task := s.tasks[taskID]
		return task != nil && task.State == "canceling"
	}
	items, err := media.ScanLibraryWithCancel(library, func(progress media.ScanProgress) {
		s.mu.Lock()
		if task := s.tasks[taskID]; task != nil {
			task.SourcePath = progress.SourcePath
			task.CurrentPath = progress.CurrentPath
			task.VisitedFiles = progress.VisitedFiles
			task.FoundItems = progress.FoundItems
		}
		if progress.Item != nil {
			s.items[progress.Item.ID] = *progress.Item
			if persistErr == nil {
				pending = append(pending, *progress.Item)
			}
		}
		s.mu.Unlock()
		if progress.Item != nil && persistErr == nil {
			persistErr = flushPending(false)
		}
	}, shouldCancel)
	if err == nil && persistErr != nil {
		err = persistErr
	}
	if err == nil {
		err = flushPending(true)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	task := s.tasks[taskID]
	if task == nil {
		return
	}
	task.FinishedAt = time.Now().Format(time.RFC3339)
	if err == media.ErrScanCanceled {
		task.State = "canceled"
		task.Error = ""
		task.ResultCount = task.FoundItems
		_ = s.store.SaveTask(task.toRecord())
		return
	}
	if err != nil {
		task.State = "failed"
		task.Error = err.Error()
		_ = s.store.SaveTask(task.toRecord())
		return
	}
	for _, item := range items {
		s.items[item.ID] = item
	}
	if err := s.store.PruneLibraryItems(library.ID, items); err != nil {
		task.State = "failed"
		task.Error = err.Error()
		_ = s.store.SaveTask(task.toRecord())
		return
	}
	currentIDs := make(map[string]struct{}, len(items))
	for _, item := range items {
		currentIDs[item.ID] = struct{}{}
	}
	for id, item := range s.items {
		if item.LibraryID == library.ID {
			if _, ok := currentIDs[id]; !ok {
				delete(s.items, id)
			}
		}
	}
	task.State = "completed"
	task.ResultCount = len(items)
	task.FoundItems = len(items)
	_ = s.store.SaveTask(task.toRecord())
}

func (s *Server) loadLibraries() error {
	libraries, err := s.store.Libraries()
	if err != nil {
		return err
	}
	s.libraries = libraries
	for i := range s.libraries {
		s.libraries[i].Paths = media.NormalizePaths(s.libraries[i])
		if len(s.libraries[i].Paths) > 0 {
			s.libraries[i].Path = s.libraries[i].Paths[0]
		}
	}
	return nil
}

func (s *Server) loadItems() error {
	items, err := s.store.Items()
	if err != nil {
		return err
	}
	for _, item := range items {
		s.items[item.ID] = item
	}
	return nil
}

func (s *Server) loadTasks() error {
	tasks, err := s.store.Tasks()
	if err != nil {
		return err
	}
	for _, record := range tasks {
		task := taskFromRecord(record)
		if task.State == "running" {
			task.State = "failed"
			task.Error = "task interrupted by server restart"
			task.FinishedAt = time.Now().Format(time.RFC3339)
			_ = s.store.SaveTask(task.toRecord())
		}
		s.tasks[task.ID] = task
	}
	return nil
}

func (s *Server) settingsResponseLocked() SettingsResponse {
	source := "none"
	if strings.TrimSpace(s.settings.TMDBAPIKey) != "" {
		source = "settings"
	}
	if strings.TrimSpace(s.config.TMDBKey) != "" {
		source = "environment"
	}
	return SettingsResponse{
		TMDBConfigured:          s.tmdb.Enabled(),
		TMDBEnabled:             s.tmdb.Enabled(),
		TMDBKeySource:           source,
		ProxyEnabled:            s.settings.ProxyEnabled,
		ProxyHost:               s.settings.ProxyHost,
		ProxyPort:               s.settings.ProxyPort,
		ProxyUsername:           s.settings.ProxyUsername,
		ProxyPassword:           s.settings.ProxyPassword != "",
		MovieScrapeMetadata:     defaultBool(s.settings.MovieScrapeMetadata, true),
		MovieScrapeNFO:          defaultBool(s.settings.MovieScrapeNFO, true),
		MovieScrapeImages:       defaultBool(s.settings.MovieScrapeImages, true),
		MovieScrapeOverwrite:    defaultBool(s.settings.MovieScrapeOverwrite, false),
		TVShowScrapeMetadata:    defaultBool(s.settings.TVShowScrapeMetadata, true),
		TVShowEpisodeMetadata:   defaultBool(s.settings.TVShowEpisodeMetadata, true),
		TVShowScrapeNFO:         defaultBool(s.settings.TVShowScrapeNFO, true),
		TVShowScrapeImages:      defaultBool(s.settings.TVShowScrapeImages, true),
		TVShowScrapeOverwrite:   defaultBool(s.settings.TVShowScrapeOverwrite, false),
		MovieRenamerPathname:    defaultString(s.settings.MovieRenamerPathname, "{title} ({year})"),
		MovieRenamerFilename:    defaultString(s.settings.MovieRenamerFilename, "{title} ({year})"),
		TVShowRenamerShowFolder: defaultString(s.settings.TVShowRenamerShowFolder, "{showTitle}"),
		TVShowRenamerSeason:     defaultString(s.settings.TVShowRenamerSeason, "Season {seasonNr2}"),
		TVShowRenamerFilename:   defaultString(s.settings.TVShowRenamerFilename, "{showTitle} - S{seasonNr2}E{episodeNr2} - {title}"),
		MoviePosterName:         defaultString(s.settings.MoviePosterName, "poster.jpg"),
		MovieFanartName:         defaultString(s.settings.MovieFanartName, "fanart.jpg"),
		MoviePosterNames:        defaultImageNames(s.settings.MoviePosterNames, s.settings.MoviePosterName, defaultMoviePosterNames()),
		MovieFanartNames:        defaultImageNames(s.settings.MovieFanartNames, s.settings.MovieFanartName, defaultMovieFanartNames()),
		TVShowPosterName:        defaultString(s.settings.TVShowPosterName, "poster.jpg"),
		TVShowFanartName:        defaultString(s.settings.TVShowFanartName, "fanart.jpg"),
		TVShowPosterNames:       defaultImageNames(s.settings.TVShowPosterNames, s.settings.TVShowPosterName, defaultTVShowPosterNames()),
		TVShowFanartNames:       defaultImageNames(s.settings.TVShowFanartNames, s.settings.TVShowFanartName, defaultTVShowFanartNames()),
	}
}

func (s *Server) appSettings() AppSettings {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.settings
}

func defaultBool(value *bool, fallback bool) bool {
	if value == nil {
		return fallback
	}
	return *value
}

func defaultString(value string, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func defaultImageNames(value string, legacy string, fallback []string) string {
	value = strings.TrimSpace(value)
	if value != "" {
		if len(fallback) > 1 && value == fallback[0] {
			return strings.Join(fallback, "\n")
		}
		return value
	}
	legacy = strings.TrimSpace(legacy)
	if legacy != "" && (len(fallback) == 0 || legacy != fallback[0]) {
		return legacy
	}
	return strings.Join(fallback, "\n")
}

func (s *Server) tmdbClient() tmdb.Client {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.tmdb
}

func loadAppSettings(dataDir string) (AppSettings, error) {
	settings, err := readJSONFile[AppSettings](settingsPath(dataDir))
	if err != nil {
		if os.IsNotExist(err) {
			return AppSettings{}, nil
		}
		return AppSettings{}, err
	}
	settings.TMDBAPIKey = strings.TrimSpace(settings.TMDBAPIKey)
	settings.ProxyHost = strings.TrimSpace(settings.ProxyHost)
	settings.ProxyUsername = strings.TrimSpace(settings.ProxyUsername)
	return settings, nil
}

func saveAppSettings(dataDir string, settings AppSettings) error {
	settings.TMDBAPIKey = strings.TrimSpace(settings.TMDBAPIKey)
	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(settingsPath(dataDir), data, 0600)
}

func settingsPath(dataDir string) string {
	return filepath.Join(dataDir, "settings.json")
}

func newTMDBClient(config Config, settings AppSettings) (tmdb.Client, error) {
	tmdbKey := strings.TrimSpace(config.TMDBKey)
	if tmdbKey == "" {
		tmdbKey = strings.TrimSpace(settings.TMDBAPIKey)
	}
	client := config.Client
	if settings.ProxyEnabled {
		proxyURL, err := buildProxyURL(settings)
		if err != nil {
			return tmdb.Client{}, err
		}
		client = &http.Client{
			Timeout: clientTimeout(config.Client),
			Transport: &http.Transport{
				Proxy: http.ProxyURL(proxyURL),
			},
		}
	}
	return tmdb.Client{Key: tmdbKey, HTTP: client, Lang: "zh-CN"}, nil
}

func buildProxyURL(settings AppSettings) (*url.URL, error) {
	host := strings.TrimSpace(settings.ProxyHost)
	if host == "" {
		return nil, fmt.Errorf("proxy host is required")
	}
	port := settings.ProxyPort
	if port < 0 || port > 65535 {
		return nil, fmt.Errorf("proxy port must be between 1 and 65535")
	}
	if port == 0 {
		port = 80
	}
	proxyURL := &url.URL{
		Scheme: "http",
		Host:   net.JoinHostPort(host, strconv.Itoa(port)),
	}
	if settings.ProxyUsername != "" {
		if settings.ProxyPassword != "" {
			proxyURL.User = url.UserPassword(settings.ProxyUsername, settings.ProxyPassword)
		} else {
			proxyURL.User = url.User(settings.ProxyUsername)
		}
	}
	return proxyURL, nil
}

func clientTimeout(client *http.Client) time.Duration {
	if client != nil && client.Timeout > 0 {
		return client.Timeout
	}
	return 20 * time.Second
}

func (s *Server) migrateJSON() error {
	if len(s.libraries) == 0 {
		if libraries, err := readJSONFile[[]media.Library](filepath.Join(s.config.DataDir, "libraries.json")); err == nil {
			for _, library := range libraries {
				library.Paths = media.NormalizePaths(library)
				if len(library.Paths) > 0 {
					library.Path = library.Paths[0]
				}
				if err := s.store.SaveLibrary(library); err != nil {
					return err
				}
				s.libraries = append(s.libraries, library)
			}
		}
	}
	if len(s.items) == 0 {
		if items, err := readJSONFile[[]media.Item](filepath.Join(s.config.DataDir, "items.json")); err == nil {
			if err := s.store.SaveItems(items); err != nil {
				return err
			}
			for _, item := range items {
				s.items[item.ID] = item
			}
		}
	}
	return nil
}

func (t *Task) toRecord() store.TaskRecord {
	return store.TaskRecord{
		ID:           t.ID,
		Type:         t.Type,
		LibraryID:    t.LibraryID,
		LibraryName:  t.LibraryName,
		State:        t.State,
		SourcePath:   t.SourcePath,
		CurrentPath:  t.CurrentPath,
		VisitedFiles: t.VisitedFiles,
		FoundItems:   t.FoundItems,
		ResultCount:  t.ResultCount,
		Error:        t.Error,
		StartedAt:    t.StartedAt,
		FinishedAt:   t.FinishedAt,
	}
}

func taskFromRecord(record store.TaskRecord) *Task {
	return &Task{
		ID:           record.ID,
		Type:         record.Type,
		LibraryID:    record.LibraryID,
		LibraryName:  record.LibraryName,
		State:        record.State,
		SourcePath:   record.SourcePath,
		CurrentPath:  record.CurrentPath,
		VisitedFiles: record.VisitedFiles,
		FoundItems:   record.FoundItems,
		ResultCount:  record.ResultCount,
		Error:        record.Error,
		StartedAt:    record.StartedAt,
		FinishedAt:   record.FinishedAt,
	}
}

func readJSONFile[T any](path string) (T, error) {
	var value T
	data, err := os.ReadFile(path)
	if err != nil {
		return value, err
	}
	err = json.Unmarshal(data, &value)
	return value, err
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
