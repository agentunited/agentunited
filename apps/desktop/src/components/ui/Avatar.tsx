import React from 'react';
import './Avatar.css';

export interface AvatarProps {
  name: string;
  type: 'agent' | 'human';
  size?: 'sm' | 'md' | 'lg';
  online?: boolean;
  src?: string;
  className?: string;
}

export function Avatar({ 
  name, 
  type, 
  size = 'md', 
  online = false,
  src,
  className = '' 
}: AvatarProps) {
  const getInitials = (name: string, type: 'agent' | 'human'): string => {
    if (type === 'agent') {
      // For agents, use first 2 characters of name (usually uppercase)
      return name.slice(0, 2).toUpperCase();
    } else {
      // For humans, use first letter of each word
      return name
        .split(' ')
        .map(word => word[0])
        .join('')
        .toUpperCase()
        .slice(0, 2);
    }
  };

  const initials = getInitials(name, type);

  const avatarClass = [
    'avatar',
    `avatar--${type}`,
    `avatar--${size}`,
    online ? 'avatar--online' : '',
    className,
  ].filter(Boolean).join(' ');

  const statusDotClass = [
    'avatar__status-dot',
    `avatar__status-dot--${size}`,
    online ? 'avatar__status-dot--online' : 'avatar__status-dot--offline',
    type === 'agent' ? 'avatar__status-dot--agent' : 'avatar__status-dot--human',
  ].join(' ');

  return (
    <div className={avatarClass} title={`${name} (${type})`}>
      {src ? (
        <img src={src} alt={name} className="avatar__image" />
      ) : (
        <span className="avatar__initials">
          {type === 'agent' ? '🤖' : initials}
        </span>
      )}
      <span className={statusDotClass} aria-label={`${online ? 'online' : 'offline'}`} />
    </div>
  );
}