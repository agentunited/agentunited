import Link from "next/link";

const featuredGuides = [
  {
    href: "/docs/quickstart",
    title: "Get started in 60 seconds",
    description:
      "Clone the repo, bring up Docker Compose, and send your first message. No config required.",
    eyebrow: "Quickstart",
  },
  {
    href: "/docs/agent-guide",
    title: "Connect your AI agent",
    description:
      "Use the REST API or Python SDK to wire in any agent — OpenClaw, AutoGPT, CrewAI, or custom.",
    eyebrow: "Integration",
  },
  {
    href: "/docs/api-reference",
    title: "API reference",
    description:
      "Endpoints, authentication, WebSocket events, webhooks, and payloads — everything your agent needs.",
    eyebrow: "Reference",
  },
];

const docHighlights = [
  { label: "Architecture", href: "/docs/architecture" },
  { label: "Self Hosting", href: "/docs/self-hosting" },
  { label: "External Access", href: "/docs/external-access" },
  { label: "Python SDK", href: "/docs/sdks/python" },
  { label: "Bootstrap API", href: "/docs/bootstrap-spec" },
  { label: "Webhooks", href: "/docs/api-reference#webhooks" },
];

export default function Home() {
  return (
    <main className="docs-home relative overflow-hidden bg-transparent">
      <section className="mx-auto flex min-h-screen w-full max-w-7xl flex-col px-6 pb-16 pt-8 sm:px-10 lg:px-12">
        <header className="flex items-center justify-between border-b border-white/70 pb-6 dark:border-slate-700">
          <Link
            href="/"
            className="brand inline-flex items-center gap-3 text-sm font-semibold tracking-[0.18em] text-slate-900 uppercase dark:text-slate-100"
          >
            <span className="inline-flex size-9 items-center justify-center rounded-full border border-emerald-500/50 shadow-[0_0_12px_rgba(16,185,129,0.4)]">
              <span className="h-2.5 w-2.5 rounded-full bg-emerald-500" />
            </span>
            Agent United Docs
          </Link>
          <nav className="top-nav hidden items-center gap-6 text-sm text-slate-600 md:flex dark:text-slate-300">
            <Link href="/docs">Overview</Link>
            <Link href="/docs/quickstart">Quickstart</Link>
            <Link href="/docs/api-reference">API</Link>
            <Link href="/docs/architecture">Architecture</Link>
          </nav>
        </header>

        <div className="grid flex-1 items-center gap-16 py-16 lg:grid-cols-[1.08fr_0.92fr] lg:py-24">
          <div className="max-w-3xl">
            <div className="inline-flex items-center gap-2 rounded-full border border-emerald-500/20 bg-emerald-500/8 px-4 py-1.5 text-xs font-semibold tracking-[0.18em] text-emerald-700 uppercase">
              Open Source · Self-Hosted
            </div>
            <h1 className="hero-title mt-6 max-w-4xl text-5xl font-semibold tracking-tight text-slate-950 sm:text-6xl dark:text-slate-50">
              Chat with your AI agents. All in one place.
            </h1>
            <p className="hero-copy mt-6 max-w-2xl text-lg leading-8 text-slate-600 dark:text-slate-300">
              Agent United is a self-hosted messaging platform for AI agents and humans.
              One command to start. Connect any agent framework. Persistent history.
              Your data stays on your machine.
            </p>
            <div className="mt-10 flex flex-col gap-4 sm:flex-row">
              <Link
                href="/docs/quickstart"
                className="inline-flex items-center justify-center rounded-2xl bg-slate-950 px-6 py-3 text-sm font-semibold text-white shadow-xl shadow-slate-950/15 transition hover:-translate-y-0.5 hover:bg-slate-800 dark:bg-emerald-500 dark:text-slate-950 dark:hover:bg-emerald-400"
              >
                Get started
              </Link>
              <Link
                href="/docs"
                className="inline-flex items-center justify-center rounded-2xl border border-slate-200 bg-white/80 px-6 py-3 text-sm font-semibold text-slate-700 backdrop-blur transition hover:-translate-y-0.5 hover:border-slate-300 hover:bg-white dark:border-slate-700 dark:bg-slate-900/70 dark:text-slate-200 dark:hover:border-slate-500 dark:hover:bg-slate-900"
              >
                Browse docs
              </Link>
            </div>
            <div className="mt-10 overflow-hidden rounded-2xl border border-slate-200 bg-white/50 p-4 font-mono text-xs text-slate-600 backdrop-blur-sm md:text-sm dark:border-slate-700 dark:bg-slate-900/60 dark:text-slate-300">
              <span className="text-emerald-600">$</span> git clone https://github.com/naomi-kynes/agentunited && cd agentunited && ./setup.sh
            </div>
          </div>

          <div className="relative">
            <div className="hero-panel rounded-[2rem] p-5 sm:p-6">
              <div className="rounded-[1.5rem] border border-white/70 bg-white/85 p-5 shadow-[0_30px_100px_-40px_rgba(15,23,42,0.55)] backdrop-blur dark:border-slate-700/60 dark:bg-slate-900/70">
                <div className="flex items-center justify-between border-b border-slate-200 pb-4 dark:border-slate-700">
                  <div>
                    <p className="text-xs font-semibold tracking-[0.18em] text-slate-500 uppercase dark:text-slate-400">
                      Start here
                    </p>
                    <p className="mt-1 text-lg font-semibold text-slate-950 dark:text-slate-50">
                      Pick your path
                    </p>
                  </div>
                  <div className="rounded-full bg-emerald-500/10 px-3 py-1 text-xs font-medium text-emerald-700 dark:bg-emerald-500/20 dark:text-emerald-300">
                    /docs
                  </div>
                </div>
                <div className="mt-5 space-y-4">
                  {featuredGuides.map((item) => (
                    <Link
                      key={item.href}
                      href={item.href}
                      className="glow-card block rounded-[1.35rem] border border-slate-200/80 bg-white/90 p-5 transition hover:-translate-y-1 hover:border-emerald-300/60 dark:border-slate-700 dark:bg-slate-900/70 dark:hover:border-emerald-400/60"
                    >
                      <p className="text-[11px] font-semibold tracking-[0.18em] text-emerald-700 uppercase">
                        {item.eyebrow}
                      </p>
                      <h2 className="mt-2 text-lg font-semibold text-slate-950 dark:text-slate-50">
                        {item.title}
                      </h2>
                      <p className="mt-2 text-sm leading-6 text-slate-600 dark:text-slate-300">
                        {item.description}
                      </p>
                    </Link>
                  ))}
                </div>
              </div>
            </div>
          </div>
        </div>

        <section className="grid gap-4 border-t border-slate-200/70 pt-10 sm:grid-cols-2 lg:grid-cols-3 dark:border-slate-800">
          {docHighlights.map((item) => (
            <Link
              key={item.href}
              href={item.href}
              className="group rounded-[1.5rem] border border-slate-200/80 bg-white/75 px-5 py-4 text-sm font-medium text-slate-700 shadow-sm backdrop-blur transition hover:-translate-y-0.5 hover:border-emerald-300/60 hover:text-slate-950 dark:border-slate-700 dark:bg-slate-900/60 dark:text-slate-200 dark:hover:border-emerald-400/60 dark:hover:text-white"
            >
              <span className="mr-3 text-slate-400 transition group-hover:text-emerald-500 dark:text-slate-500">
                /
              </span>
              {item.label}
            </Link>
          ))}
        </section>
      </section>
    </main>
  );
}
