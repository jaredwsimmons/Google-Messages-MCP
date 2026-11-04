# Code signing & notarization

The current DMG is **ad-hoc signed**, which means macOS Gatekeeper warns users
on first launch ("OpenMessage cannot be opened because the developer cannot be
verified") and they must right-click → Open. This is the single biggest
adoption blocker.

Fixing it requires an Apple Developer account and a one-time setup. Once
configured, every release is signed and notarized automatically by the GitHub
Actions workflow.

## One-time setup (~30 minutes)

### 1. Apple Developer Program enrollment

[developer.apple.com/programs/enroll](https://developer.apple.com/programs/enroll)
($99/year, individual). Required for:
- Distributing apps outside the Mac App Store
- Notarization through Apple's service
- Sparkle auto-update (signing the appcast)

### 2. Generate a Developer ID Application certificate

In Xcode → Settings → Accounts → Manage Certificates → `+` → "Developer ID
Application". This installs a certificate in your login keychain.

The certificate's identity name looks like `Developer ID Application: Max
Ghenis (TEAMID)`. Find it with:

```bash
security find-identity -v -p codesigning
```

### 3. App-specific password for notarization

Apple ID → [appleid.apple.com](https://appleid.apple.com) → Sign-In and
Security → App-Specific Passwords → Generate. Label it "OpenMessage notary".

### 4. Store credentials in GitHub Secrets

In the repo: Settings → Secrets and variables → Actions → New repository
secret. Add four secrets:

| Secret | Value |
|---|---|
| `DEVELOPER_ID` | The full identity name from `find-identity` (e.g. `Developer ID Application: Max Ghenis (ABCD123EFG)`) |
| `AC_USERNAME` | Your Apple ID email |
| `AC_PASSWORD` | The app-specific password from step 3 |
| `AC_TEAM_ID` | Your Apple Developer Team ID (10-char string, found at developer.apple.com/account) |
| `MACOS_CERT_P12_BASE64` | Your Developer ID Application cert exported as `.p12`, then base64-encoded |
| `MACOS_CERT_PASSWORD` | The password you set when exporting the `.p12` |

To export the cert:
```bash
# In Keychain Access, right-click the cert → Export → .p12, set a password
base64 -i DeveloperID.p12 | pbcopy
```

### 5. Update the build script

`macos/build.sh` already checks for `DEVELOPER_ID`. Once it's set, it signs
the bundle. The release workflow imports the cert into a temporary keychain
before building:

```yaml
- name: Import signing certificate
  run: |
    echo "${{ secrets.MACOS_CERT_P12_BASE64 }}" | base64 -d > cert.p12
    security create-keychain -p actions build.keychain
    security default-keychain -s build.keychain
    security unlock-keychain -p actions build.keychain
    security import cert.p12 -k build.keychain \
      -P "${{ secrets.MACOS_CERT_PASSWORD }}" \
      -T /usr/bin/codesign
    security set-key-partition-list -S apple-tool:,apple:,codesign: \
      -s -k actions build.keychain

- name: Notarize and staple DMG
  env:
    AC_USERNAME: ${{ secrets.AC_USERNAME }}
    AC_PASSWORD: ${{ secrets.AC_PASSWORD }}
    AC_TEAM_ID: ${{ secrets.AC_TEAM_ID }}
  run: |
    xcrun notarytool submit macos/build/OpenMessage.dmg \
      --apple-id "$AC_USERNAME" \
      --password "$AC_PASSWORD" \
      --team-id "$AC_TEAM_ID" \
      --wait
    xcrun stapler staple macos/build/OpenMessage.dmg
```

## Verifying

After notarization, this should pass:

```bash
spctl --assess --type install --verbose macos/build/OpenMessage.dmg
# Expected: macos/build/OpenMessage.dmg: accepted
#           source=Notarized Developer ID
```
