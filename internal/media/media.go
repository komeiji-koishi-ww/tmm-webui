package media

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
	"unicode"

	"tmmweb/internal/nfo"
)

var ErrScanCanceled = errors.New("scan canceled")

const (
	defaultMovieRenamePathPattern  = "${title} ${- ,edition,} (${year}) ${videoFormat} - ${fileSize}"
	defaultMovieRenameFilePattern  = "${title} ${- ,edition,} (${year}) ${videoFormat} ${audioCodec} ${fileSize}"
	defaultTVShowRenamePathPattern = "${showTitle} (${showYear})"
	defaultTVSeasonRenamePattern   = "Season ${seasonNr2}"
	defaultTVEpisodeRenamePattern  = "${showTitle}.S${seasonNr2}E${episodeNr2}.${title}"
)

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
var unresolvedRenameTokenPattern = regexp.MustCompile(`\$\{[^}]+\}`)
var videoFormatPattern = regexp.MustCompile(`(?i)(4320p|2160p|1080p|720p|576p|540p|480p|4k|8k|uhd)`)
var audioCodecPatterns = []struct {
	pattern *regexp.Regexp
	value   string
}{
	{regexp.MustCompile(`(?i)(true[ ._-]?hd)(?:[ ._-]?atmos)?`), "TrueHD"},
	{regexp.MustCompile(`(?i)(dts[ ._-]?hd[ ._-]?ma|dts[ ._-]?xll)`), "DTS-HD MA"},
	{regexp.MustCompile(`(?i)(dts[ ._-]?x)`), "DTS:X"},
	{regexp.MustCompile(`(?i)(e[ ._-]?ac[ ._-]?3|ddp|dolby[ ._-]?digital[ ._-]?plus)`), "EAC3"},
	{regexp.MustCompile(`(?i)(ac[ ._-]?3|dolby[ ._-]?digital)`), "AC3"},
	{regexp.MustCompile(`(?i)(dts)`), "DTS"},
	{regexp.MustCompile(`(?i)(flac)`), "FLAC"},
	{regexp.MustCompile(`(?i)(opus)`), "OPUS"},
	{regexp.MustCompile(`(?i)(aac)`), "AAC"},
	{regexp.MustCompile(`(?i)(mp3)`), "MP3"},
	{regexp.MustCompile(`(?i)(pcm|lpcm)`), "PCM"},
}

var (
	ffprobeOnce sync.Once
	ffprobePath string
)

type Library struct {
	ID    string   `json:"id"`
	Name  string   `json:"name"`
	Path  string   `json:"path,omitempty"`
	Paths []string `json:"paths"`
	Type  string   `json:"type"`
}

type Item struct {
	ID               string           `json:"id"`
	LibraryID        string           `json:"libraryId"`
	SourcePath       string           `json:"sourcePath"`
	Kind             string           `json:"kind"`
	Path             string           `json:"path"`
	Dir              string           `json:"dir"`
	FileName         string           `json:"fileName"`
	TitleGuess       string           `json:"titleGuess"`
	YearGuess        string           `json:"yearGuess,omitempty"`
	Original         string           `json:"originalTitle,omitempty"`
	Overview         string           `json:"overview,omitempty"`
	Runtime          int              `json:"runtime,omitempty"`
	Rating           float64          `json:"rating,omitempty"`
	ShowRating       float64          `json:"showRating,omitempty"`
	Genres           []string         `json:"genres,omitempty"`
	Premiered        string           `json:"premiered,omitempty"`
	DateAdded        string           `json:"dateAdded,omitempty"`
	ModTimeUnix      int64            `json:"modTimeUnix,omitempty"`
	NFOModTimeUnix   int64            `json:"nfoModTimeUnix,omitempty"`
	FileSize         string           `json:"fileSize,omitempty"`
	FileSizeBytes    int64            `json:"fileSizeBytes,omitempty"`
	VideoFormat      string           `json:"videoFormat,omitempty"`
	AudioCodec       string           `json:"audioCodec,omitempty"`
	VideoStreams     []VideoStream    `json:"videoStreams,omitempty"`
	AudioStreams     []AudioStream    `json:"audioStreams,omitempty"`
	SubtitleStreams  []SubtitleStream `json:"subtitleStreams,omitempty"`
	MediaInfoScanned bool             `json:"mediaInfoScanned,omitempty"`
	IMDBID           string           `json:"imdbId,omitempty"`
	ShowGuess        string           `json:"showGuess,omitempty"`
	Season           int              `json:"season,omitempty"`
	Episode          int              `json:"episode,omitempty"`
	Episodes         []int            `json:"episodes,omitempty"`
	AirDate          string           `json:"airDate,omitempty"`
	MediaType        string           `json:"mediaType"`
	HasNFO           bool             `json:"hasNfo"`
	HasPoster        bool             `json:"hasPoster"`
	HasFanart        bool             `json:"hasFanart"`
	HasSubtitle      bool             `json:"hasSubtitle"`
	MatchedID        int              `json:"matchedId,omitempty"`
	MatchedName      string           `json:"matchedName,omitempty"`
}

type RenamePreview struct {
	SourceFile string            `json:"sourceFile"`
	TargetFile string            `json:"targetFile"`
	SourceDir  string            `json:"sourceDir"`
	TargetDir  string            `json:"targetDir"`
	Operations []RenameOperation `json:"operations,omitempty"`
}

type RenameOperation struct {
	Kind   string `json:"kind"`
	Source string `json:"source"`
	Target string `json:"target"`
}

type RenameTemplateData struct {
	Title        string
	ShowTitle    string
	EpisodeTitle string
	Year         string
	TMDBID       int
	VideoFormat  string
	AudioCodec   string
	FileSize     string
	Season       int
	Episode      int
	AirDate      string
}

type RenameOptions struct {
	SpaceSubstitution               bool
	SpaceReplacement                string
	ColonReplacement                string
	ColonReplacementDefined         bool
	ASCIIReplacement                bool
	FirstCharacterNumberReplacement string
}

type MediaInfoProbe struct {
	FileSizeBytes   int64
	FileSize        string
	VideoFormat     string
	AudioCodec      string
	VideoStreams    []VideoStream
	AudioStreams    []AudioStream
	SubtitleStreams []SubtitleStream
	Scanned         bool
}

type VideoStream struct {
	Codec           string  `json:"codec,omitempty"`
	Aspect          float64 `json:"aspect,omitempty"`
	Width           int     `json:"width,omitempty"`
	Height          int     `json:"height,omitempty"`
	DurationSeconds int     `json:"durationSeconds,omitempty"`
	StereoMode      string  `json:"stereoMode,omitempty"`
	HDRType         string  `json:"hdrType,omitempty"`
}

type AudioStream struct {
	Codec    string `json:"codec,omitempty"`
	Language string `json:"language,omitempty"`
	Channels int    `json:"channels,omitempty"`
}

type SubtitleStream struct {
	Language string `json:"language,omitempty"`
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

type ScanOptions struct {
	Existing       map[string]Item
	Workers        int
	ProbeMediaInfo bool
}

type scanJob struct {
	source string
	path   string
	info   os.FileInfo
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
	return ScanLibraryWithOptions(library, ScanOptions{}, progress, shouldCancel)
}

func ScanLibraryWithOptions(library Library, options ScanOptions, progress func(ScanProgress), shouldCancel func() bool) ([]Item, error) {
	var items []Item
	visited := 0
	found := 0
	if options.Workers <= 0 {
		options.Workers = defaultScanWorkers()
	}
	library.Paths = NormalizePaths(library)
	for _, source := range library.Paths {
		root := filepath.Clean(source)
		jobs := make(chan scanJob, options.Workers*2)
		results := make(chan Item, options.Workers*2)
		var wg sync.WaitGroup
		for i := 0; i < options.Workers; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for job := range jobs {
					if shouldCancel != nil && shouldCancel() {
						continue
					}
					results <- NewItemFromFileInfoWithOptions(library, job.source, job.path, job.info, options.ProbeMediaInfo)
				}
			}()
		}
		go func() {
			wg.Wait()
			close(results)
		}()
		drainResults := func() {
			for {
				select {
				case item, ok := <-results:
					if !ok {
						return
					}
					items = append(items, item)
					if progress != nil {
						progress(ScanProgress{SourcePath: root, CurrentPath: item.Path, VisitedFiles: visited, FoundItems: len(items), Item: &item})
					}
				default:
					return
				}
			}
		}
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
					progress(ScanProgress{SourcePath: root, CurrentPath: path, VisitedFiles: visited, FoundItems: found})
				}
				return nil
			}
			found++
			id := stableID(path)
			if existing, ok := options.Existing[id]; ok && itemUnchanged(existing, path, info, options.ProbeMediaInfo) {
				items = append(items, existing)
				if progress != nil && found%25 == 0 {
					progress(ScanProgress{SourcePath: root, CurrentPath: path, VisitedFiles: visited, FoundItems: found})
				}
				return nil
			}
			jobs <- scanJob{source: root, path: path, info: info}
			drainResults()
			return nil
		})
		close(jobs)
		for item := range results {
			items = append(items, item)
			if progress != nil {
				progress(ScanProgress{SourcePath: root, CurrentPath: item.Path, VisitedFiles: visited, FoundItems: len(items), Item: &item})
			}
		}
		if err != nil {
			return nil, err
		}
		if progress != nil {
			progress(ScanProgress{SourcePath: root, CurrentPath: root, VisitedFiles: visited, FoundItems: found})
		}
	}
	sort.Slice(items, func(i, j int) bool { return items[i].Path < items[j].Path })
	return items, nil
}

func defaultScanWorkers() int {
	if configured, err := strconv.Atoi(strings.TrimSpace(os.Getenv("TMMWEB_SCAN_WORKERS"))); err == nil && configured > 0 {
		if configured > 32 {
			return 32
		}
		return configured
	}
	workers := runtime.NumCPU()
	if workers < 2 {
		return 2
	}
	if workers > 8 {
		return 8
	}
	return workers
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
	info, _ := os.Stat(path)
	return newItem(library, sourcePath, path, fileName, info, false)
}

func NewItemFromFileInfo(library Library, sourcePath string, path string, info os.FileInfo) Item {
	return NewItemFromFileInfoWithOptions(library, sourcePath, path, info, false)
}

func NewItemFromFileInfoWithOptions(library Library, sourcePath string, path string, info os.FileInfo, probeMediaInfo bool) Item {
	fileName := filepath.Base(path)
	if info != nil {
		fileName = info.Name()
	}
	return newItem(library, sourcePath, path, fileName, info, probeMediaInfo)
}

func newItem(library Library, sourcePath string, path string, fileName string, info os.FileInfo, probeMediaInfo bool) Item {
	title, year := GuessTitleYear(fileName)
	dir := filepath.Dir(path)
	fileSizeBytes, fileSize, videoFormat, audioCodec := LightweightMediaInfo(path, info)
	item := Item{
		ID:            stableID(path),
		LibraryID:     library.ID,
		SourcePath:    sourcePath,
		Kind:          library.Type,
		Path:          path,
		Dir:           dir,
		FileName:      fileName,
		TitleGuess:    title,
		YearGuess:     year,
		DateAdded:     FileDate(info).UTC().Format(time.RFC3339),
		ModTimeUnix:   fileModTimeUnix(info),
		FileSize:      fileSize,
		FileSizeBytes: fileSizeBytes,
		VideoFormat:   videoFormat,
		AudioCodec:    audioCodec,
		MediaType:     ClassifyMediaFile(path),
		HasNFO:        exists(filepath.Join(dir, "movie.nfo")) || exists(strings.TrimSuffix(path, filepath.Ext(path))+".nfo"),
		HasPoster:     hasAny(dir, []string{"poster.jpg", "folder.jpg", strings.TrimSuffix(fileName, filepath.Ext(fileName)) + "-poster.jpg"}),
		HasFanart:     hasAny(dir, []string{"fanart.jpg", "backdrop.jpg", strings.TrimSuffix(fileName, filepath.Ext(fileName)) + "-fanart.jpg"}),
		HasSubtitle:   hasSubtitle(path),
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
		item.HasPoster = hasAny(showDir, tvShowPosterCandidates(item.Season, fileName))
		item.HasFanart = hasAny(showDir, tvShowFanartCandidates(item.Season, fileName))
	}
	if probeMediaInfo {
		var wg sync.WaitGroup
		nfoItem := item
		var probed MediaInfoProbe
		wg.Add(2)
		go func() {
			defer wg.Done()
			applyNFOSummary(&nfoItem)
		}()
		go func() {
			defer wg.Done()
			probed = ProbeMediaInfo(path, info)
		}()
		wg.Wait()
		item = nfoItem
		if probed.Scanned {
			item.FileSizeBytes = probed.FileSizeBytes
			item.FileSize = probed.FileSize
			if probed.VideoFormat != "" {
				item.VideoFormat = probed.VideoFormat
			}
			if probed.AudioCodec != "" {
				item.AudioCodec = probed.AudioCodec
			}
			item.VideoStreams = probed.VideoStreams
			item.AudioStreams = probed.AudioStreams
			item.SubtitleStreams = probed.SubtitleStreams
			item.MediaInfoScanned = true
		}
	} else {
		applyNFOSummary(&item)
	}
	return item
}

func itemUnchanged(existing Item, path string, info os.FileInfo, requireMediaInfo bool) bool {
	if existing.Path != path {
		return false
	}
	if existing.FileSizeBytes != fileSizeBytes(info) {
		return false
	}
	if existing.ModTimeUnix == 0 || existing.ModTimeUnix != fileModTimeUnix(info) {
		return false
	}
	if existing.NFOModTimeUnix != firstNFOModTime(existing) {
		return false
	}
	if requireMediaInfo && !existing.MediaInfoScanned {
		return false
	}
	return true
}

func fileModTimeUnix(info os.FileInfo) int64 {
	if info == nil {
		return 0
	}
	return info.ModTime().UnixNano()
}

func fileSizeBytes(info os.FileInfo) int64 {
	if info == nil {
		return 0
	}
	return info.Size()
}

func MergeScannedItem(existing Item, scanned Item) Item {
	if scanned.MatchedID == 0 && existing.MatchedID != 0 {
		scanned.MatchedID = existing.MatchedID
	}
	if scanned.MatchedName == "" && existing.MatchedName != "" {
		scanned.MatchedName = existing.MatchedName
	}
	if scanned.Original == "" && existing.Original != "" {
		scanned.Original = existing.Original
	}
	if scanned.Overview == "" && existing.Overview != "" {
		scanned.Overview = existing.Overview
	}
	if scanned.Runtime == 0 && existing.Runtime != 0 {
		scanned.Runtime = existing.Runtime
	}
	if scanned.Kind != "tvshow" && scanned.Rating == 0 && existing.Rating != 0 {
		scanned.Rating = existing.Rating
	}
	if scanned.ShowRating == 0 && existing.ShowRating != 0 {
		scanned.ShowRating = existing.ShowRating
	}
	if len(scanned.Genres) == 0 && len(existing.Genres) > 0 {
		scanned.Genres = existing.Genres
	}
	if scanned.Premiered == "" && existing.Premiered != "" {
		scanned.Premiered = existing.Premiered
	}
	if scanned.FileSize == "" && existing.FileSize != "" {
		scanned.FileSize = existing.FileSize
	}
	if scanned.FileSizeBytes == 0 && existing.FileSizeBytes != 0 {
		scanned.FileSizeBytes = existing.FileSizeBytes
	}
	if scanned.VideoFormat == "" && existing.VideoFormat != "" {
		scanned.VideoFormat = existing.VideoFormat
	}
	if scanned.AudioCodec == "" && existing.AudioCodec != "" {
		scanned.AudioCodec = existing.AudioCodec
	}
	if len(scanned.VideoStreams) == 0 && len(existing.VideoStreams) > 0 {
		scanned.VideoStreams = existing.VideoStreams
	}
	if len(scanned.AudioStreams) == 0 && len(existing.AudioStreams) > 0 {
		scanned.AudioStreams = existing.AudioStreams
	}
	if len(scanned.SubtitleStreams) == 0 && len(existing.SubtitleStreams) > 0 {
		scanned.SubtitleStreams = existing.SubtitleStreams
	}
	if scanned.IMDBID == "" && existing.IMDBID != "" {
		scanned.IMDBID = existing.IMDBID
	}
	if scanned.YearGuess == "" && existing.YearGuess != "" {
		scanned.YearGuess = existing.YearGuess
	}
	if scanned.TitleGuess == "" && existing.TitleGuess != "" {
		scanned.TitleGuess = existing.TitleGuess
	}
	return scanned
}

func LightweightMediaInfo(path string, info os.FileInfo) (int64, string, string, string) {
	var size int64
	if info != nil {
		size = info.Size()
	}
	text := strings.ToLower(strings.NewReplacer(".", " ", "_", " ", "-", " ", "[", " ", "]", " ", "(", " ", ")", " ").Replace(path))
	return size, FormatFileSize(size), inferVideoFormat(text), inferAudioCodec(text)
}

func ProbeMediaInfo(path string, info os.FileInfo) MediaInfoProbe {
	fileSizeBytes, fileSize, videoFormat, audioCodec := LightweightMediaInfo(path, info)
	probe := MediaInfoProbe{
		FileSizeBytes: fileSizeBytes,
		FileSize:      fileSize,
		VideoFormat:   videoFormat,
		AudioCodec:    audioCodec,
	}
	ffprobe := findFFProbe()
	if ffprobe == "" {
		return probe
	}
	ctx, cancel := context.WithTimeout(context.Background(), 12*time.Second)
	defer cancel()
	output, err := exec.CommandContext(ctx, ffprobe,
		"-v", "error",
		"-show_entries", "stream=codec_type,codec_name,width,height,display_aspect_ratio,duration,channels:stream_tags=language",
		"-of", "json",
		path,
	).Output()
	if err != nil {
		return probe
	}
	var payload struct {
		Streams []struct {
			CodecType          string            `json:"codec_type"`
			CodecName          string            `json:"codec_name"`
			Width              int               `json:"width"`
			Height             int               `json:"height"`
			DisplayAspectRatio string            `json:"display_aspect_ratio"`
			Duration           string            `json:"duration"`
			Channels           int               `json:"channels"`
			Tags               map[string]string `json:"tags"`
		} `json:"streams"`
	}
	if err := json.Unmarshal(output, &payload); err != nil {
		return probe
	}
	for _, stream := range payload.Streams {
		switch stream.CodecType {
		case "video":
			codec := videoCodecFromFFProbe(stream.CodecName)
			duration := durationSeconds(stream.Duration)
			probe.VideoStreams = append(probe.VideoStreams, VideoStream{
				Codec:           codec,
				Aspect:          aspectRatio(stream.DisplayAspectRatio, stream.Width, stream.Height),
				Width:           stream.Width,
				Height:          stream.Height,
				DurationSeconds: duration,
			})
			if probe.VideoFormat == "" {
				probe.VideoFormat = videoFormatFromHeight(stream.Height)
			}
		case "audio":
			codec := audioCodecFromFFProbe(stream.CodecName)
			probe.AudioStreams = append(probe.AudioStreams, AudioStream{
				Codec:    strings.ReplaceAll(codec, "-", "_"),
				Language: languageFromTags(stream.Tags),
				Channels: stream.Channels,
			})
			if probe.AudioCodec == "" {
				probe.AudioCodec = codec
			}
		case "subtitle":
			probe.SubtitleStreams = append(probe.SubtitleStreams, SubtitleStream{
				Language: languageFromTags(stream.Tags),
			})
		}
	}
	probe.Scanned = true
	return probe
}

func videoCodecFromFFProbe(codec string) string {
	switch strings.ToLower(strings.TrimSpace(codec)) {
	case "h265":
		return "HEVC"
	case "hevc":
		return "HEVC"
	case "h264":
		return "H.264"
	case "mpeg4":
		return "MPEG-4"
	case "vp9":
		return "VP9"
	case "av1":
		return "AV1"
	default:
		return strings.ToUpper(strings.TrimSpace(codec))
	}
}

func aspectRatio(display string, width, height int) float64 {
	display = strings.TrimSpace(display)
	if parts := strings.Split(display, ":"); len(parts) == 2 {
		left, leftErr := strconv.ParseFloat(parts[0], 64)
		right, rightErr := strconv.ParseFloat(parts[1], 64)
		if leftErr == nil && rightErr == nil && right > 0 {
			return roundAspect(left / right)
		}
	}
	if width > 0 && height > 0 {
		return roundAspect(float64(width) / float64(height))
	}
	return 0
}

func roundAspect(value float64) float64 {
	if value <= 0 {
		return 0
	}
	return float64(int(value*100+0.5)) / 100
}

func durationSeconds(value string) int {
	seconds, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
	if err != nil || seconds <= 0 {
		return 0
	}
	return int(seconds + 0.5)
}

func languageFromTags(tags map[string]string) string {
	for key, value := range tags {
		if strings.EqualFold(key, "language") {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func findFFProbe() string {
	ffprobeOnce.Do(func() {
		if path, err := exec.LookPath("ffprobe"); err == nil {
			ffprobePath = path
		}
	})
	return ffprobePath
}

func videoFormatFromHeight(height int) string {
	switch {
	case height >= 4000:
		return "4320p"
	case height >= 2000:
		return "2160p"
	case height >= 1000:
		return "1080p"
	case height >= 700:
		return "720p"
	case height >= 560:
		return "576p"
	case height >= 470:
		return "480p"
	default:
		return ""
	}
}

func audioCodecFromFFProbe(codec string) string {
	switch strings.ToLower(strings.TrimSpace(codec)) {
	case "truehd":
		return "TrueHD"
	case "dts":
		return "DTS"
	case "eac3":
		return "EAC3"
	case "ac3":
		return "AC3"
	case "aac":
		return "AAC"
	case "flac":
		return "FLAC"
	case "opus":
		return "OPUS"
	case "mp3":
		return "MP3"
	case "pcm_s16le", "pcm_s24le", "pcm_s32le", "pcm_f32le", "pcm_f64le":
		return "PCM"
	default:
		return strings.ToUpper(codec)
	}
}

func EnsureLightweightMediaInfo(item Item) Item {
	if item.FileSize != "" && item.FileSizeBytes > 0 && item.VideoFormat != "" && item.AudioCodec != "" {
		return item
	}
	info, _ := os.Stat(item.Path)
	fileSizeBytes, fileSize, videoFormat, audioCodec := LightweightMediaInfo(item.Path, info)
	if item.FileSizeBytes == 0 {
		item.FileSizeBytes = fileSizeBytes
	}
	if item.FileSize == "" {
		item.FileSize = fileSize
	}
	if item.VideoFormat == "" {
		item.VideoFormat = videoFormat
	}
	if item.AudioCodec == "" {
		item.AudioCodec = audioCodec
	}
	return item
}

func FormatFileSize(size int64) string {
	if size <= 0 {
		return ""
	}
	const unit = 1024
	if size < unit*unit {
		return fmt.Sprintf("%.0fKB", float64(size)/unit)
	}
	if size < unit*unit*unit {
		return fmt.Sprintf("%.1fMB", float64(size)/(unit*unit))
	}
	return fmt.Sprintf("%.1fGB", float64(size)/(unit*unit*unit))
}

func inferVideoFormat(text string) string {
	match := videoFormatPattern.FindString(text)
	switch strings.ToLower(match) {
	case "8k", "4320p":
		return "4320p"
	case "4k", "uhd", "2160p":
		return "2160p"
	case "1080p":
		return "1080p"
	case "720p":
		return "720p"
	case "576p":
		return "576p"
	case "540p":
		return "540p"
	case "480p":
		return "480p"
	default:
		return ""
	}
}

func inferAudioCodec(text string) string {
	for _, candidate := range audioCodecPatterns {
		if candidate.pattern.MatchString(text) {
			return candidate.value
		}
	}
	return ""
}

func FileDate(info os.FileInfo) time.Time {
	if info == nil {
		return time.Now()
	}
	if created, ok := statTime(info.Sys(), "Birthtimespec", "Birthtim", "Btim"); ok {
		return created
	}
	if modified := info.ModTime(); !modified.IsZero() {
		return modified
	}
	return time.Now()
}

func statTime(sys interface{}, fieldNames ...string) (time.Time, bool) {
	value := reflect.ValueOf(sys)
	if !value.IsValid() {
		return time.Time{}, false
	}
	if value.Kind() == reflect.Pointer {
		if value.IsNil() {
			return time.Time{}, false
		}
		value = value.Elem()
	}
	if value.Kind() != reflect.Struct {
		return time.Time{}, false
	}
	for _, fieldName := range fieldNames {
		field := value.FieldByName(fieldName)
		if !field.IsValid() {
			continue
		}
		seconds, nanoseconds, ok := timespecValues(field)
		if !ok || seconds <= 0 {
			continue
		}
		return time.Unix(seconds, nanoseconds), true
	}
	return time.Time{}, false
}

func timespecValues(value reflect.Value) (int64, int64, bool) {
	if value.Kind() == reflect.Pointer {
		if value.IsNil() {
			return 0, 0, false
		}
		value = value.Elem()
	}
	if value.Kind() != reflect.Struct {
		return 0, 0, false
	}
	seconds, ok := intField(value, "Sec", "Tv_sec")
	if !ok {
		return 0, 0, false
	}
	nanoseconds, _ := intField(value, "Nsec", "Tv_nsec")
	return seconds, nanoseconds, true
}

func intField(value reflect.Value, names ...string) (int64, bool) {
	for _, name := range names {
		field := value.FieldByName(name)
		if !field.IsValid() {
			continue
		}
		switch field.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return field.Int(), true
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return int64(field.Uint()), true
		}
	}
	return 0, false
}

func applyNFOSummary(item *Item) {
	item.NFOModTimeUnix = firstNFOModTime(*item)
	for _, path := range nfoCandidates(*item) {
		stat, statErr := os.Stat(path)
		if statErr != nil || stat.IsDir() {
			continue
		}
		summary, err := nfo.ReadSummary(path)
		if err != nil {
			continue
		}
		if summary.Type == "" && summary.Title == "" && summary.Rating == 0 && summary.TMDBID == 0 && summary.IMDBID == "" {
			continue
		}
		if summary.Title != "" {
			if item.Kind == "tvshow" && summary.Type == "tvshow" {
				item.ShowGuess = summary.Title
			}
			item.TitleGuess = summary.Title
		}
		if summary.Year != "" {
			item.YearGuess = summary.Year
		}
		item.Original = summary.Original
		item.Overview = summary.Plot
		item.Runtime = summary.Runtime
		if item.Kind == "tvshow" && summary.Type == "tvshow" {
			item.ShowRating = summary.Rating
		} else {
			item.Rating = summary.Rating
		}
		item.Genres = summary.Genres
		item.Premiered = summary.Premiered
		item.MatchedID = summary.TMDBID
		item.IMDBID = summary.IMDBID
		if len(summary.VideoStreams) > 0 {
			item.VideoStreams = videoStreamsFromNFO(summary.VideoStreams)
			if item.VideoFormat == "" {
				item.VideoFormat = videoFormatFromHeight(item.VideoStreams[0].Height)
			}
		}
		if len(summary.AudioStreams) > 0 {
			item.AudioStreams = audioStreamsFromNFO(summary.AudioStreams)
			if item.AudioCodec == "" {
				item.AudioCodec = item.AudioStreams[0].Codec
			}
		}
		if len(summary.SubtitleStreams) > 0 {
			item.SubtitleStreams = subtitleStreamsFromNFO(summary.SubtitleStreams)
		}
		if summary.Title != "" {
			item.MatchedName = summary.Title
		}
		if summary.Type != "tvshow" || item.Kind != "tvshow" {
			return
		}
	}
}

func videoStreamsFromNFO(values []nfo.VideoStream) []VideoStream {
	streams := make([]VideoStream, 0, len(values))
	for _, value := range values {
		streams = append(streams, VideoStream{
			Codec: value.Codec, Aspect: value.Aspect, Width: value.Width, Height: value.Height,
			DurationSeconds: value.DurationSeconds, StereoMode: value.StereoMode, HDRType: value.HDRType,
		})
	}
	return streams
}

func audioStreamsFromNFO(values []nfo.AudioStream) []AudioStream {
	streams := make([]AudioStream, 0, len(values))
	for _, value := range values {
		streams = append(streams, AudioStream{Codec: value.Codec, Language: value.Language, Channels: value.Channels})
	}
	return streams
}

func subtitleStreamsFromNFO(values []nfo.SubtitleStream) []SubtitleStream {
	streams := make([]SubtitleStream, 0, len(values))
	for _, value := range values {
		streams = append(streams, SubtitleStream{Language: value.Language})
	}
	return streams
}

func firstNFOModTime(item Item) int64 {
	var modTime int64
	for _, path := range nfoCandidates(item) {
		stat, err := os.Stat(path)
		if err == nil && !stat.IsDir() {
			if value := stat.ModTime().UnixNano(); value > modTime {
				modTime = value
			}
		}
	}
	return modTime
}

func nfoCandidates(item Item) []string {
	if item.Kind == "tvshow" {
		showDir := showRoot(item.SourcePath, item.Path)
		return []string{
			filepath.Join(showDir, "tvshow.nfo"),
			strings.TrimSuffix(item.Path, filepath.Ext(item.Path)) + ".nfo",
		}
	}
	return []string{
		filepath.Join(item.Dir, "movie.nfo"),
		strings.TrimSuffix(item.Path, filepath.Ext(item.Path)) + ".nfo",
	}
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
		pattern = defaultMovieRenameFilePattern
	}
	return BuildMovieRenameWithPatterns(item, title, year, movieID, defaultMovieRenamePathPattern, pattern)
}

func BuildMovieRenameWithPatterns(item Item, title string, year string, movieID int, folderPattern string, filePattern string) RenamePreview {
	return BuildMovieRenameWithOptions(item, title, year, movieID, folderPattern, filePattern, RenameOptions{}, RenameOptions{})
}

func BuildMovieRenameWithOptions(item Item, title string, year string, movieID int, folderPattern string, filePattern string, folderOptions RenameOptions, fileOptions RenameOptions) RenamePreview {
	item = EnsureLightweightMediaInfo(item)
	if strings.TrimSpace(folderPattern) == "" {
		folderPattern = defaultMovieRenamePathPattern
	}
	if strings.TrimSpace(filePattern) == "" {
		filePattern = defaultMovieRenameFilePattern
	}
	data := RenameTemplateData{
		Title:       title,
		Year:        year,
		TMDBID:      movieID,
		VideoFormat: item.VideoFormat,
		AudioCodec:  item.AudioCodec,
		FileSize:    item.FileSize,
	}
	targetFolder := renderRenamePatternWithOptions(folderPattern, data, "Movie", folderOptions)
	targetFileBase := renderRenamePatternWithOptions(filePattern, data, targetFolder, fileOptions)
	ext := filepath.Ext(item.Path)
	targetDir := filepath.Join(filepath.Dir(item.Dir), targetFolder)
	targetFile := filepath.Join(targetDir, targetFileBase+ext)
	return RenamePreview{
		SourceFile: item.Path,
		TargetFile: targetFile,
		SourceDir:  item.Dir,
		TargetDir:  targetDir,
		Operations: buildMovieRenameOperations(item, targetDir, targetFileBase, ext),
	}
}

func BuildTVShowRename(item Item, showTitle string, episodeTitle string, year string, showID int, showFolderPattern string, seasonPattern string, filePattern string) RenamePreview {
	return BuildTVShowRenameWithOptions(item, showTitle, episodeTitle, year, showID, showFolderPattern, seasonPattern, filePattern, RenameOptions{}, RenameOptions{}, RenameOptions{})
}

func BuildTVShowRenameWithOptions(item Item, showTitle string, episodeTitle string, year string, showID int, showFolderPattern string, seasonPattern string, filePattern string, showFolderOptions RenameOptions, seasonOptions RenameOptions, fileOptions RenameOptions) RenamePreview {
	item = EnsureLightweightMediaInfo(item)
	if strings.TrimSpace(showFolderPattern) == "" {
		showFolderPattern = defaultTVShowRenamePathPattern
	}
	if strings.TrimSpace(seasonPattern) == "" {
		seasonPattern = defaultTVSeasonRenamePattern
	}
	if strings.TrimSpace(filePattern) == "" {
		filePattern = defaultTVEpisodeRenamePattern
	}
	if strings.TrimSpace(showTitle) == "" {
		showTitle = item.ShowGuess
	}
	if strings.TrimSpace(episodeTitle) == "" {
		if item.Episode > 0 {
			episodeTitle = fmt.Sprintf("%02d", item.Episode)
		} else {
			episodeTitle = strings.TrimSuffix(item.FileName, filepath.Ext(item.FileName))
		}
	}
	data := RenameTemplateData{
		Title:        episodeTitle,
		ShowTitle:    showTitle,
		EpisodeTitle: episodeTitle,
		Year:         year,
		TMDBID:       showID,
		VideoFormat:  item.VideoFormat,
		AudioCodec:   item.AudioCodec,
		FileSize:     item.FileSize,
		Season:       item.Season,
		Episode:      item.Episode,
		AirDate:      item.AirDate,
	}
	showFolder := renderRenamePatternWithOptions(showFolderPattern, data, "TV Show", showFolderOptions)
	seasonFolder := renderRenamePatternWithOptions(seasonPattern, data, "Season", seasonOptions)
	fileBase := renderRenamePatternWithOptions(filePattern, data, strings.TrimSuffix(item.FileName, filepath.Ext(item.FileName)), fileOptions)
	root := strings.TrimSpace(item.SourcePath)
	if root == "" {
		root = filepath.Dir(item.Dir)
		if item.Season > 0 || item.Episode > 0 {
			root = filepath.Dir(root)
		}
	}
	targetDir := filepath.Join(root, showFolder, seasonFolder)
	ext := filepath.Ext(item.Path)
	targetFile := filepath.Join(targetDir, fileBase+ext)
	return RenamePreview{
		SourceFile: item.Path,
		TargetFile: targetFile,
		SourceDir:  item.Dir,
		TargetDir:  targetDir,
		Operations: buildEpisodeRenameOperations(item, targetDir, fileBase, ext),
	}
}

func renderRenamePattern(pattern string, data RenameTemplateData, fallback string) string {
	return renderRenamePatternWithOptions(pattern, data, fallback, RenameOptions{})
}

func renderRenamePatternWithOptions(pattern string, data RenameTemplateData, fallback string, options RenameOptions) string {
	value := strings.TrimSpace(pattern)
	season := strconv.Itoa(data.Season)
	episode := strconv.Itoa(data.Episode)
	replacements := map[string]string{
		"{title}":         safePath(data.Title),
		"{title[0]}":      safePath(firstRunes(data.Title, 1)),
		"{title;first}":   safePath(firstLetterOrReplacement(data.Title, options.FirstCharacterNumberReplacement)),
		"{title[0,2]}":    safePath(firstRunes(data.Title, 2)),
		"{showTitle}":     safePath(data.ShowTitle),
		"{episodeTitle}":  safePath(data.EpisodeTitle),
		"{year}":          data.Year,
		"{showYear}":      data.Year,
		"{tmdb-{tmdbid}}": fmt.Sprintf("{tmdb-%d}", data.TMDBID),
		"{tmdbid}":        fmt.Sprintf("%d", data.TMDBID),
		"{edition}":       "",
		"{videoFormat}":   data.VideoFormat,
		"{audioCodec}":    data.AudioCodec,
		"{fileSize}":      data.FileSize,
		"{seasonNr}":      season,
		"{seasonNr2}":     fmt.Sprintf("%02d", data.Season),
		"{episodeNr}":     episode,
		"{episodeNr2}":    fmt.Sprintf("%02d", data.Episode),
		"{aired}":         data.AirDate,
	}
	for key, replacement := range map[string]string{
		"${title}":        safePath(data.Title),
		"${title[0]}":     safePath(firstRunes(data.Title, 1)),
		"${title;first}":  safePath(firstLetterOrReplacement(data.Title, options.FirstCharacterNumberReplacement)),
		"${title[0,2]}":   safePath(firstRunes(data.Title, 2)),
		"${showTitle}":    safePath(data.ShowTitle),
		"${episodeTitle}": safePath(data.EpisodeTitle),
		"${year}":         data.Year,
		"${showYear}":     data.Year,
		"${tmdb}":         fmt.Sprintf("%d", data.TMDBID),
		"${tmdbid}":       fmt.Sprintf("%d", data.TMDBID),
		"${- ,edition,}":  "",
		"${edition}":      "",
		"${videoFormat}":  data.VideoFormat,
		"${audioCodec}":   data.AudioCodec,
		"${fileSize}":     data.FileSize,
		"${seasonNr}":     season,
		"${seasonNr2}":    fmt.Sprintf("%02d", data.Season),
		"${episodeNr}":    episode,
		"${episodeNr2}":   fmt.Sprintf("%02d", data.Episode),
		"${aired}":        data.AirDate,
	} {
		replacements[key] = replacement
	}
	keys := make([]string, 0, len(replacements))
	for key := range replacements {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool { return len(keys[i]) > len(keys[j]) })
	for _, key := range keys {
		value = strings.ReplaceAll(value, key, replacements[key])
	}
	value = cleanGeneratedName(value, options)
	if value == "" {
		return cleanGeneratedName(fallback, options)
	}
	return value
}

func firstRunes(value string, count int) string {
	value = strings.TrimSpace(value)
	if count <= 0 || value == "" {
		return ""
	}
	runes := []rune(value)
	if len(runes) < count {
		count = len(runes)
	}
	return string(runes[:count])
}

func firstLetterOrReplacement(value string, replacement string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	r := []rune(value)[0]
	if unicode.IsLetter(r) {
		return string(r)
	}
	if strings.TrimSpace(replacement) == "" {
		return "#"
	}
	return replacement
}

func cleanGeneratedName(value string, options RenameOptions) string {
	colonReplacement := options.ColonReplacement
	if !options.ColonReplacementDefined {
		colonReplacement = "-"
	}
	if colonReplacement == " " {
		colonReplacement = " "
	}
	replacer := strings.NewReplacer("/", " ", "\\", " ", ":", colonReplacement, "*", "", "?", "", "\"", "'", "<", "", ">", "", "|", " ")
	value = unresolvedRenameTokenPattern.ReplaceAllString(value, "")
	if options.ASCIIReplacement {
		value = replaceNonASCII(value)
	}
	value = strings.Join(strings.Fields(replacer.Replace(value)), " ")
	value = strings.ReplaceAll(value, "( )", "")
	value = strings.ReplaceAll(value, "[ ]", "")
	value = strings.Trim(value, " .-_")
	if options.SpaceSubstitution {
		replacement := strings.TrimSpace(options.SpaceReplacement)
		if replacement == "" {
			replacement = "_"
		}
		value = strings.ReplaceAll(value, " ", replacement)
		value = regexp.MustCompile(regexp.QuoteMeta(replacement)+"+").ReplaceAllString(value, replacement)
		value = strings.Trim(value, " .-_")
	}
	return value
}

func replaceNonASCII(value string) string {
	replacements := strings.NewReplacer(
		"Ä", "Ae", "Ö", "Oe", "Ü", "Ue", "ä", "ae", "ö", "oe", "ü", "ue", "ß", "ss",
		"Á", "A", "À", "A", "Â", "A", "Ã", "A", "Å", "A", "Æ", "Ae", "Ç", "C", "É", "E", "È", "E", "Ê", "E", "Ë", "E",
		"Í", "I", "Ì", "I", "Î", "I", "Ï", "I", "Ñ", "N", "Ó", "O", "Ò", "O", "Ô", "O", "Õ", "O", "Ø", "Oe",
		"Ú", "U", "Ù", "U", "Û", "U", "Ý", "Y", "á", "a", "à", "a", "â", "a", "ã", "a", "å", "a", "æ", "ae",
		"ç", "c", "é", "e", "è", "e", "ê", "e", "ë", "e", "í", "i", "ì", "i", "î", "i", "ï", "i", "ñ", "n",
		"ó", "o", "ò", "o", "ô", "o", "õ", "o", "ø", "oe", "ú", "u", "ù", "u", "û", "u", "ý", "y", "ÿ", "y",
	)
	return replacements.Replace(value)
}

func buildMovieRenameOperations(item Item, targetDir string, targetFileBase string, ext string) []RenameOperation {
	sourceBase := strings.TrimSuffix(item.FileName, filepath.Ext(item.FileName))
	ops := []RenameOperation{{
		Kind:   "video",
		Source: item.Path,
		Target: filepath.Join(targetDir, targetFileBase+ext),
	}}
	addRenameSidecars(&ops, item.Dir, targetDir, sourceBase, targetFileBase)
	for _, name := range []string{"movie.nfo", "poster.jpg", "poster.jpeg", "poster.png", "folder.jpg", "folder.jpeg", "folder.png", "fanart.jpg", "fanart.jpeg", "fanart.png", "backdrop.jpg", "backdrop.jpeg", "backdrop.png"} {
		source := filepath.Join(item.Dir, name)
		if exists(source) {
			ops = appendRenameOperation(ops, sidecarKind(name), source, filepath.Join(targetDir, name))
		}
	}
	return ops
}

func buildEpisodeRenameOperations(item Item, targetDir string, targetFileBase string, ext string) []RenameOperation {
	sourceBase := strings.TrimSuffix(item.FileName, filepath.Ext(item.FileName))
	ops := []RenameOperation{{
		Kind:   "video",
		Source: item.Path,
		Target: filepath.Join(targetDir, targetFileBase+ext),
	}}
	addRenameSidecars(&ops, item.Dir, targetDir, sourceBase, targetFileBase)
	addTVShowRootSidecars(&ops, item, filepath.Dir(targetDir))
	return ops
}

func addTVShowRootSidecars(ops *[]RenameOperation, item Item, targetShowDir string) {
	if item.Kind != "tvshow" || targetShowDir == "" {
		return
	}
	sourceShowDir := showRoot(item.SourcePath, item.Path)
	if sourceShowDir == "" || samePath(sourceShowDir, targetShowDir) {
		return
	}
	names := []string{
		"tvshow.nfo",
		"poster.jpg", "poster.jpeg", "poster.png",
		"folder.jpg", "folder.jpeg", "folder.png",
		"fanart.jpg", "fanart.jpeg", "fanart.png",
		"backdrop.jpg", "backdrop.jpeg", "backdrop.png",
		"banner.jpg", "banner.jpeg", "banner.png",
		"theme.mp3",
	}
	if item.Season > 0 {
		for _, suffix := range []string{"poster", "fanart", "banner"} {
			names = append(names, fmt.Sprintf("season%02d-%s.jpg", item.Season, suffix), fmt.Sprintf("season%d-%s.jpg", item.Season, suffix))
		}
	}
	for _, name := range names {
		source := filepath.Join(sourceShowDir, name)
		if exists(source) {
			*ops = appendRenameOperation(*ops, sidecarKind(name), source, filepath.Join(targetShowDir, name))
		}
	}
}

func addRenameSidecars(ops *[]RenameOperation, sourceDir string, targetDir string, sourceBase string, targetBase string) {
	entries, err := os.ReadDir(sourceDir)
	if err != nil {
		return
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		ext := strings.ToLower(filepath.Ext(name))
		base := strings.TrimSuffix(name, filepath.Ext(name))
		if base != sourceBase && !strings.HasPrefix(base, sourceBase+"-") && !strings.HasPrefix(base, sourceBase+".") {
			continue
		}
		if !subtitleExts[ext] && !imageExts[ext] && ext != ".nfo" {
			continue
		}
		suffix := strings.TrimPrefix(base, sourceBase)
		source := filepath.Join(sourceDir, name)
		target := filepath.Join(targetDir, targetBase+suffix+filepath.Ext(name))
		*ops = appendRenameOperation(*ops, sidecarKind(name), source, target)
	}
}

func appendRenameOperation(ops []RenameOperation, kind string, source string, target string) []RenameOperation {
	if source == "" || target == "" {
		return ops
	}
	for _, op := range ops {
		if op.Source == source && op.Target == target {
			return ops
		}
	}
	return append(ops, RenameOperation{Kind: kind, Source: source, Target: target})
}

func sidecarKind(name string) string {
	ext := strings.ToLower(filepath.Ext(name))
	switch {
	case ext == ".nfo":
		return "nfo"
	case subtitleExts[ext]:
		return "subtitle"
	case imageExts[ext]:
		return "artwork"
	default:
		return "sidecar"
	}
}

func ApplyRename(preview RenamePreview) error {
	operations := preview.Operations
	if len(operations) == 0 {
		operations = []RenameOperation{{
			Kind:   "video",
			Source: preview.SourceFile,
			Target: preview.TargetFile,
		}}
	}
	operations = normalizeRenameOperations(operations)
	if len(operations) == 0 {
		return nil
	}
	for _, op := range operations {
		if samePath(op.Source, op.Target) {
			continue
		}
		if !exists(op.Source) {
			continue
		}
		if exists(op.Target) {
			return fmt.Errorf("target already exists: %s", op.Target)
		}
	}
	for _, op := range operations {
		if samePath(op.Source, op.Target) || !exists(op.Source) {
			continue
		}
		if err := moveFileSafe(op.Source, op.Target); err != nil {
			return err
		}
	}
	removeEmptyParents(preview.SourceDir, filepath.Dir(preview.SourceDir))
	return nil
}

func normalizeRenameOperations(operations []RenameOperation) []RenameOperation {
	seen := map[string]bool{}
	normalized := make([]RenameOperation, 0, len(operations))
	for _, op := range operations {
		source := filepath.Clean(strings.TrimSpace(op.Source))
		target := filepath.Clean(strings.TrimSpace(op.Target))
		if source == "." || target == "." || source == "" || target == "" {
			continue
		}
		key := source + "\x00" + target
		if seen[key] {
			continue
		}
		seen[key] = true
		if op.Kind == "" {
			op.Kind = "file"
		}
		op.Source = source
		op.Target = target
		normalized = append(normalized, op)
	}
	return normalized
}

func moveFileSafe(source string, target string) error {
	if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
		return err
	}
	var lastErr error
	for i := 0; i < 5; i++ {
		if err := os.Rename(source, target); err == nil {
			return nil
		} else {
			lastErr = err
			if isCrossDeviceRename(err) {
				break
			}
			time.Sleep(time.Duration(i+1) * 100 * time.Millisecond)
		}
	}
	if !isCrossDeviceRename(lastErr) {
		return lastErr
	}
	if err := copyFile(source, target); err != nil {
		return err
	}
	return os.Remove(source)
}

func copyFile(source string, target string) error {
	input, err := os.Open(source)
	if err != nil {
		return err
	}
	defer input.Close()
	info, err := input.Stat()
	if err != nil {
		return err
	}
	output, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_EXCL, info.Mode())
	if err != nil {
		return err
	}
	_, copyErr := io.Copy(output, input)
	closeErr := output.Close()
	if copyErr != nil {
		_ = os.Remove(target)
		return copyErr
	}
	if closeErr != nil {
		_ = os.Remove(target)
		return closeErr
	}
	return nil
}

func isCrossDeviceRename(err error) bool {
	var linkErr *os.LinkError
	if errors.As(err, &linkErr) {
		return errors.Is(linkErr.Err, syscall.EXDEV)
	}
	return false
}

func samePath(a string, b string) bool {
	aa, errA := filepath.Abs(a)
	bb, errB := filepath.Abs(b)
	if errA == nil && errB == nil {
		return aa == bb
	}
	return filepath.Clean(a) == filepath.Clean(b)
}

func removeEmptyParents(dir string, stop string) {
	dir = filepath.Clean(dir)
	stop = filepath.Clean(stop)
	for dir != "." && dir != string(filepath.Separator) && dir != stop {
		if err := os.Remove(dir); err != nil {
			return
		}
		dir = filepath.Dir(dir)
	}
}

func RefreshItemPath(item Item, targetFile string) Item {
	info, _ := os.Stat(targetFile)
	fileSizeBytes, fileSize, videoFormat, audioCodec := LightweightMediaInfo(targetFile, info)
	item.Path = targetFile
	item.Dir = filepath.Dir(targetFile)
	item.FileName = filepath.Base(targetFile)
	item.ID = stableID(targetFile)
	item.MediaType = ClassifyMediaFile(targetFile)
	item.FileSizeBytes = fileSizeBytes
	item.FileSize = fileSize
	item.VideoFormat = videoFormat
	item.AudioCodec = audioCodec
	item.HasNFO = exists(strings.TrimSuffix(targetFile, filepath.Ext(targetFile))+".nfo") || exists(filepath.Join(item.Dir, "movie.nfo"))
	if item.Kind == "tvshow" {
		showDir := showRoot(item.SourcePath, item.Path)
		item.HasPoster = hasAny(showDir, tvShowPosterCandidates(item.Season, item.FileName)) || hasAny(item.Dir, tvEpisodeThumbCandidates(item.FileName))
		item.HasFanart = hasAny(showDir, tvShowFanartCandidates(item.Season, item.FileName))
	} else {
		item.HasPoster = hasAny(item.Dir, []string{"poster.jpg", "folder.jpg", strings.TrimSuffix(item.FileName, filepath.Ext(item.FileName)) + "-poster.jpg"})
		item.HasFanart = hasAny(item.Dir, []string{"fanart.jpg", "backdrop.jpg", strings.TrimSuffix(item.FileName, filepath.Ext(item.FileName)) + "-fanart.jpg"})
	}
	item.HasSubtitle = hasSubtitle(targetFile)
	return item
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

func tvShowPosterCandidates(season int, fileName string) []string {
	names := []string{"poster.jpg", "folder.jpg"}
	if season > 0 {
		names = append([]string{fmt.Sprintf("season%02d-poster.jpg", season), fmt.Sprintf("season%d-poster.jpg", season)}, names...)
	}
	names = append(names, tvEpisodeThumbCandidates(fileName)...)
	return names
}

func tvShowFanartCandidates(season int, fileName string) []string {
	names := []string{"fanart.jpg", "backdrop.jpg"}
	if season > 0 {
		names = append([]string{fmt.Sprintf("season%02d-fanart.jpg", season), fmt.Sprintf("season%d-fanart.jpg", season)}, names...)
	}
	fileBase := strings.TrimSuffix(fileName, filepath.Ext(fileName))
	names = append(names, fileBase+"-fanart.jpg")
	return names
}

func tvEpisodeThumbCandidates(fileName string) []string {
	fileBase := strings.TrimSuffix(fileName, filepath.Ext(fileName))
	return []string{fileBase + "-thumb.jpg", fileBase + "-poster.jpg"}
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
