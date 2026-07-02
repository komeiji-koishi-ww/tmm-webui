package app

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"runtime/debug"
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
	defaultMovieRenamerPath  = "${title} ${- ,edition,} (${year}) ${videoFormat} - ${fileSize}"
	defaultMovieRenamerFile  = "${title} ${- ,edition,} (${year}) ${videoFormat} ${audioCodec} ${fileSize}"
	defaultTVShowRenamerPath = "${showTitle} (${showYear})"
	defaultTVSeasonRenamer   = "Season ${seasonNr2}"
	defaultTVEpisodeRenamer  = "${showTitle}.S${seasonNr2}E${episodeNr2}.${title}"
)

type Config struct {
	DataDir string
	TMDBKey string
	Client  *http.Client
}

type AppSettings struct {
	TMDBAPIKey                                 string   `json:"tmdbApiKey"`
	ProxyEnabled                               bool     `json:"proxyEnabled"`
	ProxyHost                                  string   `json:"proxyHost"`
	ProxyPort                                  int      `json:"proxyPort"`
	ProxyUsername                              string   `json:"proxyUsername"`
	ProxyPassword                              string   `json:"proxyPassword"`
	MovieScrapeMetadata                        *bool    `json:"movieScrapeMetadata,omitempty"`
	MovieScrapeNFO                             *bool    `json:"movieScrapeNfo,omitempty"`
	MovieScrapeImages                          *bool    `json:"movieScrapeImages,omitempty"`
	MovieScrapeOverwrite                       *bool    `json:"movieScrapeOverwrite,omitempty"`
	TVShowScrapeMetadata                       *bool    `json:"tvShowScrapeMetadata,omitempty"`
	TVShowEpisodeMetadata                      *bool    `json:"tvShowEpisodeMetadata,omitempty"`
	TVShowScrapeNFO                            *bool    `json:"tvShowScrapeNfo,omitempty"`
	TVShowScrapeImages                         *bool    `json:"tvShowScrapeImages,omitempty"`
	TVShowScrapeOverwrite                      *bool    `json:"tvShowScrapeOverwrite,omitempty"`
	MovieRenameAfterScrape                     *bool    `json:"movieRenameAfterScrape,omitempty"`
	TVShowRenameAfterScrape                    *bool    `json:"tvShowRenameAfterScrape,omitempty"`
	MovieScraperFields                         []string `json:"movieScraperFields,omitempty"`
	TVShowScraperFields                        []string `json:"tvShowScraperFields,omitempty"`
	TVEpisodeScraperFields                     []string `json:"tvEpisodeScraperFields,omitempty"`
	MovieRenamerPathname                       string   `json:"movieRenamerPathname"`
	MovieRenamerFilename                       string   `json:"movieRenamerFilename"`
	MovieRenamerPathSpaceSubstitution          *bool    `json:"movieRenamerPathSpaceSubstitution,omitempty"`
	MovieRenamerPathSpaceReplacement           string   `json:"movieRenamerPathSpaceReplacement"`
	MovieRenamerFilenameSpaceSubstitution      *bool    `json:"movieRenamerFilenameSpaceSubstitution,omitempty"`
	MovieRenamerFilenameSpaceReplacement       string   `json:"movieRenamerFilenameSpaceReplacement"`
	MovieRenamerColonReplacement               string   `json:"movieRenamerColonReplacement"`
	MovieRenamerAsciiReplacement               *bool    `json:"movieRenamerAsciiReplacement,omitempty"`
	MovieRenamerFirstCharacterReplacement      string   `json:"movieRenamerFirstCharacterReplacement"`
	MovieRenamerCreateSingleMovieSet           *bool    `json:"movieRenamerCreateSingleMovieSet,omitempty"`
	MovieRenamerNFOCleanup                     *bool    `json:"movieRenamerNfoCleanup,omitempty"`
	MovieRenamerCleanupUnwanted                *bool    `json:"movieRenamerCleanupUnwanted,omitempty"`
	MovieRenamerAllowMerge                     *bool    `json:"movieRenamerAllowMerge,omitempty"`
	TVShowRenamerShowFolder                    string   `json:"tvShowRenamerShowFolder"`
	TVShowRenamerSeason                        string   `json:"tvShowRenamerSeason"`
	TVShowRenamerFilename                      string   `json:"tvShowRenamerFilename"`
	TVShowRenamerShowFolderSpaceSubstitution   *bool    `json:"tvShowRenamerShowFolderSpaceSubstitution,omitempty"`
	TVShowRenamerShowFolderSpaceReplacement    string   `json:"tvShowRenamerShowFolderSpaceReplacement"`
	TVShowRenamerSeasonFolderSpaceSubstitution *bool    `json:"tvShowRenamerSeasonFolderSpaceSubstitution,omitempty"`
	TVShowRenamerSeasonFolderSpaceReplacement  string   `json:"tvShowRenamerSeasonFolderSpaceReplacement"`
	TVShowRenamerFilenameSpaceSubstitution     *bool    `json:"tvShowRenamerFilenameSpaceSubstitution,omitempty"`
	TVShowRenamerFilenameSpaceReplacement      string   `json:"tvShowRenamerFilenameSpaceReplacement"`
	TVShowRenamerColonReplacement              string   `json:"tvShowRenamerColonReplacement"`
	TVShowRenamerAsciiReplacement              *bool    `json:"tvShowRenamerAsciiReplacement,omitempty"`
	TVShowRenamerFirstCharacterReplacement     string   `json:"tvShowRenamerFirstCharacterReplacement"`
	TVShowRenamerCleanupUnwanted               *bool    `json:"tvShowRenamerCleanupUnwanted,omitempty"`
	MoviePosterName                            string   `json:"moviePosterName"`
	MovieFanartName                            string   `json:"movieFanartName"`
	MoviePosterNames                           string   `json:"moviePosterNames"`
	MovieFanartNames                           string   `json:"movieFanartNames"`
	TVShowPosterName                           string   `json:"tvShowPosterName"`
	TVShowFanartName                           string   `json:"tvShowFanartName"`
	TVShowPosterNames                          string   `json:"tvShowPosterNames"`
	TVShowFanartNames                          string   `json:"tvShowFanartNames"`
}

type SettingsResponse struct {
	TMDBConfigured                             bool     `json:"tmdbConfigured"`
	TMDBEnabled                                bool     `json:"tmdbEnabled"`
	TMDBKeySource                              string   `json:"tmdbKeySource"`
	ProxyEnabled                               bool     `json:"proxyEnabled"`
	ProxyHost                                  string   `json:"proxyHost"`
	ProxyPort                                  int      `json:"proxyPort"`
	ProxyUsername                              string   `json:"proxyUsername"`
	ProxyPassword                              bool     `json:"proxyPassword"`
	MovieScrapeMetadata                        bool     `json:"movieScrapeMetadata"`
	MovieScrapeNFO                             bool     `json:"movieScrapeNfo"`
	MovieScrapeImages                          bool     `json:"movieScrapeImages"`
	MovieScrapeOverwrite                       bool     `json:"movieScrapeOverwrite"`
	TVShowScrapeMetadata                       bool     `json:"tvShowScrapeMetadata"`
	TVShowEpisodeMetadata                      bool     `json:"tvShowEpisodeMetadata"`
	TVShowScrapeNFO                            bool     `json:"tvShowScrapeNfo"`
	TVShowScrapeImages                         bool     `json:"tvShowScrapeImages"`
	TVShowScrapeOverwrite                      bool     `json:"tvShowScrapeOverwrite"`
	MovieRenameAfterScrape                     bool     `json:"movieRenameAfterScrape"`
	TVShowRenameAfterScrape                    bool     `json:"tvShowRenameAfterScrape"`
	MovieScraperFields                         []string `json:"movieScraperFields"`
	TVShowScraperFields                        []string `json:"tvShowScraperFields"`
	TVEpisodeScraperFields                     []string `json:"tvEpisodeScraperFields"`
	MovieRenamerPathname                       string   `json:"movieRenamerPathname"`
	MovieRenamerFilename                       string   `json:"movieRenamerFilename"`
	MovieRenamerPathSpaceSubstitution          bool     `json:"movieRenamerPathSpaceSubstitution"`
	MovieRenamerPathSpaceReplacement           string   `json:"movieRenamerPathSpaceReplacement"`
	MovieRenamerFilenameSpaceSubstitution      bool     `json:"movieRenamerFilenameSpaceSubstitution"`
	MovieRenamerFilenameSpaceReplacement       string   `json:"movieRenamerFilenameSpaceReplacement"`
	MovieRenamerColonReplacement               string   `json:"movieRenamerColonReplacement"`
	MovieRenamerAsciiReplacement               bool     `json:"movieRenamerAsciiReplacement"`
	MovieRenamerFirstCharacterReplacement      string   `json:"movieRenamerFirstCharacterReplacement"`
	MovieRenamerCreateSingleMovieSet           bool     `json:"movieRenamerCreateSingleMovieSet"`
	MovieRenamerNFOCleanup                     bool     `json:"movieRenamerNfoCleanup"`
	MovieRenamerCleanupUnwanted                bool     `json:"movieRenamerCleanupUnwanted"`
	MovieRenamerAllowMerge                     bool     `json:"movieRenamerAllowMerge"`
	TVShowRenamerShowFolder                    string   `json:"tvShowRenamerShowFolder"`
	TVShowRenamerSeason                        string   `json:"tvShowRenamerSeason"`
	TVShowRenamerFilename                      string   `json:"tvShowRenamerFilename"`
	TVShowRenamerShowFolderSpaceSubstitution   bool     `json:"tvShowRenamerShowFolderSpaceSubstitution"`
	TVShowRenamerShowFolderSpaceReplacement    string   `json:"tvShowRenamerShowFolderSpaceReplacement"`
	TVShowRenamerSeasonFolderSpaceSubstitution bool     `json:"tvShowRenamerSeasonFolderSpaceSubstitution"`
	TVShowRenamerSeasonFolderSpaceReplacement  string   `json:"tvShowRenamerSeasonFolderSpaceReplacement"`
	TVShowRenamerFilenameSpaceSubstitution     bool     `json:"tvShowRenamerFilenameSpaceSubstitution"`
	TVShowRenamerFilenameSpaceReplacement      string   `json:"tvShowRenamerFilenameSpaceReplacement"`
	TVShowRenamerColonReplacement              string   `json:"tvShowRenamerColonReplacement"`
	TVShowRenamerAsciiReplacement              bool     `json:"tvShowRenamerAsciiReplacement"`
	TVShowRenamerFirstCharacterReplacement     string   `json:"tvShowRenamerFirstCharacterReplacement"`
	TVShowRenamerCleanupUnwanted               bool     `json:"tvShowRenamerCleanupUnwanted"`
	MoviePosterName                            string   `json:"moviePosterName"`
	MovieFanartName                            string   `json:"movieFanartName"`
	MoviePosterNames                           string   `json:"moviePosterNames"`
	MovieFanartNames                           string   `json:"movieFanartNames"`
	TVShowPosterName                           string   `json:"tvShowPosterName"`
	TVShowFanartName                           string   `json:"tvShowFanartName"`
	TVShowPosterNames                          string   `json:"tvShowPosterNames"`
	TVShowFanartNames                          string   `json:"tvShowFanartNames"`
}

type SettingsUpdate struct {
	TMDBAPIKey                                 *string  `json:"tmdbApiKey"`
	ClearTMDBKey                               bool     `json:"clearTmdbKey"`
	ProxyEnabled                               bool     `json:"proxyEnabled"`
	ProxyHost                                  string   `json:"proxyHost"`
	ProxyPort                                  int      `json:"proxyPort"`
	ProxyUsername                              string   `json:"proxyUsername"`
	ProxyPassword                              *string  `json:"proxyPassword"`
	ClearProxyPassword                         bool     `json:"clearProxyPassword"`
	MovieScrapeMetadata                        *bool    `json:"movieScrapeMetadata"`
	MovieScrapeNFO                             *bool    `json:"movieScrapeNfo"`
	MovieScrapeImages                          *bool    `json:"movieScrapeImages"`
	MovieScrapeOverwrite                       *bool    `json:"movieScrapeOverwrite"`
	TVShowScrapeMetadata                       *bool    `json:"tvShowScrapeMetadata"`
	TVShowEpisodeMetadata                      *bool    `json:"tvShowEpisodeMetadata"`
	TVShowScrapeNFO                            *bool    `json:"tvShowScrapeNfo"`
	TVShowScrapeImages                         *bool    `json:"tvShowScrapeImages"`
	TVShowScrapeOverwrite                      *bool    `json:"tvShowScrapeOverwrite"`
	MovieRenameAfterScrape                     *bool    `json:"movieRenameAfterScrape"`
	TVShowRenameAfterScrape                    *bool    `json:"tvShowRenameAfterScrape"`
	MovieScraperFields                         []string `json:"movieScraperFields"`
	TVShowScraperFields                        []string `json:"tvShowScraperFields"`
	TVEpisodeScraperFields                     []string `json:"tvEpisodeScraperFields"`
	MovieRenamerPathname                       string   `json:"movieRenamerPathname"`
	MovieRenamerFilename                       string   `json:"movieRenamerFilename"`
	MovieRenamerPathSpaceSubstitution          *bool    `json:"movieRenamerPathSpaceSubstitution"`
	MovieRenamerPathSpaceReplacement           string   `json:"movieRenamerPathSpaceReplacement"`
	MovieRenamerFilenameSpaceSubstitution      *bool    `json:"movieRenamerFilenameSpaceSubstitution"`
	MovieRenamerFilenameSpaceReplacement       string   `json:"movieRenamerFilenameSpaceReplacement"`
	MovieRenamerColonReplacement               string   `json:"movieRenamerColonReplacement"`
	MovieRenamerAsciiReplacement               *bool    `json:"movieRenamerAsciiReplacement"`
	MovieRenamerFirstCharacterReplacement      string   `json:"movieRenamerFirstCharacterReplacement"`
	MovieRenamerCreateSingleMovieSet           *bool    `json:"movieRenamerCreateSingleMovieSet"`
	MovieRenamerNFOCleanup                     *bool    `json:"movieRenamerNfoCleanup"`
	MovieRenamerCleanupUnwanted                *bool    `json:"movieRenamerCleanupUnwanted"`
	MovieRenamerAllowMerge                     *bool    `json:"movieRenamerAllowMerge"`
	TVShowRenamerShowFolder                    string   `json:"tvShowRenamerShowFolder"`
	TVShowRenamerSeason                        string   `json:"tvShowRenamerSeason"`
	TVShowRenamerFilename                      string   `json:"tvShowRenamerFilename"`
	TVShowRenamerShowFolderSpaceSubstitution   *bool    `json:"tvShowRenamerShowFolderSpaceSubstitution"`
	TVShowRenamerShowFolderSpaceReplacement    string   `json:"tvShowRenamerShowFolderSpaceReplacement"`
	TVShowRenamerSeasonFolderSpaceSubstitution *bool    `json:"tvShowRenamerSeasonFolderSpaceSubstitution"`
	TVShowRenamerSeasonFolderSpaceReplacement  string   `json:"tvShowRenamerSeasonFolderSpaceReplacement"`
	TVShowRenamerFilenameSpaceSubstitution     *bool    `json:"tvShowRenamerFilenameSpaceSubstitution"`
	TVShowRenamerFilenameSpaceReplacement      string   `json:"tvShowRenamerFilenameSpaceReplacement"`
	TVShowRenamerColonReplacement              string   `json:"tvShowRenamerColonReplacement"`
	TVShowRenamerAsciiReplacement              *bool    `json:"tvShowRenamerAsciiReplacement"`
	TVShowRenamerFirstCharacterReplacement     string   `json:"tvShowRenamerFirstCharacterReplacement"`
	TVShowRenamerCleanupUnwanted               *bool    `json:"tvShowRenamerCleanupUnwanted"`
	MoviePosterName                            string   `json:"moviePosterName"`
	MovieFanartName                            string   `json:"movieFanartName"`
	MoviePosterNames                           string   `json:"moviePosterNames"`
	MovieFanartNames                           string   `json:"movieFanartNames"`
	TVShowPosterName                           string   `json:"tvShowPosterName"`
	TVShowFanartName                           string   `json:"tvShowFanartName"`
	TVShowPosterNames                          string   `json:"tvShowPosterNames"`
	TVShowFanartNames                          string   `json:"tvShowFanartNames"`
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
	go server.refreshCachedItems()
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
	mux.HandleFunc("/api/artwork", s.handleArtwork)
	mux.HandleFunc("/api/browse", s.handleBrowse)
	mux.HandleFunc("/api/search", s.handleSearch)
	mux.HandleFunc("/api/metadata", s.handleMetadata)
	mux.HandleFunc("/api/scrape", s.handleScrape)
	mux.HandleFunc("/api/rename/preview", s.handleRenamePreview)
	mux.HandleFunc("/api/rename/apply", s.handleRenameApply)
	mux.HandleFunc("/api/local-rename", s.handleLocalRename)
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
	w.Header().Set("Content-Type", "application/json")
	_, _ = io.WriteString(w, `{"items":[`)
	s.mu.Lock()
	count := 0
	for _, item := range s.items {
		if libraryID == "" || item.LibraryID == libraryID {
			if count > 0 {
				_, _ = io.WriteString(w, ",")
			}
			data, err := json.Marshal(item)
			if err == nil {
				_, _ = w.Write(data)
				count++
			}
		}
	}
	s.mu.Unlock()
	_, _ = fmt.Fprintf(w, `],"count":%d}`+"\n", count)
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

func writeMovieNFO(dir string, movie tmdb.Movie, item media.Item, overwrite bool) error {
	path := filepath.Join(dir, "movie.nfo")
	if !overwrite {
		if _, err := os.Stat(path); err == nil {
			return nil
		}
	}
	return nfo.WriteMovie(dir, movie, nfoMediaFileInfo(item))
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

func writeTVSeasonNFO(path string, season tmdb.TVSeason, fanartPath string, overwrite bool) error {
	if !overwrite {
		if _, err := os.Stat(path); err == nil {
			return nil
		}
	}
	return nfo.WriteTVSeason(path, season, fanartPath)
}

func writeTVEpisodeNFO(path string, show tmdb.TVShow, episode tmdb.TVEpisode, item media.Item, overwrite bool) error {
	if !overwrite {
		if _, err := os.Stat(path); err == nil {
			return nil
		}
	}
	return nfo.WriteTVEpisode(path, show, episode, nfoMediaFileInfo(item))
}

func nfoMediaFileInfo(item media.Item) nfo.EpisodeFileInfo {
	info := nfo.EpisodeFileInfo{
		FileName:  item.FileName,
		DateAdded: nfoDateTime(item.DateAdded),
	}
	for _, stream := range item.VideoStreams {
		info.VideoStreams = append(info.VideoStreams, nfo.VideoStream{
			Codec: stream.Codec, Aspect: stream.Aspect, Width: stream.Width, Height: stream.Height,
			DurationSeconds: stream.DurationSeconds, StereoMode: stream.StereoMode, HDRType: stream.HDRType,
		})
	}
	for _, stream := range item.AudioStreams {
		info.AudioStreams = append(info.AudioStreams, nfo.AudioStream{
			Codec: stream.Codec, Language: stream.Language, Channels: stream.Channels,
		})
	}
	for _, stream := range item.SubtitleStreams {
		info.SubtitleStreams = append(info.SubtitleStreams, nfo.SubtitleStream{Language: stream.Language})
	}
	return info
}

func nfoDateTime(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	if parsed, err := time.Parse(time.RFC3339, value); err == nil {
		return parsed.Local().Format("2006-01-02 15:04:05")
	}
	return value
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

func downloadTVTheme(dir string, tvdbID int, overwrite bool) error {
	if tvdbID <= 0 {
		return nil
	}
	path := filepath.Join(dir, "theme.mp3")
	if !overwrite {
		if _, err := os.Stat(path); err == nil {
			return nil
		}
	}
	client := http.Client{Timeout: 20 * time.Second}
	response, err := client.Get(fmt.Sprintf("http://tvthemes.plexapp.com/%d.mp3", tvdbID))
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode == http.StatusNotFound || response.StatusCode == http.StatusForbidden {
		return nil
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return fmt.Errorf("theme download status %d", response.StatusCode)
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	temp, err := os.CreateTemp(dir, "theme.*.part")
	if err != nil {
		return err
	}
	tempPath := temp.Name()
	_, copyErr := io.Copy(temp, response.Body)
	closeErr := temp.Close()
	if copyErr != nil {
		_ = os.Remove(tempPath)
		return copyErr
	}
	if closeErr != nil {
		_ = os.Remove(tempPath)
		return closeErr
	}
	if overwrite {
		_ = os.Remove(path)
	}
	return os.Rename(tempPath, path)
}

func artworkPath(item media.Item, artType string, scope string) string {
	var dirs []string
	var names []string
	fileBase := strings.TrimSuffix(item.FileName, filepath.Ext(item.FileName))
	if item.Kind == "tvshow" {
		showDir := tvShowRootDir(item)
		seasonDir := tvSeasonDir(item, showDir)
		switch scope {
		case "show":
			dirs = append(dirs, showDir)
			if artType == "poster" {
				names = append(names, defaultTVShowPosterNames()...)
			} else {
				names = append(names, defaultTVShowFanartNames()...)
			}
		case "season":
			if artType == "poster" && item.Season > 0 {
				dirs = append(dirs, seasonDir, showDir)
				names = append(names, seasonPosterNames(item.Season)...)
				names = append(names, defaultTVShowPosterNames()...)
			} else if artType == "fanart" && item.Season > 0 {
				dirs = append(dirs, seasonDir, showDir)
				names = append(names, seasonFanartNames(item.Season)...)
				names = append(names, defaultTVShowFanartNames()...)
			}
		default:
			if artType == "poster" && item.Season > 0 {
				dirs = append(dirs, seasonDir, showDir)
				names = append(names, seasonPosterNames(item.Season)...)
				names = append(names, defaultTVShowPosterNames()...)
			} else if artType == "fanart" {
				dirs = append(dirs, item.Dir, seasonDir, showDir)
				names = append(names, episodeThumbNames(item)...)
			}
		}
		if len(dirs) == 0 {
			dirs = append(dirs, showDir)
		}
		if len(names) == 0 {
			if artType == "poster" {
				names = append(names, defaultTVShowPosterNames()...)
			} else {
				names = append(names, defaultTVShowFanartNames()...)
			}
		}
	} else {
		dirs = append(dirs, item.Dir)
		if artType == "poster" {
			names = append(names, "poster.jpg", "folder.jpg", fileBase+"-poster.jpg")
		} else {
			names = append(names, "fanart.jpg", "backdrop.jpg", fileBase+"-fanart.jpg")
		}
	}
	seen := map[string]bool{}
	for _, dir := range dirs {
		if dir == "" {
			continue
		}
		for _, name := range names {
			name = strings.ReplaceAll(name, "{filename}", fileBase)
			path := filepath.Join(dir, filepath.Base(name))
			if seen[path] {
				continue
			}
			seen[path] = true
			if stat, err := os.Stat(path); err == nil && !stat.IsDir() {
				return path
			}
		}
	}
	return ""
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

func seasonNFOPath(item media.Item, season int) string {
	showDir := tvShowRootDir(item)
	seasonDir := tvSeasonDir(item, showDir)
	if seasonDir != "" && seasonDir != showDir {
		return filepath.Join(seasonDir, "season.nfo")
	}
	return filepath.Join(showDir, seasonFilename(season, "nfo"))
}

func tvShowRootDir(item media.Item) string {
	source := strings.TrimSpace(item.SourcePath)
	if source != "" {
		if rel, err := filepath.Rel(source, item.Path); err == nil && rel != "." && !strings.HasPrefix(rel, "..") {
			parts := strings.Split(rel, string(filepath.Separator))
			if len(parts) > 1 && parts[0] != "" {
				return filepath.Join(source, parts[0])
			}
		}
	}
	if item.Season > 0 || item.Episode > 0 {
		parent := filepath.Dir(item.Dir)
		if parent != "." && parent != string(filepath.Separator) {
			return parent
		}
	}
	return item.Dir
}

func tvSeasonDir(item media.Item, showDir string) string {
	if item.Season > 0 || item.Episode > 0 {
		return item.Dir
	}
	return showDir
}

func seasonPosterNames(season int) []string {
	return []string{seasonFilename(season, "jpg", "-poster"), fmt.Sprintf("season%d-poster.jpg", season)}
}

func seasonFanartNames(season int) []string {
	return []string{seasonFilename(season, "jpg", "-fanart"), fmt.Sprintf("season%d-fanart.jpg", season)}
}

func episodeThumbNames(item media.Item) []string {
	fileBase := strings.TrimSuffix(item.FileName, filepath.Ext(item.FileName))
	return []string{
		fileBase + "-thumb.jpg",
	}
}

func episodeNFOPath(item media.Item) string {
	return strings.TrimSuffix(item.Path, filepath.Ext(item.Path)) + ".nfo"
}

func seasonFilename(season int, extension string, suffix ...string) string {
	name := ""
	if season == 0 {
		name = "season-specials"
	} else {
		name = fmt.Sprintf("season%02d", season)
	}
	if len(suffix) > 0 {
		name += suffix[0]
	}
	return name + "." + extension
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
	defer releaseScanMemory()

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
	s.mu.Lock()
	existing := make(map[string]media.Item)
	for id, item := range s.items {
		if item.LibraryID == library.ID {
			existing[id] = item
		}
	}
	s.mu.Unlock()
	currentIDs, err := media.ScanLibraryIDsWithOptions(library, media.ScanOptions{
		Existing:       existing,
		ProbeMediaInfo: scanMediaInfoEnabled(),
	}, func(progress media.ScanProgress) {
		s.mu.Lock()
		if task := s.tasks[taskID]; task != nil {
			task.SourcePath = progress.SourcePath
			task.CurrentPath = progress.CurrentPath
			task.VisitedFiles = progress.VisitedFiles
			task.FoundItems = progress.FoundItems
		}
		if progress.Item != nil {
			item := *progress.Item
			if existing, ok := s.items[item.ID]; ok {
				item = media.MergeScannedItem(existing, item)
			}
			s.items[item.ID] = item
			if persistErr == nil {
				pending = append(pending, item)
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
	existing = nil
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
	if err := s.store.PruneLibraryItemIDs(library.ID, currentIDs); err != nil {
		task.State = "failed"
		task.Error = err.Error()
		_ = s.store.SaveTask(task.toRecord())
		return
	}
	currentIDSet := make(map[string]struct{}, len(currentIDs))
	for _, id := range currentIDs {
		currentIDSet[id] = struct{}{}
	}
	for id, item := range s.items {
		if item.LibraryID == library.ID {
			if _, ok := currentIDSet[id]; !ok {
				delete(s.items, id)
			}
		}
	}
	task.State = "completed"
	task.ResultCount = len(currentIDs)
	task.FoundItems = len(currentIDs)
	_ = s.store.SaveTask(task.toRecord())
}

func releaseScanMemory() {
	runtime.GC()
	debug.FreeOSMemory()
}

func scanMediaInfoEnabled() bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv("TMMWEB_SCAN_MEDIAINFO"))) {
	case "", "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	default:
		return true
	}
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
		if item.DateAdded == "" {
			item.DateAdded = time.Now().UTC().Format(time.RFC3339)
		}
		s.items[item.ID] = item
	}
	return nil
}

func (s *Server) refreshCachedItems() {
	s.mu.Lock()
	items := make([]media.Item, 0, len(s.items))
	for _, item := range s.items {
		items = append(items, item)
	}
	s.mu.Unlock()

	var batch []media.Item
	flush := func() {
		if len(batch) == 0 {
			return
		}
		_ = s.store.SaveItems(batch)
		batch = batch[:0]
	}

	for _, item := range items {
		info, err := os.Stat(item.Path)
		if err != nil {
			continue
		}
		library := media.Library{
			ID:    item.LibraryID,
			Path:  item.SourcePath,
			Paths: []string{item.SourcePath},
			Type:  item.Kind,
		}
		local := media.NewItemFromFileInfo(library, item.SourcePath, item.Path, info)
		merged := media.MergeScannedItem(item, local)
		if !itemChanged(item, merged) {
			continue
		}
		s.mu.Lock()
		s.items[merged.ID] = merged
		s.mu.Unlock()
		batch = append(batch, merged)
		if len(batch) >= scanPersistBatchSize {
			flush()
		}
	}
	flush()
}

func itemChanged(a media.Item, b media.Item) bool {
	return a.DateAdded != b.DateAdded ||
		a.TitleGuess != b.TitleGuess ||
		a.YearGuess != b.YearGuess ||
		a.Original != b.Original ||
		a.Overview != b.Overview ||
		a.Runtime != b.Runtime ||
		a.Rating != b.Rating ||
		a.ShowRating != b.ShowRating ||
		a.ModTimeUnix != b.ModTimeUnix ||
		a.NFOModTimeUnix != b.NFOModTimeUnix ||
		a.FileSize != b.FileSize ||
		a.FileSizeBytes != b.FileSizeBytes ||
		a.VideoFormat != b.VideoFormat ||
		a.AudioCodec != b.AudioCodec ||
		!reflect.DeepEqual(a.VideoStreams, b.VideoStreams) ||
		!reflect.DeepEqual(a.AudioStreams, b.AudioStreams) ||
		!reflect.DeepEqual(a.SubtitleStreams, b.SubtitleStreams) ||
		a.MediaInfoScanned != b.MediaInfoScanned ||
		strings.Join(a.Genres, "\x00") != strings.Join(b.Genres, "\x00") ||
		strings.Join(a.Actors, "\x00") != strings.Join(b.Actors, "\x00") ||
		a.Premiered != b.Premiered ||
		a.IMDBID != b.IMDBID ||
		a.MatchedID != b.MatchedID ||
		a.MatchedName != b.MatchedName ||
		a.HasNFO != b.HasNFO ||
		a.HasPoster != b.HasPoster ||
		a.HasFanart != b.HasFanart ||
		a.HasSubtitle != b.HasSubtitle
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
		TMDBConfigured:                             s.tmdb.Enabled(),
		TMDBEnabled:                                s.tmdb.Enabled(),
		TMDBKeySource:                              source,
		ProxyEnabled:                               s.settings.ProxyEnabled,
		ProxyHost:                                  s.settings.ProxyHost,
		ProxyPort:                                  s.settings.ProxyPort,
		ProxyUsername:                              s.settings.ProxyUsername,
		ProxyPassword:                              s.settings.ProxyPassword != "",
		MovieScrapeMetadata:                        defaultBool(s.settings.MovieScrapeMetadata, true),
		MovieScrapeNFO:                             defaultBool(s.settings.MovieScrapeNFO, true),
		MovieScrapeImages:                          defaultBool(s.settings.MovieScrapeImages, true),
		MovieScrapeOverwrite:                       defaultBool(s.settings.MovieScrapeOverwrite, false),
		TVShowScrapeMetadata:                       defaultBool(s.settings.TVShowScrapeMetadata, true),
		TVShowEpisodeMetadata:                      defaultBool(s.settings.TVShowEpisodeMetadata, true),
		TVShowScrapeNFO:                            defaultBool(s.settings.TVShowScrapeNFO, true),
		TVShowScrapeImages:                         defaultBool(s.settings.TVShowScrapeImages, true),
		TVShowScrapeOverwrite:                      defaultBool(s.settings.TVShowScrapeOverwrite, false),
		MovieRenameAfterScrape:                     defaultBool(s.settings.MovieRenameAfterScrape, true),
		TVShowRenameAfterScrape:                    defaultBool(s.settings.TVShowRenameAfterScrape, true),
		MovieScraperFields:                         normalizeScraperFields(s.settings.MovieScraperFields, defaultMovieScraperFields()),
		TVShowScraperFields:                        normalizeScraperFields(s.settings.TVShowScraperFields, defaultTVShowScraperFields()),
		TVEpisodeScraperFields:                     normalizeScraperFields(s.settings.TVEpisodeScraperFields, defaultTVEpisodeScraperFields()),
		MovieRenamerPathname:                       defaultString(s.settings.MovieRenamerPathname, defaultMovieRenamerPath),
		MovieRenamerFilename:                       defaultString(s.settings.MovieRenamerFilename, defaultMovieRenamerFile),
		MovieRenamerPathSpaceSubstitution:          defaultBool(s.settings.MovieRenamerPathSpaceSubstitution, false),
		MovieRenamerPathSpaceReplacement:           defaultString(s.settings.MovieRenamerPathSpaceReplacement, "_"),
		MovieRenamerFilenameSpaceSubstitution:      defaultBool(s.settings.MovieRenamerFilenameSpaceSubstitution, false),
		MovieRenamerFilenameSpaceReplacement:       defaultString(s.settings.MovieRenamerFilenameSpaceReplacement, "_"),
		MovieRenamerColonReplacement:               defaultString(s.settings.MovieRenamerColonReplacement, "-"),
		MovieRenamerAsciiReplacement:               defaultBool(s.settings.MovieRenamerAsciiReplacement, false),
		MovieRenamerFirstCharacterReplacement:      defaultString(s.settings.MovieRenamerFirstCharacterReplacement, "#"),
		MovieRenamerCreateSingleMovieSet:           defaultBool(s.settings.MovieRenamerCreateSingleMovieSet, false),
		MovieRenamerNFOCleanup:                     defaultBool(s.settings.MovieRenamerNFOCleanup, false),
		MovieRenamerCleanupUnwanted:                defaultBool(s.settings.MovieRenamerCleanupUnwanted, false),
		MovieRenamerAllowMerge:                     defaultBool(s.settings.MovieRenamerAllowMerge, false),
		TVShowRenamerShowFolder:                    defaultString(s.settings.TVShowRenamerShowFolder, defaultTVShowRenamerPath),
		TVShowRenamerSeason:                        defaultString(s.settings.TVShowRenamerSeason, defaultTVSeasonRenamer),
		TVShowRenamerFilename:                      defaultString(s.settings.TVShowRenamerFilename, defaultTVEpisodeRenamer),
		TVShowRenamerShowFolderSpaceSubstitution:   defaultBool(s.settings.TVShowRenamerShowFolderSpaceSubstitution, false),
		TVShowRenamerShowFolderSpaceReplacement:    defaultString(s.settings.TVShowRenamerShowFolderSpaceReplacement, "_"),
		TVShowRenamerSeasonFolderSpaceSubstitution: defaultBool(s.settings.TVShowRenamerSeasonFolderSpaceSubstitution, false),
		TVShowRenamerSeasonFolderSpaceReplacement:  defaultString(s.settings.TVShowRenamerSeasonFolderSpaceReplacement, "_"),
		TVShowRenamerFilenameSpaceSubstitution:     defaultBool(s.settings.TVShowRenamerFilenameSpaceSubstitution, false),
		TVShowRenamerFilenameSpaceReplacement:      defaultString(s.settings.TVShowRenamerFilenameSpaceReplacement, "_"),
		TVShowRenamerColonReplacement:              defaultString(s.settings.TVShowRenamerColonReplacement, " "),
		TVShowRenamerAsciiReplacement:              defaultBool(s.settings.TVShowRenamerAsciiReplacement, false),
		TVShowRenamerFirstCharacterReplacement:     defaultString(s.settings.TVShowRenamerFirstCharacterReplacement, "#"),
		TVShowRenamerCleanupUnwanted:               defaultBool(s.settings.TVShowRenamerCleanupUnwanted, false),
		MoviePosterName:                            defaultString(s.settings.MoviePosterName, "poster.jpg"),
		MovieFanartName:                            defaultString(s.settings.MovieFanartName, "fanart.jpg"),
		MoviePosterNames:                           defaultImageNames(s.settings.MoviePosterNames, s.settings.MoviePosterName, defaultMoviePosterNames()),
		MovieFanartNames:                           defaultImageNames(s.settings.MovieFanartNames, s.settings.MovieFanartName, defaultMovieFanartNames()),
		TVShowPosterName:                           defaultString(s.settings.TVShowPosterName, "poster.jpg"),
		TVShowFanartName:                           defaultString(s.settings.TVShowFanartName, "fanart.jpg"),
		TVShowPosterNames:                          defaultImageNames(s.settings.TVShowPosterNames, s.settings.TVShowPosterName, defaultTVShowPosterNames()),
		TVShowFanartNames:                          defaultImageNames(s.settings.TVShowFanartNames, s.settings.TVShowFanartName, defaultTVShowFanartNames()),
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

func normalizeReplacement(value string, fallback string, allowed []string) string {
	for _, candidate := range allowed {
		if value == candidate {
			return value
		}
	}
	return fallback
}

func movieFolderRenameOptions(settings AppSettings) media.RenameOptions {
	return media.RenameOptions{
		SpaceSubstitution:               defaultBool(settings.MovieRenamerPathSpaceSubstitution, false),
		SpaceReplacement:                defaultString(settings.MovieRenamerPathSpaceReplacement, "_"),
		ColonReplacement:                defaultString(settings.MovieRenamerColonReplacement, "-"),
		ColonReplacementDefined:         true,
		ASCIIReplacement:                defaultBool(settings.MovieRenamerAsciiReplacement, false),
		FirstCharacterNumberReplacement: defaultString(settings.MovieRenamerFirstCharacterReplacement, "#"),
	}
}

func movieFileRenameOptions(settings AppSettings) media.RenameOptions {
	return media.RenameOptions{
		SpaceSubstitution:               defaultBool(settings.MovieRenamerFilenameSpaceSubstitution, false),
		SpaceReplacement:                defaultString(settings.MovieRenamerFilenameSpaceReplacement, "_"),
		ColonReplacement:                defaultString(settings.MovieRenamerColonReplacement, "-"),
		ColonReplacementDefined:         true,
		ASCIIReplacement:                defaultBool(settings.MovieRenamerAsciiReplacement, false),
		FirstCharacterNumberReplacement: defaultString(settings.MovieRenamerFirstCharacterReplacement, "#"),
	}
}

func tvShowFolderRenameOptions(settings AppSettings) media.RenameOptions {
	return media.RenameOptions{
		SpaceSubstitution:               defaultBool(settings.TVShowRenamerShowFolderSpaceSubstitution, false),
		SpaceReplacement:                defaultString(settings.TVShowRenamerShowFolderSpaceReplacement, "_"),
		ColonReplacement:                defaultString(settings.TVShowRenamerColonReplacement, " "),
		ColonReplacementDefined:         true,
		ASCIIReplacement:                defaultBool(settings.TVShowRenamerAsciiReplacement, false),
		FirstCharacterNumberReplacement: defaultString(settings.TVShowRenamerFirstCharacterReplacement, "#"),
	}
}

func tvSeasonFolderRenameOptions(settings AppSettings) media.RenameOptions {
	return media.RenameOptions{
		SpaceSubstitution:               defaultBool(settings.TVShowRenamerSeasonFolderSpaceSubstitution, false),
		SpaceReplacement:                defaultString(settings.TVShowRenamerSeasonFolderSpaceReplacement, "_"),
		ColonReplacement:                defaultString(settings.TVShowRenamerColonReplacement, " "),
		ColonReplacementDefined:         true,
		ASCIIReplacement:                defaultBool(settings.TVShowRenamerAsciiReplacement, false),
		FirstCharacterNumberReplacement: defaultString(settings.TVShowRenamerFirstCharacterReplacement, "#"),
	}
}

func tvEpisodeFileRenameOptions(settings AppSettings) media.RenameOptions {
	return media.RenameOptions{
		SpaceSubstitution:               defaultBool(settings.TVShowRenamerFilenameSpaceSubstitution, false),
		SpaceReplacement:                defaultString(settings.TVShowRenamerFilenameSpaceReplacement, "_"),
		ColonReplacement:                defaultString(settings.TVShowRenamerColonReplacement, " "),
		ColonReplacementDefined:         true,
		ASCIIReplacement:                defaultBool(settings.TVShowRenamerAsciiReplacement, false),
		FirstCharacterNumberReplacement: defaultString(settings.TVShowRenamerFirstCharacterReplacement, "#"),
	}
}

func tvEpisodeMetadataForItem(season tmdb.TVSeason, item media.Item) (tmdb.TVEpisode, bool) {
	targets := map[int]bool{}
	if item.Episode > 0 {
		targets[item.Episode] = true
	}
	for _, episode := range item.Episodes {
		if episode > 0 {
			targets[episode] = true
		}
	}
	if len(targets) == 0 {
		return tmdb.TVEpisode{}, false
	}
	for _, episode := range season.Episodes {
		if targets[episode.EpisodeNumber] {
			return episode, true
		}
	}
	return tmdb.TVEpisode{}, false
}

func tvEpisodeRenameTitle(item media.Item, showTitle string, loadSeasonMetadataFor func(int) (tmdb.TVSeason, error)) string {
	if item.Season > 0 {
		if seasonData, err := loadSeasonMetadataFor(item.Season); err == nil {
			if episode, ok := tvEpisodeMetadataForItem(seasonData, item); ok && strings.TrimSpace(episode.Title) != "" {
				return episode.Title
			}
		}
	}
	if item.Episode > 0 {
		return fmt.Sprintf("%02d", item.Episode)
	}
	for _, episode := range item.Episodes {
		if episode > 0 {
			return fmt.Sprintf("%02d", episode)
		}
	}
	title := strings.TrimSpace(item.MatchedName)
	if title != "" && !strings.EqualFold(title, strings.TrimSpace(showTitle)) && !strings.EqualFold(title, strings.TrimSpace(item.ShowGuess)) {
		return title
	}
	return ""
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func yearFromDate(value string) string {
	if len(value) >= 4 {
		return value[:4]
	}
	return ""
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

func defaultMovieScraperFields() []string {
	return []string{
		"ID", "TITLE", "ORIGINAL_TITLE", "TAGLINE", "PLOT", "YEAR", "RELEASE_DATE", "RATING", "TOP250", "RUNTIME",
		"CERTIFICATION", "GENRES", "SPOKEN_LANGUAGES", "COUNTRY", "PRODUCTION_COMPANY", "TAGS", "COLLECTION", "TRAILER",
		"ACTORS", "PRODUCERS", "DIRECTORS", "WRITERS",
		"POSTER", "FANART", "BANNER", "CLEARART", "THUMB", "CLEARLOGO", "DISCART", "KEYART", "EXTRAFANART", "EXTRATHUMB",
	}
}

func defaultTVShowScraperFields() []string {
	return []string{
		"ID", "TITLE", "ORIGINAL_TITLE", "PLOT", "YEAR", "AIRED", "STATUS", "RATING", "TOP250", "RUNTIME", "CERTIFICATION",
		"GENRES", "COUNTRY", "STUDIO", "TAGS", "TRAILER", "SEASON_NAMES", "SEASON_OVERVIEW",
		"ACTORS",
		"POSTER", "FANART", "BANNER", "CLEARART", "THUMB", "CLEARLOGO", "DISCART", "KEYART", "CHARACTERART", "EXTRAFANART",
		"SEASON_POSTER", "SEASON_FANART", "SEASON_BANNER", "SEASON_THUMB", "THEME",
	}
}

func defaultTVEpisodeScraperFields() []string {
	return []string{
		"TITLE", "ORIGINAL_TITLE", "PLOT", "SEASON_EPISODE", "AIRED", "RATING", "TAGS",
		"ACTORS", "DIRECTORS", "WRITERS", "THUMB",
	}
}

func normalizeScraperFields(values []string, fallback []string) []string {
	if values == nil {
		return append([]string(nil), fallback...)
	}
	allowed := make(map[string]bool, len(fallback))
	for _, value := range fallback {
		allowed[value] = true
	}
	seen := map[string]bool{}
	var normalized []string
	for _, value := range values {
		value = strings.ToUpper(strings.TrimSpace(value))
		if value == "" || !allowed[value] || seen[value] {
			continue
		}
		seen[value] = true
		normalized = append(normalized, value)
	}
	if len(normalized) == 0 && len(values) > 0 {
		return []string{}
	}
	if len(normalized) == 0 {
		return append([]string(nil), fallback...)
	}
	return normalized
}

func scraperFieldEnabled(fields []string, key string) bool {
	if len(fields) == 0 {
		return false
	}
	key = strings.ToUpper(strings.TrimSpace(key))
	for _, field := range fields {
		if strings.EqualFold(field, key) {
			return true
		}
	}
	return false
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
	devStatic := strings.TrimSpace(os.Getenv("TMMWEB_DEV_STATIC"))
	if devStatic != "" && devStatic != "0" && !strings.EqualFold(devStatic, "false") {
		return staticFileHandler(os.DirFS("internal/app/static"), true)
	}
	sub, err := fs.Sub(staticFS, "static")
	if err != nil {
		panic(err)
	}
	return staticFileHandler(sub, false)
}

func staticFileHandler(fileSystem fs.FS, noStore bool) http.Handler {
	fileServer := http.FileServer(http.FS(fileSystem))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if noStore {
			w.Header().Set("Cache-Control", "no-store, max-age=0")
			w.Header().Set("Pragma", "no-cache")
			w.Header().Set("Expires", "0")
		}
		path := strings.TrimPrefix(r.URL.Path, "/")
		if path == "" {
			r.URL.Path = "/"
			fileServer.ServeHTTP(w, r)
			return
		}
		if _, err := fs.Stat(fileSystem, path); err != nil {
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
