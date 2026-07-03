# Google Messages MCP

A local-first messaging workspace for **Google Messages** (SMS + RCS), with live
**WhatsApp** and **Signal** companions, and a built-in **MCP server**. Runs
headless or behind a localhost web UI, so you can read and send from any
MCP-compatible client (Claude Code, etc.) on **Windows**, Linux, or macOS.

Built on [mautrix/gmessages](https://github.com/mautrix/gmessages) (libgm) for
the Google Messages protocol and [mcp-go](https://github.com/mark3labs/mcp-go)
for the MCP server.

> This is a Windows-oriented fork of the upstream
> [OpenMessage](https://github.com/MaxGhenis/openmessage) project. The native
> macOS app, App Store tooling, and marketing site have been removed; what
> remains is the cross-platform Go backend, web UI, and MCP server.

## What it does

- **Google Messages** — pair your Android phone and read/send SMS + RCS locally
- **Live WhatsApp** — link WhatsApp as a companion device on your machine
- **Live Signal** — link Signal as a linked device and keep its threads in the same inbox
- **One local inbox** — search, route-aware threads, media, reactions, drafts, grouped contacts
- **Web UI** — a localhost UI with a real-time conversation view, media, and typing/read state
- **MCP-ready** — expose the same local inbox to Claude Code and other MCP clients
- **Local storage** — everything lives in a local SQLite database; your data stays on your machine

## Quick start

### Prerequisites

- **Go 1.25+** ([install](https://go.dev/dl/))
- **Google Messages** on your Android phone
- *Optional* — **signal-cli ≥ 0.14.5** on your `PATH` (requires a **Java 17+** runtime) if you want live Signal
- *Optional* — **Node.js 18+** only to run the Playwright web-UI e2e tests
- *Optional* — **Docker** only for the headless-server path below

### 1. Clone and build

```powershell
git clone https://github.com/jaredwsimmons/google-messages-mcp.git
cd google-messages-mcp
go build -o gmessages.exe .
```

On Linux/macOS, build `-o gmessages` instead.

### 2. Pair with your phone

```powershell
.\gmessages.exe pair
```

By default, a QR code appears in your terminal. On your phone, open
**Google Messages > Settings > Device pairing > Pair a device** and scan it. The
session saves to your data dir as `session.json`.

If Google only offers account pairing (QR device-pairing is disabled for many
accounts), pair with Google account cookies copied from your browser's devtools.
Copy the cookie JSON (or a full `curl` command for
`messages.google.com/web/config`), then either pipe it in or point at a file:

```powershell
Get-Clipboard | .\gmessages.exe pair --google      # read cookies from stdin
.\gmessages.exe pair --google-file cookies.txt      # read cookies from a file
```

Either way, it prints an emoji and prompts you to confirm it in Google Messages
on your phone.

### 3. Start the server

```powershell
.\gmessages.exe serve
```

This starts:
- **Web UI** at [http://127.0.0.1:7007](http://127.0.0.1:7007)
- **MCP** over HTTP — SSE at `http://127.0.0.1:7007/mcp/sse` and streamable HTTP at `http://127.0.0.1:7007/mcp`

To expose MCP over **stdio** instead (for a client that launches the binary and
talks over pipes), run `gmessages serve --mcp-stdio`. Passing `--mcp-stdio`
switches `serve` to stdio-only — it does **not** also start the web/SSE server or
bind port 7007.

### 3a. Optional: link WhatsApp or Signal

After `serve` is running, open the local UI and link WhatsApp or Signal from the
Connections surface. Those bridges stay local and sync into the same inbox as
Google Messages. (Live Signal requires `signal-cli` — see Prerequisites.)

### 3b. Safe demo mode for screenshots and recordings

```powershell
.\gmessages.exe demo
```

Demo mode starts the same local UI on the normal port, but uses a fresh
temporary data directory, seeds fake conversations, and disables all live sync —
the safest way to capture screenshots without real messages bleeding in.

### 4. Connect to Claude Code

An `~/.mcp.json` `command`/`args` entry is a **stdio** MCP server, so it must
launch `serve` in stdio-only mode. On Windows this file lives at
`%USERPROFILE%\.mcp.json`:

```json
{
  "mcpServers": {
    "gmessages": {
      "command": "C:\\path\\to\\gmessages.exe",
      "args": ["serve", "--mcp-stdio"]
    }
  }
}
```

Alternatively, if you already run `gmessages serve` (web + SSE), point Claude
Code at the SSE endpoint instead of launching a second process:

```bash
claude mcp add -s user --transport sse gmessages http://127.0.0.1:7007/mcp/sse
```

Restart Claude Code and the MCP tools appear automatically.

## Command-line usage

Beyond `pair`/`serve`/`demo`, the binary is a full local CLI over the message
store. The read commands (`read`/`search`/`thread`/`threads`/`status`) only
touch the local SQLite DB and start no live transports; `send`/`send-group`
connect to Google Messages to actually deliver:

```powershell
.\gmessages.exe read "<query>" [--limit N] [--phone NUMBER] [--since YYYY-MM-DD] [--until YYYY-MM-DD] [--json]
.\gmessages.exe search "<query>" ...                 # alias for read
.\gmessages.exe thread <name|number|conversation_id> [--limit N] [--json]
.\gmessages.exe threads [--limit N] [--json]         # list recent conversations
.\gmessages.exe status [--json]                      # per-platform counts + sync freshness
.\gmessages.exe send <conversation_id> "<message>"   # connects to Google Messages to send
.\gmessages.exe send-group <phone1,phone2,...> "<message>"
.\gmessages.exe import <gchat|gchat-conversation|whatsapp|signal> [args...]
```

## MCP tools

All tools are registered in `internal/tools/tools.go` (`Register` is the
authoritative list). Cross-platform unless noted.

### Messaging

| Tool | Description |
|------|-------------|
| `get_messages` | Recent messages with filters (phone, date range, limit) |
| `get_conversation` | Messages in a specific conversation |
| `search_messages` | Full-text search across all messages |
| `send_message` | Send a direct text by platform. Defaults to SMS/RCS and also supports direct WhatsApp/Signal recipients |
| `send_to_conversation` | Send a text reply directly to an existing conversation ID |
| `send_media_to_conversation` | Send a local file attachment to an existing conversation ID |
| `send_group_message` | Send a text to a group/MMS conversation, creating it if needed |
| `react_to_message` | Add, remove, or switch a reaction on an existing message |
| `draft_message` | Save a draft for the local app to review/send later |
| `download_media` | Download an attachment from a message to a local temp file |
| `set_message_transcript` | Save/overwrite a transcript for an existing message (original body preserved) |

### Contacts, routing & status

| Tool | Description |
|------|-------------|
| `list_conversations` | List recent conversations (optional `source_platform` filter) |
| `list_contacts` | List/search contacts |
| `resolve_contact_routes` | Resolve a person/number/name to existing conversation routes + preferred reply route |
| `get_status` | Google Messages, WhatsApp, and Signal connection status |
| `import_messages` | Import Google Chat, WhatsApp export, or Signal Desktop history |

### Analytics, story & visualization

| Tool | Description |
|------|-------------|
| `get_person_messages` | All messages with a person across all platforms |
| `get_person_messages_range` | Date-ranged version of `get_person_messages` |
| `conversation_stats` | Volume, heatmap, phrases, response times, gaps (single conversation) |
| `person_stats` | Cross-platform stats for all 1:1 messages with a person |
| `generate_story` | Narrative chapters with optional Claude API enhancement |
| `generate_person_story` | Cross-platform narrative story for a person |
| `generate_viz` | Self-contained HTML visualization combining dashboards + narrative |
| `render_story` | Render a pre-built Story JSON into an HTML viz |

## MCP examples

- List recent Signal threads: `list_conversations(source_platform="signal")`
- Search WhatsApp for a keyword: `search_messages(query="airbnb")`
- Send a direct Signal message: `send_message(platform="signal", recipient="+15551230000", message="On my way")`
- Send a text into a route-aware thread: `send_to_conversation(conversation_id="whatsapp:15551234567@s.whatsapp.net", message="On my way")`
- Send a photo from disk: `send_media_to_conversation(conversation_id="signal-group:abc123", file_path="C:\\tmp\\photo.jpg", caption="Here")`
- React to a message: `react_to_message(conversation_id="signal-group:abc123", message_id="signal:...", emoji="🔥")`

## Web UI

The web UI runs at `http://127.0.0.1:7007` when the server is started. It provides:

- Conversation list with search and grouped multi-route contacts
- Message view with images, video, audio, reactions, and reply threads
- Route-aware compose and send
- Google Messages + WhatsApp + Signal connection controls
- Live typing indicators, read-state rendering, and notifications

## Configuration

| Env var | Default | Purpose |
|---------|---------|---------|
| `GMESSAGES_DATA_DIR` | `~/.local/share/gmessages` | Data directory (DB + session). On Windows this resolves under your home dir — `C:\Users\<you>\.local\share\gmessages` — **not** `%APPDATA%`. |
| `GMESSAGES_LOG_LEVEL` | `info` | Log level (debug/info/warn/error/trace) |
| `GMESSAGES_PORT` | `7007` | Web UI / MCP HTTP port |
| `GMESSAGES_HOST` | `127.0.0.1` | Host/interface to bind the local server to (loopback-only by default) |
| `GMESSAGES_MY_NAME` | OS user name (falls back to `Me`) | Display name attributed to your own outgoing WhatsApp/Signal messages (live sends and imports) |
| `GMESSAGES_STARTUP_BACKFILL` | `auto` | Startup history sync mode: `auto`, `shallow`, `deep`, or `off` |
| `GMESSAGES_BACKFILL_DISCOVER_ORPHANS` | `0` | Opt in to deep backfill's Phase C (contact-based orphan discovery). **Off by default** because it creates an empty SMS thread on your phone for each contact without prior message history. Enable with `1`/`true`/`yes`/`on` only if you understand the side effect. |
| `GMESSAGES_GOOGLE_AVATAR_SYNC` | enabled | Set to `0`/`false`/`off`/`disabled` to stop syncing Google contact avatars |
| `GMESSAGES_SIGNAL_CLI` | auto-detected | Override the path/command for the `signal-cli` binary used by the live Signal bridge |
| `GMESSAGES_SIGNAL_TMP_SWEEP` | enabled | Set to `0` to disable cleanup of stale signal-cli temp directories |
| `GMESSAGES_COOKIE_REFRESH_SCRIPT` | unset | Path to an executable that refreshes Google auth cookies; when set, the reconnect watchdog auto-refreshes cookies instead of requiring a manual re-pair |
| `GMESSAGES_KLIPY_API_KEY` | unset | Optional Klipy API key for the web compose GIF picker. `KLIPY_API_KEY` is also accepted as a fallback. |
| `GMESSAGES_EXPORT_DIR` | `~/Documents/GoogleMessagesMCP` | Directory for `generate_viz` / `render_story` HTML outputs and photo inputs when using the default confined export mode |
| `GMESSAGES_ALLOW_ANY_EXPORT_PATH` | unset (off) | Set to `1`/`true`/`yes`/`on` to allow viz/story tools to read photos from, or write HTML to, arbitrary local paths |

### GIF picker setup

The web compose GIF picker uses Klipy search. To enable it:

1. Create a Klipy API key.
2. Start the server with `GMESSAGES_KLIPY_API_KEY=<your key>`.
3. Open the web UI and use the `GIF` button in a media-capable conversation.

Without `GMESSAGES_KLIPY_API_KEY`, GIF search endpoints return a setup error and
the rest of messaging continues to work normally.

## Architecture

- **libgm** handles the Google Messages protocol (pairing, encryption, long-polling)
- **whatsmeow** handles live WhatsApp pairing, sync, text/media send, receipts, typing, and avatars through a separate local session store
- **signal-cli** powers the local Signal linked-device bridge, message sync, media, and reactions
- **SQLite** (WAL mode, pure Go) stores messages, conversations, and contacts locally
- Real-time events from the phone are written to SQLite as they arrive
- WhatsApp/Signal Desktop import remains as a fallback/repair path when a live bridge is not active
- On first run, a deep backfill fetches full SMS/RCS history in the background; later runs do a lighter incremental sync by default
- MCP tool handlers read from SQLite for queries and route sends through the same local runtime
- Auth tokens auto-refresh and persist to `session.json`

## Privacy & security

- All messages and contacts are stored **locally** in a SQLite database in your data dir.
- The server binds to `127.0.0.1` by default — it is not reachable from other machines. Changing `GMESSAGES_HOST` (e.g. to `0.0.0.0`) exposes the web UI and MCP endpoint to your network; do so only behind trusted access controls.
- Google account-pairing cookies are sensitive credentials. Delete any cookie file you created after pairing succeeds.

## Running as a headless server (Docker)

```bash
docker compose up -d
# One-time pairing. QR pairing may be unavailable for your account; the account
# (cookie) method is usually more reliable in a headless container:
docker compose exec gmessages gmessages pair --google-file /data/cookies.txt
# then connect Claude / other MCP clients to http://127.0.0.1:7007/mcp/sse
```

## Development

```bash
go test ./...                 # Run all tests
go build .                    # Build the binary (targets the host OS)
npm install                   # Install the Playwright test runner
npx playwright install chromium
npm run test:e2e              # Run browser-level web UI tests
```

To cross-compile a Windows binary from Linux/macOS: `GOOS=windows go build .`
(on Windows, a plain `go build` already targets Windows). CI verifies the
Windows build on every push.

Debugging a live install (failing sends, re-pairing)? See
[docs/agent-runbook.md](docs/agent-runbook.md).

## License

MIT
