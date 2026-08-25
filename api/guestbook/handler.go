package guestbook

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

const storeFile = "guestbook.json"

type Entry struct {
	Name    string `json:"name"`
	Message string `json:"message"`
	At      string `json:"at"` // RFC3339
}

var mu sync.Mutex

func load() ([]Entry, error) {
	data, err := os.ReadFile(storeFile)
	if os.IsNotExist(err) {
		return []Entry{}, nil
	}
	if err != nil {
		return nil, err
	}
	var entries []Entry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, err
	}
	return entries, nil
}

func save(entries []Entry) error {
	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(storeFile, data, 0644)
}

func Handler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		handleGet(w, r)
	case http.MethodPost:
		handlePost(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleGet(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	entries, err := load()
	mu.Unlock()
	if err != nil {
		http.Error(w, "could not load entries", http.StatusInternalServerError)
		return
	}
	// return newest first
	reversed := make([]Entry, len(entries))
	for i, e := range entries {
		reversed[len(entries)-1-i] = e
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	json.NewEncoder(w).Encode(reversed)
}

func handlePost(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name    string `json:"name"`
		Message string `json:"message"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	body.Name = strings.TrimSpace(body.Name)
	body.Message = strings.TrimSpace(body.Message)

	if body.Name == "" || body.Message == "" {
		http.Error(w, "name and message required", http.StatusBadRequest)
		return
	}
	if len(body.Name) > 40 {
		http.Error(w, "name too long (max 40)", http.StatusBadRequest)
		return
	}
	if len(body.Message) > 300 {
		http.Error(w, "message too long (max 300)", http.StatusBadRequest)
		return
	}

	entry := Entry{
		Name:    body.Name,
		Message: body.Message,
		At:      time.Now().UTC().Format(time.RFC3339),
	}

	mu.Lock()
	entries, err := load()
	if err == nil {
		entries = append(entries, entry)
		err = save(entries)
	}
	mu.Unlock()

	if err != nil {
		http.Error(w, "could not save entry", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(entry)
}
