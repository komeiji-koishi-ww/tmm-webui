package nfo

import (
	"encoding/xml"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"tmmweb/internal/tmdb"
)

type movieNFO struct {
	XMLName   xml.Name `xml:"movie"`
	Title     string   `xml:"title"`
	Original  string   `xml:"originaltitle,omitempty"`
	SortTitle string   `xml:"sorttitle,omitempty"`
	Year      string   `xml:"year,omitempty"`
	Plot      string   `xml:"plot,omitempty"`
	Runtime   int      `xml:"runtime,omitempty"`
	Rating    string   `xml:"rating,omitempty"`
	UniqueIDs []unique `xml:"uniqueid"`
	Genres    []string `xml:"genre,omitempty"`
	Premiered string   `xml:"premiered,omitempty"`
	Thumb     string   `xml:"thumb,omitempty"`
	Fanart    string   `xml:"fanart>thumb,omitempty"`
}

type tvShowNFO struct {
	XMLName   xml.Name `xml:"tvshow"`
	Title     string   `xml:"title"`
	Original  string   `xml:"originaltitle,omitempty"`
	SortTitle string   `xml:"sorttitle,omitempty"`
	Year      string   `xml:"year,omitempty"`
	Plot      string   `xml:"plot,omitempty"`
	Rating    string   `xml:"rating,omitempty"`
	UniqueIDs []unique `xml:"uniqueid"`
	Genres    []string `xml:"genre,omitempty"`
	Premiered string   `xml:"premiered,omitempty"`
	Thumb     string   `xml:"thumb,omitempty"`
	Fanart    string   `xml:"fanart>thumb,omitempty"`
}

type tvSeasonNFO struct {
	XMLName      xml.Name `xml:"season"`
	SeasonNumber int      `xml:"seasonnumber"`
	Title        string   `xml:"title"`
	ShowTitle    string   `xml:"showtitle,omitempty"`
	SortTitle    string   `xml:"sorttitle,omitempty"`
	Year         string   `xml:"year,omitempty"`
	Plot         string   `xml:"plot,omitempty"`
	Rating       string   `xml:"rating,omitempty"`
	TMDBID       int      `xml:"tmdbid,omitempty"`
	UniqueIDs    []unique `xml:"uniqueid"`
	Premiered    string   `xml:"premiered,omitempty"`
	Thumb        string   `xml:"thumb,omitempty"`
	Fanart       string   `xml:"fanart>thumb,omitempty"`
}

type unique struct {
	Type    string `xml:"type,attr"`
	Default string `xml:"default,attr,omitempty"`
	Value   string `xml:",chardata"`
}

type Summary struct {
	Title        string
	Original     string
	Year         string
	Plot         string
	Runtime      int
	Rating       float64
	Genres       []string
	Premiered    string
	TMDBID       int
	IMDBID       string
	SeasonNumber int
}

func WriteMovie(dir string, movie tmdb.Movie) error {
	value := movieNFO{
		Title: movie.Title, Original: movie.Original, SortTitle: movie.Title,
		Year: year(movie.ReleaseDate), Plot: movie.Overview, Runtime: movie.Runtime,
		Rating: strconv.FormatFloat(movie.VoteAverage, 'f', 1, 64),
		Genres: movie.Genres, Premiered: movie.ReleaseDate,
	}
	value.UniqueIDs = append(value.UniqueIDs, unique{Type: "tmdb", Default: "true", Value: strconv.Itoa(movie.ID)})
	if movie.ImdbID != "" {
		value.UniqueIDs = append(value.UniqueIDs, unique{Type: "imdb", Value: movie.ImdbID})
	}
	output, err := xml.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	output = append([]byte(xml.Header), output...)
	output = append(output, '\n')
	return os.WriteFile(filepath.Join(dir, "movie.nfo"), output, 0644)
}

func WriteTVShow(dir string, show tmdb.TVShow) error {
	value := tvShowNFO{
		Title: show.Title, Original: show.Original, SortTitle: show.Title,
		Year: year(show.FirstAirDate), Plot: show.Overview,
		Rating: strconv.FormatFloat(show.VoteAverage, 'f', 1, 64),
		Genres: show.Genres, Premiered: show.FirstAirDate,
	}
	value.UniqueIDs = append(value.UniqueIDs, unique{Type: "tmdb", Default: "true", Value: strconv.Itoa(show.ID)})
	output, err := xml.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	output = append([]byte(xml.Header), output...)
	output = append(output, '\n')
	return os.WriteFile(filepath.Join(dir, "tvshow.nfo"), output, 0644)
}

func WriteTVSeason(path string, season tmdb.TVSeason, fanartPath string) error {
	value := tvSeasonNFO{
		SeasonNumber: season.SeasonNumber,
		Title:        seasonTitle(season),
		ShowTitle:    season.ShowTitle,
		SortTitle:    seasonTitle(season),
		Year:         year(season.AirDate),
		Plot:         season.Overview,
		Rating:       strconv.FormatFloat(season.VoteAverage, 'f', 1, 64),
		TMDBID:       season.ShowID,
		Premiered:    season.AirDate,
		Thumb:        season.PosterPath,
		Fanart:       fanartPath,
	}
	value.UniqueIDs = append(value.UniqueIDs, unique{Type: "tmdb", Default: "true", Value: strconv.Itoa(season.ShowID)})
	output, err := xml.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	output = append([]byte(xml.Header), output...)
	output = append(output, '\n')
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	return os.WriteFile(path, output, 0644)
}

func ReadSummary(path string) (Summary, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Summary{}, err
	}
	var probe struct {
		XMLName xml.Name
	}
	if err := xml.Unmarshal(data, &probe); err != nil {
		return Summary{}, err
	}
	switch probe.XMLName.Local {
	case "movie":
		var value movieNFO
		if err := xml.Unmarshal(data, &value); err != nil {
			return Summary{}, err
		}
		return summaryFromMovie(value), nil
	case "tvshow":
		var value tvShowNFO
		if err := xml.Unmarshal(data, &value); err != nil {
			return Summary{}, err
		}
		return summaryFromTVShow(value), nil
	case "season":
		var value tvSeasonNFO
		if err := xml.Unmarshal(data, &value); err != nil {
			return Summary{}, err
		}
		return summaryFromSeason(value), nil
	default:
		return Summary{}, nil
	}
}

func summaryFromMovie(value movieNFO) Summary {
	summary := Summary{
		Title:     strings.TrimSpace(value.Title),
		Original:  strings.TrimSpace(value.Original),
		Year:      firstNonEmpty(strings.TrimSpace(value.Year), year(value.Premiered)),
		Plot:      strings.TrimSpace(value.Plot),
		Runtime:   value.Runtime,
		Rating:    parseRating(value.Rating),
		Genres:    compactStrings(value.Genres),
		Premiered: strings.TrimSpace(value.Premiered),
	}
	applyUniqueIDs(&summary, value.UniqueIDs)
	return summary
}

func summaryFromTVShow(value tvShowNFO) Summary {
	summary := Summary{
		Title:     strings.TrimSpace(value.Title),
		Original:  strings.TrimSpace(value.Original),
		Year:      firstNonEmpty(strings.TrimSpace(value.Year), year(value.Premiered)),
		Plot:      strings.TrimSpace(value.Plot),
		Rating:    parseRating(value.Rating),
		Genres:    compactStrings(value.Genres),
		Premiered: strings.TrimSpace(value.Premiered),
	}
	applyUniqueIDs(&summary, value.UniqueIDs)
	return summary
}

func summaryFromSeason(value tvSeasonNFO) Summary {
	summary := Summary{
		Title:        strings.TrimSpace(value.Title),
		Original:     strings.TrimSpace(value.SortTitle),
		Year:         firstNonEmpty(strings.TrimSpace(value.Year), year(value.Premiered)),
		Plot:         strings.TrimSpace(value.Plot),
		Rating:       parseRating(value.Rating),
		Premiered:    strings.TrimSpace(value.Premiered),
		TMDBID:       value.TMDBID,
		SeasonNumber: value.SeasonNumber,
	}
	applyUniqueIDs(&summary, value.UniqueIDs)
	return summary
}

func applyUniqueIDs(summary *Summary, values []unique) {
	for _, value := range values {
		idType := strings.ToLower(strings.TrimSpace(value.Type))
		idValue := strings.TrimSpace(value.Value)
		switch idType {
		case "tmdb":
			if id, err := strconv.Atoi(idValue); err == nil {
				summary.TMDBID = id
			}
		case "imdb":
			summary.IMDBID = idValue
		}
	}
}

func parseRating(value string) float64 {
	rating, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
	if err != nil {
		return 0
	}
	return rating
}

func compactStrings(values []string) []string {
	result := make([]string, 0, len(values))
	seen := map[string]bool{}
	for _, value := range values {
		value = strings.TrimSpace(value)
		key := strings.ToLower(value)
		if value == "" || seen[key] {
			continue
		}
		seen[key] = true
		result = append(result, value)
	}
	return result
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func seasonTitle(season tmdb.TVSeason) string {
	if strings.TrimSpace(season.Title) != "" {
		return season.Title
	}
	return "Season " + strconv.Itoa(season.SeasonNumber)
}

func year(date string) string {
	if len(date) >= 4 {
		return date[:4]
	}
	return ""
}

func WriteImage(dir string, name string, data []byte) error {
	name = strings.TrimSpace(name)
	if name == "" || len(data) == 0 {
		return nil
	}
	return os.WriteFile(filepath.Join(dir, name), data, 0644)
}
