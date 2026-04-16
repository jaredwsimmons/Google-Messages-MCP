import { buildDownloadUrl } from "../data/site-content";

const navItems = [
  { href: "/#features", label: "Product" },
  { href: "/#how-it-works", label: "How it works" },
  { href: "/#setup", label: "Setup" },
  { href: "/blog", label: "Blog" }
];

export function SiteHeader({ compact = false, overlay = false }) {
  const downloadMinWidthClassName = compact ? "min-w-[112px]" : "min-w-[132px]";
  const headerClassName = overlay
    ? "absolute inset-x-0 top-0 z-50"
    : "sticky top-0 z-50 border-b border-[var(--border)] bg-[color:rgba(9,17,28,0.82)] backdrop-blur-xl";

  return (
    <header className={headerClassName}>
      <div className="mx-auto flex max-w-[1520px] items-center justify-between gap-6 px-6 py-5 lg:px-10">
        <a
          href="/"
          className="text-[1.45rem] font-semibold tracking-[-0.05em] text-[var(--text-primary)] transition-colors hover:text-[var(--accent-strong)]"
        >
          OpenMessage
        </a>
        <div className="flex items-center gap-3 md:gap-5">
          <nav className="hidden items-center gap-6 text-sm text-[var(--text-secondary)] md:flex">
            {navItems.map((item) => (
              <a
                key={item.href}
                href={item.href}
                className="transition-colors hover:text-[var(--text-primary)]"
              >
                {item.label}
              </a>
            ))}
          </nav>
          <a
            href={buildDownloadUrl("site_header")}
            className={`inline-flex items-center justify-center rounded-full border border-[var(--border-strong)] ${overlay ? "bg-[color:rgba(8,13,24,0.58)] text-[var(--text-primary)] backdrop-blur-xl hover:bg-[color:rgba(14,22,37,0.78)]" : "bg-[var(--accent)] text-[var(--bg-deep)]"} px-4 py-2 text-sm font-medium transition-transform hover:-translate-y-0.5 ${downloadMinWidthClassName}`}
          >
            Download
          </a>
        </div>
      </div>
    </header>
  );
}
