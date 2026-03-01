import React, { useState, useCallback } from 'react';
import { ChannelListItem, Input, Button } from './ui';
import type { ChannelData } from './ui';
import { NoChannelsState, NoDirectMessagesState } from './EmptyStates';
import { useWebSocket } from '../hooks/useWebSocket';
import '../styles/sidebar.css';

interface SidebarProps {
  width: number;
  onWidthChange: (width: number) => void;
  webSocketHook?: ReturnType<typeof useWebSocket>;
  onChannelSelect?: (channelId: string) => void;
  onSettingsOpen?: () => void;
  onCommandPaletteOpen?: () => void;
}

function Sidebar({ 
  width, 
  onWidthChange, 
  webSocketHook,
  onChannelSelect,
  onSettingsOpen,
  onCommandPaletteOpen
}: SidebarProps) {
  const [isResizing, setIsResizing] = useState(false);
  const [searchQuery, setSearchQuery] = useState('');

  // Use WebSocket hook for real data
  const wsHook = webSocketHook || useWebSocket();
  const { channels: wsChannels, users, activeChannel, connected } = wsHook;

  // Helper function to format last activity time
  function formatLastActivity(timestamp: string): string {
    const now = new Date();
    const time = new Date(timestamp);
    const diffMs = now.getTime() - time.getTime();
    const diffMins = Math.floor(diffMs / 60000);
    
    if (diffMins < 1) return 'Just now';
    if (diffMins < 60) return `${diffMins}m ago`;
    
    const diffHours = Math.floor(diffMins / 60);
    if (diffHours < 24) return `${diffHours}h ago`;
    
    const diffDays = Math.floor(diffHours / 24);
    return `${diffDays}d ago`;
  }

  // Convert WebSocket data to component format
  const channels: ChannelData[] = wsChannels
    .filter(ch => ch.type === 'channel')
    .map(ch => ({
      id: ch.id,
      name: ch.name,
      type: 'channel',
      unreadCount: ch.unreadCount || 0,
      lastMessage: ch.lastMessage?.content || 'No messages yet',
      lastActivity: ch.lastMessage?.timestamp ? formatLastActivity(ch.lastMessage.timestamp) : 'Never'
    }));

  const directMessages: ChannelData[] = wsChannels
    .filter(ch => ch.type === 'dm')
    .map(ch => {
      // Find the other user in DM
      const otherMember = ch.members?.find(m => m.id !== wsHook.currentUser?.id);
      return {
        id: ch.id,
        name: otherMember?.name || 'Unknown User',
        type: 'dm',
        memberName: otherMember?.name || 'Unknown User',
        memberType: otherMember?.type || 'human',
        memberOnline: otherMember?.status === 'online',
        unreadCount: ch.unreadCount || 0,
        lastMessage: ch.lastMessage?.content || 'No messages yet',
        lastActivity: ch.lastMessage?.timestamp ? formatLastActivity(ch.lastMessage.timestamp) : 'Never'
      };
    });

  // Filter by search query
  const filteredChannels = searchQuery 
    ? channels.filter(ch => ch.name.toLowerCase().includes(searchQuery.toLowerCase()))
    : channels;
  
  const filteredDMs = searchQuery
    ? directMessages.filter(dm => dm.name.toLowerCase().includes(searchQuery.toLowerCase()))
    : directMessages;

  const handleMouseDown = useCallback((e: React.MouseEvent) => {
    e.preventDefault();
    setIsResizing(true);
    
    const startX = e.clientX;
    const startWidth = width;
    
    const handleMouseMove = (e: MouseEvent) => {
      const newWidth = Math.min(
        Math.max(startWidth + (e.clientX - startX), 200), // min 200px
        320 // max 320px
      );
      onWidthChange(newWidth);
    };
    
    const handleMouseUp = () => {
      setIsResizing(false);
      document.removeEventListener('mousemove', handleMouseMove);
      document.removeEventListener('mouseup', handleMouseUp);
    };
    
    document.addEventListener('mousemove', handleMouseMove);
    document.addEventListener('mouseup', handleMouseUp);
  }, [width, onWidthChange]);

  const handleChannelClick = (channelId: string) => {
    onChannelSelect?.(channelId);
  };

  const handleNewChannel = () => {
    console.log('Create new channel');
  };

  const handleNewDM = () => {
    console.log('Start new DM');
  };

  const handleSearchChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setSearchQuery(e.target.value);
  };

  const handleSearchKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' || (e.metaKey && e.key === 'k')) {
      e.preventDefault();
      onCommandPaletteOpen?.();
    }
  };

  return (
    <div className="sidebar" style={{ width: `${width}px` }}>
      {/* Brand header */}
      <div className="sidebar-header">
        <div className="brand-icon">
          <span>🤖</span>
        </div>
        <div className="brand-text">
          <div className="brand-name">AGENTS</div>
          <div className="brand-tagline">UNITED</div>
        </div>
        <div className="header-actions">
          <Button 
            variant="icon" 
            size="sm" 
            ariaLabel="Settings"
            onClick={onSettingsOpen}
          >
            ⚙️
          </Button>
        </div>
      </div>

      {/* Search */}
      <div className="sidebar-search">
        <Input 
          variant="search"
          placeholder="🔍 Search (Cmd+K)"
          value={searchQuery}
          onChange={handleSearchChange}
          onKeyDown={handleSearchKeyDown}
          size="sm"
        />
      </div>

      {/* Connection status */}
      {!connected && (
        <div className="connection-status offline">
          <span className="status-dot"></span>
          Connecting...
        </div>
      )}

      {/* Channels section */}
      <div className="sidebar-section">
        <div className="section-header">
          <span className="section-title">CHANNELS</span>
          <Button 
            variant="icon" 
            size="sm" 
            ariaLabel="Add channel"
            onClick={handleNewChannel}
          >
            +
          </Button>
        </div>
        <div className="channel-list">
          {filteredChannels.length > 0 ? (
            filteredChannels.map((channel) => (
              <ChannelListItem
                key={channel.id}
                channel={channel}
                isActive={activeChannel === channel.id}
                onClick={handleChannelClick}
              />
            ))
          ) : searchQuery ? (
            <div className="no-results">No channels found</div>
          ) : (
            <NoChannelsState onCreateChannel={handleNewChannel} />
          )}
        </div>
      </div>

      {/* Direct Messages section */}
      <div className="sidebar-section">
        <div className="section-header">
          <span className="section-title">DIRECT MESSAGES</span>
          <Button 
            variant="icon" 
            size="sm" 
            ariaLabel="New message"
            onClick={handleNewDM}
          >
            +
          </Button>
        </div>
        <div className="dm-list">
          {filteredDMs.length > 0 ? (
            filteredDMs.map((dm) => (
              <ChannelListItem
                key={dm.id}
                channel={dm}
                isActive={activeChannel === dm.id}
                onClick={handleChannelClick}
              />
            ))
          ) : searchQuery ? (
            <div className="no-results">No conversations found</div>
          ) : (
            <NoDirectMessagesState onNewMessage={handleNewDM} />
          )}
        </div>
      </div>

      {/* Resize handle */}
      <div 
        className="sidebar-resize-handle"
        onMouseDown={handleMouseDown}
        style={{ cursor: isResizing ? 'col-resize' : 'default' }}
      />
    </div>
  );
}

export default Sidebar;