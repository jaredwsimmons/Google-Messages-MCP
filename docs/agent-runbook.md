# Agent & operator runbook

Hard-won operational knowledge for working on a **live** Google Messages MCP
install (debugging sends, re-pairing, reading real messages). If you are an
automated agent doing a support task, read this first — most of it cost hours to
learn the hard way.

## Data layout

There is a single data directory, resolved by `internal/app/app.go`:

| Used by | Path | Notes |
|---|---|---|
| CLI + server | `GMESSAGES_DATA_DIR`, else `~/.local/share/gmessages/` | Holds `messages.db` + `session.json` and the per-platform session stores. |

`gmessages read/status/pair/serve` all use this directory. If you set
`GMESSAGES_DATA_DIR`, set it consistently across every invocation so you are
always looking at the same store.

## Reading live messages

A running `serve` holds `messages.db` open in WAL mode, so a second SQLite
reader often fails with `unable to open database file (14)`, and `?immutable=1`
opens but misses WAL-only (recent) writes. **Prefer the running server's HTTP
API** (loopback-guarded; `curl` from localhost passes the origin check):

```
GET /api/status
GET /api/conversations?limit=500
GET /api/conversations/<conversation_id>/messages?limit=N
GET /api/search?q=<term>
```

Outgoing message rows carry a `Status`: `OUTGOING_SENDING` → `OUTGOING_SENT`/
`OUTGOING_DELIVERED`, or `OUTGOING_FAILED:<STATUS>` when a send is rejected.

## Pairing & the "zombie session"

**Symptom:** sends fail with `OUTGOING_FAILED:UNKNOWN`; `/api/status` shows
`google.connected=true`; reconnects and restarts don't help. The Google Messages
**linked-device session has lapsed** — the phone silently unlinked the device
(common after travel / network changes). The connection flag lies; the session
is dead for sends.

Key facts:

- The web UI surfaces a "Google Messages isn't sending — Re-pair" banner when
  `google.needs_repair` is set (3 consecutive Google send failures while
  connected).
- **QR pairing may be unavailable** — Google has disabled device-pairing QR for
  many accounts. Use **Google Account pairing** (`--google`) in that case.

### Re-pair recipe (the one that works)

1. Stop the server.
2. Force the pairing flow by removing `session.json` from the data dir (back it
   up first): `<data-dir>/session.json`. Other platforms' sessions
   (`whatsapp-session.db`, `signal-cli/`) are independent — leave them.
3. **Clear the stale session FIRST (don't skip).** Running `pair --google` while
   a dead `session.json` is still present floods the pairing with
   `failed to decrypt data event: HMAC mismatch` and yields a new session that
   401s on token refresh **immediately** (dead on arrival). Removing
   `session.json` before pairing is what produces a healthy session that
   connects *and* syncs (`/api/status` freshness `behind_days` drops to 0). Some
   HMAC-mismatch lines are normal noise; the tell for a bad pair is an immediate
   post-pair 401, not the noise itself.
4. Use the **cookie method**: open `messages.google.com/web` in a signed-in
   browser, copy the request cookies for `messages.google.com` from devtools (or
   a full `curl` command for `messages.google.com/web/config`), and run:
   ```
   Get-Clipboard | gmessages pair --google      # PowerShell
   # or
   gmessages pair --google-file <cookiefile>
   ```
   **Google cookies rotate roughly every 30 minutes** → extract fresh
   **immediately** before pairing, or `pair --google` returns HTTP 401. On Linux
   the helpers in `scripts/` can refresh cookies from a local Chrome profile.
5. `pair --google` prints `EMOJI: <emoji>`. Tap that emoji in Google Messages
   **on the phone** (notification shade, or profile → Device pairing) to confirm.
   The Gaia client init can time out once — just retry.
6. On confirmation the session saves to the data dir; restart `serve` and sends
   work. Wipe any cookie file afterwards.

### Don't over-reconnect

Connecting/disconnecting the Google web session many times in a short window
(repeated restarts, `reconnect` calls, multiple `pair` runs) gets the account
**throttled** — the long-poll drops and `/api/status` shows
`"Google Messages connection lost; reconnecting…"` in a loop with a perfectly
valid session. The fix is to **stop and let it cool down** (minutes up to ~1h),
not to hammer reconnect. Sends may land in brief connected windows meanwhile.

## signal-cli

Require **signal-cli ≥ 0.14.5**. 0.14.1 throws
`NullPointerException: …getSender() … content is null` on certain inbound
envelopes, exits non-zero, never ACKs, and re-hits the same poison message every
poll — a crash loop that flaps the Signal `connected` flag every few seconds and
makes the whole UI flicker. Upgrading signal-cli fixes it.

## Verifying after support work

- `curl -s -o /dev/null -w '%{http_code}' http://127.0.0.1:7007/` → 200
- `/api/status` → `google/whatsapp/signal` connection + `google.needs_repair`
- A real send shows `OUTGOING_DELIVERED` in
  `/api/conversations/<id>/messages`. Don't re-send a user's real message as a
  "test" (duplicate risk on `UNKNOWN`, which is ambiguous about whether it sent);
  if you must test connectivity, get explicit per-send permission.
