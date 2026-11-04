import { createCipheriv, randomBytes } from "node:crypto";
import { put } from "@vercel/blob";
import { NextResponse } from "next/server";

const allowedInterests = new Set(["", "mac-app", "whatsapp", "signal", "mcp", "general"]);

function isValidEmail(value) {
  return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(value);
}

export async function POST(request) {
  let payload;

  try {
    payload = await request.json();
  } catch {
    return NextResponse.json({ error: "Invalid request body." }, { status: 400 });
  }

  const email = `${payload?.email || ""}`.trim().toLowerCase();
  const interest = `${payload?.interest || ""}`.trim();
  const website = `${payload?.website || ""}`.trim();
  const blobToken = process.env.BLOB_READ_WRITE_TOKEN;
  const encryptionKey = process.env.WAITLIST_ENCRYPTION_KEY;

  if (website) {
    return NextResponse.json({ ok: true, message: "Thanks." });
  }

  if (!isValidEmail(email)) {
    return NextResponse.json({ error: "Enter a valid email address." }, { status: 400 });
  }

  if (!allowedInterests.has(interest)) {
    return NextResponse.json({ error: "Invalid interest selection." }, { status: 400 });
  }

  if (!blobToken || !encryptionKey) {
    console.error("Waitlist storage is not configured.");
    return NextResponse.json(
      { error: "Updates are not configured right now. Email max@maxghenis.com instead." },
      { status: 503 }
    );
  }

  try {
    const submittedAt = new Date().toISOString();
    const pathname = `waitlist/${submittedAt.slice(0, 10)}/${crypto.randomUUID()}.json`;
    const iv = randomBytes(12);
    const key = Buffer.from(encryptionKey, "hex");

    if (key.length !== 32) {
      throw new Error("WAITLIST_ENCRYPTION_KEY must be 32 bytes (64 hex chars).");
    }

    const cipher = createCipheriv("aes-256-gcm", key, iv);
    const plaintext = JSON.stringify({
      email,
      interest: interest || "general",
      submittedAt,
      source: "openmessage.ai"
    });
    const ciphertext = Buffer.concat([cipher.update(plaintext, "utf8"), cipher.final()]);
    const tag = cipher.getAuthTag();

    await put(
      pathname,
      JSON.stringify(
        {
          version: 1,
          algorithm: "aes-256-gcm",
          submittedAt,
          iv: iv.toString("base64"),
          tag: tag.toString("base64"),
          ciphertext: ciphertext.toString("base64")
        },
        null,
        2
      ),
      {
        access: "public",
        addRandomSuffix: false,
        contentType: "application/json"
      }
    );
  } catch (error) {
    console.error("Waitlist submit failed:", error);
    return NextResponse.json(
      { error: "Could not save your email right now. Try again shortly." },
      { status: 502 }
    );
  }

  return NextResponse.json({
    ok: true,
    message: "Thanks. I’ll send product updates when there’s something real to share."
  });
}
