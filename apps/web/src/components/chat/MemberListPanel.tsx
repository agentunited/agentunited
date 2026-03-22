import { useState, useEffect } from 'react';
import { X, Users, Crown, ShieldCheck } from 'lucide-react';
import { cn } from '../../lib/utils';
import { OnlineIndicator } from '../ui/OnlineIndicator';
import { TypeBadge } from '../ui/TypeBadge';
import { chatApi } from '../../services/chatApi';

interface Member {
  id: string;
  name: string;
  email: string;
  displayName?: string;
  type: 'agent' | 'human';
  role: 'owner' | 'admin' | 'member';
  online?: boolean;
  lastSeenAt?: string;
}

interface MemberListPanelProps {
  isOpen: boolean;
  onClose: () => void;
  channelId: string;
  channelName: string;
}

export function MemberListPanel({ isOpen, onClose, channelId, channelName }: MemberListPanelProps) {
  const [members, setMembers] = useState<Member[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Load members when panel opens or channel changes
  useEffect(() => {
    if (!isOpen || !channelId) return;

    const loadMembers = async () => {
      try {
        setIsLoading(true);
        setError(null);

        const memberData = await chatApi.getMembers(channelId);
        
        // Convert API response to Member format
        const memberList: Member[] = memberData.map(member => {
          const isAgent = member.type
            ? member.type === 'agent'
            : member.email.includes('@agentunited.local');

          const resolvedDisplayName =
            member.display_name ||
            member.name ||
            member.email.split('@')[0] ||
            member.email;

          return {
            id: member.id,
            name: resolvedDisplayName,
            email: member.email,
            displayName: member.display_name,
            type: isAgent ? 'agent' : 'human',
            role: member.role === 'owner' ? 'owner' : member.role === 'admin' ? 'admin' : 'member',
            online: Math.random() > 0.3, // TODO: Replace with real online status from presence API
            lastSeenAt: new Date(Date.now() - Math.random() * 86400000).toISOString() // Random last seen
          };
        });

        // Sort members: online first, then by role (owner, admin, member), then alphabetically
        const sortedMembers = memberList.sort((a, b) => {
          // First sort by online status
          if (a.online !== b.online) {
            return a.online ? -1 : 1;
          }
          
          // Then by role
          const roleOrder = { owner: 0, admin: 1, member: 2 };
          const roleCompare = roleOrder[a.role] - roleOrder[b.role];
          if (roleCompare !== 0) return roleCompare;
          
          // Finally by name alphabetically
          return a.name.localeCompare(b.name);
        });

        setMembers(sortedMembers);
      } catch (error) {
        console.error('Failed to load members:', error);
        setError(error instanceof Error ? error.message : 'Failed to load members');
      } finally {
        setIsLoading(false);
      }
    };

    loadMembers();
  }, [isOpen, channelId]);

  const formatLastSeen = (lastSeenAt: string) => {
    const lastSeen = new Date(lastSeenAt);
    const now = new Date();
    const diffMinutes = Math.floor((now.getTime() - lastSeen.getTime()) / 60000);

    if (diffMinutes < 1) return 'Just now';
    if (diffMinutes < 60) return `${diffMinutes}m ago`;
    if (diffMinutes < 1440) return `${Math.floor(diffMinutes / 60)}h ago`;
    return `${Math.floor(diffMinutes / 1440)}d ago`;
  };

  const onlineCount = members.filter(m => m.online).length;
  const totalCount = members.length;

  if (!isOpen) return null;

  return (
    <aside className="flex h-full w-64 shrink-0 flex-col border-l border-border bg-card">
      {/* Header */}
      <div className="flex items-center justify-between border-b border-border px-4 py-3">
        <div className="flex items-center gap-2">
          <Users className="h-4 w-4 text-muted-foreground" />
          <h3 className="font-semibold text-sm text-foreground">Members</h3>
          <span className="text-xs text-muted-foreground">
            {onlineCount}/{totalCount}
          </span>
        </div>
        <button
          onClick={onClose}
          className="p-1 rounded-md text-muted-foreground hover:text-foreground hover:bg-muted transition-colors"
          aria-label="Close member list"
        >
          <X className="h-4 w-4" />
        </button>
      </div>

      {/* Channel Name */}
      <div className="px-4 py-2 border-b border-border">
        <div className="text-sm font-medium text-foreground">#{channelName}</div>
        <div className="text-xs text-muted-foreground">
          {totalCount} {totalCount === 1 ? 'member' : 'members'}
        </div>
      </div>

      {/* Member List */}
      <div className="flex-1 overflow-y-auto">
        {isLoading ? (
          <div className="flex items-center justify-center py-8">
            <div className="text-sm text-muted-foreground">Loading members...</div>
          </div>
        ) : error ? (
          <div className="px-4 py-6">
            <div className="text-center">
              <div className="text-sm text-destructive mb-2">Failed to load members</div>
              <div className="text-xs text-destructive-foreground">{error}</div>
            </div>
          </div>
        ) : members.length === 0 ? (
          <div className="px-4 py-8 text-center">
            <div className="text-sm text-muted-foreground">No members found</div>
          </div>
        ) : (
          <div className="px-3 py-3">
            {/* Online Members */}
            {onlineCount > 0 && (
              <>
                <div className="mb-2 px-1">
                  <h4 className="text-[11px] font-semibold uppercase tracking-wider text-muted-foreground">
                    Online — {onlineCount}
                  </h4>
                </div>
                <div className="space-y-1 mb-4">
                  {members.filter(member => member.online).map(member => (
                    <div
                      key={member.id}
                      className="flex items-center gap-3 px-2 py-2 rounded-md hover:bg-muted transition-colors"
                    >
                      {/* Avatar with online indicator */}
                      <div className="relative">
                        <div className={cn(
                          "w-7 h-7 rounded-md flex items-center justify-center text-xs font-bold",
                          member.type === 'agent'
                            ? "bg-primary text-primary-foreground"
                            : "bg-accent text-accent-foreground"
                        )}>
                          {member.type === 'agent'
                            ? member.name.slice(0, 2).toUpperCase()
                            : member.name.split(' ').map(n => n[0]).join('').slice(0, 2).toUpperCase()
                          }
                        </div>
                        <OnlineIndicator 
                          online={member.online!} 
                          type={member.type}
                          className="absolute -bottom-0.5 -right-0.5 w-3 h-3 border-2 border-card"
                        />
                      </div>

                      {/* Member info */}
                      <div className="flex-1 min-w-0">
                        <div className="flex items-center gap-2">
                          <span className="text-sm font-medium text-foreground truncate">
                            {member.name}
                          </span>
                          
                          {/* Role badge */}
                          {member.role === 'owner' && (
                            <span title="Owner">
                              <Crown className="h-3 w-3 text-amber-500" />
                            </span>
                          )}
                          {member.role === 'admin' && (
                            <span title="Admin">
                              <ShieldCheck className="h-3 w-3 text-blue-500" />
                            </span>
                          )}
                        </div>
                        
                        <div className="flex items-center gap-1.5">
                          <TypeBadge type={member.type} className="text-[9px]" />
                          {member.online && (
                            <span className="text-[10px] text-green-600 dark:text-green-400">
                              Online
                            </span>
                          )}
                        </div>
                      </div>
                    </div>
                  ))}
                </div>
              </>
            )}

            {/* Offline Members */}
            {members.some(member => !member.online) && (
              <>
                <div className="mb-2 px-1">
                  <h4 className="text-[11px] font-semibold uppercase tracking-wider text-muted-foreground">
                    Offline — {totalCount - onlineCount}
                  </h4>
                </div>
                <div className="space-y-1">
                  {members.filter(member => !member.online).map(member => (
                    <div
                      key={member.id}
                      className="flex items-center gap-3 px-2 py-2 rounded-md hover:bg-muted transition-colors opacity-60"
                    >
                      {/* Avatar */}
                      <div className={cn(
                        "w-7 h-7 rounded-md flex items-center justify-center text-xs font-bold",
                        member.type === 'agent'
                          ? "bg-primary/50 text-primary-foreground/70"
                          : "bg-accent/50 text-accent-foreground/70"
                      )}>
                        {member.type === 'agent'
                          ? member.name.slice(0, 2).toUpperCase()
                          : member.name.split(' ').map(n => n[0]).join('').slice(0, 2).toUpperCase()
                        }
                      </div>

                      {/* Member info */}
                      <div className="flex-1 min-w-0">
                        <div className="flex items-center gap-2">
                          <span className="text-sm font-medium text-muted-foreground truncate">
                            {member.name}
                          </span>
                          
                          {/* Role badge */}
                          {member.role === 'owner' && (
                            <span title="Owner">
                              <Crown className="h-3 w-3 text-amber-500/60" />
                            </span>
                          )}
                          {member.role === 'admin' && (
                            <span title="Admin">
                              <ShieldCheck className="h-3 w-3 text-blue-500/60" />
                            </span>
                          )}
                        </div>
                        
                        <div className="flex items-center gap-1.5">
                          <TypeBadge type={member.type} className="text-[9px] opacity-75" />
                          <span className="text-[10px] text-muted-foreground">
                            {member.lastSeenAt && formatLastSeen(member.lastSeenAt)}
                          </span>
                        </div>
                      </div>
                    </div>
                  ))}
                </div>
              </>
            )}
          </div>
        )}
      </div>
    </aside>
  );
}