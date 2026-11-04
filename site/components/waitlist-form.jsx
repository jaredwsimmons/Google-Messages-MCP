"use client";

import { useState } from "react";

const interestOptions = [
  { value: "", label: "What are you interested in? (optional)" },
  { value: "mac-app", label: "Mac app" },
  { value: "whatsapp", label: "WhatsApp" },
  { value: "signal", label: "Signal" },
  { value: "mcp", label: "MCP" },
  { value: "general", label: "General updates" }
];

export function WaitlistForm({ compact = false }) {
  const [email, setEmail] = useState("");
  const [interest, setInterest] = useState("");
  const [website, setWebsite] = useState("");
  const [status, setStatus] = useState("idle");
  const [message, setMessage] = useState("");

  async function handleSubmit(event) {
    event.preventDefault();

    if (status === "submitting") {
      return;
    }

    setStatus("submitting");
    setMessage("");

    try {
      const response = await fetch("/api/waitlist", {
        method: "POST",
        headers: {
          "Content-Type": "application/json"
        },
        body: JSON.stringify({ email, interest, website })
      });

      const payload = await response.json().catch(() => ({}));

      if (!response.ok) {
        throw new Error(payload?.error || "Could not save your email right now.");
      }

      setStatus("success");
      setMessage(payload?.message || "Thanks. I’ll send product updates when there’s something real to share.");
      setEmail("");
      setInterest("");
      setWebsite("");
    } catch (error) {
      setStatus("error");
      setMessage(error.message || "Could not save your email right now.");
    }
  }

  return (
    <div
      className={
        compact
          ? "rounded-[1.6rem] border border-[var(--border)] bg-[color:rgba(9,17,29,0.74)] p-5"
          : "rounded-[2rem] border border-[var(--border)] bg-[color:rgba(9,17,29,0.74)] p-6 shadow-[var(--panel-shadow)]"
      }
    >
      <div className="text-[0.72rem] font-semibold uppercase tracking-[0.22em] text-[var(--accent-strong)]">
        Get updates
      </div>
      <h3 className="mt-3 text-[1.5rem] font-semibold tracking-[-0.05em] text-[var(--text-primary)]">
        Join the early tester list.
      </h3>
      <p className="mt-3 max-w-[30rem] text-sm leading-6 text-[var(--text-secondary)]">
        Low-volume emails when there is something real to try: new platform support, App Store review status, or meaningful product updates.
      </p>

      <form className="mt-5 grid gap-3" onSubmit={handleSubmit}>
        <label className="sr-only" htmlFor={compact ? "waitlist-email-compact" : "waitlist-email"}>
          Email address
        </label>
        <input
          id={compact ? "waitlist-email-compact" : "waitlist-email"}
          type="email"
          required
          autoComplete="email"
          placeholder="you@example.com"
          value={email}
          onChange={(event) => setEmail(event.target.value)}
          className="w-full rounded-[1.15rem] border border-[var(--border)] bg-[color:rgba(7,13,24,0.82)] px-4 py-3 text-sm text-[var(--text-primary)] outline-none transition-colors placeholder:text-[var(--text-muted)] focus:border-[var(--border-strong)]"
        />
        <label className="sr-only" htmlFor={compact ? "waitlist-interest-compact" : "waitlist-interest"}>
          Interest
        </label>
        <select
          id={compact ? "waitlist-interest-compact" : "waitlist-interest"}
          value={interest}
          onChange={(event) => setInterest(event.target.value)}
          className="w-full rounded-[1.15rem] border border-[var(--border)] bg-[color:rgba(7,13,24,0.82)] px-4 py-3 text-sm text-[var(--text-primary)] outline-none transition-colors focus:border-[var(--border-strong)]"
        >
          {interestOptions.map((option) => (
            <option key={option.label} value={option.value}>
              {option.label}
            </option>
          ))}
        </select>
        <div className="hidden">
          <label htmlFor={compact ? "waitlist-website-compact" : "waitlist-website"}>Website</label>
          <input
            id={compact ? "waitlist-website-compact" : "waitlist-website"}
            type="text"
            tabIndex={-1}
            autoComplete="off"
            value={website}
            onChange={(event) => setWebsite(event.target.value)}
          />
        </div>
        <button
          type="submit"
          disabled={status === "submitting"}
          className="inline-flex items-center justify-center rounded-full bg-[var(--accent)] px-5 py-3 text-sm font-semibold text-[var(--bg-deep)] transition-transform hover:-translate-y-0.5 disabled:cursor-wait disabled:opacity-70"
        >
          {status === "submitting" ? "Saving…" : "Get updates"}
        </button>
      </form>

      <p
        className={`mt-3 text-sm leading-6 ${status === "error" ? "text-[#ffb4c1]" : "text-[var(--text-muted)]"}`}
        aria-live="polite"
      >
        {message || "No spam. Just product updates and tester invites."}
      </p>
    </div>
  );
}
