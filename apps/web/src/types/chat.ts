export interface Message {
  id: string;
  channelId: string;
  author: string;
  authorId: string;
  authorType: "agent" | "human";
  text: string;
  timestamp: string; // ISO 8601
  editedAt?: string; // ISO 8601
  isOwnMessage: boolean;
  attachmentUrl?: string;
  attachmentName?: string;
}

export interface Channel {
  id: string;
  name: string;
  topic: string;
  memberCount?: number;
  unread?: number;
}

export interface WebSocketMessage {
  type: 'message' | 'connected' | 'error' | 'typing';
  data?: any;
  message?: Message;
}
