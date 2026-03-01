"use client"

import { Hash, Phone, Pin, Users } from "lucide-react"
import { members } from "@/lib/chat-data"
import { cn } from "@/lib/utils"

function MiniAvatar({ name, avatarColor, type }: { name: string; avatarColor: string; type: string }) {
  const initials = type === "agent"
    ? name.slice(0, 2).toUpperCase()
    : name.split(" ").map((n) => n[0]).join("").toUpperCase()

  return (
    <div
      className={cn(
        "flex h-7 w-7 items-center justify-center rounded-full text-[10px] font-bold ring-2 ring-card",
        avatarColor
      )}
      title={name}
    >
      {initials}
    </div>
  )
}

export function ChatHeader() {
  const activeMembers = members.filter((m) => m.online)

  return (
    <header className="flex items-center justify-between border-b border-border bg-card px-5 py-3">
      <div className="flex items-center gap-3">
        <div className="flex items-center gap-1.5">
          <Hash className="h-4.5 w-4.5 text-muted-foreground" />
          <h1 className="text-base font-semibold text-card-foreground">general</h1>
        </div>
        <span className="text-xs text-muted-foreground">
          Team-wide collaboration channel
        </span>
      </div>

      <div className="flex items-center gap-4">
        {/* Active member cluster */}
        <div className="flex -space-x-2">
          {activeMembers.map((m) => (
            <MiniAvatar key={m.id} name={m.name} avatarColor={m.avatarColor} type={m.type} />
          ))}
          {members.length > activeMembers.length && (
            <div className="flex h-7 w-7 items-center justify-center rounded-full bg-muted text-[10px] font-bold text-muted-foreground ring-2 ring-card">
              +{members.length - activeMembers.length}
            </div>
          )}
        </div>

        <div className="flex items-center gap-1">
          <button className="rounded-md p-1.5 text-muted-foreground transition-colors hover:bg-muted hover:text-foreground" aria-label="Members">
            <Users className="h-4 w-4" />
          </button>
          <button className="rounded-md p-1.5 text-muted-foreground transition-colors hover:bg-muted hover:text-foreground" aria-label="Pin">
            <Pin className="h-4 w-4" />
          </button>
          <button className="rounded-md p-1.5 text-muted-foreground transition-colors hover:bg-muted hover:text-foreground" aria-label="Call">
            <Phone className="h-4 w-4" />
          </button>
        </div>
      </div>
    </header>
  )
}
