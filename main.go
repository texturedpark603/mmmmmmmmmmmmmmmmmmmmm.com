package main

import (
	"log"
	"net/http"

	"m/api/lastfm"
	"m/config"
	"m/middleware"
)

func main() {
	cfg := config.Load(".env")

	if !cfg.Valid() {
		log.Fatal("LASTFM_API_KEY is not set")
	}

	mux := http.NewServeMux()

	mux.Handle("/api/now-playing", lastfm.NewHandler(cfg))

	files := http.FileServer(http.Dir("./static"))
	mux.Handle("/", files)

	handler := middleware.Chain(mux,
		middleware.Logger,
		middleware.CORS,
	)

	log.Printf("listening on http://localhost:%s", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, handler))
}
