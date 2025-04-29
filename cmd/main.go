package main

import (
	config "cloudcamp/configs"
	"cloudcamp/internal/schedule"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"os/signal"
	"time"
)

func main() {

	cfg, err := config.Load("../config.json")
	if err != nil {
		log.Fatalf("No such file in path: %v", err)
	}

	backendURLs := make([]string, len(cfg.Backends))
	weights := make([]int, len(cfg.Backends))
	for i, b := range cfg.Backends {
		backendURLs[i] = b.URL
		weights[i] = b.Weight
	}

	manager, err := schedule.NewManager(backendURLs, weights)
	if err != nil {
		log.Fatalf("Failed to create backend manager: %v", err)
	}

	proxy := &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			backendURL, err := manager.GetNextBackend()
			if err != nil {
				log.Printf("No backends available: %v", err)
				return
			}

			log.Printf("Forwarding to: %s", backendURL.String())
			req.URL.Scheme = backendURL.Scheme
			req.URL.Host = backendURL.Host
			req.Header.Set("X-Forwarded-For", req.RemoteAddr)
		},
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			log.Printf("Proxy error: %v", err)
			http.Error(w, "Service unavailable", http.StatusServiceUnavailable)
		},
		Transport: &http.Transport{
			DisableKeepAlives: false,
		},
	}

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      proxy,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
	}

	log.Printf("Starting server on port %d", cfg.Server.Port)
	log.Printf("Read timeout: %d sec, Write timeout: %d sec",
		cfg.Server.ReadTimeout, cfg.Server.WriteTimeout)

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt)

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	<-done
	log.Println("Server is shutting down...")
}
