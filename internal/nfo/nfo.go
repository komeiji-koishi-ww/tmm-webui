package nfo

import (
	"encoding/xml"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"tmmweb/internal/tmdb"
)

type xmlInt int

func (v *xmlInt) UnmarshalXML(decoder *xml.Decoder, start xml.StartElement) error {
	var raw string
	if err := decoder.DecodeElement(&raw, &start); err != nil {
		return err
	}
	*v = parseXMLInt(raw)
	return nil
}

func (v *xmlInt) UnmarshalXMLAttr(attr xml.Attr) error {
	*v = parseXMLInt(attr.Value)
	return nil
}

func (v xmlInt) MarshalText() ([]byte, error) {
	return []byte(strconv.Itoa(int(v))), nil
}

func parseXMLInt(raw string) xmlInt {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0
	}
	if parsed, err := strconv.Atoi(raw); err == nil {
		return xmlInt(parsed)
	}
	if parsed, err := strconv.ParseFloat(raw, 64); err == nil {
		return xmlInt(parsed)
	}
	return 0
}

type movieNFO struct {
	XMLName          xml.Name     `xml:"movie"`
	Title            string       `xml:"title"`
	Original         string       `xml:"originaltitle,omitempty"`
	SortTitle        string       `xml:"sorttitle,omitempty"`
	Year             string       `xml:"year,omitempty"`
	Plot             string       `xml:"plot,omitempty"`
	Runtime          xmlInt       `xml:"runtime,omitempty"`
	Rating           string       `xml:"rating,omitempty"`
	Ratings          ratings      `xml:"ratings,omitempty"`
	UserRating       string       `xml:"userrating"`
	ID               string       `xml:"id,omitempty"`
	TMDBID           xmlInt       `xml:"tmdbid,omitempty"`
	UniqueIDs        []unique     `xml:"uniqueid"`
	Genres           []string     `xml:"genre,omitempty"`
	Premiered        string       `xml:"premiered,omitempty"`
	Thumb            string       `xml:"thumb,omitempty"`
	Fanart           string       `xml:"fanart>thumb,omitempty"`
	Watched          bool         `xml:"watched"`
	PlayCount        string       `xml:"playcount"`
	Studios          []string     `xml:"studio,omitempty"`
	Actors           []actorNFO   `xml:"actor,omitempty"`
	Trailer          string       `xml:"trailer"`
	DateAdded        string       `xml:"dateadded,omitempty"`
	FileInfo         *fileInfoNFO `xml:"fileinfo,omitempty"`
	TMMComment       string       `xml:",comment"`
	Source           string       `xml:"source"`
	OriginalFileName string       `xml:"original_filename,omitempty"`
	UserNote         string       `xml:"user_note"`
}

type tvShowNFO struct {
	XMLName   xml.Name `xml:"tvshow"`
	Title     string   `xml:"title"`
	Original  string   `xml:"originaltitle,omitempty"`
	SortTitle string   `xml:"sorttitle,omitempty"`
	Year      string   `xml:"year,omitempty"`
	Plot      string   `xml:"plot,omitempty"`
	Rating    string   `xml:"rating,omitempty"`
	Ratings   ratings  `xml:"ratings,omitempty"`
	ID        string   `xml:"id,omitempty"`
	TMDBID    xmlInt   `xml:"tmdbid,omitempty"`
	UniqueIDs []unique `xml:"uniqueid"`
	Genres    []string `xml:"genre,omitempty"`
	Premiered string   `xml:"premiered,omitempty"`
	Thumb     string   `xml:"thumb,omitempty"`
	Fanart    string   `xml:"fanart>thumb,omitempty"`
}

type tvSeasonNFO struct {
	XMLName      xml.Name `xml:"season"`
	SeasonNumber xmlInt   `xml:"seasonnumber"`
	Title        string   `xml:"title"`
	ShowTitle    string   `xml:"showtitle,omitempty"`
	SortTitle    string   `xml:"sorttitle,omitempty"`
	Year         string   `xml:"year,omitempty"`
	Plot         string   `xml:"plot,omitempty"`
	Rating       string   `xml:"rating,omitempty"`
	Ratings      ratings  `xml:"ratings,omitempty"`
	TMDBID       xmlInt   `xml:"tmdbid,omitempty"`
	UniqueIDs    []unique `xml:"uniqueid"`
	Premiered    string   `xml:"premiered,omitempty"`
	Thumb        string   `xml:"thumb,omitempty"`
	Fanart       string   `xml:"fanart>thumb,omitempty"`
}

type tvEpisodeNFO struct {
	XMLName          xml.Name         `xml:"episodedetails"`
	Title            string           `xml:"title"`
	ShowTitle        string           `xml:"showtitle,omitempty"`
	Original         string           `xml:"originaltitle,omitempty"`
	Season           xmlInt           `xml:"season"`
	Episode          xmlInt           `xml:"episode"`
	DisplaySeason    xmlInt           `xml:"displayseason,omitempty"`
	DisplayEpisode   xmlInt           `xml:"displayepisode,omitempty"`
	ID               string           `xml:"id,omitempty"`
	Plot             string           `xml:"plot,omitempty"`
	Runtime          xmlInt           `xml:"runtime,omitempty"`
	Aired            string           `xml:"aired,omitempty"`
	Premiered        string           `xml:"premiered,omitempty"`
	Rating           string           `xml:"rating,omitempty"`
	Ratings          ratings          `xml:"ratings,omitempty"`
	UserRating       string           `xml:"userrating"`
	TMDBID           xmlInt           `xml:"tmdbid,omitempty"`
	UniqueIDs        []unique         `xml:"uniqueid"`
	Thumb            string           `xml:"thumb,omitempty"`
	MPAA             string           `xml:"mpaa"`
	Watched          bool             `xml:"watched"`
	PlayCount        string           `xml:"playcount"`
	Studios          []string         `xml:"studio,omitempty"`
	Credits          []personRef      `xml:"credits,omitempty"`
	Directors        []personRef      `xml:"director,omitempty"`
	Actors           []actorNFO       `xml:"actor,omitempty"`
	Trailer          string           `xml:"trailer"`
	DateAdded        string           `xml:"dateadded,omitempty"`
	EpBookmark       string           `xml:"epbookmark"`
	Code             string           `xml:"code"`
	FileInfo         *fileInfoNFO     `xml:"fileinfo,omitempty"`
	TMMComment       string           `xml:",comment"`
	Source           string           `xml:"source"`
	OriginalFileName string           `xml:"original_filename,omitempty"`
	UserNote         string           `xml:"user_note"`
	EpisodeGroups    episodeGroupsNFO `xml:"episode_groups"`
}

type unique struct {
	Type    string `xml:"type,attr"`
	Default string `xml:"default,attr,omitempty"`
	Value   string `xml:",chardata"`
}

type rating struct {
	Name    string `xml:"name,attr"`
	Default string `xml:"default,attr"`
	Max     string `xml:"max,attr,omitempty"`
	Value   string `xml:"value"`
	Votes   xmlInt `xml:"votes,omitempty"`
}

type ratings struct {
	Items []rating `xml:"rating"`
}

type personRef struct {
	TMDBID xmlInt `xml:"tmdbid,attr,omitempty"`
	Value  string `xml:",chardata"`
}

type actorNFO struct {
	Name    string `xml:"name"`
	Role    string `xml:"role,omitempty"`
	Thumb   string `xml:"thumb,omitempty"`
	Profile string `xml:"profile,omitempty"`
	Type    string `xml:"type,omitempty"`
	TMDBID  xmlInt `xml:"tmdbid,omitempty"`
}

type fileInfoNFO struct {
	StreamDetails streamDetailsNFO `xml:"streamdetails"`
}

type streamDetailsNFO struct {
	Videos    []videoNFO    `xml:"video,omitempty"`
	Audios    []audioNFO    `xml:"audio,omitempty"`
	Subtitles []subtitleNFO `xml:"subtitle,omitempty"`
}

type videoNFO struct {
	Codec             string  `xml:"codec,omitempty"`
	Aspect            float64 `xml:"aspect,omitempty"`
	Width             xmlInt  `xml:"width,omitempty"`
	Height            xmlInt  `xml:"height,omitempty"`
	DurationInSeconds xmlInt  `xml:"durationinseconds,omitempty"`
	HDRType           string  `xml:"hdrtype,omitempty"`
	StereoMode        string  `xml:"stereomode"`
}

type audioNFO struct {
	Codec    string `xml:"codec,omitempty"`
	Language string `xml:"language,omitempty"`
	Channels xmlInt `xml:"channels,omitempty"`
}

type subtitleNFO struct {
	Language string `xml:"language,omitempty"`
}

type episodeGroupsNFO struct {
	Groups []episodeGroupNFO `xml:"group"`
}

type episodeGroupNFO struct {
	ID      string `xml:"id,attr"`
	Name    string `xml:"name,attr"`
	Season  xmlInt `xml:"season,attr"`
	Episode xmlInt `xml:"episode,attr"`
}

type EpisodeFileInfo struct {
	FileName        string
	DateAdded       string
	VideoStreams    []VideoStream
	AudioStreams    []AudioStream
	SubtitleStreams []SubtitleStream
}

type VideoStream struct {
	Codec           string
	Aspect          float64
	Width           int
	Height          int
	DurationSeconds int
	StereoMode      string
	HDRType         string
}

type AudioStream struct {
	Codec    string
	Language string
	Channels int
}

type SubtitleStream struct {
	Language string
}

type Summary struct {
	Type            string
	Title           string
	Original        string
	Year            string
	Plot            string
	Runtime         int
	Rating          float64
	Genres          []string
	Actors          []string
	Premiered       string
	TMDBID          int
	IMDBID          string
	SeasonNumber    int
	VideoStreams    []VideoStream
	AudioStreams    []AudioStream
	SubtitleStreams []SubtitleStream
}

func WriteMovie(dir string, movie tmdb.Movie, fileInfo EpisodeFileInfo) error {
	value := movieNFO{
		Title: movie.Title, Original: movie.Original, SortTitle: movie.Title,
		Year: year(movie.ReleaseDate), Plot: movie.Overview, Runtime: xmlInt(movie.Runtime),
		Rating: strconv.FormatFloat(movie.VoteAverage, 'f', 1, 64),
		ID:     strconv.Itoa(movie.ID), TMDBID: xmlInt(movie.ID),
		Genres: movie.Genres, Premiered: movie.ReleaseDate,
		Thumb:            tmdbImageURL(movie.PosterPath, "original"),
		Fanart:           tmdbImageURL(movie.BackdropPath, "original"),
		Studios:          movie.Studios,
		Actors:           actorRefs(movie.CastPeople),
		DateAdded:        fileInfo.DateAdded,
		FileInfo:         episodeFileInfo(fileInfo),
		TMMComment:       "tinyMediaManager meta data",
		Source:           "UNKNOWN",
		OriginalFileName: fileInfo.FileName,
	}
	value.UniqueIDs = append(value.UniqueIDs, unique{Type: "tmdb", Default: "true", Value: strconv.Itoa(movie.ID)})
	if movie.ImdbID != "" {
		value.UniqueIDs = append(value.UniqueIDs, unique{Type: "imdb", Value: movie.ImdbID})
	}
	if movie.VoteAverage > 0 {
		value.Ratings.Items = append(value.Ratings.Items, rating{
			Name: "themoviedb", Default: "true", Max: "10",
			Value: strconv.FormatFloat(movie.VoteAverage, 'f', 1, 64),
			Votes: xmlInt(movie.VoteCount),
		})
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
		SeasonNumber: xmlInt(season.SeasonNumber),
		Title:        seasonTitle(season),
		ShowTitle:    season.ShowTitle,
		SortTitle:    seasonTitle(season),
		Year:         year(season.AirDate),
		Plot:         season.Overview,
		Rating:       strconv.FormatFloat(season.VoteAverage, 'f', 1, 64),
		TMDBID:       xmlInt(season.ShowID),
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

func WriteTVEpisode(path string, show tmdb.TVShow, episode tmdb.TVEpisode, fileInfo EpisodeFileInfo) error {
	value := tvEpisodeNFO{
		Title:            episode.Title,
		Original:         episode.Title,
		ShowTitle:        show.Title,
		Season:           xmlInt(episode.SeasonNumber),
		Episode:          xmlInt(episode.EpisodeNumber),
		DisplaySeason:    -1,
		DisplayEpisode:   -1,
		ID:               strconv.Itoa(episode.ID),
		Plot:             episode.Overview,
		Aired:            episode.AirDate,
		Premiered:        episode.AirDate,
		Rating:           strconv.FormatFloat(episode.VoteAverage, 'f', 1, 64),
		Runtime:          xmlInt(episode.Runtime),
		TMDBID:           xmlInt(episode.ID),
		Thumb:            tmdbImageURL(episode.StillPath, "original"),
		Studios:          show.Studios,
		DateAdded:        fileInfo.DateAdded,
		Source:           "UNKNOWN",
		OriginalFileName: fileInfo.FileName,
		TMMComment:       "tinyMediaManager meta data",
	}
	if value.Runtime == 0 && len(fileInfo.VideoStreams) > 0 {
		value.Runtime = xmlInt(fileInfo.VideoStreams[0].DurationSeconds / 60)
	}
	value.UniqueIDs = append(value.UniqueIDs, unique{Type: "tmdb", Default: "true", Value: strconv.Itoa(episode.ID)})
	if episode.VoteAverage > 0 {
		value.Ratings.Items = append(value.Ratings.Items, rating{
			Name: "themoviedb", Default: "true", Max: "10",
			Value: strconv.FormatFloat(episode.VoteAverage, 'f', 1, 64),
			Votes: xmlInt(episode.VoteCount),
		})
	}
	value.Credits = personRefs(episode.Writers)
	value.Directors = personRefs(episode.Directors)
	value.Actors = actorRefs(append(append([]tmdb.Person{}, show.CastPeople...), episode.Actors...))
	value.FileInfo = episodeFileInfo(fileInfo)
	value.EpisodeGroups.Groups = append(value.EpisodeGroups.Groups,
		episodeGroupNFO{ID: "AIRED", Season: xmlInt(episode.SeasonNumber), Episode: xmlInt(episode.EpisodeNumber)},
		episodeGroupNFO{ID: "DISPLAY", Season: -1, Episode: -1},
	)
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

func personRefs(values []tmdb.Person) []personRef {
	refs := make([]personRef, 0, len(values))
	seen := map[int]bool{}
	for _, value := range values {
		name := strings.TrimSpace(value.Name)
		if name == "" {
			continue
		}
		if value.ID > 0 {
			if seen[value.ID] {
				continue
			}
			seen[value.ID] = true
		}
		refs = append(refs, personRef{TMDBID: xmlInt(value.ID), Value: name})
	}
	return refs
}

func actorRefs(values []tmdb.Person) []actorNFO {
	actors := make([]actorNFO, 0, len(values))
	seen := map[int]bool{}
	for _, value := range values {
		name := strings.TrimSpace(value.Name)
		if name == "" {
			continue
		}
		if value.ID > 0 {
			if seen[value.ID] {
				continue
			}
			seen[value.ID] = true
		}
		actor := actorNFO{
			Name: name, Role: strings.TrimSpace(value.Role),
			Thumb:   tmdbImageURL(value.ProfilePath, "h632"),
			Profile: tmdbPersonURL(value.ID),
			TMDBID:  xmlInt(value.ID),
		}
		if value.Guest {
			actor.Type = "GuestStar"
		}
		actors = append(actors, actor)
	}
	return actors
}

func episodeFileInfo(info EpisodeFileInfo) *fileInfoNFO {
	if len(info.VideoStreams) == 0 && len(info.AudioStreams) == 0 && len(info.SubtitleStreams) == 0 {
		return nil
	}
	fileInfo := &fileInfoNFO{}
	for _, stream := range info.VideoStreams {
		fileInfo.StreamDetails.Videos = append(fileInfo.StreamDetails.Videos, videoNFO{
			Codec: stream.Codec, Aspect: stream.Aspect, Width: xmlInt(stream.Width), Height: xmlInt(stream.Height),
			DurationInSeconds: xmlInt(stream.DurationSeconds), HDRType: stream.HDRType, StereoMode: stream.StereoMode,
		})
	}
	for _, stream := range info.AudioStreams {
		fileInfo.StreamDetails.Audios = append(fileInfo.StreamDetails.Audios, audioNFO{
			Codec: stream.Codec, Language: stream.Language, Channels: xmlInt(stream.Channels),
		})
	}
	for _, stream := range info.SubtitleStreams {
		fileInfo.StreamDetails.Subtitles = append(fileInfo.StreamDetails.Subtitles, subtitleNFO{
			Language: stream.Language,
		})
	}
	return fileInfo
}

func tmdbImageURL(path string, size string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return ""
	}
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return path
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return "https://image.tmdb.org/t/p/" + size + path
}

func tmdbPersonURL(id int) string {
	if id <= 0 {
		return ""
	}
	return "https://www.themoviedb.org/person/" + strconv.Itoa(id)
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
		summary := summaryFromMovie(value)
		summary.Type = "movie"
		return summary, nil
	case "tvshow":
		var value tvShowNFO
		if err := xml.Unmarshal(data, &value); err != nil {
			return Summary{}, err
		}
		summary := summaryFromTVShow(value)
		summary.Type = "tvshow"
		return summary, nil
	case "season":
		var value tvSeasonNFO
		if err := xml.Unmarshal(data, &value); err != nil {
			return Summary{}, err
		}
		summary := summaryFromSeason(value)
		summary.Type = "season"
		return summary, nil
	case "episodedetails":
		var value tvEpisodeNFO
		if err := xml.Unmarshal(data, &value); err != nil {
			return Summary{}, err
		}
		summary := summaryFromEpisode(value)
		summary.Type = "episode"
		return summary, nil
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
		Runtime:   int(value.Runtime),
		Rating:    ratingValue(value.Rating, value.Ratings.Items),
		Genres:    compactStrings(value.Genres),
		Actors:    actorNames(value.Actors),
		Premiered: strings.TrimSpace(value.Premiered),
		TMDBID:    int(value.TMDBID),
	}
	applyUniqueIDs(&summary, value.UniqueIDs)
	applyLegacyID(&summary, value.ID)
	applyFileInfoSummary(&summary, value.FileInfo)
	return summary
}

func summaryFromTVShow(value tvShowNFO) Summary {
	summary := Summary{
		Title:     strings.TrimSpace(value.Title),
		Original:  strings.TrimSpace(value.Original),
		Year:      firstNonEmpty(strings.TrimSpace(value.Year), year(value.Premiered)),
		Plot:      strings.TrimSpace(value.Plot),
		Rating:    ratingValue(value.Rating, value.Ratings.Items),
		Genres:    compactStrings(value.Genres),
		Premiered: strings.TrimSpace(value.Premiered),
		TMDBID:    int(value.TMDBID),
	}
	applyUniqueIDs(&summary, value.UniqueIDs)
	applyLegacyID(&summary, value.ID)
	return summary
}

func summaryFromSeason(value tvSeasonNFO) Summary {
	summary := Summary{
		Title:        strings.TrimSpace(value.Title),
		Original:     strings.TrimSpace(value.SortTitle),
		Year:         firstNonEmpty(strings.TrimSpace(value.Year), year(value.Premiered)),
		Plot:         strings.TrimSpace(value.Plot),
		Rating:       ratingValue(value.Rating, value.Ratings.Items),
		Premiered:    strings.TrimSpace(value.Premiered),
		TMDBID:       int(value.TMDBID),
		SeasonNumber: int(value.SeasonNumber),
	}
	applyUniqueIDs(&summary, value.UniqueIDs)
	return summary
}

func summaryFromEpisode(value tvEpisodeNFO) Summary {
	summary := Summary{
		Title:     strings.TrimSpace(value.Title),
		Original:  strings.TrimSpace(value.Original),
		Plot:      strings.TrimSpace(value.Plot),
		Runtime:   int(value.Runtime),
		Rating:    ratingValue(value.Rating, value.Ratings.Items),
		Actors:    actorNames(value.Actors),
		Premiered: firstNonEmpty(strings.TrimSpace(value.Premiered), strings.TrimSpace(value.Aired)),
		TMDBID:    int(value.TMDBID),
	}
	summary.Year = year(summary.Premiered)
	applyUniqueIDs(&summary, value.UniqueIDs)
	applyLegacyID(&summary, value.ID)
	applyFileInfoSummary(&summary, value.FileInfo)
	return summary
}

func applyFileInfoSummary(summary *Summary, fileInfo *fileInfoNFO) {
	if fileInfo == nil {
		return
	}
	for _, stream := range fileInfo.StreamDetails.Videos {
		summary.VideoStreams = append(summary.VideoStreams, VideoStream{
			Codec: stream.Codec, Aspect: stream.Aspect, Width: int(stream.Width), Height: int(stream.Height),
			DurationSeconds: int(stream.DurationInSeconds), StereoMode: stream.StereoMode, HDRType: stream.HDRType,
		})
	}
	for _, stream := range fileInfo.StreamDetails.Audios {
		summary.AudioStreams = append(summary.AudioStreams, AudioStream{
			Codec: stream.Codec, Language: stream.Language, Channels: int(stream.Channels),
		})
	}
	for _, stream := range fileInfo.StreamDetails.Subtitles {
		summary.SubtitleStreams = append(summary.SubtitleStreams, SubtitleStream{Language: stream.Language})
	}
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

func applyLegacyID(summary *Summary, value string) {
	id := strings.TrimSpace(value)
	if id == "" {
		return
	}
	if strings.HasPrefix(id, "tt") {
		if summary.IMDBID == "" {
			summary.IMDBID = id
		}
		return
	}
	if summary.TMDBID == 0 {
		if parsed, err := strconv.Atoi(id); err == nil {
			summary.TMDBID = parsed
		}
	}
}

func actorNames(values []actorNFO) []string {
	names := make([]string, 0, len(values))
	seen := map[string]struct{}{}
	for _, value := range values {
		name := strings.TrimSpace(value.Name)
		if name == "" {
			continue
		}
		key := strings.ToLower(name)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		names = append(names, name)
		if len(names) >= 12 {
			break
		}
	}
	return names
}

func parseRating(value string) float64 {
	rating, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
	if err != nil {
		return 0
	}
	return rating
}

func ratingValue(legacy string, ratings []rating) float64 {
	for _, rating := range ratings {
		if strings.EqualFold(strings.TrimSpace(rating.Default), "true") {
			if value := parseRating(rating.Value); value > 0 {
				return value
			}
		}
	}
	for _, name := range []string{"themoviedb", "tmdb", "imdb"} {
		for _, rating := range ratings {
			if strings.EqualFold(strings.TrimSpace(rating.Name), name) {
				if value := parseRating(rating.Value); value > 0 {
					return value
				}
			}
		}
	}
	return parseRating(legacy)
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
