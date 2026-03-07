import { useState, useEffect, useCallback } from 'react';
import { ChatSidebar } from '../components/chat/ChatSidebar';
import { ChatHeader } from '../components/chat/ChatHeader';
import { MessageList } from '../components/chat/MessageList';
import { MessageInput } from '../components/chat/MessageInput';
import { CreateChannelModal } from '../components/chat/CreateChannelModal';
import { NewDMModal } from '../components/chat/NewDMModal';
import { MemberListPanel } from '../components/chat/MemberListPanel';
import { SearchResultsPanel } from '../components/chat/SearchResultsPanel';
import { ErrorBoundary } from '../components/ui/ErrorBoundary';
import { useWebSocket } from '../hooks/useWebSocket';
import { chatApi } from '../services/chatApi';
import { sendMessageWithAttachment } from '../services/api';
import type { Channel } from '../types/chat';

interface DirectMessage {
  id: string;
  name: string;
  type: 'agent' | 'human';
  online: boolean;
  unread?: number;
}

function getDisplayName(name: string): string {
  if (!name.includes('@')) return name;
  const localPart = name.split('@')[0] ?? name;
  return localPart.replace(/[._-]/g, ' ').replace(/\b\w/g, (c) => c.toUpperCase());
}

export function ChatPage() {
  const [channels, setChannels] = useState<Channel[]>([]);
  const [directMessages, setDirectMessages] = useState<DirectMessage[]>([]);
  const [selectedChannelId, setSelectedChannelId] = useState<string>('');
  const [selectedDMId, setSelectedDMId] = useState<string>('');
  const [channelsLoading, setChannelsLoading] = useState(true);
  const [channelsError, setChannelsError] = useState<string | null>(null);
  const [showCreateChannel, setShowCreateChannel] = useState(false);
  const [showNewDM, setShowNewDM] = useState(false);
  const [showMembersPanel, setShowMembersPanel] = useState(false);
  const [sidebarOpen, setSidebarOpen] = useState(false);
  const [searchQuery, setSearchQuery] = useState('');
  const [isSearching, setIsSearching] = useState(false);
  const [channelMembers, setChannelMembers] = useState<{ id: string; name: string; email: string; type: 'agent' | 'human'; online: boolean }[]>([]);
  
  // Determine if we're viewing a channel or DM
  const isViewingDM = !!selectedDMId;
  const activeConversationId = isViewingDM ? selectedDMId : selectedChannelId;
  
  const { isConnected, messages, sendMessage, error: wsError } = useWebSocket('ws://localhost:8080/ws', activeConversationId);

  // Load channels on component mount
  useEffect(() => {
    const loadChannels = async () => {
      try {
        setChannelsLoading(true);
        setChannelsError(null);
        
        const fetchedChannels = await chatApi.getChannels();
        setChannels(fetchedChannels);
        
        // Auto-select first channel if none selected
        if (fetchedChannels.length > 0 && !selectedChannelId && !selectedDMId) {
          setSelectedChannelId(fetchedChannels[0].id);
        }
      } catch (error) {
        console.error('Failed to load channels:', error);
        setChannelsError(error instanceof Error ? error.message : 'Failed to load channels');
      } finally {
        setChannelsLoading(false);
      }
    };

    loadChannels();
  }, [selectedChannelId, selectedDMId]);

  // Load DMs on component mount
  useEffect(() => {
    const loadDMs = async () => {
      try {
        const dms = await chatApi.listDMs();
        
        // Convert channels to DirectMessage format
        const directMessageList: DirectMessage[] = dms.map(dm => {
          // TODO: Parse DM name and type from API response when available
          // For now, assume DM names contain user info
          const isAgent = dm.name.includes('agent') || dm.name.includes('Agent');
          return {
            id: dm.id,
            name: dm.name,
            type: isAgent ? 'agent' : 'human',
            online: Math.random() > 0.3, // TODO: Replace with real online status
            unread: dm.unread
          };
        });
        
        setDirectMessages(directMessageList);
      } catch (error) {
        console.error('Failed to load DMs:', error);
        // Don't set error for DMs, just log it
      }
    };

    loadDMs();
  }, []);

  // Load members for the selected channel
  useEffect(() => {
    if (!selectedChannelId || isViewingDM) return;

    const loadChannelMembers = async () => {
      try {
        const members = await chatApi.getMembers(selectedChannelId);
        
        // Convert API response to Member format
        const memberList = members.map(member => {
          const isAgent = member.email.includes('@agentunited.local');
          return {
            id: member.id,
            name: isAgent 
              ? member.email.split('@')[0] 
              : member.email.split('@')[0] || member.email,
            email: member.email,
            type: isAgent ? 'agent' as const : 'human' as const,
            online: Math.random() > 0.3 // TODO: Replace with real online status
          };
        });

        setChannelMembers(memberList);
      } catch (error) {
        console.error('Failed to load channel members:', error);
        setChannelMembers([]);
      }
    };

    loadChannelMembers();
  }, [selectedChannelId, isViewingDM]);

  const selectedChannel = channels.find(ch => ch.id === selectedChannelId) || null;
  const selectedDM = directMessages.find(dm => dm.id === selectedDMId) || null;

  const handleSendMessage = async (text: string, mentions?: { userId: string; display: string }[], attachment?: File) => {
    if (!activeConversationId) return;
    
    try {
      if (attachment) {
        // Use API directly for file uploads
        await sendMessageWithAttachment(activeConversationId, text, attachment);
      } else {
        // Use WebSocket for text-only messages  
        sendMessage(text);
      }
      
      if (mentions && mentions.length > 0) {
        console.log('Mentions in message:', mentions);
      }
    } catch (error) {
      console.error('Failed to send message:', error);
      // Let MessageInput handle the error display
      throw error;
    }
  };

  const handleSelectChannel = (channelId: string) => {
    setSelectedChannelId(channelId);
    setSelectedDMId(''); // Clear DM selection
    setSidebarOpen(false);
  };

  const handleDMSelect = (dmId: string) => {
    setSelectedDMId(dmId);
    setSelectedChannelId(''); // Clear channel selection
    setSidebarOpen(false);
  };

  const handleCreateChannel = useCallback(async (name: string, description: string) => {
    const newChannel = await chatApi.createChannel(name, description);
    setChannels(prev => [...prev, newChannel]);
    setSelectedChannelId(newChannel.id);
    setSelectedDMId(''); // Clear DM selection
  }, []);

  const handleSearch = useCallback((query: string) => {
    setSearchQuery(query);
    setIsSearching(true);
  }, []);

  const handleCloseSearch = useCallback(() => {
    setSearchQuery('');
    setIsSearching(false);
  }, []);

  const handleSearchResultClick = useCallback((channelId: string, messageId: string) => {
    // Navigate to the channel containing the message
    setSelectedChannelId(channelId);
    setSelectedDMId(''); // Clear DM selection
    setIsSearching(false); // Close search
    setSearchQuery('');
    
    // TODO: Scroll to specific message when message list supports it
    console.log(`Navigate to message ${messageId} in channel ${channelId}`);
  }, []);

  const handleChannelUpdate = useCallback((updatedChannel: Channel) => {
    setChannels(prev => prev.map(ch => ch.id === updatedChannel.id ? updatedChannel : ch));
  }, []);

  const handleChannelDelete = useCallback((channelId: string) => {
    setChannels(prev => prev.filter(ch => ch.id !== channelId));
    if (selectedChannelId === channelId && channels.length > 1) {
      const remaining = channels.filter(ch => ch.id !== channelId);
      setSelectedChannelId(remaining[0]?.id || '');
    }
  }, [selectedChannelId, channels]);

  const handleDMCreated = useCallback(async (dmId: string) => {
    try {
      // Reload DMs to get the new one
      const dms = await chatApi.listDMs();
      const directMessageList: DirectMessage[] = dms.map(dm => {
        const isAgent = dm.name.includes('agent') || dm.name.includes('Agent');
        return {
          id: dm.id,
          name: dm.name,
          type: isAgent ? 'agent' : 'human',
          online: Math.random() > 0.3,
          unread: dm.unread
        };
      });
      
      setDirectMessages(directMessageList);
      
      // Select the new DM
      setSelectedDMId(dmId);
      setSelectedChannelId(''); // Clear channel selection
    } catch (error) {
      console.error('Failed to refresh DMs:', error);
      // Still select the DM even if refresh failed
      setSelectedDMId(dmId);
      setSelectedChannelId('');
    }
  }, []);

  const handleToggleMembers = useCallback(() => {
    setShowMembersPanel(prev => !prev);
  }, []);

  // Keyboard Shortcuts
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      // CMD/CTRL + K to open search
      if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
        e.preventDefault();
        setIsSearching(true);
        // Dispatch custom event to focus the sidebar input
        window.dispatchEvent(new CustomEvent('focus-chat-search'));
      }
      
      // ALT + N to create new channel
      if (e.altKey && e.key === 'n') {
        e.preventDefault();
        setShowCreateChannel(true);
      }

      // ALT + D to start new DM
      if (e.altKey && e.key === 'd') {
        e.preventDefault();
        setShowNewDM(true);
      }

      // ESC to close modals/search/sidebar
      if (e.key === 'Escape') {
        if (sidebarOpen) setSidebarOpen(false);
        if (isSearching) handleCloseSearch();
        if (showCreateChannel) setShowCreateChannel(false);
        if (showNewDM) setShowNewDM(false);
        if (showMembersPanel) setShowMembersPanel(false);
      }

      // '/' to focus chat input
      if (e.key === '/' && !isSearching && !showCreateChannel && !showNewDM) {
        // Only if we aren't already typing in an input
        const activeElement = document.activeElement;
        const isTyping = activeElement instanceof HTMLInputElement || 
                         activeElement instanceof HTMLTextAreaElement;
        
        if (!isTyping) {
          e.preventDefault();
          const chatInput = document.querySelector('.chat-message-input textarea') as HTMLTextAreaElement;
          chatInput?.focus();
        }
      }
    };

    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [isSearching, showCreateChannel, showNewDM, showMembersPanel, sidebarOpen, handleCloseSearch]);

  const handleMessageUpdated = useCallback((messageId: string, newContent: string) => {
    // TODO: Update the message in local state when WebSocket integration is enhanced
    // For now, we rely on the API call in MessageItem
    console.log(`Message ${messageId} updated to: ${newContent}`);
  }, []);

  const handleMessageDeleted = useCallback((messageId: string) => {
    // TODO: Remove the message from local state when WebSocket integration is enhanced
    // For now, we rely on the API call in MessageItem
    console.log(`Message ${messageId} deleted`);
  }, []);

  // Show loading state while channels are loading
  if (channelsLoading) {
    return (
      <div className="h-screen flex items-center justify-center bg-background">
        <div className="text-center">
          <div className="text-lg font-medium text-foreground mb-2">Loading channels...</div>
          <div className="text-sm text-muted-foreground">Connecting to Agent United</div>
        </div>
      </div>
    );
  }

  // Show error state if channels failed to load
  if (channelsError) {
    return (
      <div className="h-screen flex items-center justify-center bg-background">
        <div className="text-center max-w-md mx-auto p-6">
          <div className="text-lg font-medium text-destructive mb-2">Failed to load channels</div>
          <div className="text-sm text-destructive-foreground mb-4">{channelsError}</div>
          <button 
            onClick={() => window.location.reload()} 
            className="px-4 py-2 bg-primary text-primary-foreground rounded hover:bg-primary/90"
          >
            Retry
          </button>
        </div>
      </div>
    );
  }

  // Show empty state if no channels
  if (channels.length === 0) {
    return (
      <div className="flex h-screen items-center justify-center bg-background px-4">
        <div className="w-full max-w-md rounded-2xl border border-emerald-500/25 bg-white/85 p-7 text-center shadow-[0_20px_60px_-45px_rgba(16,185,129,0.55)] backdrop-blur dark:bg-card/90">
          <div className="mx-auto mb-4 flex h-12 w-12 items-center justify-center rounded-full border border-emerald-500/45 shadow-[0_0_18px_rgba(16,185,129,0.35)]">
            <div className="h-2.5 w-2.5 rounded-full bg-emerald-500" />
          </div>
          <h2 className="text-lg font-semibold text-foreground">No channels yet</h2>
          <p className="mt-1 text-sm text-muted-foreground">
            Waiting for an agent to connect. Run <span className="font-mono text-foreground">./setup.sh</span> to get started.
          </p>
          <button
            onClick={() => setShowCreateChannel(true)}
            className="mt-5 rounded-xl bg-emerald-500 px-4 py-2 text-sm font-semibold text-white transition-colors hover:bg-emerald-600"
          >
            Create a channel
          </button>
        </div>

        <CreateChannelModal
          isOpen={showCreateChannel}
          onClose={() => setShowCreateChannel(false)}
          onSubmit={handleCreateChannel}
        />
      </div>
    );
  }

  // Determine the active conversation name and type
  const activeConversationName = isViewingDM 
    ? getDisplayName(selectedDM?.name || 'Unknown')
    : selectedChannel?.name || 'unknown';
  
  const activeConversationTopic = isViewingDM 
    ? `Direct message with ${selectedDM?.name}`
    : selectedChannel?.topic;

  return (
    <div className="h-screen flex overflow-hidden bg-background">
      {sidebarOpen && (
        <div
          className="fixed inset-0 z-30 bg-black/40 md:hidden"
          onClick={() => setSidebarOpen(false)}
          aria-hidden="true"
        />
      )}

      <ErrorBoundary
        title="Sidebar error"
        message="Navigation hit an error. Reload to recover the sidebar."
        className="w-64 shrink-0"
      >
        <ChatSidebar
          channels={channels}
          directMessages={directMessages}
          activeChannelId={activeConversationId}
          isOpen={sidebarOpen}
          onChannelSelect={handleSelectChannel}
          onDMSelect={handleDMSelect}
          onCreateChannel={() => setShowCreateChannel(true)}
          onNewDM={() => setShowNewDM(true)}
          onSearch={handleSearch}
          onChannelUpdate={handleChannelUpdate}
          onChannelDelete={handleChannelDelete}
        />
      </ErrorBoundary>

      <CreateChannelModal
        isOpen={showCreateChannel}
        onClose={() => setShowCreateChannel(false)}
        onSubmit={handleCreateChannel}
      />

      <NewDMModal
        isOpen={showNewDM}
        onClose={() => setShowNewDM(false)}
        onDMCreated={handleDMCreated}
      />

      <ErrorBoundary
        title="Chat panel error"
        message="The chat panel hit an error. Sidebar navigation should still work."
        className="flex-1"
      >
        <div className="flex-1 flex flex-col">
          {isSearching ? (
            // Show search results instead of normal chat
            <SearchResultsPanel
              query={searchQuery}
              channels={channels}
              onClose={handleCloseSearch}
              onResultClick={handleSearchResultClick}
            />
          ) : (
            // Show normal chat interface
            <>
              {!isConnected && (
                <div className={`sticky top-0 z-20 border-b px-4 py-2 text-sm ${wsError ? 'border-destructive/25 bg-destructive/10 text-destructive' : 'border-amber-400/25 bg-amber-500/10 text-amber-700 dark:text-amber-300'}`}>
                  <div className="flex items-center gap-2 font-medium">
                    <span className={`inline-block h-2 w-2 rounded-full ${wsError ? 'bg-destructive' : 'bg-amber-500 animate-pulse'}`} />
                    {wsError ? wsError : 'Reconnecting to live updates…'}
                  </div>
                </div>
              )}

              <ChatHeader
                channelName={activeConversationName}
                topic={activeConversationTopic}
                isDM={isViewingDM}
                onToggleMembers={handleToggleMembers}
                showMembersPanel={showMembersPanel}
                onToggleSidebar={() => setSidebarOpen((prev) => !prev)}
                sidebarOpen={sidebarOpen}
              />

              <MessageList
                messages={messages}
                channelId={activeConversationId}
                channelName={activeConversationName}
                isDM={isViewingDM}
                onMessageUpdated={handleMessageUpdated}
                onMessageDeleted={handleMessageDeleted}
              />

              <MessageInput
                onSend={handleSendMessage}
                members={isViewingDM ? [] : channelMembers}
                placeholder={isViewingDM
                  ? `Message ${selectedDM?.name || 'user'}`
                  : `Message #${selectedChannel?.name || 'general'}`
                }
                className="chat-message-input"
              />
            </>
          )}
        </div>
      </ErrorBoundary>

      {/* Member List Panel - only show for channels, not DMs, and not during search */}
      {!isViewingDM && !isSearching && showMembersPanel && selectedChannel && (
        <MemberListPanel
          isOpen={showMembersPanel}
          onClose={() => setShowMembersPanel(false)}
          channelId={selectedChannel.id}
          channelName={selectedChannel.name}
        />
      )}
    </div>
  );
}