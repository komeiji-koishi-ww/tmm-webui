package media

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

var ErrScanCanceled = errors.New("scan canceled")

var videoExts = map[string]bool{
	".mkv": true, ".mp4": true, ".avi": true, ".mov": true, ".m4v": true,
	".ts": true, ".m2ts": true, ".wmv": true, ".flv": true, ".webm": true,
}

var subtitleExts = map[string]bool{
	".aqt": true, ".ass": true, ".idx": true, ".smi": true, ".srt": true, ".ssa": true,
	".sub": true, ".sup": true, ".vtt": true,
}

var imageExts = map[string]bool{
	".jpg": true, ".jpeg": true, ".png": true, ".webp": true, ".gif": true, ".tbn": true,
}

var yearPattern = regexp.MustCompile(`(?i)(19\d{2}|20\d{2})`)
var bracketPattern = regexp.MustCompile(`[\[\]\(\)\{\}]`)
var noisyTokens = regexp.MustCompile(`(?i)\b(2160p|1080p|720p|480p|uhd|hdr|hdr10|dv|dolby|vision|bluray|blu-ray|bdrip|webrip|web-dl|hdtv|remux|x264|x265|h264|h265|hevc|aac|dts|truehd|atmos|proper|repack)\b`)
var datePatternYMD = regexp.MustCompile(`(?i)(19\d{2}|20\d{2})[.-](\d{2})[.-](\d{2})`)
var datePatternMDY = regexp.MustCompile(`(?i)(\d{2})[.-](\d{2})[.-](19\d{2}|20\d{2})`)
var seasonLongPattern = regexp.MustCompile(`(?i)(staffel|season|saison|series|temporada|第)[\s_.-]?(\d{1,4})`)
var seasonOnlyPattern = regexp.MustCompile(`(?i)(^|[\s_.-])s[\s_.-]?(\d{1,4})([\s_.-]|$)`)
var seasonFolderPattern = regexp.MustCompile(`(?i)^(season|staffel|saison|series|temporada|第)?[\s_.-]*(\d{1,4})(季)?$`)
var seasonMultiEpisodePattern = regexp.MustCompile(`(?i)s(\d{1,4})[\s_]?((?:[epx_.-]+\d{1,4})+)`)
var xMultiEpisodePattern = regexp.MustCompile(`(?i)(\d{1,4})((?:x\d{1,4})+)`)
var episodeTokenPattern = regexp.MustCompile(`(?i)(?:episode|ep|e|x|_|-)+[\s_.-]?(\d{1,4})`)
var romanPartPattern = regexp.MustCompile(`(?i)(?:part|pt)[._\s]+([mdclxvi]+)`)
var digitsPattern = regexp.MustCompile(`^\d{1,4}$`)

type Library struct {
	ID    string   `json:"id"`
	Name  string   `json:"name"`
	Path  string   `json:"path,omitempty"`
	Paths []string `json:"paths"`
	Type  string   `json:"type"`
}

type Item struct {
	ID          string `json:"id"`
	LibraryID   string `json:"libraryId"`
	SourcePath  string `json:"sourcePath"`
	Kind        string `json:"kind"`
	Path        string `json:"path"`
	Dir         string `json:"dir"`
	FileName    string `json:"fileName"`
	TitleGuess  string `json:"titleGuess"`
	YearGuess   string `json:"yearGuess,omitempty"`
	ShowGuess   string `json:"showGuess,omitempty"`
	Season      int    `json:"season,omitempty"`
	Episode     int    `json:"episode,omitempty"`
	Episodes    []int  `json:"episodes,omitempty"`
	AirDate     string `json:"airDate,omitempty"`
	MediaType   string `json:"mediaType"`
	HasNFO      bool   `json:"hasNfo"`
	HasPoster   bool   `json:"hasPoster"`
	HasFanart   bool   `json:"hasFanart"`
	HasSubtitle bool   `json:"hasSubtitle"`
	MatchedID   int    `json:"matchedId,omitempty"`
	MatchedName string `json:"matchedName,omitempty"`
}

type RenamePreview struct {
	SourceFile string `json:"sourceFile"`
	TargetFile string `json:"targetFile"`
	SourceDir  string `json:"sourceDir"`
	TargetDir  string `json:"targetDir"`
}

type RenameTemplateData struct {
	Title        string
	ShowTitle    string
	EpisodeTitle string
	Year         string
	TMDBID       int
	Season       int
	Episode      int
	AirDate      string
}

type EpisodeMatch struct {
	Season   int
	Episodes []int
	AirDate  string
}

type ScanProgress struct {
	SourcePath   string `json:"sourcePath"`
	CurrentPath  string `json:"currentPath"`
	VisitedFiles int    `json:"visitedFiles"`
	FoundItems   int    `json:"foundItems"`
	Item         *Item  `json:"-"`
}

func (m EpisodeMatch) PrimaryEpisode() int {
	if len(m.Episodes) == 0 {
		return 0
	}
	return m.Episodes[0]
}

func ScanLibrary(library Library) ([]Item, error) {
	return ScanLibraryWithProgress(library, nil)
}

func ScanLibraryWithProgress(library Library, progress func(ScanProgress)) ([]Item, error) {
	return ScanLibraryWithCancel(library, progress, nil)
}

func ScanLibraryWithCancel(library Library, progress func(ScanProgress), shouldCancel func() bool) ([]Item, error) {
	var items []Item
	visited := 0
	library.Paths = NormalizePaths(library)
	for _, source := range library.Paths {
		root := filepath.Clean(source)
		err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			if shouldCancel != nil && shouldCancel() {
				return ErrScanCanceled
			}
			if err != nil {
				return nil
			}
			if info.IsDir() {
				if ShouldSkipDir(path, info.Name(), library.Type) && path != root {
					return filepath.SkipDir
				}
				return nil
			}
			visited++
			if ClassifyMediaFile(path) != "VIDEO" {
				if progress != nil && visited%50 == 0 {
					progress(ScanProgress{SourcePath: root, CurrentPath: path, VisitedFiles: visited, FoundItems: len(items)})
				}
				return nil
			}
			item := NewItem(library, root, path, info.Name())
			items = append(items, item)
			if progress != nil {
				progress(ScanProgress{SourcePath: root, CurrentPath: path, VisitedFiles: visited, FoundItems: len(items), Item: &item})
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
		if progress != nil {
			progress(ScanProgress{SourcePath: root, CurrentPath: root, VisitedFiles: visited, FoundItems: len(items)})
		}
	}
	sort.Slice(items, func(i, j int) bool { return items[i].Path < items[j].Path })
	return items, nil
}

func NormalizePaths(library Library) []string {
	seen := map[string]bool{}
	var paths []string
	for _, path := range append(library.Paths, library.Path) {
		path = strings.TrimSpace(path)
		if path == "" {
			continue
		}
		clean := filepath.Clean(path)
		if seen[clean] {
			continue
		}
		seen[clean] = true
		paths = append(paths, clean)
	}
	return paths
}

func NewItem(library Library, sourcePath string, path string, fileName string) Item {
	title, year := GuessTitleYear(fileName)
	dir := filepath.Dir(path)
	item := Item{
		ID:          stableID(path),
		LibraryID:   library.ID,
		SourcePath:  sourcePath,
		Kind:        library.Type,
		Path:        path,
		Dir:         dir,
		FileName:    fileName,
		TitleGuess:  title,
		YearGuess:   year,
		MediaType:   ClassifyMediaFile(path),
		HasNFO:      exists(filepath.Join(dir, "movie.nfo")) || exists(strings.TrimSuffix(path, filepath.Ext(path))+".nfo"),
		HasPoster:   hasAny(dir, []string{"poster.jpg", "folder.jpg", strings.TrimSuffix(fileName, filepath.Ext(fileName)) + "-poster.jpg"}),
		HasFanart:   hasAny(dir, []string{"fanart.jpg", "backdrop.jpg", strings.TrimSuffix(fileName, filepath.Ext(fileName)) + "-fanart.jpg"}),
		HasSubtitle: hasSubtitle(path),
	}
	if library.Type == "tvshow" {
		item.ShowGuess = GuessShowName(sourcePath, path)
		item.TitleGuess = item.ShowGuess
		match := GuessSeasonEpisode(sourcePath, path, item.ShowGuess)
		item.Season = match.Season
		item.Episode = match.PrimaryEpisode()
		item.Episodes = match.Episodes
		item.AirDate = match.AirDate
		item.HasNFO = exists(strings.TrimSuffix(path, filepath.Ext(path)) + ".nfo")
		showDir := showRoot(sourcePath, path)
		item.HasPoster = hasAny(showDir, []string{"poster.jpg", "folder.jpg"})
		item.HasFanart = hasAny(showDir, []string{"fanart.jpg", "backdrop.jpg"})
	}
	return item
}

func ClassifyMediaFile(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	base := strings.ToLower(strings.TrimSuffix(filepath.Base(path), filepath.Ext(path)))
	dir := strings.ToLower(filepath.Base(filepath.Dir(path)))
	if videoExts[ext] {
		if strings.Contains(base, "sample") || dir == "sample" || dir == "samples" {
			return "SAMPLE"
		}
		if strings.Contains(base, "trailer") || dir == "trailer" || dir == "trailers" {
			return "TRAILER"
		}
		return "VIDEO"
	}
	if subtitleExts[ext] {
		return "SUBTITLE"
	}
	if ext == ".nfo" {
		return "NFO"
	}
	if ext == ".vsmeta" {
		return "VSMETA"
	}
	if imageExts[ext] {
		switch {
		case strings.Contains(base, "poster") || base == "folder" || base == "movie" || strings.Contains(base, "cover"):
			return "POSTER"
		case strings.Contains(base, "fanart") || strings.Contains(base, "backdrop"):
			return "FANART"
		case strings.Contains(base, "banner"):
			return "BANNER"
		case strings.Contains(base, "clearlogo") || base == "logo":
			return "CLEARLOGO"
		case strings.Contains(base, "clearart"):
			return "CLEARART"
		case strings.Contains(base, "disc") || strings.Contains(base, "cdart"):
			return "DISC"
		case strings.Contains(base, "keyart"):
			return "KEYART"
		default:
			return "GRAPHIC"
		}
	}
	return "UNKNOWN"
}

func GuessTitleYear(name string) (string, string) {
	base := strings.TrimSuffix(name, filepath.Ext(name))
	year := ""
	if matches := yearPattern.FindAllString(base, -1); len(matches) > 0 {
		match := matches[len(matches)-1]
		year = match
		base = base[:strings.LastIndex(base, match)]
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

func GuessShowName(sourcePath string, path string) string {
	rel, err := filepath.Rel(sourcePath, path)
	if err != nil || strings.HasPrefix(rel, "..") {
		return filepath.Base(filepath.Dir(path))
	}
	parts := strings.Split(rel, string(filepath.Separator))
	if len(parts) > 1 && parts[0] != "." {
		return strings.TrimSpace(parts[0])
	}
	return filepath.Base(filepath.Dir(path))
}

func GuessSeasonEpisode(sourcePath string, path string, showName string) EpisodeMatch {
	rel, err := filepath.Rel(sourcePath, path)
	if err != nil {
		rel = filepath.Base(path)
	}
	fileOnly := filepath.Base(path)
	match := detectEpisode(fileOnly, showName)
	if season := seasonFromRelativePath(rel); season > 0 {
		match.Season = season
	}
	if len(match.Episodes) == 0 && match.AirDate == "" {
		match = detectEpisode(rel, showName)
	} else if match.Season == 0 {
		fromPath := detectEpisode(rel, showName)
		match.Season = fromPath.Season
	}
	if len(match.Episodes) > 0 && match.Season == 0 {
		match.Season = 1
	}
	return match
}

func seasonFromRelativePath(rel string) int {
	dir := filepath.Dir(rel)
	if dir == "." || dir == "" {
		return 0
	}
	parts := strings.Split(dir, string(filepath.Separator))
	for i := len(parts) - 1; i >= 0; i-- {
		part := strings.TrimSpace(parts[i])
		for _, pattern := range []*regexp.Regexp{seasonLongPattern, seasonFolderPattern, seasonOnlyPattern} {
			if match := pattern.FindStringSubmatch(part); len(match) > 0 {
				for j := len(match) - 1; j >= 1; j-- {
					if season, err := strconv.Atoi(match[j]); err == nil && season >= 0 {
						return season
					}
				}
			}
		}
	}
	return 0
}

func detectEpisode(name string, showName string) EpisodeMatch {
	if date := parseAirDate(name); date != "" {
		season, _ := strconv.Atoi(date[:4])
		return EpisodeMatch{Season: season, AirDate: date}
	}
	clean := normalizeEpisodeName(name, showName)
	result := EpisodeMatch{}
	if date := parseAirDate(clean); date != "" {
		result.AirDate = date
		result.Season, _ = strconv.Atoi(date[:4])
		return result
	}
	result = parseSeasonAndEpisodes(result, clean)
	if len(result.Episodes) > 0 {
		return result
	}
	result = parseSeasonOnly(result, clean)
	result = parseEpisodeOnly(result, clean)
	if len(result.Episodes) > 0 {
		return result
	}
	result = parseRomanEpisode(result, clean)
	if len(result.Episodes) > 0 {
		return result
	}
	result = parseBareNumbers(result, clean)
	return result
}

func normalizeEpisodeName(name string, showName string) string {
	name = strings.TrimSuffix(name, filepath.Ext(name))
	name = strings.ReplaceAll(name, string(filepath.Separator), " ")
	name = strings.ReplaceAll(name, "\\", " ")
	name = strings.ReplaceAll(name, "/", " ")
	name = strings.ReplaceAll(name, "_", " ")
	name = strings.ReplaceAll(name, ".", " ")
	name = strings.ReplaceAll(name, "-", " ")
	name = bracketPattern.ReplaceAllString(name, " ")
	name = noisyTokens.ReplaceAllString(name, " ")
	if showName != "" {
		name = regexp.MustCompile(`(?i)`+regexp.QuoteMeta(showName)).ReplaceAllString(name, " ")
	}
	return strings.Join(strings.Fields(name), " ")
}

func parseAirDate(name string) string {
	if match := datePatternYMD.FindStringSubmatch(name); len(match) == 4 {
		value := fmt.Sprintf("%s-%s-%s", match[1], match[2], match[3])
		if _, err := time.Parse("2006-01-02", value); err == nil {
			return value
		}
	}
	if match := datePatternMDY.FindStringSubmatch(name); len(match) == 4 {
		value := fmt.Sprintf("%s-%s-%s", match[3], match[1], match[2])
		if _, err := time.Parse("2006-01-02", value); err == nil {
			return value
		}
	}
	return ""
}

func parseSeasonAndEpisodes(result EpisodeMatch, name string) EpisodeMatch {
	for _, pattern := range []*regexp.Regexp{seasonMultiEpisodePattern, xMultiEpisodePattern} {
		for _, match := range pattern.FindAllStringSubmatch(name, -1) {
			if len(match) < 3 {
				continue
			}
			season, _ := strconv.Atoi(match[1])
			if result.Season == 0 {
				result.Season = season
			}
			if result.Season != season {
				continue
			}
			for _, epMatch := range episodeTokenPattern.FindAllStringSubmatch(match[2], -1) {
				if len(epMatch) >= 2 {
					addEpisode(&result, epMatch[1])
				}
			}
		}
	}
	return result
}

func parseSeasonOnly(result EpisodeMatch, name string) EpisodeMatch {
	for _, pattern := range []*regexp.Regexp{seasonLongPattern, seasonOnlyPattern, seasonFolderPattern} {
		if match := pattern.FindStringSubmatch(name); len(match) > 0 {
			for i := len(match) - 1; i >= 1; i-- {
				if season, err := strconv.Atoi(match[i]); err == nil && season >= 0 {
					result.Season = season
					return result
				}
			}
		}
	}
	return result
}

func parseEpisodeOnly(result EpisodeMatch, name string) EpisodeMatch {
	for _, match := range episodeTokenPattern.FindAllStringSubmatch(" "+name+" ", -1) {
		if len(match) >= 2 {
			addEpisode(&result, match[1])
		}
	}
	return result
}

func parseRomanEpisode(result EpisodeMatch, name string) EpisodeMatch {
	if match := romanPartPattern.FindStringSubmatch(name); len(match) == 2 {
		if episode := romanToInt(match[1]); episode > 0 {
			result.Episodes = append(result.Episodes, episode)
		}
	}
	return result
}

func parseBareNumbers(result EpisodeMatch, name string) EpisodeMatch {
	parts := strings.FieldsFunc(name, func(r rune) bool {
		return r == ' ' || r == '.' || r == '_' || r == '-' || r == '|'
	})
	for i := len(parts) - 1; i >= 0; i-- {
		part := parts[i]
		if !digitsPattern.MatchString(part) {
			continue
		}
		switch len(part) {
		case 4:
			season, _ := strconv.Atoi(part[:2])
			episode, _ := strconv.Atoi(part[2:])
			if result.Season == season && episode > 0 {
				result.Episodes = append(result.Episodes, episode)
				return result
			}
		case 3:
			season, _ := strconv.Atoi(part[:1])
			episode, _ := strconv.Atoi(part[1:])
			if episode > 0 {
				result.Season = season
				result.Episodes = append(result.Episodes, episode)
				return result
			}
		case 1, 2:
			episode, _ := strconv.Atoi(part)
			if episode > 0 {
				result.Episodes = append(result.Episodes, episode)
				return result
			}
		}
	}
	return result
}

func addEpisode(result *EpisodeMatch, value string) {
	episode, err := strconv.Atoi(value)
	if err != nil || episode < 0 {
		return
	}
	for _, existing := range result.Episodes {
		if existing == episode {
			return
		}
	}
	result.Episodes = append(result.Episodes, episode)
}

func romanToInt(value string) int {
	numbers := map[rune]int{'I': 1, 'V': 5, 'X': 10, 'L': 50, 'C': 100, 'D': 500, 'M': 1000}
	value = strings.ToUpper(value)
	total, previous := 0, 0
	for i := len(value) - 1; i >= 0; i-- {
		current := numbers[rune(value[i])]
		if current < previous {
			total -= current
		} else {
			total += current
			previous = current
		}
	}
	return total
}

func BuildMovieRename(item Item, title string, year string, movieID int, pattern string) RenamePreview {
	if pattern == "" {
		pattern = "{title} ({year}) {tmdb-{tmdbid}}"
	}
	return BuildMovieRenameWithPatterns(item, title, year, movieID, pattern, pattern)
}

func BuildMovieRenameWithPatterns(item Item, title string, year string, movieID int, folderPattern string, filePattern string) RenamePreview {
	if strings.TrimSpace(folderPattern) == "" {
		folderPattern = "{title} ({year})"
	}
	if strings.TrimSpace(filePattern) == "" {
		filePattern = folderPattern
	}
	data := RenameTemplateData{
		Title:  title,
		Year:   year,
		TMDBID: movieID,
	}
	targetFolder := renderRenamePattern(folderPattern, data, "Movie")
	targetFileBase := renderRenamePattern(filePattern, data, targetFolder)
	ext := filepath.Ext(item.Path)
	targetDir := filepath.Join(filepath.Dir(item.Dir), targetFolder)
	return RenamePreview{
		SourceFile: item.Path,
		TargetFile: filepath.Join(targetDir, targetFileBase+ext),
		SourceDir:  item.Dir,
		TargetDir:  targetDir,
	}
}

func BuildTVShowRename(item Item, showTitle string, episodeTitle string, year string, showID int, showFolderPattern string, seasonPattern string, filePattern string) RenamePreview {
	if strings.TrimSpace(showFolderPattern) == "" {
		showFolderPattern = "{showTitle}"
	}
	if strings.TrimSpace(seasonPattern) == "" {
		seasonPattern = "Season {seasonNr2}"
	}
	if strings.TrimSpace(filePattern) == "" {
		filePattern = "{showTitle} - S{seasonNr2}E{episodeNr2} - {title}"
	}
	if strings.TrimSpace(showTitle) == "" {
		showTitle = item.ShowGuess
	}
	if strings.TrimSpace(episodeTitle) == "" {
		episodeTitle = strings.TrimSuffix(item.FileName, filepath.Ext(item.FileName))
	}
	data := RenameTemplateData{
		Title:        episodeTitle,
		ShowTitle:    showTitle,
		EpisodeTitle: episodeTitle,
		Year:         year,
		TMDBID:       showID,
		Season:       item.Season,
		Episode:      item.Episode,
		AirDate:      item.AirDate,
	}
	showFolder := renderRenamePattern(showFolderPattern, data, "TV Show")
	seasonFolder := renderRenamePattern(seasonPattern, data, "Season")
	fileBase := renderRenamePattern(filePattern, data, strings.TrimSuffix(item.FileName, filepath.Ext(item.FileName)))
	root := strings.TrimSpace(item.SourcePath)
	if root == "" {
		root = filepath.Dir(item.Dir)
		if item.Season > 0 || item.Episode > 0 {
			root = filepath.Dir(root)
		}
	}
	targetDir := filepath.Join(root, showFolder, seasonFolder)
	ext := filepath.Ext(item.Path)
	return RenamePreview{
		SourceFile: item.Path,
		TargetFile: filepath.Join(targetDir, fileBase+ext),
		SourceDir:  item.Dir,
		TargetDir:  targetDir,
	}
}

func renderRenamePattern(pattern string, data RenameTemplateData, fallback string) string {
	value := strings.TrimSpace(pattern)
	season := strconv.Itoa(data.Season)
	episode := strconv.Itoa(data.Episode)
	replacements := map[string]string{
		"{title}":         safePath(data.Title),
		"{showTitle}":     safePath(data.ShowTitle),
		"{episodeTitle}":  safePath(data.EpisodeTitle),
		"{year}":          data.Year,
		"{tmdb-{tmdbid}}": fmt.Sprintf("{tmdb-%d}", data.TMDBID),
		"{tmdbid}":        fmt.Sprintf("%d", data.TMDBID),
		"{seasonNr}":      season,
		"{seasonNr2}":     fmt.Sprintf("%02d", data.Season),
		"{episodeNr}":     episode,
		"{episodeNr2}":    fmt.Sprintf("%02d", data.Episode),
		"{aired}":         data.AirDate,
	}
	keys := make([]string, 0, len(replacements))
	for key := range replacements {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool { return len(keys[i]) > len(keys[j]) })
	for _, key := range keys {
		value = strings.ReplaceAll(value, key, replacements[key])
	}
	value = cleanGeneratedName(value)
	if value == "" {
		return cleanGeneratedName(fallback)
	}
	return value
}

func cleanGeneratedName(value string) string {
	replacer := strings.NewReplacer("/", " ", "\\", " ", ":", " -", "*", "", "?", "", "\"", "'", "<", "", ">", "", "|", " ")
	return strings.Join(strings.Fields(replacer.Replace(value)), " ")
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

func ShouldSkipDir(path string, name string, kind string) bool {
	if exists(filepath.Join(path, ".tmmignore")) {
		return true
	}
	upper := strings.ToUpper(name)
	if strings.HasPrefix(name, ".") && name != ".45" {
		return true
	}
	switch upper {
	case ".", "..", "CERTIFICATE", "$RECYCLE.BIN", "RECYCLER", "SYSTEM VOLUME INFORMATION", "@EADIR", "ADV_OBJ", "PLEX VERSIONS", "LOST.DIR":
		return true
	}
	if kind == "tvshow" && upper == "EXTRATHUMB" {
		return true
	}
	return false
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func hasAny(dir string, names []string) bool {
	for _, name := range names {
		if exists(filepath.Join(dir, name)) {
			return true
		}
	}
	return false
}

func hasSubtitle(path string) bool {
	base := strings.TrimSuffix(path, filepath.Ext(path))
	for ext := range subtitleExts {
		if exists(base + ext) {
			return true
		}
	}
	return false
}

func safePath(value string) string {
	replacer := strings.NewReplacer("/", " ", "\\", " ", ":", " -", "*", "", "?", "", "\"", "'", "<", "", ">", "", "|", " ")
	return strings.Join(strings.Fields(replacer.Replace(value)), " ")
}

func stableID(path string) string {
	return fmt.Sprintf("%x", strings.ToLower(filepath.Clean(path)))
}

func showRoot(sourcePath string, path string) string {
	rel, err := filepath.Rel(sourcePath, path)
	if err != nil || strings.HasPrefix(rel, "..") {
		return filepath.Dir(path)
	}
	parts := strings.Split(rel, string(filepath.Separator))
	if len(parts) == 0 || parts[0] == "." {
		return filepath.Dir(path)
	}
	return filepath.Join(sourcePath, parts[0])
}
