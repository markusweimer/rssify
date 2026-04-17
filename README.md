# rssify

A lightweight Go server that generates RSS feeds from websites that don't provide them.

Ships as a single static binary (~8 MB container image) with no runtime dependencies. Designed to run on Linux hosts as a container with minimal resource usage.

## Included feeds

| Feed | Route | Source |
|------|-------|--------|
| Anthropic News | `/feeds/anthropic-news` | [anthropic.com/news](https://www.anthropic.com/news) |
| Anthropic Research | `/feeds/anthropic-research` | [anthropic.com/research](https://www.anthropic.com/research) |

## Quick start

### Run locally

```sh
go build -o rssify .
./rssify
```

### Run with Docker

```sh
docker build -t rssify .
docker run -p 8080:8080 rssify
```

Then add `http://localhost:8080/feeds/anthropic-news` to your RSS reader.

## Routes

| Route | Description |
|-------|-------------|
| `GET /` | Index page listing all available feeds |
| `GET /feeds/{id}` | RSS 2.0 feed for the given scraper |
| `GET /health` | Health check (returns `ok`) |

## Configuration

rssify reads from `config.yaml` (if present) and allows environment variable overrides:

```yaml
server:
  port: 8080
  read_timeout: 5s
  write_timeout: 10s

cache:
  ttl: 30m

scraper:
  user_agent: "rssify/1.0"
  request_timeout: 15s
```

| Environment variable | Overrides | Example |
|---------------------|-----------|---------|
| `RSS_PORT` | `server.port` | `9090` |
| `RSS_CACHE_TTL` | `cache.ttl` | `15m` |
| `RSS_USER_AGENT` | `scraper.user_agent` | `my-reader/1.0` |
| `RSS_REQUEST_TIMEOUT` | `scraper.request_timeout` | `30s` |

## Adding a new scraper

1. Create a file in `internal/scraper/` that implements the `Scraper` interface:

```go
type Scraper interface {
    ID() string                                          // URL-safe route ID
    FeedTitle() string                                   // human-readable title
    FeedDescription() string                             // short description
    FeedLink() string                                    // source website URL
    Scrape(ctx context.Context) ([]FeedItem, error)      // fetch and parse
}
```

2. Register it in `main.go` by adding to the `scrapers` slice.

## Architecture

- **Scraper plugin pattern** — each site implements a common interface
- **In-memory TTL cache** — avoids re-fetching on every request (default 30 min)
- **Singleflight** — concurrent requests for the same expired feed share one upstream fetch
- **Stale-on-error** — serves cached data if the upstream site is unreachable
- **Graceful shutdown** — handles `SIGINT`/`SIGTERM` cleanly

## Testing

```sh
go test ./...
```
