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
const imageBaseURL = "https://image.tmdb.org/t/p/original"

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
	ImdbID       string   `json:"imdbId"`
	Cast         []string `json:"cast,omitempty"`
}

type TVShow struct {
	ID           int      `json:"id"`
	Title        string   `json:"title"`
	Original     string   `json:"originalTitle"`
	Overview     string   `json:"overview"`
	FirstAirDate string   `json:"firstAirDate"`
	PosterPath   string   `json:"posterPath"`
	BackdropPath string   `json:"backdropPath"`
	Genres       []string `json:"genres"`
	VoteAverage  float64  `json:"voteAverage"`
	Cast         []string `json:"cast,omitempty"`
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
		Genres       []struct {
			Name string `json:"name"`
		} `json:"genres"`
		ExternalIDs struct {
			ImdbID string `json:"imdb_id"`
		} `json:"external_ids"`
		Credits struct {
			Cast []struct {
				Name string `json:"name"`
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
		VoteAverage: response.VoteAverage, ImdbID: response.ExternalIDs.ImdbID,
	}
	for _, genre := range response.Genres {
		movie.Genres = append(movie.Genres, genre.Name)
	}
	for _, person := range response.Credits.Cast {
		if person.Name != "" && len(movie.Cast) < 12 {
			movie.Cast = append(movie.Cast, person.Name)
		}
	}
	return movie, nil
}

func (c Client) TVShow(id int) (TVShow, error) {
	values := url.Values{}
	values.Set("api_key", c.Key)
	values.Set("language", c.language())
	values.Set("append_to_response", "credits")
	var response struct {
		ID           int     `json:"id"`
		Name         string  `json:"name"`
		OriginalName string  `json:"original_name"`
		Overview     string  `json:"overview"`
		FirstAirDate string  `json:"first_air_date"`
		PosterPath   string  `json:"poster_path"`
		BackdropPath string  `json:"backdrop_path"`
		VoteAverage  float64 `json:"vote_average"`
		Genres       []struct {
			Name string `json:"name"`
		} `json:"genres"`
		Credits struct {
			Cast []struct {
				Name string `json:"name"`
			} `json:"cast"`
		} `json:"credits"`
	}
	if err := c.get(fmt.Sprintf("/tv/%d?%s", id, values.Encode()), &response); err != nil {
		return TVShow{}, err
	}
	show := TVShow{
		ID: response.ID, Title: response.Name, Original: response.OriginalName,
		Overview: response.Overview, FirstAirDate: response.FirstAirDate,
		PosterPath: response.PosterPath, BackdropPath: response.BackdropPath,
		VoteAverage: response.VoteAverage,
	}
	for _, genre := range response.Genres {
		show.Genres = append(show.Genres, genre.Name)
	}
	for _, person := range response.Credits.Cast {
		if person.Name != "" && len(show.Cast) < 12 {
			show.Cast = append(show.Cast, person.Name)
		}
	}
	return show, nil
}

func (c Client) DownloadImage(path string) ([]byte, error) {
	if path == "" {
		return nil, fmt.Errorf("empty image path")
	}
	req, err := http.NewRequest("GET", imageBaseURL+path, nil)
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
