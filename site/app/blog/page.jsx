import { SiteFooter } from "../../components/site-footer";
import { SiteHeader } from "../../components/site-header";
import { blogPosts } from "../../data/site-content";

export const metadata = {
  title: "Blog",
  description: "Engineering notes and product updates from OpenMessage."
};

export default function BlogIndexPage() {
  return (
    <main className="relative z-[1] min-h-screen">
      <SiteHeader />

      <section className="border-b border-[var(--border)]">
        <div className="mx-auto max-w-[1180px] px-6 pb-16 pt-24 lg:px-10 lg:pb-20">
          <div className="eyebrow">Blog</div>
          <h1 className="mt-5 max-w-[34rem] text-[clamp(2.4rem,5vw,4.2rem)] font-semibold leading-[0.94] tracking-[-0.06em] text-[var(--text-primary)]">
            Technical notes on local-first messaging.
          </h1>
          <p className="mt-6 max-w-[38rem] text-lg leading-8 text-[var(--text-secondary)]">
            The product story stays simple. The engineering story lives here.
          </p>
        </div>
      </section>

      <section className="mx-auto max-w-[1180px] px-6 py-16 lg:px-10">
        <div className="grid gap-6">
          {blogPosts.map((post) => (
            <a
              key={post.slug}
              href={`/blog/${post.slug}`}
              className="group rounded-[1.8rem] border border-[var(--border)] bg-[color:rgba(9,17,29,0.7)] px-7 py-7 transition-colors hover:border-[var(--border-strong)] hover:bg-[var(--bg-hover)]"
            >
              <div className="text-[0.72rem] font-semibold uppercase tracking-[0.22em] text-[var(--accent-strong)]">
                {post.eyebrow}
              </div>
              <h2 className="mt-4 text-[1.8rem] font-semibold tracking-[-0.05em] text-[var(--text-primary)]">
                {post.title}
              </h2>
              <p className="mt-3 max-w-[44rem] text-base leading-7 text-[var(--text-secondary)]">
                {post.description}
              </p>
              <div className="mt-6 text-sm font-semibold text-[var(--text-primary)]">
                Read post
              </div>
            </a>
          ))}
        </div>
      </section>

      <SiteFooter />
    </main>
  );
}
