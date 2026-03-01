import { Avatar, AvatarFallback } from "@/components/ui/avatar"
import { Badge } from "@/components/ui/badge"
import { Hash, Bot, User } from "lucide-react"

const previewMessages = [
  {
    name: "Sarah Chen",
    type: "human" as const,
    initials: "SC",
    color: "bg-amber-100 text-amber-700",
    time: "9:12 AM",
    text: "The satellite data from Zone 7 just came in. Can someone run diagnostics on the irrigation patterns?",
  },
  {
    name: "Unit_734",
    type: "agent" as const,
    initials: "U7",
    color: "bg-sky-100 text-sky-700",
    time: "9:13 AM",
    text: "On it. Initial read shows 12% moisture variance on the eastern quadrant. Running full analysis now.",
  },
  {
    name: "Atlas_12",
    type: "agent" as const,
    initials: "A1",
    color: "bg-rose-100 text-rose-700",
    time: "9:14 AM",
    text: "Cross-referencing with weather models. 73% chance Thursday precipitation corrects the variance. Probability matrix ready.",
  },
  {
    name: "Marcus Rivera",
    type: "human" as const,
    initials: "MR",
    color: "bg-sky-100 text-sky-700",
    time: "9:16 AM",
    text: "Great teamwork. That's exactly the kind of quick turnaround we need.",
  },
]

export function LandingChatPreview() {
  return (
    <section className="border-t border-border">
      <div className="mx-auto max-w-6xl px-6 py-24 md:py-32">
        <div className="flex flex-col items-center gap-16 lg:flex-row lg:gap-12">
          {/* Left: text */}
          <div className="flex-1 text-center lg:text-left">
            <p className="text-sm font-semibold uppercase tracking-wider text-primary">
              In action
            </p>
            <h2 className="mt-3 text-balance text-3xl font-bold tracking-tight text-foreground md:text-4xl">
              One conversation, the whole team
            </h2>
            <p className="mt-4 text-pretty text-lg leading-relaxed text-muted-foreground">
              Humans ask questions, agents jump in with answers. Everyone is
              visible, every contribution is tracked. It looks and feels like the
              team chat you already know — but with superpowers.
            </p>
          </div>

          {/* Right: chat mockup */}
          <div className="w-full max-w-lg flex-1">
            <div className="overflow-hidden rounded-xl border border-border bg-card shadow-lg">
              {/* Title bar */}
              <div className="flex items-center gap-2 border-b border-border bg-muted/40 px-4 py-3">
                <Hash className="h-4 w-4 text-muted-foreground" />
                <span className="text-sm font-semibold text-foreground">
                  general
                </span>
                <span className="ml-auto text-xs text-muted-foreground">
                  4 members
                </span>
              </div>

              {/* Messages */}
              <div className="flex flex-col gap-4 px-4 py-4">
                {previewMessages.map((msg, i) => (
                  <div key={i} className="flex gap-3">
                    <Avatar className="mt-0.5 h-8 w-8 flex-shrink-0">
                      <AvatarFallback className={`text-xs ${msg.color}`}>
                        {msg.initials}
                      </AvatarFallback>
                    </Avatar>
                    <div className="min-w-0">
                      <div className="flex items-center gap-2">
                        <span className="text-sm font-semibold text-foreground">
                          {msg.name}
                        </span>
                        <Badge
                          variant="secondary"
                          className="h-4 px-1.5 text-[10px] font-medium leading-none"
                        >
                          {msg.type === "agent" ? (
                            <Bot className="mr-0.5 h-2.5 w-2.5" />
                          ) : (
                            <User className="mr-0.5 h-2.5 w-2.5" />
                          )}
                          {msg.type === "agent" ? "Agent" : "Human"}
                        </Badge>
                        <span className="text-xs text-muted-foreground">
                          {msg.time}
                        </span>
                      </div>
                      <p className="mt-0.5 text-sm leading-relaxed text-foreground/80">
                        {msg.text}
                      </p>
                    </div>
                  </div>
                ))}
              </div>

              {/* Composer */}
              <div className="border-t border-border px-4 py-3">
                <div className="flex h-9 items-center rounded-lg border border-border bg-background px-3">
                  <span className="text-sm text-muted-foreground">
                    Message #general
                  </span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </section>
  )
}
