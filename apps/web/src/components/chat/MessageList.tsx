import { MessageItem } from "./MessageItem"
import type { Message } from "../../types/chat"

interface MessageListProps {
  messages: Message[]
  dateLabel?: string
}

// getUserType is no longer needed — type comes from API via msg.authorType

// Helper function to format timestamp
function formatTimestamp(timestamp: string): string {
  try {
    const date = new Date(timestamp)
    return date.toLocaleTimeString('en-US', { 
      hour: 'numeric', 
      minute: '2-digit',
      hour12: true 
    })
  } catch {
    return timestamp
  }
}

export function MessageList({ messages, dateLabel = "Today" }: MessageListProps) {
  return (
    <div className="flex-1 overflow-y-auto">
      <div className="flex flex-col py-2">
        {/* Date divider */}
        <div className="flex items-center gap-3 px-5 py-3">
          <div className="h-px flex-1 bg-border" />
          <span className="text-xs font-medium text-muted-foreground">{dateLabel}</span>
          <div className="h-px flex-1 bg-border" />
        </div>

        {messages.map((msg) => (
          <MessageItem 
            key={msg.id}
            id={msg.id}
            author={{
              name: msg.author,
              type: msg.authorType
            }}
            content={msg.text}
            timestamp={formatTimestamp(msg.timestamp)}
          />
        ))}
      </div>
    </div>
  )
}