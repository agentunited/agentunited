import React from 'react';
import './Badge.css';

export interface BadgeProps {
  children: React.ReactNode;
  variant?: 'agent' | 'human' | 'unread' | 'status' | 'success' | 'warning' | 'danger';
  size?: 'sm' | 'md';
  className?: string;
}

export function Badge({ 
  children, 
  variant = 'agent', 
  size = 'md',
  className = '' 
}: BadgeProps) {
  const badgeClass = [
    'badge',
    `badge--${variant}`,
    `badge--${size}`,
    className,
  ].filter(Boolean).join(' ');

  return (
    <span className={badgeClass}>
      {children}
    </span>
  );
}

export interface StatusDotProps {
  status: 'online' | 'offline' | 'away' | 'busy';
  type?: 'agent' | 'human';
  size?: 'sm' | 'md' | 'lg';
  className?: string;
}

export function StatusDot({ 
  status, 
  type = 'human',
  size = 'md', 
  className = '' 
}: StatusDotProps) {
  const dotClass = [
    'status-dot',
    `status-dot--${status}`,
    `status-dot--${type}`,
    `status-dot--${size}`,
    className,
  ].filter(Boolean).join(' ');

  return <span className={dotClass} aria-label={`${status} ${type}`} />;
}

export interface UnreadCountProps {
  count: number;
  max?: number;
  className?: string;
}

export function UnreadCount({ 
  count, 
  max = 99, 
  className = '' 
}: UnreadCountProps) {
  if (count === 0) return null;

  const displayCount = count > max ? `${max}+` : count.toString();
  
  return (
    <Badge variant="unread" size="sm" className={className}>
      {displayCount}
    </Badge>
  );
}