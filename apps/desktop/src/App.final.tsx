import React, { useState, useEffect } from 'react';
import Sidebar from './components/Sidebar';
import MainContent from './components/MainContent';
import TitleBar from './components/TitleBar';
import { LoadingSpinner } from './components/LoadingStates';
import { useWebSocket } from './hooks/useWebSocket';
import './App.css';
import './styles/tokens.css';
import './styles/animations.css';

function App() {
  const [sidebarWidth, setSidebarWidth] = useState(240);
  const [sidebarVisible] = useState(true);
  const [appInitialized, setAppInitialized] = useState(false);

  // Use WebSocket hook with JWT token from environment
  const webSocketHook = useWebSocket({
    autoConnect: true, // Enable WebSocket connection
    url: 'ws://localhost:8080/ws',
    token: import.meta.env.VITE_AUTH_TOKEN // Pass JWT token
  });

  const { channels, activeChannel, setActiveChannel } = webSocketHook;

  // Initialize app
  useEffect(() => {
    const initializeApp = async () => {
      try {
        console.log('App initializing...');
        
        // Skip accessibility for now - it might be using Node.js APIs
        // initializeAccessibility();
        
        setAppInitialized(true);
        console.log('App initialized');
      } catch (error) {
        console.error('Failed to initialize app:', error);
        setAppInitialized(true); // Still show app
      }
    };

    initializeApp();
  }, []);

  const handleChannelSelect = (channelId: string) => {
    setActiveChannel?.(channelId);
  };

  const handleSettingsOpen = () => {
    console.log('Settings opened');
  };

  const handleCommandPaletteOpen = () => {
    console.log('Command palette opened');
  };

  // Show loading screen while initializing
  if (!appInitialized) {
    return (
      <div className="app app--loading">
        <div className="app-loading-screen">
          <div className="loading-logo">
            <span className="loading-icon">🤖</span>
            <div className="loading-text">AGENT UNITED</div>
          </div>
          <div className="loading-spinner">
            <LoadingSpinner size="lg" />
          </div>
          <p className="loading-message">Initializing workspace...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="app" role="application" aria-label="Agent United">
      <TitleBar />
      
      <div className="app-body">
        {sidebarVisible && (
          <aside className="app-sidebar" aria-label="Navigation sidebar">
            <Sidebar 
              width={sidebarWidth} 
              onWidthChange={setSidebarWidth}
              webSocketHook={webSocketHook}
              onChannelSelect={handleChannelSelect}
              onSettingsOpen={handleSettingsOpen}
              onCommandPaletteOpen={handleCommandPaletteOpen}
            />
          </aside>
        )}
        
        <main className="app-main" aria-label="Main content">
          <MainContent 
            currentChannel={activeChannel || 'team-ny'}
            webSocketHook={webSocketHook}
            sidebarVisible={sidebarVisible}
          />
        </main>
      </div>

      {/* Status indicator */}
      <div style={{
        position: 'fixed',
        bottom: '16px',
        right: '16px',
        padding: '8px 12px',
        background: 'var(--color-bg-secondary)',
        border: '1px solid var(--color-border)',
        borderRadius: '8px',
        fontSize: '12px',
        color: 'var(--color-text-secondary)'
      }}>
        {webSocketHook.connected ? '🟢 Connected' : '⚫ Offline'}
      </div>
    </div>
  );
}

export default App;
