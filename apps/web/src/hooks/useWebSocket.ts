import { useState, useEffect, useRef, useCallback } from 'react';
import type { Message } from '../types/chat';
import { getApiBaseUrl } from '../services/apiConfig';
import { fetchMessages } from '../services/api';
import { getDisplayName } from '../lib/displayName';
import { chatApi } from '../services/chatApi';

interface UseWebSocketReturn {
  isConnected: boolean;
  connectionStatus: 'connected' | 'reconnecting' | 'disconnected';
  messages: Message[];
  sendMessage: (text: string) => Promise<boolean>;
  error: string | null;
}

export function useWebSocket(_url: string, channelId: string): UseWebSocketReturn {
  const [isConnected, setIsConnected] = useState(false);
  const [connectionStatus, setConnectionStatus] = useState<'connected' | 'reconnecting' | 'disconnected'>(
    () => (localStorage.getItem('auth-token') ? 'reconnecting' : 'disconnected')
  );
  const [messages, setMessages] = useState<Message[]>([]);
  const [error, setError] = useState<string | null>(() => (localStorage.getItem('auth-token') ? null : 'Not authenticated'));

  const wsRef = useRef<WebSocket | null>(null);
  const reconnectTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const reconnectAttemptsRef = useRef(0);

  // Load initial messages from API
  useEffect(() => {
    let cancelled = false;
    
    async function loadMessages() {
      if (!channelId) {
        setMessages([]);
        return;
      }
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
    if (!channelId) return;
    
    const token = localStorage.getItem('auth-token');
    if (!token) {
      return;
    }

    const baseUrl = getApiBaseUrl();
    const wsProtocol = baseUrl.startsWith('https') ? 'wss' : 'ws';
    const wsHost = baseUrl.replace(/^https?:\/\//, '');
    const wsUrl = `${wsProtocol}://${wsHost}/ws?token=${token}`;

    function connect() {
      setConnectionStatus('reconnecting');
      const ws = new WebSocket(wsUrl);
      wsRef.current = ws;

      ws.onopen = () => {
        setIsConnected(true);
        setConnectionStatus('connected');
        setError(null);
        reconnectAttemptsRef.current = 0;
        console.log('WebSocket connected');
        // Subscribe to the current channel
        if (channelId) {
          ws.send(JSON.stringify({ type: 'subscribe', channel_id: channelId }));
        }
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
            const authorId = d.author_id || d.message?.author_id || '';
            const isOwnMessage = authorId === localStorage.getItem('user-id');

            const rawAuthorLabel =
              d.author_email ||
              d.author_display_name ||
              d.author_name ||
              d.message?.author_email ||
              d.message?.author_display_name ||
              d.message?.author_name ||
              (isOwnMessage ? localStorage.getItem('user-email') : '') ||
              authorId ||
              'Unknown';

            const msg: Message = {
              id: d.id || d.message_id || d.message?.id || `ws-${Date.now()}`,
              channelId: d.channel_id || d.message?.channel_id || channelId,
              author: getDisplayName(rawAuthorLabel),
              authorId,
              authorType: ((d.author_type || d.message?.author_type) === 'agent') ? 'agent' : 'human',
              text: d.text || d.message?.text || '',
              timestamp: d.created_at || d.message?.created_at || new Date().toISOString(),
              isOwnMessage,
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
        setConnectionStatus('disconnected');
      };

      ws.onclose = () => {
        setIsConnected(false);

        // Reconnect with exponential backoff
        const attempts = reconnectAttemptsRef.current;
        if (attempts < 10) {
          const delay = Math.min(30000, 1000 * Math.pow(2, attempts));
          reconnectAttemptsRef.current++;
          setConnectionStatus('reconnecting');
          console.log(`Reconnecting in ${delay}ms (attempt ${attempts + 1})`);
          reconnectTimeoutRef.current = setTimeout(connect, delay);
        } else {
          setConnectionStatus('disconnected');
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

  const sendMessage = useCallback(async (text: string) => {
    // Try WebSocket first
    if (wsRef.current && wsRef.current.readyState === WebSocket.OPEN) {
      wsRef.current.send(JSON.stringify({
        type: 'send_message',
        channel_id: channelId,
        text,
      }));
      return true;
    }
    
    // HTTP fallback when WebSocket unavailable (e.g., relay doesn't support WS upgrades)
    try {
      const message = await chatApi.sendMessage(channelId, text);
      // Add the sent message to local state so it appears immediately
      setMessages((prev) => {
        if (prev.some(m => m.id === message.id)) return prev;
        return [...prev, message];
      });
      return true;
    } catch (err) {
      console.error('Failed to send message via HTTP fallback:', err);
      return false;
    }
  }, [channelId]);

  return {
    isConnected,
    connectionStatus,
    messages,
    sendMessage,
    error
  };
}
