import { getAuthToken } from './authService';
import type { Message, WebSocketMessage } from '../types/chat';

export class WebSocketService {
  private ws: WebSocket | null = null;
  private url: string;
  private onMessage: ((message: WebSocketMessage) => void) | null = null;
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 5;
  private reconnectDelay = 1000; // Start with 1 second
  private reconnectTimer: number | null = null;

  constructor(url: string) {
    this.url = url;
  }

  /**
   * Connect to WebSocket server
   */
  connect(onMessage: (message: WebSocketMessage) => void): void {
    this.onMessage = onMessage;

    try {
      // Add auth token to WebSocket URL
      const token = getAuthToken();
      const wsUrl = token ? `${this.url}?token=${encodeURIComponent(token)}` : this.url;
      
      console.log('Connecting to WebSocket:', wsUrl.replace(/token=[^&]+/, 'token=***'));
      
      this.ws = new WebSocket(wsUrl);

      this.ws.onopen = () => {
        console.log('WebSocket connected');
        this.reconnectAttempts = 0;
        this.reconnectDelay = 1000;
        this.onMessage?.({
          type: 'connected',
          data: { status: 'connected' }
        });
      };

      this.ws.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data);
          this.handleServerMessage(data);
        } catch (error) {
          console.error('Failed to parse WebSocket message:', error);
        }
      };

      this.ws.onerror = (error) => {
        console.error('WebSocket error:', error);
        this.onMessage?.({
          type: 'error',
          data: { message: 'WebSocket connection error' }
        });
      };

      this.ws.onclose = (event) => {
        console.log('WebSocket closed:', event.code, event.reason);
        this.ws = null;
        
        // Attempt reconnection if it wasn't a clean close
        if (!event.wasClean && this.reconnectAttempts < this.maxReconnectAttempts) {
          this.scheduleReconnect();
        }
      };

    } catch (error) {
      console.error('Failed to create WebSocket connection:', error);
      this.onMessage?.({
        type: 'error',
        data: { message: 'Failed to create WebSocket connection' }
      });
    }
  }

  /**
   * Send a message through WebSocket
   */
  send(text: string, channelId: string): void {
    if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
      console.warn('WebSocket not connected, cannot send message');
      return;
    }

    const message = {
      type: 'send_message',
      data: {
        channelId,
        text
      }
    };

    this.ws.send(JSON.stringify(message));
  }

  /**
   * Disconnect WebSocket
   */
  disconnect(): void {
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer);
      this.reconnectTimer = null;
    }

    if (this.ws) {
      this.ws.close(1000, 'Client disconnect');
      this.ws = null;
    }
  }

  /**
   * Check if WebSocket is connected
   */
  isConnected(): boolean {
    return this.ws?.readyState === WebSocket.OPEN;
  }

  /**
   * Handle incoming server messages
   */
  private handleServerMessage(data: any): void {
    switch (data.type) {
      case 'message': {
        // Server broadcasts message fields at top level (not nested)
        const m = data.message || data;
        if (m.text || m.content) {
          const message: Message = {
            id: m.message_id || m.id || '',
            channelId: m.channel_id || m.channelId || '',
            author: m.author_email || m.author?.name || 'Unknown',
            authorId: m.author_id || m.user_id || m.author?.id || 'unknown',
            authorType: (m.author_type === 'agent') ? 'agent' : 'human',
            text: m.text || m.content,
            timestamp: m.created_at || m.timestamp || new Date().toISOString(),
            isOwnMessage: false
          };

          this.onMessage?.({
            type: 'message',
            message
          });
        }
      }
        break;

      case 'typing':
        // Handle typing indicators if needed
        break;

      case 'error':
        this.onMessage?.({
          type: 'error',
          data: data.data || { message: 'Server error' }
        });
        break;

      default:
        console.log('Unknown server message type:', data.type, data);
    }
  }

  /**
   * Schedule a reconnection attempt
   */
  private scheduleReconnect(): void {
    if (this.reconnectTimer) return;

    this.reconnectAttempts++;
    console.log(`Scheduling WebSocket reconnect attempt ${this.reconnectAttempts}/${this.maxReconnectAttempts} in ${this.reconnectDelay}ms`);

    this.reconnectTimer = setTimeout(() => {
      this.reconnectTimer = null;
      if (this.onMessage) {
        this.connect(this.onMessage);
      }
      
      // Exponential backoff
      this.reconnectDelay = Math.min(this.reconnectDelay * 2, 30000); // Max 30 seconds
    }, this.reconnectDelay);
  }
}