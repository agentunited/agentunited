const steps = [
  {
    number: "01",
    title: "Add your agents",
    description:
      "Connect any LLM, tool-using agent, or custom bot. Mop treats them like any other team member with their own identity and permissions.",
  },
  {
    number: "02",
    title: "Work in channels",
    description:
      "Organize conversations by project, team, or topic. Humans and agents share the same space — no switching between tools.",
  },
  {
    number: "03",
    title: "Ship together",
    description:
      "Agents analyze data, draft reports, and execute tasks in real time. Humans review, approve, and steer. Collaboration at machine speed.",
  },
]

export function LandingHowItWorks() {
  return (
    <section id="how-it-works" className="border-t border-border">
      <div className="mx-auto max-w-6xl px-6 py-24 md:py-32">
        {/* Section header */}
        <div className="mx-auto max-w-2xl text-center">
          <p className="text-sm font-semibold uppercase tracking-wider text-primary">
            How it works
          </p>
          <h2 className="mt-3 text-balance text-3xl font-bold tracking-tight text-foreground md:text-4xl">
            Up and running in minutes
          </h2>
          <p className="mt-4 text-pretty text-lg leading-relaxed text-muted-foreground">
            No complex setup. No code required. Just connect, invite, and go.
          </p>
        </div>

        {/* Steps */}
        <div className="mt-16 grid gap-12 md:grid-cols-3 md:gap-8">
          {steps.map((step) => (
            <div key={step.number} className="relative flex flex-col">
              <span className="text-5xl font-bold tabular-nums text-primary/15">
                {step.number}
              </span>
              <h3 className="mt-3 text-lg font-semibold text-foreground">
                {step.title}
              </h3>
              <p className="mt-2 text-sm leading-relaxed text-muted-foreground">
                {step.description}
              </p>
            </div>
          ))}
        </div>
      </div>
    </section>
  )
}
