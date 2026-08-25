package main

import (
	"log"
	"net/http"
	"strings"

	"m/api/guestbook"
	"m/api/lanyard"
	"m/api/lastfm"
	"m/api/status"
	"m/api/visitors"
	"m/config"
	"m/middleware"
)

func main() {
	cfg := config.Load(".env")

	mux := http.NewServeMux()

	// APIs
	mux.HandleFunc("/api/presence", lanyard.Handler)
	mux.Handle("/api/visitors", visitors.NewHandler("visitors.count"))
	mux.HandleFunc("/api/lastfm", lastfm.Handler)
	mux.HandleFunc("/api/guestbook", guestbook.Handler)
	mux.HandleFunc("/api/status", status.Handler)

	// Static assets (css, images, etc.)
	files := http.FileServer(http.Dir("./static"))
	mux.Handle("/css/", files)
	mux.Handle("/static/", files)

	// SPA: serve index.html for all page routes
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		// let through anything that looks like a real file (has an extension)
		if strings.Contains(path, ".") {
			files.ServeHTTP(w, r)
			return
		}
		http.ServeFile(w, r, "./static/index.html")
	})

	handler := middleware.Chain(mux,
		middleware.Logger,
		middleware.CORS,
		middleware.NoCache,
	)

	log.Printf("listening on http://localhost:%s", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, handler))
}
