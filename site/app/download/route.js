import { NextResponse } from "next/server";

import { releaseDownloadUrl } from "../../data/site-content";
import { trackGoogleAnalyticsEvent } from "../../lib/analytics";

export const dynamic = "force-dynamic";

function normalizeValue(value, fallback) {
  const normalized = `${value || ""}`.trim().toLowerCase().replace(/[^a-z0-9-_]/g, "-");
  return normalized || fallback;
}

export async function GET(request) {
  const url = new URL(request.url);
  const source = normalizeValue(url.searchParams.get("source"), "site");
  const platform = normalizeValue(url.searchParams.get("platform"), "macos-dmg");
  const referrer = request.headers.get("referer") || "";

  await trackGoogleAnalyticsEvent({
    request,
    name: "download_redirect",
    params: {
      download_source: source,
      download_platform: platform,
      download_target: "github_release_dmg",
      referrer_url: referrer
    }
  });

  const response = NextResponse.redirect(releaseDownloadUrl, 307);
  response.headers.set("Cache-Control", "no-store, max-age=0");
  return response;
}
