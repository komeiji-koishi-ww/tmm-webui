package media

import "testing"

func TestGuessTitleYear(t *testing.T) {
	title, year := GuessTitleYear("Blade.Runner.2049.2017.2160p.UHD.BluRay.x265.mkv")
	if title != "Blade Runner" {
		t.Fatalf("title = %q", title)
	}
	if year != "2049" {
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
