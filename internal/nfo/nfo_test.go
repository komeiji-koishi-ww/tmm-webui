package nfo

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadSummaryUsesLegacyNumericIDAndActors(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "movie.nfo")
	data := `<?xml version="1.0" encoding="UTF-8"?>
<movie>
  <title>Test Movie</title>
  <id>12345</id>
  <plot>Movie plot.</plot>
  <actor><name>Actor One</name></actor>
  <actor><name>Actor Two</name></actor>
  <actor><name>Actor One</name></actor>
</movie>`
	if err := os.WriteFile(path, []byte(data), 0644); err != nil {
		t.Fatal(err)
	}

	summary, err := ReadSummary(path)
	if err != nil {
		t.Fatal(err)
	}
	if summary.TMDBID != 12345 {
		t.Fatalf("TMDBID = %d, want 12345", summary.TMDBID)
	}
	if summary.IMDBID != "" {
		t.Fatalf("IMDBID = %q, want empty", summary.IMDBID)
	}
	if got, want := len(summary.Actors), 2; got != want {
		t.Fatalf("len(Actors) = %d, want %d: %#v", got, want, summary.Actors)
	}
}

func TestReadSummaryKeepsIMDBLegacyID(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "movie.nfo")
	data := `<?xml version="1.0" encoding="UTF-8"?>
<movie>
  <title>Test Movie</title>
  <id>tt1234567</id>
</movie>`
	if err := os.WriteFile(path, []byte(data), 0644); err != nil {
		t.Fatal(err)
	}

	summary, err := ReadSummary(path)
	if err != nil {
		t.Fatal(err)
	}
	if summary.IMDBID != "tt1234567" {
		t.Fatalf("IMDBID = %q, want tt1234567", summary.IMDBID)
	}
	if summary.TMDBID != 0 {
		t.Fatalf("TMDBID = %d, want 0", summary.TMDBID)
	}
}
