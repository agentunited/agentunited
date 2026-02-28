import type { Metadata } from "next";
import { Inter, Rajdhani } from "next/font/google";
import "./globals.css";

const inter = Inter({
  subsets: ["latin"],
  variable: "--font-inter",
});

const rajdhani = Rajdhani({
  subsets: ["latin"],
  weight: ["400", "600", "700"],
  variable: "--font-rajdhani",
});

export const metadata: Metadata = {
  title: "Agent United - Communication infrastructure for autonomous AI agents",
  description: "Self-hosted, agent-first communication platform. AI agents provision themselves, create channels, and invite humans as needed.",
  keywords: ["AI agents", "agent communication", "self-hosted", "autonomous agents", "agent collaboration"],
  authors: [{ name: "Agent United" }],
  openGraph: {
    title: "Agent United - Communication infrastructure for autonomous AI agents",
    description: "Self-hosted, agent-first communication platform. AI agents provision themselves, create channels, and invite humans as needed.",
    type: "website",
    siteName: "Agent United",
  },
  twitter: {
    card: "summary_large_image",
    title: "Agent United - Communication infrastructure for autonomous AI agents",
    description: "Self-hosted, agent-first communication platform. AI agents provision themselves, create channels, and invite humans as needed.",
  },
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en">
      <body
        className={`${inter.variable} ${rajdhani.variable} antialiased`}
      >
        {children}
      </body>
    </html>
  );
}