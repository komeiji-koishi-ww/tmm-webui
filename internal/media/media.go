package media

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

var videoExts = map[string]bool{
	".mkv":  true,
	".mp4":  true,
	".avi":  true,
	".mov":  true,
	".m4v":  true,
	".ts":   true,
	".m2ts": true,
	".wmv":  true,
	".flv":  true,
	".webm": true,
}

var yearPattern = regexp.MustCompile(`(?i)(19\d{2}|20\d{2})`)
var bracketPattern = regexp.MustCompile(`[\[\]\(\)\{\}]`)
var noisyTokens = regexp.MustCompile(`(?i)\b(2160p|1080p|720p|480p|uhd|hdr|hdr10|dv|dolby|vision|bluray|blu-ray|bdrip|webrip|web-dl|hdtv|remux|x264|x265|h264|h265|hevc|aac|dts|truehd|atmos|proper|repack)\b`)

type Library struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Path string `json:"path"`
	Type string `json:"type"`
}

type Item struct {
	ID          string `json:"id"`
	LibraryID   string `json:"libraryId"`
	Kind        string `json:"kind"`
	Path        string `json:"path"`
	Dir         string `json:"dir"`
	FileName    string `json:"fileName"`
	TitleGuess  string `json:"titleGuess"`
	YearGuess   string `json:"yearGuess,omitempty"`
	HasNFO      bool   `json:"hasNfo"`
	HasPoster   bool   `json:"hasPoster"`
	MatchedID   int    `json:"matchedId,omitempty"`
	MatchedName string `json:"matchedName,omitempty"`
}

type RenamePreview struct {
	SourceFile string `json:"sourceFile"`
	TargetFile string `json:"targetFile"`
	SourceDir  string `json:"sourceDir"`
	TargetDir  string `json:"targetDir"`
}

func ScanLibrary(library Library) ([]Item, error) {
	var items []Item
	root := filepath.Clean(library.Path)
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			if shouldSkipDir(info.Name()) && path != root {
				return filepath.SkipDir
			}
			return nil
		}
		if !videoExts[strings.ToLower(filepath.Ext(info.Name()))] {
			return nil
		}
		title, year := GuessTitleYear(info.Name())
		dir := filepath.Dir(path)
		item := Item{
			ID:         stableID(path),
			LibraryID:  library.ID,
			Kind:       library.Type,
			Path:       path,
			Dir:        dir,
			FileName:   info.Name(),
			TitleGuess: title,
			YearGuess:  year,
			HasNFO:     exists(filepath.Join(dir, "movie.nfo")) || exists(strings.TrimSuffix(path, filepath.Ext(path))+".nfo"),
			HasPoster:  exists(filepath.Join(dir, "poster.jpg")) || exists(filepath.Join(dir, "folder.jpg")),
		}
		items = append(items, item)
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(items, func(i, j int) bool { return items[i].Path < items[j].Path })
	return items, nil
}

func GuessTitleYear(name string) (string, string) {
	base := strings.TrimSuffix(name, filepath.Ext(name))
	year := ""
	if match := yearPattern.FindString(base); match != "" {
		year = match
		base = base[:strings.Index(base, match)]
	}
	base = strings.ReplaceAll(base, ".", " ")
	base = strings.ReplaceAll(base, "_", " ")
	base = strings.ReplaceAll(base, "-", " ")
	base = bracketPattern.ReplaceAllString(base, " ")
	base = noisyTokens.ReplaceAllString(base, " ")
	base = strings.Join(strings.Fields(base), " ")
	if base == "" {
		base = strings.TrimSuffix(name, filepath.Ext(name))
	}
	return base, year
}

func BuildMovieRename(item Item, title string, year string, movieID int, pattern string) RenamePreview {
	if pattern == "" {
		pattern = "{title} ({year}) {tmdb-{tmdbid}}"
	}
	safeTitle := safePath(title)
	targetBase := strings.TrimSpace(pattern)
	targetBase = strings.ReplaceAll(targetBase, "{title}", safeTitle)
	targetBase = strings.ReplaceAll(targetBase, "{year}", year)
	targetBase = strings.ReplaceAll(targetBase, "{tmdbid}", fmt.Sprintf("%d", movieID))
	targetBase = strings.ReplaceAll(targetBase, "{tmdb-{tmdbid}}", fmt.Sprintf("{tmdb-%d}", movieID))
	targetBase = strings.Join(strings.Fields(targetBase), " ")
	ext := filepath.Ext(item.Path)
	targetDir := filepath.Join(filepath.Dir(item.Dir), targetBase)
	return RenamePreview{
		SourceFile: item.Path,
		TargetFile: filepath.Join(targetDir, targetBase+ext),
		SourceDir:  item.Dir,
		TargetDir:  targetDir,
	}
}

func ApplyRename(preview RenamePreview) error {
	if preview.SourceFile == preview.TargetFile {
		return nil
	}
	if exists(preview.TargetFile) {
		return fmt.Errorf("target already exists: %s", preview.TargetFile)
	}
	if err := os.MkdirAll(preview.TargetDir, 0755); err != nil {
		return err
	}
	return os.Rename(preview.SourceFile, preview.TargetFile)
}

func shouldSkipDir(name string) bool {
	switch strings.ToLower(name) {
	case "@eadir", ".snapshot", ".recycle", "#recycle", "sample", "extras", "featurettes":
		return true
	default:
		return false
	}
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func safePath(value string) string {
	replacer := strings.NewReplacer("/", " ", "\\", " ", ":", " -", "*", "", "?", "", "\"", "'", "<", "", ">", "", "|", " ")
	return strings.Join(strings.Fields(replacer.Replace(value)), " ")
}

func stableID(path string) string {
	return fmt.Sprintf("%x", strings.ToLower(filepath.Clean(path)))
}
