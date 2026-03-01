import { useState, useEffect, useRef, useCallback } from 'react';
import type { Message } from '../types/chat';
import { getApiBaseUrl } from '../services/apiConfig';
import { fetchMessages } from '../services/api';

interface UseWebSocketReturn {
  isConnected: boolean;
  messages: Message[];
  sendMessage: (text: string) => void;
  error: string | null;
}

export function useWebSocket(_url: string, channelId: string): UseWebSocketReturn {
  const [isConnected, setIsConnected] = useState(false);
  const [messages, setMessages] = useState<Message[]>([]);
  const [error, setError] = useState<string | null>(null);
  const wsRef = useRef<WebSocket | null>(null);
  const reconnectTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const reconnectAttemptsRef = useRef(0);

  // Load initial messages from API
  useEffect(() => {
    let cancelled = false;
    
    async function loadMessages() {
      try {
        const msgs = await fetchMessages(channelId);
        if (!cancelled) {
          setMessages(msgs);
        }
      } catch (err) {
        console.error('Failed to load messages:', err);
        if (!cancelled) {
          setMessages([]);
        }
      }
    }

    loadMessages();
    return () => { cancelled = true; };
  }, [channelId]);

  // WebSocket connection
  useEffect(() => {
    const token = localStorage.getItem('auth-token');
    if (!token) {
      setError('Not authenticated');
      return;
    }

    const baseUrl = getApiBaseUrl();
    const wsProtocol = baseUrl.startsWith('https') ? 'wss' : 'ws';
    const wsHost = baseUrl.replace(/^https?:\/\//, '');
    const wsUrl = `${wsProtocol}://${wsHost}/ws?token=${token}`;

    function connect() {
      const ws = new WebSocket(wsUrl);
      wsRef.current = ws;

      ws.onopen = () => {
        setIsConnected(true);
        setError(null);
        reconnectAttemptsRef.current = 0;
        console.log('WebSocket connected');
      };

      ws.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data);
          
          if (data.type === 'connected') {
            // Store user_id for own-message detection
            if (data.user_id) {
              localStorage.setItem('user-id', data.user_id);
            }
            return;
          }

          if (data.type === 'message' || data.type === 'new_message' || data.type === 'message.created') {
            const d = data.data || data; // message.created wraps in { type, data }
            const msg: Message = {
              id: d.id || d.message?.id || `ws-${Date.now()}`,
              channelId: d.channel_id || d.message?.channel_id || channelId,
              author: d.author_email || d.author_type || 'Unknown',
              authorId: d.author_id || d.message?.author_id || '',
              authorType: (d.author_type === 'agent') ? 'agent' : 'human',
              text: d.text || d.message?.text || '',
              timestamp: d.created_at || d.message?.created_at || new Date().toISOString(),
              isOwnMessage: d.author_id === localStorage.getItem('user-id'),
            };

            // Only add if it's for the current channel
            if (msg.channelId === channelId) {
              setMessages((prev) => {
                // Deduplicate by id
                if (prev.some(m => m.id === msg.id)) return prev;
                return [...prev, msg];
              });
            }
          }
        } catch (err) {
          console.error('Failed to parse WS message:', err);
        }
      };

      ws.onerror = () => {
        setError('WebSocket error');
        setIsConnected(false);
      };

      ws.onclose = () => {
        setIsConnected(false);
        
        // Reconnect with exponential backoff
        const attempts = reconnectAttemptsRef.current;
        if (attempts < 10) {
          const delay = Math.min(30000, 1000 * Math.pow(2, attempts));
          reconnectAttemptsRef.current++;
          console.log(`Reconnecting in ${delay}ms (attempt ${attempts + 1})`);
          reconnectTimeoutRef.current = setTimeout(connect, delay);
        } else {
          setError('Connection lost. Please refresh.');
        }
      };
    }

    connect();

    return () => {
      if (reconnectTimeoutRef.current) {
        clearTimeout(reconnectTimeoutRef.current);
      }
      if (wsRef.current) {
        wsRef.current.close();
        wsRef.current = null;
      }
    };
  }, [channelId]);

  const sendMessage = useCallback((text: string) => {
    if (wsRef.current && wsRef.current.readyState === WebSocket.OPEN) {
      wsRef.current.send(JSON.stringify({
        type: 'message',
        channel_id: channelId,
        text,
      }));
    }
  }, [channelId]);

  return {
    isConnected,
    messages,
    sendMessage,
    error
  };
}
