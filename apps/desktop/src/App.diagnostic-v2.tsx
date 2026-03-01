import React, { useState, useEffect } from 'react';
import { useWebSocket } from './hooks/useWebSocket';
import './styles/tokens.css';

function App() {
  const [debugInfo, setDebugInfo] = useState<any>({});
  
  const webSocketHook = useWebSocket({
    autoConnect: true,
    url: 'ws://localhost:8080/ws',
    token: import.meta.env.VITE_AUTH_TOKEN
  });

  useEffect(() => {
    setDebugInfo({
      connected: webSocketHook.connected,
      channelsCount: webSocketHook.channels.length,
      channels: webSocketHook.channels.map(ch => ({ id: ch.id, name: ch.name, type: ch.type })),
      usersCount: webSocketHook.users.length,
      currentUser: webSocketHook.currentUser,
      activeChannel: webSocketHook.activeChannel,
      token: import.meta.env.VITE_AUTH_TOKEN ? 'present' : 'missing',
      tokenLength: import.meta.env.VITE_AUTH_TOKEN?.length || 0
    });
  }, [webSocketHook.connected, webSocketHook.channels, webSocketHook.users, webSocketHook.currentUser]);

  const testFetchChannels = async () => {
    try {
      const res = await fetch('http://localhost:8080/api/v1/channels', {
        headers: {
          'Authorization': `Bearer ${import.meta.env.VITE_AUTH_TOKEN}`
        }
      });
      const data = await res.json();
      setDebugInfo((prev: any) => ({ ...prev, fetchTest: data }));
    } catch (err: any) {
      setDebugInfo((prev: any) => ({ ...prev, fetchError: err.message }));
    }
  };

  return (
    <div style={{
      padding: '40px',
      background: 'var(--color-bg-primary)',
      color: 'var(--color-text-primary)',
      fontFamily: 'monospace',
      minHeight: '100vh'
    }}>
      <h1 style={{ color: 'var(--color-agent)' }}>🔍 Diagnostic Mode</h1>
      
      <div style={{ marginTop: '20px' }}>
        <h2>WebSocket Hook State:</h2>
        <pre style={{ 
          background: 'var(--color-bg-secondary)', 
          padding: '20px', 
          borderRadius: '8px',
          overflow: 'auto'
        }}>
          {JSON.stringify(debugInfo, null, 2)}
        </pre>
      </div>

      <button 
        onClick={testFetchChannels}
        style={{
          marginTop: '20px',
          padding: '10px 20px',
          background: 'var(--color-rust)',
          color: 'white',
          border: 'none',
          borderRadius: '4px',
          cursor: 'pointer'
        }}
      >
        Test Fetch Channels
      </button>

      <div style={{ marginTop: '20px', color: 'var(--color-text-secondary)', fontSize: '14px' }}>
        <p>This diagnostic page shows the actual state of the WebSocket connection and data.</p>
        <p>If channels array is empty but WebSocket is connected, the fetch on connect isn't working.</p>
      </div>
    </div>
  );
}

export default App;
