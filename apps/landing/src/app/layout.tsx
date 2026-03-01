import type { Metadata } from "next";
import { Inter, Rajdhani } from "next/font/google";
import "./globals.css";

const inter = Inter({
  subsets: ["latin"],
  variable: "--font-inter",
});

const rajdhani = Rajdhani({
  subsets: ["latin"],
  variable: "--font-rajdhani",
  weight: ["400", "500", "600", "700"],
});

export const metadata: Metadata = {
  title: "AgentUnited - The simplest way to chat with your AI agents",
  description: "The simplest way to chat with your AI agents. One command to start. All your agents in one place. Works with OpenClaw, AutoGPT, CrewAI, any agent.",
  keywords: ["AI agents", "agent chat", "OpenClaw", "AutoGPT", "CrewAI", "AI assistant", "chat interface", "self-hosted"],
  openGraph: {
    title: "AgentUnited - The simplest way to chat with your AI agents",
    description: "The simplest way to chat with your AI agents. One command to start. All your agents in one place.",
    url: "https://agentunited.ai",
    siteName: "AgentUnited",
    type: "website",
  },
  twitter: {
    card: "summary_large_image",
    title: "AgentUnited - The simplest way to chat with your AI agents",
    description: "The simplest way to chat with your AI agents. One command to start. All your agents in one place.",
  },
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en" className={`${inter.variable} ${rajdhani.variable}`}>
      <body className="antialiased font-sans">
        {children}
      </body>
    </html>
  );
}