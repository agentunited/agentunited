import React, { useState } from 'react';
import { Avatar } from './Avatar';
import { Badge } from './Badge';
import { Button } from './Button';
import './Message.css';

export interface MessageData {
  id: string;
  authorId: string;
  authorName: string;
  authorType: 'agent' | 'human';
  content: string;
  timestamp: string;
  mentions?: string[];
  attachments?: Array<{
    id: string;
    name: string;
    url: string;
    type: 'image' | 'file';
    size?: number;
  }>;
  replyTo?: string;
  edited?: boolean;
}

export interface MessageProps {
  message: MessageData;
  isOwn?: boolean;
  showActions?: boolean;
  onReply?: (messageId: string) => void;
  onEdit?: (messageId: string) => void;
  onDelete?: (messageId: string) => void;
  onReact?: (messageId: string, reaction: string) => void;
  className?: string;
}

export function Message({
  message,
  isOwn = false,
  showActions = false,
  onReply,
  onEdit,
  onDelete,
  onReact,
  className = ''
}: MessageProps) {
  const [isHovered, setIsHovered] = useState(false);

  const messageClass = [
    'message',
    isOwn ? 'message--own' : '',
    isHovered ? 'message--hovered' : '',
    className,
  ].filter(Boolean).join(' ');

  const formatTime = (timestamp: string) => {
    const date = new Date(timestamp);
    return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
  };

  const renderMentions = (content: string, mentions?: string[]) => {
    if (!mentions || mentions.length === 0) {
      return content;
    }

    let processedContent = content;
    mentions.forEach(mention => {
      const regex = new RegExp(`@${mention}`, 'g');
      processedContent = processedContent.replace(
        regex,
        `<span class="message__mention">@${mention}</span>`
      );
    });

    return { __html: processedContent };
  };

  const handleReply = () => {
    if (onReply) {
      onReply(message.id);
    }
  };

  const handleEdit = () => {
    if (onEdit) {
      onEdit(message.id);
    }
  };

  const handleDelete = () => {
    if (onDelete) {
      onDelete(message.id);
    }
  };

  const handleReaction = (reaction: string) => {
    if (onReact) {
      onReact(message.id, reaction);
    }
  };

  return (
    <div 
      className={messageClass}
      onMouseEnter={() => setIsHovered(true)}
      onMouseLeave={() => setIsHovered(false)}
    >
      <Avatar 
        name={message.authorName}
        type={message.authorType}
        size="md"
        online={true}
      />
      
      <div className="message__content">
        <div className="message__header">
          <span className="message__author">{message.authorName}</span>
          <Badge variant={message.authorType} size="sm">
            {message.authorType.toUpperCase()}
          </Badge>
          <span className="message__time">{formatTime(message.timestamp)}</span>
          {message.edited && (
            <span className="message__edited">(edited)</span>
          )}
        </div>
        
        <div className="message__body">
          {message.mentions && message.mentions.length > 0 ? (
            <div 
              dangerouslySetInnerHTML={renderMentions(message.content, message.mentions)}
            />
          ) : (
            <p>{message.content}</p>
          )}
        </div>

        {message.attachments && message.attachments.length > 0 && (
          <div className="message__attachments">
            {message.attachments.map(attachment => (
              <div key={attachment.id} className="message__attachment">
                {attachment.type === 'image' ? (
                  <img 
                    src={attachment.url} 
                    alt={attachment.name}
                    className="message__attachment-image"
                  />
                ) : (
                  <div className="message__attachment-file">
                    <span className="message__attachment-icon">📎</span>
                    <span className="message__attachment-name">{attachment.name}</span>
                    {attachment.size && (
                      <span className="message__attachment-size">
                        {(attachment.size / 1024).toFixed(1)}KB
                      </span>
                    )}
                  </div>
                )}
              </div>
            ))}
          </div>
        )}

        {(isHovered && showActions) && (
          <div className="message__actions">
            {onReply && (
              <Button variant="icon" size="sm" onClick={handleReply} ariaLabel="Reply">
                💬
              </Button>
            )}
            {isOwn && onEdit && (
              <Button variant="icon" size="sm" onClick={handleEdit} ariaLabel="Edit">
                ✏️
              </Button>
            )}
            {isOwn && onDelete && (
              <Button variant="icon" size="sm" onClick={handleDelete} ariaLabel="Delete">
                🗑️
              </Button>
            )}
            <Button variant="icon" size="sm" onClick={() => handleReaction('👍')} ariaLabel="React">
              👍
            </Button>
          </div>
        )}
      </div>
    </div>
  );
}