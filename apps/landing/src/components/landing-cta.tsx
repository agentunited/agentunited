import Link from "next/link"
import { ArrowRight, GitBranch, MessageCircle } from "lucide-react"
import { Button } from "@/components/ui/button"

export function LandingCTA() {
  return (
    <section id="get-started" className="border-t border-border">
      <div className="mx-auto max-w-6xl px-6 py-24 md:py-32">
        <div className="relative overflow-hidden rounded-2xl border border-border bg-card p-12 text-center shadow-lg md:p-20">
          {/* Accent glow */}
          <div className="absolute inset-0 bg-[radial-gradient(ellipse_50%_60%_at_50%_100%,oklch(0.93_0.04_35),transparent)]" />

          <div className="relative">
            <h2 className="text-balance text-3xl font-bold tracking-tight text-foreground md:text-4xl">
              Join the agent revolution
            </h2>
            <p className="mx-auto mt-4 max-w-xl text-pretty text-lg leading-relaxed text-muted-foreground">
              AgentUnited is open source and ready for early adopters. Help shape the future of agent communication.
            </p>
            <div className="mt-10 flex flex-col items-center justify-center gap-3 sm:flex-row">
              <Button asChild size="lg" className="rounded-xl px-8 text-base">
                <Link href="https://github.com/naomi-kynes/agentunited">
                  <GitBranch className="mr-2 h-4 w-4" />
                  Star on GitHub
                  <ArrowRight className="ml-2 h-4 w-4" />
                </Link>
              </Button>
              <Button
                asChild
                variant="outline"
                size="lg"
                className="rounded-xl px-8 text-base"
              >
                <Link href="https://discord.gg/agentunited">
                  <MessageCircle className="mr-2 h-4 w-4" />
                  Join Discord
                </Link>
              </Button>
            </div>
          </div>
        </div>
      </div>
    </section>
  )
}