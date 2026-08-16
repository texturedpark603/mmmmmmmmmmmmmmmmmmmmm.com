package lanyard

import (
	"encoding/json"
	"io"
	"net/http"
)

const userID = "1442791669220249631"

type presence struct {
	Status          string  `json:"status"`
	StatusText      string  `json:"status_text"`
	Activity        string  `json:"activity"`
	SpotifySong     string  `json:"spotify_song"`
	SpotifyArtist   string  `json:"spotify_artist"`
	SpotifyAlbumArt string  `json:"spotify_album_art"`
	SpotifyEnd      int64   `json:"spotify_end"`
	SpotifyStart    int64   `json:"spotify_start"`
	IsListening     bool    `json:"is_listening"`
}

type lanyardResp struct {
	Data struct {
		DiscordStatus string `json:"discord_status"`
		Activities    []struct {
			Name    string `json:"name"`
			Details string `json:"details"`
			State   string `json:"state"`
			Type    int    `json:"type"`
		} `json:"activities"`
		ListeningToSpotify bool `json:"listening_to_spotify"`
		Spotify            *struct {
			Song        string `json:"song"`
			Artist      string `json:"artist"`
			AlbumArtURL string `json:"album_art_url"`
			Timestamps  struct {
				End   int64 `json:"end"`
				Start int64 `json:"start"`
			} `json:"timestamps"`
		} `json:"spotify"`
	} `json:"data"`
	Success bool `json:"success"`
}

func Handler(w http.ResponseWriter, r *http.Request) {
	resp, err := http.Get("https://api.lanyard.rest/v1/users/" + userID)
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

	for _, a := range d.Activities {
		if a.Type == 2 {
			continue
		}
		if a.Name != "" {
			out.Activity = a.Name
			if a.Details != "" {
				out.Activity += " — " + a.Details
			}
			break
		}
	}

	if d.ListeningToSpotify && d.Spotify != nil {
		out.SpotifySong = d.Spotify.Song
		out.SpotifyArtist = d.Spotify.Artist
		out.SpotifyAlbumArt = d.Spotify.AlbumArtURL
		out.SpotifyEnd = d.Spotify.Timestamps.End
		out.SpotifyStart = d.Spotify.Timestamps.Start
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	json.NewEncoder(w).Encode(out)
}
