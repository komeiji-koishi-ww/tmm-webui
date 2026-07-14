package media

import (
	"os"
	"path/filepath"
	"testing"
)

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
	if preview.TargetDir != "/media/Movies/Blade Runner (1982)" {
		t.Fatalf("target dir = %q", preview.TargetDir)
	}
	if preview.TargetFile != "/media/Movies/Blade Runner (1982)/Blade Runner (1982).mkv" {
		t.Fatalf("target file = %q", preview.TargetFile)
	}
}

func TestLightweightMediaInfoFromFilename(t *testing.T) {
	_, _, videoFormat, audioCodec := LightweightMediaInfo("/media/Movie.2024.2160p.WEB-DL.DTS-HD.MA.mkv", nil)
	if videoFormat != "2160p" {
		t.Fatalf("videoFormat = %q", videoFormat)
	}
	if audioCodec != "DTS-HD MA" {
		t.Fatalf("audioCodec = %q", audioCodec)
	}
	if size := FormatFileSize(13 * 1024 * 1024 * 1024); size != "13.0GB" {
		t.Fatalf("fileSize = %q", size)
	}
}

func TestRenamePatternUsesLightweightMediaInfo(t *testing.T) {
	item := Item{
		Path:        "/media/Movies/Old/Movie.1080p.DTS.mkv",
		Dir:         "/media/Movies/Old",
		FileName:    "Movie.1080p.DTS.mkv",
		VideoFormat: "1080p",
		AudioCodec:  "DTS",
		FileSize:    "8.4GB",
	}
	preview := BuildMovieRenameWithPatterns(item, "喜剧之王", "1999", 0, "${title} (${year}) ${videoFormat} - ${fileSize}", "${title} (${year}) ${videoFormat} ${audioCodec} ${fileSize}")
	if preview.TargetDir != "/media/Movies/喜剧之王 (1999) 1080p - 8.4GB" {
		t.Fatalf("target dir = %q", preview.TargetDir)
	}
	if preview.TargetFile != "/media/Movies/喜剧之王 (1999) 1080p - 8.4GB/喜剧之王 (1999) 1080p DTS 8.4GB.mkv" {
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

func TestUnchangedCachedDirRequiresCurrentFileState(t *testing.T) {
	root := t.TempDir()
	movieDir := filepath.Join(root, "Movie One")
	if err := os.Mkdir(movieDir, 0o755); err != nil {
		t.Fatal(err)
	}
	videoPath := filepath.Join(movieDir, "Movie.One.2024.mkv")
	if err := os.WriteFile(videoPath, []byte("video"), 0o644); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(videoPath)
	if err != nil {
		t.Fatal(err)
	}
	dirInfo, err := os.Stat(movieDir)
	if err != nil {
		t.Fatal(err)
	}

	library := Library{ID: "movies", Type: "movie", Path: root, Paths: []string{root}}
	item := NewItemFromFileInfo(library, root, videoPath, info)
	item.DirModTimeUnix = dirInfo.ModTime().UnixNano()
	existing := BuildScanExistingIndex(map[string]Item{item.ID: item})
	cache := buildExistingDirCache(existing, root)

	items, ok := unchangedCachedDir(movieDir, dirInfo, cache, existing, false)
	if !ok || len(items) != 1 || items[0].ID != item.ID {
		t.Fatalf("expected unchanged directory cache hit, ok=%v items=%d", ok, len(items))
	}

	if err := os.WriteFile(videoPath, []byte("changed video"), 0o644); err != nil {
		t.Fatal(err)
	}
	dirInfo, err = os.Stat(movieDir)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := unchangedCachedDir(movieDir, dirInfo, cache, existing, false); ok {
		t.Fatal("changed media file should invalidate unchanged directory cache")
	}
}

func TestCompactCachedItemDropsStreamDetails(t *testing.T) {
	item := Item{
		TitleGuess:           "Example",
		MediaDurationSeconds: 3600,
		VideoStreams:         []VideoStream{{Codec: "HEVC", DurationSeconds: 3600}},
		AudioStreams:         []AudioStream{{Codec: "EAC3", Language: "eng", Channels: 6}},
		SubtitleStreams:      []SubtitleStream{{Language: "zho"}},
	}
	compact := CompactCachedItem(item)
	if HasDetailedMediaInfo(compact) {
		t.Fatal("compact item retained stream details")
	}
	if compact.TitleGuess != item.TitleGuess || compact.MediaDurationSeconds != item.MediaDurationSeconds {
		t.Fatalf("compact item lost summary fields: %#v", compact)
	}
}

func TestScanLibrarySkipsUnchangedDirAndFindsNewDir(t *testing.T) {
	root := t.TempDir()
	oldDir := filepath.Join(root, "Old Movie")
	newDir := filepath.Join(root, "New Movie")
	if err := os.Mkdir(oldDir, 0o755); err != nil {
		t.Fatal(err)
	}
	oldPath := filepath.Join(oldDir, "Old.Movie.2020.mkv")
	if err := os.WriteFile(oldPath, []byte("old"), 0o644); err != nil {
		t.Fatal(err)
	}
	oldInfo, err := os.Stat(oldPath)
	if err != nil {
		t.Fatal(err)
	}
	oldDirInfo, err := os.Stat(oldDir)
	if err != nil {
		t.Fatal(err)
	}

	library := Library{ID: "movies", Type: "movie", Path: root, Paths: []string{root}}
	oldItem := NewItemFromFileInfo(library, root, oldPath, oldInfo)
	oldItem.DirModTimeUnix = oldDirInfo.ModTime().UnixNano()

	if err := os.Mkdir(newDir, 0o755); err != nil {
		t.Fatal(err)
	}
	newPath := filepath.Join(newDir, "New.Movie.2024.mkv")
	if err := os.WriteFile(newPath, []byte("new"), 0o644); err != nil {
		t.Fatal(err)
	}

	var progressPaths []string
	items, err := ScanLibraryWithOptions(library, ScanOptions{
		Existing:          map[string]Item{oldItem.ID: oldItem},
		SkipUnchangedDirs: true,
	}, func(progress ScanProgress) {
		progressPaths = append(progressPaths, progress.CurrentPath)
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 2 {
		t.Fatalf("items = %d, want 2", len(items))
	}
	ids := map[string]bool{}
	for _, item := range items {
		ids[item.ID] = true
	}
	if !ids[oldItem.ID] || !ids[stableID(newPath)] {
		t.Fatalf("scan did not keep old cached item and find new item: %#v", ids)
	}
	for _, path := range progressPaths {
		if path == oldPath {
			t.Fatal("unchanged cached media file should not be walked as a scan item")
		}
	}
}
