import { getApiBaseUrl } from './apiConfig';
import { getAuthToken } from './authService';
import type { Channel, Message } from '../types/chat';
import { getDisplayName } from '../lib/displayName';

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
  unread_count?: number;
}

interface ApiAgent {
  id: string;
  name: string;
  display_name?: string;
  email?: string;
}

interface ApiUser {
  id: string;
  email?: string;
  display_name?: string;
  type?: 'human' | 'agent';
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
    author: getDisplayName(apiMsg.author_email || apiMsg.author_id),
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
    memberCount: apiChannel.member_count,
    unread: apiChannel.unread_count ?? 0,
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
   * Create a new channel
   */
  async createChannel(name: string, description?: string): Promise<Channel> {
    try {
      const body: Record<string, string> = { name };
      if (description) body.topic = description;
      const response = await apiRequest<{ channel: ApiChannel }>(
        '/api/v1/channels',
        {
          method: 'POST',
          body: JSON.stringify(body)
        }
      );
      return mapApiChannel(response.channel);
    } catch (error) {
      console.error('Failed to create channel:', error);
      throw error;
    }
  },

  /**
   * Update a channel's description
   */
  async updateChannel(channelId: string, updates: { description?: string }): Promise<Channel> {
    const response = await apiRequest<{ channel: ApiChannel }>(
      `/api/v1/channels/${channelId}`,
      { method: 'PATCH', body: JSON.stringify(updates) }
    );
    return mapApiChannel(response.channel);
  },

  /**
   * Delete a channel
   */
  async deleteChannel(channelId: string): Promise<void> {
    await apiRequest<void>(`/api/v1/channels/${channelId}`, { method: 'DELETE' });
  },

  /**
   * Get channel members
   */
  async getMembers(channelId: string): Promise<{ id: string; email: string; role: string }[]> {
    const response = await apiRequest<{ members: { id: string; email: string; role: string }[] }>(
      `/api/v1/channels/${channelId}/members`
    );
    return response.members || [];
  },

  /**
   * Search messages
   */
  async searchMessages(query: string, channelId?: string): Promise<Message[]> {
    const params = new URLSearchParams({ q: query });
    if (channelId) params.set('channel_id', channelId);
    const response = await apiRequest<{ messages: ApiMessage[]; count: number }>(
      `/api/v1/messages/search?${params}`
    );
    const currentUserId = localStorage.getItem('user-id') || 'current-user';
    return (response.messages || []).map(msg => mapApiMessage(msg, currentUserId));
  },

  /**
   * Edit a message
   */
  async editMessage(channelId: string, messageId: string, text: string): Promise<Message> {
    const response = await apiRequest<{ message: ApiMessage }>(
      `/api/v1/channels/${channelId}/messages/${messageId}`,
      { method: 'PATCH', body: JSON.stringify({ text }) }
    );
    const currentUserId = localStorage.getItem('user-id') || 'current-user';
    return mapApiMessage(response.message, currentUserId);
  },

  /**
   * Delete a message
   */
  async deleteMessage(channelId: string, messageId: string): Promise<void> {
    await apiRequest<void>(
      `/api/v1/channels/${channelId}/messages/${messageId}`,
      { method: 'DELETE' }
    );
  },

  /**
   * Create a DM conversation
   */
  async createDM(userId: string): Promise<Channel> {
    const response = await apiRequest<{ channel: ApiChannel }>(
      '/api/v1/dm',
      { method: 'POST', body: JSON.stringify({ user_id: userId }) }
    );
    return mapApiChannel(response.channel);
  },

  /**
   * List DM conversations
   */
  async listDMs(): Promise<Channel[]> {
    const response = await apiRequest<{ channels: ApiChannel[] } | ApiChannel[]>('/api/v1/dm');
    const channels = Array.isArray(response) ? response : response.channels || [];
    return channels.map(mapApiChannel);
  },

  /**
   * Mark a channel as read (clear unread count)
   */
  async markChannelRead(channelId: string, lastReadMessageId?: string): Promise<void> {
    const body = lastReadMessageId ? { last_read_message_id: lastReadMessageId } : {};
    await apiRequest<void>(
      `/api/v1/channels/${channelId}/read`,
      { method: 'POST', body: JSON.stringify(body) }
    );
  },

  async markDMRead(dmId: string, lastReadMessageId?: string): Promise<void> {
    const body = lastReadMessageId ? { last_read_message_id: lastReadMessageId } : {};
    await apiRequest<void>(
      `/api/v1/dm/${dmId}/read`,
      { method: 'POST', body: JSON.stringify(body) }
    );
  },

  async addChannelMember(channelId: string, userId: string, role: 'member' | 'observer' = 'member'): Promise<void> {
    await apiRequest<void>(
      `/api/v1/channels/${channelId}/members`,
      { method: 'POST', body: JSON.stringify({ user_id: userId, role }) }
    );
  },

  async getAgents(): Promise<ApiAgent[]> {
    const response = await apiRequest<{ agents?: ApiAgent[] } | ApiAgent[]>('/api/v1/agents');
    if (Array.isArray(response)) return response;
    return response.agents || [];
  },

  async getUsers(): Promise<ApiUser[]> {
    const response = await apiRequest<{ users?: ApiUser[] } | ApiUser[]>('/api/v1/users');
    if (Array.isArray(response)) return response;
    return response.users || [];
  },

  /**
   * Get unread counts for all conversations
   */
  async getUnreadCounts(): Promise<{
    total_unread: number;
    channels: { id: string; name: string; unread_count: number }[];
    dms: { id: string; unread_count: number }[];
  }> {
    return apiRequest<{
      total_unread: number;
      channels: { id: string; name: string; unread_count: number }[];
      dms: { id: string; unread_count: number }[];
    }>('/api/v1/unread');
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