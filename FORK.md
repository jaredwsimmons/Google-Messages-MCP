# About this fork

`jaredwsimmons/openmessage` is upstream
[MaxGhenis/openmessage](https://github.com/MaxGhenis/openmessage) plus a small
patch that makes it build and run on Windows. It keeps upstream's names
(`openmessage` binary, `OPENMESSAGES_*` variables, `~/.local/share/openmessage`
data directory) so upstream updates merge without conflicts.

## What the fork changes

Everything else is upstream's code, unchanged.

| Change | Files |
|---|---|
| Windows versions of the OS calls that differ: file lock, free-space check, directory and file sync, umask, refused-connection errors, whether file mode bits are meaningful, SQLite file URIs for drive-letter paths, and session-file reads and renames that retry while Windows reports the file busy | `internal/osutil/` (new), and its callers in `cmd/backup.go`, `cmd/migrate.go`, `cmd/serve.go`, `cmd/status_silence.go`, `internal/client/session.go`, `internal/googlecookies/googlecookies.go`, `internal/localapi/localapi.go`, `internal/migration/source.go`, `internal/migration/validate.go`, `internal/storage/blob/store.go`, `internal/whatsapplid/sources.go` |
| Embedded migration SQL keeps LF line endings on every checkout | `.gitattributes` (new) |
| Blob-store and session tests skip Unix mode assertions where Windows has no mode bits; a messaging test helper uses the SQLite URI helper | `internal/storage/blob/store_test.go`, `internal/client/session_test.go`, `internal/messaging/dispatch_reconcile_test.go` |
| Windows CI job: build, vet, platform-layer tests, and the full suite reported without failing the job | `.github/workflows/test.yml` |
| Windows amd64/arm64 zips in GitHub releases | `.github/workflows/release.yml` |
| Docker Compose publishes port 7007 on loopback only | `docker-compose.yml` |
| Windows setup notes | `README.md` ("Windows"), `CLAUDE.md` (header), this file |

Without the `internal/osutil` patch, upstream does not compile for Windows
(`syscall.Flock`, `syscall.Statfs` and `syscall.Umask` do not exist there). It
would also fail at runtime in three ways. Go does not map Winsock's
`WSAECONNREFUSED` to `syscall.ECONNREFUSED`, so `pair`, `backup` and `migrate`
would refuse to run while the app is stopped. `File.Sync` on a directory fails
with "Access is denied", which breaks attachment storage, backups and
migration. And a check that blob files are mode 0600 can never pass, so
re-storing an existing attachment would fail.

The first two Windows CI runs found four more:

- **SQLite file URIs.** `backup`, `migrate` and the WhatsApp LID repair
  built `file://C:/...` for a drive path. SQLite reads `C:` as the URI's
  authority and refuses the file, which `modernc.org/sqlite` reports as
  "out of memory". The v2 store's own `storeDSN` already added the slash
  that makes `file:///C:/...`; `osutil.SQLiteFileURI` now does the same
  for the others.
- **CRLF migrations.** Git for Windows checks text files out with CRLF by
  default. The v2 store embeds its migration SQL and records each file's
  SHA-256 in the store, so a store written by a CRLF build was refused by an
  LF build (`ErrMigrationChecksumMismatch`), and the reverse. `.gitattributes`
  pins `*.sql` to LF.
- **Busy session file.** Replacing `session.json` by rename while another
  handle reads it fails on Windows with "Access is denied" or a sharing
  violation. Session reads and renames now retry for up to two seconds,
  as Go's own toolchain does (`cmd/go/internal/robustio`).
- **Read-only file flush.** The v2 migration's last step opened the staged
  store read-only and flushed it. Windows needs write access to flush a
  file, so the step failed with "Access is denied"; `osutil.SyncFile` opens
  it for writing there.

The patch is written to be offered upstream. If upstream accepts it, this fork
can track upstream with no changes of its own.

## Pulling upstream updates

```bash
git remote add upstream https://github.com/MaxGhenis/openmessage   # once
git fetch upstream main
git merge upstream/main
GOOS=windows GOARCH=amd64 go vet ./...
go test ./...
```

If upstream adds another Unix-only call, the Windows vet step or the Windows
CI job fails. Add the Windows variant to `internal/osutil` rather than
patching the caller.

## History

Before 2026-11-03 this fork renamed the project to "Google Messages MCP" (binary
`gmessages`, `GMESSAGES_*` variables, data directory
`~/.local/share/gmessages`) and removed the macOS app, website and iMessage
importer. That made upstream's later fixes impossible to merge. Upstream
OpenMessage had 124 commits after the fork point, including fixes for unsafe
attachment serving, HTTP MCP being on by default, stdio MCP sessions logging
out WhatsApp, and a stale Google Messages library. So the fork was rebuilt on
upstream's main branch, and the rename was dropped. The GitHub repository
itself was renamed from `jaredwsimmons/Google-Messages-MCP` to
`jaredwsimmons/openmessage` on 2026-11-03; GitHub redirects the old URL.

## Upgrading an install built before 2026-11-03

**MCP no longer syncs on its own.** An MCP entry that runs
`openmessage serve --mcp-stdio` (with or without `--no-mcp-sse`) used to start
its own Google, WhatsApp and Signal connections. Upstream now runs that shape
as a client of a separate server (`cmd/serve.go`, `runServeMCPClient`): it
reads the local store and sends through the running app, but fetches no new
messages itself. Do one of these:

1. Recommended: keep the MCP entry as it is, and keep `openmessage serve`
   running separately, for example from Task Scheduler at log on.
2. Add `--transports` to the MCP entry's args to restore the one-process
   behavior. Every Claude session then opens its own connections. Upstream
   found this logs WhatsApp out ("logged out from another device") and
   corrupts Signal when two processes share credentials.

Before the first run of the new build, stop the old process and run
`openmessage backup`, which writes a verified copy of the store under
`<data-dir>/migration-backups/`. The new build keeps using the existing
store. Moving to upstream's v2 store is opt-in: run `openmessage migrate`,
then set `OPENMESSAGES_V2_PRIMARY=1` (`cmd/v2stack.go`).

If an install ran the renamed `gmessages` build, its data is in
`%USERPROFILE%\.local\share\gmessages` (Windows) or `~/.local/share/gmessages`.
Point `OPENMESSAGES_DATA_DIR` at that folder, or move it to the `openmessage`
path.

## Known issues

- **The full test suite does not pass on Windows yet**, so CI runs it as
  information only. The remaining failures are tests that assume Unix, not
  product bugs. Most have been checked against the log; the last two causes
  are inferred from their messages rather than reproduced:
  - Mode assertions (0600/0700) that Windows cannot satisfy: `cmd`
    (`pair_guard_test.go`, `v2stack_test.go`), `internal/app`,
    `internal/web` (control token), `internal/tools` (saved attachments).
  - Tests that build the binary as `openmessage` without `.exe`, or run
    `.sh` and shell-script stubs: `cmd` (`binary_test.go`,
    `TestRefreshGoogleSessionCookiesUsesEnvScript`), `internal/signallive`.
  - macOS-only WhatsApp Desktop media paths: `internal/whatsapplive`.
  - A `?` in a test filename, which Windows forbids:
    `internal/storage/sqlite` (`TestOpenInitializesBlankDatabase`).
  - A read-only-directory trick Windows ignores
    (`TestRunPairStopsBeforeThePhoneWhenTheSessionCannotBeBackedUp`) and Unix
    signals (`TestPairSignalContext`).
  - `TestRunServeMCPClientAdoptsDaemonDataDir` (likely short `RUNNER~1` vs
    long path names) and `TestLinkPreviewServiceEvictsLeastRecentlyUsedEntries`
    (cause not yet identified).
- **Existing Windows clones keep CRLF migration files** until they are
  checked out again. After pulling the `.gitattributes` change, run this once
  in PowerShell from the repository root, then rebuild:

  ```powershell
  Remove-Item internal\storage\sqlite\migrations\*.sql
  git checkout -- internal/storage/sqlite/migrations
  ```

- `TestRunSignalCLIHoldsAReceiveStoringAnAttachmentUntilTheStallWindow` in
  `internal/signallive` fails intermittently on Linux. It fails on unmodified
  upstream too, so it is not caused by this patch.
- Upstream's release workflow stamps manual (`workflow_dispatch`) releases
  with the branch name instead of the tag: `${GITHUB_REF_NAME:-...}` is never
  empty. Tag pushes are unaffected.
