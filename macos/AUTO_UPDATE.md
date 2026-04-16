# Auto-update setup (Sparkle)

OpenMessage doesn't ship auto-update yet. Adding it is straightforward but
requires an Apple Developer account (for code signing) and an EdDSA keypair.

This doc walks through the full setup so the next release can include it.

## Prerequisites

1. **Apple Developer Program** ($99/year). Without this, the DMG is ad-hoc
   signed and macOS Gatekeeper warns users on every install. Auto-update is
   pointless without trusted signatures.
2. **EdDSA keypair** for signing update payloads:
   ```bash
   brew install --cask sparkle
   generate_keys
   ```
   This prints a public key (goes in `Info.plist`) and writes the private key
   to the macOS keychain (used by the build script to sign each update).

## Add Sparkle to the Swift package

In `macos/OpenMessage/Package.swift`:

```swift
let package = Package(
    name: "OpenMessage",
    platforms: [.macOS(.v14)],
    dependencies: [
        .package(url: "https://github.com/sparkle-project/Sparkle", from: "2.6.0"),
    ],
    targets: [
        .executableTarget(
            name: "OpenMessage",
            dependencies: [
                .product(name: "Sparkle", package: "Sparkle"),
            ],
            path: "Sources",
            exclude: ["Info.plist"],
            resources: [.copy("Assets.xcassets")]
        ),
    ]
)
```

## Wire the updater controller

Add `Sources/Updater.swift`:

```swift
import Sparkle
import SwiftUI

final class Updater: NSObject, ObservableObject {
    let controller: SPUStandardUpdaterController

    override init() {
        controller = SPUStandardUpdaterController(
            startingUpdater: true,
            updaterDelegate: nil,
            userDriverDelegate: nil
        )
        super.init()
    }

    func checkForUpdates() {
        controller.checkForUpdates(nil)
    }
}
```

Inject it into `OpenMessageApp.swift` and add a "Check for Updates…" menu item
in `MenuBarView.swift`.

## Info.plist additions

```xml
<key>SUFeedURL</key>
<string>https://openmessage.ai/appcast.xml</string>
<key>SUPublicEDKey</key>
<string>YOUR_PUBLIC_KEY_HERE</string>
<key>SUEnableAutomaticChecks</key>
<true/>
<key>SUScheduledCheckInterval</key>
<integer>86400</integer>
```

## Appcast feed

The release workflow writes `appcast.xml` to `site/public/appcast.xml`. Each
GitHub release becomes one `<item>` entry in the feed. Use the existing
`SHA256SUMS` file to verify download integrity.

A minimal `appcast.xml` looks like:

```xml
<?xml version="1.0" standalone="yes"?>
<rss version="2.0" xmlns:sparkle="http://www.andymatuschak.org/xml-namespaces/sparkle">
  <channel>
    <title>OpenMessage</title>
    <link>https://openmessage.ai/appcast.xml</link>
    <description>OpenMessage updates</description>
    <language>en</language>
    <item>
      <title>Version 0.2.0</title>
      <pubDate>Sat, 12 Apr 2026 12:00:00 +0000</pubDate>
      <sparkle:version>0.2.0</sparkle:version>
      <sparkle:shortVersionString>0.2.0</sparkle:shortVersionString>
      <description><![CDATA[<ul><li>WhatsApp + Signal support</li>...</ul>]]></description>
      <enclosure
        url="https://github.com/MaxGhenis/openmessage/releases/download/v0.2.0/OpenMessage.dmg"
        sparkle:edSignature="..."
        length="..."
        type="application/octet-stream"/>
    </item>
  </channel>
</rss>
```

## Build script signing

Update `macos/build.sh` to sign each DMG with the EdDSA private key:

```bash
sign_update macos/build/OpenMessage.dmg
```

Pipe the output into the appcast generator and commit `site/public/appcast.xml`
as part of the release workflow.
