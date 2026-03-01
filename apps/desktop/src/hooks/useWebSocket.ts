import { useState, useEffect, useCallback, useRef } from 'react';
import { getWebSocketService, initializeWebSocketService, MessageData, ChannelData, UserData } from '../services/WebSocketService';

export interface UseWebSocketOptions {
  autoConnect?: boolean;
  token?: string;
  url?: string;
}

export interface UseWebSocketReturn {
  // Connection state
  connected: boolean;
  connecting: boolean;
  connect: () => void;
  disconnect: () => void;
  
  // Message operations
  sendMessage: (channelId: string, content: string, mentions?: Array<{id: string; name: string; type: 'agent' | 'human'}>) => void;
  messages: { [channelId: string]: MessageData[] };
  
  // Channel operations
  channels: ChannelData[];
  activeChannel: string | null;
  setActiveChannel: (channelId: string | null) => void;
  
  // User operations
  users: UserData[];
  currentUser: UserData | null;
  
  // Typing indicators
  typingUsers: { [channelId: string]: UserData[] };
  startTyping: (channelId: string) => void;
  stopTyping: (channelId: string) => void;
  
  // Utility
  getChannelById: (channelId: string) => ChannelData | undefined;
  getUserById: (userId: string) => UserData | undefined;
}

export const useWebSocket = (options: UseWebSocketOptions = {}): UseWebSocketReturn => {
  const { autoConnect = true, token, url } = options;
  
  const [connected, setConnected] = useState(false);
  const [connecting, setConnecting] = useState(false);
  const [messages, setMessages] = useState<{ [channelId: string]: MessageData[] }>({});
  const [channels, setChannels] = useState<ChannelData[]>([]);
  const [users, setUsers] = useState<UserData[]>([]);
  const [currentUser, setCurrentUser] = useState<UserData | null>(null);
  const [activeChannel, setActiveChannel] = useState<string | null>(null);
  const [typingUsers, setTypingUsers] = useState<{ [channelId: string]: UserData[] }>({});
  
  const wsService = getWebSocketService();
  const typingTimeoutRefs = useRef<{ [key: string]: NodeJS.Timeout }>({});
  
  // Initialize WebSocket service with config if provided
  useEffect(() => {
    if (url || token) {
      // Re-initialize service with new config
      initializeWebSocketService({
        url: url || 'ws://localhost:8080/ws',
        token
      });
    }
  }, [url, token]);

  // Connection management
  const connect = useCallback(() => {
    setConnecting(true);
    wsService.connect();
  }, [wsService]);

  const disconnect = useCallback(() => {
    wsService.disconnect();
  }, [wsService]);

  // Message operations
  const sendMessage = useCallback((channelId: string, content: string, mentions?: Array<{id: string; name: string; type: 'agent' | 'human'}>) => {
    wsService.sendMessage(channelId, content, mentions);
  }, [wsService]);

  // Typing indicators
  const startTyping = useCallback((channelId: string) => {
    wsService.startTyping(channelId);
  }, [wsService]);

  const stopTyping = useCallback((channelId: string) => {
    wsService.stopTyping(channelId);
  }, [wsService]);

  // Utility functions
  const getChannelById = useCallback((channelId: string): ChannelData | undefined => {
    return channels.find(channel => channel.id === channelId);
  }, [channels]);

  const getUserById = useCallback((userId: string): UserData | undefined => {
    return users.find(user => user.id === userId);
  }, [users]);

  // WebSocket event handlers
  useEffect(() => {
    // Connection status
    const unsubscribeConnection = wsService.onConnectionChange(async (isConnected) => {
      setConnected(isConnected);
      setConnecting(false);
      
      // Fetch initial data when connected
      if (isConnected && token) {
        try {
          // Fetch channels
          const channelsRes = await fetch('http://localhost:8080/api/v1/channels', {
            headers: {
              'Authorization': `Bearer ${token}`
            }
          });
          console.log('Channels fetch response:', channelsRes.status);
          if (channelsRes.ok) {
            const channelsData = await channelsRes.json();
            console.log('Channels data:', channelsData);
            
            // Map API response to ChannelData interface
            const mappedChannels: ChannelData[] = (channelsData.channels || []).map((ch: any) => ({
              id: ch.id,
              name: ch.name,
              type: 'channel' as const, // API doesn't return type, assume all are channels
              topic: ch.topic,
              unreadCount: 0, // API doesn't return unread count yet
              lastMessage: undefined,
              members: []
            }));
            
            console.log('Mapped channels:', mappedChannels);
            setChannels(mappedChannels);
            console.log('Channels set:', mappedChannels.length, 'channels');
          } else {
            console.error('Failed to fetch channels:', channelsRes.status, await channelsRes.text());
          }
          
          // Fetch current user info
          const meRes = await fetch('http://localhost:8080/api/v1/users/me', {
            headers: {
              'Authorization': `Bearer ${token}`
            }
          });
          if (meRes.ok) {
            const meData = await meRes.json();
            setCurrentUser(meData.user || null);
          }
        } catch (error) {
          console.error('Failed to fetch initial data:', error);
        }
      }
    });

    // Message events
    const unsubscribeMessage = wsService.on('message.created', (data: MessageData) => {
      setMessages(prev => ({
        ...prev,
        [data.channelId]: [...(prev[data.channelId] || []), data]
      }));

      // Update channel last message
      setChannels(prev => prev.map(channel => 
        channel.id === data.channelId
          ? {
              ...channel,
              lastMessage: {
                content: data.content,
                timestamp: data.timestamp,
                author: data.author.name
              },
              unreadCount: channel.id === activeChannel ? 0 : channel.unreadCount + 1
            }
          : channel
      ));
    });

    const unsubscribeMessageUpdate = wsService.on('message.updated', (data: MessageData) => {
      setMessages(prev => ({
        ...prev,
        [data.channelId]: (prev[data.channelId] || []).map(msg => 
          msg.id === data.id ? data : msg
        )
      }));
    });

    // Channel events
    const unsubscribeChannelCreated = wsService.on('channel.created', (data: ChannelData) => {
      setChannels(prev => [...prev, data]);
    });

    // User presence events
    const unsubscribeUserPresence = wsService.on('user.presence', (data: { user: UserData; status: 'online' | 'offline' | 'away' }) => {
      setUsers(prev => prev.map(user => 
        user.id === data.user.id 
          ? { ...user, status: data.status }
          : user
      ));
    });

    // Typing events
    const unsubscribeTypingStart = wsService.on('typing.start', (data: { channelId: string; user: UserData }) => {
      setTypingUsers(prev => {
        const channelTyping = prev[data.channelId] || [];
        if (channelTyping.find(u => u.id === data.user.id)) {
          return prev; // User already typing
        }
        return {
          ...prev,
          [data.channelId]: [...channelTyping, data.user]
        };
      });

      // Clear typing after timeout
      const key = `${data.channelId}-${data.user.id}`;
      if (typingTimeoutRefs.current[key]) {
        clearTimeout(typingTimeoutRefs.current[key]);
      }
      typingTimeoutRefs.current[key] = setTimeout(() => {
        setTypingUsers(prev => ({
          ...prev,
          [data.channelId]: (prev[data.channelId] || []).filter(u => u.id !== data.user.id)
        }));
        delete typingTimeoutRefs.current[key];
      }, 3000);
    });

    const unsubscribeTypingStop = wsService.on('typing.stop', (data: { channelId: string; user: UserData }) => {
      setTypingUsers(prev => ({
        ...prev,
        [data.channelId]: (prev[data.channelId] || []).filter(u => u.id !== data.user.id)
      }));

      const key = `${data.channelId}-${data.user.id}`;
      if (typingTimeoutRefs.current[key]) {
        clearTimeout(typingTimeoutRefs.current[key]);
        delete typingTimeoutRefs.current[key];
      }
    });

    return () => {
      unsubscribeConnection();
      unsubscribeMessage();
      unsubscribeMessageUpdate();
      unsubscribeChannelCreated();
      unsubscribeUserPresence();
      unsubscribeTypingStart();
      unsubscribeTypingStop();
      
      // Clear all typing timeouts
      Object.values(typingTimeoutRefs.current).forEach(timeout => {
        clearTimeout(timeout);
      });
      typingTimeoutRefs.current = {};
    };
  }, [wsService, activeChannel]);

  // Auto-connect on mount
  useEffect(() => {
    if (autoConnect) {
      connect();
    }

    return () => {
      disconnect();
    };
  }, [autoConnect, connect, disconnect]);

  // Mark channel as read when active channel changes
  useEffect(() => {
    if (activeChannel) {
      setChannels(prev => prev.map(channel => 
        channel.id === activeChannel
          ? { ...channel, unreadCount: 0 }
          : channel
      ));
    }
  }, [activeChannel]);

  // Load initial data when connected
  useEffect(() => {
    if (connected) {
      // Mock initial data - in real app, this would come from API/WebSocket
      setChannels([
        {
          id: 'ch_general',
          name: 'general',
          type: 'channel',
          topic: 'General discussion and updates',
          unreadCount: 0,
          lastMessage: {
            content: 'Welcome to the workspace!',
            timestamp: new Date().toISOString(),
            author: 'System'
          }
        },
        {
          id: 'ch_research',
          name: 'research',
          type: 'channel',
          topic: 'Research findings and data analysis',
          unreadCount: 2
        }
      ]);

      setUsers([
        {
          id: 'user_1',
          name: 'Research Agent',
          type: 'agent',
          status: 'online'
        },
        {
          id: 'user_2',
          name: 'Alice Smith',
          type: 'human',
          status: 'online'
        }
      ]);

      setCurrentUser({
        id: 'user_current',
        name: 'Current User',
        type: 'human',
        status: 'online'
      });

      // Set default active channel
      if (!activeChannel) {
        setActiveChannel('ch_general');
      }
    }
  }, [connected, activeChannel]);

  return {
    connected,
    connecting,
    connect,
    disconnect,
    sendMessage,
    messages,
    channels,
    activeChannel,
    setActiveChannel,
    users,
    currentUser,
    typingUsers,
    startTyping,
    stopTyping,
    getChannelById,
    getUserById
  };
};