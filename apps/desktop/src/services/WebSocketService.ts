export interface MessageData {
  id: string;
  channelId: string;
  content: string;
  author: {
    id: string;
    name: string;
    type: 'agent' | 'human';
    avatar?: string;
  };
  timestamp: string;
  mentions?: Array<{
    id: string;
    name: string;
    type: 'agent' | 'human';
    start: number;
    end: number;
  }>;
  attachments?: Array<{
    id: string;
    name: string;
    url: string;
    type: string;
  }>;
  reactions?: Array<{
    emoji: string;
    count: number;
    users: string[];
  }>;
}

export interface ChannelData {
  id: string;
  name: string;
  type: 'channel' | 'dm';
  topic?: string;
  lastMessage?: {
    content: string;
    timestamp: string;
    author: string;
  };
  unreadCount: number;
  members?: Array<{
    id: string;
    name: string;
    type: 'agent' | 'human';
    status: 'online' | 'offline' | 'away';
  }>;
}

export interface UserData {
  id: string;
  name: string;
  type: 'agent' | 'human';
  status: 'online' | 'offline' | 'away';
  avatar?: string;
}

export interface WebSocketMessage {
  type: 'message.created' | 'message.updated' | 'channel.created' | 'user.presence' | 'typing.start' | 'typing.stop';
  data: any;
}

export interface WebSocketServiceConfig {
  url: string;
  token?: string;
  reconnectInterval?: number;
  maxReconnectAttempts?: number;
}

export class WebSocketService {
  private ws: WebSocket | null = null;
  private config: WebSocketServiceConfig;
  private reconnectAttempts = 0;
  private reconnectTimer: NodeJS.Timeout | null = null;
  private isConnecting = false;
  private listeners: { [key: string]: Array<(data: any) => void> } = {};
  private connectionListeners: Array<(connected: boolean) => void> = [];

  constructor(config: WebSocketServiceConfig) {
    this.config = {
      reconnectInterval: 1000,
      maxReconnectAttempts: 10,
      ...config
    };
  }

  connect(): void {
    if (this.ws?.readyState === WebSocket.OPEN || this.isConnecting) {
      return;
    }

    this.isConnecting = true;

    try {
      const wsUrl = new URL(this.config.url);
      if (this.config.token) {
        wsUrl.searchParams.set('token', this.config.token);
      }
      
      console.log('Connecting to WebSocket:', wsUrl.toString());
      console.log('Token present:', !!this.config.token);
      console.log('Token length:', this.config.token?.length || 0);

      this.ws = new WebSocket(wsUrl.toString());
      
      this.ws.onopen = () => {
        console.log('WebSocket connected');
        this.isConnecting = false;
        this.reconnectAttempts = 0;
        this.notifyConnectionListeners(true);
      };

      this.ws.onmessage = (event) => {
        try {
          const message: WebSocketMessage = JSON.parse(event.data);
          this.handleMessage(message);
        } catch (error) {
          console.error('Error parsing WebSocket message:', error);
        }
      };

      this.ws.onerror = (error) => {
        console.error('WebSocket error:', error);
        this.isConnecting = false;
      };

      this.ws.onclose = (event) => {
        console.log('WebSocket closed:', event.code, event.reason);
        this.isConnecting = false;
        this.notifyConnectionListeners(false);
        
        if (!event.wasClean && this.reconnectAttempts < this.config.maxReconnectAttempts!) {
          this.scheduleReconnect();
        }
      };

    } catch (error) {
      console.error('Error creating WebSocket connection:', error);
      this.isConnecting = false;
      this.scheduleReconnect();
    }
  }

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

  isConnected(): boolean {
    return this.ws?.readyState === WebSocket.OPEN;
  }

  send(message: any): void {
    if (!this.isConnected()) {
      console.error('Cannot send message: WebSocket not connected');
      return;
    }

    try {
      this.ws!.send(JSON.stringify(message));
    } catch (error) {
      console.error('Error sending WebSocket message:', error);
    }
  }

  // Message sending methods
  sendMessage(channelId: string, content: string, mentions?: Array<{id: string; name: string; type: 'agent' | 'human'}>): void {
    this.send({
      type: 'message.send',
      data: {
        channelId,
        content,
        mentions: mentions?.map(m => ({ id: m.id, type: m.type }))
      }
    });
  }

  startTyping(channelId: string): void {
    this.send({
      type: 'typing.start',
      data: { channelId }
    });
  }

  stopTyping(channelId: string): void {
    this.send({
      type: 'typing.stop',
      data: { channelId }
    });
  }

  // Event listeners
  on(eventType: string, callback: (data: any) => void): () => void {
    if (!this.listeners[eventType]) {
      this.listeners[eventType] = [];
    }
    this.listeners[eventType].push(callback);

    // Return unsubscribe function
    return () => {
      const callbacks = this.listeners[eventType];
      if (callbacks) {
        const index = callbacks.indexOf(callback);
        if (index > -1) {
          callbacks.splice(index, 1);
        }
      }
    };
  }

  onConnectionChange(callback: (connected: boolean) => void): () => void {
    this.connectionListeners.push(callback);
    
    // Return unsubscribe function
    return () => {
      const index = this.connectionListeners.indexOf(callback);
      if (index > -1) {
        this.connectionListeners.splice(index, 1);
      }
    };
  }

  private handleMessage(message: WebSocketMessage): void {
    const callbacks = this.listeners[message.type] || [];
    callbacks.forEach(callback => {
      try {
        callback(message.data);
      } catch (error) {
        console.error('Error in WebSocket message handler:', error);
      }
    });
  }

  private scheduleReconnect(): void {
    if (this.reconnectTimer) {
      return;
    }

    const delay = Math.min(
      this.config.reconnectInterval! * Math.pow(2, this.reconnectAttempts),
      30000 // Max 30 seconds
    );

    console.log(`Reconnecting in ${delay}ms (attempt ${this.reconnectAttempts + 1}/${this.config.maxReconnectAttempts})`);

    this.reconnectTimer = setTimeout(() => {
      this.reconnectTimer = null;
      this.reconnectAttempts++;
      this.connect();
    }, delay);
  }

  private notifyConnectionListeners(connected: boolean): void {
    this.connectionListeners.forEach(callback => {
      try {
        callback(connected);
      } catch (error) {
        console.error('Error in connection listener:', error);
      }
    });
  }
}

// Singleton instance
let webSocketService: WebSocketService | null = null;

export const getWebSocketService = (): WebSocketService => {
  if (!webSocketService) {
    // Default config - should be overridden by app initialization
    // Use import.meta.env for Vite (browser environment)
    webSocketService = new WebSocketService({
      url: import.meta.env.VITE_WEBSOCKET_URL || 'ws://localhost:8080/ws',
      token: import.meta.env.VITE_AUTH_TOKEN
    });
  }
  return webSocketService;
};

export const initializeWebSocketService = (config: WebSocketServiceConfig): WebSocketService => {
  webSocketService = new WebSocketService(config);
  return webSocketService;
};