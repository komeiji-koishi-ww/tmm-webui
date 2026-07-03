package app

import (
	"net/http"
	"os"
	"sync"
	"time"

	"tmmweb/internal/media"
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
