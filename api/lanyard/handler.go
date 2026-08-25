package lanyard

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
)

const userID = "1270520220737339544"

var client = &http.Client{}

type presence struct {
	Status          string `json:"status"`
	StatusText      string `json:"status_text"`
	Activity        string `json:"activity"`
	SpotifySong     string `json:"spotify_song"`
	SpotifyArtist   string `json:"spotify_artist"`
	SpotifyAlbumArt string `json:"spotify_album_art"`
	SpotifyEnd      int64  `json:"spotify_end"`
	SpotifyStart    int64  `json:"spotify_start"`
	IsListening     bool   `json:"is_listening"`
	NowPlayingURL   string `json:"now_playing_url"`   // link for the song title
	ArtistURL       string `json:"artist_url"`        // link for the artist
	NowPlayingLabel string `json:"now_playing_label"` // e.g. "soundcloud" or "spotify"
}

type lanyardResp struct {
	Data struct {
		DiscordStatus string `json:"discord_status"`
		Activities    []struct {
			Name    string `json:"name"`
			Details string `json:"details"`
			State   string `json:"state"`
			Type    int    `json:"type"`
			Assets  *struct {
				LargeImage string `json:"large_image"`
			} `json:"assets"`
			DetailsURL string `json:"details_url"`
			StateURL   string `json:"state_url"`
			CreatedAt  int64  `json:"created_at"`
			Timestamps *struct {
				End   int64 `json:"end"`
				Start int64 `json:"start"`
			} `json:"timestamps"`
		} `json:"activities"`
		ListeningToSpotify bool `json:"listening_to_spotify"`
		Spotify            *struct {
			Song        string `json:"song"`
			Artist      string `json:"artist"`
			AlbumArtURL string `json:"album_art_url"`
			TrackID     string `json:"track_id"`
			Timestamps  struct {
				End   int64 `json:"end"`
				Start int64 `json:"start"`
			} `json:"timestamps"`
		} `json:"spotify"`
	} `json:"data"`
	Success bool `json:"success"`
}

// resolveImage converts Lanyard's mp:external/... proxy URLs into real HTTPS URLs.
// e.g. "mp:external/<hash>/https/i1.sndcdn.com/..." → "https://i1.sndcdn.com/..."
func resolveImage(raw string) string {
	if strings.HasPrefix(raw, "mp:external/") {
		withoutPrefix := strings.TrimPrefix(raw, "mp:external/")
		idx := strings.Index(withoutPrefix, "/")
		if idx == -1 {
			return ""
		}
		urlPart := withoutPrefix[idx+1:]
		slashIdx := strings.Index(urlPart, "/")
		if slashIdx == -1 {
			return ""
		}
		scheme := urlPart[:slashIdx]
		rest := urlPart[slashIdx+1:]
		return scheme + "://" + rest
	}
	return raw
}

func Handler(w http.ResponseWriter, r *http.Request) {
	req, _ := http.NewRequest("GET", "https://api.lanyard.rest/v1/users/"+userID, nil)
	req.Header.Set("Cache-Control", "no-cache, no-store")
	req.Header.Set("Pragma", "no-cache")

	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "could not reach lanyard", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "could not read response", http.StatusInternalServerError)
		return
	}

	var raw lanyardResp
	if err := json.Unmarshal(body, &raw); err != nil {
		http.Error(w, "could not parse response", http.StatusInternalServerError)
		return
	}

	d := raw.Data
	out := presence{
		Status:      d.DiscordStatus,
		IsListening: d.ListeningToSpotify,
	}

	statusLabels := map[string]string{
		"online":    "online",
		"idle":      "idle",
		"dnd":       "do not disturb",
		"offline":   "offline",
		"invisible": "offline",
	}
	if label, ok := statusLabels[d.DiscordStatus]; ok {
		out.StatusText = label
	}

	// Non-listening activities (games, custom status, etc.)
	for _, a := range d.Activities {
		if a.Type == 2 {
			continue
		}
		if out.Activity == "" && a.Name != "" {
			out.Activity = a.Name
			if a.Details != "" {
				out.Activity += " — " + a.Details
			}
		}
	}

	// Spotify takes priority; fall back to any type-2 external activity
	if d.ListeningToSpotify && d.Spotify != nil {
		s := d.Spotify
		out.SpotifySong = s.Song
		out.SpotifyArtist = s.Artist
		out.SpotifyAlbumArt = s.AlbumArtURL
		out.SpotifyEnd = s.Timestamps.End
		out.SpotifyStart = s.Timestamps.Start
		out.NowPlayingLabel = "spotify"
		if s.TrackID != "" {
			out.NowPlayingURL = "https://open.spotify.com/track/" + s.TrackID
		}
	} else {
		// Pick the most recently created external listening activity (highest created_at)
		best := -1
		for i, a := range d.Activities {
			if a.Type != 2 || a.Details == "" {
				continue
			}
			if best == -1 || a.CreatedAt > d.Activities[best].CreatedAt {
				best = i
			}
		}
		if best != -1 {
			a := d.Activities[best]
			out.IsListening = true
			out.SpotifySong = a.Details
			out.SpotifyArtist = a.State
			out.NowPlayingURL = a.DetailsURL
			out.ArtistURL = a.StateURL
			out.NowPlayingLabel = strings.ToLower(a.Name)
			if a.Assets != nil && a.Assets.LargeImage != "" {
				out.SpotifyAlbumArt = resolveImage(a.Assets.LargeImage)
			}
			if a.Timestamps != nil {
				out.SpotifyEnd = a.Timestamps.End
				out.SpotifyStart = a.Timestamps.Start
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	json.NewEncoder(w).Encode(out)
}
