import type { Channel, Message } from '../types/chat';
import { getApiBaseUrl } from './apiConfig';

function getAuthToken(): string | null {
  return localStorage.getItem('auth-token');
}

function getUserId(): string | null {
  return localStorage.getItem('user-id');
}

async function apiFetch(path: string, options: RequestInit = {}): Promise<Response> {
  const token = getAuthToken();
  const baseUrl = getApiBaseUrl();
  
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    ...(options.headers as Record<string, string> || {}),
  };
  
  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }

  const response = await fetch(`${baseUrl}${path}`, {
    ...options,
    headers,
  });

  if (response.status === 401) {
    localStorage.removeItem('auth-token');
    window.location.href = '/login';
    throw new Error('Unauthorized');
  }

  return response;
}

export async function fetchChannels(): Promise<Channel[]> {
  const response = await apiFetch('/api/v1/channels');
  if (!response.ok) {
    throw new Error(`Failed to fetch channels: ${response.status}`);
  }
  const data = await response.json();
  const channels = data.channels || data;
  if (!Array.isArray(channels)) return [];
  return channels.map((ch: any) => ({
    id: ch.id || ch.channel_id,
    name: ch.name,
    topic: ch.topic || '',
    memberCount: ch.member_count || ch.members?.length || 0,
  }));
}

export async function fetchMessages(channelId: string): Promise<Message[]> {
  const response = await apiFetch(`/api/v1/channels/${channelId}/messages`);
  if (!response.ok) {
    throw new Error(`Failed to fetch messages: ${response.status}`);
  }
  const data = await response.json();
  const messages = data.messages || data;
  if (!Array.isArray(messages)) return [];
  const currentUserId = getUserId();
  
  return messages.map((msg: any) => ({
    id: msg.id,
    channelId: msg.channel_id || channelId,
    author: msg.author_email || msg.author_type || 'Unknown',
    authorId: msg.author_id,
    authorType: msg.author_type === 'agent' ? 'agent' as const : 'human' as const,
    text: msg.text,
    timestamp: msg.created_at,
    isOwnMessage: msg.author_id === currentUserId,
    attachmentUrl: msg.attachment_url,
    attachmentName: msg.attachment_name,
  }));
}

export async function sendMessageApi(channelId: string, text: string): Promise<Message> {
  const response = await apiFetch(`/api/v1/channels/${channelId}/messages`, {
    method: 'POST',
    body: JSON.stringify({ text }),
  });
  if (!response.ok) {
    throw new Error(`Failed to send message: ${response.status}`);
  }
  const data = await response.json();
  const msg = data.message || data;
  const currentUserId = getUserId();
  
  return {
    id: msg.id,
    channelId: msg.channel_id || channelId,
    author: msg.author_email || 'You',
    authorId: msg.author_id,
    authorType: msg.author_type === 'agent' ? 'agent' as const : 'human' as const,
    text: msg.text,
    timestamp: msg.created_at,
    isOwnMessage: msg.author_id === currentUserId || true,
    attachmentUrl: msg.attachment_url,
    attachmentName: msg.attachment_name,
  };
}

export async function sendMessageWithAttachment(channelId: string, text: string, file: File): Promise<Message> {
  const token = getAuthToken();
  const baseUrl = getApiBaseUrl();
  
  // Validate file size (max 10MB)
  const maxSizeBytes = 10 * 1024 * 1024; // 10MB
  if (file.size > maxSizeBytes) {
    throw new Error('File size must be less than 10MB');
  }

  const formData = new FormData();
  formData.append('text', text);
  formData.append('file', file);

  const headers: Record<string, string> = {};
  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }

  const response = await fetch(`${baseUrl}/api/v1/channels/${channelId}/messages`, {
    method: 'POST',
    body: formData,
    headers,
  });

  if (response.status === 401) {
    localStorage.removeItem('auth-token');
    window.location.href = '/login';
    throw new Error('Unauthorized');
  }

  if (!response.ok) {
    throw new Error(`Failed to send message: ${response.status}`);
  }

  const data = await response.json();
  const msg = data.message || data;
  const currentUserId = getUserId();
  
  return {
    id: msg.id,
    channelId: msg.channel_id || channelId,
    author: msg.author_email || 'You',
    authorId: msg.author_id,
    authorType: msg.author_type === 'agent' ? 'agent' as const : 'human' as const,
    text: msg.text,
    timestamp: msg.created_at,
    isOwnMessage: msg.author_id === currentUserId || true,
    attachmentUrl: msg.attachment_url,
    attachmentName: msg.attachment_name,
  };
}
