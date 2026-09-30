package main

import (
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	cfg, err := loadConfig()
	if err != nil {
		log.Fatal(err)
	}

	handler := newHandler(cfg)

	server := &http.Server{
		Addr:              cfg.ListenAddr,
		Handler:           loggingMiddleware(handler),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	log.Printf("instatus-discord listening on %s", cfg.ListenAddr)
	log.Printf("status page name: %s", cfg.StatusPageName)

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s from %s (%s)", r.Method, r.URL.Path, r.RemoteAddr, time.Since(start))
	})
}

func init() {
	log.SetFlags(log.Ldate | log.Ltime | log.LUTC | log.Lmicroseconds)
	_ = os.Stdout
}
