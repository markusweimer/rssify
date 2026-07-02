# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

rssify is a lightweight Go server that generates RSS feeds from websites that don't provide them. It ships as a single static binary in a scratch container image with no runtime dependencies.

## Commands

```sh
make build          # CGO_ENABLED=0 build with version ldflags -> ./rssify
make test           # go test ./...
make test-verbose   # go test -v ./...
make test-cover     # go test with coverage summary
make vet            # go vet ./...
make fmt            # gofmt -w .
make lint           # vet + staticcheck (if installed)
make run            # build and run the server
make podman-build   # build the container image
make podman-run     # build and run the container
```

Run a single test: `go test ./internal/scraper/ -run TestExtractItems -v`

There is no separate lint config; `make lint` is just `go vet` plus `staticcheck` if present on PATH.

## Architecture

The whole app is wired together in `main.go`; the `internal/` packages are otherwise independent of each other.

- **`internal/scraper`** — Defines the `Scraper` interface (`ID`, `FeedTitle`, `FeedDescription`, `FeedLink`, `Scrape`). Each site is a plugin implementing this interface, instantiated and registered in the `scrapers` slice in `main.go`. `anthropic.go` holds shared HTML-parsing logic (targets `<ul>` elements with a `PublicationList`+`__list` class, extracts `<li>` children) used by both `anthropic_news.go` and `anthropic_research.go`, which only differ in URL/title/description. Adding a new feed source means creating a new file here implementing `Scraper` and registering it in `main.go` — see the README's "Adding a new scraper" section for the interface shape.
- **`internal/cache`** — Generic in-memory TTL cache (`Cache[T]`), keyed by scraper ID. `Get` returns `(value, exists, fresh)` so callers can distinguish "missing" from "stale."
- **`internal/feed`** — Builds RSS 2.0 XML (`feed.Build`) from a `Scraper` and its `[]FeedItem`.
- **`internal/config`** — Loads `config.yaml` (optional; falls back to `config.Defaults()` if missing/unreadable) and applies `RSS_*` environment variable overrides on top.

### Request flow (`main.go`)

1. `GET /feeds/{id}` looks up the scraper by ID and calls `getItems`.
2. `getItems` checks the cache first. If fresh, return immediately.
3. On a miss/stale entry, it uses `golang.org/x/sync/singleflight` (keyed by scraper ID) so concurrent requests for the same expired feed share a single upstream fetch instead of a thundering herd.
4. **Stale-on-error**: if the fresh scrape fails (or returns zero items) and a stale cached copy exists, that stale copy is served instead of an error.
5. Successful non-empty results are written back into the cache.
6. `feed.Build` renders the items to RSS 2.0 XML, which is returned with `Cache-Control: max-age=<cache TTL>`.

A single shared `*http.Client` (with connection pooling/keep-alive) is passed into every scraper — new scrapers should reuse this client rather than creating their own.

### Routes

| Route | Description |
|-------|-------------|
| `GET /` | Index page listing all available feeds |
| `GET /feeds/{id}` | RSS 2.0 feed for the given scraper |
| `GET /health` | Health check (returns `ok`) |

## Container image

The `Dockerfile` is a two-stage build: `golang:1.25-alpine` (must match the `go.mod` Go version) compiles a static binary, then it's copied into `scratch` along with CA certs and `config.yaml`, running as unprivileged UID/GID `65534`. CI (`.github/workflows/*.yml`) builds and pushes this image to `ghcr.io` on pushes to `main` and on `v*` tags.
