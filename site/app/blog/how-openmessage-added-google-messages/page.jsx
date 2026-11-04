import { SiteFooter } from "../../../components/site-footer";
import { SiteHeader } from "../../../components/site-header";

export const metadata = {
  title: "How OpenMessage added Google Messages",
  description:
    "A technical note on pairing, live sync, and the shared local runtime behind Google Messages in OpenMessage."
};

export default function GoogleMessagesPostPage() {
  return (
    <main className="relative z-[1] min-h-screen">
      <SiteHeader />

      <section className="border-b border-[var(--border)]">
        <div className="mx-auto max-w-[900px] px-6 pb-14 pt-24 lg:px-10 lg:pb-18">
          <div className="eyebrow">Engineering note</div>
          <h1 className="mt-5 max-w-[42rem] text-[clamp(2.5rem,5vw,4.3rem)] font-semibold leading-[0.94] tracking-[-0.06em] text-[var(--text-primary)]">
            How OpenMessage added Google Messages
          </h1>
          <p className="mt-6 max-w-[42rem] text-lg leading-8 text-[var(--text-secondary)]">
            Google Messages was the first real route in OpenMessage, so the product architecture
            was built around making that connection feel local, live, and reliable.
          </p>
        </div>
      </section>

      <article className="mx-auto max-w-[900px] px-6 py-16 lg:px-10">
        <div className="article-copy">
          <p>
            OpenMessage did not start as a generic chat shell. It started from one concrete
            question: what would it look like if your own phone-backed messages were available
            in a local desktop app and in MCP, without turning into a hosted messaging service?
            Google Messages was the first route that made that idea tractable.
          </p>

          <h2>Why Google Messages came first</h2>
          <p>
            The route already has a real linked-device model, and a large share of desktop
            messaging demand is still plain SMS and RCS. That made Google Messages the right
            place to prove the product. If OpenMessage could pair a real phone, keep a local
            inbox current, and let an assistant operate on the same state, the rest of the
            architecture would stop being hypothetical.
          </p>

          <h2>The pairing model</h2>
          <p>
            OpenMessage pairs against the user&rsquo;s own Google Messages setup rather than
            asking the user to create a second messaging account. In practice that means the app
            can connect through the same pairing flow people already understand from Google
            Messages for Web, while still keeping the OpenMessage runtime local-first.
          </p>
          <p>
            The important product constraint is that pairing should lead directly into a usable
            desktop workspace, not into an account system of our own. That is why the local app,
            the CLI, and the MCP runtime all converge on the same paired session and message
            store.
          </p>

          <h2>The live event path</h2>
          <p>
            A messaging route only feels real if sends and receives show up immediately. So
            Google Messages in OpenMessage is not just a periodic sync job. The backend keeps a
            live client session, listens for new events, writes them into the local database,
            and publishes UI invalidations so the thread list, active thread, typing state, and
            notifications update without waiting for a manual refresh.
          </p>

          <h2>What lands in the shared local store</h2>
          <p>
            The local database is not a Google-only cache. It is the shared inbox model that the
            rest of the product builds on. Google Messages normalizes into the same core
            conversation and message schema that WhatsApp now uses too. That is why grouped
            contacts, unified search, MCP tooling, route-aware replies, diagnostics, and desktop
            notifications can all work across multiple transports without a second architecture.
          </p>

          <h2>Why reliability work mattered so much</h2>
          <p>
            The first version of a route is never the hard part. The hard part is everything
            that happens after the happy path: reconnects, stale sessions, duplicate optimistic
            sends, missed listener windows, deep backfills, UI invalidation drops, and all the
            ways a desktop app can make the user think a message sent when it did not.
          </p>
          <p>
            A large amount of OpenMessage&rsquo;s recent hardening work came directly from making
            Google Messages trustworthy as a real daily driver rather than a demo. That includes
            stricter send-path error handling, better reconnect behavior, catch-up sync after
            interruptions, browser and macOS notification fixes, and regression coverage around
            the thread UI.
          </p>

          <h2>How MCP fits into the same runtime</h2>
          <p>
            One of the points of OpenMessage is that the assistant is not operating on a fake
            export or a second cloud mirror. The same local backend that pairs Google Messages
            also exposes the MCP surface. So when Claude searches, drafts, or sends, it is doing
            that against the same local state you see in the desktop app.
          </p>

          <h2>What Google Messages support means now</h2>
          <p>
            Today Google Messages is not just the original route. It is the reliability anchor
            for the whole product. It proves the local-first model, the paired-device model, the
            shared inbox, and the idea that messaging plus MCP can live in one desktop runtime
            without requiring a new hosted service in the middle.
          </p>

          <h2>Why this mattered before adding more routes</h2>
          <p>
            WhatsApp made OpenMessage much more ambitious, but Google Messages is where the core
            discipline was built. If the first route had stayed fragile, adding more transports
            would have just multiplied instability. Instead, Google Messages established the
            basic contract: pair locally, update live, store locally, expose the same state to
            the UI and MCP, and recover cleanly when the real world gets messy.
          </p>
          <p>
            That contract is the real product. New transports only matter if they fit into it.
          </p>
        </div>
      </article>

      <SiteFooter />
    </main>
  );
}
