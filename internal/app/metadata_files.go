package app

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"tmmweb/internal/media"
	"tmmweb/internal/nfo"
	"tmmweb/internal/tmdb"
)

// Metadata file helpers keep the scraper handlers focused on workflow rather
// than Kodi/NFO naming conventions and local artwork placement.
func writeMovieNFO(dir string, movie tmdb.Movie, item media.Item, overwrite bool) error {
	path := filepath.Join(dir, "movie.nfo")
	if !overwrite {
		if _, err := os.Stat(path); err == nil {
			return nil
		}
	}
	return nfo.WriteMovie(dir, movie, nfoMediaFileInfo(item))
}

func writeTVShowNFO(dir string, show tmdb.TVShow, overwrite bool) error {
	path := filepath.Join(dir, "tvshow.nfo")
	if !overwrite {
		if _, err := os.Stat(path); err == nil {
			return nil
		}
	}
	return nfo.WriteTVShow(dir, show)
}

func writeTVSeasonNFO(path string, season tmdb.TVSeason, fanartPath string, overwrite bool) error {
	if !overwrite {
		if _, err := os.Stat(path); err == nil {
			return nil
		}
	}
	return nfo.WriteTVSeason(path, season, fanartPath)
}

func writeTVEpisodeNFO(path string, show tmdb.TVShow, episode tmdb.TVEpisode, item media.Item, overwrite bool) error {
	if !overwrite {
		if _, err := os.Stat(path); err == nil {
			return nil
		}
	}
	return nfo.WriteTVEpisode(path, show, episode, nfoMediaFileInfo(item))
}

func nfoMediaFileInfo(item media.Item) nfo.EpisodeFileInfo {
	info := nfo.EpisodeFileInfo{
		FileName:  item.FileName,
		DateAdded: nfoDateTime(item.DateAdded),
	}
	for _, stream := range item.VideoStreams {
		info.VideoStreams = append(info.VideoStreams, nfo.VideoStream{
			Codec: stream.Codec, Aspect: stream.Aspect, Width: stream.Width, Height: stream.Height,
			DurationSeconds: stream.DurationSeconds, StereoMode: stream.StereoMode, HDRType: stream.HDRType,
		})
	}
	for _, stream := range item.AudioStreams {
		info.AudioStreams = append(info.AudioStreams, nfo.AudioStream{
			Codec: stream.Codec, Language: stream.Language, Channels: stream.Channels,
		})
	}
	for _, stream := range item.SubtitleStreams {
		info.SubtitleStreams = append(info.SubtitleStreams, nfo.SubtitleStream{Language: stream.Language})
	}
	return info
}

func nfoDateTime(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	if parsed, err := time.Parse(time.RFC3339, value); err == nil {
		return parsed.Local().Format("2006-01-02 15:04:05")
	}
	return value
}

func writeImage(dir string, name string, data []byte, overwrite bool) error {
	if !overwrite {
		if _, err := os.Stat(filepath.Join(dir, name)); err == nil {
			return nil
		}
	}
	return nfo.WriteImage(dir, name, data)
}

func writeImages(dir string, names []string, data []byte, overwrite bool) error {
	var firstErr error
	for _, name := range names {
		if err := writeImage(dir, name, data, overwrite); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

func downloadTVTheme(dir string, tvdbID int, overwrite bool) error {
	if tvdbID <= 0 {
		return nil
	}
	path := filepath.Join(dir, "theme.mp3")
	if !overwrite {
		if _, err := os.Stat(path); err == nil {
			return nil
		}
	}
	client := http.Client{Timeout: 20 * time.Second}
	response, err := client.Get(fmt.Sprintf("http://tvthemes.plexapp.com/%d.mp3", tvdbID))
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode == http.StatusNotFound || response.StatusCode == http.StatusForbidden {
		return nil
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return fmt.Errorf("theme download status %d", response.StatusCode)
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	temp, err := os.CreateTemp(dir, "theme.*.part")
	if err != nil {
		return err
	}
	tempPath := temp.Name()
	_, copyErr := io.Copy(temp, response.Body)
	closeErr := temp.Close()
	if copyErr != nil {
		_ = os.Remove(tempPath)
		return copyErr
	}
	if closeErr != nil {
		_ = os.Remove(tempPath)
		return closeErr
	}
	if overwrite {
		_ = os.Remove(path)
	}
	return os.Rename(tempPath, path)
}

func artworkPath(item media.Item, artType string, scope string) string {
	var dirs []string
	var names []string
	fileBase := strings.TrimSuffix(item.FileName, filepath.Ext(item.FileName))
	if item.Kind == "tvshow" {
		showDir := tvShowRootDir(item)
		seasonDir := tvSeasonDir(item, showDir)
		tvDirs := tvArtworkDirs(item, showDir, seasonDir)
		switch scope {
		case "show":
			dirs = append(dirs, showDir)
			if artType == "poster" {
				names = append(names, defaultTVShowPosterNames()...)
			} else {
				names = append(names, defaultTVShowFanartNames()...)
			}
		case "season":
			if artType == "poster" && item.Season > 0 {
				dirs = append(dirs, tvDirs...)
				names = append(names, seasonPosterNames(item.Season)...)
				names = append(names, defaultTVShowPosterNames()...)
			} else if artType == "fanart" && item.Season > 0 {
				dirs = append(dirs, tvDirs...)
				names = append(names, seasonFanartNames(item.Season)...)
				names = append(names, defaultTVShowFanartNames()...)
			}
		default:
			if artType == "poster" && item.Season > 0 {
				dirs = append(dirs, tvDirs...)
				names = append(names, seasonPosterNames(item.Season)...)
				names = append(names, defaultTVShowPosterNames()...)
			} else if artType == "fanart" {
				dirs = append(dirs, tvDirs...)
				names = append(names, episodeThumbNames(item)...)
			}
		}
		if len(dirs) == 0 {
			dirs = append(dirs, showDir)
		}
		if len(names) == 0 {
			if artType == "poster" {
				names = append(names, defaultTVShowPosterNames()...)
			} else {
				names = append(names, defaultTVShowFanartNames()...)
			}
		}
	} else {
		dirs = append(dirs, item.Dir)
		if artType == "poster" {
			names = append(names, "poster.jpg", "folder.jpg", fileBase+"-poster.jpg")
		} else {
			names = append(names, "fanart.jpg", "backdrop.jpg", fileBase+"-fanart.jpg")
		}
	}
	seen := map[string]bool{}
	for _, dir := range dirs {
		if dir == "" {
			continue
		}
		for _, name := range names {
			name = strings.ReplaceAll(name, "{filename}", fileBase)
			path := filepath.Join(dir, filepath.Base(name))
			if seen[path] {
				continue
			}
			seen[path] = true
			if stat, err := os.Stat(path); err == nil && !stat.IsDir() {
				return path
			}
		}
	}
	return ""
}

func imageNames(configured string, legacy string, defaults []string, item media.Item) []string {
	values := splitConfiguredNames(configured)
	if len(values) == 0 && strings.TrimSpace(legacy) != "" {
		values = []string{legacy}
	}
	if len(values) == 0 {
		values = defaults
	}
	fileBase := strings.TrimSuffix(item.FileName, filepath.Ext(item.FileName))
	seen := map[string]bool{}
	var names []string
	for _, value := range values {
		name := strings.TrimSpace(value)
		if name == "" {
			continue
		}
		name = strings.ReplaceAll(name, "{filename}", fileBase)
		name = strings.ReplaceAll(name, "{moviefilename}", fileBase)
		name = strings.ReplaceAll(name, "{movieFilename}", fileBase)
		if filepath.Ext(name) == "" {
			name += ".jpg"
		}
		name = filepath.Base(name)
		if name == "." || seen[name] {
			continue
		}
		seen[name] = true
		names = append(names, name)
	}
	if len(names) == 0 {
		return defaults
	}
	return names
}

func splitConfiguredNames(value string) []string {
	return strings.FieldsFunc(value, func(r rune) bool {
		return r == '\n' || r == '\r' || r == ',' || r == ';'
	})
}

func defaultMoviePosterNames() []string {
	return []string{"poster.jpg", "folder.jpg", "{filename}-poster.jpg"}
}

func defaultMovieFanartNames() []string {
	return []string{"fanart.jpg", "{filename}-fanart.jpg"}
}

func defaultTVShowPosterNames() []string {
	return []string{"poster.jpg", "folder.jpg"}
}

func defaultTVShowFanartNames() []string {
	return []string{"fanart.jpg", "backdrop.jpg"}
}

func seasonNFOPath(item media.Item, season int) string {
	showDir := tvShowRootDir(item)
	seasonDir := tvSeasonDir(item, showDir)
	if seasonDir != "" && seasonDir != showDir {
		return filepath.Join(seasonDir, "season.nfo")
	}
	return filepath.Join(showDir, seasonFilename(season, "nfo"))
}

func tvShowRootDir(item media.Item) string {
	source := strings.TrimSpace(item.SourcePath)
	if source != "" {
		if rel, err := filepath.Rel(source, item.Path); err == nil && rel != "." && !strings.HasPrefix(rel, "..") {
			parts := strings.Split(rel, string(filepath.Separator))
			if len(parts) > 1 && parts[0] != "" {
				return filepath.Join(source, parts[0])
			}
		}
	}
	if item.Season > 0 || item.Episode > 0 {
		parent := filepath.Dir(item.Dir)
		if parent != "." && parent != string(filepath.Separator) {
			return parent
		}
	}
	return item.Dir
}

func tvSeasonDir(item media.Item, showDir string) string {
	if item.Season > 0 || item.Episode > 0 {
		return item.Dir
	}
	return showDir
}

func tvArtworkDirs(item media.Item, showDir string, seasonDir string) []string {
	candidates := []string{
		item.Dir,
		seasonDir,
		showDir,
	}
	if item.Dir != "" {
		parent := filepath.Dir(item.Dir)
		if parent != "." && parent != string(filepath.Separator) {
			candidates = append(candidates, parent)
			grandparent := filepath.Dir(parent)
			if grandparent != "." && grandparent != string(filepath.Separator) {
				candidates = append(candidates, grandparent)
			}
		}
	}
	seen := map[string]bool{}
	dirs := make([]string, 0, len(candidates))
	for _, dir := range candidates {
		dir = strings.TrimSpace(dir)
		if dir == "" || seen[dir] {
			continue
		}
		seen[dir] = true
		dirs = append(dirs, dir)
	}
	return dirs
}

func seasonPosterNames(season int) []string {
	return []string{seasonFilename(season, "jpg", "-poster"), fmt.Sprintf("season%d-poster.jpg", season)}
}

func seasonFanartNames(season int) []string {
	return []string{seasonFilename(season, "jpg", "-fanart"), fmt.Sprintf("season%d-fanart.jpg", season)}
}

func episodeThumbNames(item media.Item) []string {
	fileBase := strings.TrimSuffix(item.FileName, filepath.Ext(item.FileName))
	return []string{
		fileBase + "-thumb.jpg",
		fileBase + "-poster.jpg",
	}
}

func episodeNFOPath(item media.Item) string {
	return strings.TrimSuffix(item.Path, filepath.Ext(item.Path)) + ".nfo"
}

func seasonFilename(season int, extension string, suffix ...string) string {
	name := ""
	if season == 0 {
		name = "season-specials"
	} else {
		name = fmt.Sprintf("season%02d", season)
	}
	if len(suffix) > 0 {
		name += suffix[0]
	}
	return name + "." + extension
}
