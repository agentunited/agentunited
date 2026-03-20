import { useState, useEffect, useCallback, useMemo } from 'react';
import { useNavigate, useParams, useSearchParams } from 'react-router-dom';
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
import { getDisplayName } from '../lib/displayName';

interface DirectMessage {
  id: string;
  name: string;
  type: 'agent' | 'human';
  online: boolean;
  unread?: number;
}

export function ChatPage() {
  const navigate = useNavigate();
  const { dmId } = useParams<{ dmId?: string }>();
  const [searchParams, setSearchParams] = useSearchParams();
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
  const [channelCreateWarning, setChannelCreateWarning] = useState<string | null>(null);
  const [startupToast, setStartupToast] = useState<string | null>(null);
  const [searchQuery, setSearchQuery] = useState('');
  const [isSearching, setIsSearching] = useState(false);
  const [channelMembers, setChannelMembers] = useState<{ id: string; name: string; email: string; type: 'agent' | 'human'; online: boolean }[]>([]);
  const [userDirectory, setUserDirectory] = useState<Record<string, { display: string; type: 'agent' | 'human' }>>({});
  
  // Determine if we're viewing a channel or DM
  const isViewingDM = !!selectedDMId;
  const activeConversationId = isViewingDM ? selectedDMId : selectedChannelId;
  
  const { connectionStatus, messages, sendMessage } = useWebSocket('', activeConversationId);

  // Load channels on component mount
  useEffect(() => {
    const loadChannels = async () => {
      try {
        setChannelsLoading(true);
        setChannelsError(null);

        const fetchedChannels = await chatApi.getChannels();
        setChannels(fetchedChannels);

        // Auto-select first channel if nothing selected yet.
        if (fetchedChannels.length > 0 && !selectedChannelId && !selectedDMId && !dmId) {
          setSelectedChannelId(fetchedChannels[0].id);
        }
      } catch (error) {
        console.error('Failed to load channels:', error);
        setChannelsError(error instanceof Error ? error.message : 'Failed to load channels');
      } finally {
        setChannelsLoading(false);
      }
    };

    void loadChannels();
    // load once on mount; downstream selection changes should not refetch channel list
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  // Load DMs on component mount
  useEffect(() => {
    const loadDMs = async () => {
      try {
        const dms = await chatApi.listDMs();

        const directMessageList: DirectMessage[] = dms.map(dm => ({
          id: dm.id,
          name: dm.name,
          type: /agent/i.test(dm.name) ? 'agent' : 'human',
          online: false,
          unread: dm.unread
        }));

        setDirectMessages(directMessageList);

        // Auto-select first DM if user has no channels and no route target.
        if (!dmId && directMessageList.length > 0 && !selectedChannelId && !selectedDMId) {
          setSelectedDMId(directMessageList[0].id);
        }
      } catch (error) {
        console.error('Failed to load DMs:', error);
      }
    };

    void loadDMs();
    // load once on mount; selection changes should not refetch DM list
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  // Load users directory for sender display-name resolution
  useEffect(() => {
    const loadUsersDirectory = async () => {
      try {
        const users = await chatApi.getUsers();
        const next: Record<string, { display: string; type: 'agent' | 'human' }> = {};
        for (const user of users) {
          const display = user.display_name || user.email?.split('@')[0];
          const entry = { display: display || 'Unknown', type: user.type === 'agent' ? 'agent' as const : 'human' as const };
          // Index by id (for message author resolution) AND by email (for DM name resolution)
          if (user.id) next[user.id] = entry;
          if (user.email) next[user.email] = entry;
        }
        setUserDirectory(next);
      } catch (error) {
        console.error('Failed to load users directory:', error);
      }
    };

    void loadUsersDirectory();
  }, []);

  // Load members for the selected channel
  useEffect(() => {
    if (!selectedChannelId || isViewingDM) return;

    const loadChannelMembers = async () => {
      try {
        const members = await chatApi.getMembers(selectedChannelId);
        
        // Convert API response to Member format
        const memberList = members.map(member => {
          const isAgent = member.type
            ? member.type === 'agent'
            : member.email.includes('@agentunited.local');

          return {
            id: member.id,
            name: member.display_name || member.name || member.email.split('@')[0] || member.email,
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

  const hasUuid = (value?: string) =>
    !!value && /[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}/i.test(value);

  const normalizeLabel = (name: string) =>
    name
      .replace(/\b(Agent)\s+\1\b/gi, '$1')
      .replace(/\s{2,}/g, ' ')
      .trim();

  const displayDirectMessages = useMemo(() => {
    return directMessages.map((dm) => {
      const resolved = userDirectory[dm.name];
      if (resolved) {
        return { ...dm, name: normalizeLabel(getDisplayName(resolved.display)), type: resolved.type };
      }

      if (hasUuid(dm.name)) {
        return { ...dm, name: 'Direct Message' };
      }

      return { ...dm, name: normalizeLabel(getDisplayName(dm.name)) };
    });
  }, [directMessages, userDirectory]);

  const selectedDM = displayDirectMessages.find(dm => dm.id === selectedDMId) || null;

  const channelMemberDirectory = useMemo(() => {
    const directory: Record<string, string> = {};

    for (const member of channelMembers) {
      const resolvedName = getDisplayName(member.name);
      directory[member.id] = resolvedName;
      directory[member.email] = resolvedName;
      directory[member.email.toLowerCase()] = resolvedName;

      const localPart = member.email.split('@')[0];
      if (localPart) {
        directory[localPart] = resolvedName;
        directory[localPart.toLowerCase()] = resolvedName;
      }
    }

    return directory;
  }, [channelMembers]);

  const displayMessages = useMemo(() => {
    return messages.map((msg) => {
      const candidates = [msg.authorId, msg.author].filter(Boolean) as string[];

      let preferredName: string | undefined;

      for (const rawCandidate of candidates) {
        const candidate = rawCandidate.trim();
        const lowerCandidate = candidate.toLowerCase();

        const directoryHit =
          userDirectory[candidate] ||
          userDirectory[lowerCandidate] ||
          (candidate.includes('@') ? userDirectory[candidate.toLowerCase()] : undefined);

        if (directoryHit?.display) {
          preferredName = getDisplayName(directoryHit.display);
          break;
        }

        const memberHit =
          channelMemberDirectory[candidate] ||
          channelMemberDirectory[lowerCandidate];

        if (memberHit) {
          preferredName = memberHit;
          break;
        }
      }

      if (!preferredName) return msg;

      // Prefer resolved display_name over username/email prefixes in bubbles.
      if (preferredName && msg.author !== preferredName) {
        return { ...msg, author: preferredName };
      }

      if (msg.author === msg.authorId || hasUuid(msg.author)) {
        return { ...msg, author: preferredName };
      }

      return msg;
    });
  }, [messages, userDirectory, channelMemberDirectory]);

  const markConversationRead = useCallback(async (kind: 'channel' | 'dm', id: string) => {
    try {
      if (kind === 'channel') {
        await chatApi.markChannelRead(id);
        setChannels((prev) => prev.map((ch) => (ch.id === id ? { ...ch, unread: 0 } : ch)));
      } else {
        await chatApi.markDMRead(id);
        setDirectMessages((prev) => prev.map((dm) => (dm.id === id ? { ...dm, unread: 0 } : dm)));
      }
    } catch (error) {
      console.error(`Failed to mark ${kind} as read:`, error);
    }
  }, []);

  const handleSendMessage = async (text: string, mentions?: { userId: string; display: string }[], attachment?: File) => {
    if (!activeConversationId) return;
    
    try {
      let sent = false;
      if (attachment) {
        // Use API directly for file uploads
        await sendMessageWithAttachment(activeConversationId, text, attachment);
        sent = true;
      } else {
        // Prefer WebSocket, but fallback to HTTP when reconnecting/disconnected.
        sent = await sendMessage(text);
      }

      if (!sent) {
        throw new Error('Failed to send message');
      }

      if (isViewingDM && selectedDM?.type === 'agent' && !selectedDM.online) {
        setStartupToast('Agent is offline — your message was delivered and will be read when they reconnect.');
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
    navigate('/chat');
    void markConversationRead('channel', channelId);
  };

  const handleDMSelect = (dmId: string) => {
    setSelectedDMId(dmId);
    setSelectedChannelId(''); // Clear channel selection
    setSidebarOpen(false);
    navigate(`/chat/dm/${dmId}`);
    void markConversationRead('dm', dmId);
  };

  const handleCreateChannel = useCallback(async (name: string, description: string, selectedAgentIds: string[] = []) => {
    const newChannel = await chatApi.createChannel(name, description);

    let failedInvites = 0;
    for (const agentId of selectedAgentIds) {
      try {
        await chatApi.addChannelMember(newChannel.id, agentId, 'member');
      } catch (error) {
        failedInvites += 1;
        console.error(`Failed to invite agent ${agentId}:`, error);
      }
    }

    setChannels(prev => [...prev, newChannel]);
    setSelectedChannelId(newChannel.id);
    setSelectedDMId(''); // Clear DM selection

    if (failedInvites > 0) {
      setChannelCreateWarning(`Channel created, but ${failedInvites} agent invite(s) failed. You can retry from channel settings.`);
    } else {
      setChannelCreateWarning(null);
    }
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
    if (!dmId) {
      setStartupToast('Workspace loading, please try again');
      return;
    }

    try {
      // Reload DMs to get the new one
      const dms = await chatApi.listDMs();
      const directMessageList: DirectMessage[] = dms.map(dm => ({
        id: dm.id,
        name: dm.name,
        type: /agent/i.test(dm.name) ? 'agent' : 'human',
        online: false,
        unread: dm.unread
      }));

      setDirectMessages(directMessageList);

      const created = directMessageList.find((dm) => dm.id === dmId);
      if (!created) {
        setStartupToast('Workspace loading, please try again');
        return;
      }

      // Select the new DM
      setSelectedDMId(dmId);
      setSelectedChannelId(''); // Clear channel selection
    } catch (error) {
      console.error('Failed to refresh DMs:', error);
      setStartupToast('Workspace loading, please try again');
    }
  }, []);

  const handleToggleMembers = useCallback(() => {
    setShowMembersPanel(prev => !prev);
  }, []);

  useEffect(() => {
    if (dmId) {
      setSelectedDMId(dmId);
      setSelectedChannelId('');
    }
  }, [dmId]);

  // First-login onboarding fallback: if invite accept did not return dm_channel_id,
  // auto-open a DM with the first available agent once users list is loaded.
  useEffect(() => {
    if (searchParams.get('first_login') !== '1') return;
    if (selectedChannelId || selectedDMId || dmId) return;

    const firstAgent = Object.entries(userDirectory).find(([, user]) => user.type === 'agent');
    if (!firstAgent) return;

    const createStarterDm = async () => {
      try {
        const dm = await chatApi.createDM(firstAgent[0]);
        setSelectedDMId(dm.id);
        setSelectedChannelId('');
        navigate(`/chat/dm/${dm.id}`);
        const next = new URLSearchParams(searchParams);
        next.delete('first_login');
        setSearchParams(next, { replace: true });
      } catch {
        setStartupToast('You\'re in. Start a conversation with your agent ↓');
      }
    };

    void createStarterDm();
  }, [searchParams, selectedChannelId, selectedDMId, dmId, userDirectory, navigate, setSearchParams]);

  useEffect(() => {
    if (selectedDMId) {
      void markConversationRead('dm', selectedDMId);
    } else if (selectedChannelId) {
      void markConversationRead('channel', selectedChannelId);
    }
  }, [selectedChannelId, selectedDMId, markConversationRead]);

  useEffect(() => {
    if (!startupToast) return;
    const t = window.setTimeout(() => setStartupToast(null), 3000);
    return () => window.clearTimeout(t);
  }, [startupToast]);

  useEffect(() => {
    if (!channelCreateWarning) return;
    const timeout = window.setTimeout(() => setChannelCreateWarning(null), 5000);
    return () => window.clearTimeout(timeout);
  }, [channelCreateWarning]);

  useEffect(() => {
    const onFocus = () => {
      if (selectedDMId) {
        void markConversationRead('dm', selectedDMId);
      } else if (selectedChannelId) {
        void markConversationRead('channel', selectedChannelId);
      }
    };

    window.addEventListener('focus', onFocus);
    return () => window.removeEventListener('focus', onFocus);
  }, [selectedChannelId, selectedDMId, markConversationRead]);

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

  // No early return for channels error or empty channels — render the full sidebar so DMs are always accessible.

  // Determine the active conversation name and type
  const activeConversationName = isViewingDM 
    ? (selectedDM?.name || 'Direct Message')
    : selectedChannel?.name || 'unknown';
  
  const activeConversationTopic = isViewingDM 
    ? `Direct message with ${selectedDM?.name || 'participant'}`
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
          directMessages={displayDirectMessages}
          activeChannelId={activeConversationId}
          isOpen={sidebarOpen}
          onChannelSelect={handleSelectChannel}
          onDMSelect={handleDMSelect}
          connectionStatus={connectionStatus}
          onCreateChannel={() => setShowCreateChannel(true)}
          onNewDM={() => setShowNewDM(true)}
          onSearch={handleSearch}
          onChannelUpdate={handleChannelUpdate}
          onChannelDelete={handleChannelDelete}
          onOpenSettings={() => {
            setSidebarOpen(false);
            navigate('/settings/profile');
          }}
          onOpenIntegrations={() => {
            setSidebarOpen(false);
            navigate('/settings/integrations');
          }}
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
          {channelCreateWarning && (
            <div className="border-b border-amber-400/25 bg-amber-500/10 px-4 py-2 text-sm text-amber-700 dark:text-amber-300">
              {channelCreateWarning}
            </div>
          )}

          {channelsError && (
            <div className="border-b border-red-400/25 bg-red-500/10 px-4 py-2 text-sm text-red-700 dark:text-red-300">
              Couldn&apos;t load channel list. You can still use direct messages.
            </div>
          )}

          {isSearching ? (
            // Show search results instead of normal chat
            <SearchResultsPanel
              query={searchQuery}
              channels={channels}
              onClose={handleCloseSearch}
              onResultClick={handleSearchResultClick}
            />
          ) : !activeConversationId ? (
            // No channel or DM selected — show welcome state instead of broken #unknown
            <div className="flex flex-1 items-center justify-center bg-background">
              <div className="text-center px-6 max-w-sm">
                <div className="mx-auto mb-4 flex h-12 w-12 items-center justify-center rounded-full border border-emerald-500/45 shadow-[0_0_18px_rgba(16,185,129,0.35)]">
                  <div className="h-2.5 w-2.5 rounded-full bg-emerald-500" />
                </div>
                <h2 className="text-lg font-semibold text-foreground mb-2">No conversations yet</h2>
                <p className="text-sm text-muted-foreground mb-4">
                  Start a direct message with your agent using the <strong>+</strong> button in the sidebar.
                </p>
                <button
                  onClick={() => setShowNewDM(true)}
                  className="inline-flex items-center gap-2 rounded-lg bg-emerald-500 px-4 py-2 text-sm font-medium text-white hover:bg-emerald-600 transition-colors"
                >
                  Start a conversation
                </button>
              </div>
            </div>
          ) : (
            // Show normal chat interface
            <>
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
                messages={displayMessages}
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
                  ? `Message ${selectedDM?.name || 'participant'}`
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

      {startupToast && (
        <div className="fixed right-4 bottom-4 z-50 rounded-md border border-slate-300 bg-slate-900 px-3 py-2 text-sm text-white shadow-lg">
          {startupToast}
        </div>
      )}
    </div>
  );
}