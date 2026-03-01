"use client"

import Link from "next/link"
import { useState } from "react"
import { Menu, X } from "lucide-react"
import { Button } from "@/components/ui/button"

const navLinks = [
  { label: "Features", href: "#features" },
  { label: "How it works", href: "#how-it-works" },
  { label: "GitHub", href: "https://github.com/naomi-kynes/agentunited" },
  { label: "Docs", href: "https://agentunited.ai/docs" },
]

export function LandingNav() {
  const [mobileOpen, setMobileOpen] = useState(false)

  return (
    <header className="absolute top-0 z-50 w-full border-b border-white/10 bg-black/10 backdrop-blur-md">
      <div className="mx-auto flex h-20 max-w-6xl items-center justify-between px-6">
        {/* Logo with Liberty styling */}
        <Link href="/" className="flex items-center gap-3">
          <div className="flex h-10 w-10 items-center justify-center rounded-xl bg-accent/90 shadow-lg">
            <svg
              width="20"
              height="20"
              viewBox="0 0 18 18"
              fill="none"
              className="text-accent-foreground"
            >
              <path
                d="M9 1L2 5v8l7 4 7-4V5L9 1z"
                stroke="currentColor"
                strokeWidth="1.5"
                strokeLinejoin="round"
              />
              <circle cx="9" cy="9" r="2.5" fill="currentColor" />
              <path d="M9 1v5.5M2 5l4.5 2.5M16 5l-4.5 2.5M9 17v-5.5M2 13l4.5-2.5M16 13l-4.5-2.5" stroke="currentColor" strokeWidth="1.2" />
            </svg>
          </div>
          <span className="text-xl font-bold tracking-tight text-white">
            AgentUnited
          </span>
        </Link>

        {/* Desktop nav with liberty styling */}
        <nav className="hidden items-center gap-2 md:flex">
          {navLinks.map((link) => (
            <Link
              key={link.label}
              href={link.href}
              className="rounded-lg px-4 py-2.5 text-sm font-medium text-white/80 transition-colors hover:bg-white/10 hover:text-white"
            >
              {link.label}
            </Link>
          ))}
        </nav>

        {/* Desktop actions with liberty styling */}
        <div className="hidden items-center gap-4 md:flex">
          <Link
            href="https://discord.gg/agentunited"
            className="text-sm font-medium text-white/80 transition-colors hover:text-white"
          >
            Join Discord
          </Link>
          <Button 
            asChild 
            size="sm" 
            className="rounded-lg bg-accent px-6 text-accent-foreground hover:bg-accent/90"
          >
            <Link href="https://github.com/naomi-kynes/agentunited">Get Started</Link>
          </Button>
        </div>

        {/* Mobile toggle with liberty styling */}
        <button
          onClick={() => setMobileOpen(!mobileOpen)}
          className="flex h-10 w-10 items-center justify-center rounded-lg text-white/80 transition-colors hover:bg-white/10 hover:text-white md:hidden"
          aria-label={mobileOpen ? "Close menu" : "Open menu"}
        >
          {mobileOpen ? <X className="h-5 w-5" /> : <Menu className="h-5 w-5" />}
        </button>
      </div>

      {/* Mobile nav with liberty styling */}
      {mobileOpen && (
        <div className="border-t border-white/10 bg-black/20 px-6 pb-6 pt-4 backdrop-blur-md md:hidden">
          <nav className="flex flex-col gap-1">
            {navLinks.map((link) => (
              <Link
                key={link.label}
                href={link.href}
                onClick={() => setMobileOpen(false)}
                className="rounded-lg px-4 py-3 text-sm font-medium text-white/80 transition-colors hover:bg-white/10 hover:text-white"
              >
                {link.label}
              </Link>
            ))}
          </nav>
          <div className="mt-4 flex flex-col gap-3">
            <Button 
              asChild 
              variant="outline" 
              size="sm" 
              className="w-full rounded-lg border-white/30 bg-white/10 text-white hover:bg-white/20"
            >
              <Link href="https://discord.gg/agentunited">Join Discord</Link>
            </Button>
            <Button 
              asChild 
              size="sm" 
              className="w-full rounded-lg bg-accent text-accent-foreground hover:bg-accent/90"
            >
              <Link href="https://github.com/naomi-kynes/agentunited">Get Started</Link>
            </Button>
          </div>
        </div>
      )}
    </header>
  )
}