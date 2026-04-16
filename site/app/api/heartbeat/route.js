// Anonymous heartbeat receiver.
//
// The OpenMessage app sends one POST per install per 24h with:
//   { install_id, version, os, arch, platforms, sent_at, schema_version }
//
// We only persist counts. Install IDs are written to Vercel Blob bucketed by
// day so we can compute DAU without retaining individual identifiers long-term.
//
// Required env vars:
//   BLOB_READ_WRITE_TOKEN   — Vercel Blob token (already set for waitlist)
//
// Telemetry is opt-in client-side. This endpoint accepts whatever is sent;
// rate-limiting + per-install dedup is the client's responsibility.

import { put } from "@vercel/blob";
import { NextResponse } from "next/server";

const SCHEMA_VERSION = 1;

function safeString(value, max = 64) {
  return `${value || ""}`.replace(/[^A-Za-z0-9._-]/g, "").slice(0, max);
}

export async function POST(request) {
  let payload;
  try {
    payload = await request.json();
  } catch {
    return NextResponse.json({ ok: false }, { status: 400 });
  }

  if (!payload || payload.schema_version !== SCHEMA_VERSION) {
    return NextResponse.json({ ok: false }, { status: 400 });
  }

  const installId = safeString(payload.install_id, 64);
  if (installId.length < 16) {
    return NextResponse.json({ ok: false }, { status: 400 });
  }

  const day = new Date().toISOString().slice(0, 10);
  const record = {
    install_id: installId,
    version: safeString(payload.version, 32),
    os: safeString(payload.os, 16),
    arch: safeString(payload.arch, 16),
    platforms: {
      google_messages: !!payload?.platforms?.google_messages,
      whatsapp: !!payload?.platforms?.whatsapp,
      signal: !!payload?.platforms?.signal,
    },
    received_at: new Date().toISOString(),
  };

  if (!process.env.BLOB_READ_WRITE_TOKEN) {
    // No-op if not configured; respond OK so the app doesn't retry.
    return NextResponse.json({ ok: true, persisted: false });
  }

  try {
    await put(
      `heartbeats/${day}/${installId}.json`,
      JSON.stringify(record),
      {
        access: "public",
        addRandomSuffix: false,
        contentType: "application/json",
      },
    );
  } catch (err) {
    console.error("heartbeat persist failed:", err);
    return NextResponse.json({ ok: false }, { status: 502 });
  }

  return NextResponse.json({ ok: true });
}

export async function GET() {
  return NextResponse.json({ ok: true, schema_version: SCHEMA_VERSION });
}
