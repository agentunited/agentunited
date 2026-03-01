import React, { useState, useEffect, useRef, useCallback } from 'react';
import { UserData } from '../../services/WebSocketService';
import { Avatar } from './Avatar';
import './MentionAutocomplete.css';

export interface MentionMatch {
  id: string;
  name: string;
  type: 'agent' | 'human';
  status?: 'online' | 'offline' | 'away';
}

export interface MentionAutocompleteProps {
  users: UserData[];
  onSelectMention: (mention: MentionMatch, startPos: number, endPos: number) => void;
  onClose: () => void;
  inputRef: React.RefObject<HTMLInputElement | HTMLTextAreaElement>;
  isOpen: boolean;
  query: string;
  position: { top: number; left: number };
}

export function MentionAutocomplete({
  users,
  onSelectMention,
  onClose,
  inputRef,
  isOpen,
  query,
  position
}: MentionAutocompleteProps) {
  const [selectedIndex, setSelectedIndex] = useState(0);
  const dropdownRef = useRef<HTMLDivElement>(null);

  // Filter users based on query
  const filteredUsers = users.filter(user =>
    user.name.toLowerCase().includes(query.toLowerCase())
  ).slice(0, 8); // Limit to 8 results

  // Reset selection when query changes
  useEffect(() => {
    setSelectedIndex(0);
  }, [query]);

  // Handle keyboard navigation
  useEffect(() => {
    if (!isOpen) return;

    const handleKeyDown = (e: KeyboardEvent) => {
      switch (e.key) {
        case 'ArrowDown':
          e.preventDefault();
          setSelectedIndex(prev => 
            prev < filteredUsers.length - 1 ? prev + 1 : 0
          );
          break;
          
        case 'ArrowUp':
          e.preventDefault();
          setSelectedIndex(prev => 
            prev > 0 ? prev - 1 : filteredUsers.length - 1
          );
          break;
          
        case 'Enter':
        case 'Tab':
          e.preventDefault();
          if (filteredUsers[selectedIndex]) {
            handleSelectUser(filteredUsers[selectedIndex]);
          }
          break;
          
        case 'Escape':
          e.preventDefault();
          onClose();
          break;
      }
    };

    document.addEventListener('keydown', handleKeyDown);
    return () => document.removeEventListener('keydown', handleKeyDown);
  }, [isOpen, filteredUsers, selectedIndex, onClose]);

  // Auto-scroll selected item into view
  useEffect(() => {
    if (!dropdownRef.current) return;

    const selectedElement = dropdownRef.current.children[selectedIndex] as HTMLElement;
    if (selectedElement) {
      selectedElement.scrollIntoView({ block: 'nearest' });
    }
  }, [selectedIndex]);

  const handleSelectUser = useCallback((user: UserData) => {
    const input = inputRef.current;
    if (!input) return;

    const text = input.value;
    const cursorPos = input.selectionStart || 0;
    
    // Find the start of the mention (look backwards for @)
    let startPos = cursorPos;
    while (startPos > 0 && text[startPos - 1] !== '@') {
      startPos--;
    }
    if (startPos > 0) startPos--; // Include the @

    onSelectMention(
      {
        id: user.id,
        name: user.name,
        type: user.type,
        status: user.status
      },
      startPos,
      cursorPos
    );
  }, [inputRef, onSelectMention]);

  if (!isOpen || filteredUsers.length === 0) {
    return null;
  }

  return (
    <div 
      ref={dropdownRef}
      className="mention-autocomplete"
      style={{
        position: 'absolute',
        top: position.top,
        left: position.left,
      }}
    >
      <div className="mention-autocomplete__header">
        <span className="mention-autocomplete__title">People</span>
        <span className="mention-autocomplete__hint">
          ↵ to select · ↑↓ to navigate
        </span>
      </div>
      
      <div className="mention-autocomplete__list">
        {filteredUsers.map((user, index) => (
          <button
            key={user.id}
            className={`mention-autocomplete__item ${
              index === selectedIndex ? 'mention-autocomplete__item--selected' : ''
            }`}
            onClick={() => handleSelectUser(user)}
            onMouseEnter={() => setSelectedIndex(index)}
          >
            <div className="mention-autocomplete__avatar">
              <Avatar 
                name={user.name} 
                type={user.type} 
                status={user.status}
                size="sm"
              />
            </div>
            
            <div className="mention-autocomplete__info">
              <span className="mention-autocomplete__name">
                {user.name}
              </span>
              <span className="mention-autocomplete__type">
                {user.type === 'agent' ? '🤖 Agent' : '👤 Human'}
                {user.status === 'online' && (
                  <span className="mention-autocomplete__status"> • Online</span>
                )}
              </span>
            </div>
            
            <div className="mention-autocomplete__shortcut">
              @{user.name.toLowerCase().replace(/\s+/g, '')}
            </div>
          </button>
        ))}
      </div>
    </div>
  );
}

// Hook for managing mention autocomplete state
export function useMentionAutocomplete(inputRef: React.RefObject<HTMLInputElement | HTMLTextAreaElement>) {
  const [isOpen, setIsOpen] = useState(false);
  const [query, setQuery] = useState('');
  const [position, setPosition] = useState({ top: 0, left: 0 });

  const checkForMention = useCallback(() => {
    const input = inputRef.current;
    if (!input) return;

    const text = input.value;
    const cursorPos = input.selectionStart || 0;
    
    // Look backwards from cursor for @
    let mentionStart = -1;
    for (let i = cursorPos - 1; i >= 0; i--) {
      if (text[i] === '@') {
        mentionStart = i;
        break;
      }
      if (text[i] === ' ' || text[i] === '\n') {
        break; // Space or newline breaks the mention
      }
    }

    if (mentionStart === -1) {
      setIsOpen(false);
      return;
    }

    // Extract query after @
    const mentionQuery = text.slice(mentionStart + 1, cursorPos);
    
    // Only show autocomplete if query is not too long and doesn't contain spaces
    if (mentionQuery.length <= 20 && !mentionQuery.includes(' ') && !mentionQuery.includes('\n')) {
      setQuery(mentionQuery);
      
      // Calculate position for dropdown
      const rect = input.getBoundingClientRect();
      const style = window.getComputedStyle(input);
      const lineHeight = parseInt(style.lineHeight) || 20;
      
      setPosition({
        top: rect.bottom + 4,
        left: rect.left
      });
      
      setIsOpen(true);
    } else {
      setIsOpen(false);
    }
  }, [inputRef]);

  const insertMention = useCallback((mention: MentionMatch, startPos: number, endPos: number) => {
    const input = inputRef.current;
    if (!input) return;

    const text = input.value;
    const mentionText = `@${mention.name}`;
    const newText = text.slice(0, startPos) + mentionText + ' ' + text.slice(endPos);
    
    // Update input value
    input.value = newText;
    
    // Set cursor position after the mention
    const newCursorPos = startPos + mentionText.length + 1;
    input.setSelectionRange(newCursorPos, newCursorPos);
    
    // Trigger input event so React sees the change
    const event = new Event('input', { bubbles: true });
    input.dispatchEvent(event);
    
    setIsOpen(false);
    input.focus();
  }, [inputRef]);

  const closeMention = useCallback(() => {
    setIsOpen(false);
  }, []);

  return {
    isOpen,
    query,
    position,
    checkForMention,
    insertMention,
    closeMention
  };
}