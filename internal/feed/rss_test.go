package feed

import (
	"context"
	"encoding/xml"
	"strings"
	"testing"
	"time"

	"github.com/mweimer/rssify/internal/scraper"
)

// mockScraper implements scraper.Scraper for testing.
type mockScraper struct{}

func (m *mockScraper) ID() string                                            { return "test-feed" }
func (m *mockScraper) FeedTitle() string                                     { return "Test Feed" }
func (m *mockScraper) FeedDescription() string                               { return "A test feed" }
func (m *mockScraper) FeedLink() string                                      { return "https://example.com" }
func (m *mockScraper) Scrape(_ context.Context) ([]scraper.FeedItem, error)  { return nil, nil }

func TestBuildEmptyItems(t *testing.T) {
	s := &mockScraper{}
	data, err := Build(s, nil, 30)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	str := string(data)
	if !strings.HasPrefix(str, "<?xml") {
		t.Error("expected XML declaration")
	}
	if !strings.Contains(str, "<title>Test Feed</title>") {
		t.Error("expected feed title")
	}
	if !strings.Contains(str, `<rss version="2.0">`) {
		t.Error("expected RSS 2.0 version")
	}
}

func TestBuildWithItems(t *testing.T) {
	s := &mockScraper{}
	items := []scraper.FeedItem{
		{
			Title:    "First Post",
			Link:     "https://example.com/first",
			Category: "News",
			PubDate:  time.Date(2026, 4, 16, 0, 0, 0, 0, time.UTC),
			Desc:     "Description of first post",
		},
		{
			Title: "Second Post",
			Link:  "https://example.com/second",
		},
	}

	data, err := Build(s, items, 30)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify valid XML by parsing it back
	var rss RSS
	if err := xml.Unmarshal(data, &rss); err != nil {
		t.Fatalf("produced invalid XML: %v", err)
	}

	if rss.Version != "2.0" {
		t.Errorf("expected version 2.0, got %q", rss.Version)
	}
	if rss.Channel.Title != "Test Feed" {
		t.Errorf("expected 'Test Feed', got %q", rss.Channel.Title)
	}
	if rss.Channel.TTL != 30 {
		t.Errorf("expected TTL 30, got %d", rss.Channel.TTL)
	}
	if len(rss.Channel.Items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(rss.Channel.Items))
	}

	item := rss.Channel.Items[0]
	if item.Title != "First Post" {
		t.Errorf("expected 'First Post', got %q", item.Title)
	}
	if item.Category != "News" {
		t.Errorf("expected category 'News', got %q", item.Category)
	}
	if item.GUID.Value != "https://example.com/first" {
		t.Errorf("expected GUID matching link, got %q", item.GUID.Value)
	}
	if !item.GUID.IsPermaLink {
		t.Error("expected isPermaLink=true")
	}
	if item.Desc != "Description of first post" {
		t.Errorf("expected description, got %q", item.Desc)
	}

	// Second item should have no category or description
	item2 := rss.Channel.Items[1]
	if item2.Category != "" {
		t.Errorf("expected empty category, got %q", item2.Category)
	}
	if item2.PubDate != "" {
		t.Errorf("expected empty pubDate for zero time, got %q", item2.PubDate)
	}
}

func TestBuildXMLDeclaration(t *testing.T) {
	s := &mockScraper{}
	data, err := Build(s, nil, 10)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(string(data), `<?xml version="1.0" encoding="UTF-8"?>`) {
		t.Error("missing or incorrect XML declaration")
	}
}
