package scraper

import (
	"context"
	"time"
)

// FeedItem represents a single item in an RSS feed.
type FeedItem struct {
	Title    string
	Link     string
	Category string
	PubDate  time.Time
	Desc     string
}

// Scraper defines the interface for a website scraper that produces RSS feed items.
// To add a new feed source, implement this interface and register it in main.go.
type Scraper interface {
	// ID returns a URL-safe identifier used in the feed route (e.g., "anthropic-news").
	ID() string
	// FeedTitle returns the human-readable title for the RSS feed.
	FeedTitle() string
	// FeedDescription returns a short description of the feed.
	FeedDescription() string
	// FeedLink returns the URL of the source website.
	FeedLink() string
	// Scrape fetches and parses the source page, returning feed items.
	Scrape(ctx context.Context) ([]FeedItem, error)
}
