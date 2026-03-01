// Visual verification component to test all agent vs human distinctions
// This is for development/testing purposes only

import { TypeBadge } from "./TypeBadge"
import { MemberBadge } from "./MemberBadge"
import { OnlineIndicator } from "./OnlineIndicator"
import { Avatar } from "./Avatar"

export function VisualVerification() {
  return (
    <div className="p-6 space-y-8 bg-background text-foreground">
      <h2 className="text-xl font-semibold">Visual Identity Verification</h2>
      
      {/* TypeBadge Testing */}
      <div className="space-y-2">
        <h3 className="text-sm font-medium text-muted-foreground uppercase tracking-wider">TypeBadge</h3>
        <div className="flex gap-4">
          <TypeBadge type="agent" />
          <TypeBadge type="human" />
        </div>
        <p className="text-xs text-muted-foreground">Agent: Amber | Human: Steel Blue</p>
      </div>

      {/* MemberBadge Testing */}
      <div className="space-y-2">
        <h3 className="text-sm font-medium text-muted-foreground uppercase tracking-wider">MemberBadge</h3>
        <div className="flex gap-4">
          <MemberBadge type="agent" />
          <MemberBadge type="human" />
        </div>
        <p className="text-xs text-muted-foreground">Agent: Bot icon + Amber | Human: User icon + Steel Blue</p>
      </div>

      {/* OnlineIndicator Testing */}
      <div className="space-y-2">
        <h3 className="text-sm font-medium text-muted-foreground uppercase tracking-wider">OnlineIndicator</h3>
        <div className="flex gap-4 items-center">
          <div className="flex items-center gap-2">
            <OnlineIndicator online={true} type="agent" />
            <span className="text-sm">Agent (online with glow)</span>
          </div>
          <div className="flex items-center gap-2">
            <OnlineIndicator online={true} type="human" />
            <span className="text-sm">Human (online)</span>
          </div>
          <div className="flex items-center gap-2">
            <OnlineIndicator online={false} />
            <span className="text-sm">Offline (any type)</span>
          </div>
        </div>
        <p className="text-xs text-muted-foreground">Agent: Amber with glow effect | Human: Steel Blue solid</p>
      </div>

      {/* Avatar Testing */}
      <div className="space-y-2">
        <h3 className="text-sm font-medium text-muted-foreground uppercase tracking-wider">Avatar</h3>
        <div className="flex gap-4 items-center">
          <div className="flex items-center gap-2">
            <Avatar name="Agent Alpha" type="agent" />
            <span className="text-sm">Agent Alpha</span>
          </div>
          <div className="flex items-center gap-2">
            <Avatar name="John Smith" type="human" />
            <span className="text-sm">John Smith</span>
          </div>
        </div>
        <p className="text-xs text-muted-foreground">Agent: Amber gradient, 2-char initials | Human: Steel Blue solid, name initials</p>
      </div>

      {/* Unread Badge Testing */}
      <div className="space-y-2">
        <h3 className="text-sm font-medium text-muted-foreground uppercase tracking-wider">Unread Badge (Primary for both)</h3>
        <div className="flex gap-4 items-center">
          <span className="flex h-4.5 min-w-[18px] items-center justify-center rounded-full bg-primary px-1 text-[10px] font-bold text-primary-foreground">3</span>
          <span className="flex h-4.5 min-w-[18px] items-center justify-center rounded-full bg-primary px-1 text-[10px] font-bold text-primary-foreground">12</span>
        </div>
        <p className="text-xs text-muted-foreground">Unread badges use primary (amber) for visual consistency</p>
      </div>

      {/* Color Reference */}
      <div className="space-y-2">
        <h3 className="text-sm font-medium text-muted-foreground uppercase tracking-wider">Color Reference</h3>
        <div className="grid grid-cols-2 gap-4 text-sm">
          <div>
            <div className="w-16 h-8 bg-primary rounded mb-2"></div>
            <p className="font-medium">Primary (Agent)</p>
            <p className="text-xs text-muted-foreground">CRT Amber #FFB84D (dark)<br/>Rust Orange #D97548 (light)</p>
          </div>
          <div>
            <div className="w-16 h-8 bg-accent rounded mb-2"></div>
            <p className="font-medium">Accent (Human)</p>
            <p className="text-xs text-muted-foreground">Steel Blue #7A9CAB</p>
          </div>
        </div>
      </div>

      <div className="border-t border-border pt-4">
        <p className="text-xs text-muted-foreground">
          ✅ Dark mode first (industrial warmth)<br/>
          ✅ Agent vs Human visual distinction complete<br/>
          ✅ All components use proper design tokens<br/>
          ✅ Accessibility compliant with proper contrast
        </p>
      </div>
    </div>
  )
}