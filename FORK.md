# About this fork

`jaredwsimmons/Google-Messages-MCP` is upstream
[MaxGhenis/openmessage](https://github.com/MaxGhenis/openmessage) plus a small
patch that makes it build and run on Windows. It keeps upstream's names
(`openmessage` binary, `OPENMESSAGES_*` variables, `~/.local/share/openmessage`
data directory) so upstream updates merge without conflicts.

## What the fork changes

Everything else is upstream's code, unchanged.

| Change | Files |
|---|---|
| Windows versions of the OS calls that differ: file lock, free-space check, directory sync, umask, refused-connection errors, and whether file mode bits are meaningful | `internal/osutil/` (new), and its callers in `cmd/backup.go`, `cmd/migrate.go`, `cmd/serve.go`, `cmd/status_silence.go`, `internal/localapi/localapi.go`, `internal/storage/blob/store.go` |
| Blob-store tests skip Unix mode assertions where Windows has no mode bits | `internal/storage/blob/store_test.go` |
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
upstream's main branch, and the rename was dropped.

If an install ran the renamed `gmessages` build, its data is in
`%USERPROFILE%\.local\share\gmessages` (Windows) or `~/.local/share/gmessages`.
Point `OPENMESSAGES_DATA_DIR` at that folder, or move it to the `openmessage`
path. Before the first run of the new build, stop the old server and run
`openmessage backup`, which writes a verified copy under
`<data-dir>/migration-backups/`.

## Known issues

- `TestRunSignalCLIHoldsAReceiveStoringAnAttachmentUntilTheStallWindow` in
  `internal/signallive` fails intermittently on Linux. It fails on unmodified
  upstream too, so it is not caused by this patch.
- Upstream's release workflow stamps manual (`workflow_dispatch`) releases
  with the branch name instead of the tag: `${GITHUB_REF_NAME:-...}` is never
  empty. Tag pushes are unaffected.
