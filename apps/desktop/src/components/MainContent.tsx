import React, { useState, useRef, useEffect } from 'react';
import { Message, Button, Input } from './ui';
import type { MessageData } from './ui';
import { NoMessagesState, LoadingState } from './EmptyStates';
import { MentionAutocomplete, useMentionAutocomplete } from './ui/MentionAutocomplete';
import { useWebSocket } from '../hooks/useWebSocket';
import '../styles/main-content.css';

interface MainContentProps {
  currentChannel?: string;
  webSocketHook?: ReturnType<typeof useWebSocket>;
}

function MainContent({ 
  currentChannel = 'general', 
  webSocketHook 
}: MainContentProps) {
  const [message, setMessage] = useState('');
  const [isSending, setIsSending] = useState(false);
  const messageInputRef = useRef<HTMLInputElement>(null);
  const messagesEndRef = useRef<HTMLDivElement>(null);
  
  // Use provided WebSocket hook or create a new one
  const wsHook = webSocketHook || useWebSocket();
  const {
    connected,
    connecting,
    messages,
    channels,
    users,
    activeChannel,
    sendMessage,
    startTyping,
    stopTyping,
    typingUsers,
    getChannelById
  } = wsHook;

  // Mention autocomplete
  const {
    isOpen: mentionOpen,
    query: mentionQuery,
    position: mentionPosition,
    checkForMention,
    insertMention,
    closeMention
  } = useMentionAutocomplete(messageInputRef as React.RefObject<HTMLInputElement>);

  // Get current channel data
  const channelData = getChannelById(activeChannel || currentChannel);
  const channelMessages = messages[activeChannel || currentChannel] || [];
  const channelTypingUsers = typingUsers[activeChannel || currentChannel] || [];

  // Auto-scroll to bottom when new messages arrive
  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [channelMessages]);

  // Handle typing indicators
  useEffect(() => {
    let typingTimer: NodeJS.Timeout;
    
    const handleInputChange = () => {
      if (message.trim() && connected) {
        startTyping(activeChannel || currentChannel);
        
        // Stop typing after 2 seconds of inactivity
        clearTimeout(typingTimer);
        typingTimer = setTimeout(() => {
          stopTyping(activeChannel || currentChannel);
        }, 2000);
      }
    };

    if (message.trim()) {
      handleInputChange();
    }
    
    return () => {
      clearTimeout(typingTimer);
      if (connected && message.trim()) {
        stopTyping(activeChannel || currentChannel);
      }
    };
  }, [message, connected, activeChannel, currentChannel, startTyping, stopTyping]);

  const handleSendMessage = async () => {
    if (!message.trim() || !connected || isSending) return;

    setIsSending(true);
    
    try {
      // Extract mentions from message
      const mentionRegex = /@(\w+(?:\s+\w+)*)/g;
      const mentions: Array<{id: string; name: string; type: 'agent' | 'human'}> = [];
      let match;
      
      while ((match = mentionRegex.exec(message)) !== null) {
        const mentionName = match[1];
        const user = users.find(u => u.name.toLowerCase().includes(mentionName.toLowerCase()));
        if (user) {
          mentions.push({
            id: user.id,
            name: user.name,
            type: user.type
          });
        }
      }

      sendMessage(activeChannel || currentChannel, message, mentions);
      setMessage('');
      if (connected) {
        stopTyping(activeChannel || currentChannel);
      }
    } catch (error) {
      console.error('Failed to send message:', error);
    } finally {
      setIsSending(false);
    }
  };

  const handleKeyPress = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey && !mentionOpen) {
      e.preventDefault();
      handleSendMessage();
    }
  };

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setMessage(e.target.value);
    checkForMention();
  };

  const handleMentionSelect = (mention: any, startPos: number, endPos: number) => {
    insertMention(mention, startPos, endPos);
  };

  if (connecting) {
    return (
      <div className="main-content">
        <LoadingState message="Connecting to workspace..." />
      </div>
    );
  }

  if (!connected) {
    return (
      <div className="main-content">
        <div className="main-content__offline">
          <div className="offline-icon">📡</div>
          <h2>Connection Lost</h2>
          <p>Unable to connect to Agent United. Check your connection and try again.</p>
        </div>
      </div>
    );
  }

  return (
    <div className="main-content">
      {/* Channel header */}
      <div className="channel-header">
        <div className="channel-info">
          <span className="channel-icon">#</span>
          <h1 className="channel-name">{channelData?.name || currentChannel}</h1>
          {channelData?.topic && (
            <span className="channel-topic">{channelData.topic}</span>
          )}
        </div>
        <div className="channel-actions">
          <Button variant="icon" size="sm" ariaLabel="More options">
            ⋮
          </Button>
          <Button variant="icon" size="sm" ariaLabel="Search">
            🔍
          </Button>
        </div>
      </div>

      {/* Message area */}
      <div className="message-area">
        <div className="message-list">
          {channelMessages.length > 0 ? (
            <>
              {channelMessages.map(msg => (
                <Message
                  key={msg.id}
                  message={{
                    id: msg.id,
                    authorId: msg.author.id,
                    authorName: msg.author.name,
                    authorType: msg.author.type,
                    content: msg.content,
                    timestamp: msg.timestamp,
                    mentions: msg.mentions?.map(m => m.name) || [],
                    attachments: msg.attachments?.map(att => ({
                      ...att,
                      type: att.type as 'image' | 'file' // Cast to expected type
                    }))
                  }}
                  showActions={true}
                  onReply={(messageId) => console.log('Reply to:', messageId)}
                  onReact={(messageId, reaction) => console.log('React to:', messageId, reaction)}
                />
              ))}
              
              {/* Typing indicators */}
              {channelTypingUsers.length > 0 && (
                <div className="typing-indicator">
                  <div className="typing-users">
                    {channelTypingUsers.map(user => user.name).join(', ')} 
                    {channelTypingUsers.length === 1 ? ' is' : ' are'} typing...
                  </div>
                </div>
              )}
              
              <div ref={messagesEndRef} />
            </>
          ) : (
            <NoMessagesState channelName={channelData?.name || currentChannel} />
          )}
        </div>

        {/* Message composer */}
        <div className="message-composer">
          <div className="composer-input-area">
            <div className="input-wrapper">
              <Input
                ref={messageInputRef}
                placeholder={`Type a message in #${channelData?.name || currentChannel}...`}
                value={message}
                onChange={handleInputChange}
                onKeyDown={handleKeyPress}
                size="md"
                disabled={!connected}
                rightIcon={
                  <Button
                    variant="primary"
                    size="sm"
                    onClick={handleSendMessage}
                    disabled={!message.trim() || !connected || isSending}
                    loading={isSending}
                  >
                    {isSending ? 'Sending...' : 'Send'}
                  </Button>
                }
              />
              
              {/* Mention autocomplete */}
              <MentionAutocomplete
                users={users}
                onSelectMention={handleMentionSelect}
                onClose={closeMention}
                inputRef={messageInputRef as React.RefObject<HTMLInputElement | HTMLTextAreaElement>}
                isOpen={mentionOpen}
                query={mentionQuery}
                position={mentionPosition}
              />
            </div>
          </div>
          
          {message.length > 1900 && (
            <div className="character-count">
              {message.length}/2000
            </div>
          )}
        </div>
      </div>
    </div>
  );
}

export default MainContent;