import { useState, useCallback, useRef, useEffect } from "react"
import { Search, ChevronDown, Hash, Plus, X } from "lucide-react"
import { cn } from "../../lib/utils"
import { OnlineIndicator } from "../ui/OnlineIndicator"
import { TypeBadge } from "../ui/TypeBadge"
import { ThemeToggle } from "../ui/ThemeToggle"
import { ChannelContextMenu } from "./ChannelContextMenu"
import type { Channel } from "../../types/chat"

interface DirectMessage {
  id: string
  name: string
  type: "human" | "agent"
  online: boolean
  unread?: number
}

interface ChatSidebarProps {
  channels: Channel[]
  directMessages: DirectMessage[]
  activeChannelId?: string
  isOpen?: boolean
  onChannelSelect: (id: string) => void
  onDMSelect: (id: string) => void
  onCreateChannel?: () => void
  onNewDM?: () => void
  onSearch?: (query: string) => void
  onChannelUpdate?: (channel: Channel) => void
  onChannelDelete?: (channelId: string) => void
}

function getDisplayName(name: string): string {
  if (name.includes('@')) {
    const localPart = name.split('@')[0] ?? name
    return localPart
      .replace(/[._-]/g, ' ')
      .replace(/\b\w/g, (c) => c.toUpperCase())
  }

  return name
}

export function ChatSidebar({
  channels,
  directMessages,
  activeChannelId,
  isOpen = false,
  onChannelSelect,
  onDMSelect,
  onCreateChannel,
  onNewDM,
  onSearch,
  onChannelUpdate,
  onChannelDelete,
}: ChatSidebarProps) {
  const [searchQuery, setSearchQuery] = useState("")
  const searchInputRef = useRef<HTMLInputElement>(null)

  useEffect(() => {
    const handleFocusSearch = () => {
      searchInputRef.current?.focus()
    }
    window.addEventListener('focus-chat-search', handleFocusSearch)
    return () => window.removeEventListener('focus-chat-search', handleFocusSearch)
  }, [])

  const handleSearch = useCallback((e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === "Enter" && searchQuery.trim() && onSearch) {
      onSearch(searchQuery.trim())
    }
  }, [searchQuery, onSearch])

  const handleChannelUpdate = (updatedChannel: Channel) => {
    onChannelUpdate?.(updatedChannel)
  }

  const handleChannelDelete = (channelId: string) => {
    onChannelDelete?.(channelId)
  }

  const totalUnreadCount = channels.reduce((sum, ch) => sum + (ch.unread || 0), 0) +
    directMessages.reduce((sum, dm) => sum + (dm.unread || 0), 0)

  return (
    <aside
      aria-label="Navigation"
      className={cn(
        "flex h-full w-64 flex-col border-r border-sidebar-border bg-sidebar text-sidebar-foreground",
        "fixed inset-y-0 left-0 z-40 transition-transform duration-200 ease-in-out",
        isOpen ? "translate-x-0" : "-translate-x-full",
        "md:static md:z-auto md:shrink-0 md:translate-x-0 md:!transform-none"
      )}
    >
      <div className="flex items-center gap-2.5 border-b border-sidebar-border px-4 py-3.5">
        <div className="relative flex h-8 w-8 items-center justify-center rounded-full border border-emerald-500/50 shadow-[0_0_10px_rgba(16,185,129,0.35)]">
          <div className="h-2 w-2 rounded-full bg-emerald-500" />
          {totalUnreadCount > 0 && (
            <span className="absolute -top-0.5 -right-0.5 h-2 w-2 rounded-full bg-emerald-500 ring-2 ring-sidebar" />
          )}
        </div>

        <div className="flex flex-col leading-tight">
          <span className="text-sm font-semibold text-sidebar-foreground">Agent United</span>
          <span className="text-[10px] font-medium tracking-[0.08em] text-emerald-600/70 uppercase">workspace</span>
        </div>

        <div className="ml-auto flex items-center gap-1">
          <ThemeToggle />
          <button className="rounded-md p-1 text-muted-foreground transition-colors hover:bg-sidebar-accent hover:text-sidebar-foreground" aria-label="Workspace options">
            <ChevronDown className="h-3.5 w-3.5" />
          </button>
        </div>
      </div>

      <div className="px-3 py-2.5">
        <div className="flex items-center gap-2 rounded-lg border border-sidebar-border/60 bg-white/70 px-3 py-1.5 shadow-sm transition-all focus-within:border-emerald-400/50 focus-within:ring-2 focus-within:ring-emerald-500/10 dark:bg-background/40">
          <Search className="h-3.5 w-3.5 shrink-0 text-muted-foreground" />
          <input
            type="text"
            placeholder="Search…"
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            onKeyDown={handleSearch}
            ref={searchInputRef}
            className="w-full bg-transparent text-sm outline-none placeholder:text-muted-foreground/60"
          />
          {searchQuery && (
            <button onClick={() => { setSearchQuery("") }} className="text-muted-foreground transition-colors hover:text-foreground" aria-label="Clear search">
              <X className="h-3 w-3" />
            </button>
          )}
        </div>
      </div>

      <div className="px-3 pt-2 pb-1">
        <div className="flex items-center justify-between px-1">
          <span className="text-[10px] font-semibold tracking-[0.14em] text-emerald-700/60 uppercase">Channels</span>
          <button className="rounded p-0.5 transition-colors hover:bg-sidebar-accent" aria-label="Add channel" onClick={onCreateChannel}>
            <Plus className="h-3.5 w-3.5 text-muted-foreground" />
          </button>
        </div>

        <nav className="mt-1 flex flex-col gap-0.5">
          {channels.map((ch) => {
            const hasUnread = (ch.unread || 0) > 0
            const isActive = ch.id === activeChannelId

            return (
              <button
                key={ch.id}
                onClick={() => onChannelSelect(ch.id)}
                className={cn(
                  "group flex w-full items-center gap-2 rounded-lg px-2 py-1.5 text-sm transition-colors",
                  isActive
                    ? "border-l-2 border-emerald-500 bg-emerald-500/10 pl-[6px] font-semibold text-emerald-800 dark:text-emerald-300"
                    : hasUnread
                      ? "font-semibold text-sidebar-foreground hover:bg-sidebar-accent/50"
                      : "text-sidebar-foreground/75 hover:bg-sidebar-accent/50"
                )}
              >
                <Hash className={cn("h-3.5 w-3.5 shrink-0", isActive ? "text-emerald-600" : "text-muted-foreground")} />

                <span className="flex-1 truncate text-left">{ch.name}</span>

                {hasUnread ? (
                  <span className="ml-auto flex h-[18px] min-w-[18px] items-center justify-center rounded-full bg-emerald-500 px-1 text-[10px] font-bold text-white">
                    {ch.unread! > 99 ? '99+' : ch.unread}
                  </span>
                ) : (
                  <ChannelContextMenu
                    channel={ch}
                    onChannelUpdate={handleChannelUpdate}
                    onChannelDelete={handleChannelDelete}
                    className="ml-auto"
                  />
                )}
              </button>
            )
          })}
        </nav>
      </div>

      <div className="flex-1 overflow-y-auto px-3 pt-4 pb-3">
        <div className="flex items-center justify-between px-1">
          <span className="text-[10px] font-semibold tracking-[0.14em] text-emerald-700/60 uppercase">Direct Messages</span>
          <button className="rounded p-0.5 transition-colors hover:bg-sidebar-accent" aria-label="New message" onClick={onNewDM}>
            <Plus className="h-3.5 w-3.5 text-muted-foreground" />
          </button>
        </div>

        <nav className="mt-1 flex flex-col gap-0.5">
          {directMessages.map((dm) => {
            const hasUnread = (dm.unread || 0) > 0
            const isActive = dm.id === activeChannelId
            const displayName = getDisplayName(dm.name)

            return (
              <button
                key={dm.id}
                onClick={() => onDMSelect(dm.id)}
                className={cn(
                  "group flex w-full items-center gap-2 rounded-lg px-2 py-1.5 text-sm transition-colors",
                  isActive
                    ? "bg-emerald-500/10 font-semibold text-emerald-800 dark:text-emerald-300"
                    : hasUnread
                      ? "font-semibold text-sidebar-foreground hover:bg-sidebar-accent/50"
                      : "text-sidebar-foreground/75 hover:bg-sidebar-accent/50"
                )}
              >
                <div className="relative flex h-6 w-6 shrink-0 items-center justify-center rounded-full text-[10px] font-bold">
                  <div className={cn(
                    "flex h-6 w-6 items-center justify-center rounded-full",
                    dm.type === 'agent' ? 'bg-emerald-500/15 text-emerald-700 dark:text-emerald-300' : 'bg-slate-200 text-slate-600 dark:bg-slate-700 dark:text-slate-200'
                  )}>
                    {displayName.charAt(0).toUpperCase()}
                  </div>
                  <OnlineIndicator
                    online={dm.online}
                    type={dm.type}
                    className="absolute -right-0.5 -bottom-0.5 h-2 w-2 ring-2 ring-sidebar"
                  />
                </div>

                <span className="flex-1 truncate text-left">{displayName}</span>

                {dm.type === 'agent' && <TypeBadge type={dm.type} />}

                {hasUnread && (
                  <span className="ml-1 flex h-[18px] min-w-[18px] items-center justify-center rounded-full bg-emerald-500 px-1 text-[10px] font-bold text-white">
                    {dm.unread! > 99 ? '99+' : dm.unread}
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
