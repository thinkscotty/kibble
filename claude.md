# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project: Kibble

Kibble is an AI-powered facts generator web application. Users define topics, and the app calls the Gemini Flash 2.5 API to generate interesting facts, which are cached in SQLite and displayed via a web UI. Facts are also served to external client devices (LED matrices, smart displays) via a JSON API.

Target hardware: Cloud VPS (Rocky Linux 10, ARM64 Ampere, 4 cores, 8GB RAM). Previously Raspberry Pi 3B+.

## Build & Run Commands

```bash
make build           # Build for current platform -> bin/kibble
make build-arm64     # Cross-compile for Raspberry Pi (Linux ARM64)
make build-arm       # Cross-compile for older Raspberry Pi (Linux ARM 32-bit)
make build-all       # Build for all supported platforms
make run             # Build and run locally
make test            # Run all tests with race detector
make lint            # Run go vet
make clean           # Remove build artifacts
```

Run directly: `go run ./cmd/kibble -config config.yaml`

Run a single test: `go test -v -run TestFunctionName ./internal/package/`

## Architecture & Structure

**Stack**: Go backend, SQLite (pure-Go via `modernc.org/sqlite`), Go `html/template` + HTMX frontend, YAML config (`gopkg.in/yaml.v3`). All static assets embedded via `go:embed` for single-binary deployment. Only 2 external dependencies, no CGO.

**Key packages under `internal/`**:
- `config` — Loads YAML config with defaults. Config file is optional; all settings have sensible defaults.
- `models` — Shared domain types (`Topic`, `Fact`, `Setting`, `Stats`, etc.). No logic, just structs.
- `database` — All SQLite operations. WAL mode, `MaxOpenConns(2)`. Migrations run on startup. Tables: topics, facts, settings (key-value), api_usage_log.
- `gemini` — HTTP client for Gemini REST API. `prompt.go` builds prompts and parses numbered-list responses. API key read from settings table.
- `similarity` — Trigram Jaccard similarity. Precomputes trigrams on fact insert, stores as JSON in `facts.trigrams` column.
- `scheduler` — Single-goroutine background loop (60s tick). Generates facts for topics due for refresh. Mutex prevents overlap with manual refreshes.
- `server` — HTTP server with Go 1.22+ method-based routing. Handlers split by domain: `handlers_{dashboard,topics,facts,settings,stats,api}.go`. Templates loaded from embedded FS at startup.

**Embedded assets** (`embed.go` at project root):
- `web/static/` — CSS, vendored HTMX JS
- `web/templates/` — `layouts/base.html`, `pages/*.html`, `partials/*.html`

**HTMX patterns**: CRUD returns HTML partials (not full pages). Inline editing via swap. Toast notifications via OOB swap. Search debounce at 300ms.

**External JSON API** at `/api/v1/`: `topics`, `facts?topic_id=X&limit=N`, `facts/random`

**Runtime settings** (theme, colors, API key, AI instructions) stored in `settings` DB table, not YAML config. YAML is for infrastructure tuning (port, DB path, log level).

## Important Constraints

- Target: Raspberry Pi 3B+ with <1GB RAM. Minimize memory usage.
- Pure-Go SQLite enables cross-compilation with `GOOS=linux GOARCH=arm64` from any host.
- Binary uses `-ldflags "-s -w"` to strip debug info for smaller size.
- No CGO required.
