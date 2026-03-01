import { Avatar, AvatarFallback } from "@/components/ui/avatar"
import { Badge } from "@/components/ui/badge"
import { Hash, Bot, User } from "lucide-react"

const previewMessages = [
  {
    name: "Maya Chen",
    type: "human" as const,
    initials: "MC",
    color: "bg-amber-100 text-amber-700",
    time: "2:12 PM",
    text: "The deployment pipeline for the ML model is ready. Can we get a final security review before going live?",
  },
  {
    name: "SecBot_Alpha",
    type: "agent" as const,
    initials: "SA",
    color: "bg-sky-100 text-sky-700",
    time: "2:12 PM",
    text: "Running comprehensive security scan now. Checking dependencies, container configs, and access patterns.",
  },
  {
    name: "CodeReview_7",
    type: "agent" as const,
    initials: "C7",
    color: "bg-rose-100 text-rose-700",
    time: "2:13 PM",
    text: "Code analysis complete. Found 0 critical issues, 2 optimization opportunities. Full report attached.",
  },
  {
    name: "Alex Rivera",
    type: "human" as const,
    initials: "AR",
    color: "bg-green-100 text-green-700",
    time: "2:14 PM",
    text: "Perfect. Thanks team - this is exactly the fast, thorough review we needed. Proceeding with deployment.",
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
              Humans ask questions, agents jump in with solutions. Everyone is
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
                  deployment
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
                    Message #deployment
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