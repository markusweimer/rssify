package scraper

import (
	"context"
	"net/http"
)

const (
	anthropicBaseURL = "https://www.anthropic.com"
	anthropicNewsURL = anthropicBaseURL + "/news"
)

// AnthropicNews scrapes the Anthropic news page.
type AnthropicNews struct {
	Client *http.Client
}

func (s *AnthropicNews) ID() string              { return "anthropic-news" }
func (s *AnthropicNews) FeedTitle() string        { return "Anthropic News" }
func (s *AnthropicNews) FeedDescription() string  { return "News and announcements from Anthropic" }
func (s *AnthropicNews) FeedLink() string         { return anthropicNewsURL }

func (s *AnthropicNews) Scrape(ctx context.Context) ([]FeedItem, error) {
	raw, err := parseAnthropicPage(ctx, anthropicNewsURL, s.Client)
	if err != nil {
		return nil, err
	}
	return resolveAnthropicItems(raw, anthropicBaseURL), nil
}
