# Google Messages MCP

Local-first message database with a built-in MCP server, oriented for Windows
(cross-compiles cleanly for Linux/macOS too). Ingests SMS/RCS (Google Messages),
and supports live WhatsApp and Signal plus Google Chat / WhatsApp / Signal
Desktop import.

The binary is `gmessages`; all environment variables use the `GMESSAGES_` prefix.

## Architecture

```
├── cmd/              Go CLI commands (pair, serve, send, read, status, import)
├── internal/
│   ├── app/          Bootstrap, data dir, backfill
│   ├── client/       libgm Google Messages protocol
│   ├── db/           SQLite store (conversations, messages, contacts, unified_contacts, drafts)
│   ├── importer/     Import adapters (gchat, whatsapp, signal_desktop)
│   ├── signallive/   Live Signal bridge (signal-cli linked device)
│   ├── whatsapplive/ Live WhatsApp bridge (whatsmeow companion device)
│   ├── story/        Stats computation + narrative story generation
│   ├── tools/        MCP tools
│   ├── viz/          Relationship visualization renderer (self-contained HTML)
│   └── web/          HTTP API + embedded React UI
├── scripts/          Linux cookie-refresh / watchdog helpers
├── Dockerfile        Headless MCP server image
└── docker-compose.yml
```

## Supporting a live install (READ FIRST for support/debug tasks)

If you are debugging a real install — sends failing, re-pairing, reading actual
messages — read **[docs/agent-runbook.md](docs/agent-runbook.md)** before
touching anything. The traps that cost the most:

- **Read live messages via the HTTP API** (`/api/conversations/<id>/messages`,
  `/api/search`, `/api/status`) — a running `serve` holds the WAL'd DB, so a
  direct `sqlite3` reader hits "unable to open database file (14)".
- **Re-pairing Google Messages:** QR pairing may not be offered; use Google
  Account pairing via the cookie method (`gmessages pair --google`); clear
  `session.json` from the data dir to reach the pairing screen; don't
  over-reconnect (it throttles the account). Full recipe in the runbook.

## Local CLI (read-only, no transports)

These commands open the store directly and start no live transports, so they
work in a one-shot terminal session without pairing:

```bash
gmessages read "<query>" [--limit N] [--phone NUMBER] [--since YYYY-MM-DD] [--until YYYY-MM-DD] [--json]
gmessages search ...                                            # alias for read
gmessages status [--json]                                       # per-platform counts + sync freshness
```

`status` is the fast way to check coverage before trusting a search: it lists
each platform's message count and latest sent/received timestamps, and flags any
platform whose latest message trails the newest overall by ≥3 days ("Nd behind").
A stale row means the daemon isn't syncing that platform — searches over that
window will miss messages. `read` resolves each hit's sender (name → number →
conversation id) so results are legible without a second lookup, and accepts
`--since`/`--until` (YYYY-MM-DD, local time; `--until` is inclusive to end of
day) to scope a search to a date window. Date filtering lives in the store via
`SearchFilter`/`SearchMessagesFiltered`; the legacy `SearchMessages(query,
phone, limit)` wrapper is preserved for the MCP tool and HTTP API.

## Multi-platform import

```bash
gmessages import gchat /path/to/Takeout/Google\ Chat/Groups/ --email you@gmail.com
gmessages import gchat-conversation /path/to/messages.json --email you@gmail.com
gmessages import whatsapp /path/to/chat.txt --name "Your Name"
gmessages import signal /path/to/Signal --name "Your Name"
```

### MCP tools

Tools registered in `internal/tools/tools.go` (`Register` is the authoritative list):
- `get_messages`, `get_conversation`, `search_messages` — cross-platform by default
- `list_conversations` — optional `source_platform` filter (sms, gchat, whatsapp, signal)
- `get_person_messages` — all messages with a person across all platforms
- `get_person_messages_range` — date-filtered version of get_person_messages (for deep-diving into specific periods)
- `import_messages` — import from any supported source (gchat, gchat_conversation, whatsapp, signal)
- `conversation_stats` — volume, heatmap, phrases, response times, gaps (single conversation)
- `generate_story` — narrative chapters with optional Claude API enhancement (single conversation)
- `person_stats` — cross-platform stats for all 1:1 messages with a person (merges + deduplicates)
- `generate_person_story` — cross-platform narrative story for a person (merges + deduplicates)
- `generate_viz` — self-contained HTML visualization combining data dashboards + narrative (see below)
- `render_story` — render a pre-built Story JSON into HTML viz; supports `photo_paths` (curated list) or `photos_dir`
- `send_message`, `send_to_conversation`, `send_media_to_conversation`, `react_to_message`
- `draft_message`, `download_media`, `list_contacts`, `get_status`

### HTTP API

- `GET /api/stats/{conversation_id}` — conversation statistics JSON
- `GET /api/story/{conversation_id}?style=intimate&api_key=...` — generated story JSON
- `GET /api/conversations?limit=50` — list all conversations (all platforms)
- `GET /api/search?q=...` — search across all platforms

### Schema

Messages and conversations have `source_platform` (sms/gchat/whatsapp/signal/imessage/telegram) and messages have `source_id` for dedup. Unified contacts table maps people across platforms. (`imessage` remains a valid platform value for previously-imported data, though the iMessage importer is not part of this fork.)

## Testing

```bash
go test ./cmd/ -v             # Unit + integration tests
go test ./... -v              # All tests
GOOS=windows go build ./...   # Verify the Windows build compiles
```

## Relationship visualization (`generate_viz`)

Generates a self-contained HTML file combining data dashboards with narrative chapters. Output is viewable locally.

**Sections**: password gate, hero, timeline nav, narrative chapters (early/middle/late), monthly volume chart (Chart.js), sender split donut, response times, hour-of-week heatmap, phrase cloud (colored by sender ratio), longest gap callout, interspersed photo breaks (chronologically aligned), interludes, closing.

**Key parameters**: `name` (person to search), `output_path` (relative to `GMESSAGES_EXPORT_DIR`, default `~/Documents/GoogleMessagesMCP`, unless `GMESSAGES_ALLOW_ANY_EXPORT_PATH=1` is set), `timezone` (default ET), `password`, `api_key` (for Claude-generated narrative), colors (`primary_color`, `secondary_color`, etc.).

**Architecture**:
- `internal/viz/config.go` — `VizConfig` struct, section ordering, color theming
- `internal/viz/render.go` — `RenderHTML()` orchestrator, Chart.js data building
- `internal/viz/template.go` — Go html/template with all CSS/JS inline (except CDN fonts + Chart.js)
- `internal/viz/photos.go` — `Photo` struct, `EncodePhotosFromDir/Paths()`, date parsing from filenames, chronological sorting
- `internal/tools/viz.go` — MCP tool handler

**Stats engine extensions** (`internal/story/stats.go`):
- `PhraseCount.BySender` — per-sender phrase counts for colored word cloud
- `ComputeStats(messages, tz)` — timezone parameter for TZ-shifted heatmap

## Agentic story generation (`/generate-story`)

Claude Code slash command that produces fact-grounded relationship visualizations. Instead of a single-pass API call that hallucinates, the agent explores conversations agentically:

1. `person_stats` → identify 4-8 pivotal periods from volume patterns
2. `get_person_messages_range` → deep-dive into each period's actual messages
2.5. Photo curation → visually inspect candidate photos, select best 15-25
3. Write chapters grounded in real quotes and events
4. `render_story` → combine narrative with data dashboards into HTML

**Usage:** `/generate-story Jenn` from Claude Code in this project.

**Command file:** `.claude/commands/generate-story.md`

## Key files

- `internal/app/app.go` — data dir resolution (`GMESSAGES_DATA_DIR` env var; default is `~/.local/share/gmessages`)
- `internal/db/db.go` — schema, structs, migration
- `internal/importer/` — gchat.go, whatsapp.go, whatsapp_native.go, signal_desktop.go
- `internal/story/stats.go` — conversation statistics computation (with timezone + per-sender phrases)
- `internal/story/generate.go` — narrative story generation (local or Claude API)
- `internal/viz/` — relationship visualization renderer (config, template, render, photos)
- `internal/client/events.go` — handles Google Messages protocol events
- `cmd/serve.go` — starts the web UI, MCP transports, and live bridges
