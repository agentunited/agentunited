import React, { useState, useEffect, useRef, useMemo, useCallback } from 'react';
import { Avatar, Input } from './ui';
import type { ChannelData, UserData } from '../services/WebSocketService';
import './CommandPalette.css';

export interface CommandPaletteItem {
  id: string;
  type: 'channel' | 'user' | 'action';
  title: string;
  subtitle?: string;
  icon?: React.ReactNode;
  avatar?: {
    name: string;
    type: 'agent' | 'human';
    status?: 'online' | 'offline' | 'away';
  };
  action: () => void;
}

interface CommandPaletteProps {
  isOpen: boolean;
  onClose: () => void;
  channels: ChannelData[];
  users: UserData[];
  onChannelSelect: (channelId: string) => void;
  onUserSelect: (userId: string) => void;
  onSettingsOpen: () => void;
  onNewChannel: () => void;
  onNewDM: () => void;
}

export function CommandPalette({
  isOpen,
  onClose,
  channels,
  users,
  onChannelSelect,
  onUserSelect,
  onSettingsOpen,
  onNewChannel,
  onNewDM
}: CommandPaletteProps) {
  const [query, setQuery] = useState('');
  const [selectedIndex, setSelectedIndex] = useState(0);
  const inputRef = useRef<HTMLInputElement>(null);
  const listRef = useRef<HTMLDivElement>(null);

  // Generate command palette items
  const items = useMemo((): CommandPaletteItem[] => {
    const allItems: CommandPaletteItem[] = [];

    // Add actions first (always visible)
    allItems.push(
      {
        id: 'action-new-channel',
        type: 'action',
        title: 'Create new channel',
        subtitle: 'Start a new channel for team collaboration',
        icon: <span className="command-icon">📢</span>,
        action: () => {
          onNewChannel();
          onClose();
        }
      },
      {
        id: 'action-new-dm',
        type: 'action',
        title: 'Start direct message',
        subtitle: 'Send a private message to someone',
        icon: <span className="command-icon">💬</span>,
        action: () => {
          onNewDM();
          onClose();
        }
      },
      {
        id: 'action-settings',
        type: 'action',
        title: 'Open settings',
        subtitle: 'Configure your app preferences',
        icon: <span className="command-icon">⚙️</span>,
        action: () => {
          onSettingsOpen();
          onClose();
        }
      }
    );

    // Filter and add channels
    const filteredChannels = channels.filter(channel =>
      query === '' || 
      channel.name.toLowerCase().includes(query.toLowerCase()) ||
      channel.topic?.toLowerCase().includes(query.toLowerCase())
    );

    filteredChannels.forEach(channel => {
      allItems.push({
        id: `channel-${channel.id}`,
        type: 'channel',
        title: `#${channel.name}`,
        subtitle: channel.topic || 'No topic set',
        icon: <span className="command-icon">#</span>,
        action: () => {
          onChannelSelect(channel.id);
          onClose();
        }
      });
    });

    // Filter and add users
    const filteredUsers = users.filter(user =>
      query === '' || 
      user.name.toLowerCase().includes(query.toLowerCase())
    );

    filteredUsers.forEach(user => {
      allItems.push({
        id: `user-${user.id}`,
        type: 'user',
        title: user.name,
        subtitle: user.type === 'agent' ? '🤖 Agent' : '👤 Human',
        avatar: {
          name: user.name,
          type: user.type,
          status: user.status
        },
        action: () => {
          onUserSelect(user.id);
          onClose();
        }
      });
    });

    // Apply query filter for non-action items
    if (query) {
      return allItems.filter(item => {
        if (item.type === 'action') {
          return item.title.toLowerCase().includes(query.toLowerCase()) ||
                 item.subtitle?.toLowerCase().includes(query.toLowerCase());
        }
        return item.title.toLowerCase().includes(query.toLowerCase()) ||
               item.subtitle?.toLowerCase().includes(query.toLowerCase());
      });
    }

    return allItems.slice(0, 50); // Limit results
  }, [query, channels, users, onChannelSelect, onUserSelect, onSettingsOpen, onNewChannel, onNewDM, onClose]);

  // Reset selection when items change
  useEffect(() => {
    setSelectedIndex(0);
  }, [items]);

  // Focus input when opened
  useEffect(() => {
    if (isOpen) {
      setQuery('');
      setSelectedIndex(0);
      setTimeout(() => {
        inputRef.current?.focus();
      }, 50);
    }
  }, [isOpen]);

  // Handle keyboard navigation
  useEffect(() => {
    if (!isOpen) return;

    const handleKeyDown = (e: KeyboardEvent) => {
      switch (e.key) {
        case 'ArrowDown':
          e.preventDefault();
          setSelectedIndex(prev => 
            prev < items.length - 1 ? prev + 1 : 0
          );
          break;
          
        case 'ArrowUp':
          e.preventDefault();
          setSelectedIndex(prev => 
            prev > 0 ? prev - 1 : items.length - 1
          );
          break;
          
        case 'Enter':
          e.preventDefault();
          if (items[selectedIndex]) {
            items[selectedIndex].action();
          }
          break;
          
        case 'Escape':
          e.preventDefault();
          onClose();
          break;
      }
    };

    document.addEventListener('keydown', handleKeyDown);
    return () => document.removeEventListener('keydown', handleKeyDown);
  }, [isOpen, items, selectedIndex, onClose]);

  // Auto-scroll selected item into view
  useEffect(() => {
    if (!listRef.current) return;

    const selectedElement = listRef.current.children[selectedIndex] as HTMLElement;
    if (selectedElement) {
      selectedElement.scrollIntoView({ block: 'nearest' });
    }
  }, [selectedIndex]);

  // Handle click outside to close
  useEffect(() => {
    if (!isOpen) return;

    const handleClickOutside = (e: MouseEvent) => {
      const target = e.target as HTMLElement;
      if (!target.closest('.command-palette')) {
        onClose();
      }
    };

    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, [isOpen, onClose]);

  // Keyboard shortcut registration
  useEffect(() => {
    const handleGlobalKeyDown = (e: KeyboardEvent) => {
      if ((e.metaKey || e.ctrlKey) && e.key === 'k' && !isOpen) {
        e.preventDefault();
        // This would be handled by parent component
      }
    };

    document.addEventListener('keydown', handleGlobalKeyDown);
    return () => document.removeEventListener('keydown', handleGlobalKeyDown);
  }, [isOpen]);

  if (!isOpen) {
    return null;
  }

  return (
    <div className="command-palette-overlay">
      <div className="command-palette">
        <div className="command-palette__header">
          <Input
            ref={inputRef}
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder="Search channels, people, or commands..."
            className="command-palette__input"
            variant="search"
          />
        </div>

        <div className="command-palette__content">
          {items.length > 0 ? (
            <div ref={listRef} className="command-palette__list">
              {items.map((item, index) => (
                <button
                  key={item.id}
                  className={`command-palette__item ${
                    index === selectedIndex ? 'command-palette__item--selected' : ''
                  }`}
                  onClick={item.action}
                  onMouseEnter={() => setSelectedIndex(index)}
                >
                  <div className="command-palette__icon">
                    {item.avatar ? (
                      <Avatar 
                        name={item.avatar.name} 
                        type={item.avatar.type}
                        status={item.avatar.status}
                        size="sm"
                      />
                    ) : (
                      item.icon
                    )}
                  </div>
                  
                  <div className="command-palette__info">
                    <div className="command-palette__title">
                      {item.title}
                    </div>
                    {item.subtitle && (
                      <div className="command-palette__subtitle">
                        {item.subtitle}
                      </div>
                    )}
                  </div>

                  <div className="command-palette__type">
                    {item.type === 'channel' && 'Channel'}
                    {item.type === 'user' && 'Person'}
                    {item.type === 'action' && 'Action'}
                  </div>
                </button>
              ))}
            </div>
          ) : (
            <div className="command-palette__empty">
              <div className="command-palette__empty-icon">🔍</div>
              <div className="command-palette__empty-title">No results found</div>
              <div className="command-palette__empty-subtitle">
                Try a different search term
              </div>
            </div>
          )}
        </div>

        <div className="command-palette__footer">
          <div className="command-palette__hints">
            <kbd>↵</kbd> to select
            <kbd>↑↓</kbd> to navigate
            <kbd>Esc</kbd> to close
          </div>
        </div>
      </div>
    </div>
  );
}

// Hook for managing command palette state
export function useCommandPalette() {
  const [isOpen, setIsOpen] = useState(false);

  const openPalette = useCallback(() => {
    setIsOpen(true);
  }, []);

  const closePalette = useCallback(() => {
    setIsOpen(false);
  }, []);

  // Global keyboard shortcut
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
        e.preventDefault();
        openPalette();
      }
    };

    document.addEventListener('keydown', handleKeyDown);
    return () => document.removeEventListener('keydown', handleKeyDown);
  }, [openPalette]);

  return {
    isOpen,
    openPalette,
    closePalette
  };
}