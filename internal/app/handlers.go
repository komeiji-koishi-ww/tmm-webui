package app

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"tmmweb/internal/media"
	"tmmweb/internal/tmdb"
)

// HTTP handlers intentionally stay thin: validate request input, call the
// domain helpers, persist changed records, and return JSON.
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
		if input.MovieRenameAfterScrape != nil {
			settings.MovieRenameAfterScrape = input.MovieRenameAfterScrape
		}
		if input.TVShowRenameAfterScrape != nil {
			settings.TVShowRenameAfterScrape = input.TVShowRenameAfterScrape
		}
		settings.MovieScraperFields = normalizeScraperFields(input.MovieScraperFields, defaultMovieScraperFields())
		settings.TVShowScraperFields = normalizeScraperFields(input.TVShowScraperFields, defaultTVShowScraperFields())
		settings.TVEpisodeScraperFields = normalizeScraperFields(input.TVEpisodeScraperFields, defaultTVEpisodeScraperFields())
		settings.MovieRenamerPathname = strings.TrimSpace(input.MovieRenamerPathname)
		settings.MovieRenamerFilename = strings.TrimSpace(input.MovieRenamerFilename)
		if input.MovieRenamerPathSpaceSubstitution != nil {
			settings.MovieRenamerPathSpaceSubstitution = input.MovieRenamerPathSpaceSubstitution
		}
		settings.MovieRenamerPathSpaceReplacement = normalizeReplacement(input.MovieRenamerPathSpaceReplacement, "_", []string{"_", ".", "-"})
		if input.MovieRenamerFilenameSpaceSubstitution != nil {
			settings.MovieRenamerFilenameSpaceSubstitution = input.MovieRenamerFilenameSpaceSubstitution
		}
		settings.MovieRenamerFilenameSpaceReplacement = normalizeReplacement(input.MovieRenamerFilenameSpaceReplacement, "_", []string{"_", ".", "-"})
		settings.MovieRenamerColonReplacement = normalizeReplacement(input.MovieRenamerColonReplacement, "-", []string{" ", "-", "_", ""})
		if input.MovieRenamerAsciiReplacement != nil {
			settings.MovieRenamerAsciiReplacement = input.MovieRenamerAsciiReplacement
		}
		settings.MovieRenamerFirstCharacterReplacement = defaultString(strings.TrimSpace(input.MovieRenamerFirstCharacterReplacement), "#")
		if input.MovieRenamerCreateSingleMovieSet != nil {
			settings.MovieRenamerCreateSingleMovieSet = input.MovieRenamerCreateSingleMovieSet
		}
		if input.MovieRenamerNFOCleanup != nil {
			settings.MovieRenamerNFOCleanup = input.MovieRenamerNFOCleanup
		}
		if input.MovieRenamerCleanupUnwanted != nil {
			settings.MovieRenamerCleanupUnwanted = input.MovieRenamerCleanupUnwanted
		}
		if input.MovieRenamerAllowMerge != nil {
			settings.MovieRenamerAllowMerge = input.MovieRenamerAllowMerge
		}
		settings.TVShowRenamerShowFolder = strings.TrimSpace(input.TVShowRenamerShowFolder)
		settings.TVShowRenamerSeason = strings.TrimSpace(input.TVShowRenamerSeason)
		settings.TVShowRenamerFilename = strings.TrimSpace(input.TVShowRenamerFilename)
		if input.TVShowRenamerShowFolderSpaceSubstitution != nil {
			settings.TVShowRenamerShowFolderSpaceSubstitution = input.TVShowRenamerShowFolderSpaceSubstitution
		}
		settings.TVShowRenamerShowFolderSpaceReplacement = normalizeReplacement(input.TVShowRenamerShowFolderSpaceReplacement, "_", []string{"_", ".", "-"})
		if input.TVShowRenamerSeasonFolderSpaceSubstitution != nil {
			settings.TVShowRenamerSeasonFolderSpaceSubstitution = input.TVShowRenamerSeasonFolderSpaceSubstitution
		}
		settings.TVShowRenamerSeasonFolderSpaceReplacement = normalizeReplacement(input.TVShowRenamerSeasonFolderSpaceReplacement, "_", []string{"_", ".", "-"})
		if input.TVShowRenamerFilenameSpaceSubstitution != nil {
			settings.TVShowRenamerFilenameSpaceSubstitution = input.TVShowRenamerFilenameSpaceSubstitution
		}
		settings.TVShowRenamerFilenameSpaceReplacement = normalizeReplacement(input.TVShowRenamerFilenameSpaceReplacement, "_", []string{"_", ".", "-"})
		settings.TVShowRenamerColonReplacement = normalizeReplacement(input.TVShowRenamerColonReplacement, " ", []string{" ", "-", "_"})
		if input.TVShowRenamerAsciiReplacement != nil {
			settings.TVShowRenamerAsciiReplacement = input.TVShowRenamerAsciiReplacement
		}
		settings.TVShowRenamerFirstCharacterReplacement = defaultString(strings.TrimSpace(input.TVShowRenamerFirstCharacterReplacement), "#")
		if input.TVShowRenamerCleanupUnwanted != nil {
			settings.TVShowRenamerCleanupUnwanted = input.TVShowRenamerCleanupUnwanted
		}
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
		libraries := s.libraries
		if libraries == nil {
			libraries = []media.Library{}
		}
		writeJSON(w, libraries)
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
	full := strings.EqualFold(strings.TrimSpace(r.URL.Query().Get("full")), "1") ||
		strings.EqualFold(strings.TrimSpace(r.URL.Query().Get("full")), "true")
	w.Header().Set("Content-Type", "application/json")
	_, _ = io.WriteString(w, `{"items":[`)
	s.mu.Lock()
	count := 0
	for _, item := range s.items {
		if libraryID == "" || item.LibraryID == libraryID {
			if count > 0 {
				_, _ = io.WriteString(w, ",")
			}
			var value interface{} = item
			if !full {
				value = itemListEntryFromMedia(item)
			}
			data, err := json.Marshal(value)
			if err == nil {
				_, _ = w.Write(data)
				count++
			}
		}
	}
	s.mu.Unlock()
	_, _ = fmt.Fprintf(w, `],"count":%d}`+"\n", count)
}

type itemListEntry struct {
	ID            string   `json:"id"`
	LibraryID     string   `json:"libraryId"`
	SourcePath    string   `json:"sourcePath"`
	Kind          string   `json:"kind"`
	Path          string   `json:"path"`
	Dir           string   `json:"dir"`
	FileName      string   `json:"fileName"`
	TitleGuess    string   `json:"titleGuess"`
	YearGuess     string   `json:"yearGuess,omitempty"`
	Original      string   `json:"originalTitle,omitempty"`
	Overview      string   `json:"overview,omitempty"`
	Runtime       int      `json:"runtime,omitempty"`
	Rating        float64  `json:"rating,omitempty"`
	ShowRating    float64  `json:"showRating,omitempty"`
	Genres        []string `json:"genres,omitempty"`
	Actors        []string `json:"actors,omitempty"`
	Premiered     string   `json:"premiered,omitempty"`
	DateAdded     string   `json:"dateAdded,omitempty"`
	FileSize      string   `json:"fileSize,omitempty"`
	FileSizeBytes int64    `json:"fileSizeBytes,omitempty"`
	VideoFormat   string   `json:"videoFormat,omitempty"`
	AudioCodec    string   `json:"audioCodec,omitempty"`
	IMDBID        string   `json:"imdbId,omitempty"`
	ShowGuess     string   `json:"showGuess,omitempty"`
	Season        int      `json:"season,omitempty"`
	Episode       int      `json:"episode,omitempty"`
	Episodes      []int    `json:"episodes,omitempty"`
	AirDate       string   `json:"airDate,omitempty"`
	MediaType     string   `json:"mediaType"`
	HasNFO        bool     `json:"hasNfo"`
	HasPoster     bool     `json:"hasPoster"`
	HasFanart     bool     `json:"hasFanart"`
	HasSubtitle   bool     `json:"hasSubtitle"`
	MatchedID     int      `json:"matchedId,omitempty"`
	MatchedName   string   `json:"matchedName,omitempty"`
}

func itemListEntryFromMedia(item media.Item) itemListEntry {
	return itemListEntry{
		ID:            item.ID,
		LibraryID:     item.LibraryID,
		SourcePath:    item.SourcePath,
		Kind:          item.Kind,
		Path:          item.Path,
		Dir:           item.Dir,
		FileName:      item.FileName,
		TitleGuess:    item.TitleGuess,
		YearGuess:     item.YearGuess,
		Original:      item.Original,
		Overview:      item.Overview,
		Runtime:       item.Runtime,
		Rating:        item.Rating,
		ShowRating:    item.ShowRating,
		Genres:        limitStrings(item.Genres, 24),
		Actors:        limitStrings(item.Actors, 40),
		Premiered:     item.Premiered,
		DateAdded:     item.DateAdded,
		FileSize:      item.FileSize,
		FileSizeBytes: item.FileSizeBytes,
		VideoFormat:   item.VideoFormat,
		AudioCodec:    item.AudioCodec,
		IMDBID:        item.IMDBID,
		ShowGuess:     item.ShowGuess,
		Season:        item.Season,
		Episode:       item.Episode,
		Episodes:      limitInts(item.Episodes, 12),
		AirDate:       item.AirDate,
		MediaType:     item.MediaType,
		HasNFO:        item.HasNFO,
		HasPoster:     item.HasPoster,
		HasFanart:     item.HasFanart,
		HasSubtitle:   item.HasSubtitle,
		MatchedID:     item.MatchedID,
		MatchedName:   item.MatchedName,
	}
}

func limitStrings(values []string, limit int) []string {
	if len(values) <= limit {
		return values
	}
	return values[:limit]
}

func limitInts(values []int, limit int) []int {
	if len(values) <= limit {
		return values
	}
	return values[:limit]
}

func (s *Server) handleArtwork(w http.ResponseWriter, r *http.Request) {
	itemID := r.URL.Query().Get("id")
	artType := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("type")))
	scope := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("scope")))
	if itemID == "" || (artType != "poster" && artType != "fanart") {
		http.Error(w, "id and type=poster|fanart are required", http.StatusBadRequest)
		return
	}
	item, ok := s.findItem(itemID)
	if !ok {
		http.Error(w, "item not found", http.StatusNotFound)
		return
	}
	path := artworkPath(item, artType, scope)
	if path == "" {
		http.NotFound(w, r)
		return
	}
	http.ServeFile(w, r, path)
}

func (s *Server) handleTMDBImage(w http.ResponseWriter, r *http.Request) {
	imagePath := strings.TrimSpace(r.URL.Query().Get("path"))
	size := strings.TrimSpace(r.URL.Query().Get("size"))
	if imagePath == "" {
		http.Error(w, "path is required", http.StatusBadRequest)
		return
	}
	allowedSizes := map[string]bool{
		"w92": true, "w154": true, "w185": true, "w300": true, "w342": true,
		"w500": true, "w780": true, "w1280": true, "original": true,
	}
	if size == "" {
		size = "w342"
	}
	if !allowedSizes[size] {
		http.Error(w, "invalid image size", http.StatusBadRequest)
		return
	}
	data, err := s.tmdbClient().DownloadImageSized(imagePath, size)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	w.Header().Set("Cache-Control", "public, max-age=86400")
	w.Header().Set("Content-Type", http.DetectContentType(data))
	_, _ = w.Write(data)
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
		ItemID            int      `json:"-"`
		Item              string   `json:"itemId"`
		Scope             string   `json:"scope"`
		LibraryID         string   `json:"libraryId"`
		ShowName          string   `json:"showName"`
		Season            int      `json:"season"`
		TMDBID            int      `json:"tmdbId"`
		MediaType         string   `json:"mediaType"`
		WriteNFO          bool     `json:"writeNfo"`
		WriteImages       bool     `json:"writeImages"`
		WriteMeta         bool     `json:"writeMeta"`
		WriteShowMeta     *bool    `json:"writeShowMeta"`
		WriteEpisodeMeta  *bool    `json:"writeEpisodeMeta"`
		Overwrite         bool     `json:"overwrite"`
		RenameAfterScrape *bool    `json:"renameAfterScrape"`
		MetadataFields    []string `json:"metadataFields"`
		EpisodeFields     []string `json:"episodeFields"`
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
		metadataFields := normalizeScraperFields(input.MetadataFields, normalizeScraperFields(settings.TVShowScraperFields, defaultTVShowScraperFields()))
		episodeFields := normalizeScraperFields(input.EpisodeFields, normalizeScraperFields(settings.TVEpisodeScraperFields, defaultTVEpisodeScraperFields()))
		writeShowMeta := defaultBool(input.WriteShowMeta, input.WriteMeta)
		writeEpisodeMeta := defaultBool(input.WriteEpisodeMeta, input.WriteMeta && defaultBool(settings.TVShowEpisodeMetadata, true))
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
		showDir := tvShowRootDir(item)
		var seasonMetadata tmdb.TVSeason
		var seasonMetadataLoaded bool
		seasonMetadataByNumber := map[int]tmdb.TVSeason{}
		seasonPosterAvailable := map[int]bool{}
		seasonFanartAvailable := map[int]bool{}
		episodeThumbAvailable := map[string]bool{}
		loadSeasonMetadataFor := func(seasonNumber int) (tmdb.TVSeason, error) {
			if seasonNumber <= 0 {
				return tmdb.TVSeason{}, nil
			}
			if cached, ok := seasonMetadataByNumber[seasonNumber]; ok {
				return cached, nil
			}
			loaded, err := client.TVSeason(show.ID, seasonNumber, show.Title)
			if err != nil {
				return tmdb.TVSeason{}, err
			}
			seasonMetadataByNumber[seasonNumber] = loaded
			if seasonNumber == season {
				seasonMetadata = loaded
				seasonMetadataLoaded = true
			}
			return loaded, nil
		}
		loadSeasonMetadata := func() (tmdb.TVSeason, error) {
			if seasonMetadataLoaded {
				return seasonMetadata, nil
			}
			loaded, err := loadSeasonMetadataFor(season)
			if err != nil {
				return tmdb.TVSeason{}, err
			}
			seasonMetadata = loaded
			seasonMetadataLoaded = true
			return seasonMetadata, nil
		}
		if input.WriteNFO && writeShowMeta {
			if scope == "season" && season > 0 {
				seasonMetadata, err := loadSeasonMetadata()
				if err != nil {
					http.Error(w, err.Error(), 500)
					return
				}
				if err := writeTVSeasonNFO(seasonNFOPath(item, season), seasonMetadata, show.BackdropPath, input.Overwrite); err != nil {
					http.Error(w, err.Error(), 500)
					return
				}
			} else {
				if err := writeTVShowNFO(showDir, show, input.Overwrite); err != nil {
					http.Error(w, err.Error(), 500)
					return
				}
			}
		}
		if input.WriteImages {
			if show.PosterPath != "" && scraperFieldEnabled(metadataFields, "POSTER") {
				if data, err := client.DownloadImage(show.PosterPath); err == nil {
					_ = writeImages(showDir, imageNames(settings.TVShowPosterNames, settings.TVShowPosterName, defaultTVShowPosterNames(), item), data, input.Overwrite)
				}
			}
			if show.BackdropPath != "" && scraperFieldEnabled(metadataFields, "FANART") {
				if data, err := client.DownloadImage(show.BackdropPath); err == nil {
					_ = writeImages(showDir, imageNames(settings.TVShowFanartNames, settings.TVShowFanartName, defaultTVShowFanartNames(), item), data, input.Overwrite)
				}
			}
			if show.TVDBID > 0 && scraperFieldEnabled(metadataFields, "THEME") {
				_ = downloadTVTheme(showDir, show.TVDBID, input.Overwrite)
			}
		}
		updated := []media.Item{item}
		if scope == "show" || scope == "season" {
			updated = s.matchTVItems(libraryID, showName, season, scope, show)
		}
		if len(updated) == 0 {
			updated = append(updated, item)
		}
		if input.WriteImages {
			seasonsToFetch := map[int]bool{}
			for _, changed := range updated {
				if changed.Season > 0 {
					seasonsToFetch[changed.Season] = true
				}
			}
			for seasonNumber := range seasonsToFetch {
				seasonData, err := loadSeasonMetadataFor(seasonNumber)
				if err != nil {
					continue
				}
				if seasonData.PosterPath != "" && scraperFieldEnabled(metadataFields, "SEASON_POSTER") {
					if data, err := client.DownloadImage(seasonData.PosterPath); err == nil {
						seasonPosterAvailable[seasonNumber] = true
						_ = writeImages(showDir, seasonPosterNames(seasonNumber), data, input.Overwrite)
					}
				}
				if show.BackdropPath != "" && scraperFieldEnabled(metadataFields, "SEASON_FANART") {
					if data, err := client.DownloadImage(show.BackdropPath); err == nil {
						seasonFanartAvailable[seasonNumber] = true
						_ = writeImages(showDir, seasonFanartNames(seasonNumber), data, input.Overwrite)
					}
				}
			}
			if writeEpisodeMeta && scraperFieldEnabled(episodeFields, "THUMB") {
				for _, changed := range updated {
					if changed.Season <= 0 {
						continue
					}
					seasonData, err := loadSeasonMetadataFor(changed.Season)
					if err != nil {
						continue
					}
					episode, ok := tvEpisodeMetadataForItem(seasonData, changed)
					if !ok || episode.StillPath == "" {
						continue
					}
					if data, err := client.DownloadImage(episode.StillPath); err == nil {
						episodeThumbAvailable[changed.ID] = true
						_ = writeImages(changed.Dir, episodeThumbNames(changed), data, input.Overwrite)
					}
				}
			}
		}
		if input.WriteNFO && writeEpisodeMeta {
			for i := range updated {
				if updated[i].Season <= 0 {
					continue
				}
				seasonData, err := loadSeasonMetadataFor(updated[i].Season)
				if err != nil {
					continue
				}
				episode, ok := tvEpisodeMetadataForItem(seasonData, updated[i])
				if !ok {
					continue
				}
				if err := writeTVEpisodeNFO(episodeNFOPath(updated[i]), show, episode, updated[i], input.Overwrite); err == nil {
					updated[i].HasNFO = true
				}
			}
		}
		renameWarnings := []string{}
		renamePreviews := []media.RenamePreview{}
		oldIDs := []string{}
		for i := range updated {
			if writeShowMeta {
				if scraperFieldEnabled(metadataFields, "ID") {
					updated[i].MatchedID = show.ID
				}
				if scraperFieldEnabled(metadataFields, "TITLE") {
					updated[i].MatchedName = show.Title
					updated[i].ShowGuess = firstNonEmpty(show.Title, updated[i].ShowGuess)
					updated[i].TitleGuess = firstNonEmpty(updated[i].ShowGuess, updated[i].TitleGuess)
				}
				if scraperFieldEnabled(metadataFields, "ORIGINAL_TITLE") {
					updated[i].Original = show.Original
				}
				if scraperFieldEnabled(metadataFields, "PLOT") {
					updated[i].Overview = show.Overview
				}
				if scraperFieldEnabled(metadataFields, "RATING") {
					updated[i].ShowRating = show.VoteAverage
				}
				if scraperFieldEnabled(metadataFields, "GENRES") {
					updated[i].Genres = show.Genres
				}
				if scraperFieldEnabled(metadataFields, "AIRED") {
					updated[i].Premiered = show.FirstAirDate
				}
				if scraperFieldEnabled(metadataFields, "YEAR") {
					updated[i].YearGuess = firstNonEmpty(yearFromDate(show.FirstAirDate), updated[i].YearGuess)
				}
			}
			if writeEpisodeMeta && updated[i].Season > 0 {
				if seasonData, err := loadSeasonMetadataFor(updated[i].Season); err == nil {
					if episode, ok := tvEpisodeMetadataForItem(seasonData, updated[i]); ok {
						if scraperFieldEnabled(episodeFields, "TITLE") && strings.TrimSpace(episode.Title) != "" {
							updated[i].MatchedName = episode.Title
						}
						if scraperFieldEnabled(episodeFields, "PLOT") {
							updated[i].Overview = firstNonEmpty(episode.Overview, updated[i].Overview)
						}
						if scraperFieldEnabled(episodeFields, "AIRED") {
							updated[i].AirDate = firstNonEmpty(episode.AirDate, updated[i].AirDate)
						}
						if scraperFieldEnabled(episodeFields, "RATING") {
							updated[i].Rating = episode.VoteAverage
						}
					}
				}
			}
			if input.WriteImages {
				updated[i].HasPoster = updated[i].HasPoster ||
					(show.PosterPath != "" && scraperFieldEnabled(metadataFields, "POSTER")) ||
					seasonPosterAvailable[updated[i].Season] ||
					episodeThumbAvailable[updated[i].ID]
				updated[i].HasFanart = updated[i].HasFanart ||
					(show.BackdropPath != "" && scraperFieldEnabled(metadataFields, "FANART")) ||
					seasonFanartAvailable[updated[i].Season]
			}
			if input.WriteNFO && scope == "season" {
				updated[i].HasNFO = true
			}
		}
		if defaultBool(input.RenameAfterScrape, defaultBool(settings.TVShowRenameAfterScrape, true)) {
			for i := range updated {
				episodeTitle := tvEpisodeRenameTitle(updated[i], show.Title, loadSeasonMetadataFor)
				preview := media.BuildTVShowRenameWithOptions(
					updated[i],
					firstNonEmpty(show.Title, updated[i].ShowGuess),
					episodeTitle,
					yearFromDate(show.FirstAirDate),
					show.ID,
					settings.TVShowRenamerShowFolder,
					settings.TVShowRenamerSeason,
					settings.TVShowRenamerFilename,
					tvShowFolderRenameOptions(settings),
					tvSeasonFolderRenameOptions(settings),
					tvEpisodeFileRenameOptions(settings),
				)
				renamePreviews = append(renamePreviews, preview)
				oldID := updated[i].ID
				if err := media.ApplyRename(preview); err != nil {
					renameWarnings = append(renameWarnings, err.Error())
					continue
				}
				updated[i] = media.RefreshItemPath(updated[i], preview.TargetFile)
				if updated[i].ID != oldID {
					oldIDs = append(oldIDs, oldID)
				}
			}
		}
		s.mu.Lock()
		for _, oldID := range oldIDs {
			delete(s.items, oldID)
		}
		for _, changed := range updated {
			s.items[changed.ID] = changed
		}
		s.mu.Unlock()
		if err := s.store.SaveItems(updated); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		for _, oldID := range oldIDs {
			if err := s.store.DeleteItem(oldID); err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
		}
		response := map[string]interface{}{"item": updated[0], "items": updated, "show": show, "scope": scope}
		if len(renamePreviews) > 0 {
			response["renamePreviews"] = renamePreviews
			response["renamed"] = len(renameWarnings) == 0
		}
		if len(renameWarnings) > 0 {
			response["renameWarnings"] = renameWarnings
		}
		if seasonMetadataLoaded {
			response["season"] = seasonMetadata
		}
		writeJSON(w, response)
		return
	}
	metadataFields := normalizeScraperFields(input.MetadataFields, normalizeScraperFields(settings.MovieScraperFields, defaultMovieScraperFields()))
	movie, err := client.Movie(input.TMDBID)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	if input.WriteNFO {
		if err := writeMovieNFO(item.Dir, movie, item, input.Overwrite); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
	}
	if input.WriteImages {
		if movie.PosterPath != "" && scraperFieldEnabled(metadataFields, "POSTER") {
			if data, err := client.DownloadImage(movie.PosterPath); err == nil {
				_ = writeImages(item.Dir, imageNames(settings.MoviePosterNames, settings.MoviePosterName, defaultMoviePosterNames(), item), data, input.Overwrite)
			}
		}
		if movie.BackdropPath != "" && scraperFieldEnabled(metadataFields, "FANART") {
			if data, err := client.DownloadImage(movie.BackdropPath); err == nil {
				_ = writeImages(item.Dir, imageNames(settings.MovieFanartNames, settings.MovieFanartName, defaultMovieFanartNames(), item), data, input.Overwrite)
			}
		}
	}
	if input.WriteMeta {
		if scraperFieldEnabled(metadataFields, "ID") {
			item.MatchedID = movie.ID
			item.IMDBID = movie.ImdbID
		}
		if scraperFieldEnabled(metadataFields, "TITLE") {
			item.MatchedName = movie.Title
			item.TitleGuess = firstNonEmpty(movie.Title, item.TitleGuess)
		}
		if scraperFieldEnabled(metadataFields, "ORIGINAL_TITLE") {
			item.Original = movie.Original
		}
		if scraperFieldEnabled(metadataFields, "PLOT") {
			item.Overview = movie.Overview
		}
		if scraperFieldEnabled(metadataFields, "RUNTIME") {
			item.Runtime = movie.Runtime
		}
		if scraperFieldEnabled(metadataFields, "RATING") {
			item.Rating = movie.VoteAverage
		}
		if scraperFieldEnabled(metadataFields, "GENRES") {
			item.Genres = movie.Genres
		}
		if scraperFieldEnabled(metadataFields, "RELEASE_DATE") {
			item.Premiered = movie.ReleaseDate
		}
		if scraperFieldEnabled(metadataFields, "YEAR") {
			item.YearGuess = firstNonEmpty(yearFromDate(movie.ReleaseDate), item.YearGuess)
		}
	}
	if input.WriteNFO {
		item.HasNFO = true
	}
	if input.WriteImages {
		item.HasPoster = item.HasPoster || (movie.PosterPath != "" && scraperFieldEnabled(metadataFields, "POSTER"))
		item.HasFanart = item.HasFanart || (movie.BackdropPath != "" && scraperFieldEnabled(metadataFields, "FANART"))
	}
	var renamePreview *media.RenamePreview
	var renameWarnings []string
	oldID := ""
	if defaultBool(input.RenameAfterScrape, defaultBool(settings.MovieRenameAfterScrape, true)) {
		preview := media.BuildMovieRenameWithOptions(
			item,
			firstNonEmpty(movie.Title, item.TitleGuess),
			firstNonEmpty(yearFromDate(movie.ReleaseDate), item.YearGuess),
			movie.ID,
			settings.MovieRenamerPathname,
			settings.MovieRenamerFilename,
			movieFolderRenameOptions(settings),
			movieFileRenameOptions(settings),
		)
		renamePreview = &preview
		previousID := item.ID
		if err := media.ApplyRename(preview); err != nil {
			renameWarnings = append(renameWarnings, err.Error())
		} else {
			item = media.RefreshItemPath(item, preview.TargetFile)
			if item.ID != previousID {
				oldID = previousID
			}
		}
	}
	s.mu.Lock()
	if oldID != "" {
		delete(s.items, oldID)
	}
	s.items[item.ID] = item
	s.mu.Unlock()
	if err := s.store.SaveItem(item); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	if oldID != "" {
		if err := s.store.DeleteItem(oldID); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
	}
	response := map[string]interface{}{"item": item, "movie": movie}
	if renamePreview != nil {
		response["renamePreview"] = renamePreview
		response["renamed"] = len(renameWarnings) == 0
	}
	if len(renameWarnings) > 0 {
		response["renameWarnings"] = renameWarnings
	}
	writeJSON(w, response)
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
		preview = media.BuildTVShowRenameWithOptions(
			item,
			input.Title,
			"",
			input.Year,
			input.TMDBID,
			defaultString(input.TVShowRenamerShowFolder, defaultString(settings.TVShowRenamerShowFolder, defaultTVShowRenamerPath)),
			defaultString(input.TVShowRenamerSeason, defaultString(settings.TVShowRenamerSeason, defaultTVSeasonRenamer)),
			defaultString(input.TVShowRenamerFilename, defaultString(settings.TVShowRenamerFilename, defaultTVEpisodeRenamer)),
			tvShowFolderRenameOptions(settings),
			tvSeasonFolderRenameOptions(settings),
			tvEpisodeFileRenameOptions(settings),
		)
	} else {
		folderPattern := input.MovieRenamerPathname
		filePattern := input.MovieRenamerFilename
		if strings.TrimSpace(folderPattern) == "" && strings.TrimSpace(filePattern) == "" {
			folderPattern = input.Pattern
			filePattern = input.Pattern
		}
		preview = media.BuildMovieRenameWithOptions(
			item,
			input.Title,
			input.Year,
			input.TMDBID,
			defaultString(folderPattern, defaultString(settings.MovieRenamerPathname, defaultMovieRenamerPath)),
			defaultString(filePattern, defaultString(settings.MovieRenamerFilename, defaultMovieRenamerFile)),
			movieFolderRenameOptions(settings),
			movieFileRenameOptions(settings),
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
	var updated *media.Item
	oldID := ""
	s.mu.Lock()
	for id, item := range s.items {
		if item.Path == preview.SourceFile {
			changed := media.RefreshItemPath(item, preview.TargetFile)
			oldID = id
			delete(s.items, id)
			s.items[changed.ID] = changed
			updated = &changed
			break
		}
	}
	s.mu.Unlock()
	if updated != nil {
		if err := s.store.SaveItem(*updated); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		if oldID != "" && oldID != updated.ID {
			if err := s.store.DeleteItem(oldID); err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
		}
		writeJSON(w, map[string]interface{}{"ok": true, "item": updated})
		return
	}
	writeJSON(w, map[string]interface{}{"ok": true})
}

func (s *Server) handleLocalRename(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var input struct {
		Items []struct {
			ItemID      string `json:"itemId"`
			NewFileName string `json:"newFileName"`
		} `json:"items"`
	}
	if !decodeJSON(w, r, &input) {
		return
	}
	if len(input.Items) == 0 {
		http.Error(w, "items are required", http.StatusBadRequest)
		return
	}
	type renamePlan struct {
		item       media.Item
		oldID      string
		targetFile string
	}
	plans := make([]renamePlan, 0, len(input.Items))
	targets := map[string]string{}
	for _, request := range input.Items {
		item, ok := s.findItem(request.ItemID)
		if !ok {
			http.Error(w, "item not found", http.StatusNotFound)
			return
		}
		newFileName := strings.TrimSpace(request.NewFileName)
		if newFileName == "" {
			http.Error(w, "new filename is required", http.StatusBadRequest)
			return
		}
		if strings.ContainsAny(newFileName, `/\`) || filepath.Base(newFileName) != newFileName {
			http.Error(w, "new filename must not contain path separators", http.StatusBadRequest)
			return
		}
		targetFile := filepath.Join(item.Dir, newFileName)
		cleanTarget := filepath.Clean(targetFile)
		cleanSource := filepath.Clean(item.Path)
		if owner, exists := targets[cleanTarget]; exists && owner != item.ID {
			http.Error(w, "duplicate target filename: "+newFileName, http.StatusConflict)
			return
		}
		targets[cleanTarget] = item.ID
		if cleanTarget != cleanSource {
			if _, err := os.Stat(cleanTarget); err == nil {
				http.Error(w, "target file already exists: "+newFileName, http.StatusConflict)
				return
			} else if err != nil && !os.IsNotExist(err) {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
		plans = append(plans, renamePlan{item: item, oldID: item.ID, targetFile: targetFile})
	}

	updated := make([]media.Item, 0, len(plans))
	oldIDs := []string{}
	for _, plan := range plans {
		changed := plan.item
		if filepath.Clean(plan.item.Path) != filepath.Clean(plan.targetFile) {
			if err := os.Rename(plan.item.Path, plan.targetFile); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			changed = media.RefreshItemPath(plan.item, plan.targetFile)
		}
		updated = append(updated, changed)
		if changed.ID != plan.oldID {
			oldIDs = append(oldIDs, plan.oldID)
		}
	}

	s.mu.Lock()
	for _, oldID := range oldIDs {
		delete(s.items, oldID)
	}
	for _, item := range updated {
		s.items[item.ID] = item
	}
	s.mu.Unlock()

	if err := s.store.SaveItems(updated); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	for _, oldID := range oldIDs {
		if err := s.store.DeleteItem(oldID); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	writeJSON(w, map[string]interface{}{"ok": true, "items": updated})
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
		updated = append(updated, item)
	}
	return updated
}
