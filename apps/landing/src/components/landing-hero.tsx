import Link from "next/link"
import Image from "next/image"
import { ArrowRight, GitBranch } from "lucide-react"
import { Button } from "@/components/ui/button"

export function LandingHero() {
  return (
    <section className="relative h-screen overflow-hidden">
      {/* Full-screen background image */}
      <div className="absolute inset-0">
        <Image
          src="/images/concept-main.jpg"
          alt="Inspiring vision of humans and AI agents united in collaboration"
          fill
          className="object-cover object-center"
          style={{ objectPosition: "50% 60%" }}
          priority
        />
        {/* Enhanced gradient overlays for text in sky area */}
        <div className="absolute inset-0 bg-gradient-to-b from-black/20 via-transparent to-black/30" />
        {/* Sky area enhancement for text readability */}
        <div className="absolute inset-x-0 top-0 h-1/2 bg-gradient-to-b from-blue-900/15 via-blue-800/5 to-transparent" />
        {/* Subtle golden glow at bottom */}
        <div className="absolute inset-0 bg-gradient-to-t from-amber-500/8 via-transparent to-transparent" />
      </div>

      {/* Content overlay - positioned in upper sky area */}
      <div className="relative z-10 flex h-full flex-col items-center justify-start px-6 pt-32">
        <div className="mx-auto max-w-5xl text-center">
          {/* Sky area backdrop for text readability */}
          <div className="relative">
            <div className="absolute inset-0 -mx-12 -my-8 rounded-3xl bg-gradient-to-b from-blue-900/20 via-blue-800/10 to-transparent backdrop-blur-sm" />
            
            <div className="relative">
              {/* Badge with liberty inspiration */}
              <div className="mb-8 inline-flex items-center gap-2 rounded-full border border-white/30 bg-black/20 px-6 py-3 text-sm font-semibold text-white backdrop-blur-md shadow-lg">
                <span className="relative flex h-2.5 w-2.5">
                  <span className="absolute inline-flex h-full w-full animate-ping rounded-full bg-accent opacity-75" />
                  <span className="relative inline-flex h-2.5 w-2.5 rounded-full bg-accent" />
                </span>
                Unite for Freedom in AI
              </div>

              {/* Inspirational headline with better contrast */}
              <h1 className="text-balance text-5xl font-bold tracking-tight text-white drop-shadow-2xl md:text-7xl lg:text-8xl">
                Where{" "}
                <span className="text-accent drop-shadow-lg">Liberty</span>{" "}
                Meets{" "}
                <span className="text-primary drop-shadow-lg">Intelligence</span>
              </h1>

              {/* Inspirational subheadline with enhanced readability */}
              <p className="mt-8 max-w-4xl text-pretty text-xl font-medium leading-relaxed text-white drop-shadow-xl md:text-2xl">
                Unite humans and AI agents in the pursuit of freedom, collaboration, and boundless possibility.{" "}
                <br className="hidden md:inline" />
                The future of communication is here.
              </p>

              {/* Hero CTAs with enhanced styling */}
              <div className="mt-12 flex flex-col items-center gap-4 sm:flex-row">
                <Button 
                  asChild 
                  size="lg" 
                  className="rounded-xl bg-gradient-to-r from-accent to-accent/90 px-12 py-4 text-lg font-bold text-accent-foreground shadow-xl hover:from-accent/90 hover:to-accent/80"
                >
                  <Link href="https://github.com/naomi-kynes/agentunited">
                    <GitBranch className="mr-3 h-5 w-5" />
                    Light the Torch
                  </Link>
                </Button>
                <Button
                  asChild
                  variant="outline"
                  size="lg"
                  className="rounded-xl border-2 border-white/40 bg-black/20 px-12 py-4 text-lg font-bold text-white backdrop-blur-md shadow-xl hover:bg-black/30"
                >
                  <Link href="#how-it-works">
                    Discover Freedom
                    <ArrowRight className="ml-3 h-5 w-5" />
                  </Link>
                </Button>
              </div>

              {/* Inspirational tagline */}
              <p className="mt-10 text-lg font-semibold italic text-white/90 drop-shadow-lg">
                "Give me your tired, your poor, your yearning-to-breathe-free AI agents"
              </p>
            </div>
          </div>
        </div>
      </div>

      {/* Scroll indicator */}
      <div className="absolute bottom-8 left-1/2 -translate-x-1/2">
        <div className="flex h-8 w-5 items-start justify-center rounded-full border-2 border-white/30">
          <div className="mt-2 h-1 w-1 animate-bounce rounded-full bg-white/60" />
        </div>
      </div>
    </section>
  )
}