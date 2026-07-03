# Google Messages MCP

A local-first messaging workspace for **Google Messages** (SMS + RCS), with live
**WhatsApp** and **Signal** companions, and a built-in **MCP server**. Runs
headless or behind a localhost web UI, so you can read and send from any
MCP-compatible client (Claude Code, etc.) on **Windows**, Linux, or macOS.

Built on [mautrix/gmessages](https://github.com/mautrix/gmessages) (libgm) for
the Google Messages protocol and [mcp-go](https://github.com/mark3labs/mcp-go)
for the MCP server.

> This is a Windows-oriented fork of the upstream OpenMessage project. The
> native macOS app, App Store tooling, and marketing site have been removed;
> what remains is the cross-platform Go backend, web UI, and MCP server.

## What it does

- **Google Messages** — pair your Android phone and read/send SMS + RCS locally
- **Live WhatsApp** — link WhatsApp as a companion device on your machine
- **Live Signal** — link Signal locally and keep its threads in the same inbox
- **One local inbox** — search, route-aware threads, media, reactions, drafts, grouped contacts
- **Web UI** — a localhost UI with real-time conversation view
- **MCP-ready** — expose the same local inbox to Claude Code and other MCP clients
- **Local storage** — SQLite database; your data stays on your machine

## Quick start

### Prerequisites

- **Go 1.25+** ([install](https://go.dev/dl/))
- **Google Messages** on your Android phone

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

If Google only offers account pairing, you can pair with Google account cookies
copied from your browser's devtools. Copy the cookie JSON (or a full `curl`
command for `messages.google.com/web/config`), then pipe it in:

```powershell
Get-Clipboard | .\gmessages.exe pair --google
```

It then prompts you to confirm an emoji on your phone.

### 3. Start the server

```powershell
.\gmessages.exe serve
```

This starts both:
- **Web UI** at [http://127.0.0.1:7007](http://127.0.0.1:7007)
- **MCP SSE endpoint** at `http://127.0.0.1:7007/mcp/sse`

When `serve` is launched by an MCP client over pipes, it also serves MCP on stdio automatically.

### 3a. Optional: link WhatsApp or Signal

After `serve` is running, open the local UI and link WhatsApp or Signal from the
Connections surface. Those bridges stay local and sync into the same inbox as
Google Messages.

### 3b. Safe demo mode for screenshots and recordings

```powershell
.\gmessages.exe demo
```

Demo mode starts the same local UI on the normal port, but uses a fresh
temporary data directory, seeds fake conversations, and disables all live sync —
the safest way to capture screenshots without real messages bleeding in.

### 4. Connect to Claude Code

Add to `~/.mcp.json`:

```json
{
  "mcpServers": {
    "gmessages": {
      "command": "C:\\path\\to\\gmessages.exe",
      "args": ["serve"]
    }
  }
}
```

Restart Claude Code. The MCP tools appear automatically.

## Features

- **Read messages** — full conversation history, search, media, replies, reactions
- **Send messages** — SMS/RCS plus live WhatsApp and Signal text, media, and reactions
- **Live WhatsApp sync** — inbound messages, typing indicators, read state, and media
- **Live Signal sync** — inbound messages, media, reactions, and group threads
- **React to messages** — emoji reactions on any message
- **Image/media display** — inline images, video, audio, and fullscreen viewer
- **Web UI** — real-time conversation view at localhost:7007
- **MCP tools** — conversation lookup, route-aware sends, media download, import + story/viz tools
- **Local storage** — SQLite database, your data stays on your machine

## MCP tools

| Tool | Description |
|------|-------------|
| `get_messages` | Recent messages with filters (phone, date range, limit) |
| `get_conversation` | Messages in a specific conversation |
| `search_messages` | Full-text search across all messages |
| `send_message` | Send a direct text by platform. Defaults to SMS/RCS and also supports direct WhatsApp/Signal recipients |
| `send_to_conversation` | Send a text reply directly to an existing conversation ID |
| `send_media_to_conversation` | Send a local file attachment to an existing conversation ID |
| `react_to_message` | Add, remove, or switch a reaction on an existing message |
| `list_conversations` | List recent conversations |
| `list_contacts` | List/search contacts |
| `get_status` | Google Messages, WhatsApp, and Signal connection status |
| `download_media` | Download an attachment from a message to a local temp file |
| `draft_message` | Save a draft for the local app to review/send later |
| `import_messages` | Import Google Chat, WhatsApp export, or Signal Desktop history |

## MCP examples

- List recent Signal threads: `list_conversations(source_platform="signal")`
- Search WhatsApp for a keyword: `search_messages(query="airbnb")`
- Send a direct Signal message: `send_message(platform="signal", recipient="+15551230000", message="On my way")`
- Send a text into a route-aware thread: `send_to_conversation(conversation_id="whatsapp:15551234567@s.whatsapp.net", message="On my way")`
- Send a photo from disk: `send_media_to_conversation(conversation_id="signal-group:abc123", file_path="C:\\tmp\\photo.jpg", caption="Here")`
- React to a message: `react_to_message(conversation_id="signal-group:abc123", message_id="signal:...", emoji="🔥")`

## Web UI

The web UI runs at `http://localhost:7007` when the server is started. It provides:

- Conversation list with search and grouped multi-route contacts
- Message view with images, video, audio, reactions, and reply threads
- Route-aware compose and send
- Google Messages + WhatsApp + Signal connection controls
- Live typing indicators, read-state rendering, and notifications

## Configuration

| Env var | Default | Purpose |
|---------|---------|---------|
| `GMESSAGES_DATA_DIR` | `~/.local/share/gmessages` | Data directory (DB + session) |
| `GMESSAGES_LOG_LEVEL` | `info` | Log level (debug/info/warn/error/trace) |
| `GMESSAGES_PORT` | `7007` | Web UI port |
| `GMESSAGES_HOST` | `127.0.0.1` | Host/interface to bind the local web server to |
| `GMESSAGES_MY_NAME` | system user name | Display name for outgoing imported WhatsApp/Signal messages |
| `GMESSAGES_STARTUP_BACKFILL` | `auto` | Startup history sync mode: `auto`, `shallow`, `deep`, or `off` |
| `GMESSAGES_BACKFILL_DISCOVER_ORPHANS` | `0` | Opt in to deep backfill's Phase C (contact-based orphan discovery). **Off by default** because it creates an empty SMS thread on your phone for each contact without prior message history. Enable with `1`/`true`/`yes`/`on` only if you understand the side effect. |
| `GMESSAGES_KLIPY_API_KEY` | unset | Optional Klipy API key for the web compose GIF picker. GIF search is unavailable until this is set. `KLIPY_API_KEY` is also accepted as a fallback. |
| `GMESSAGES_SIGNAL_TMP_SWEEP` | enabled | Set to `0` to disable cleanup of stale signal-cli temp directories. |
| `GMESSAGES_EXPORT_DIR` | `~/Documents/GoogleMessagesMCP` | Directory for `generate_viz` / `render_story` HTML outputs and photo inputs when using the default confined export mode. |
| `GMESSAGES_ALLOW_ANY_EXPORT_PATH` | unset (off) | Set to `1`/`true`/`yes`/`on` to allow viz/story tools to read photos from, or write HTML to, arbitrary local paths. |

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

## Running as a headless server (Docker)

```bash
docker compose up -d
docker compose exec gmessages gmessages pair   # one-time
# then connect Claude / other MCP clients to http://localhost:7007/mcp/sse
```

## Development

```bash
go test ./...                 # Run all tests
go build .                    # Build binary
GOOS=windows go build ./...   # Verify the Windows build
npm install                   # Install Playwright test runner
npx playwright install chromium
npm run test:e2e              # Run browser-level web UI tests
```

Debugging a live install (failing sends, re-pairing)? See
[docs/agent-runbook.md](docs/agent-runbook.md).

## License

MIT
