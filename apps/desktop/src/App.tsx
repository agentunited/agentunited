import React, { useState, useEffect } from 'react';
import Sidebar from './components/Sidebar';
import MainContent from './components/MainContent';
import TitleBar from './components/TitleBar';
import { InviteWindow, SettingsWindow } from './screens';
import { CommandPalette, useCommandPalette } from './components/CommandPalette';
import { useWebSocket } from './hooks/useWebSocket';
import './App.css';

function App() {
  const [sidebarWidth, setSidebarWidth] = useState(240);
  const [showInviteWindow, setShowInviteWindow] = useState(false);
  const [showSettingsWindow, setShowSettingsWindow] = useState(false);
  
  // WebSocket hook for global state management
  const webSocketHook = useWebSocket({
    autoConnect: true,
    url: process.env.WEBSOCKET_URL || 'ws://localhost:8080/ws'
  });

  const {
    channels,
    users,
    activeChannel,
    setActiveChannel
  } = webSocketHook;

  // Command palette
  const {
    isOpen: commandPaletteOpen,
    openPalette: openCommandPalette,
    closePalette: closeCommandPalette
  } = useCommandPalette();

  // Handle channel selection
  const handleChannelSelect = (channelId: string) => {
    setActiveChannel(channelId);
  };

  // Handle user selection (for DMs)
  const handleUserSelect = (userId: string) => {
    // Create or find DM channel with user
    const dmChannelId = `dm_${userId}`;
    setActiveChannel(dmChannelId);
    console.log('Opening DM with user:', userId);
  };

  // Handle settings
  const handleSettingsOpen = () => {
    setShowSettingsWindow(true);
  };

  // Handle new channel creation
  const handleNewChannel = () => {
    console.log('Create new channel');
  };

  // Handle new DM creation
  const handleNewDM = () => {
    console.log('Start new DM');
  };

  // Global keyboard shortcuts
  useEffect(() => {
    const handleGlobalKeyDown = (e: KeyboardEvent) => {
      // Cmd/Ctrl + K for command palette
      if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
        e.preventDefault();
        openCommandPalette();
      }
      
      // Cmd/Ctrl + , for settings
      if ((e.metaKey || e.ctrlKey) && e.key === ',') {
        e.preventDefault();
        handleSettingsOpen();
      }
      
      // Cmd/Ctrl + Shift + N for new channel
      if ((e.metaKey || e.ctrlKey) && e.shiftKey && e.key === 'N') {
        e.preventDefault();
        handleNewChannel();
      }
      
      // Cmd/Ctrl + Shift + D for new DM
      if ((e.metaKey || e.ctrlKey) && e.shiftKey && e.key === 'D') {
        e.preventDefault();
        handleNewDM();
      }
    };

    document.addEventListener('keydown', handleGlobalKeyDown);
    return () => document.removeEventListener('keydown', handleGlobalKeyDown);
  }, [openCommandPalette]);

  return (
    <div className="app">
      <TitleBar />
      
      <div className="app-body">
        <Sidebar 
          width={sidebarWidth} 
          onWidthChange={setSidebarWidth}
          webSocketHook={webSocketHook}
          onChannelSelect={handleChannelSelect}
          onSettingsOpen={handleSettingsOpen}
          onCommandPaletteOpen={openCommandPalette}
        />
        
        <MainContent 
          currentChannel={activeChannel || 'general'}
          webSocketHook={webSocketHook}
        />
      </div>

      {/* Command Palette */}
      <CommandPalette
        isOpen={commandPaletteOpen}
        onClose={closeCommandPalette}
        channels={channels}
        users={users}
        onChannelSelect={handleChannelSelect}
        onUserSelect={handleUserSelect}
        onSettingsOpen={handleSettingsOpen}
        onNewChannel={handleNewChannel}
        onNewDM={handleNewDM}
      />

      {/* Modal Windows */}
      {showInviteWindow && (
        <InviteWindow
          token="example_token"
          onSuccess={() => setShowInviteWindow(false)}
          onError={(error) => {
            console.error('Invite error:', error);
            setShowInviteWindow(false);
          }}
        />
      )}

      {showSettingsWindow && (
        <SettingsWindow
          onClose={() => setShowSettingsWindow(false)}
        />
      )}
    </div>
  );
}

export default App;