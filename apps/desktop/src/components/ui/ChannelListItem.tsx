import React from 'react';
import { StatusDot, UnreadCount } from './Badge';
import './ChannelListItem.css';

export interface ChannelData {
  id: string;
  name: string;
  type: 'channel' | 'dm';
  unreadCount?: number;
  lastMessage?: string;
  lastActivity?: string;
  // For DMs
  memberName?: string;
  memberType?: 'agent' | 'human';
  memberOnline?: boolean;
}

export interface ChannelListItemProps {
  channel: ChannelData;
  isActive?: boolean;
  onClick?: (channelId: string) => void;
  onContextMenu?: (channelId: string, event: React.MouseEvent) => void;
  className?: string;
}

export function ChannelListItem({
  channel,
  isActive = false,
  onClick,
  onContextMenu,
  className = ''
}: ChannelListItemProps) {
  const itemClass = [
    'channel-list-item',
    `channel-list-item--${channel.type}`,
    isActive ? 'channel-list-item--active' : '',
    channel.unreadCount ? 'channel-list-item--unread' : '',
    className,
  ].filter(Boolean).join(' ');

  const handleClick = () => {
    if (onClick) {
      onClick(channel.id);
    }
  };

  const handleContextMenu = (event: React.MouseEvent) => {
    event.preventDefault();
    if (onContextMenu) {
      onContextMenu(channel.id, event);
    }
  };

  const renderChannelIcon = () => {
    if (channel.type === 'channel') {
      return <span className="channel-list-item__icon">#</span>;
    }
    
    return (
      <StatusDot 
        status={channel.memberOnline ? 'online' : 'offline'}
        type={channel.memberType || 'human'}
        size="sm"
      />
    );
  };

  const renderChannelName = () => {
    if (channel.type === 'channel') {
      return channel.name;
    }
    
    return channel.memberName || channel.name;
  };

  const renderTypeBadge = () => {
    if (channel.type === 'dm' && channel.memberType) {
      return (
        <span className={`channel-list-item__type-badge channel-list-item__type-badge--${channel.memberType}`}>
          {channel.memberType.toUpperCase()}
        </span>
      );
    }
    return null;
  };

  return (
    <div
      className={itemClass}
      onClick={handleClick}
      onContextMenu={handleContextMenu}
      role="button"
      tabIndex={0}
      aria-label={`${channel.type === 'channel' ? 'Channel' : 'Direct message'} ${renderChannelName()}${channel.unreadCount ? `, ${channel.unreadCount} unread messages` : ''}`}
      onKeyDown={(e) => {
        if (e.key === 'Enter' || e.key === ' ') {
          e.preventDefault();
          handleClick();
        }
      }}
    >
      <div className="channel-list-item__icon-container">
        {renderChannelIcon()}
      </div>
      
      <div className="channel-list-item__content">
        <div className="channel-list-item__header">
          <span className="channel-list-item__name">
            {renderChannelName()}
          </span>
          {renderTypeBadge()}
        </div>
        
        {channel.lastMessage && (
          <div className="channel-list-item__preview">
            {channel.lastMessage}
          </div>
        )}
      </div>
      
      <div className="channel-list-item__meta">
        {channel.lastActivity && (
          <span className="channel-list-item__time">
            {channel.lastActivity}
          </span>
        )}
        {channel.unreadCount && (
          <UnreadCount count={channel.unreadCount} />
        )}
      </div>
    </div>
  );
}