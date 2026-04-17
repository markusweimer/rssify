package scraper

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/net/html"
)

// anthropicItem is a raw parsed item from an Anthropic publication list.
type anthropicItem struct {
	href     string
	title    string
	category string
	date     string
}

// parseAnthropicPage fetches a URL and extracts items from the PublicationList.
// Both /news and /research use the same HTML structure.
func parseAnthropicPage(ctx context.Context, pageURL string, client *http.Client) ([]anthropicItem, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, pageURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching %s: %w", pageURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d from %s", resp.StatusCode, pageURL)
	}

	// Limit body to 2MB to prevent runaway reads
	body, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	return extractItems(string(body))
}

// extractItems parses HTML and finds publication list items.
// Targets <ul> elements whose class contains "PublicationList" and "__list",
// then extracts <li> children with the expected structure.
func extractItems(htmlContent string) ([]anthropicItem, error) {
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return nil, fmt.Errorf("parsing HTML: %w", err)
	}

	var items []anthropicItem
	seen := make(map[string]bool)

	var findList func(*html.Node)
	findList = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "ul" {
			cls := getAttr(n, "class")
			// Match only PublicationList <ul> elements
			if strings.Contains(cls, "PublicationList") && strings.Contains(cls, "__list") {
				for c := n.FirstChild; c != nil; c = c.NextSibling {
					if c.Type == html.ElementNode && c.Data == "li" {
						if item, ok := extractListItem(c); ok && !seen[item.href] {
							items = append(items, item)
							seen[item.href] = true
						}
					}
				}
				return
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			findList(c)
		}
	}
	findList(doc)

	return items, nil
}

// extractListItem tries to extract a publication item from an <li> element.
// Expected structure:
//
//	<li><a href="..."><div class="...meta"><time>...</time><span>category</span></div><span>title</span></a></li>
func extractListItem(li *html.Node) (anthropicItem, bool) {
	// Find the <a> child
	a := findChild(li, "a")
	if a == nil {
		return anthropicItem{}, false
	}

	href := getAttr(a, "href")
	if href == "" {
		return anthropicItem{}, false
	}

	item := anthropicItem{href: href}

	// Find <time> element for date and spans for category/title
	var findParts func(*html.Node)
	findParts = func(n *html.Node) {
		if n.Type == html.ElementNode {
			switch n.Data {
			case "time":
				// Prefer datetime attribute over visible text
				if dt := getAttr(n, "datetime"); dt != "" {
					item.date = dt
				} else {
					item.date = textContent(n)
				}
				return
			case "span":
				cls := getAttr(n, "class")
				text := strings.TrimSpace(textContent(n))
				if text == "" {
					break
				}
				if strings.Contains(cls, "title") {
					item.title = text
				} else if strings.Contains(cls, "subject") {
					item.category = text
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			findParts(c)
		}
	}
	findParts(a)

	if item.title == "" {
		return anthropicItem{}, false
	}

	return item, true
}

// resolveAnthropicItems converts raw parsed items into FeedItems with full URLs and parsed dates.
func resolveAnthropicItems(items []anthropicItem, baseURL string) []FeedItem {
	base, _ := url.Parse(baseURL)

	result := make([]FeedItem, 0, len(items))
	for _, raw := range items {
		fi := FeedItem{
			Title:    raw.title,
			Category: raw.category,
		}

		// Resolve URL against base
		if ref, err := url.Parse(raw.href); err == nil && base != nil {
			fi.Link = base.ResolveReference(ref).String()
		} else {
			fi.Link = raw.href
		}

		// Parse date — try ISO format first (from datetime attr), then display format
		if raw.date != "" {
			fi.PubDate = parseDate(raw.date)
		}

		result = append(result, fi)
	}
	return result
}

// parseDate tries several date formats.
func parseDate(s string) time.Time {
	formats := []string{
		time.RFC3339,
		"2006-01-02",
		"Jan 2, 2006",
		"January 2, 2006",
	}
	for _, f := range formats {
		if t, err := time.Parse(f, s); err == nil {
			return t
		}
	}
	return time.Time{}
}

// HTML helper functions

func findChild(n *html.Node, tag string) *html.Node {
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode && c.Data == tag {
			return c
		}
	}
	return nil
}

func getAttr(n *html.Node, key string) string {
	for _, a := range n.Attr {
		if a.Key == key {
			return a.Val
		}
	}
	return ""
}

func textContent(n *html.Node) string {
	var sb strings.Builder
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.TextNode {
			sb.WriteString(n.Data)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(n)
	return strings.TrimSpace(sb.String())
}
