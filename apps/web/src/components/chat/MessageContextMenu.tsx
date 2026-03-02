import { useState } from 'react';
import { MoreVertical, Edit3, Trash2, Copy, Link } from 'lucide-react';
import { cn } from '../../lib/utils';

interface MessageContextMenuProps {
  messageId: string;
  isOwnMessage: boolean;
  messageText: string;
  onEdit?: () => void;
  onDelete?: () => void;
  onCopy?: () => void;
  className?: string;
}

export function MessageContextMenu({ 
  messageId, 
  isOwnMessage, 
  messageText, 
  onEdit, 
  onDelete, 
  onCopy,
  className 
}: MessageContextMenuProps) {
  const [isOpen, setIsOpen] = useState(false);

  const handleCopyText = () => {
    navigator.clipboard.writeText(messageText);
    onCopy?.();
    setIsOpen(false);
  };

  const handleCopyLink = () => {
    // TODO: Implement message link copying when message permalinks are supported
    navigator.clipboard.writeText(`#message-${messageId}`);
    setIsOpen(false);
  };

  const handleEdit = () => {
    onEdit?.();
    setIsOpen(false);
  };

  const handleDelete = () => {
    onDelete?.();
    setIsOpen(false);
  };

  return (
    <div className={cn("relative", className)}>
      <button
        onClick={() => setIsOpen(!isOpen)}
        className="p-1.5 rounded-md text-muted-foreground hover:text-foreground hover:bg-muted transition-colors opacity-0 group-hover:opacity-100"
        aria-label="Message options"
      >
        <MoreVertical className="h-4 w-4" />
      </button>

      {isOpen && (
        <>
          {/* Overlay to close menu when clicking outside */}
          <div 
            className="fixed inset-0 z-10" 
            onClick={() => setIsOpen(false)}
          />
          
          {/* Menu */}
          <div className="absolute right-0 top-8 z-20 w-48 bg-card border border-border rounded-md shadow-lg py-1">
            {isOwnMessage && onEdit && (
              <button
                onClick={handleEdit}
                className="flex items-center gap-2 w-full px-3 py-2 text-sm text-foreground hover:bg-muted transition-colors"
              >
                <Edit3 className="h-4 w-4" />
                Edit Message
              </button>
            )}
            
            <button
              onClick={handleCopyText}
              className="flex items-center gap-2 w-full px-3 py-2 text-sm text-foreground hover:bg-muted transition-colors"
            >
              <Copy className="h-4 w-4" />
              Copy Text
            </button>
            
            <button
              onClick={handleCopyLink}
              className="flex items-center gap-2 w-full px-3 py-2 text-sm text-foreground hover:bg-muted transition-colors"
            >
              <Link className="h-4 w-4" />
              Copy Message Link
            </button>
            
            {isOwnMessage && onDelete && (
              <>
                <hr className="my-1 border-border" />
                <button
                  onClick={handleDelete}
                  className="flex items-center gap-2 w-full px-3 py-2 text-sm text-destructive hover:bg-destructive/10 transition-colors"
                >
                  <Trash2 className="h-4 w-4" />
                  Delete Message
                </button>
              </>
            )}
          </div>
        </>
      )}
    </div>
  );
}