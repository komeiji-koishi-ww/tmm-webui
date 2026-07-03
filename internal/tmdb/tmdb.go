package tmdb

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

const baseURL = "https://api.themoviedb.org/3"
const imageBaseURL = "https://image.tmdb.org/t/p"

type Client struct {
	Key    string
	HTTP   *http.Client
	Lang   string
	Region string
}

type SearchResult struct {
	ID           int      `json:"id"`
	Title        string   `json:"title"`
	Original     string   `json:"originalTitle"`
	Overview     string   `json:"overview"`
	ReleaseDate  string   `json:"releaseDate"`
	PosterPath   string   `json:"posterPath"`
	BackdropPath string   `json:"backdropPath"`
	VoteAverage  float64  `json:"voteAverage"`
	MediaType    string   `json:"mediaType"`
	Cast         []string `json:"cast,omitempty"`
}

type Movie struct {
	ID           int      `json:"id"`
	Title        string   `json:"title"`
	Original     string   `json:"originalTitle"`
	Overview     string   `json:"overview"`
	ReleaseDate  string   `json:"releaseDate"`
	Runtime      int      `json:"runtime"`
	PosterPath   string   `json:"posterPath"`
	BackdropPath string   `json:"backdropPath"`
	Genres       []string `json:"genres"`
	VoteAverage  float64  `json:"voteAverage"`
	VoteCount    int      `json:"voteCount,omitempty"`
	ImdbID       string   `json:"imdbId"`
	Studios      []string `json:"studios,omitempty"`
	Cast         []string `json:"cast,omitempty"`
	CastPeople   []Person `json:"castPeople,omitempty"`
}

type TVShow struct {
	ID           int      `json:"id"`
	TVDBID       int      `json:"tvdbId,omitempty"`
	Title        string   `json:"title"`
	Original     string   `json:"originalTitle"`
	Overview     string   `json:"overview"`
	FirstAirDate string   `json:"firstAirDate"`
	PosterPath   string   `json:"posterPath"`
	BackdropPath string   `json:"backdropPath"`
	Genres       []string `json:"genres"`
	VoteAverage  float64  `json:"voteAverage"`
	VoteCount    int      `json:"voteCount,omitempty"`
	Studios      []string `json:"studios,omitempty"`
	Cast         []string `json:"cast,omitempty"`
	CastPeople   []Person `json:"castPeople,omitempty"`
}

type TVSeason struct {
	ID           int         `json:"id"`
	ShowID       int         `json:"showId"`
	ShowTitle    string      `json:"showTitle"`
	SeasonNumber int         `json:"seasonNumber"`
	Title        string      `json:"title"`
	Overview     string      `json:"overview"`
	AirDate      string      `json:"airDate"`
	PosterPath   string      `json:"posterPath"`
	VoteAverage  float64     `json:"voteAverage"`
	Episodes     []TVEpisode `json:"episodes,omitempty"`
}

type TVEpisode struct {
	ID            int      `json:"id"`
	SeasonNumber  int      `json:"seasonNumber"`
	EpisodeNumber int      `json:"episodeNumber"`
	Title         string   `json:"title"`
	Overview      string   `json:"overview"`
	AirDate       string   `json:"airDate"`
	StillPath     string   `json:"stillPath"`
	VoteAverage   float64  `json:"voteAverage"`
	VoteCount     int      `json:"voteCount,omitempty"`
	Runtime       int      `json:"runtime,omitempty"`
	Directors     []Person `json:"directors,omitempty"`
	Writers       []Person `json:"writers,omitempty"`
	Actors        []Person `json:"actors,omitempty"`
}

type Person struct {
	ID          int    `json:"id,omitempty"`
	Name        string `json:"name"`
	Role        string `json:"role,omitempty"`
	ProfilePath string `json:"profilePath,omitempty"`
	Guest       bool   `json:"guest,omitempty"`
}

func (c Client) Enabled() bool {
	return c.Key != ""
}

func (c Client) SearchMovie(query, year string) ([]SearchResult, error) {
	if c.Key == "" {
		return nil, fmt.Errorf("TMDB_API_KEY is not configured")
	}
	values := url.Values{}
	values.Set("api_key", c.Key)
	values.Set("query", query)
	values.Set("include_adult", "false")
	values.Set("language", c.language())
	if year != "" {
		values.Set("year", year)
	}
	var response struct {
		Results []struct {
			ID           int     `json:"id"`
			Title        string  `json:"title"`
			Original     string  `json:"original_title"`
			Overview     string  `json:"overview"`
			ReleaseDate  string  `json:"release_date"`
			PosterPath   string  `json:"poster_path"`
			BackdropPath string  `json:"backdrop_path"`
			VoteAverage  float64 `json:"vote_average"`
		} `json:"results"`
	}
	if err := c.get("/search/movie?"+values.Encode(), &response); err != nil {
		return nil, err
	}
	results := make([]SearchResult, 0, len(response.Results))
	for _, result := range response.Results {
		results = append(results, SearchResult{
			ID: result.ID, Title: result.Title, Original: result.Original,
			Overview: result.Overview, ReleaseDate: result.ReleaseDate,
			PosterPath: result.PosterPath, BackdropPath: result.BackdropPath,
			VoteAverage: result.VoteAverage, MediaType: "movie",
		})
	}
	return results, nil
}

func (c Client) SearchTV(query, year string) ([]SearchResult, error) {
	if c.Key == "" {
		return nil, fmt.Errorf("TMDB_API_KEY is not configured")
	}
	values := url.Values{}
	values.Set("api_key", c.Key)
	values.Set("query", query)
	values.Set("include_adult", "false")
	values.Set("language", c.language())
	if year != "" {
		values.Set("first_air_date_year", year)
	}
	var response struct {
		Results []struct {
			ID           int     `json:"id"`
			Name         string  `json:"name"`
			OriginalName string  `json:"original_name"`
			Overview     string  `json:"overview"`
			FirstAirDate string  `json:"first_air_date"`
			PosterPath   string  `json:"poster_path"`
			BackdropPath string  `json:"backdrop_path"`
			VoteAverage  float64 `json:"vote_average"`
		} `json:"results"`
	}
	if err := c.get("/search/tv?"+values.Encode(), &response); err != nil {
		return nil, err
	}
	results := make([]SearchResult, 0, len(response.Results))
	for _, result := range response.Results {
		results = append(results, SearchResult{
			ID: result.ID, Title: result.Name, Original: result.OriginalName,
			Overview: result.Overview, ReleaseDate: result.FirstAirDate,
			PosterPath: result.PosterPath, BackdropPath: result.BackdropPath,
			VoteAverage: result.VoteAverage, MediaType: "tvshow",
		})
	}
	return results, nil
}

func (c Client) Movie(id int) (Movie, error) {
	values := url.Values{}
	values.Set("api_key", c.Key)
	values.Set("language", c.language())
	values.Set("append_to_response", "external_ids,credits")
	var response struct {
		ID           int     `json:"id"`
		Title        string  `json:"title"`
		Original     string  `json:"original_title"`
		Overview     string  `json:"overview"`
		ReleaseDate  string  `json:"release_date"`
		Runtime      int     `json:"runtime"`
		PosterPath   string  `json:"poster_path"`
		BackdropPath string  `json:"backdrop_path"`
		VoteAverage  float64 `json:"vote_average"`
		VoteCount    int     `json:"vote_count"`
		Genres       []struct {
			Name string `json:"name"`
		} `json:"genres"`
		ProductionCompanies []struct {
			Name string `json:"name"`
		} `json:"production_companies"`
		Networks []struct {
			Name string `json:"name"`
		} `json:"networks"`
		ExternalIDs struct {
			ImdbID string `json:"imdb_id"`
		} `json:"external_ids"`
		Credits struct {
			Cast []struct {
				ID          int    `json:"id"`
				Name        string `json:"name"`
				Character   string `json:"character"`
				ProfilePath string `json:"profile_path"`
			} `json:"cast"`
		} `json:"credits"`
	}
	if err := c.get(fmt.Sprintf("/movie/%d?%s", id, values.Encode()), &response); err != nil {
		return Movie{}, err
	}
	movie := Movie{
		ID: response.ID, Title: response.Title, Original: response.Original,
		Overview: response.Overview, ReleaseDate: response.ReleaseDate, Runtime: response.Runtime,
		PosterPath: response.PosterPath, BackdropPath: response.BackdropPath,
		VoteAverage: response.VoteAverage, VoteCount: response.VoteCount, ImdbID: response.ExternalIDs.ImdbID,
	}
	for _, genre := range response.Genres {
		movie.Genres = append(movie.Genres, genre.Name)
	}
	for _, person := range response.Credits.Cast {
		if person.Name != "" && len(movie.Cast) < 12 {
			movie.Cast = append(movie.Cast, person.Name)
		}
		if person.Name != "" {
			movie.CastPeople = append(movie.CastPeople, Person{
				ID: person.ID, Name: person.Name, Role: person.Character, ProfilePath: person.ProfilePath,
			})
		}
	}
	for _, company := range response.ProductionCompanies {
		if company.Name != "" {
			movie.Studios = append(movie.Studios, company.Name)
		}
	}
	return movie, nil
}

func (c Client) TVShow(id int) (TVShow, error) {
	values := url.Values{}
	values.Set("api_key", c.Key)
	values.Set("language", c.language())
	values.Set("append_to_response", "external_ids,credits")
	var response struct {
		ID           int     `json:"id"`
		Name         string  `json:"name"`
		OriginalName string  `json:"original_name"`
		Overview     string  `json:"overview"`
		FirstAirDate string  `json:"first_air_date"`
		PosterPath   string  `json:"poster_path"`
		BackdropPath string  `json:"backdrop_path"`
		VoteAverage  float64 `json:"vote_average"`
		VoteCount    int     `json:"vote_count"`
		Genres       []struct {
			Name string `json:"name"`
		} `json:"genres"`
		ProductionCompanies []struct {
			Name string `json:"name"`
		} `json:"production_companies"`
		Networks []struct {
			Name string `json:"name"`
		} `json:"networks"`
		Credits struct {
			Cast []struct {
				ID          int    `json:"id"`
				Name        string `json:"name"`
				Character   string `json:"character"`
				ProfilePath string `json:"profile_path"`
			} `json:"cast"`
		} `json:"credits"`
		ExternalIDs struct {
			TVDBID int `json:"tvdb_id"`
		} `json:"external_ids"`
	}
	if err := c.get(fmt.Sprintf("/tv/%d?%s", id, values.Encode()), &response); err != nil {
		return TVShow{}, err
	}
	show := TVShow{
		ID: response.ID, TVDBID: response.ExternalIDs.TVDBID, Title: response.Name, Original: response.OriginalName,
		Overview: response.Overview, FirstAirDate: response.FirstAirDate,
		PosterPath: response.PosterPath, BackdropPath: response.BackdropPath,
		VoteAverage: response.VoteAverage, VoteCount: response.VoteCount,
	}
	for _, genre := range response.Genres {
		show.Genres = append(show.Genres, genre.Name)
	}
	for _, person := range response.Credits.Cast {
		if person.Name != "" && len(show.Cast) < 12 {
			show.Cast = append(show.Cast, person.Name)
		}
		if person.Name != "" {
			show.CastPeople = append(show.CastPeople, Person{
				ID: person.ID, Name: person.Name, Role: person.Character, ProfilePath: person.ProfilePath,
			})
		}
	}
	for _, company := range response.ProductionCompanies {
		if company.Name != "" {
			show.Studios = append(show.Studios, company.Name)
		}
	}
	for _, network := range response.Networks {
		if network.Name != "" {
			show.Studios = append(show.Studios, network.Name)
		}
	}
	return show, nil
}

func (c Client) TVSeason(showID int, seasonNumber int, showTitle string) (TVSeason, error) {
	values := url.Values{}
	values.Set("api_key", c.Key)
	values.Set("language", c.language())
	var response struct {
		ID           int     `json:"id"`
		Name         string  `json:"name"`
		Overview     string  `json:"overview"`
		AirDate      string  `json:"air_date"`
		PosterPath   string  `json:"poster_path"`
		SeasonNumber int     `json:"season_number"`
		VoteAverage  float64 `json:"vote_average"`
		Episodes     []struct {
			ID            int     `json:"id"`
			Name          string  `json:"name"`
			Overview      string  `json:"overview"`
			AirDate       string  `json:"air_date"`
			EpisodeNumber int     `json:"episode_number"`
			SeasonNumber  int     `json:"season_number"`
			StillPath     string  `json:"still_path"`
			VoteAverage   float64 `json:"vote_average"`
			VoteCount     int     `json:"vote_count"`
			Runtime       int     `json:"runtime"`
			Crew          []struct {
				ID          int    `json:"id"`
				Name        string `json:"name"`
				Job         string `json:"job"`
				Department  string `json:"department"`
				ProfilePath string `json:"profile_path"`
			} `json:"crew"`
			GuestStars []struct {
				ID          int    `json:"id"`
				Name        string `json:"name"`
				Character   string `json:"character"`
				ProfilePath string `json:"profile_path"`
			} `json:"guest_stars"`
		} `json:"episodes"`
	}
	if err := c.get(fmt.Sprintf("/tv/%d/season/%d?%s", showID, seasonNumber, values.Encode()), &response); err != nil {
		return TVSeason{}, err
	}
	season := TVSeason{
		ID: response.ID, ShowID: showID, ShowTitle: showTitle, SeasonNumber: response.SeasonNumber,
		Title: response.Name, Overview: response.Overview, AirDate: response.AirDate,
		PosterPath: response.PosterPath, VoteAverage: response.VoteAverage,
	}
	for _, episode := range response.Episodes {
		item := TVEpisode{
			ID: episode.ID, SeasonNumber: episode.SeasonNumber, EpisodeNumber: episode.EpisodeNumber,
			Title: episode.Name, Overview: episode.Overview, AirDate: episode.AirDate,
			StillPath: episode.StillPath, VoteAverage: episode.VoteAverage, VoteCount: episode.VoteCount,
			Runtime: episode.Runtime,
		}
		for _, person := range episode.Crew {
			member := Person{ID: person.ID, Name: person.Name, Role: person.Job, ProfilePath: person.ProfilePath}
			switch strings.ToLower(strings.TrimSpace(person.Job)) {
			case "director":
				item.Directors = append(item.Directors, member)
			case "writer", "screenplay", "story", "teleplay":
				item.Writers = append(item.Writers, member)
			default:
				if strings.EqualFold(strings.TrimSpace(person.Department), "Writing") {
					item.Writers = append(item.Writers, member)
				}
			}
		}
		for _, person := range episode.GuestStars {
			item.Actors = append(item.Actors, Person{
				ID: person.ID, Name: person.Name, Role: person.Character, ProfilePath: person.ProfilePath, Guest: true,
			})
		}
		season.Episodes = append(season.Episodes, item)
	}
	return season, nil
}

func (c Client) DownloadImage(path string) ([]byte, error) {
	return c.DownloadImageSized(path, "original")
}

func (c Client) DownloadImageSized(path string, size string) ([]byte, error) {
	if path == "" {
		return nil, fmt.Errorf("empty image path")
	}
	if !strings.HasPrefix(path, "/") || strings.Contains(path, "..") {
		return nil, fmt.Errorf("invalid image path")
	}
	size = strings.TrimSpace(size)
	if size == "" {
		size = "original"
	}
	req, err := http.NewRequest("GET", imageBaseURL+"/"+size+path, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.httpClient().Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("tmdb image status %d", resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
}

func (c Client) get(path string, out interface{}) error {
	resp, err := c.httpClient().Get(baseURL + path)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("tmdb status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

func (c Client) httpClient() *http.Client {
	if c.HTTP != nil {
		return c.HTTP
	}
	return http.DefaultClient
}

func (c Client) language() string {
	if c.Lang != "" {
		return c.Lang
	}
	return "zh-CN"
}
