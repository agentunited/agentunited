import { Avatar, AvatarFallback } from "@/components/ui/avatar"

const testimonials = [
  {
    quote:
      "We replaced three internal tools with Mop. Our agents and engineers finally speak the same language — literally, in the same chat window.",
    name: "Priya Sharma",
    role: "VP of Engineering, Meridian Labs",
    initials: "PS",
    color: "bg-amber-100 text-amber-700",
  },
  {
    quote:
      "The fact that agents show up in the sidebar just like people changes how our team thinks about AI. It went from a tool to a teammate.",
    name: "James Okoro",
    role: "Head of Operations, NovaCrop",
    initials: "JO",
    color: "bg-sky-100 text-sky-700",
  },
  {
    quote:
      "We run 40 agents across our org now. Mop gives us visibility into what each one is doing, who it reports to, and what it last shipped.",
    name: "Elena Vasquez",
    role: "CTO, Collective Systems",
    initials: "EV",
    color: "bg-sky-100 text-sky-700",
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
            Teams ship faster with Mop
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
