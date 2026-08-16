package main

import (
	"log"
	"net/http"

	"m/api/lanyard"
	"m/api/visitors"
	"m/config"
	"m/middleware"
)

func main() {
	cfg := config.Load(".env")

	mux := http.NewServeMux()

	mux.HandleFunc("/api/presence", lanyard.Handler)
	mux.Handle("/api/visitors", visitors.NewHandler("visitors.count"))

	files := http.FileServer(http.Dir("./static"))
	mux.Handle("/", files)

	handler := middleware.Chain(mux,
		middleware.Logger,
		middleware.CORS,
		middleware.NoCache,
	)

	log.Printf("listening on http://localhost:%s", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, handler))
}
