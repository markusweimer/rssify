package feed

import (
	"encoding/xml"
	"time"

	"github.com/mweimer/rssify/internal/scraper"
)

// RSS represents an RSS 2.0 document.
type RSS struct {
	XMLName xml.Name `xml:"rss"`
	Version string   `xml:"version,attr"`
	Channel Channel  `xml:"channel"`
}

// Channel is the RSS channel element.
type Channel struct {
	Title         string `xml:"title"`
	Link          string `xml:"link"`
	Description   string `xml:"description"`
	LastBuildDate string `xml:"lastBuildDate,omitempty"`
	TTL           int    `xml:"ttl,omitempty"`
	Items         []Item `xml:"item"`
}

// Item is an RSS item element.
type Item struct {
	Title    string `xml:"title"`
	Link     string `xml:"link"`
	Category string `xml:"category,omitempty"`
	PubDate  string `xml:"pubDate,omitempty"`
	GUID     GUID   `xml:"guid"`
	Desc     string `xml:"description,omitempty"`
}

// GUID is the globally unique identifier for an item.
type GUID struct {
	IsPermaLink bool   `xml:"isPermaLink,attr"`
	Value       string `xml:",chardata"`
}

// Build generates RSS 2.0 XML bytes from a scraper and its items.
func Build(s scraper.Scraper, items []scraper.FeedItem, cacheTTLMinutes int) ([]byte, error) {
	rssItems := make([]Item, 0, len(items))
	for _, fi := range items {
		item := Item{
			Title: fi.Title,
			Link:  fi.Link,
			GUID: GUID{
				IsPermaLink: true,
				Value:       fi.Link,
			},
		}
		if fi.Category != "" {
			item.Category = fi.Category
		}
		if !fi.PubDate.IsZero() {
			item.PubDate = fi.PubDate.Format(time.RFC1123Z)
		}
		if fi.Desc != "" {
			item.Desc = fi.Desc
		}
		rssItems = append(rssItems, item)
	}

	rss := RSS{
		Version: "2.0",
		Channel: Channel{
			Title:         s.FeedTitle(),
			Link:          s.FeedLink(),
			Description:   s.FeedDescription(),
			LastBuildDate: time.Now().UTC().Format(time.RFC1123Z),
			TTL:           cacheTTLMinutes,
			Items:         rssItems,
		},
	}

	output, err := xml.MarshalIndent(rss, "", "  ")
	if err != nil {
		return nil, err
	}

	return append([]byte(xml.Header), output...), nil
}
