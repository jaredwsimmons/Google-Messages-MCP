"use client";

import { useState } from "react";

export function CommandBlock({ children, label }) {
  const [copied, setCopied] = useState(false);

  const text = typeof children === "string" ? children : String(children);

  const handleCopy = async () => {
    try {
      await navigator.clipboard.writeText(text);
      setCopied(true);
      setTimeout(() => setCopied(false), 1600);
    } catch {
      // ignore
    }
  };

  return (
    <div className="group relative min-w-0 rounded-3xl border border-[var(--border)] bg-[color:rgba(13,23,40,0.78)] p-4 shadow-[0_18px_60px_rgba(4,12,24,0.24)]">
      {label ? (
        <div className="mb-3 flex items-center justify-between gap-3">
          <div className="text-[0.68rem] font-semibold uppercase tracking-[0.24em] text-[var(--text-muted)]">
            {label}
          </div>
        </div>
      ) : null}
      <div className="flex items-start gap-3">
        <pre className="min-w-0 flex-1 overflow-x-auto whitespace-pre-wrap break-all text-sm leading-7 text-[var(--text-primary)] sm:whitespace-pre sm:break-normal">
          <code>{children}</code>
        </pre>
        <button
          type="button"
          onClick={handleCopy}
          aria-label={copied ? "Copied" : "Copy to clipboard"}
          className="flex-shrink-0 rounded-lg border border-[var(--border)] bg-[color:rgba(8,13,24,0.5)] px-2.5 py-1.5 text-[0.7rem] font-medium uppercase tracking-[0.12em] text-[var(--text-secondary)] opacity-0 transition-all hover:border-[var(--border-strong)] hover:bg-[color:rgba(20,33,57,0.7)] hover:text-[var(--text-primary)] focus-visible:opacity-100 group-hover:opacity-100"
        >
          {copied ? "Copied" : "Copy"}
        </button>
      </div>
    </div>
  );
}
