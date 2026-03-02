import { useState, useRef, useCallback } from "react"
import { Plus, Send, Smile, Paperclip, AtSign, X } from "lucide-react"
import { cn } from "../../lib/utils"
import { MentionAutocomplete } from "./MentionAutocomplete"
import { useMentions } from "../../hooks/useMentions"

interface Member {
  id: string;
  name: string;
  email: string;
  type: 'agent' | 'human';
  online?: boolean;
}

interface MessageInputProps {
  onSend: (message: string, mentions?: { userId: string; display: string }[], attachment?: File) => void;
  members?: Member[];
  placeholder?: string;
}

export function MessageInput({ 
  onSend, 
  members = [], 
  placeholder = "Message #general" 
}: MessageInputProps) {
  const [value, setValue] = useState("")
  const [cursorPosition, setCursorPosition] = useState(0);
  const [currentMentions, setCurrentMentions] = useState<{ userId: string; display: string }[]>([]);
  const [selectedFile, setSelectedFile] = useState<File | null>(null);
  const [isUploading, setIsUploading] = useState(false);
  const [uploadError, setUploadError] = useState<string | null>(null);
  const textareaRef = useRef<HTMLTextAreaElement>(null)
  const fileInputRef = useRef<HTMLInputElement>(null);

  const {
    isShowingMentions,
    mentionQuery,
    mentionPosition,
    selectedMemberIndex,
    inputRef,
    handleInputChange,
    handleKeyDown,
    selectMention,
    filteredMembers
  } = useMentions({ 
    members,
    onMentionSelected: (mention) => {
      setCurrentMentions(prev => [...prev, { userId: mention.userId, display: mention.display }]);
    }
  });

  // Sync refs
  if (textareaRef.current && inputRef.current !== textareaRef.current) {
    (inputRef as any).current = textareaRef.current;
  }

  const handleValueChange = useCallback((newValue: string) => {
    setValue(newValue);
    if (textareaRef.current) {
      const cursor = textareaRef.current.selectionStart || 0;
      setCursorPosition(cursor);
      handleInputChange(newValue, cursor);
    }
  }, [handleInputChange]);

  const handleFileSelect = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    setUploadError(null);
    
    if (file) {
      // Validate file size (max 10MB)
      const maxSizeBytes = 10 * 1024 * 1024; // 10MB
      if (file.size > maxSizeBytes) {
        setUploadError('File size must be less than 10MB');
        return;
      }
      
      setSelectedFile(file);
    }
    
    // Reset file input
    if (fileInputRef.current) {
      fileInputRef.current.value = '';
    }
  };

  const handleRemoveFile = () => {
    setSelectedFile(null);
    setUploadError(null);
  };

  const handleSubmit = async () => {
    if (!value.trim() && !selectedFile) return;
    
    try {
      setIsUploading(true);
      setUploadError(null);
      
      await onSend(value, currentMentions, selectedFile || undefined);
      
      // Reset form
      setValue("");
      setCurrentMentions([]);
      setSelectedFile(null);
      if (textareaRef.current) {
        textareaRef.current.style.height = "auto"
      }
    } catch (error) {
      console.error('Failed to send message:', error);
      setUploadError(error instanceof Error ? error.message : 'Failed to send message');
    } finally {
      setIsUploading(false);
    }
  }

  const handleTextareaKeyDown = (e: React.KeyboardEvent) => {
    // Let mentions handler try first
    const mentionHandled = handleKeyDown(
      e, 
      value, 
      cursorPosition, 
      handleValueChange,
      (pos) => {
        if (textareaRef.current) {
          textareaRef.current.setSelectionRange(pos, pos);
          setCursorPosition(pos);
        }
      }
    );

    if (mentionHandled) return;

    // Handle regular send
    if (e.key === "Enter" && !e.shiftKey) {
      e.preventDefault()
      handleSubmit()
    }
  }

  const handleMentionSelect = (member: Member) => {
    if (textareaRef.current) {
      const { newValue, newCursor } = selectMention(member, value, cursorPosition);
      setValue(newValue);
      setCursorPosition(newCursor);
      textareaRef.current.setSelectionRange(newCursor, newCursor);
      textareaRef.current.focus();
    }
  };

  const handleCursorChange = () => {
    if (textareaRef.current) {
      const cursor = textareaRef.current.selectionStart || 0;
      setCursorPosition(cursor);
      handleInputChange(value, cursor);
    }
  };

  const insertMention = () => {
    if (textareaRef.current) {
      const cursor = textareaRef.current.selectionStart || 0;
      const beforeCursor = value.substring(0, cursor);
      const afterCursor = value.substring(cursor);
      const newValue = beforeCursor + "@" + afterCursor;
      const newCursor = cursor + 1;
      
      setValue(newValue);
      setCursorPosition(newCursor);
      setTimeout(() => {
        if (textareaRef.current) {
          textareaRef.current.setSelectionRange(newCursor, newCursor);
          textareaRef.current.focus();
        }
      }, 0);
      
      handleInputChange(newValue, newCursor);
    }
  };

  const formatFileSize = (bytes: number): string => {
    if (bytes === 0) return '0 Bytes';
    const k = 1024;
    const sizes = ['Bytes', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  };

  const canSend = (value.trim() || selectedFile) && !isUploading;

  return (
    <>
      <div className="border-t border-border bg-card px-5 py-3">
        {/* File attachment preview */}
        {selectedFile && (
          <div className="mb-3 flex items-center gap-2 rounded-lg border border-border bg-muted/30 px-3 py-2">
            <Paperclip className="h-4 w-4 text-muted-foreground" />
            <div className="flex-1 min-w-0">
              <span className="text-sm text-foreground truncate block">{selectedFile.name}</span>
              <span className="text-xs text-muted-foreground">{formatFileSize(selectedFile.size)}</span>
            </div>
            <button
              onClick={handleRemoveFile}
              className="shrink-0 rounded-md p-1 text-muted-foreground hover:bg-muted hover:text-foreground transition-colors"
              aria-label="Remove attachment"
            >
              <X className="h-4 w-4" />
            </button>
          </div>
        )}

        {/* Error message */}
        {uploadError && (
          <div className="mb-3 rounded-md border border-destructive/20 bg-destructive/10 px-3 py-2">
            <span className="text-sm text-destructive">{uploadError}</span>
          </div>
        )}

        <div className="flex items-end gap-2 rounded-xl border border-input bg-background px-3 py-2 focus-within:border-primary/40 focus-within:ring-2 focus-within:ring-primary/10 transition-all">
          <button
            className="mb-0.5 shrink-0 rounded-md p-1.5 text-muted-foreground transition-colors hover:bg-muted hover:text-foreground"
            aria-label="Attach file"
          >
            <Plus className="h-4.5 w-4.5" />
          </button>

          <div className="relative flex-1">
            <textarea
              ref={textareaRef}
              value={value}
              onChange={(e) => handleValueChange(e.target.value)}
              onKeyDown={handleTextareaKeyDown}
              onSelect={handleCursorChange}
              onClick={handleCursorChange}
              placeholder={placeholder}
              rows={1}
              disabled={isUploading}
              className="max-h-32 min-h-[24px] w-full resize-none bg-transparent text-sm leading-relaxed text-foreground outline-none placeholder:text-muted-foreground disabled:opacity-60"
              onInput={(e) => {
                const target = e.target as HTMLTextAreaElement
                target.style.height = "auto"
                target.style.height = target.scrollHeight + "px"
              }}
            />
          </div>

          <div className="mb-0.5 flex shrink-0 items-center gap-0.5">
            <button
              onClick={insertMention}
              disabled={isUploading}
              className="rounded-md p-1.5 text-muted-foreground transition-colors hover:bg-muted hover:text-foreground disabled:opacity-60"
              aria-label="Mention someone"
            >
              <AtSign className="h-4 w-4" />
            </button>
            
            <input
              ref={fileInputRef}
              type="file"
              onChange={handleFileSelect}
              className="hidden"
              accept="*/*"
              disabled={isUploading}
            />
            <button
              onClick={() => fileInputRef.current?.click()}
              disabled={isUploading}
              className="rounded-md p-1.5 text-muted-foreground transition-colors hover:bg-muted hover:text-foreground disabled:opacity-60"
              aria-label="Add attachment"
            >
              <Paperclip className="h-4 w-4" />
            </button>
            
            <button
              className="rounded-md p-1.5 text-muted-foreground transition-colors hover:bg-muted hover:text-foreground disabled:opacity-60"
              aria-label="Add emoji"
              disabled={isUploading}
            >
              <Smile className="h-4 w-4" />
            </button>
            <button
              onClick={handleSubmit}
              disabled={!canSend}
              className={cn(
                "ml-1 rounded-lg p-2 transition-colors",
                canSend
                  ? "bg-primary text-primary-foreground hover:bg-primary/90"
                  : "bg-muted text-muted-foreground cursor-not-allowed"
              )}
              aria-label={isUploading ? "Uploading..." : "Send message"}
            >
              <Send className="h-4 w-4" />
            </button>
          </div>
        </div>
        <p className="mt-1.5 px-1 text-[11px] text-muted-foreground">
          <span className="font-medium">Shift + Enter</span> for a new line • <span className="font-medium">@</span> to mention • Max 10MB files
        </p>
      </div>

      {/* Mention Autocomplete */}
      <MentionAutocomplete
        isOpen={isShowingMentions}
        query={mentionQuery}
        members={filteredMembers}
        selectedIndex={selectedMemberIndex}
        onSelect={handleMentionSelect}
        onKeyDown={() => {}} // Handled in textarea
        position={mentionPosition}
      />
    </>
  )
}