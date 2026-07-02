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

func TestReadSummaryAcceptsTMM3MovieCounters(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "movie.nfo")
	data := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<movie>
  <title>导火线</title>
  <originaltitle>Fuze</originaltitle>
  <year>2026</year>
  <ratings>
    <rating default="true" max="10" name="themoviedb">
      <value>6.471</value>
      <votes>86</votes>
    </rating>
  </ratings>
  <userrating>0.0</userrating>
  <plot>伦敦市中心的建筑工地，一枚未引爆的二战炸弹意外出土。</plot>
  <runtime>96</runtime>
  <id>tt31189814</id>
  <tmdbid>1242265</tmdbid>
  <uniqueid default="false" type="tmdb">1242265</uniqueid>
  <uniqueid default="true" type="imdb">tt31189814</uniqueid>
  <premiered>2026-03-23</premiered>
  <playcount/>
  <actor>
    <name>亚伦·泰勒-约翰逊</name>
    <role>Will</role>
  </actor>
</movie>`
	if err := os.WriteFile(path, []byte(data), 0644); err != nil {
		t.Fatal(err)
	}

	summary, err := ReadSummary(path)
	if err != nil {
		t.Fatal(err)
	}
	if summary.Title != "导火线" {
		t.Fatalf("Title = %q, want 导火线", summary.Title)
	}
	if summary.TMDBID != 1242265 {
		t.Fatalf("TMDBID = %d, want 1242265", summary.TMDBID)
	}
	if summary.IMDBID != "tt31189814" {
		t.Fatalf("IMDBID = %q, want tt31189814", summary.IMDBID)
	}
	if summary.Plot == "" {
		t.Fatal("Plot is empty")
	}
	if got, want := len(summary.Actors), 1; got != want {
		t.Fatalf("len(Actors) = %d, want %d: %#v", got, want, summary.Actors)
	}
}
