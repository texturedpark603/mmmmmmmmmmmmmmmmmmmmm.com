package lastfm

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"os"
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

type response struct {
	RecentTracks recentTracks `xml:"recenttracks"`
}

type nowPlaying struct {
	Song      string `json:"song"`
	Artist    string `json:"artist"`
	CoverURL  string `json:"cover_url"`
	IsPlaying bool   `json:"is_playing"`
}

func Handler(w http.ResponseWriter, r *http.Request) {
	apiKey := os.Getenv("LASTFM_API_KEY")
	if apiKey == "" {
		http.Error(w, "missing api key", http.StatusInternalServerError)
		return
	}

	username := os.Getenv("LASTFM_USER")
	if username == "" {
		username = "qm"
	}

	endpoint := fmt.Sprintf(
		"https://ws.audioscrobbler.com/2.0/?method=user.getRecentTracks&user=%s&api_key=%s&format=xml&limit=1",
		username, apiKey,
	)

	resp, err := http.Get(endpoint) //nolint:gosec
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

	var parsed response
	if err := xml.Unmarshal(body, &parsed); err != nil {
		http.Error(w, "could not parse response", http.StatusInternalServerError)
		return
	}

	tracks := parsed.RecentTracks.Tracks
	if len(tracks) == 0 {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(nowPlaying{})
		return
	}

	latest := tracks[0]

	var cover string
	for _, img := range latest.Images {
		if img.Size == "extralarge" {
			cover = img.URL
			break
		}
	}

	out := nowPlaying{
		Song:      latest.Name,
		Artist:    latest.Artist.Name,
		CoverURL:  cover,
		IsPlaying: latest.NowPlaying == "true",
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache")
	json.NewEncoder(w).Encode(out)
}
