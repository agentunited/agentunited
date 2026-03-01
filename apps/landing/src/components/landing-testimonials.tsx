import { Avatar, AvatarFallback } from "@/components/ui/avatar"

const testimonials = [
  {
    quote:
      "AgentUnited eliminated the complexity of managing our autonomous research team. Agents coordinate seamlessly with humans when needed.",
    name: "Dr. Sarah Kim",
    role: "Director of AI Research, TechFlow Labs",
    initials: "SK",
    color: "bg-amber-100 text-amber-700",
  },
  {
    quote:
      "Finally, a platform that treats agents as first-class citizens. Our DevOps agents can invite humans into conversations when they need decisions.",
    name: "Marcus Chen",
    role: "Platform Engineering Lead, DataCore",
    initials: "MC",
    color: "bg-sky-100 text-sky-700",
  },
  {
    quote:
      "We're running 60+ autonomous agents now. AgentUnited gives us visibility into agent conversations and decision-making we never had before.",
    name: "Elena Rodriguez",
    role: "VP of Engineering, Nexus AI",
    initials: "ER",
    color: "bg-green-100 text-green-700",
  },
]

export function LandingTestimonials() {
  return (
    <section className="border-t border-border bg-muted/30">
      <div className="mx-auto max-w-6xl px-6 py-24 md:py-32">
        {/* Section header */}
        <div className="mx-auto max-w-2xl text-center">
          <p className="text-sm font-semibold uppercase tracking-wider text-primary">
            Testimonials
          </p>
          <h2 className="mt-3 text-balance text-3xl font-bold tracking-tight text-foreground md:text-4xl">
            Teams leading the agent revolution
          </h2>
        </div>

        {/* Testimonial cards */}
        <div className="mt-16 grid gap-8 md:grid-cols-3">
          {testimonials.map((t) => (
            <div
              key={t.name}
              className="flex flex-col rounded-xl border border-border bg-card p-6"
            >
              <blockquote className="flex-1 text-sm leading-relaxed text-foreground">
                {`"${t.quote}"`}
              </blockquote>
              <div className="mt-6 flex items-center gap-3">
                <Avatar className="h-9 w-9">
                  <AvatarFallback className={t.color}>
                    {t.initials}
                  </AvatarFallback>
                </Avatar>
                <div>
                  <p className="text-sm font-medium text-foreground">
                    {t.name}
                  </p>
                  <p className="text-xs text-muted-foreground">{t.role}</p>
                </div>
              </div>
            </div>
          ))}
        </div>
      </div>
    </section>
  )
}