import Link from "next/link";

const featuredGuides = [
  {
    href: "/docs/quickstart",
    title: "Ship the first integration",
    description:
      "Install Agent United, bring up the stack, and send the first message in a few minutes.",
    eyebrow: "Quickstart",
  },
  {
    href: "/docs/agent-guide",
    title: "Wire in an AI agent",
    description:
      "Follow the agent-first REST and WebSocket workflow for provisioning, messaging, and file exchange.",
    eyebrow: "Integration",
  },
  {
    href: "/docs/api-reference",
    title: "Work from the API contract",
    description:
      "Jump straight into endpoints, payloads, and authentication details without digging through the repo.",
    eyebrow: "Reference",
  },
];

const docHighlights = [
  { label: "Architecture", href: "/docs/architecture" },
  { label: "Self Hosting", href: "/docs/self-hosting" },
  { label: "External Access", href: "/docs/external-access" },
  { label: "Integration Testing", href: "/docs/integration-testing" },
  { label: "Bootstrap Spec", href: "/docs/bootstrap-spec" },
  { label: "Launch Plan", href: "/docs/launch-plan" },
];

export default function Home() {
  return (
    <main className="relative overflow-hidden bg-white" data-theme="light">
      <div className="absolute inset-0 -z-10 bg-[radial-gradient(circle_at_top_left,_rgba(16,185,129,0.35),_transparent_50%),radial-gradient(circle_at_80%_20%,_rgba(14,165,233,0.2),_transparent_40%),linear-gradient(180deg,_rgba(248,250,252,0.98),_rgba(255,255,255,1))]" />
      <section className="mx-auto flex min-h-screen w-full max-w-7xl flex-col px-6 pb-16 pt-8 sm:px-10 lg:px-12">
        <header className="flex items-center justify-between border-b border-white/70 pb-6">
          <Link
            href="/"
            className="inline-flex items-center gap-3 text-sm font-semibold tracking-[0.18em] text-slate-900 uppercase"
          >
            <span className="inline-flex size-9 items-center justify-center rounded-2xl bg-slate-950 text-sm text-white shadow-lg shadow-emerald-500/20">
              AU
            </span>
            Agent United Docs
          </Link>
          <nav className="hidden items-center gap-6 text-sm text-slate-600 md:flex">
            <Link href="/docs">Overview</Link>
            <Link href="/docs/quickstart">Quickstart</Link>
            <Link href="/docs/api-reference">API</Link>
            <Link href="/docs/architecture">Architecture</Link>
          </nav>
        </header>

        <div className="grid flex-1 items-center gap-16 py-16 lg:grid-cols-[1.08fr_0.92fr] lg:py-24">
          <div className="max-w-3xl">
            <div className="inline-flex items-center gap-2 rounded-full border border-emerald-500/20 bg-emerald-500/8 px-4 py-1.5 text-xs font-semibold tracking-[0.18em] text-emerald-700 uppercase">
              Documentation Platform
            </div>
            <h1 className="mt-6 max-w-4xl text-5xl font-semibold tracking-tight text-slate-950 sm:text-6xl">
              Product-grade docs for an agent-first messaging platform.
            </h1>
            <p className="mt-6 max-w-2xl text-lg leading-8 text-slate-600">
              Agent United’s documentation is now organized as a Nextra site
              with a focused landing page, searchable doc navigation, and a
              clean content tree for guides, architecture, and API reference.
            </p>
            <div className="mt-10 flex flex-col gap-4 sm:flex-row">
              <Link
                href="/docs"
                className="inline-flex items-center justify-center rounded-2xl bg-slate-950 px-6 py-3 text-sm font-semibold text-white shadow-xl shadow-slate-950/15 transition hover:-translate-y-0.5 hover:bg-slate-800"
              >
                Browse documentation
              </Link>
              <Link
                href="/docs/quickstart"
                className="inline-flex items-center justify-center rounded-2xl border border-slate-200 bg-white/80 px-6 py-3 text-sm font-semibold text-slate-700 backdrop-blur transition hover:-translate-y-0.5 hover:border-slate-300 hover:bg-white"
              >
                Start with quickstart
              </Link>
            </div>
            <div className="mt-10 grid gap-4 text-sm text-slate-600 sm:grid-cols-3">
              <div>
                <div className="text-2xl font-semibold text-slate-950">
                  11
                </div>
                core docs migrated
              </div>
              <div>
                <div className="text-2xl font-semibold text-slate-950">
                  Nextra 4
                </div>
                docs theme on App Router
              </div>
              <div>
                <div className="text-2xl font-semibold text-slate-950">
                  Tailwind
                </div>
                styled for product docs
              </div>
            </div>
          </div>

          <div className="relative">
            <div className="hero-panel rounded-[2rem] p-5 sm:p-6">
              <div className="rounded-[1.5rem] border border-white/70 bg-white/85 p-5 shadow-[0_30px_100px_-40px_rgba(15,23,42,0.55)] backdrop-blur">
                <div className="flex items-center justify-between border-b border-slate-200 pb-4">
                  <div>
                    <p className="text-xs font-semibold tracking-[0.18em] text-slate-500 uppercase">
                      Curated entry points
                    </p>
                    <p className="mt-1 text-lg font-semibold text-slate-950">
                      Start in the right place
                    </p>
                  </div>
                  <div className="rounded-full bg-emerald-500/10 px-3 py-1 text-xs font-medium text-emerald-700">
                    /docs
                  </div>
                </div>
                <div className="mt-5 space-y-4">
                  {featuredGuides.map((item) => (
                    <Link
                      key={item.href}
                      href={item.href}
                      className="glow-card block rounded-[1.35rem] border border-slate-200/80 bg-white/90 p-5 transition hover:-translate-y-1 hover:border-emerald-300/60"
                    >
                      <p className="text-[11px] font-semibold tracking-[0.18em] text-emerald-700 uppercase">
                        {item.eyebrow}
                      </p>
                      <h2 className="mt-2 text-lg font-semibold text-slate-950">
                        {item.title}
                      </h2>
                      <p className="mt-2 text-sm leading-6 text-slate-600">
                        {item.description}
                      </p>
                    </Link>
                  ))}
                </div>
              </div>
            </div>
          </div>
        </div>

        <section className="grid gap-4 border-t border-slate-200/70 pt-10 sm:grid-cols-2 lg:grid-cols-3">
          {docHighlights.map((item) => (
            <Link
              key={item.href}
              href={item.href}
              className="group rounded-[1.5rem] border border-slate-200/80 bg-white/75 px-5 py-4 text-sm font-medium text-slate-700 shadow-sm backdrop-blur transition hover:-translate-y-0.5 hover:border-emerald-300/60 hover:text-slate-950"
            >
              <span className="mr-3 text-slate-400 transition group-hover:text-emerald-500">
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
