package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"tmmweb/internal/app"
)

func main() {
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
