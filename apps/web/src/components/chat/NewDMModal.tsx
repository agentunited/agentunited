import { useState, useEffect, useCallback } from 'react';
import { Search, User, Bot, X } from 'lucide-react';
import { cn } from '../../lib/utils';
import { chatApi } from '../../services/chatApi';

interface User {
  id: string;
  name: string;
  email: string;
  type: 'agent' | 'human';
  online?: boolean;
}

interface NewDMModalProps {
  isOpen: boolean;
  onClose: () => void;
  onDMCreated: (dmId: string) => void;
}

export function NewDMModal({ isOpen, onClose, onDMCreated }: NewDMModalProps) {
  const [searchQuery, setSearchQuery] = useState('');
  const [users, setUsers] = useState<User[]>([]);
  const [filteredUsers, setFilteredUsers] = useState<User[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [isCreating, setIsCreating] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Load users when modal opens
  useEffect(() => {
    if (!isOpen) return;

    const loadUsers = async () => {
      try {
        setIsLoading(true);
        setError(null);
        
        const apiUsers = await chatApi.getUsers();

        const userList: User[] = apiUsers.map((user) => {
          const type: 'agent' | 'human' = user.type === 'agent' ? 'agent' : 'human';
          const baseName = user.display_name || (user.email ? user.email.split('@')[0] : 'Unknown');
          const normalizedName = type === 'agent' ? baseName.replace(/\s+agent$/i, '') : baseName;

          return {
            id: user.id,
            name: normalizedName,
            email: user.email || '',
            type,
            online: false,
          };
        });
        
        // Remove current user from the list
        const currentUserId = localStorage.getItem('user-id');
        const filteredUserList = userList.filter(u => u.id !== currentUserId);
        
        setUsers(filteredUserList);
        setFilteredUsers(filteredUserList);
      } catch (error) {
        console.error('Failed to load users:', error);
        setError(error instanceof Error ? error.message : 'Failed to load users');
      } finally {
        setIsLoading(false);
      }
    };

    loadUsers();
  }, [isOpen]);

  // Filter users based on search query
  useEffect(() => {
    if (!searchQuery.trim()) {
      // Sort by online status first, then alphabetically
      const sorted = [...users].sort((a, b) => {
        if (a.online !== b.online) {
          return a.online ? -1 : 1;
        }
        return a.name.localeCompare(b.name);
      });
      setFilteredUsers(sorted);
    } else {
      const query = searchQuery.toLowerCase();
      const filtered = users.filter(user => 
        user.name.toLowerCase().includes(query) ||
        user.email.toLowerCase().includes(query)
      );
      setFilteredUsers(filtered);
    }
  }, [searchQuery, users]);

  const handleCreateDM = async (userId: string) => {
    try {
      setIsCreating(true);
      setError(null);
      
      const dm = await chatApi.createDM(userId);
      onDMCreated(dm.id);
      onClose();
    } catch (error) {
      console.error('Failed to create DM:', error);
      setError(error instanceof Error ? error.message : 'Failed to start conversation');
    } finally {
      setIsCreating(false);
    }
  };

  const handleKeyDown = useCallback((e: React.KeyboardEvent) => {
    if (e.key === 'Escape') {
      onClose();
    }
  }, [onClose]);

  const handleSearchKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && filteredUsers.length === 1) {
      handleCreateDM(filteredUsers[0].id);
    }
  };

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50" onClick={onClose}>
      <div 
        className="w-full max-w-md bg-card border border-border rounded-lg p-6 m-4 max-h-[80vh] flex flex-col"
        onClick={e => e.stopPropagation()}
        onKeyDown={handleKeyDown}
      >
        <div className="flex items-center justify-between mb-4">
          <h2 className="text-lg font-semibold text-foreground">New Message</h2>
          <button
            onClick={onClose}
            className="p-1 rounded-md text-muted-foreground hover:text-foreground hover:bg-muted transition-colors"
          >
            <X className="h-5 w-5" />
          </button>
        </div>

        {/* Search Input */}
        <div className="mb-4">
          <label className="block text-sm font-medium text-foreground mb-2">To:</label>
          <div className="flex items-center gap-2 rounded-md bg-background border border-input px-3 py-2">
            <Search className="h-4 w-4 text-muted-foreground" />
            <input
              type="text"
              placeholder="Search people..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              onKeyDown={handleSearchKeyDown}
              className="w-full bg-transparent text-sm outline-none placeholder:text-muted-foreground"
              autoFocus
            />
            {searchQuery && (
              <button 
                onClick={() => setSearchQuery('')}
                className="text-muted-foreground hover:text-foreground"
              >
                <X className="h-3 w-3" />
              </button>
            )}
          </div>
        </div>

        {error && (
          <div className="mb-4 p-3 bg-destructive/10 border border-destructive/20 rounded-md text-destructive text-sm">
            {error}
          </div>
        )}

        {/* User List */}
        <div className="flex-1 min-h-0">
          {isLoading ? (
            <div className="flex items-center justify-center py-8">
              <div className="text-sm text-muted-foreground">Loading people...</div>
            </div>
          ) : filteredUsers.length === 0 ? (
            <div className="text-center py-8">
              <div className="text-sm text-muted-foreground">
                {searchQuery ? 'No people found' : 'No people available'}
              </div>
            </div>
          ) : (
            <div className="space-y-1 max-h-64 overflow-y-auto">
              {filteredUsers.map((user) => (
                <button
                  key={user.id}
                  onClick={() => handleCreateDM(user.id)}
                  disabled={isCreating}
                  className="w-full flex items-center gap-3 p-3 rounded-md hover:bg-muted transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  {/* Online Indicator */}
                  <div className="relative">
                    <div className={cn(
                      "w-8 h-8 rounded-lg flex items-center justify-center text-xs font-bold",
                      user.type === 'agent' 
                        ? "bg-primary text-primary-foreground" 
                        : "bg-accent text-accent-foreground"
                    )}>
                      {user.type === 'agent' 
                        ? user.name.slice(0, 2).toUpperCase()
                        : user.name.split(' ').map(n => n[0]).join('').slice(0, 2).toUpperCase()
                      }
                    </div>
                    {user.online && (
                      <div className={cn(
                        "absolute -bottom-0.5 -right-0.5 w-3 h-3 rounded-full border-2 border-card",
                        user.type === 'agent' ? "bg-primary" : "bg-accent"
                      )} />
                    )}
                  </div>

                  {/* User Info */}
                  <div className="flex-1 text-left">
                    <div className="flex items-center gap-2">
                      <span className="text-sm font-medium text-foreground">{user.name}</span>
                      <div className={cn(
                        "inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-[10px] font-semibold uppercase tracking-wide",
                        user.type === 'agent' 
                          ? "bg-primary/10 text-primary" 
                          : "bg-accent/25 text-accent-foreground"
                      )}>
                        {user.type === 'agent' ? (
                          <Bot className="h-2.5 w-2.5" />
                        ) : (
                          <User className="h-2.5 w-2.5" />
                        )}
                        {user.type}
                      </div>
                    </div>
                    <div className="text-xs text-muted-foreground">
                      {user.online ? 'Online' : 'Offline'}
                    </div>
                  </div>
                </button>
              ))}
            </div>
          )}
        </div>

        {/* Footer */}
        <div className="flex justify-end gap-2 pt-4 border-t border-border">
          <button
            onClick={onClose}
            className="px-4 py-2 text-sm font-medium text-muted-foreground hover:text-foreground transition-colors"
          >
            Cancel
          </button>
        </div>
      </div>
    </div>
  );
}