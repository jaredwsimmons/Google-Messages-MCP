#!/usr/bin/env bash
# Reproduce the Swift 6 concurrency crash that fresh installs hit when the
# Contacts TCC prompt fires. Resets Contacts + Notifications grants for the
# bundle, launches the installed app, and verifies the process is still
# alive 15 seconds later.
#
# Run after building and installing a new .app bundle. Exits 0 on pass.
set -euo pipefail

BUNDLE_ID="${BUNDLE_ID:-com.openmessage.app}"
APP_PATH="${APP_PATH:-/Applications/OpenMessage.app}"
WAIT_SEC="${WAIT_SEC:-15}"

if [ ! -d "$APP_PATH" ]; then
    echo "FAIL: $APP_PATH not found. Install the DMG first."
    exit 1
fi

echo "==> Resetting TCC grants for $BUNDLE_ID"
tccutil reset AddressBook "$BUNDLE_ID" || true
tccutil reset UserNotifications "$BUNDLE_ID" || true

echo "==> Killing any existing instance"
pkill -x OpenMessage || true
sleep 1

echo "==> Launching $APP_PATH"
open "$APP_PATH"

echo "==> Waiting ${WAIT_SEC}s for TCC callbacks to settle"
sleep "$WAIT_SEC"

if pgrep -x OpenMessage > /dev/null; then
    echo "PASS: OpenMessage still running after ${WAIT_SEC}s"
    exit 0
fi

echo "FAIL: OpenMessage exited. Check the latest crash log:"
ls -t ~/Library/Logs/DiagnosticReports/OpenMessage-*.ips 2>/dev/null | head -1
exit 1
