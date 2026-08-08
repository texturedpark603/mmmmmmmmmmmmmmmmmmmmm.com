package lastfm

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"time"

	"m/cache"
	"m/config"
)

type track struct {
	Name       string  `xml:"name"`
	Artist     artist  `xml:"artist"`
	Images     []image `xml:"image"`
	NowPlaying string  `xml:"nowplaying,attr"`
}

type artist struct {
	Name string `xml:",chardata"`
}

type image struct {
	Size string `xml:"size,attr"`
	URL  string `xml:",chardata"`
}

type recentTracks struct {
	Tracks []track `xml:"track"`
}

type lfmResponse struct {
	RecentTracks recentTracks `xml:"recenttracks"`
}

type nowPlaying struct {
	Song      string `json:"song"`
	Artist    string `json:"artist"`
	CoverURL  string `json:"cover_url"`
	IsPlaying bool   `json:"is_playing"`
}

const cacheKey = "now-playing"

type Handler struct {
	cfg   *config.Config
	store *cache.Cache
}

func NewHandler(cfg *config.Config) *Handler {
	return &Handler{
		cfg:   cfg,
		store: cache.New(30 * time.Second),
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if cached, ok := h.store.Get(cacheKey); ok {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(cached)
		return
	}

	result, err := h.fetch()
	if err != nil {
		http.Error(w, "could not fetch track", http.StatusBadGateway)
		return
	}

	h.store.Set(cacheKey, result)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (h *Handler) fetch() (*nowPlaying, error) {
	endpoint := fmt.Sprintf(
		"https://ws.audioscrobbler.com/2.0/?method=user.getRecentTracks&user=%s&api_key=%s&format=xml&limit=1",
		h.cfg.LastFMUser, h.cfg.LastFMKey,
	)

	resp, err := http.Get(endpoint) //nolint:gosec
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var parsed lfmResponse
	if err := xml.Unmarshal(body, &parsed); err != nil {
		return nil, err
	}

	tracks := parsed.RecentTracks.Tracks
	if len(tracks) == 0 {
		return &nowPlaying{}, nil
	}

	latest := tracks[0]

	var cover string
	for _, img := range latest.Images {
		if img.Size == "extralarge" {
			cover = img.URL
			break
		}
	}

	return &nowPlaying{
		Song:      latest.Name,
		Artist:    latest.Artist.Name,
		CoverURL:  cover,
		IsPlaying: latest.NowPlaying == "true",
	}, nil
}
