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

type unique struct {
	Type    string `xml:"type,attr"`
	Default string `xml:"default,attr,omitempty"`
	Value   string `xml:",chardata"`
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
