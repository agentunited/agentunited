import React, { useState, useEffect } from 'react';
import Sidebar from './components/Sidebar';
import MainContent from './components/MainContent';
import TitleBar from './components/TitleBar';
import { InviteWindow, SettingsWindow } from './screens';
import { CommandPalette, useCommandPalette } from './components/CommandPalette';
import { useWebSocket } from './hooks/useWebSocket';
import { useNativeIntegration, useMenuAction } from './hooks/useNativeIntegration';
import './App.css';

function App() {
  const [sidebarWidth, setSidebarWidth] = useState(240);
  const [showInviteWindow, setShowInviteWindow] = useState(false);
  const [showSettingsWindow, setShowSettingsWindow] = useState(false);
  const [sidebarVisible, setSidebarVisible] = useState(true);
  
  // WebSocket hook for global state management
  const webSocketHook = useWebSocket({
    autoConnect: true,
    url: process.env.WEBSOCKET_URL || 'ws://localhost:8080/ws'
  });

  const {
    channels,
    users,
    activeChannel,
    setActiveChannel,
    messages,
    connected
  } = webSocketHook;

  // Native integration hook
  const {
    isElectron,
    isMacOS,
    showNotification,
    updateDockBadge,
    clearDockBadge,
    updateAppConfig,
    appConfig
  } = useNativeIntegration();

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

  // Handle sidebar toggle
  const handleToggleSidebar = () => {
    setSidebarVisible(!sidebarVisible);
  };

  // Handle channel jumping
  const handleJumpToChannel = (index: number) => {
    if (index < channels.length) {
      setActiveChannel(channels[index].id);
    }
  };

  // Handle channel navigation
  const handlePreviousChannel = () => {
    const currentIndex = channels.findIndex(ch => ch.id === activeChannel);
    if (currentIndex > 0) {
      setActiveChannel(channels[currentIndex - 1].id);
    } else if (channels.length > 0) {
      setActiveChannel(channels[channels.length - 1].id); // Wrap to last
    }
  };

  const handleNextChannel = () => {
    const currentIndex = channels.findIndex(ch => ch.id === activeChannel);
    if (currentIndex < channels.length - 1) {
      setActiveChannel(channels[currentIndex + 1].id);
    } else if (channels.length > 0) {
      setActiveChannel(channels[0].id); // Wrap to first
    }
  };

  // Menu action handlers
  useMenuAction('preferences', handleSettingsOpen);
  useMenuAction('new-channel', handleNewChannel);
  useMenuAction('new-dm', handleNewDM);
  useMenuAction('command-palette', openCommandPalette);
  useMenuAction('toggle-sidebar', handleToggleSidebar);
  useMenuAction('jump-to-channel', handleJumpToChannel);
  useMenuAction('previous-channel', handlePreviousChannel);
  useMenuAction('next-channel', handleNextChannel);
  useMenuAction('keyboard-shortcuts', () => {
    // Show keyboard shortcuts modal/help
    console.log('Show keyboard shortcuts');
  });

  // Update dock badge with unread count
  useEffect(() => {
    if (!isElectron) return;

    const totalUnread = channels.reduce((sum, channel) => sum + (channel.unreadCount || 0), 0);
    updateDockBadge(totalUnread);
  }, [channels, isElectron, updateDockBadge]);

  // Clear dock badge when app becomes active
  useEffect(() => {
    if (!isElectron) return;

    const handleVisibilityChange = () => {
      if (!document.hidden && activeChannel) {
        // Clear dock badge when app is visible and focused
        setTimeout(() => clearDockBadge(), 1000);
      }
    };

    const handleFocus = () => {
      if (activeChannel) {
        setTimeout(() => clearDockBadge(), 1000);
      }
    };

    document.addEventListener('visibilitychange', handleVisibilityChange);
    window.addEventListener('focus', handleFocus);

    return () => {
      document.removeEventListener('visibilitychange', handleVisibilityChange);
      window.removeEventListener('focus', handleFocus);
    };
  }, [isElectron, activeChannel, clearDockBadge]);

  // Show native notifications for new messages
  useEffect(() => {
    if (!isElectron || !connected || !appConfig?.notifications.enabled) return;

    // Listen for new messages and show notifications
    const allMessages = Object.values(messages).flat();
    const recentMessages = allMessages.filter(msg => {
      const msgTime = new Date(msg.timestamp).getTime();
      const now = Date.now();
      return now - msgTime < 5000; // Messages from last 5 seconds
    });

    recentMessages.forEach(msg => {
      // Check if message contains mentions
      const isMention = msg.mentions?.some(mention => 
        mention.name.toLowerCase().includes('current user') // Replace with actual current user check
      ) || false;

      // Check if it's a direct message
      const isDM = msg.channelId.startsWith('dm_');

      // Show notification if conditions are met
      if (isMention || (isDM && appConfig.notifications.directMessages)) {
        const channelName = channels.find(ch => ch.id === msg.channelId)?.name || 'Unknown Channel';
        
        showNotification({
          title: isDM ? `${msg.author.name}` : `#${channelName}`,
          body: `${msg.author.name}: ${msg.content.slice(0, 100)}${msg.content.length > 100 ? '...' : ''}`,
          channelId: msg.channelId,
          messageId: msg.id,
          isMention: isMention
        });
      }
    });
  }, [messages, isElectron, connected, appConfig, channels, showNotification]);

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

      // Cmd/Ctrl + Shift + S for sidebar toggle
      if ((e.metaKey || e.ctrlKey) && e.shiftKey && e.key === 'S') {
        e.preventDefault();
        handleToggleSidebar();
      }

      // Cmd/Ctrl + [ for previous channel
      if ((e.metaKey || e.ctrlKey) && e.key === '[') {
        e.preventDefault();
        handlePreviousChannel();
      }

      // Cmd/Ctrl + ] for next channel
      if ((e.metaKey || e.ctrlKey) && e.key === ']') {
        e.preventDefault();
        handleNextChannel();
      }

      // Cmd/Ctrl + 1-9 for channel jumping
      if ((e.metaKey || e.ctrlKey) && e.key >= '1' && e.key <= '9') {
        e.preventDefault();
        const channelIndex = parseInt(e.key) - 1;
        handleJumpToChannel(channelIndex);
      }

      // Cmd/Ctrl + / for keyboard shortcuts help
      if ((e.metaKey || e.ctrlKey) && e.key === '/') {
        e.preventDefault();
        console.log('Show keyboard shortcuts');
      }
    };

    document.addEventListener('keydown', handleGlobalKeyDown);
    return () => document.removeEventListener('keydown', handleGlobalKeyDown);
  }, [openCommandPalette, channels]);

  return (
    <div className="app">
      <TitleBar />
      
      <div className="app-body">
        {sidebarVisible && (
          <Sidebar 
            width={sidebarWidth} 
            onWidthChange={setSidebarWidth}
            webSocketHook={webSocketHook}
            onChannelSelect={handleChannelSelect}
            onSettingsOpen={handleSettingsOpen}
            onCommandPaletteOpen={openCommandPalette}
          />
        )}
        
        <MainContent 
          currentChannel={activeChannel || 'general'}
          webSocketHook={webSocketHook}
          sidebarVisible={sidebarVisible}
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