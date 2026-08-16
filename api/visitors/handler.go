package visitors

import (
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
)

type Handler struct {
	mu      sync.Mutex
	file    string
	visited map[string]bool
}

func NewHandler(file string) *Handler {
	return &Handler{
		file:    file,
		visited: make(map[string]bool),
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ip := r.Header.Get("CF-Connecting-IP")
	if ip == "" {
		ip = r.RemoteAddr
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	count := h.readCount()

	if !h.visited[ip] {
		h.visited[ip] = true
		count++
		h.writeCount(count)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]int{"visitors": count})
}

func (h *Handler) readCount() int {
	data, err := os.ReadFile(h.file)
	if err != nil {
		return 0
	}
	n, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		return 0
	}
	return n
}

func (h *Handler) writeCount(n int) {
	os.WriteFile(h.file, []byte(strconv.Itoa(n)), 0644) //nolint:errcheck
}
