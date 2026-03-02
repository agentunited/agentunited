import { useCallback } from "react"
import { MessageItem } from "./MessageItem"
import type { Message } from "../../types/chat"

interface MessageListProps {
  messages: Message[]
  channelId: string
  dateLabel?: string
  onMessageUpdated?: (messageId: string, newContent: string) => void
  onMessageDeleted?: (messageId: string) => void
}

// formatTimestamp removed — MessageItem handles its own formatting

export function MessageList({ 
  messages, 
  channelId, 
  dateLabel = "Today",
  onMessageUpdated,
  onMessageDeleted
}: MessageListProps) {

  const handleMessageUpdated = useCallback((messageId: string, newContent: string) => {
    onMessageUpdated?.(messageId, newContent);
  }, [onMessageUpdated]);

  const handleMessageDeleted = useCallback((messageId: string) => {
    onMessageDeleted?.(messageId);
  }, [onMessageDeleted]);

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
            channelId={channelId}
            author={{
              name: msg.author,
              type: msg.authorType
            }}
            content={msg.text}
            timestamp={msg.timestamp}
            editedAt={msg.editedAt}
            isOwnMessage={msg.isOwnMessage}
            attachment={msg.attachmentUrl && msg.attachmentName ? {
              url: msg.attachmentUrl,
              name: msg.attachmentName
            } : undefined}
            onMessageUpdated={handleMessageUpdated}
            onMessageDeleted={handleMessageDeleted}
          />
        ))}
      </div>
    </div>
  )
}