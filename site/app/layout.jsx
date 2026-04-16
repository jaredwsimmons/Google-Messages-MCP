import { IBM_Plex_Mono, Public_Sans } from "next/font/google";

import { GoogleAnalytics } from "../components/google-analytics";
import { getGoogleAnalyticsMeasurementId } from "../lib/analytics";
import "./globals.css";

const sans = Public_Sans({
  subsets: ["latin"],
  variable: "--font-sans",
  display: "swap"
});

const mono = IBM_Plex_Mono({
  subsets: ["latin"],
  weight: ["400", "500"],
  variable: "--font-mono",
  display: "swap"
});

export const metadata = {
  metadataBase: new URL("https://openmessage.ai"),
  title: {
    default: "OpenMessage",
    template: "%s | OpenMessage"
  },
  description:
    "Local-first messaging for Google Messages, WhatsApp, and Signal, with an AI-native control layer and local MCP access.",
  openGraph: {
    title: "OpenMessage",
    description:
      "One local workspace for Google Messages, WhatsApp, Signal, and AI-assisted messaging.",
    url: "https://openmessage.ai",
    siteName: "OpenMessage",
    images: [
      {
        url: "/hero-product-dark.png",
        width: 1600,
        height: 1100,
        alt: "OpenMessage desktop workspace"
      }
    ],
    locale: "en_US",
    type: "website"
  },
  twitter: {
    card: "summary_large_image",
    title: "OpenMessage",
    description:
      "Local-first messaging for Google Messages, WhatsApp, and Signal, with an AI-native control layer.",
    images: ["/hero-product-dark.png"]
  }
};

export default function RootLayout({ children }) {
  const gaMeasurementId = getGoogleAnalyticsMeasurementId();

  return (
    <html lang="en" className={`${sans.variable} ${mono.variable}`}>
      <body>
        <GoogleAnalytics measurementId={gaMeasurementId} />
        {children}
      </body>
    </html>
  );
}
