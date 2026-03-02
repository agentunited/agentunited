import { useState, useEffect, useCallback } from 'react';
import { ChatSidebar } from '../components/chat/ChatSidebar';
import { ChatHeader } from '../components/chat/ChatHeader';
import { MessageList } from '../components/chat/MessageList';
import { MessageInput } from '../components/chat/MessageInput';
import { CreateChannelModal } from '../components/chat/CreateChannelModal';
import { NewDMModal } from '../components/chat/NewDMModal';
import { MemberListPanel } from '../components/chat/MemberListPanel';
import { SearchResultsPanel } from '../components/chat/SearchResultsPanel';
import { useWebSocket } from '../hooks/useWebSocket';
import { chatApi } from '../services/chatApi';
import type { Channel } from '../types/chat';

interface DirectMessage {
  id: string;
  name: string;
  type: 'agent' | 'human';
  online: boolean;
  unread?: number;
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
  const [searchQuery, setSearchQuery] = useState('');
  const [isSearching, setIsSearching] = useState(false);
  
  // Determine if we're viewing a channel or DM
  const isViewingDM = !!selectedDMId;
  const activeConversationId = isViewingDM ? selectedDMId : selectedChannelId;
  
  const { messages, sendMessage, error: wsError } = useWebSocket('ws://localhost:8080/ws', activeConversationId);

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

  const selectedChannel = channels.find(ch => ch.id === selectedChannelId) || null;
  const selectedDM = directMessages.find(dm => dm.id === selectedDMId) || null;

  const handleSendMessage = (text: string) => {
    sendMessage(text);
  };

  const handleSelectChannel = (channelId: string) => {
    setSelectedChannelId(channelId);
    setSelectedDMId(''); // Clear DM selection
  };

  const handleDMSelect = (dmId: string) => {
    setSelectedDMId(dmId);
    setSelectedChannelId(''); // Clear channel selection
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
      <div className="h-screen flex items-center justify-center bg-background">
        <div className="text-center max-w-md mx-auto p-6">
          <div className="text-lg font-medium text-foreground mb-2">No channels available</div>
          <div className="text-sm text-muted-foreground">Contact an agent to get access to channels</div>
        </div>
      </div>
    );
  }

  // Determine the active conversation name and type
  const activeConversationName = isViewingDM 
    ? selectedDM?.name || 'Unknown'
    : selectedChannel?.name || 'unknown';
  
  const activeConversationTopic = isViewingDM 
    ? `Direct message with ${selectedDM?.name}`
    : selectedChannel?.topic;

  return (
    <div className="h-screen flex overflow-hidden bg-background">
      <ChatSidebar
        channels={channels}
        directMessages={directMessages}
        activeChannelId={selectedChannelId}
        onChannelSelect={handleSelectChannel}
        onDMSelect={handleDMSelect}
        onCreateChannel={() => setShowCreateChannel(true)}
        onNewDM={() => setShowNewDM(true)}
        onSearch={handleSearch}
        onChannelUpdate={handleChannelUpdate}
        onChannelDelete={handleChannelDelete}
      />

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
            <ChatHeader
              channelName={activeConversationName}
              topic={activeConversationTopic}
              isDM={isViewingDM}
              onToggleMembers={handleToggleMembers}
              showMembersPanel={showMembersPanel}
            />

            <MessageList messages={messages} />

            {wsError && (
              <div className="px-4 py-2 bg-destructive/10 border-t border-destructive/20">
                <div className="text-sm text-destructive">{wsError}</div>
              </div>
            )}

            <MessageInput
              onSend={handleSendMessage}
              placeholder={isViewingDM 
                ? `Message ${selectedDM?.name || 'user'}`
                : `Message #${selectedChannel?.name || 'general'}`
              }
            />
          </>
        )}
      </div>

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