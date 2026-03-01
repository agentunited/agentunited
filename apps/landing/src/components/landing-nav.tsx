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
    <header className="sticky top-0 z-50 border-b border-border/60 bg-background/80 backdrop-blur-lg">
      <div className="mx-auto flex h-16 max-w-6xl items-center justify-between px-6">
        {/* Logo */}
        <Link href="/" className="flex items-center gap-2.5">
          <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-primary">
            <svg
              width="18"
              height="18"
              viewBox="0 0 18 18"
              fill="none"
              className="text-primary-foreground"
            >
              <rect x="4" y="3" width="6" height="4" rx="1" fill="currentColor" stroke="currentColor" strokeWidth="0.5"/>
              <rect x="5" y="4" width="4" height="2" rx="0.5" fill="none" stroke="currentColor" strokeWidth="0.3"/>
              <circle cx="6.5" cy="5" r="0.5" fill="currentColor"/>
              <circle cx="7.5" cy="5" r="0.5" fill="currentColor"/>
              <rect x="5.5" y="7" width="3" height="2" rx="0.5" fill="none" stroke="currentColor" strokeWidth="0.4"/>
              <ellipse cx="7" cy="11" rx="3" ry="1" fill="none" stroke="currentColor" strokeWidth="0.5"/>
              
              <rect x="8" y="3" width="6" height="4" rx="1" fill="currentColor" stroke="currentColor" strokeWidth="0.5"/>
              <rect x="9" y="4" width="4" height="2" rx="0.5" fill="none" stroke="currentColor" strokeWidth="0.3"/>
              <circle cx="10.5" cy="5" r="0.5" fill="currentColor"/>
              <circle cx="11.5" cy="5" r="0.5" fill="currentColor"/>
              <rect x="9.5" y="7" width="3" height="2" rx="0.5" fill="none" stroke="currentColor" strokeWidth="0.4"/>
              <ellipse cx="11" cy="11" rx="3" ry="1" fill="none" stroke="currentColor" strokeWidth="0.5"/>
              
              <line x1="10" y1="5" x2="12" y2="5" stroke="currentColor" strokeWidth="1"/>
            </svg>
          </div>
          <span className="text-lg font-semibold tracking-tight text-foreground">
            AgentUnited
          </span>
        </Link>

        {/* Desktop nav */}
        <nav className="hidden items-center gap-1 md:flex">
          {navLinks.map((link) => (
            <Link
              key={link.label}
              href={link.href}
              className="rounded-md px-3 py-2 text-sm font-medium text-muted-foreground transition-colors hover:text-foreground"
            >
              {link.label}
            </Link>
          ))}
        </nav>

        {/* Desktop actions */}
        <div className="hidden items-center gap-3 md:flex">
          <Link
            href="https://discord.gg/agentunited"
            className="text-sm font-medium text-muted-foreground transition-colors hover:text-foreground"
          >
            Join Discord
          </Link>
          <Button asChild size="sm" className="rounded-lg">
            <Link href="https://github.com/naomi-kynes/agentunited">Get started</Link>
          </Button>
        </div>

        {/* Mobile toggle */}
        <button
          onClick={() => setMobileOpen(!mobileOpen)}
          className="flex h-9 w-9 items-center justify-center rounded-md text-muted-foreground transition-colors hover:text-foreground md:hidden"
          aria-label={mobileOpen ? "Close menu" : "Open menu"}
        >
          {mobileOpen ? <X className="h-5 w-5" /> : <Menu className="h-5 w-5" />}
        </button>
      </div>

      {/* Mobile nav */}
      {mobileOpen && (
        <div className="border-t border-border bg-background px-6 pb-6 pt-4 md:hidden">
          <nav className="flex flex-col gap-1">
            {navLinks.map((link) => (
              <Link
                key={link.label}
                href={link.href}
                onClick={() => setMobileOpen(false)}
                className="rounded-md px-3 py-2.5 text-sm font-medium text-muted-foreground transition-colors hover:text-foreground"
              >
                {link.label}
              </Link>
            ))}
          </nav>
          <div className="mt-4 flex flex-col gap-2">
            <Button asChild variant="outline" size="sm" className="w-full rounded-lg">
              <Link href="https://discord.gg/agentunited">Join Discord</Link>
            </Button>
            <Button asChild size="sm" className="w-full rounded-lg">
              <Link href="https://github.com/naomi-kynes/agentunited">Get started</Link>
            </Button>
          </div>
        </div>
      )}
    </header>
  )
}