import { useEffect, useRef, useCallback, useState } from 'react';
import { KeyboardNavigation, FocusManager } from '../accessibility';

// Keyboard navigation hook for lists and menus
export function useKeyboardNavigation<T extends HTMLElement>({
  items,
  orientation = 'vertical',
  gridColumns,
  loop = true,
  onSelect,
  onEscape,
  typeahead = false,
  getItemText,
}: {
  items: T[];
  orientation?: 'horizontal' | 'vertical' | 'grid';
  gridColumns?: number;
  loop?: boolean;
  onSelect?: (item: T, index: number) => void;
  onEscape?: () => void;
  typeahead?: boolean;
  getItemText?: (item: T) => string;
}) {
  const [focusedIndex, setFocusedIndex] = useState(-1);
  const typeaheadRef = useRef('');
  const typeaheadTimeoutRef = useRef<NodeJS.Timeout>();

  const handleKeyDown = useCallback((event: KeyboardEvent) => {
    const { key } = event;

    // Handle escape key
    if (key === 'Escape') {
      event.preventDefault();
      onEscape?.();
      return;
    }

    // Handle enter/space for selection
    if ((key === 'Enter' || key === ' ') && focusedIndex >= 0) {
      event.preventDefault();
      const item = items[focusedIndex];
      if (item) {
        onSelect?.(item, focusedIndex);
      }
      return;
    }

    // Handle arrow navigation
    if (['ArrowUp', 'ArrowDown', 'ArrowLeft', 'ArrowRight', 'Home', 'End'].includes(key)) {
      const newIndex = KeyboardNavigation.handleArrowNavigation(
        event,
        items,
        focusedIndex,
        orientation,
        gridColumns
      );
      
      if (newIndex !== focusedIndex) {
        setFocusedIndex(newIndex);
      }
      return;
    }

    // Handle typeahead search
    if (typeahead && getItemText && key.length === 1 && !event.altKey && !event.ctrlKey && !event.metaKey) {
      event.preventDefault();
      
      // Clear previous timeout
      if (typeaheadTimeoutRef.current) {
        clearTimeout(typeaheadTimeoutRef.current);
      }

      // Add to search string
      typeaheadRef.current += key.toLowerCase();

      // Find matching item
      const startIndex = focusedIndex + 1;
      let matchIndex = -1;

      for (let i = 0; i < items.length; i++) {
        const index = loop ? (startIndex + i) % items.length : startIndex + i;
        if (index >= items.length) break;

        const item = items[index];
        const itemText = getItemText(item).toLowerCase();
        
        if (itemText.startsWith(typeaheadRef.current)) {
          matchIndex = index;
          break;
        }
      }

      if (matchIndex >= 0) {
        setFocusedIndex(matchIndex);
        items[matchIndex]?.focus();
      }

      // Reset typeahead after delay
      typeaheadTimeoutRef.current = setTimeout(() => {
        typeaheadRef.current = '';
      }, 500);
    }
  }, [items, focusedIndex, orientation, gridColumns, loop, onSelect, onEscape, typeahead, getItemText]);

  // Focus management
  useEffect(() => {
    if (focusedIndex >= 0 && focusedIndex < items.length) {
      items[focusedIndex]?.focus();
    }
  }, [focusedIndex, items]);

  // Cleanup
  useEffect(() => {
    return () => {
      if (typeaheadTimeoutRef.current) {
        clearTimeout(typeaheadTimeoutRef.current);
      }
    };
  }, []);

  return {
    focusedIndex,
    setFocusedIndex,
    handleKeyDown,
  };
}

// Focus trap hook for modals and dialogs
export function useFocusTrap(
  isActive: boolean,
  containerRef: React.RefObject<HTMLElement>
) {
  const previouslyFocusedElementRef = useRef<HTMLElement | null>(null);
  const cleanupRef = useRef<(() => void) | null>(null);

  useEffect(() => {
    if (!isActive || !containerRef.current) return;

    // Store previously focused element
    previouslyFocusedElementRef.current = document.activeElement as HTMLElement;

    // Create focus trap
    cleanupRef.current = FocusManager.createFocusTrap(containerRef.current);

    return () => {
      // Cleanup focus trap
      cleanupRef.current?.();
      
      // Restore focus to previously focused element
      FocusManager.restoreFocus(previouslyFocusedElementRef.current);
    };
  }, [isActive, containerRef]);

  // Also cleanup on unmount
  useEffect(() => {
    return () => {
      cleanupRef.current?.();
    };
  }, []);
}

// Global keyboard shortcuts hook
export function useKeyboardShortcuts(
  shortcuts: Record<string, (event: KeyboardEvent) => void>,
  enabled = true
) {
  useEffect(() => {
    if (!enabled) return;

    const handleKeyDown = (event: KeyboardEvent) => {
      // Build key combination string
      const modifiers = [];
      if (event.ctrlKey || event.metaKey) modifiers.push('cmd');
      if (event.altKey) modifiers.push('alt');
      if (event.shiftKey) modifiers.push('shift');
      
      const key = event.key.toLowerCase();
      const combination = [...modifiers, key].join('+');

      // Check for exact match
      if (shortcuts[combination]) {
        event.preventDefault();
        shortcuts[combination](event);
        return;
      }

      // Check for key-only shortcuts
      if (shortcuts[key] && !event.ctrlKey && !event.metaKey && !event.altKey && !event.shiftKey) {
        shortcuts[key](event);
      }
    };

    document.addEventListener('keydown', handleKeyDown);
    
    return () => {
      document.removeEventListener('keydown', handleKeyDown);
    };
  }, [shortcuts, enabled]);
}

// Command palette navigation hook
export function useCommandPalette() {
  const [isOpen, setIsOpen] = useState(false);
  const [query, setQuery] = useState('');
  const [selectedIndex, setSelectedIndex] = useState(0);

  const open = useCallback(() => {
    setIsOpen(true);
    setQuery('');
    setSelectedIndex(0);
  }, []);

  const close = useCallback(() => {
    setIsOpen(false);
    setQuery('');
    setSelectedIndex(0);
  }, []);

  // Global keyboard shortcut to open command palette
  useKeyboardShortcuts({
    'cmd+k': open,
    'cmd+/': open,
    'ctrl+k': open,
  }, !isOpen);

  return {
    isOpen,
    query,
    selectedIndex,
    setQuery,
    setSelectedIndex,
    open,
    close,
  };
}

// Skip links hook for accessibility
export function useSkipLinks() {
  const skipLinksRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const handleFocus = (event: FocusEvent) => {
      const target = event.target as HTMLElement;
      
      // Show skip links when focused
      if (target?.closest('[data-skip-link]')) {
        skipLinksRef.current?.classList.remove('sr-only');
      }
    };

    const handleBlur = (event: FocusEvent) => {
      const target = event.target as HTMLElement;
      
      // Hide skip links when focus leaves
      if (target?.closest('[data-skip-link]')) {
        setTimeout(() => {
          if (!document.activeElement?.closest('[data-skip-link]')) {
            skipLinksRef.current?.classList.add('sr-only');
          }
        }, 0);
      }
    };

    document.addEventListener('focusin', handleFocus);
    document.addEventListener('focusout', handleBlur);

    return () => {
      document.removeEventListener('focusin', handleFocus);
      document.removeEventListener('focusout', handleBlur);
    };
  }, []);

  return skipLinksRef;
}

// Roving tabindex hook for composite widgets
export function useRovingTabindex<T extends HTMLElement>({
  items,
  defaultIndex = 0,
  orientation = 'horizontal',
}: {
  items: T[];
  defaultIndex?: number;
  orientation?: 'horizontal' | 'vertical';
}) {
  const [activeIndex, setActiveIndex] = useState(defaultIndex);

  // Update tabindex attributes
  useEffect(() => {
    items.forEach((item, index) => {
      if (index === activeIndex) {
        item.setAttribute('tabindex', '0');
      } else {
        item.setAttribute('tabindex', '-1');
      }
    });
  }, [items, activeIndex]);

  const handleKeyDown = useCallback((event: KeyboardEvent, currentIndex: number) => {
    const { key } = event;
    let newIndex = currentIndex;

    switch (key) {
      case 'ArrowLeft':
        if (orientation === 'horizontal') {
          newIndex = currentIndex > 0 ? currentIndex - 1 : items.length - 1;
        }
        break;
        
      case 'ArrowRight':
        if (orientation === 'horizontal') {
          newIndex = currentIndex < items.length - 1 ? currentIndex + 1 : 0;
        }
        break;
        
      case 'ArrowUp':
        if (orientation === 'vertical') {
          newIndex = currentIndex > 0 ? currentIndex - 1 : items.length - 1;
        }
        break;
        
      case 'ArrowDown':
        if (orientation === 'vertical') {
          newIndex = currentIndex < items.length - 1 ? currentIndex + 1 : 0;
        }
        break;
        
      case 'Home':
        newIndex = 0;
        break;
        
      case 'End':
        newIndex = items.length - 1;
        break;
        
      default:
        return;
    }

    if (newIndex !== currentIndex) {
      event.preventDefault();
      setActiveIndex(newIndex);
      items[newIndex]?.focus();
    }
  }, [items, orientation]);

  const handleFocus = useCallback((index: number) => {
    setActiveIndex(index);
  }, []);

  return {
    activeIndex,
    setActiveIndex,
    handleKeyDown,
    handleFocus,
  };
}

// Escape key handler hook
export function useEscapeKey(
  callback: () => void,
  enabled = true
) {
  useEffect(() => {
    if (!enabled) return;

    const handleKeyDown = (event: KeyboardEvent) => {
      if (event.key === 'Escape') {
        callback();
      }
    };

    document.addEventListener('keydown', handleKeyDown);
    
    return () => {
      document.removeEventListener('keydown', handleKeyDown);
    };
  }, [callback, enabled]);
}

// Outside click and escape handler
export function useClickOutside<T extends HTMLElement>(
  ref: React.RefObject<T>,
  callback: () => void,
  enabled = true
) {
  useEffect(() => {
    if (!enabled) return;

    const handleClick = (event: MouseEvent) => {
      if (ref.current && !ref.current.contains(event.target as Node)) {
        callback();
      }
    };

    document.addEventListener('mousedown', handleClick);
    
    return () => {
      document.removeEventListener('mousedown', handleClick);
    };
  }, [ref, callback, enabled]);

  useEscapeKey(callback, enabled);
}