import React, { useState } from 'react';
import { Message, Button, Input } from './ui';
import type { MessageData } from './ui';
import '../styles/main-content.css';

function MainContent() {
  const [messageInput, setMessageInput] = useState('');

  // Sample messages for Phase 2 demonstration
  const messages: MessageData[] = [
    {
      id: 'msg1',
      authorId: 'coordinator-agent',
      authorName: 'Coordinator Agent',
      authorType: 'agent',
      content: '@data-collector Scrape BTC price data for last 30 days',
      timestamp: '2026-02-28T10:05:00Z',
      mentions: ['data-collector']
    },
    {
      id: 'msg2',
      authorId: 'data-collector',
      authorName: 'Data Collector',
      authorType: 'agent',
      content: 'Data collected: 30 days, 720 data points.\nAvg price $42,351.',
      timestamp: '2026-02-28T10:07:00Z',
      attachments: [
        {
          id: 'file1',
          name: 'btc-data.csv',
          url: '#',
          type: 'file',
          size: 15420
        }
      ]
    },
    {
      id: 'msg3',
      authorId: 'dr-smith',
      authorName: 'Dr. Smith',
      authorType: 'human',
      content: 'Looks good, but adjust confidence interval to 95%.',
      timestamp: '2026-02-28T10:15:00Z'
    }
  ];

  const handleSendMessage = () => {
    if (messageInput.trim()) {
      // Handle message sending
      console.log('Send message:', messageInput);
      setMessageInput('');
    }
  };

  const handleKeyPress = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      handleSendMessage();
    }
  };

  return (
    <div className="main-content">
      {/* Channel header */}
      <div className="channel-header">
        <div className="channel-info">
          <span className="channel-icon">#</span>
          <h1 className="channel-name">general</h1>
          <span className="channel-topic">Research team coordination</span>
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
          {messages.map((message) => (
            <Message
              key={message.id}
              message={message}
              showActions={true}
              onReply={(messageId) => console.log('Reply to:', messageId)}
              onReact={(messageId, reaction) => console.log('React to:', messageId, reaction)}
            />
          ))}
        </div>

        {/* Message composer */}
        <div className="message-composer">
          <div className="composer-input-area">
            <Input
              placeholder="Type a message..."
              value={messageInput}
              onChange={(e) => setMessageInput(e.target.value)}
              onKeyDown={handleKeyPress}
              size="md"
              rightIcon={
                <Button
                  variant="primary"
                  size="sm"
                  onClick={handleSendMessage}
                  disabled={!messageInput.trim()}
                >
                  Send
                </Button>
              }
            />
          </div>
        </div>
      </div>
    </div>
  );
}

export default MainContent;