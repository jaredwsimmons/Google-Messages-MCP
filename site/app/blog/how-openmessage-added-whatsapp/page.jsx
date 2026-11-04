import { SiteFooter } from "../../../components/site-footer";
import { SiteHeader } from "../../../components/site-header";

export const metadata = {
  title: "How OpenMessage added live WhatsApp support",
  description:
    "A technical note on the local bridge, linked-device model, and shared runtime behind WhatsApp in OpenMessage."
};

export default function WhatsAppPostPage() {
  return (
    <main className="relative z-[1] min-h-screen">
      <SiteHeader />

      <section className="border-b border-[var(--border)]">
        <div className="mx-auto max-w-[900px] px-6 pb-14 pt-24 lg:px-10 lg:pb-18">
          <div className="eyebrow">Engineering note</div>
          <h1 className="mt-5 max-w-[42rem] text-[clamp(2.5rem,5vw,4.3rem)] font-semibold leading-[0.94] tracking-[-0.06em] text-[var(--text-primary)]">
            How OpenMessage added live WhatsApp support
          </h1>
          <p className="mt-6 max-w-[42rem] text-lg leading-8 text-[var(--text-secondary)]">
            The short version is simple: WhatsApp became a real local route, not a slow
            desktop import.
          </p>
        </div>
      </section>

      <article className="mx-auto max-w-[900px] px-6 py-16 lg:px-10">
        <div className="article-copy">
          <p>
            OpenMessage started with Google Messages because Android already has a clear
            linked-device model. WhatsApp needed to feel equally native inside the same inbox,
            or it would just become another side tab. So the goal was not &ldquo;show WhatsApp
            history eventually.&rdquo; The goal was a live route: new messages, typing, media,
            receipts, search, and replies flowing through the same local runtime as SMS and
            RCS.
          </p>

          <h2>Why a live bridge instead of a periodic importer</h2>
          <p>
            A periodic desktop import can make an archive visible, but it does not feel like a
            messaging product. Messages arrive late. Typing indicators do not exist. Receipts
            lag. Media is brittle. The UI starts to look unified while the product behavior is
            still fragmented.
          </p>
          <p>
            That is why WhatsApp in OpenMessage now runs as a live linked-device session on the
            user&rsquo;s own machine. The point was not just protocol coverage. The point was to
            make the route behave like a first-class part of the app.
          </p>

          <h2>The actual shape of the system</h2>
          <p>
            The runtime is still local-first. A Go backend owns the message store, transport
            sessions, search, and MCP surface. The macOS wrapper is mostly there to package the
            local runtime cleanly with desktop integration like notifications and contact
            access.
          </p>
          <p>
            WhatsApp support plugs into that same backend as a live bridge. In practical terms:
          </p>
          <ul>
            <li>A linked-device WhatsApp session is paired on the local machine.</li>
            <li>Inbound events normalize into the same local conversation and message schema as other routes.</li>
            <li>The web UI and native wrapper update from the same invalidation stream.</li>
            <li>MCP clients connect to the same local state, not a second AI-only cache.</li>
          </ul>

          <h2>What &ldquo;local-first&rdquo; means here</h2>
          <p>
            OpenMessage is not trying to become a hosted messaging relay. The desktop app, the
            search index, the diagnostics, and the transport sessions are designed to live on
            the user&rsquo;s machine. That matters both for privacy and for product feel. The app
            can be responsive because the source of truth is close to the UI.
          </p>

          <h2>What shipped with WhatsApp</h2>
          <p>
            The first bar was not just text send and receive. WhatsApp support only felt worth
            shipping once it handled the parts that make a route feel real inside a desktop
            workspace:
          </p>
          <ul>
            <li>live inbound and outbound messages</li>
            <li>media send and render</li>
            <li>voice notes</li>
            <li>typing indicators</li>
            <li>read-state rendering</li>
            <li>desktop notifications</li>
            <li>grouped route selection beside SMS and RCS</li>
            <li>assistant access through the built-in MCP runtime</li>
          </ul>

          <h2>Why the inbox model matters</h2>
          <p>
            The UI does not pretend every transport is identical. OpenMessage keeps route
            selection explicit when it matters, but still groups the experience by person where
            it helps. That is why one contact can clearly show both SMS and WhatsApp without the
            product collapsing into two separate apps jammed into one window.
          </p>

          <h2>The implementation stack</h2>
          <p>
            The live WhatsApp path is built into the OpenMessage runtime rather than bolted on
            as a remote service. Under the hood, the bridge uses the WhatsApp linked-device
            model and is integrated into the same Go backend that powers search, diagnostics,
            the local web app, and MCP. The result is that features like route-aware drafting,
            local notifications, and media rendering do not need a second architecture just for
            WhatsApp.
          </p>

          <h2>What comes next</h2>
          <p>
            Shipping WhatsApp changed OpenMessage from &ldquo;Google Messages with MCP&rdquo; into
            a broader local messaging product. The next work is less about adding another flashy
            transport immediately and more about trust: pairing quality, reconnect behavior,
            diagnostics, and making the local runtime easy to understand when something breaks.
          </p>
          <p>
            That is the real shape of the product now: one local message surface, multiple real
            routes, and an assistant that can operate on the same inbox you use yourself.
          </p>
        </div>
      </article>

      <SiteFooter />
    </main>
  );
}
