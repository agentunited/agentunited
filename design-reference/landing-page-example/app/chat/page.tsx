import { ChatSidebar } from "@/components/chat-sidebar"
import { ChatHeader } from "@/components/chat-header"
import { MessageFeed } from "@/components/message-feed"
import { MessageComposer } from "@/components/message-composer"

export default function ChatPage() {
  return (
    <div className="flex h-dvh overflow-hidden bg-background">
      <ChatSidebar />
      <main className="flex min-w-0 flex-1 flex-col">
        <ChatHeader />
        <MessageFeed />
        <MessageComposer />
      </main>
    </div>
  )
}
