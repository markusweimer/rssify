package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/sync/singleflight"

	"github.com/mweimer/rssify/internal/cache"
	"github.com/mweimer/rssify/internal/config"
	"github.com/mweimer/rssify/internal/feed"
	"github.com/mweimer/rssify/internal/scraper"
)

func main() {
	cfg, err := config.Load("config.yaml")
	if err != nil {
		log.Printf("Warning: could not load config.yaml, using defaults: %v", err)
		cfg = config.Defaults()
	}

	feedCache := cache.New[[]scraper.FeedItem](cfg.Cache.TTL)
	var sfGroup singleflight.Group

	// Shared HTTP client for all scrapers — enables connection reuse
	httpClient := &http.Client{
		Timeout: cfg.Scraper.RequestTimeout,
		Transport: &http.Transport{
			MaxIdleConns:        10,
			IdleConnTimeout:     90 * time.Second,
			DisableCompression:  false,
			DialContext: (&net.Dialer{
				Timeout:   10 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
		},
	}

	scrapers := []scraper.Scraper{
		&scraper.AnthropicNews{Client: httpClient},
		&scraper.AnthropicResearch{Client: httpClient},
	}

	// Build scraper lookup
	scraperMap := make(map[string]scraper.Scraper, len(scrapers))
	for _, s := range scrapers {
		scraperMap[s.ID()] = s
	}

	cacheTTLMinutes := int(cfg.Cache.TTL.Minutes())

	mux := http.NewServeMux()

	// Index page listing available feeds
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintf(w, "<html><head><title>rssify</title></head><body>")
		fmt.Fprintf(w, "<h1>rssify</h1><p>Available feeds:</p><ul>")
		for _, s := range scrapers {
			path := "/feeds/" + s.ID()
			fmt.Fprintf(w, `<li><a href="%s">%s</a> — %s</li>`, path, s.FeedTitle(), s.FeedDescription())
		}
		fmt.Fprintf(w, "</ul></body></html>")
	})

	// Feed handler
	mux.HandleFunc("GET /feeds/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		s, ok := scraperMap[id]
		if !ok {
			http.NotFound(w, r)
			return
		}

		items, err := getItems(r.Context(), feedCache, &sfGroup, s)
		if err != nil {
			log.Printf("Error scraping %s: %v", id, err)
			http.Error(w, "Failed to fetch feed", http.StatusBadGateway)
			return
		}

		xmlBytes, err := feed.Build(s, items, cacheTTLMinutes)
		if err != nil {
			log.Printf("Error building RSS for %s: %v", id, err)
			http.Error(w, "Failed to build feed", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/rss+xml; charset=utf-8")
		w.Header().Set("Cache-Control", fmt.Sprintf("public, max-age=%d", int(cfg.Cache.TTL.Seconds())))
		w.Write(xmlBytes)
	})

	// Health check
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("ok"))
	})

	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	// Graceful shutdown
	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("rssify listening on %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	<-done
	log.Println("Shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Shutdown error: %v", err)
	}
	log.Println("Stopped")
}

// getItems returns cached items or fetches fresh ones.
// Uses singleflight to coalesce concurrent refreshes for the same feed.
// Implements stale-on-error: serves stale data if a fresh fetch fails.
func getItems(ctx context.Context, c *cache.Cache[[]scraper.FeedItem], sf *singleflight.Group, s scraper.Scraper) ([]scraper.FeedItem, error) {
	items, exists, fresh := c.Get(s.ID())
	if exists && fresh {
		return items, nil
	}

	// Use singleflight to prevent thundering herd on cache expiry
	result, err, _ := sf.Do(s.ID(), func() (any, error) {
		// Double-check cache inside singleflight (another goroutine may have refreshed)
		if items, _, fresh := c.Get(s.ID()); fresh {
			return items, nil
		}

		newItems, err := s.Scrape(ctx)
		if err != nil {
			return nil, err
		}
		if len(newItems) > 0 {
			c.Set(s.ID(), newItems)
		}
		return newItems, nil
	})

	if err != nil {
		if exists {
			log.Printf("Serving stale cache for %s due to error: %v", s.ID(), err)
			return items, nil
		}
		return nil, err
	}

	newItems := result.([]scraper.FeedItem)
	if len(newItems) == 0 && exists {
		log.Printf("Scrape of %s returned 0 items, serving stale cache", s.ID())
		return items, nil
	}

	return newItems, nil
}
