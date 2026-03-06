import type { ReactNode } from "react";
import Link from "next/link";
import { getPageMap } from "nextra/page-map";
import { Footer, Layout, Navbar } from "nextra-theme-docs";

export default async function DocsLayout({
  children,
}: {
  children: ReactNode;
}) {
  const pageMap = await getPageMap("/docs");

  return (
    <Layout
      darkMode
      docsRepositoryBase="https://github.com/agentunited/agentunited-docs/tree/main/content/docs"
      editLink={null}
      feedback={{ content: null }}
      footer={
        <Footer className="border-t border-black/5 bg-white/70 px-6 py-8 text-sm text-slate-600 backdrop-blur dark:border-white/10 dark:bg-slate-950/60 dark:text-slate-300">
          <div className="mx-auto flex max-w-7xl flex-col gap-3 md:flex-row md:items-center md:justify-between">
            <p>Agent United documentation site built with Nextra and Tailwind.</p>
            <div className="flex items-center gap-4">
              <Link href="/docs">Docs</Link>
              <Link href="/docs/quickstart">Quickstart</Link>
              <Link href="/docs/api-reference">API Reference</Link>
            </div>
          </div>
        </Footer>
      }
      navbar={
        <Navbar
          logo={
            <span className="flex items-center gap-3">
              <span className="inline-flex size-9 items-center justify-center rounded-full border border-emerald-500/50 shadow-[0_0_12px_rgba(16,185,129,0.4)] dark:shadow-[0_0_12px_rgba(16,185,129,0.5)]">
                <span className="h-2.5 w-2.5 rounded-full bg-emerald-500" />
              </span>
              <span className="flex flex-col leading-tight">
                <span className="text-sm font-semibold tracking-[0.18em] uppercase">
                  Agent United
                </span>
                <span className="text-xs font-medium text-slate-500 dark:text-slate-400">
                  Documentation
                </span>
              </span>
            </span>
          }
          logoLink="/"
        >
          <div className="hidden items-center gap-5 text-sm font-medium text-slate-600 md:flex dark:text-slate-300">
            <Link href="/docs">Overview</Link>
            <Link href="/docs/quickstart">Quickstart</Link>
            <Link href="/docs/api-reference">API</Link>
            <Link href="/docs/architecture">Architecture</Link>
          </div>
        </Navbar>
      }
      navigation={{
        next: true,
        prev: true,
      }}
      pageMap={pageMap}
      sidebar={{
        autoCollapse: true,
        defaultMenuCollapseLevel: 1,
        defaultOpen: true,
        toggleButton: true,
      }}
      toc={{
        backToTop: "Back to top",
        title: "On this page",
      }}
    >
      {children}
    </Layout>
  );
}
