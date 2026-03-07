import { Hash, Menu, MoreVertical, Search, Users } from "lucide-react"
import { cn } from "../../lib/utils"

interface ChatHeaderProps {
  channelName: string
  topic?: string
  isDM?: boolean
  onToggleMembers?: () => void
  showMembersPanel?: boolean
  onToggleSidebar?: () => void
  sidebarOpen?: boolean
}

export function ChatHeader({
  channelName,
  topic,
  isDM = false,
  onToggleMembers,
  showMembersPanel = false,
  onToggleSidebar,
  sidebarOpen = false,
}: ChatHeaderProps) {
  return (
    <header className="sticky top-0 z-10 flex items-center gap-3 border-b border-border bg-white/80 px-5 py-3 backdrop-blur-sm dark:bg-card/90">
      {onToggleSidebar && (
        <button
          onClick={onToggleSidebar}
          className="rounded-lg p-1.5 text-muted-foreground transition-colors hover:bg-muted hover:text-foreground md:hidden"
          aria-label="Open navigation"
          aria-expanded={sidebarOpen}
        >
          <Menu className="h-5 w-5" />
        </button>
      )}

      {isDM ? (
        <div className="flex h-8 w-8 items-center justify-center rounded-full bg-emerald-500/15 text-xs font-bold text-emerald-700 ring-1 ring-emerald-500/30 dark:text-emerald-300">
          {channelName.charAt(0).toUpperCase()}
        </div>
      ) : (
        <Hash className="h-4 w-4 shrink-0 text-muted-foreground" />
      )}

      <div className="min-w-0 flex-1">
        <h2 className="text-sm font-semibold text-foreground">{channelName}</h2>
        {topic && <p className="truncate text-[11px] text-muted-foreground/70">{topic}</p>}
      </div>

      <div className="flex items-center gap-1">
        <button
          className="rounded-lg p-1.5 text-muted-foreground transition-colors hover:bg-muted hover:text-foreground"
          aria-label="Search messages"
        >
          <Search className="h-4 w-4" />
        </button>

        {!isDM && onToggleMembers && (
          <button
            onClick={onToggleMembers}
            className={cn(
              "rounded-lg p-1.5 transition-colors",
              showMembersPanel
                ? "bg-emerald-500/10 text-emerald-700 hover:bg-emerald-500/15 dark:text-emerald-300"
                : "text-muted-foreground hover:bg-muted hover:text-foreground"
            )}
            aria-label={showMembersPanel ? "Hide member list" : "Show member list"}
          >
            <Users className="h-4 w-4" />
          </button>
        )}

        <button
          className="rounded-lg p-1.5 text-muted-foreground transition-colors hover:bg-muted hover:text-foreground"
          aria-label={isDM ? "Conversation options" : "Channel options"}
        >
          <MoreVertical className="h-4 w-4" />
        </button>
      </div>
    </header>
  )
}
