import { useState, useEffect } from 'react';
import { X, Hash, MessageCircle, ArrowUpRight } from 'lucide-react';
import { TypeBadge } from '../ui/TypeBadge';
import { chatApi } from '../../services/chatApi';
import type { Message, Channel } from '../../types/chat';

interface SearchResult extends Message {
  channelName?: string;
  isDM?: boolean;
}

interface SearchResultsPanelProps {
  query: string;
  channels: Channel[];
  onClose: () => void;
  onResultClick: (channelId: string, messageId: string) => void;
}

export function SearchResultsPanel({ query, channels, onClose, onResultClick }: SearchResultsPanelProps) {
  const [results, setResults] = useState<SearchResult[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [hasMore, setHasMore] = useState(false);

  // Load search results when query changes
  useEffect(() => {
    if (!query.trim()) {
      setResults([]);
      return;
    }

    const searchMessages = async () => {
      try {
        setIsLoading(true);
        setError(null);

        // Search across all channels by querying each one
        const allResults: Message[] = [];
        for (const ch of channels) {
          try {
            const chResults = await chatApi.searchMessages(query.trim(), ch.id);
            allResults.push(...chResults);
          } catch { /* skip channels that fail */ }
        }
        const searchResults = allResults.sort((a, b) => 
          new Date(b.timestamp).getTime() - new Date(a.timestamp).getTime()
        );
        
        // Map channel IDs to names using loaded channels
        const enhancedResults = searchResults.map(result => {
          const channel = channels.find(ch => ch.id === result.channelId);
          return {
            ...result,
            channelName: channel?.name || result.channelId || 'Unknown Channel',
            isDM: result.channelId?.startsWith('dm_') || false
          };
        });

        setResults(enhancedResults);
        setHasMore(searchResults.length >= 50); // Assume more if we got the limit
      } catch (error) {
        console.error('Search failed:', error);
        setError(error instanceof Error ? error.message : 'Search failed');
      } finally {
        setIsLoading(false);
      }
    };

    // Debounce search by 300ms
    const timeoutId = setTimeout(searchMessages, 300);
    return () => clearTimeout(timeoutId);
  }, [query]);

  const formatTimestamp = (timestamp: string) => {
    const date = new Date(timestamp);
    const now = new Date();
    const diffDays = Math.floor((now.getTime() - date.getTime()) / (1000 * 60 * 60 * 24));
    
    if (diffDays === 0) {
      return date.toLocaleTimeString([], { hour: 'numeric', minute: '2-digit' });
    } else if (diffDays === 1) {
      return `Yesterday ${date.toLocaleTimeString([], { hour: 'numeric', minute: '2-digit' })}`;
    } else if (diffDays < 7) {
      return date.toLocaleDateString([], { weekday: 'short', hour: 'numeric', minute: '2-digit' });
    } else {
      return date.toLocaleDateString([], { month: 'short', day: 'numeric', hour: 'numeric', minute: '2-digit' });
    }
  };

  const highlightQuery = (text: string, query: string) => {
    if (!query.trim()) return text;
    
    const regex = new RegExp(`(${query.trim()})`, 'gi');
    const parts = text.split(regex);
    
    return parts.map((part, index) => 
      regex.test(part) ? (
        <mark key={index} className="bg-primary/20 text-primary rounded px-0.5">
          {part}
        </mark>
      ) : part
    );
  };

  const handleResultClick = (result: SearchResult) => {
    onResultClick(result.channelId, result.id);
  };

  if (!query.trim()) {
    return null;
  }

  return (
    <div className="flex-1 flex flex-col bg-background">
      {/* Header */}
      <div className="flex items-center justify-between border-b border-border px-5 py-3">
        <div className="flex items-center gap-2">
          <h2 className="text-base font-semibold text-foreground">Search Results</h2>
          <span className="text-sm text-muted-foreground">
            "{query}"
          </span>
        </div>
        <button
          onClick={onClose}
          className="p-1 rounded-md text-muted-foreground hover:text-foreground hover:bg-muted transition-colors"
          aria-label="Close search"
        >
          <X className="h-5 w-5" />
        </button>
      </div>

      {/* Results */}
      <div className="flex-1 overflow-y-auto">
        {isLoading ? (
          <div className="flex items-center justify-center py-12">
            <div className="text-sm text-muted-foreground">Searching...</div>
          </div>
        ) : error ? (
          <div className="px-5 py-8">
            <div className="text-center">
              <div className="text-sm text-destructive mb-2">Search failed</div>
              <div className="text-xs text-destructive-foreground">{error}</div>
            </div>
          </div>
        ) : results.length === 0 ? (
          <div className="px-5 py-12 text-center">
            <div className="text-sm text-muted-foreground mb-2">No results found</div>
            <div className="text-xs text-muted-foreground">
              Try searching for different keywords
            </div>
          </div>
        ) : (
          <>
            {/* Results summary */}
            <div className="px-5 py-3 border-b border-border">
              <div className="text-sm text-muted-foreground">
                {results.length} {results.length === 1 ? 'result' : 'results'}
                {hasMore && ' (showing first 50)'}
              </div>
            </div>

            {/* Results list */}
            <div className="px-5 py-3 space-y-3">
              {results.map((result) => (
                <button
                  key={result.id}
                  onClick={() => handleResultClick(result)}
                  className="w-full text-left p-3 rounded-lg border border-border hover:bg-muted/50 transition-colors group"
                >
                  {/* Result header */}
                  <div className="flex items-center gap-2 mb-2">
                    {result.isDM ? (
                      <MessageCircle className="h-3.5 w-3.5 text-muted-foreground" />
                    ) : (
                      <Hash className="h-3.5 w-3.5 text-muted-foreground" />
                    )}
                    <span className="text-xs font-medium text-muted-foreground">
                      {result.isDM ? 'Direct Message' : `#${result.channelName}`}
                    </span>
                    <span className="text-xs text-muted-foreground">·</span>
                    <span className="text-xs text-muted-foreground">
                      {formatTimestamp(result.timestamp)}
                    </span>
                    <ArrowUpRight className="h-3 w-3 text-muted-foreground opacity-0 group-hover:opacity-100 transition-opacity ml-auto" />
                  </div>

                  {/* Author */}
                  <div className="flex items-center gap-2 mb-1">
                    <span className="text-sm font-medium text-foreground">
                      {result.author}
                    </span>
                    <TypeBadge 
                      type={result.authorType} 
                      className="text-[9px]"
                    />
                  </div>

                  {/* Message content */}
                  <div className="text-sm text-foreground leading-relaxed">
                    {highlightQuery(result.text, query)}
                  </div>
                </button>
              ))}

              {/* Load more button */}
              {hasMore && (
                <div className="pt-4">
                  <button className="w-full py-2 text-sm text-muted-foreground hover:text-foreground transition-colors">
                    Load More Results
                  </button>
                </div>
              )}
            </div>
          </>
        )}
      </div>
    </div>
  );
}