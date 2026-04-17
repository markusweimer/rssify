package scraper

import (
	"testing"
	"time"
)

func TestExtractItemsFromNewsHTML(t *testing.T) {
	html := `<html><body>
	<ul class="PublicationList-module-scss-module__KxYrHG__list">
		<li>
			<a href="/news/test-article" class="PublicationList-module-scss-module__KxYrHG__listItem">
				<div class="PublicationList-module-scss-module__KxYrHG__meta">
					<time class="PublicationList-module-scss-module__KxYrHG__date body-3">Apr 16, 2026</time>
					<span class="PublicationList-module-scss-module__KxYrHG__subject body-3">Product</span>
				</div>
				<span class="PublicationList-module-scss-module__KxYrHG__title body-3">Test Article Title</span>
			</a>
		</li>
		<li>
			<a href="/news/second-article" class="PublicationList-module-scss-module__KxYrHG__listItem">
				<div class="PublicationList-module-scss-module__KxYrHG__meta">
					<time class="PublicationList-module-scss-module__KxYrHG__date body-3">Mar 10, 2026</time>
					<span class="PublicationList-module-scss-module__KxYrHG__subject body-3">Announcements</span>
				</div>
				<span class="PublicationList-module-scss-module__KxYrHG__title body-3">Second Article</span>
			</a>
		</li>
	</ul>
	</body></html>`

	items, err := extractItems(html)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}

	if items[0].href != "/news/test-article" {
		t.Errorf("expected /news/test-article, got %q", items[0].href)
	}
	if items[0].title != "Test Article Title" {
		t.Errorf("expected 'Test Article Title', got %q", items[0].title)
	}
	if items[0].category != "Product" {
		t.Errorf("expected 'Product', got %q", items[0].category)
	}
	if items[0].date != "Apr 16, 2026" {
		t.Errorf("expected 'Apr 16, 2026', got %q", items[0].date)
	}

	if items[1].title != "Second Article" {
		t.Errorf("expected 'Second Article', got %q", items[1].title)
	}
}

func TestExtractItemsIgnoresNonPublicationList(t *testing.T) {
	// A <ul> without "PublicationList" in the class should be ignored
	html := `<html><body>
	<ul class="SomeOther__list">
		<li>
			<a href="/nav/link">
				<span class="title">Nav Item</span>
			</a>
		</li>
	</ul>
	</body></html>`

	items, err := extractItems(html)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("expected 0 items from non-publication list, got %d", len(items))
	}
}

func TestExtractItemsDeduplicates(t *testing.T) {
	html := `<html><body>
	<ul class="PublicationList-module-scss-module__abc__list">
		<li><a href="/news/dup"><span class="title">Article</span></a></li>
		<li><a href="/news/dup"><span class="title">Article</span></a></li>
	</ul>
	</body></html>`

	items, err := extractItems(html)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 1 {
		t.Errorf("expected 1 deduplicated item, got %d", len(items))
	}
}

func TestExtractItemsSkipsNoTitle(t *testing.T) {
	html := `<html><body>
	<ul class="PublicationList-module__list">
		<li><a href="/news/no-title"><time>Apr 1, 2026</time></a></li>
	</ul>
	</body></html>`

	items, err := extractItems(html)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("expected 0 items (no title), got %d", len(items))
	}
}

func TestExtractItemsSkipsNoHref(t *testing.T) {
	html := `<html><body>
	<ul class="PublicationList-module__list">
		<li><a><span class="title">Orphan</span></a></li>
	</ul>
	</body></html>`

	items, err := extractItems(html)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("expected 0 items (no href), got %d", len(items))
	}
}

func TestExtractItemPrefersDatetimeAttr(t *testing.T) {
	html := `<html><body>
	<ul class="PublicationList-module__list">
		<li>
			<a href="/article">
				<time datetime="2026-04-16">Apr 16, 2026</time>
				<span class="title">Article</span>
			</a>
		</li>
	</ul>
	</body></html>`

	items, err := extractItems(html)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].date != "2026-04-16" {
		t.Errorf("expected datetime attr '2026-04-16', got %q", items[0].date)
	}
}

func TestResolveAnthropicItems(t *testing.T) {
	raw := []anthropicItem{
		{href: "/news/test", title: "Test", category: "News", date: "Apr 16, 2026"},
		{href: "/research/paper", title: "Paper", date: "2026-03-10"},
		{href: "https://external.com/page", title: "External"},
	}

	items := resolveAnthropicItems(raw, "https://www.anthropic.com")

	if len(items) != 3 {
		t.Fatalf("expected 3 items, got %d", len(items))
	}

	// Relative URL resolved
	if items[0].Link != "https://www.anthropic.com/news/test" {
		t.Errorf("expected full URL, got %q", items[0].Link)
	}
	if items[0].Category != "News" {
		t.Errorf("expected 'News', got %q", items[0].Category)
	}
	expected := time.Date(2026, 4, 16, 0, 0, 0, 0, time.UTC)
	if !items[0].PubDate.Equal(expected) {
		t.Errorf("expected %v, got %v", expected, items[0].PubDate)
	}

	// ISO date format
	expected2 := time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC)
	if !items[1].PubDate.Equal(expected2) {
		t.Errorf("expected %v, got %v", expected2, items[1].PubDate)
	}

	// Absolute URL preserved
	if items[2].Link != "https://external.com/page" {
		t.Errorf("expected external URL preserved, got %q", items[2].Link)
	}
	if !items[2].PubDate.IsZero() {
		t.Errorf("expected zero time for missing date, got %v", items[2].PubDate)
	}
}

func TestParseDateFormats(t *testing.T) {
	tests := []struct {
		input    string
		expected time.Time
	}{
		{"Apr 16, 2026", time.Date(2026, 4, 16, 0, 0, 0, 0, time.UTC)},
		{"2026-04-16", time.Date(2026, 4, 16, 0, 0, 0, 0, time.UTC)},
		{"January 1, 2025", time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)},
		{"2026-04-16T12:00:00Z", time.Date(2026, 4, 16, 12, 0, 0, 0, time.UTC)},
		{"garbage", time.Time{}},
		{"", time.Time{}},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := parseDate(tt.input)
			if !got.Equal(tt.expected) {
				t.Errorf("parseDate(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestTextContent(t *testing.T) {
	// Simple test via extractItems which uses textContent internally
	html := `<html><body>
	<ul class="PublicationList-module__list">
		<li>
			<a href="/test">
				<span class="title">  Spaced Title  </span>
			</a>
		</li>
	</ul>
	</body></html>`

	items, err := extractItems(html)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].title != "Spaced Title" {
		t.Errorf("expected trimmed title, got %q", items[0].title)
	}
}

func TestAnthropicNewsInterface(t *testing.T) {
	s := &AnthropicNews{}
	if s.ID() != "anthropic-news" {
		t.Errorf("unexpected ID: %s", s.ID())
	}
	if s.FeedLink() != "https://www.anthropic.com/news" {
		t.Errorf("unexpected link: %s", s.FeedLink())
	}
}

func TestAnthropicResearchInterface(t *testing.T) {
	s := &AnthropicResearch{}
	if s.ID() != "anthropic-research" {
		t.Errorf("unexpected ID: %s", s.ID())
	}
	if s.FeedLink() != "https://www.anthropic.com/research" {
		t.Errorf("unexpected link: %s", s.FeedLink())
	}
}
