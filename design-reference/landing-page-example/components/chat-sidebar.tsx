"use client"

import { Search, ChevronDown, Hash, Plus } from "lucide-react"
import { channels, dmList, getMemberById } from "@/lib/chat-data"
import { cn } from "@/lib/utils"

function OnlineIndicator({ online }: { online: boolean }) {
  return (
    <span
      className={cn(
        "inline-block h-2 w-2 rounded-full shrink-0",
        online ? "bg-primary" : "bg-muted-foreground/30"
      )}
    />
  )
}

function TypeBadge({ type }: { type: "human" | "agent" }) {
  return (
    <span
      className={cn(
        "ml-auto text-[10px] font-medium tracking-wide uppercase rounded-full px-1.5 py-0.5 shrink-0",
        type === "human"
          ? "bg-accent/30 text-accent-foreground"
          : "bg-primary/10 text-primary"
      )}
    >
      {type === "human" ? "Human" : "Agent"}
    </span>
  )
}

export function ChatSidebar() {
  return (
    <aside className="flex h-full w-64 shrink-0 flex-col border-r border-sidebar-border bg-sidebar text-sidebar-foreground">
      {/* Workspace Header */}
      <div className="flex items-center gap-2 border-b border-sidebar-border px-4 py-3.5">
        <div className="flex h-8 w-8 items-center justify-center rounded-md bg-primary text-primary-foreground font-bold text-sm">
          M
        </div>
        <div className="flex flex-col leading-tight">
          <span className="text-sm font-semibold">Mop UI</span>
          <span className="text-[11px] text-muted-foreground">Workspace</span>
        </div>
        <ChevronDown className="ml-auto h-4 w-4 text-muted-foreground" />
      </div>

      {/* Search */}
      <div className="px-3 py-2.5">
        <div className="flex items-center gap-2 rounded-md bg-muted px-3 py-1.5">
          <Search className="h-3.5 w-3.5 text-muted-foreground" />
          <input
            type="text"
            placeholder="Search..."
            className="w-full bg-transparent text-sm outline-none placeholder:text-muted-foreground"
          />
        </div>
      </div>

      {/* Channels */}
      <div className="px-3 pt-2 pb-1">
        <div className="flex items-center justify-between px-1">
          <span className="text-[11px] font-semibold uppercase tracking-wider text-muted-foreground">
            Channels
          </span>
          <button className="rounded p-0.5 hover:bg-sidebar-accent" aria-label="Add channel">
            <Plus className="h-3.5 w-3.5 text-muted-foreground" />
          </button>
        </div>
        <nav className="mt-1 flex flex-col gap-0.5">
          {channels.map((ch) => (
            <button
              key={ch.id}
              className={cn(
                "flex items-center gap-2 rounded-md px-2 py-1.5 text-sm transition-colors",
                ch.id === "general"
                  ? "bg-sidebar-accent text-sidebar-accent-foreground font-medium"
                  : "text-sidebar-foreground/80 hover:bg-sidebar-accent/60"
              )}
            >
              <Hash className="h-3.5 w-3.5 shrink-0 text-muted-foreground" />
              <span>{ch.name}</span>
              {ch.unread && (
                <span className="ml-auto flex h-4.5 min-w-[18px] items-center justify-center rounded-full bg-primary px-1 text-[10px] font-bold text-primary-foreground">
                  {ch.unread}
                </span>
              )}
            </button>
          ))}
        </nav>
      </div>

      {/* Direct Messages */}
      <div className="flex-1 overflow-y-auto px-3 pt-4 pb-3">
        <div className="flex items-center justify-between px-1">
          <span className="text-[11px] font-semibold uppercase tracking-wider text-muted-foreground">
            Direct Messages
          </span>
          <button className="rounded p-0.5 hover:bg-sidebar-accent" aria-label="New message">
            <Plus className="h-3.5 w-3.5 text-muted-foreground" />
          </button>
        </div>
        <nav className="mt-1 flex flex-col gap-0.5">
          {dmList.map((dm) => {
            const member = getMemberById(dm.memberId)
            if (!member) return null
            return (
              <button
                key={dm.memberId}
                className="flex items-center gap-2.5 rounded-md px-2 py-1.5 text-sm text-sidebar-foreground/80 transition-colors hover:bg-sidebar-accent/60"
              >
                <OnlineIndicator online={member.online} />
                <span className="truncate">{member.name}</span>
                <TypeBadge type={member.type} />
                {dm.unread && (
                  <span className="flex h-4.5 min-w-[18px] items-center justify-center rounded-full bg-primary px-1 text-[10px] font-bold text-primary-foreground">
                    {dm.unread}
                  </span>
                )}
              </button>
            )
          })}
        </nav>
      </div>
    </aside>
  )
}
