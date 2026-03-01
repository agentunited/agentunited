"use client"

import { useState } from "react"
import { Plus, Send, Smile, Paperclip, AtSign } from "lucide-react"
import { cn } from "@/lib/utils"

export function MessageComposer() {
  const [value, setValue] = useState("")

  return (
    <div className="border-t border-border bg-card px-5 py-3">
      <div className="flex items-end gap-2 rounded-xl border border-input bg-background px-3 py-2 focus-within:border-primary/40 focus-within:ring-2 focus-within:ring-primary/10 transition-all">
        <button
          className="mb-0.5 shrink-0 rounded-md p-1.5 text-muted-foreground transition-colors hover:bg-muted hover:text-foreground"
          aria-label="Attach file"
        >
          <Plus className="h-4.5 w-4.5" />
        </button>

        <textarea
          value={value}
          onChange={(e) => setValue(e.target.value)}
          placeholder="Message #general"
          rows={1}
          className="max-h-32 min-h-[24px] flex-1 resize-none bg-transparent text-sm leading-relaxed text-foreground outline-none placeholder:text-muted-foreground"
          onInput={(e) => {
            const target = e.target as HTMLTextAreaElement
            target.style.height = "auto"
            target.style.height = target.scrollHeight + "px"
          }}
        />

        <div className="mb-0.5 flex shrink-0 items-center gap-0.5">
          <button
            className="rounded-md p-1.5 text-muted-foreground transition-colors hover:bg-muted hover:text-foreground"
            aria-label="Mention someone"
          >
            <AtSign className="h-4 w-4" />
          </button>
          <button
            className="rounded-md p-1.5 text-muted-foreground transition-colors hover:bg-muted hover:text-foreground"
            aria-label="Add attachment"
          >
            <Paperclip className="h-4 w-4" />
          </button>
          <button
            className="rounded-md p-1.5 text-muted-foreground transition-colors hover:bg-muted hover:text-foreground"
            aria-label="Add emoji"
          >
            <Smile className="h-4 w-4" />
          </button>
          <button
            className={cn(
              "ml-1 rounded-lg p-2 transition-colors",
              value.trim()
                ? "bg-primary text-primary-foreground hover:bg-primary/90"
                : "bg-muted text-muted-foreground cursor-not-allowed"
            )}
            aria-label="Send message"
            disabled={!value.trim()}
          >
            <Send className="h-4 w-4" />
          </button>
        </div>
      </div>
      <p className="mt-1.5 px-1 text-[11px] text-muted-foreground">
        <span className="font-medium">Shift + Enter</span> for a new line
      </p>
    </div>
  )
}
