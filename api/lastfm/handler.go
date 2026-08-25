package lastfm

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
)

const apiBase = "https://ws.audioscrobbler.com/2.0/"

type Track struct {
	Name      string `json:"name"`
	Artist    string `json:"artist"`
	Album     string `json:"album"`
	URL       string `json:"url"`
	Art       string `json:"art"`       // album art (extralarge)
	NowPlaying bool  `json:"now_playing"`
	PlayedAt  string `json:"played_at"` // UTC string, empty if now playing
}

type lastfmResp struct {
	RecentTracks struct {
		Track []struct {
			Name   string `json:"name"`
			URL    string `json:"url"`
			Artist struct {
				Text string `json:"#text"`
			} `json:"artist"`
			Album struct {
				Text string `json:"#text"`
			} `json:"album"`
			Image []struct {
				Text string `json:"#text"`
				Size string `json:"size"`
			} `json:"image"`
			Attr struct {
				NowPlaying string `json:"nowplaying"`
			} `json:"@attr"`
			Date struct {
				UTS  string `json:"uts"`
				Text string `json:"#text"`
			} `json:"date"`
		} `json:"track"`
	} `json:"recenttracks"`
}

var client = &http.Client{}

func Handler(w http.ResponseWriter, r *http.Request) {
	apiKey := os.Getenv("LASTFM_API_KEY")
	user := os.Getenv("LASTFM_USER")
	if apiKey == "" || user == "" {
		http.Error(w, "lastfm not configured", http.StatusServiceUnavailable)
		return
	}

	limit := r.URL.Query().Get("limit")
	if limit == "" {
		limit = "20"
	}

	q := url.Values{}
	q.Set("method", "user.getrecenttracks")
	q.Set("user", user)
	q.Set("api_key", apiKey)
	q.Set("format", "json")
	q.Set("limit", limit)
	q.Set("extended", "0")

	req, _ := http.NewRequest("GET", fmt.Sprintf("%s?%s", apiBase, q.Encode()), nil)
	req.Header.Set("Cache-Control", "no-cache")

	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "could not reach last.fm", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "could not read response", http.StatusInternalServerError)
		return
	}

	var raw lastfmResp
	if err := json.Unmarshal(body, &raw); err != nil {
		http.Error(w, "could not parse response", http.StatusInternalServerError)
		return
	}

	tracks := make([]Track, 0, len(raw.RecentTracks.Track))
	for _, t := range raw.RecentTracks.Track {
		art := ""
		for _, img := range t.Image {
			if img.Size == "extralarge" && img.Text != "" {
				art = img.Text
			}
		}
		track := Track{
			Name:       t.Name,
			Artist:     t.Artist.Text,
			Album:      t.Album.Text,
			URL:        t.URL,
			Art:        art,
			NowPlaying: t.Attr.NowPlaying == "true",
			PlayedAt:   t.Date.Text,
		}
		tracks = append(tracks, track)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	json.NewEncoder(w).Encode(tracks)
}
