package app

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"tmmweb/internal/media"
	"tmmweb/internal/tmdb"
)

// AppSettings mirrors the persisted settings.json contract used by the UI.
type AppSettings struct {
	TMDBAPIKey                                 string   `json:"tmdbApiKey"`
	ProxyEnabled                               bool     `json:"proxyEnabled"`
	ProxyHost                                  string   `json:"proxyHost"`
	ProxyPort                                  int      `json:"proxyPort"`
	ProxyUsername                              string   `json:"proxyUsername"`
	ProxyPassword                              string   `json:"proxyPassword"`
	MovieScrapeMetadata                        *bool    `json:"movieScrapeMetadata,omitempty"`
	MovieScrapeNFO                             *bool    `json:"movieScrapeNfo,omitempty"`
	MovieScrapeImages                          *bool    `json:"movieScrapeImages,omitempty"`
	MovieScrapeOverwrite                       *bool    `json:"movieScrapeOverwrite,omitempty"`
	TVShowScrapeMetadata                       *bool    `json:"tvShowScrapeMetadata,omitempty"`
	TVShowEpisodeMetadata                      *bool    `json:"tvShowEpisodeMetadata,omitempty"`
	TVShowScrapeNFO                            *bool    `json:"tvShowScrapeNfo,omitempty"`
	TVShowScrapeImages                         *bool    `json:"tvShowScrapeImages,omitempty"`
	TVShowScrapeOverwrite                      *bool    `json:"tvShowScrapeOverwrite,omitempty"`
	MovieRenameAfterScrape                     *bool    `json:"movieRenameAfterScrape,omitempty"`
	TVShowRenameAfterScrape                    *bool    `json:"tvShowRenameAfterScrape,omitempty"`
	MovieScraperFields                         []string `json:"movieScraperFields,omitempty"`
	TVShowScraperFields                        []string `json:"tvShowScraperFields,omitempty"`
	TVEpisodeScraperFields                     []string `json:"tvEpisodeScraperFields,omitempty"`
	MovieRenamerPathname                       string   `json:"movieRenamerPathname"`
	MovieRenamerFilename                       string   `json:"movieRenamerFilename"`
	MovieRenamerPathSpaceSubstitution          *bool    `json:"movieRenamerPathSpaceSubstitution,omitempty"`
	MovieRenamerPathSpaceReplacement           string   `json:"movieRenamerPathSpaceReplacement"`
	MovieRenamerFilenameSpaceSubstitution      *bool    `json:"movieRenamerFilenameSpaceSubstitution,omitempty"`
	MovieRenamerFilenameSpaceReplacement       string   `json:"movieRenamerFilenameSpaceReplacement"`
	MovieRenamerColonReplacement               string   `json:"movieRenamerColonReplacement"`
	MovieRenamerAsciiReplacement               *bool    `json:"movieRenamerAsciiReplacement,omitempty"`
	MovieRenamerFirstCharacterReplacement      string   `json:"movieRenamerFirstCharacterReplacement"`
	MovieRenamerCreateSingleMovieSet           *bool    `json:"movieRenamerCreateSingleMovieSet,omitempty"`
	MovieRenamerNFOCleanup                     *bool    `json:"movieRenamerNfoCleanup,omitempty"`
	MovieRenamerCleanupUnwanted                *bool    `json:"movieRenamerCleanupUnwanted,omitempty"`
	MovieRenamerAllowMerge                     *bool    `json:"movieRenamerAllowMerge,omitempty"`
	TVShowRenamerShowFolder                    string   `json:"tvShowRenamerShowFolder"`
	TVShowRenamerSeason                        string   `json:"tvShowRenamerSeason"`
	TVShowRenamerFilename                      string   `json:"tvShowRenamerFilename"`
	TVShowRenamerShowFolderSpaceSubstitution   *bool    `json:"tvShowRenamerShowFolderSpaceSubstitution,omitempty"`
	TVShowRenamerShowFolderSpaceReplacement    string   `json:"tvShowRenamerShowFolderSpaceReplacement"`
	TVShowRenamerSeasonFolderSpaceSubstitution *bool    `json:"tvShowRenamerSeasonFolderSpaceSubstitution,omitempty"`
	TVShowRenamerSeasonFolderSpaceReplacement  string   `json:"tvShowRenamerSeasonFolderSpaceReplacement"`
	TVShowRenamerFilenameSpaceSubstitution     *bool    `json:"tvShowRenamerFilenameSpaceSubstitution,omitempty"`
	TVShowRenamerFilenameSpaceReplacement      string   `json:"tvShowRenamerFilenameSpaceReplacement"`
	TVShowRenamerColonReplacement              string   `json:"tvShowRenamerColonReplacement"`
	TVShowRenamerAsciiReplacement              *bool    `json:"tvShowRenamerAsciiReplacement,omitempty"`
	TVShowRenamerFirstCharacterReplacement     string   `json:"tvShowRenamerFirstCharacterReplacement"`
	TVShowRenamerCleanupUnwanted               *bool    `json:"tvShowRenamerCleanupUnwanted,omitempty"`
	MoviePosterName                            string   `json:"moviePosterName"`
	MovieFanartName                            string   `json:"movieFanartName"`
	MoviePosterNames                           string   `json:"moviePosterNames"`
	MovieFanartNames                           string   `json:"movieFanartNames"`
	TVShowPosterName                           string   `json:"tvShowPosterName"`
	TVShowFanartName                           string   `json:"tvShowFanartName"`
	TVShowPosterNames                          string   `json:"tvShowPosterNames"`
	TVShowFanartNames                          string   `json:"tvShowFanartNames"`
}

type SettingsResponse struct {
	TMDBConfigured                             bool     `json:"tmdbConfigured"`
	TMDBEnabled                                bool     `json:"tmdbEnabled"`
	TMDBKeySource                              string   `json:"tmdbKeySource"`
	ProxyEnabled                               bool     `json:"proxyEnabled"`
	ProxyHost                                  string   `json:"proxyHost"`
	ProxyPort                                  int      `json:"proxyPort"`
	ProxyUsername                              string   `json:"proxyUsername"`
	ProxyPassword                              bool     `json:"proxyPassword"`
	MovieScrapeMetadata                        bool     `json:"movieScrapeMetadata"`
	MovieScrapeNFO                             bool     `json:"movieScrapeNfo"`
	MovieScrapeImages                          bool     `json:"movieScrapeImages"`
	MovieScrapeOverwrite                       bool     `json:"movieScrapeOverwrite"`
	TVShowScrapeMetadata                       bool     `json:"tvShowScrapeMetadata"`
	TVShowEpisodeMetadata                      bool     `json:"tvShowEpisodeMetadata"`
	TVShowScrapeNFO                            bool     `json:"tvShowScrapeNfo"`
	TVShowScrapeImages                         bool     `json:"tvShowScrapeImages"`
	TVShowScrapeOverwrite                      bool     `json:"tvShowScrapeOverwrite"`
	MovieRenameAfterScrape                     bool     `json:"movieRenameAfterScrape"`
	TVShowRenameAfterScrape                    bool     `json:"tvShowRenameAfterScrape"`
	MovieScraperFields                         []string `json:"movieScraperFields"`
	TVShowScraperFields                        []string `json:"tvShowScraperFields"`
	TVEpisodeScraperFields                     []string `json:"tvEpisodeScraperFields"`
	MovieRenamerPathname                       string   `json:"movieRenamerPathname"`
	MovieRenamerFilename                       string   `json:"movieRenamerFilename"`
	MovieRenamerPathSpaceSubstitution          bool     `json:"movieRenamerPathSpaceSubstitution"`
	MovieRenamerPathSpaceReplacement           string   `json:"movieRenamerPathSpaceReplacement"`
	MovieRenamerFilenameSpaceSubstitution      bool     `json:"movieRenamerFilenameSpaceSubstitution"`
	MovieRenamerFilenameSpaceReplacement       string   `json:"movieRenamerFilenameSpaceReplacement"`
	MovieRenamerColonReplacement               string   `json:"movieRenamerColonReplacement"`
	MovieRenamerAsciiReplacement               bool     `json:"movieRenamerAsciiReplacement"`
	MovieRenamerFirstCharacterReplacement      string   `json:"movieRenamerFirstCharacterReplacement"`
	MovieRenamerCreateSingleMovieSet           bool     `json:"movieRenamerCreateSingleMovieSet"`
	MovieRenamerNFOCleanup                     bool     `json:"movieRenamerNfoCleanup"`
	MovieRenamerCleanupUnwanted                bool     `json:"movieRenamerCleanupUnwanted"`
	MovieRenamerAllowMerge                     bool     `json:"movieRenamerAllowMerge"`
	TVShowRenamerShowFolder                    string   `json:"tvShowRenamerShowFolder"`
	TVShowRenamerSeason                        string   `json:"tvShowRenamerSeason"`
	TVShowRenamerFilename                      string   `json:"tvShowRenamerFilename"`
	TVShowRenamerShowFolderSpaceSubstitution   bool     `json:"tvShowRenamerShowFolderSpaceSubstitution"`
	TVShowRenamerShowFolderSpaceReplacement    string   `json:"tvShowRenamerShowFolderSpaceReplacement"`
	TVShowRenamerSeasonFolderSpaceSubstitution bool     `json:"tvShowRenamerSeasonFolderSpaceSubstitution"`
	TVShowRenamerSeasonFolderSpaceReplacement  string   `json:"tvShowRenamerSeasonFolderSpaceReplacement"`
	TVShowRenamerFilenameSpaceSubstitution     bool     `json:"tvShowRenamerFilenameSpaceSubstitution"`
	TVShowRenamerFilenameSpaceReplacement      string   `json:"tvShowRenamerFilenameSpaceReplacement"`
	TVShowRenamerColonReplacement              string   `json:"tvShowRenamerColonReplacement"`
	TVShowRenamerAsciiReplacement              bool     `json:"tvShowRenamerAsciiReplacement"`
	TVShowRenamerFirstCharacterReplacement     string   `json:"tvShowRenamerFirstCharacterReplacement"`
	TVShowRenamerCleanupUnwanted               bool     `json:"tvShowRenamerCleanupUnwanted"`
	MoviePosterName                            string   `json:"moviePosterName"`
	MovieFanartName                            string   `json:"movieFanartName"`
	MoviePosterNames                           string   `json:"moviePosterNames"`
	MovieFanartNames                           string   `json:"movieFanartNames"`
	TVShowPosterName                           string   `json:"tvShowPosterName"`
	TVShowFanartName                           string   `json:"tvShowFanartName"`
	TVShowPosterNames                          string   `json:"tvShowPosterNames"`
	TVShowFanartNames                          string   `json:"tvShowFanartNames"`
}

type SettingsUpdate struct {
	TMDBAPIKey                                 *string  `json:"tmdbApiKey"`
	ClearTMDBKey                               bool     `json:"clearTmdbKey"`
	ProxyEnabled                               bool     `json:"proxyEnabled"`
	ProxyHost                                  string   `json:"proxyHost"`
	ProxyPort                                  int      `json:"proxyPort"`
	ProxyUsername                              string   `json:"proxyUsername"`
	ProxyPassword                              *string  `json:"proxyPassword"`
	ClearProxyPassword                         bool     `json:"clearProxyPassword"`
	MovieScrapeMetadata                        *bool    `json:"movieScrapeMetadata"`
	MovieScrapeNFO                             *bool    `json:"movieScrapeNfo"`
	MovieScrapeImages                          *bool    `json:"movieScrapeImages"`
	MovieScrapeOverwrite                       *bool    `json:"movieScrapeOverwrite"`
	TVShowScrapeMetadata                       *bool    `json:"tvShowScrapeMetadata"`
	TVShowEpisodeMetadata                      *bool    `json:"tvShowEpisodeMetadata"`
	TVShowScrapeNFO                            *bool    `json:"tvShowScrapeNfo"`
	TVShowScrapeImages                         *bool    `json:"tvShowScrapeImages"`
	TVShowScrapeOverwrite                      *bool    `json:"tvShowScrapeOverwrite"`
	MovieRenameAfterScrape                     *bool    `json:"movieRenameAfterScrape"`
	TVShowRenameAfterScrape                    *bool    `json:"tvShowRenameAfterScrape"`
	MovieScraperFields                         []string `json:"movieScraperFields"`
	TVShowScraperFields                        []string `json:"tvShowScraperFields"`
	TVEpisodeScraperFields                     []string `json:"tvEpisodeScraperFields"`
	MovieRenamerPathname                       string   `json:"movieRenamerPathname"`
	MovieRenamerFilename                       string   `json:"movieRenamerFilename"`
	MovieRenamerPathSpaceSubstitution          *bool    `json:"movieRenamerPathSpaceSubstitution"`
	MovieRenamerPathSpaceReplacement           string   `json:"movieRenamerPathSpaceReplacement"`
	MovieRenamerFilenameSpaceSubstitution      *bool    `json:"movieRenamerFilenameSpaceSubstitution"`
	MovieRenamerFilenameSpaceReplacement       string   `json:"movieRenamerFilenameSpaceReplacement"`
	MovieRenamerColonReplacement               string   `json:"movieRenamerColonReplacement"`
	MovieRenamerAsciiReplacement               *bool    `json:"movieRenamerAsciiReplacement"`
	MovieRenamerFirstCharacterReplacement      string   `json:"movieRenamerFirstCharacterReplacement"`
	MovieRenamerCreateSingleMovieSet           *bool    `json:"movieRenamerCreateSingleMovieSet"`
	MovieRenamerNFOCleanup                     *bool    `json:"movieRenamerNfoCleanup"`
	MovieRenamerCleanupUnwanted                *bool    `json:"movieRenamerCleanupUnwanted"`
	MovieRenamerAllowMerge                     *bool    `json:"movieRenamerAllowMerge"`
	TVShowRenamerShowFolder                    string   `json:"tvShowRenamerShowFolder"`
	TVShowRenamerSeason                        string   `json:"tvShowRenamerSeason"`
	TVShowRenamerFilename                      string   `json:"tvShowRenamerFilename"`
	TVShowRenamerShowFolderSpaceSubstitution   *bool    `json:"tvShowRenamerShowFolderSpaceSubstitution"`
	TVShowRenamerShowFolderSpaceReplacement    string   `json:"tvShowRenamerShowFolderSpaceReplacement"`
	TVShowRenamerSeasonFolderSpaceSubstitution *bool    `json:"tvShowRenamerSeasonFolderSpaceSubstitution"`
	TVShowRenamerSeasonFolderSpaceReplacement  string   `json:"tvShowRenamerSeasonFolderSpaceReplacement"`
	TVShowRenamerFilenameSpaceSubstitution     *bool    `json:"tvShowRenamerFilenameSpaceSubstitution"`
	TVShowRenamerFilenameSpaceReplacement      string   `json:"tvShowRenamerFilenameSpaceReplacement"`
	TVShowRenamerColonReplacement              string   `json:"tvShowRenamerColonReplacement"`
	TVShowRenamerAsciiReplacement              *bool    `json:"tvShowRenamerAsciiReplacement"`
	TVShowRenamerFirstCharacterReplacement     string   `json:"tvShowRenamerFirstCharacterReplacement"`
	TVShowRenamerCleanupUnwanted               *bool    `json:"tvShowRenamerCleanupUnwanted"`
	MoviePosterName                            string   `json:"moviePosterName"`
	MovieFanartName                            string   `json:"movieFanartName"`
	MoviePosterNames                           string   `json:"moviePosterNames"`
	MovieFanartNames                           string   `json:"movieFanartNames"`
	TVShowPosterName                           string   `json:"tvShowPosterName"`
	TVShowFanartName                           string   `json:"tvShowFanartName"`
	TVShowPosterNames                          string   `json:"tvShowPosterNames"`
	TVShowFanartNames                          string   `json:"tvShowFanartNames"`
}

// settingsResponseLocked must be called while s.mu is held; it normalizes stored
// values into the fully-populated response shape expected by the frontend.
func (s *Server) settingsResponseLocked() SettingsResponse {
	source := "none"
	if strings.TrimSpace(s.settings.TMDBAPIKey) != "" {
		source = "settings"
	}
	if strings.TrimSpace(s.config.TMDBKey) != "" {
		source = "environment"
	}
	return SettingsResponse{
		TMDBConfigured:                             s.tmdb.Enabled(),
		TMDBEnabled:                                s.tmdb.Enabled(),
		TMDBKeySource:                              source,
		ProxyEnabled:                               s.settings.ProxyEnabled,
		ProxyHost:                                  s.settings.ProxyHost,
		ProxyPort:                                  s.settings.ProxyPort,
		ProxyUsername:                              s.settings.ProxyUsername,
		ProxyPassword:                              s.settings.ProxyPassword != "",
		MovieScrapeMetadata:                        defaultBool(s.settings.MovieScrapeMetadata, true),
		MovieScrapeNFO:                             defaultBool(s.settings.MovieScrapeNFO, true),
		MovieScrapeImages:                          defaultBool(s.settings.MovieScrapeImages, true),
		MovieScrapeOverwrite:                       defaultBool(s.settings.MovieScrapeOverwrite, false),
		TVShowScrapeMetadata:                       defaultBool(s.settings.TVShowScrapeMetadata, true),
		TVShowEpisodeMetadata:                      defaultBool(s.settings.TVShowEpisodeMetadata, true),
		TVShowScrapeNFO:                            defaultBool(s.settings.TVShowScrapeNFO, true),
		TVShowScrapeImages:                         defaultBool(s.settings.TVShowScrapeImages, true),
		TVShowScrapeOverwrite:                      defaultBool(s.settings.TVShowScrapeOverwrite, false),
		MovieRenameAfterScrape:                     defaultBool(s.settings.MovieRenameAfterScrape, true),
		TVShowRenameAfterScrape:                    defaultBool(s.settings.TVShowRenameAfterScrape, true),
		MovieScraperFields:                         normalizeScraperFields(s.settings.MovieScraperFields, defaultMovieScraperFields()),
		TVShowScraperFields:                        normalizeScraperFields(s.settings.TVShowScraperFields, defaultTVShowScraperFields()),
		TVEpisodeScraperFields:                     normalizeScraperFields(s.settings.TVEpisodeScraperFields, defaultTVEpisodeScraperFields()),
		MovieRenamerPathname:                       defaultString(s.settings.MovieRenamerPathname, defaultMovieRenamerPath),
		MovieRenamerFilename:                       defaultString(s.settings.MovieRenamerFilename, defaultMovieRenamerFile),
		MovieRenamerPathSpaceSubstitution:          defaultBool(s.settings.MovieRenamerPathSpaceSubstitution, false),
		MovieRenamerPathSpaceReplacement:           defaultString(s.settings.MovieRenamerPathSpaceReplacement, "_"),
		MovieRenamerFilenameSpaceSubstitution:      defaultBool(s.settings.MovieRenamerFilenameSpaceSubstitution, false),
		MovieRenamerFilenameSpaceReplacement:       defaultString(s.settings.MovieRenamerFilenameSpaceReplacement, "_"),
		MovieRenamerColonReplacement:               defaultString(s.settings.MovieRenamerColonReplacement, "-"),
		MovieRenamerAsciiReplacement:               defaultBool(s.settings.MovieRenamerAsciiReplacement, false),
		MovieRenamerFirstCharacterReplacement:      defaultString(s.settings.MovieRenamerFirstCharacterReplacement, "#"),
		MovieRenamerCreateSingleMovieSet:           defaultBool(s.settings.MovieRenamerCreateSingleMovieSet, false),
		MovieRenamerNFOCleanup:                     defaultBool(s.settings.MovieRenamerNFOCleanup, false),
		MovieRenamerCleanupUnwanted:                defaultBool(s.settings.MovieRenamerCleanupUnwanted, false),
		MovieRenamerAllowMerge:                     defaultBool(s.settings.MovieRenamerAllowMerge, false),
		TVShowRenamerShowFolder:                    defaultString(s.settings.TVShowRenamerShowFolder, defaultTVShowRenamerPath),
		TVShowRenamerSeason:                        defaultString(s.settings.TVShowRenamerSeason, defaultTVSeasonRenamer),
		TVShowRenamerFilename:                      defaultString(s.settings.TVShowRenamerFilename, defaultTVEpisodeRenamer),
		TVShowRenamerShowFolderSpaceSubstitution:   defaultBool(s.settings.TVShowRenamerShowFolderSpaceSubstitution, false),
		TVShowRenamerShowFolderSpaceReplacement:    defaultString(s.settings.TVShowRenamerShowFolderSpaceReplacement, "_"),
		TVShowRenamerSeasonFolderSpaceSubstitution: defaultBool(s.settings.TVShowRenamerSeasonFolderSpaceSubstitution, false),
		TVShowRenamerSeasonFolderSpaceReplacement:  defaultString(s.settings.TVShowRenamerSeasonFolderSpaceReplacement, "_"),
		TVShowRenamerFilenameSpaceSubstitution:     defaultBool(s.settings.TVShowRenamerFilenameSpaceSubstitution, false),
		TVShowRenamerFilenameSpaceReplacement:      defaultString(s.settings.TVShowRenamerFilenameSpaceReplacement, "_"),
		TVShowRenamerColonReplacement:              defaultString(s.settings.TVShowRenamerColonReplacement, " "),
		TVShowRenamerAsciiReplacement:              defaultBool(s.settings.TVShowRenamerAsciiReplacement, false),
		TVShowRenamerFirstCharacterReplacement:     defaultString(s.settings.TVShowRenamerFirstCharacterReplacement, "#"),
		TVShowRenamerCleanupUnwanted:               defaultBool(s.settings.TVShowRenamerCleanupUnwanted, false),
		MoviePosterName:                            defaultString(s.settings.MoviePosterName, "poster.jpg"),
		MovieFanartName:                            defaultString(s.settings.MovieFanartName, "fanart.jpg"),
		MoviePosterNames:                           defaultImageNames(s.settings.MoviePosterNames, s.settings.MoviePosterName, defaultMoviePosterNames()),
		MovieFanartNames:                           defaultImageNames(s.settings.MovieFanartNames, s.settings.MovieFanartName, defaultMovieFanartNames()),
		TVShowPosterName:                           defaultString(s.settings.TVShowPosterName, "poster.jpg"),
		TVShowFanartName:                           defaultString(s.settings.TVShowFanartName, "fanart.jpg"),
		TVShowPosterNames:                          defaultImageNames(s.settings.TVShowPosterNames, s.settings.TVShowPosterName, defaultTVShowPosterNames()),
		TVShowFanartNames:                          defaultImageNames(s.settings.TVShowFanartNames, s.settings.TVShowFanartName, defaultTVShowFanartNames()),
	}
}

func (s *Server) appSettings() AppSettings {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.settings
}

func defaultBool(value *bool, fallback bool) bool {
	if value == nil {
		return fallback
	}
	return *value
}

func defaultString(value string, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func normalizeReplacement(value string, fallback string, allowed []string) string {
	for _, candidate := range allowed {
		if value == candidate {
			return value
		}
	}
	return fallback
}

func movieFolderRenameOptions(settings AppSettings) media.RenameOptions {
	return media.RenameOptions{
		SpaceSubstitution:               defaultBool(settings.MovieRenamerPathSpaceSubstitution, false),
		SpaceReplacement:                defaultString(settings.MovieRenamerPathSpaceReplacement, "_"),
		ColonReplacement:                defaultString(settings.MovieRenamerColonReplacement, "-"),
		ColonReplacementDefined:         true,
		ASCIIReplacement:                defaultBool(settings.MovieRenamerAsciiReplacement, false),
		FirstCharacterNumberReplacement: defaultString(settings.MovieRenamerFirstCharacterReplacement, "#"),
	}
}

func movieFileRenameOptions(settings AppSettings) media.RenameOptions {
	return media.RenameOptions{
		SpaceSubstitution:               defaultBool(settings.MovieRenamerFilenameSpaceSubstitution, false),
		SpaceReplacement:                defaultString(settings.MovieRenamerFilenameSpaceReplacement, "_"),
		ColonReplacement:                defaultString(settings.MovieRenamerColonReplacement, "-"),
		ColonReplacementDefined:         true,
		ASCIIReplacement:                defaultBool(settings.MovieRenamerAsciiReplacement, false),
		FirstCharacterNumberReplacement: defaultString(settings.MovieRenamerFirstCharacterReplacement, "#"),
	}
}

func tvShowFolderRenameOptions(settings AppSettings) media.RenameOptions {
	return media.RenameOptions{
		SpaceSubstitution:               defaultBool(settings.TVShowRenamerShowFolderSpaceSubstitution, false),
		SpaceReplacement:                defaultString(settings.TVShowRenamerShowFolderSpaceReplacement, "_"),
		ColonReplacement:                defaultString(settings.TVShowRenamerColonReplacement, " "),
		ColonReplacementDefined:         true,
		ASCIIReplacement:                defaultBool(settings.TVShowRenamerAsciiReplacement, false),
		FirstCharacterNumberReplacement: defaultString(settings.TVShowRenamerFirstCharacterReplacement, "#"),
	}
}

func tvSeasonFolderRenameOptions(settings AppSettings) media.RenameOptions {
	return media.RenameOptions{
		SpaceSubstitution:               defaultBool(settings.TVShowRenamerSeasonFolderSpaceSubstitution, false),
		SpaceReplacement:                defaultString(settings.TVShowRenamerSeasonFolderSpaceReplacement, "_"),
		ColonReplacement:                defaultString(settings.TVShowRenamerColonReplacement, " "),
		ColonReplacementDefined:         true,
		ASCIIReplacement:                defaultBool(settings.TVShowRenamerAsciiReplacement, false),
		FirstCharacterNumberReplacement: defaultString(settings.TVShowRenamerFirstCharacterReplacement, "#"),
	}
}

func tvEpisodeFileRenameOptions(settings AppSettings) media.RenameOptions {
	return media.RenameOptions{
		SpaceSubstitution:               defaultBool(settings.TVShowRenamerFilenameSpaceSubstitution, false),
		SpaceReplacement:                defaultString(settings.TVShowRenamerFilenameSpaceReplacement, "_"),
		ColonReplacement:                defaultString(settings.TVShowRenamerColonReplacement, " "),
		ColonReplacementDefined:         true,
		ASCIIReplacement:                defaultBool(settings.TVShowRenamerAsciiReplacement, false),
		FirstCharacterNumberReplacement: defaultString(settings.TVShowRenamerFirstCharacterReplacement, "#"),
	}
}

func tvEpisodeMetadataForItem(season tmdb.TVSeason, item media.Item) (tmdb.TVEpisode, bool) {
	targets := map[int]bool{}
	if item.Episode > 0 {
		targets[item.Episode] = true
	}
	for _, episode := range item.Episodes {
		if episode > 0 {
			targets[episode] = true
		}
	}
	if len(targets) == 0 {
		return tmdb.TVEpisode{}, false
	}
	for _, episode := range season.Episodes {
		if targets[episode.EpisodeNumber] {
			return episode, true
		}
	}
	return tmdb.TVEpisode{}, false
}

func tvEpisodeRenameTitle(item media.Item, showTitle string, loadSeasonMetadataFor func(int) (tmdb.TVSeason, error)) string {
	if item.Season > 0 {
		if seasonData, err := loadSeasonMetadataFor(item.Season); err == nil {
			if episode, ok := tvEpisodeMetadataForItem(seasonData, item); ok && strings.TrimSpace(episode.Title) != "" {
				return episode.Title
			}
		}
	}
	if item.Episode > 0 {
		return fmt.Sprintf("%02d", item.Episode)
	}
	for _, episode := range item.Episodes {
		if episode > 0 {
			return fmt.Sprintf("%02d", episode)
		}
	}
	title := strings.TrimSpace(item.MatchedName)
	if title != "" && !strings.EqualFold(title, strings.TrimSpace(showTitle)) && !strings.EqualFold(title, strings.TrimSpace(item.ShowGuess)) {
		return title
	}
	return ""
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func yearFromDate(value string) string {
	if len(value) >= 4 {
		return value[:4]
	}
	return ""
}

func defaultImageNames(value string, legacy string, fallback []string) string {
	value = strings.TrimSpace(value)
	if value != "" {
		if len(fallback) > 1 && value == fallback[0] {
			return strings.Join(fallback, "\n")
		}
		return value
	}
	legacy = strings.TrimSpace(legacy)
	if legacy != "" && (len(fallback) == 0 || legacy != fallback[0]) {
		return legacy
	}
	return strings.Join(fallback, "\n")
}

func defaultMovieScraperFields() []string {
	return []string{
		"ID", "TITLE", "ORIGINAL_TITLE", "TAGLINE", "PLOT", "YEAR", "RELEASE_DATE", "RATING", "TOP250", "RUNTIME",
		"CERTIFICATION", "GENRES", "SPOKEN_LANGUAGES", "COUNTRY", "PRODUCTION_COMPANY", "TAGS", "COLLECTION", "TRAILER",
		"ACTORS", "PRODUCERS", "DIRECTORS", "WRITERS",
		"POSTER", "FANART", "BANNER", "CLEARART", "THUMB", "CLEARLOGO", "DISCART", "KEYART", "EXTRAFANART", "EXTRATHUMB",
	}
}

func defaultTVShowScraperFields() []string {
	return []string{
		"ID", "TITLE", "ORIGINAL_TITLE", "PLOT", "YEAR", "AIRED", "STATUS", "RATING", "TOP250", "RUNTIME", "CERTIFICATION",
		"GENRES", "COUNTRY", "STUDIO", "TAGS", "TRAILER", "SEASON_NAMES", "SEASON_OVERVIEW",
		"ACTORS",
		"POSTER", "FANART", "BANNER", "CLEARART", "THUMB", "CLEARLOGO", "DISCART", "KEYART", "CHARACTERART", "EXTRAFANART",
		"SEASON_POSTER", "SEASON_FANART", "SEASON_BANNER", "SEASON_THUMB", "THEME",
	}
}

func defaultTVEpisodeScraperFields() []string {
	return []string{
		"TITLE", "ORIGINAL_TITLE", "PLOT", "SEASON_EPISODE", "AIRED", "RATING", "TAGS",
		"ACTORS", "DIRECTORS", "WRITERS", "THUMB",
	}
}

func normalizeScraperFields(values []string, fallback []string) []string {
	if values == nil {
		return append([]string(nil), fallback...)
	}
	allowed := make(map[string]bool, len(fallback))
	for _, value := range fallback {
		allowed[value] = true
	}
	seen := map[string]bool{}
	var normalized []string
	for _, value := range values {
		value = strings.ToUpper(strings.TrimSpace(value))
		if value == "" || !allowed[value] || seen[value] {
			continue
		}
		seen[value] = true
		normalized = append(normalized, value)
	}
	if len(normalized) == 0 && len(values) > 0 {
		return []string{}
	}
	if len(normalized) == 0 {
		return append([]string(nil), fallback...)
	}
	return normalized
}

func scraperFieldEnabled(fields []string, key string) bool {
	if len(fields) == 0 {
		return false
	}
	key = strings.ToUpper(strings.TrimSpace(key))
	for _, field := range fields {
		if strings.EqualFold(field, key) {
			return true
		}
	}
	return false
}

func (s *Server) tmdbClient() tmdb.Client {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.tmdb
}

func loadAppSettings(dataDir string) (AppSettings, error) {
	settings, err := readJSONFile[AppSettings](settingsPath(dataDir))
	if err != nil {
		if os.IsNotExist(err) {
			return AppSettings{}, nil
		}
		return AppSettings{}, err
	}
	settings.TMDBAPIKey = strings.TrimSpace(settings.TMDBAPIKey)
	settings.ProxyHost = strings.TrimSpace(settings.ProxyHost)
	settings.ProxyUsername = strings.TrimSpace(settings.ProxyUsername)
	return settings, nil
}

func saveAppSettings(dataDir string, settings AppSettings) error {
	settings.TMDBAPIKey = strings.TrimSpace(settings.TMDBAPIKey)
	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(settingsPath(dataDir), data, 0600)
}

func settingsPath(dataDir string) string {
	return filepath.Join(dataDir, "settings.json")
}

func newTMDBClient(config Config, settings AppSettings) (tmdb.Client, error) {
	tmdbKey := strings.TrimSpace(config.TMDBKey)
	if tmdbKey == "" {
		tmdbKey = strings.TrimSpace(settings.TMDBAPIKey)
	}
	client := config.Client
	if settings.ProxyEnabled {
		proxyURL, err := buildProxyURL(settings)
		if err != nil {
			return tmdb.Client{}, err
		}
		client = &http.Client{
			Timeout: clientTimeout(config.Client),
			Transport: &http.Transport{
				Proxy: http.ProxyURL(proxyURL),
			},
		}
	}
	return tmdb.Client{Key: tmdbKey, HTTP: client, Lang: "zh-CN"}, nil
}

func buildProxyURL(settings AppSettings) (*url.URL, error) {
	host := strings.TrimSpace(settings.ProxyHost)
	if host == "" {
		return nil, fmt.Errorf("proxy host is required")
	}
	port := settings.ProxyPort
	if port < 0 || port > 65535 {
		return nil, fmt.Errorf("proxy port must be between 1 and 65535")
	}
	if port == 0 {
		port = 80
	}
	proxyURL := &url.URL{
		Scheme: "http",
		Host:   net.JoinHostPort(host, strconv.Itoa(port)),
	}
	if settings.ProxyUsername != "" {
		if settings.ProxyPassword != "" {
			proxyURL.User = url.UserPassword(settings.ProxyUsername, settings.ProxyPassword)
		} else {
			proxyURL.User = url.User(settings.ProxyUsername)
		}
	}
	return proxyURL, nil
}

func clientTimeout(client *http.Client) time.Duration {
	if client != nil && client.Timeout > 0 {
		return client.Timeout
	}
	return 20 * time.Second
}
