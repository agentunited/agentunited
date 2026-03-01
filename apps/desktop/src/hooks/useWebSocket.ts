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
  const sendMessage = useCallback(async (channelId: string, content: string, mentions?: Array<{id: string; name: string; type: 'agent' | 'human'}>) => {
    if (!token) {
      console.error('Cannot send message: no auth token');
      return;
    }
    
    try {
      const res = await fetch(`http://localhost:8080/api/v1/channels/${channelId}/messages`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({ text: content })  // Backend expects 'text' not 'content'
      });
      
      if (!res.ok) {
        console.error('Failed to send message:', res.status, await res.text());
      } else {
        const data = await res.json();
        console.log('Message sent:', data);
        
        // Add message to local state immediately
        const newMessage: MessageData = {
          id: data.message.id,
          channelId,
          content,
          author: {
            id: data.message.author_id,
            name: data.message.author_email || 'You',  // Backend returns author_email
            type: 'agent'  // Default for now
          },
          timestamp: data.message.created_at,
          mentions: [],
          attachments: []
        };
        
        setMessages(prev => ({
          ...prev,
          [channelId]: [...(prev[channelId] || []), newMessage]
        }));
      }
    } catch (error) {
      console.error('Error sending message:', error);
    }
  }, [token]);

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
    const unsubscribeMessage = wsService.on('message.created', (data: any) => {
      // Map backend message format to MessageData
      const mappedMessage: MessageData = {
        id: data.id,
        channelId: data.channel_id,
        content: data.text,
        author: {
          id: data.author_id,
          name: data.author_email || 'Unknown',
          type: data.author_type === 'user' ? 'human' : 'agent'
        },
        timestamp: data.created_at,
        mentions: [],
        attachments: []
      };
      
      setMessages(prev => ({
        ...prev,
        [mappedMessage.channelId]: [...(prev[mappedMessage.channelId] || []), mappedMessage]
      }));

      // Update channel last message
      setChannels(prev => prev.map(channel => 
        channel.id === mappedMessage.channelId
          ? {
              ...channel,
              lastMessage: {
                content: mappedMessage.content,
                timestamp: mappedMessage.timestamp,
                author: mappedMessage.author.name
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

  // Fetch messages when active channel changes (initial load only, WebSocket handles updates)
  useEffect(() => {
    if (!connected || !activeChannel || !token) return;
    
    const fetchMessages = async () => {
      try {
        const res = await fetch(`http://localhost:8080/api/v1/channels/${activeChannel}/messages`, {
          headers: {
            'Authorization': `Bearer ${token}`
          }
        });
        
        if (res.ok) {
          const data = await res.json();
          
          // Map API response to MessageData format
          // Backend returns: { id, channel_id, author_id, author_type, author_email, text, created_at }
          const mappedMessages: MessageData[] = (data.messages || []).map((msg: any) => ({
            id: msg.id,
            channelId: activeChannel,
            content: msg.text,  // Backend uses 'text' field
            author: {
              id: msg.author_id,
              name: msg.author_email || 'Unknown',  // Backend returns author_email
              type: msg.author_type === 'user' ? 'human' : 'agent'  // Map 'user' to 'human'
            },
            timestamp: msg.created_at,
            mentions: [],
            attachments: []
          }));
          
          setMessages(prev => ({
            ...prev,
            [activeChannel]: mappedMessages
          }));
        } else {
          console.error('Failed to fetch messages:', res.status, await res.text());
        }
      } catch (error) {
        console.error('Error fetching messages:', error);
      }
    };
    
    fetchMessages();
  }, [connected, activeChannel, token]);

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