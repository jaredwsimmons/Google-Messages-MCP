# Homebrew cask for OpenMessage.
#
# To install once this lands in homebrew/cask:
#     brew install --cask openmessage
#
# To install from this repo while waiting on the upstream PR:
#     brew tap maxghenis/openmessage https://github.com/MaxGhenis/openmessage
#     brew install --cask maxghenis/openmessage/openmessage
#
# The release workflow publishes a SHA256SUMS file alongside the DMG; bump the
# `version` and `sha256` fields when cutting a new release.

cask "openmessage" do
  version "0.2.0"
  sha256 :no_check # replace with the real sha256 from SHA256SUMS once cut

  url "https://github.com/MaxGhenis/openmessage/releases/download/v#{version}/OpenMessage.dmg"
  name "OpenMessage"
  desc "Local-first desktop for SMS, RCS, WhatsApp, and Signal with built-in MCP"
  homepage "https://openmessage.ai"

  app "OpenMessage.app"

  zap trash: [
    "~/Library/Application Support/OpenMessage",
    "~/Library/Preferences/com.openmessage.OpenMessage.plist",
    "~/Library/Caches/com.openmessage.OpenMessage",
  ]
end
