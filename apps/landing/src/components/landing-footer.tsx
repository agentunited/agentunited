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