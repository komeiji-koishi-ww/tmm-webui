package media

import "testing"

func TestGuessTitleYear(t *testing.T) {
	title, year := GuessTitleYear("Blade.Runner.2049.2017.2160p.UHD.BluRay.x265.mkv")
	if title != "Blade Runner 2049" {
		t.Fatalf("title = %q", title)
	}
	if year != "2017" {
		t.Fatalf("year = %q", year)
	}
}

func TestBuildMovieRename(t *testing.T) {
	item := Item{
		Path: "/media/Movies/Old/Movie.mkv",
		Dir:  "/media/Movies/Old",
	}
	preview := BuildMovieRename(item, "Blade/Runner", "1982", 78, "")
	if preview.TargetDir != "/media/Movies/Blade Runner (1982) {tmdb-78}" {
		t.Fatalf("target dir = %q", preview.TargetDir)
	}
	if preview.TargetFile != "/media/Movies/Blade Runner (1982) {tmdb-78}/Blade Runner (1982) {tmdb-78}.mkv" {
		t.Fatalf("target file = %q", preview.TargetFile)
	}
}

func TestGuessSeasonEpisodeTmmPatterns(t *testing.T) {
	tests := []struct {
		name     string
		season   int
		episodes []int
		airDate  string
	}{
		{name: "/shows/Dark/Season 02/Dark.S02E03.mkv", season: 2, episodes: []int{3}},
		{name: "/shows/Dark/Dark.1x02.mkv", season: 1, episodes: []int{2}},
		{name: "/shows/Dark/Season 01/102.mkv", season: 1, episodes: []int{2}},
		{name: "/shows/Dark/Season 01/02.mkv", season: 1, episodes: []int{2}},
		{name: "/shows/Daily/Daily.2024.08.13.mp4", season: 2024, airDate: "2024-08-13"},
	}
	for _, tt := range tests {
		match := GuessSeasonEpisode("/shows", tt.name, "Dark")
		if match.Season != tt.season {
			t.Fatalf("%s season = %d", tt.name, match.Season)
		}
		if tt.airDate != "" {
			if match.AirDate != tt.airDate {
				t.Fatalf("%s airDate = %q", tt.name, match.AirDate)
			}
			continue
		}
		if len(match.Episodes) != len(tt.episodes) {
			t.Fatalf("%s episodes = %#v", tt.name, match.Episodes)
		}
		for i := range tt.episodes {
			if match.Episodes[i] != tt.episodes[i] {
				t.Fatalf("%s episodes = %#v", tt.name, match.Episodes)
			}
		}
	}
}

func TestClassifyMediaFile(t *testing.T) {
	if got := ClassifyMediaFile("/media/movie-poster.jpg"); got != "POSTER" {
		t.Fatalf("poster classified as %s", got)
	}
	if got := ClassifyMediaFile("/media/trailers/trailer.mp4"); got != "TRAILER" {
		t.Fatalf("trailer classified as %s", got)
	}
	if got := ClassifyMediaFile("/media/movie.en.srt"); got != "SUBTITLE" {
		t.Fatalf("subtitle classified as %s", got)
	}
}

func TestShouldSkipDir(t *testing.T) {
	if !ShouldSkipDir("/media/@eaDir", "@eaDir", "movie") {
		t.Fatal("@eaDir should be skipped")
	}
	if !ShouldSkipDir("/media/Plex Versions", "Plex Versions", "movie") {
		t.Fatal("Plex Versions should be skipped")
	}
}
