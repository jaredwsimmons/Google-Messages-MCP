import { SiteFooter } from "../../components/site-footer";
import { SiteHeader } from "../../components/site-header";

export const metadata = {
  title: "Privacy"
};

export default function PrivacyPage() {
  return (
    <main className="min-h-screen">
      <SiteHeader compact />

      <section className="mx-auto max-w-[860px] px-6 py-18 lg:px-10">
        <div className="article-copy">
          <div className="eyebrow">Privacy policy</div>
          <h1 className="mt-6 font-semibold tracking-[-0.065em]">Everything important stays on your machine.</h1>
          <p className="max-w-[42rem] text-lg leading-8">
            OpenMessage is built as a local-first messaging client. It does not require an
            OpenMessage account and does not route your message store through
            OpenMessage-operated servers. The marketing site may collect basic website analytics
            if enabled, but that is separate from the product and never includes message content.
          </p>

          <div className="mt-8 inline-flex rounded-full border border-[var(--border)] bg-[color:rgba(13,23,40,0.8)] px-4 py-2 text-sm text-[var(--text-muted)]">
            Last updated: April 11, 2026
          </div>

          <h2>Summary</h2>
          <p>
            <strong>OpenMessage does not collect, transmit, or retain your message data on our
            infrastructure.</strong> Message history, contacts, local search indexes, and transport
            sessions stay on your device unless you explicitly forward content elsewhere.
          </p>

          <h2>What data OpenMessage accesses</h2>
          <ul>
            <li>Your message history, conversations, media, and metadata for connected transports such as Google Messages, WhatsApp, and Signal.</li>
            <li>Contact names, numbers, and optional local contact photos if you grant macOS Contacts access.</li>
            <li>Local session credentials needed to keep Google Messages, WhatsApp, and Signal paired.</li>
            <li>Public URL metadata when link previews are enabled and a conversation contains a public link.</li>
          </ul>

          <h2>Where data is stored</h2>
          <ul>
            <li><strong>Messages and contacts:</strong> local SQLite databases in the app&apos;s Application Support directory.</li>
            <li><strong>Transport sessions:</strong> local session files and local bridge stores used for reconnecting paired services.</li>
            <li><strong>Media and previews:</strong> cached locally when downloaded or derived on your machine.</li>
          </ul>

          <h2>What data is transmitted</h2>
          <ul>
            <li><strong>To messaging providers:</strong> the minimum traffic required to send, receive, and pair with the connected services, using the same network surfaces those services already expose.</li>
            <li><strong>To OpenMessage:</strong> nothing for message sync. We do not operate an OpenMessage cloud backend for your message history, contacts, or transport sessions.</li>
            <li><strong>To AI tools:</strong> only when you choose to connect a local MCP client or send content to an external model provider from your own machine.</li>
            <li><strong>For link previews:</strong> preview metadata is fetched directly from your device to the public URL. Private and localhost-style addresses are intentionally refused.</li>
            <li><strong>For product updates:</strong> if you submit your email through the website&apos;s updates form, that address and your optional interest are stored as encrypted signup records in the site&apos;s Vercel project so OpenMessage can send product updates or tester invites.</li>
            <li><strong>For website analytics:</strong> if Google Analytics is enabled on openmessage.ai, the marketing site may send pageview, referral, browser/device, and download-redirect events to Google. This applies to the website only, not the local messaging runtime.</li>
          </ul>

          <h2>Third-party services</h2>
          <p>
            OpenMessage talks to the messaging networks you choose to connect, such as Google
            Messages, WhatsApp, and Signal. Their privacy policies still govern the underlying messaging
            services themselves. OpenMessage does not add a separate OpenMessage-hosted data layer on top.
          </p>

          <h2>Notifications and diagnostics</h2>
          <p>
            Native notifications are generated locally on your device. Diagnostics and issue-report
            exports are generated locally too; you decide whether to copy or share them.
          </p>

          <h2>Data deletion</h2>
          <p>
            Removing OpenMessage removes the app. To delete locally stored data, remove the app&apos;s
            Application Support directory and any local bridge session stores. Your messages remain
            with the underlying messaging services regardless of OpenMessage.
          </p>

          <h2>Changes to this policy</h2>
          <p>
            If this policy changes, the updated version will be posted here with a new date.
            OpenMessage is open source, so the codebase remains the best canonical reference for
            what the app actually does.
          </p>

          <h2>Contact</h2>
          <p>
            Questions about privacy or data handling can go to{" "}
            <a href="mailto:max@maxghenis.com">max@maxghenis.com</a> or{" "}
            <a href="https://github.com/MaxGhenis/openmessage/issues" target="_blank" rel="noreferrer">
              the GitHub issue tracker
            </a>
            .
          </p>
        </div>
      </section>

      <SiteFooter />
    </main>
  );
}
