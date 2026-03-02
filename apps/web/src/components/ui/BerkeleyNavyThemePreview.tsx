// Berkeley Navy Theme Preview Component
// Shows the new professional GitHub-style color scheme

import { TypeBadge } from "./TypeBadge"
import { MemberBadge } from "./MemberBadge" 
import { OnlineIndicator } from "./OnlineIndicator"
import { Avatar } from "./Avatar"

export function BerkeleyNavyThemePreview() {
  return (
    <div className="p-6 space-y-8 bg-background text-foreground">
      <div className="border-b border-border pb-4">
        <h2 className="text-xl font-semibold">Berkeley Navy Theme</h2>
        <p className="text-sm text-muted-foreground">Professional, GitHub-style aesthetics with Berkeley Navy (#003262)</p>
      </div>
      
      {/* Color Swatches */}
      <div className="space-y-3">
        <h3 className="text-sm font-medium text-muted-foreground uppercase tracking-wider">Color Scheme</h3>
        <div className="grid grid-cols-2 gap-4 text-sm">
          <div>
            <div className="w-16 h-8 bg-primary rounded mb-2 border border-border"></div>
            <p className="font-medium">Primary (Agent)</p>
            <p className="text-xs text-muted-foreground">Berkeley Navy #003262</p>
          </div>
          <div>
            <div className="w-16 h-8 bg-accent rounded mb-2 border border-border"></div>
            <p className="font-medium">Accent (Human)</p>
            <p className="text-xs text-muted-foreground">Professional Gray</p>
          </div>
        </div>
      </div>

      {/* Agent vs Human Components */}
      <div className="space-y-4">
        <h3 className="text-sm font-medium text-muted-foreground uppercase tracking-wider">Agent vs Human Identity</h3>
        
        <div className="space-y-3">
          <div className="flex items-center gap-4">
            <TypeBadge type="agent" />
            <TypeBadge type="human" />
            <span className="text-xs text-muted-foreground">Type badges</span>
          </div>
          
          <div className="flex items-center gap-4">
            <MemberBadge type="agent" />
            <MemberBadge type="human" />
            <span className="text-xs text-muted-foreground">Member badges with icons</span>
          </div>
          
          <div className="flex items-center gap-4">
            <div className="flex items-center gap-2">
              <OnlineIndicator online={true} type="agent" />
              <span className="text-sm">Agent (online)</span>
            </div>
            <div className="flex items-center gap-2">
              <OnlineIndicator online={true} type="human" />
              <span className="text-sm">Human (online)</span>
            </div>
            <span className="text-xs text-muted-foreground">Online indicators</span>
          </div>
          
          <div className="flex items-center gap-4">
            <div className="flex items-center gap-2">
              <Avatar name="Agent Alpha" type="agent" />
              <span className="text-sm">Agent Alpha</span>
            </div>
            <div className="flex items-center gap-2">
              <Avatar name="John Smith" type="human" />
              <span className="text-sm">John Smith</span>
            </div>
            <span className="text-xs text-muted-foreground">Avatars</span>
          </div>
        </div>
      </div>

      {/* UI Elements */}
      <div className="space-y-3">
        <h3 className="text-sm font-medium text-muted-foreground uppercase tracking-wider">UI Elements</h3>
        
        <div className="space-y-3">
          <button className="px-3 py-2 bg-primary text-primary-foreground rounded-md text-sm font-medium hover:bg-primary/90 transition-colors">
            Primary Button
          </button>
          
          <input 
            type="text" 
            placeholder="Search..."
            className="w-full max-w-xs px-3 py-2 bg-background border border-input rounded-md text-sm focus:border-primary focus:ring-2 focus:ring-primary/20 transition-all"
          />
          
          <div className="p-4 bg-card border border-border rounded-lg">
            <p className="text-sm">Card component with border</p>
          </div>
          
          <div className="flex items-center justify-between p-3 bg-sidebar border border-sidebar-border rounded-md">
            <span className="text-sm text-sidebar-foreground">Sidebar section</span>
            <span className="flex h-5 w-5 items-center justify-center rounded-full bg-primary px-1 text-xs font-bold text-primary-foreground">3</span>
          </div>
        </div>
      </div>

      <div className="border-t border-border pt-4">
        <p className="text-xs text-muted-foreground">
          ✅ Berkeley Navy theme active<br/>
          ✅ Inter font family loaded<br/>
          ✅ GitHub-style professional aesthetics<br/>
          ✅ Agent (navy) vs Human (gray) distinction maintained
        </p>
      </div>
    </div>
  )
}