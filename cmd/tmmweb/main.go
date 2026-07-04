package main

import (
	"log"
	"net/http"
	"os"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"tmmweb/internal/app"
)

func main() {
	configureRuntimeMemory()

	addr := getenv("TMMWEB_ADDR", ":8080")
	dataDir := getenv("TMMWEB_DATA", "./data")
	tmdbKey := os.Getenv("TMDB_API_KEY")

	server, err := app.NewServer(app.Config{
		DataDir: dataDir,
		TMDBKey: tmdbKey,
		Client: &http.Client{
			Timeout: 20 * time.Second,
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("tmm-webui listening on %s", addr)
	if err := http.ListenAndServe(addr, server.Routes()); err != nil {
		log.Fatal(err)
	}
}

func getenv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func configureRuntimeMemory() {
	if raw := strings.TrimSpace(os.Getenv("TMMWEB_MEMORY_LIMIT_MB")); raw != "" {
		limitMB, err := strconv.Atoi(raw)
		if err != nil {
			log.Printf("ignoring invalid TMMWEB_MEMORY_LIMIT_MB=%q", raw)
		} else if limitMB > 0 {
			debug.SetMemoryLimit(int64(limitMB) << 20)
			log.Printf("runtime memory limit set to %dMB", limitMB)
		}
	}
	if raw := strings.TrimSpace(os.Getenv("TMMWEB_GOGC")); raw != "" {
		percent, err := strconv.Atoi(raw)
		if err != nil {
			log.Printf("ignoring invalid TMMWEB_GOGC=%q", raw)
			return
		}
		if percent < 10 {
			percent = 10
		}
		debug.SetGCPercent(percent)
		log.Printf("runtime GOGC set to %d", percent)
	}
}
