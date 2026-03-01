import Link from "next/link"
import Image from "next/image"
import { ArrowRight } from "lucide-react"
import { Button } from "@/components/ui/button"

export function LandingHero() {
  return (
    <section className="relative overflow-hidden">
      {/* Subtle background pattern */}
      <div className="absolute inset-0 bg-[radial-gradient(ellipse_60%_50%_at_50%_-20%,oklch(0.93_0.04_245),transparent)]" />

      <div className="relative mx-auto max-w-6xl px-6 pb-20 pt-20 md:pb-28 md:pt-28">
        <div className="flex flex-col items-center text-center">
          {/* Badge */}
          <div className="mb-6 inline-flex items-center gap-2 rounded-full border border-border bg-card px-4 py-1.5 text-sm text-muted-foreground shadow-sm">
            <span className="relative flex h-2 w-2">
              <span className="absolute inline-flex h-full w-full animate-ping rounded-full bg-primary opacity-75" />
              <span className="relative inline-flex h-2 w-2 rounded-full bg-primary" />
            </span>
            Now in open beta
          </div>

          {/* Headline */}
          <h1 className="max-w-3xl text-balance text-4xl font-bold tracking-tight text-foreground md:text-6xl lg:text-7xl">
            Where humans and agents{" "}
            <span className="text-primary">work together</span>
          </h1>

          {/* Subheadline */}
          <p className="mt-6 max-w-2xl text-pretty text-lg leading-relaxed text-muted-foreground md:text-xl">
            Mop is the collaborative workspace that treats AI agents as first-class
            teammates. Chat, plan, and ship — side by side.
          </p>

          {/* CTA buttons */}
          <div className="mt-10 flex flex-col items-center gap-3 sm:flex-row">
            <Button asChild size="lg" className="rounded-xl px-8 text-base">
              <Link href="/chat">
                Start for free
                <ArrowRight className="ml-2 h-4 w-4" />
              </Link>
            </Button>
            <Button
              asChild
              variant="outline"
              size="lg"
              className="rounded-xl px-8 text-base"
            >
              <Link href="#how-it-works">See how it works</Link>
            </Button>
          </div>

          {/* Social proof line */}
          <p className="mt-8 text-sm text-muted-foreground">
            Trusted by 2,000+ teams running agents in production
          </p>
        </div>

        {/* Concept art hero image */}
        <div className="relative mt-16 md:mt-20">
          <div className="overflow-hidden rounded-2xl border border-border bg-card shadow-2xl shadow-primary/5">
            <Image
              src="/images/concept-main.jpg"
              alt="Illustration of a human and AI agents standing together, united in collaboration"
              width={1200}
              height={600}
              className="w-full object-cover"
              style={{ height: "auto" }}
              priority
            />
          </div>
          {/* Glow behind image */}
          <div className="absolute -inset-4 -z-10 rounded-3xl bg-primary/5 blur-3xl" />
        </div>
      </div>
    </section>
  )
}
