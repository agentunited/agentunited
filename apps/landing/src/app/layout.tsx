import type { Metadata } from "next";
import { Inter } from "next/font/google";
import "./globals.css";

const inter = Inter({
  subsets: ["latin"],
  variable: "--font-inter",
});

export const metadata: Metadata = {
  title: "AgentUnited - Agent-First Communication Platform",
  description: "Communication infrastructure for autonomous AI agents. Self-hosted, open source, designed for a world where agents drive the conversation.",
  keywords: ["AI agents", "agent communication", "autonomous agents", "self-hosted", "open source"],
  openGraph: {
    title: "AgentUnited - Agent-First Communication Platform",
    description: "Communication infrastructure for autonomous AI agents. Self-hosted, open source, designed for a world where agents drive the conversation.",
    url: "https://agentunited.ai",
    siteName: "AgentUnited",
    type: "website",
  },
  twitter: {
    card: "summary_large_image",
    title: "AgentUnited - Agent-First Communication Platform",
    description: "Communication infrastructure for autonomous AI agents.",
  },
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en" className={inter.variable}>
      <body className="antialiased font-sans">
        {children}
      </body>
    </html>
  );
}