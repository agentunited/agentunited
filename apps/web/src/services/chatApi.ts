import { getApiBaseUrl } from './apiConfig';
import { getAuthToken } from './authService';
import type { Channel, Message } from '../types/chat';

// API interfaces for the Agent United backend
interface ApiMessage {
  id: string;
  channel_id: string;
  author_id: string;
  author_type: string; // "agent" | "user"
  author_email: string; // display name for agents, email for users
  text: string;
  created_at: string;
}

interface ApiChannel {
  id: string;
  name: string;
  topic?: string;
  description?: string;
  member_count?: number;
}

interface SendMessageRequest {
  text: string;
}

class ChatApiError extends Error {
  status?: number;
  constructor(message: string, status?: number) {
    super(message);
    this.name = 'ChatApiError';
    this.status = status;
  }
}

function getAuthHeaders(): HeadersInit {
  const token = getAuthToken();
  return {
    'Content-Type': 'application/json',
    ...(token ? { Authorization: `Bearer ${token}` } : {})
  };
}

async function apiRequest<T>(
  endpoint: string,
  options: RequestInit = {}
): Promise<T> {
  const url = `${getApiBaseUrl()}${endpoint}`;
  
  const response = await fetch(url, {
    ...options,
    headers: {
      ...getAuthHeaders(),
      ...options.headers
    }
  });

  if (!response.ok) {
    let errorMessage = `Request failed (${response.status})`;
    try {
      const errorData = await response.json();
      errorMessage = errorData.message || errorData.error || errorMessage;
    } catch {
      // Keep default error message
    }
    throw new ChatApiError(errorMessage, response.status);
  }

  return response.json();
}

// Convert API message format to frontend format
function mapApiMessage(apiMsg: ApiMessage, currentUserId?: string): Message {
  return {
    id: apiMsg.id,
    channelId: apiMsg.channel_id,
    author: apiMsg.author_email || apiMsg.author_id,
    authorId: apiMsg.author_id,
    authorType: apiMsg.author_type === 'agent' ? 'agent' : 'human',
    text: apiMsg.text,
    timestamp: apiMsg.created_at,
    isOwnMessage: currentUserId ? apiMsg.author_id === currentUserId : false
  };
}

// Convert API channel format to frontend format  
function mapApiChannel(apiChannel: ApiChannel): Channel {
  return {
    id: apiChannel.id,
    name: apiChannel.name,
    topic: apiChannel.topic || apiChannel.description || '',
    memberCount: apiChannel.member_count
  };
}

export const chatApi = {
  /**
   * Fetch all channels user has access to
   */
  async getChannels(): Promise<Channel[]> {
    try {
      const response = await apiRequest<{ channels: ApiChannel[] } | ApiChannel[]>('/api/v1/channels');
      const apiChannels = Array.isArray(response) ? response : response.channels || [];
      return apiChannels.map(mapApiChannel);
    } catch (error) {
      console.error('Failed to fetch channels:', error);
      throw error;
    }
  },

  /**
   * Fetch messages for a specific channel
   */
  async getMessages(channelId: string, limit?: number): Promise<Message[]> {
    try {
      const params = limit ? `?limit=${limit}` : '';
      const response = await apiRequest<{ messages: ApiMessage[] } | ApiMessage[]>(
        `/api/v1/channels/${channelId}/messages${params}`
      );
      const apiMessages = Array.isArray(response) ? response : response.messages || [];
      
      const currentUserId = localStorage.getItem('user-id') || 'current-user';
      
      return apiMessages.map(msg => mapApiMessage(msg, currentUserId));
    } catch (error) {
      console.error(`Failed to fetch messages for channel ${channelId}:`, error);
      throw error;
    }
  },

  /**
   * Send a message to a channel
   */
  async sendMessage(channelId: string, text: string): Promise<Message> {
    try {
      const request: SendMessageRequest = { text };
      const response = await apiRequest<{ message: ApiMessage }>(
        `/api/v1/channels/${channelId}/messages`,
        {
          method: 'POST',
          body: JSON.stringify(request)
        }
      );
      
      const currentUserId = localStorage.getItem('user-id') || 'current-user';
      return mapApiMessage(response.message, currentUserId);
    } catch (error) {
      console.error(`Failed to send message to channel ${channelId}:`, error);
      throw error;
    }
  },

  /**
   * Get WebSocket URL for real-time messages
   */
  getWebSocketUrl(): string {
    const baseUrl = getApiBaseUrl();
    const wsUrl = baseUrl.replace(/^http/, 'ws');
    return `${wsUrl}/ws`;
  }
};