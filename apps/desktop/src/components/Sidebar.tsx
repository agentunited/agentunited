import React, { useState, useCallback } from 'react';
import { ChannelListItem, Input, Button } from './ui';
import type { ChannelData } from './ui';
import '../styles/sidebar.css';

interface SidebarProps {
  width: number;
  onWidthChange: (width: number) => void;
}

function Sidebar({ width, onWidthChange }: SidebarProps) {
  const [isResizing, setIsResizing] = useState(false);
  const [activeChannelId, setActiveChannelId] = useState('general');

  // Sample data for Phase 2 demonstration
  const channels: ChannelData[] = [
    {
      id: 'general',
      name: 'general',
      type: 'channel',
      unreadCount: 3,
      lastMessage: 'Data analysis complete',
      lastActivity: '2m ago'
    },
    {
      id: 'crypto',
      name: 'crypto',
      type: 'channel',
      unreadCount: 1,
      lastMessage: 'BTC price update',
      lastActivity: '5m ago'
    },
    {
      id: 'research',
      name: 'research',
      type: 'channel',
      lastMessage: 'Meeting scheduled',
      lastActivity: '1h ago'
    }
  ];

  const directMessages: ChannelData[] = [
    {
      id: 'dm-agent1',
      name: 'Agent Alpha',
      type: 'dm',
      memberName: 'Agent Alpha',
      memberType: 'agent',
      memberOnline: true,
      unreadCount: 2,
      lastMessage: 'Analysis ready for review',
      lastActivity: '1m ago'
    },
    {
      id: 'dm-human1',
      name: 'Dr. Smith',
      type: 'dm',
      memberName: 'Dr. Smith',
      memberType: 'human',
      memberOnline: true,
      lastMessage: 'Thanks for the update',
      lastActivity: '10m ago'
    },
    {
      id: 'dm-agent2',
      name: 'AnalystBot',
      type: 'dm',
      memberName: 'AnalystBot',
      memberType: 'agent',
      memberOnline: false,
      lastMessage: 'Processing complete',
      lastActivity: '2h ago'
    }
  ];

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
      </div>

      {/* Search */}
      <div className="sidebar-search">
        <Input 
          variant="search"
          placeholder="🔍 Search..."
          leftIcon={<span>🔍</span>}
          size="sm"
        />
      </div>

      {/* Channels section */}
      <div className="sidebar-section">
        <div className="section-header">
          <span className="section-title">CHANNELS</span>
          <Button variant="icon" size="sm" ariaLabel="Add channel">
            +
          </Button>
        </div>
        <div className="channel-list">
          {channels.map((channel) => (
            <ChannelListItem
              key={channel.id}
              channel={channel}
              isActive={activeChannelId === channel.id}
              onClick={(channelId) => setActiveChannelId(channelId)}
            />
          ))}
        </div>
      </div>

      {/* Direct Messages section */}
      <div className="sidebar-section">
        <div className="section-header">
          <span className="section-title">DIRECT MESSAGES</span>
          <Button variant="icon" size="sm" ariaLabel="New message">
            +
          </Button>
        </div>
        <div className="dm-list">
          {directMessages.map((dm) => (
            <ChannelListItem
              key={dm.id}
              channel={dm}
              isActive={activeChannelId === dm.id}
              onClick={(channelId) => setActiveChannelId(channelId)}
            />
          ))}
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