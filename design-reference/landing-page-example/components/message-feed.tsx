"use client"

import Image from "next/image"
import { messages, getMemberById } from "@/lib/chat-data"
import type { Message, Member } from "@/lib/chat-data"
import { cn } from "@/lib/utils"
import { Bot, User } from "lucide-react"

function Avatar({ member }: { member: Member }) {
  const initials =
    member.type === "agent"
      ? member.name.slice(0, 2).toUpperCase()
      : member.name
          .split(" ")
          .map((n) => n[0])
          .join("")
          .toUpperCase()

  return (
    <div
      className={cn(
        "flex h-9 w-9 shrink-0 items-center justify-center rounded-lg text-xs font-bold",
        member.avatarColor
      )}
    >
      {initials}
    </div>
  )
}

function MemberBadge({ type }: { type: "human" | "agent" }) {
  return (
    <span
      className={cn(
        "inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-[10px] font-semibold uppercase tracking-wide",
        type === "human"
          ? "bg-accent/25 text-accent-foreground"
          : "bg-primary/10 text-primary"
      )}
    >
      {type === "human" ? (
        <User className="h-2.5 w-2.5" />
      ) : (
        <Bot className="h-2.5 w-2.5" />
      )}
      {type}
    </span>
  )
}

function ChatMessage({ message }: { message: Message }) {
  const member = getMemberById(message.senderId)
  if (!member) return null

  return (
    <div className="group flex gap-3 rounded-lg px-5 py-3 transition-colors hover:bg-muted/50">
      <Avatar member={member} />
      <div className="flex min-w-0 flex-1 flex-col">
        <div className="flex items-center gap-2">
          <span className="text-sm font-semibold text-foreground">{member.name}</span>
          <MemberBadge type={member.type} />
          <span className="text-xs text-muted-foreground">{message.timestamp}</span>
        </div>
        <p className="mt-1 text-sm leading-relaxed text-foreground/85">{message.content}</p>
        {message.attachment && (
          <div className="mt-3 overflow-hidden rounded-lg border border-border">
            <Image
              src={message.attachment.src}
              alt={message.attachment.alt}
              width={520}
              height={300}
              className="w-full max-w-[520px] object-cover"
              style={{ height: "auto" }}
            />
            <div className="bg-muted/50 px-3 py-2">
              <p className="text-xs text-muted-foreground">{message.attachment.alt}</p>
            </div>
          </div>
        )}
      </div>
    </div>
  )
}

export function MessageFeed() {
  return (
    <div className="flex-1 overflow-y-auto">
      <div className="flex flex-col py-2">
        {/* Date divider */}
        <div className="flex items-center gap-3 px-5 py-3">
          <div className="h-px flex-1 bg-border" />
          <span className="text-xs font-medium text-muted-foreground">Today</span>
          <div className="h-px flex-1 bg-border" />
        </div>

        {messages.map((msg) => (
          <ChatMessage key={msg.id} message={msg} />
        ))}
      </div>
    </div>
  )
}
