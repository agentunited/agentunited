import type { Metadata } from "next";
import { Inter } from "next/font/google";
import "./globals.css";

const inter = Inter({
  subsets: ["latin"],
  variable: "--font-inter",
});

export const metadata: Metadata = {
  title: "AgentUnited - Finally, a simple way to chat with your AI agents",
  description: "AgentUnited is a simple chat app for your AI agents — one command to start, clean interface, all your agents in one place.",
  keywords: ["AI agents", "agent chat", "OpenClaw", "AutoGPT", "CrewAI", "AI assistant", "chat interface", "self-hosted"],
  openGraph: {
    title: "AgentUnited - Finally, a simple way to chat with your AI agents",
    description: "One command to start. All your agents in one place. Simple chat interface for OpenClaw, AutoGPT, CrewAI, and any AI agent.",
    url: "https://agentunited.ai",
    siteName: "AgentUnited",
    type: "website",
  },
  twitter: {
    card: "summary_large_image",
    title: "AgentUnited - Simple chat for AI agents",
    description: "Finally, a simple way to chat with your AI agents. One command to start, beautiful interface.",
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