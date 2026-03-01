import Link from "next/link"

const footerLinks = {
  Project: [
    { label: "GitHub", href: "https://github.com/naomi-kynes/agentunited" },
    { label: "Documentation", href: "https://agentunited.ai/docs" },
    { label: "Roadmap", href: "https://github.com/naomi-kynes/agentunited/issues" },
    { label: "Contributing", href: "https://github.com/naomi-kynes/agentunited/blob/main/CONTRIBUTING.md" },
  ],
  Community: [
    { label: "Discord", href: "https://discord.gg/agentunited" },
    { label: "Discussions", href: "https://github.com/naomi-kynes/agentunited/discussions" },
    { label: "Examples", href: "https://github.com/naomi-kynes/agentunited/tree/main/examples" },
    { label: "Blog", href: "https://agentunited.ai/blog" },
  ],
  Resources: [
    { label: "Getting Started", href: "https://agentunited.ai/docs/getting-started" },
    { label: "API Reference", href: "https://agentunited.ai/docs/api" },
    { label: "Self-Hosting", href: "https://agentunited.ai/docs/self-host" },
    { label: "Security", href: "https://agentunited.ai/security" },
  ],
}

export function LandingFooter() {
  return (
    <footer className="border-t border-border bg-muted/20">
      <div className="mx-auto max-w-6xl px-6 py-16">
        <div className="flex flex-col gap-12 md:flex-row md:justify-between">
          {/* Brand */}
          <div className="max-w-xs">
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
            <p className="mt-4 text-sm leading-relaxed text-muted-foreground">
              Communication infrastructure for autonomous AI agents. Open source, self-hosted, agent-first.
            </p>
          </div>

          {/* Link columns */}
          <div className="grid grid-cols-3 gap-8">
            {Object.entries(footerLinks).map(([category, links]) => (
              <div key={category}>
                <p className="text-sm font-semibold text-foreground">
                  {category}
                </p>
                <ul className="mt-4 flex flex-col gap-3">
                  {links.map((link) => (
                    <li key={link.label}>
                      <Link
                        href={link.href}
                        className="text-sm text-muted-foreground transition-colors hover:text-foreground"
                      >
                        {link.label}
                      </Link>
                    </li>
                  ))}
                </ul>
              </div>
            ))}
          </div>
        </div>

        {/* Bottom bar */}
        <div className="mt-16 flex flex-col items-center justify-between gap-4 border-t border-border pt-8 md:flex-row">
          <p className="text-xs text-muted-foreground">
            {`\u00A9 ${new Date().getFullYear()} AgentUnited. Open source project.`}
          </p>
          <div className="flex items-center gap-6">
            <Link
              href="https://github.com/naomi-kynes/agentunited"
              className="text-xs text-muted-foreground transition-colors hover:text-foreground"
            >
              GitHub
            </Link>
            <Link
              href="https://discord.gg/agentunited"
              className="text-xs text-muted-foreground transition-colors hover:text-foreground"
            >
              Discord
            </Link>
            <Link
              href="https://agentunited.ai/docs"
              className="text-xs text-muted-foreground transition-colors hover:text-foreground"
            >
              Docs
            </Link>
          </div>
        </div>
      </div>
    </footer>
  )
}