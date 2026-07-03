# Release Checklist

Use this before tagging a release. Pushing a `v*` tag runs the `Release`
workflow, which cross-compiles the `gmessages` binary for Windows, Linux, and
macOS and publishes them to a GitHub Release.

## 1. Worktree and privacy preflight

- Run `git status --short` and confirm every changed file is intentional.
- Confirm private recovery files, chat exports, screenshots, and local databases are not tracked. In particular, `.tmp-*.sql` dumps must stay untracked and ignored locally.
- Check any new screenshots/assets for real names, private chats, phone numbers, and account data.
- Update `main.version` expectations / release notes for the tag.

## 2. Automated checks

- `go test ./...`
- `GOOS=windows go build ./...` (verify the Windows build compiles)
- `npm run test:e2e` (web UI)
- Start `gmessages serve` from a clean data dir and confirm the backend stays alive for a basic send/receive session.

## 3. Messaging dogfood matrix

- Google Messages: pairing, reconnect, SMS send/receive, RCS attribution, image receive, and read receipts.
- WhatsApp: pairing, reconnect, text send, image plus caption send, reaction send/receive, group leave, group names, and avatar loading.
- Signal: pairing, history/backfill, group names, image receive, reactions, and stale connection recovery.
- Long threads: opening, sending, receiving, older-message scrollback, media hydration, and bottom-pin behavior.
- Multi-route contacts: contact list chips, main-pane platform tabs, send target switching, and unread state.

## 4. Diagnostics

- Open Settings and export/copy diagnostics after a clean launch.
- Confirm `/api/diagnostics` includes `schema_version`, `generated_at_iso`, `backend`, `memory`, `capabilities`, platform counts, and bridge status snapshots.
- Attach diagnostics to issues when investigating crashes, backend exits, dropped connections, stale media, or missing notifications.
- Do not paste message bodies, contact exports, or database dumps into public issues.

## 5. Release gate

- No open P1/P2 review findings or known privacy leaks.
- No failing required CI checks.
- Release artifacts are generated from the intended commit/tag.
- Rollback path is known before promoting a public build.
