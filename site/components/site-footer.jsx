import { repoUrl } from "../data/site-content";

export function SiteFooter() {
  return (
    <footer className="border-t border-[var(--border)]">
      <div className="mx-auto flex max-w-[1520px] flex-col gap-4 px-6 py-8 text-sm text-[var(--text-muted)] md:flex-row md:items-center md:justify-between lg:px-10">
        <p>
          A solo project by{" "}
          <a
            className="text-[var(--text-secondary)] transition-colors hover:text-[var(--text-primary)]"
            href="https://maxghenis.com"
            target="_blank"
            rel="noreferrer"
          >
            Max Ghenis
          </a>
          . Free and open source under the MIT license.
        </p>
        <div className="flex flex-wrap items-center gap-4">
          <a className="transition-colors hover:text-[var(--text-primary)]" href="/#updates">
            Get updates
          </a>
          <a className="transition-colors hover:text-[var(--text-primary)]" href="/blog">
            Blog
          </a>
          <a className="transition-colors hover:text-[var(--text-primary)]" href="/privacy">
            Privacy
          </a>
          <a className="transition-colors hover:text-[var(--text-primary)]" href="/thesis">
            Thesis
          </a>
          <a
            className="transition-colors hover:text-[var(--text-primary)]"
            href={repoUrl}
            target="_blank"
            rel="noreferrer"
          >
            GitHub
          </a>
          <a
            className="transition-colors hover:text-[var(--text-primary)]"
            href="https://github.com/sponsors/MaxGhenis"
            target="_blank"
            rel="noreferrer"
          >
            Sponsor
          </a>
        </div>
      </div>
    </footer>
  );
}
