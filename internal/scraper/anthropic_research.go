package scraper

import (
	"context"
	"net/http"
)

const anthropicResearchURL = anthropicBaseURL + "/research"

// AnthropicResearch scrapes the Anthropic research page.
type AnthropicResearch struct {
	Client *http.Client
}

func (s *AnthropicResearch) ID() string              { return "anthropic-research" }
func (s *AnthropicResearch) FeedTitle() string        { return "Anthropic Research" }
func (s *AnthropicResearch) FeedDescription() string  { return "Research publications from Anthropic" }
func (s *AnthropicResearch) FeedLink() string         { return anthropicResearchURL }

func (s *AnthropicResearch) Scrape(ctx context.Context) ([]FeedItem, error) {
	raw, err := parseAnthropicPage(ctx, anthropicResearchURL, s.Client)
	if err != nil {
		return nil, err
	}
	return resolveAnthropicItems(raw, anthropicBaseURL), nil
}
