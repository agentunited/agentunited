import React, { useState, useCallback } from 'react';
import '../styles/sidebar.css';

interface SidebarProps {
  width: number;
  onWidthChange: (width: number) => void;
}

function Sidebar({ width, onWidthChange }: SidebarProps) {
  const [isResizing, setIsResizing] = useState(false);

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
        <input 
          type="text" 
          placeholder="🔍 Search..." 
          className="search-input"
        />
      </div>

      {/* Channels section */}
      <div className="sidebar-section">
        <div className="section-header">
          <span className="section-title">CHANNELS</span>
          <button className="add-button">+</button>
        </div>
        <div className="channel-list">
          <div className="channel-item active">
            <span className="channel-icon">#</span>
            <span className="channel-name">general</span>
            <span className="unread-count">3</span>
          </div>
          <div className="channel-item">
            <span className="channel-icon">#</span>
            <span className="channel-name">crypto</span>
            <span className="unread-count">1</span>
          </div>
          <div className="channel-item">
            <span className="channel-icon">#</span>
            <span className="channel-name">research</span>
          </div>
        </div>
      </div>

      {/* Direct Messages section */}
      <div className="sidebar-section">
        <div className="section-header">
          <span className="section-title">DIRECT MESSAGES</span>
          <button className="add-button">+</button>
        </div>
        <div className="dm-list">
          <div className="dm-item">
            <div className="online-indicator active"></div>
            <span className="dm-name">Agent Alpha</span>
            <span className="type-badge agent">AGENT</span>
          </div>
          <div className="dm-item">
            <div className="online-indicator active"></div>
            <span className="dm-name">Dr. Smith</span>
            <span className="type-badge human">HUMAN</span>
          </div>
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