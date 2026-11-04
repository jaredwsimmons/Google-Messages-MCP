import { SiteFooter } from "../../components/site-footer";
import { SiteHeader } from "../../components/site-header";
import {
  competitionRows,
  expansionRows,
  thesisStats
} from "../../data/site-content";

export const metadata = {
  title: "Thesis",
  robots: {
    index: false,
    follow: false
  }
};

export default function ThesisPage() {
  return (
    <main className="min-h-screen">
      <SiteHeader compact />

      <section className="mx-auto max-w-[920px] px-6 py-18 lg:px-10">
        <div className="article-copy">
          <div className="eyebrow">Internal thesis</div>
          <h1 className="mt-6 font-semibold tracking-[-0.065em]">OpenMessage is the local-first messaging layer for AI.</h1>
          <p className="max-w-[44rem] text-lg leading-8">
            The market already proved demand for unified messaging. The gap now is open,
            local-first, AI-native messaging infrastructure that people can actually trust.
          </p>

          <div className="mt-8 inline-flex rounded-full border border-[var(--border)] bg-[color:rgba(13,23,40,0.8)] px-4 py-2 text-sm text-[var(--text-muted)]">
            Not for public distribution
          </div>

          <h2>The opportunity</h2>
          <p>
            Messaging is still fragmented across consumer and work networks. Unified messaging has
            proven value, but the most visible products remain cloud-dependent, closed-source, and
            weak on local developer access.
          </p>
          <p>
            At the same time, assistants are shifting toward acting on local tools through MCP and
            similar protocols. Messaging is one of the highest-value surfaces those assistants still
            cannot operate in safely and locally.
          </p>

          <div className="mt-8 grid gap-4 md:grid-cols-3">
            {thesisStats.map((stat) => (
              <div
                key={stat.label}
                className="rounded-[1.75rem] border border-[var(--border)] bg-[color:rgba(13,23,40,0.74)] p-5"
              >
                <div className="text-4xl font-semibold tracking-[-0.07em] text-[var(--accent-strong)]">
                  {stat.value}
                </div>
                <div className="mt-2 text-sm leading-6 text-[var(--text-secondary)]">{stat.label}</div>
              </div>
            ))}
          </div>

          <h2>What OpenMessage is</h2>
          <p>
            OpenMessage is a <strong>local-first, open-source messaging client</strong> with a built-in
            AI control layer. Today it supports Google Messages, live WhatsApp, and Signal, ships as a native
            macOS app or standalone Go runtime, and exposes the local workspace through MCP.
          </p>
          <ul>
            <li><strong>Local-first:</strong> message history, search, and sessions live on the user&apos;s machine.</li>
            <li><strong>AI-native:</strong> the same runtime powers both the UI and the local tool surface for assistants.</li>
            <li><strong>Open source:</strong> auditable code, extensible integrations, and low trust overhead.</li>
            <li><strong>Native where it matters:</strong> Swift wrapper on macOS instead of Electron, plus a plain Go runtime for headless or Linux usage.</li>
          </ul>

          <h2>Expansion path</h2>
          <p>
            The product gets stronger each time a new route plugs into the same thread workspace and
            local MCP surface.
          </p>

          <table>
            <thead>
              <tr>
                <th>Service</th>
                <th>Bridge</th>
                <th>Status</th>
                <th>Why it matters</th>
              </tr>
            </thead>
            <tbody>
              {expansionRows.map((row) => (
                <tr key={row[0]}>
                  {row.map((cell, index) => (
                    <td key={`${row[0]}-${index}`}>{cell}</td>
                  ))}
                </tr>
              ))}
            </tbody>
          </table>

          <h2>Business model</h2>
          <p>
            The likely model is open-core distribution with premium AI and team features on top of
            the local client.
          </p>
          <ul>
            <li><strong>Free core:</strong> local messaging, local search, route management, and MCP access.</li>
            <li><strong>Premium:</strong> hosted AI features, better cross-thread summarization, and managed credits.</li>
            <li><strong>Team / enterprise:</strong> shared context, admin controls, compliance features, and secure deployment support.</li>
          </ul>

          <h2>Competitive landscape</h2>
          <table>
            <thead>
              <tr>
                <th>Product</th>
                <th>Open source</th>
                <th>Local-first</th>
                <th>AI / MCP</th>
                <th>Price</th>
              </tr>
            </thead>
            <tbody>
              {competitionRows.map((row) => (
                <tr key={row[0]}>
                  {row.map((cell, index) => (
                    <td key={`${row[0]}-${index}`}>{cell}</td>
                  ))}
                </tr>
              ))}
            </tbody>
          </table>

          <h2>Why now</h2>
          <ul>
            <li><strong>MCP adoption:</strong> local tool protocols are becoming a default integration path for assistants.</li>
            <li><strong>Messaging proved valuable:</strong> Beeper and Texts.com validated the category, but not the local-AI angle.</li>
            <li><strong>Bridge ecosystems are maturing:</strong> the hard protocol work is increasingly reusable instead of greenfield.</li>
            <li><strong>User trust is scarce:</strong> local-first is an increasingly defensible product choice in messaging.</li>
          </ul>

          <h2>Current traction</h2>
          <ul>
            <li>Native macOS app and local web runtime shipping from the same codebase.</li>
            <li>Google Messages, WhatsApp, and Signal support in the current product.</li>
            <li>Local MCP server, diagnostics, search, notifications, and route-aware threading.</li>
            <li>Open-source distribution and a live product site at openmessage.ai.</li>
          </ul>

          <h2>Ask</h2>
          <p>
            Looking for collaborators, contributors, and people who care about local AI, messaging
            infrastructure, and developer-facing product surfaces. Reach out at{" "}
            <a href="mailto:max@maxghenis.com">max@maxghenis.com</a>.
          </p>
        </div>
      </section>

      <SiteFooter />
    </main>
  );
}
