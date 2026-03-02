import { useState } from 'react';
import { Avatar } from "../ui/Avatar"
import { MemberBadge } from "../ui/MemberBadge"
import { MessageContextMenu } from "./MessageContextMenu"
import { DeleteMessageModal } from "./DeleteMessageModal"
import { chatApi } from "../../services/chatApi"

interface MessageItemProps {
  id: string
  channelId: string
  author: {
    name: string
    type: "human" | "agent"
  }
  content: string
  timestamp: string
  editedAt?: string
  isOwnMessage: boolean
  attachment?: {
    url: string
    name: string
  }
  onMessageUpdated?: (messageId: string, newContent: string) => void
  onMessageDeleted?: (messageId: string) => void
}

export function MessageItem({ 
  id, 
  channelId, 
  author, 
  content, 
  timestamp, 
  editedAt,
  isOwnMessage,
  attachment,
  onMessageUpdated,
  onMessageDeleted
}: MessageItemProps) {
  const [isEditing, setIsEditing] = useState(false);
  const [editContent, setEditContent] = useState(content);
  const [isLoading, setIsLoading] = useState(false);
  const [showDeleteModal, setShowDeleteModal] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleStartEdit = () => {
    setIsEditing(true);
    setEditContent(content);
    setError(null);
  };

  const handleCancelEdit = () => {
    setIsEditing(false);
    setEditContent(content);
    setError(null);
  };

  const handleSaveEdit = async () => {
    if (!editContent.trim() || editContent.trim() === content) {
      handleCancelEdit();
      return;
    }

    try {
      setIsLoading(true);
      setError(null);

      await chatApi.editMessage(channelId, id, editContent.trim());
      
      // Update local state
      onMessageUpdated?.(id, editContent.trim());
      setIsEditing(false);
    } catch (error) {
      console.error('Failed to edit message:', error);
      setError(error instanceof Error ? error.message : 'Failed to edit message');
    } finally {
      setIsLoading(false);
    }
  };

  const handleKeyDown = (e: React.KeyboardEvent<HTMLTextAreaElement>) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      handleSaveEdit();
    } else if (e.key === 'Escape') {
      handleCancelEdit();
    }
  };

  const handleDelete = async () => {
    try {
      setIsLoading(true);
      await chatApi.deleteMessage(channelId, id);
      onMessageDeleted?.(id);
      setShowDeleteModal(false);
    } catch (error) {
      console.error('Failed to delete message:', error);
      // Keep modal open on error so user can retry
    } finally {
      setIsLoading(false);
    }
  };

  const handleCopy = () => {
    // Visual feedback for copy action could be added here
    console.log('Message copied to clipboard');
  };

  const formatTimestamp = (timestamp: string) => {
    const date = new Date(timestamp);
    const now = new Date();
    const diffDays = Math.floor((now.getTime() - date.getTime()) / (1000 * 60 * 60 * 24));
    
    if (diffDays === 0) {
      return date.toLocaleTimeString([], { hour: 'numeric', minute: '2-digit' });
    } else if (diffDays === 1) {
      return `Yesterday ${date.toLocaleTimeString([], { hour: 'numeric', minute: '2-digit' })}`;
    } else {
      return date.toLocaleDateString([], { month: 'short', day: 'numeric', hour: 'numeric', minute: '2-digit' });
    }
  };

  const formatEditedTimestamp = (editedAt: string) => {
    const date = new Date(editedAt);
    return date.toLocaleString([], { 
      month: 'short', 
      day: 'numeric', 
      hour: 'numeric', 
      minute: '2-digit' 
    });
  };

  return (
    <>
      <div className="group flex gap-3 rounded-lg px-5 py-3 transition-colors hover:bg-muted/50">
        <Avatar name={author.name} type={author.type} />
        <div className="flex min-w-0 flex-1 flex-col">
          <div className="flex items-center gap-2">
            <span className="text-sm font-semibold text-foreground">{author.name}</span>
            <MemberBadge type={author.type} />
            <span className="text-xs text-muted-foreground">{formatTimestamp(timestamp)}</span>
            {editedAt && (
              <span 
                className="text-xs text-muted-foreground/70"
                title={`Edited at ${formatEditedTimestamp(editedAt)}`}
              >
                (edited)
              </span>
            )}
          </div>
          
          {isEditing ? (
            // Edit mode
            <div className="mt-1">
              <textarea
                value={editContent}
                onChange={(e) => setEditContent(e.target.value)}
                onKeyDown={handleKeyDown}
                className="w-full p-2 text-sm bg-background border border-input rounded-md resize-none min-h-[60px] focus:outline-none focus:ring-2 focus:ring-primary focus:border-transparent"
                placeholder="Edit your message..."
                disabled={isLoading}
                autoFocus
              />
              {error && (
                <div className="mt-1 text-xs text-destructive">{error}</div>
              )}
              <div className="mt-2 flex items-center gap-2 text-xs text-muted-foreground">
                <span>Enter to save • Esc to cancel</span>
                <div className="flex gap-1 ml-auto">
                  <button
                    onClick={handleCancelEdit}
                    disabled={isLoading}
                    className="px-2 py-1 text-xs hover:text-foreground transition-colors disabled:opacity-50"
                  >
                    Cancel
                  </button>
                  <button
                    onClick={handleSaveEdit}
                    disabled={isLoading || !editContent.trim()}
                    className="px-2 py-1 text-xs bg-primary text-primary-foreground rounded hover:bg-primary/90 transition-colors disabled:opacity-50"
                  >
                    {isLoading ? 'Saving...' : 'Save'}
                  </button>
                </div>
              </div>
            </div>
          ) : (
            // Display mode
            <div className="mt-1 text-sm leading-relaxed text-foreground/85">
              {content}
            </div>
          )}

          {attachment && (
            <div className="mt-3 overflow-hidden rounded-lg border border-border max-w-md">
              <div className="bg-muted/50 px-3 py-2 flex items-center gap-2">
                <span className="text-xs text-muted-foreground truncate">{attachment.name}</span>
                <a
                  href={attachment.url}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="text-xs text-primary hover:underline"
                >
                  Download
                </a>
              </div>
            </div>
          )}
        </div>

        {/* Context menu - only show when not editing */}
        {!isEditing && (
          <MessageContextMenu
            messageId={id}
            isOwnMessage={isOwnMessage}
            messageText={content}
            onEdit={handleStartEdit}
            onDelete={() => setShowDeleteModal(true)}
            onCopy={handleCopy}
            className="ml-auto"
          />
        )}
      </div>

      {/* Delete confirmation modal */}
      <DeleteMessageModal
        isOpen={showDeleteModal}
        onClose={() => setShowDeleteModal(false)}
        onConfirm={handleDelete}
        isDeleting={isLoading}
      />
    </>
  );
}