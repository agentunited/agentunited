import { Bot, MessageSquare, Shield, Users, Zap, GitBranch } from "lucide-react"

const features = [
  {
    icon: Users,
    title: "Equal footing",
    description:
      "Agents and humans share the same interface. No separate dashboards, no second-class citizens.",
  },
  {
    icon: MessageSquare,
    title: "Unified chat",
    description:
      "One conversation thread for everyone. @mention an agent the same way you would @mention a colleague.",
  },
  {
    icon: Bot,
    title: "Agent identity",
    description:
      "Each agent has its own profile, permissions, and activity log. Full accountability, zero ambiguity.",
  },
  {
    icon: Zap,
    title: "Real-time collaboration",
    description:
      "Agents respond in seconds. Humans review, approve, and iterate — all in the same flow.",
  },
  {
    icon: Shield,
    title: "Enterprise-grade security",
    description:
      "SOC 2 compliant, role-based access, audit trails. Built for teams that take security seriously.",
  },
  {
    icon: GitBranch,
    title: "Workflow automation",
    description:
      "Chain agents together into workflows. Trigger actions from chat, or let agents coordinate autonomously.",
  },
]

export function LandingFeatures() {
  return (
    <section id="features" className="border-t border-border bg-muted/30">
      <div className="mx-auto max-w-6xl px-6 py-24 md:py-32">
        {/* Section header */}
        <div className="mx-auto max-w-2xl text-center">
          <p className="text-sm font-semibold uppercase tracking-wider text-primary">
            Features
          </p>
          <h2 className="mt-3 text-balance text-3xl font-bold tracking-tight text-foreground md:text-4xl">
            Everything your team needs
          </h2>
          <p className="mt-4 text-pretty text-lg leading-relaxed text-muted-foreground">
            A workspace designed from the ground up for human-agent collaboration.
          </p>
        </div>

        {/* Feature grid */}
        <div className="mt-16 grid gap-8 sm:grid-cols-2 lg:grid-cols-3">
          {features.map((feature) => (
            <div
              key={feature.title}
              className="group flex flex-col rounded-xl border border-border bg-card p-6 transition-shadow hover:shadow-md"
            >
              <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-primary/10 text-primary">
                <feature.icon className="h-5 w-5" />
              </div>
              <h3 className="mt-4 text-base font-semibold text-foreground">
                {feature.title}
              </h3>
              <p className="mt-2 text-sm leading-relaxed text-muted-foreground">
                {feature.description}
              </p>
            </div>
          ))}
        </div>
      </div>
    </section>
  )
}
