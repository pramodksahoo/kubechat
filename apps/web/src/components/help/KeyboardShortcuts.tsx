import React, { useState, useEffect, useRef, useMemo } from 'react';
import { Button, Heading, Body, Input, StatusBadge } from '../../design-system';
import { cn } from '../../lib/utils';
import { useFocusTrap, useKeyboardShortcuts } from '../../design-system/hooks/useKeyboardNavigation';
import { ScreenReaderUtils } from '../../design-system/accessibility';
import { trackEvent } from '../../lib/monitoring';

// Keyboard shortcut definition
export interface KeyboardShortcut {
  id: string;
  keys: string[];
  description: string;
  category: string;
  context?: string;
  global?: boolean;
  platform?: 'mac' | 'windows' | 'linux' | 'all';
}

// Shortcut category
export interface ShortcutCategory {
  id: string;
  name: string;
  description: string;
  icon: React.ReactNode;
  shortcuts: KeyboardShortcut[];
}

// Platform detection
const getPlatform = (): 'mac' | 'windows' | 'linux' => {
  if (typeof window === 'undefined') return 'windows';
  
  const platform = window.navigator.platform.toLowerCase();
  if (platform.includes('mac')) return 'mac';
  if (platform.includes('linux')) return 'linux';
  return 'windows';
};

// Key display mapping for different platforms
const keyDisplayMap: Record<string, { mac: string; windows: string; linux: string }> = {
  cmd: { mac: '⌘', windows: 'Ctrl', linux: 'Ctrl' },
  ctrl: { mac: '⌃', windows: 'Ctrl', linux: 'Ctrl' },
  alt: { mac: '⌥', windows: 'Alt', linux: 'Alt' },
  shift: { mac: '⇧', windows: 'Shift', linux: 'Shift' },
  tab: { mac: '⇥', windows: 'Tab', linux: 'Tab' },
  enter: { mac: '↩', windows: 'Enter', linux: 'Enter' },
  escape: { mac: '⎋', windows: 'Esc', linux: 'Esc' },
  space: { mac: '␣', windows: 'Space', linux: 'Space' },
  backspace: { mac: '⌫', windows: 'Backspace', linux: 'Backspace' },
  delete: { mac: '⌦', windows: 'Delete', linux: 'Delete' },
  up: { mac: '↑', windows: '↑', linux: '↑' },
  down: { mac: '↓', windows: '↓', linux: '↓' },
  left: { mac: '←', windows: '←', linux: '←' },
  right: { mac: '→', windows: '→', linux: '→' },
};

// Default keyboard shortcuts for KubeChat
const defaultShortcuts: ShortcutCategory[] = [
  {
    id: 'navigation',
    name: 'Navigation',
    description: 'Navigate through the application',
    icon: (
      <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 20l-5.447-2.724A1 1 0 013 16.382V5.618a1 1 0 011.447-.894L9 7m0 13l6-3m-6 3V7m6 10l4.553 2.276A1 1 0 0021 18.382V7.618a1 1 0 00-1.447-.894L15 4m0 13V4m-6 3l6-3" />
      </svg>
    ),
    shortcuts: [
      {
        id: 'open-command-palette',
        keys: ['cmd', 'k'],
        description: 'Open command palette',
        category: 'navigation',
        global: true,
      },
      {
        id: 'search',
        keys: ['cmd', '/'],
        description: 'Focus search',
        category: 'navigation',
        global: true,
      },
      {
        id: 'go-to-dashboard',
        keys: ['g', 'd'],
        description: 'Go to dashboard',
        category: 'navigation',
        global: true,
      },
      {
        id: 'go-to-clusters',
        keys: ['g', 'c'],
        description: 'Go to clusters',
        category: 'navigation',
        global: true,
      },
      {
        id: 'toggle-sidebar',
        keys: ['cmd', 'b'],
        description: 'Toggle sidebar',
        category: 'navigation',
        global: true,
      },
    ],
  },
  {
    id: 'kubectl',
    name: 'kubectl Commands',
    description: 'Execute and manage kubectl commands',
    icon: (
      <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 9l3 3-3 3m5 0h3M5 20h14a2 2 0 002-2V6a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
      </svg>
    ),
    shortcuts: [
      {
        id: 'new-command',
        keys: ['cmd', 'shift', 'k'],
        description: 'New kubectl command',
        category: 'kubectl',
        global: true,
      },
      {
        id: 'execute-command',
        keys: ['cmd', 'enter'],
        description: 'Execute command',
        category: 'kubectl',
        context: 'command-input',
      },
      {
        id: 'clear-command',
        keys: ['cmd', 'shift', 'c'],
        description: 'Clear command input',
        category: 'kubectl',
        context: 'command-input',
      },
      {
        id: 'command-history',
        keys: ['cmd', 'h'],
        description: 'Open command history',
        category: 'kubectl',
        context: 'command-input',
      },
      {
        id: 'previous-command',
        keys: ['up'],
        description: 'Previous command in history',
        category: 'kubectl',
        context: 'command-input',
      },
      {
        id: 'next-command',
        keys: ['down'],
        description: 'Next command in history',
        category: 'kubectl',
        context: 'command-input',
      },
    ],
  },
  {
    id: 'general',
    name: 'General',
    description: 'General application shortcuts',
    icon: (
      <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
      </svg>
    ),
    shortcuts: [
      {
        id: 'help',
        keys: ['?'],
        description: 'Show keyboard shortcuts',
        category: 'general',
        global: true,
      },
      {
        id: 'close-modal',
        keys: ['escape'],
        description: 'Close modal or dialog',
        category: 'general',
        global: true,
      },
      {
        id: 'save',
        keys: ['cmd', 's'],
        description: 'Save current changes',
        category: 'general',
        global: true,
      },
      {
        id: 'refresh',
        keys: ['cmd', 'r'],
        description: 'Refresh current view',
        category: 'general',
        global: true,
      },
      {
        id: 'toggle-theme',
        keys: ['cmd', 'shift', 't'],
        description: 'Toggle dark/light theme',
        category: 'general',
        global: true,
      },
    ],
  },
  {
    id: 'accessibility',
    name: 'Accessibility',
    description: 'Accessibility navigation shortcuts',
    icon: (
      <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197m13.5-9a2.5 2.5 0 11-5 0 2.5 2.5 0 015 0z" />
      </svg>
    ),
    shortcuts: [
      {
        id: 'skip-to-content',
        keys: ['alt', 'c'],
        description: 'Skip to main content',
        category: 'accessibility',
        global: true,
      },
      {
        id: 'skip-to-nav',
        keys: ['alt', 'n'],
        description: 'Skip to navigation',
        category: 'accessibility',
        global: true,
      },
      {
        id: 'focus-search',
        keys: ['alt', 's'],
        description: 'Focus search input',
        category: 'accessibility',
        global: true,
      },
      {
        id: 'next-heading',
        keys: ['alt', 'h'],
        description: 'Jump to next heading',
        category: 'accessibility',
        global: true,
      },
      {
        id: 'previous-heading',
        keys: ['alt', 'shift', 'h'],
        description: 'Jump to previous heading',
        category: 'accessibility',
        global: true,
      },
    ],
  },
];

// Keyboard shortcuts guide component
export interface KeyboardShortcutsGuideProps {
  isOpen: boolean;
  onClose: () => void;
  customShortcuts?: ShortcutCategory[];
  platform?: 'mac' | 'windows' | 'linux';
  className?: string;
}

export function KeyboardShortcutsGuide({
  isOpen,
  onClose,
  customShortcuts = [],
  platform: platformProp,
  className,
}: KeyboardShortcutsGuideProps) {
  const [searchQuery, setSearchQuery] = useState('');
  const [selectedCategory, setSelectedCategory] = useState<string | null>(null);
  const containerRef = useRef<HTMLDivElement>(null);
  
  const platform = platformProp || getPlatform();
  
  // Focus trap
  useFocusTrap(isOpen, containerRef);

  // Keyboard shortcuts for the guide itself
  useKeyboardShortcuts({
    'escape': onClose,
    '?': onClose,
  }, isOpen);

  // Filter shortcuts based on search and platform
  const filteredCategories = useMemo(() => {
    // Combine default and custom shortcuts
    const allCategories = [...defaultShortcuts, ...customShortcuts];
    
    return allCategories.map(category => {
      const filteredShortcuts = category.shortcuts.filter(shortcut => {
        // Platform filter
        if (shortcut.platform && shortcut.platform !== 'all' && shortcut.platform !== platform) {
          return false;
        }

        // Search filter
        if (searchQuery) {
          const query = searchQuery.toLowerCase();
          return (
            shortcut.description.toLowerCase().includes(query) ||
            shortcut.keys.some(key => key.toLowerCase().includes(query)) ||
            shortcut.category.toLowerCase().includes(query)
          );
        }

        return true;
      });

      return {
        ...category,
        shortcuts: filteredShortcuts,
      };
    }).filter(category => category.shortcuts.length > 0);
  }, [customShortcuts, platform, searchQuery]);

  // Track guide opening
  useEffect(() => {
    if (isOpen) {
      trackEvent('keyboard_shortcuts_opened', { platform });
      ScreenReaderUtils.announce('Keyboard shortcuts guide opened');
    }
  }, [isOpen, platform]);

  // Format key combination for display (currently unused)
  // const formatKeyCombo = (keys: string[]): string[] => {
  //   return keys.map(key => {
  //     const mapping = keyDisplayMap[key.toLowerCase()];
  //     return mapping ? mapping[platform] : key;
  //   });
  // };

  if (!isOpen) return null;

  return (
    <div className={cn(
      'fixed inset-0 z-modal flex items-center justify-center p-4 bg-black/50 backdrop-blur-sm',
      className
    )}>
      <div
        ref={containerRef}
        className="relative w-full max-w-4xl bg-background-primary rounded-modal shadow-modal max-h-[90vh] overflow-hidden"
        role="dialog"
        aria-labelledby="shortcuts-title"
        aria-describedby="shortcuts-description"
      >
        {/* Header */}
        <div className="flex items-center justify-between p-6 border-b border-border-primary">
          <div>
            <Heading level={2} id="shortcuts-title">
              Keyboard Shortcuts
            </Heading>
            <Body size="sm" color="secondary" id="shortcuts-description">
              Navigate KubeChat efficiently with these keyboard shortcuts
            </Body>
          </div>
          
          <div className="flex items-center gap-3">
            <StatusBadge variant="neutral">
              {platform === 'mac' ? 'macOS' : platform === 'windows' ? 'Windows' : 'Linux'}
            </StatusBadge>
            <Button variant="ghost" size="sm" onClick={onClose} aria-label="Close shortcuts guide">
              <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
              </svg>
            </Button>
          </div>
        </div>

        {/* Search */}
        <div className="p-4 border-b border-border-primary bg-background-secondary">
          <Input
            type="search"
            placeholder="Search shortcuts..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            leftIcon={
              <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="m21 21-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
              </svg>
            }
          />
        </div>

        {/* Content */}
        <div className="flex max-h-[60vh]">
          {/* Category sidebar */}
          <div className="w-64 bg-background-secondary border-r border-border-primary overflow-y-auto scrollbar-thin">
            <div className="p-4">
              <Body size="sm" className="font-medium mb-3 text-text-secondary uppercase tracking-wide">
                Categories
              </Body>
              
              <div className="space-y-1">
                <button
                  onClick={() => setSelectedCategory(null)}
                  className={cn(
                    'w-full text-left p-2 rounded-lg transition-colors flex items-center gap-3',
                    selectedCategory === null 
                      ? 'bg-primary-100 dark:bg-primary-900/20 text-primary-700 dark:text-primary-300'
                      : 'hover:bg-background-tertiary'
                  )}
                >
                  <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
                  </svg>
                  <Body size="sm">All Shortcuts</Body>
                </button>
                
                {filteredCategories.map((category) => (
                  <button
                    key={category.id}
                    onClick={() => setSelectedCategory(category.id)}
                    className={cn(
                      'w-full text-left p-2 rounded-lg transition-colors flex items-center gap-3',
                      selectedCategory === category.id 
                        ? 'bg-primary-100 dark:bg-primary-900/20 text-primary-700 dark:text-primary-300'
                        : 'hover:bg-background-tertiary'
                    )}
                  >
                    <div className="w-4 h-4 text-current">{category.icon}</div>
                    <div className="flex-1">
                      <Body size="sm">{category.name}</Body>
                      <Body size="xs" color="tertiary">
                        {category.shortcuts.length} shortcuts
                      </Body>
                    </div>
                  </button>
                ))}
              </div>
            </div>
          </div>

          {/* Shortcuts list */}
          <div className="flex-1 overflow-y-auto scrollbar-thin">
            <div className="p-6">
              {selectedCategory ? (
                /* Single category view */
                (() => {
                  const category = filteredCategories.find(c => c.id === selectedCategory);
                  if (!category) return null;
                  
                  return (
                    <div>
                      <div className="flex items-center gap-3 mb-6">
                        <div className="w-6 h-6 text-text-secondary">{category.icon}</div>
                        <div>
                          <Heading level={3}>{category.name}</Heading>
                          <Body size="sm" color="secondary">{category.description}</Body>
                        </div>
                      </div>
                      
                      <ShortcutsList shortcuts={category.shortcuts} platform={platform} />
                    </div>
                  );
                })()
              ) : (
                /* All categories view */
                <div className="space-y-8">
                  {filteredCategories.map((category) => (
                    <div key={category.id}>
                      <div className="flex items-center gap-3 mb-4">
                        <div className="w-5 h-5 text-text-secondary">{category.icon}</div>
                        <Heading level={4}>{category.name}</Heading>
                      </div>
                      
                      <ShortcutsList shortcuts={category.shortcuts} platform={platform} />
                    </div>
                  ))}
                </div>
              )}
              
              {filteredCategories.length === 0 && (
                <div className="text-center py-12">
                  <div className="w-16 h-16 mx-auto mb-4 text-text-tertiary">
                    <svg fill="currentColor" viewBox="0 0 20 20" className="w-full h-full">
                      <path fillRule="evenodd" d="M8 4a4 4 0 100 8 4 4 0 000-8zM2 8a6 6 0 1110.89 3.476l4.817 4.817a1 1 0 01-1.414 1.414l-4.816-4.816A6 6 0 012 8z" clipRule="evenodd" />
                    </svg>
                  </div>
                  <Heading level={4} className="mb-2">No shortcuts found</Heading>
                  <Body color="secondary">
                    Try a different search term or browse all categories
                  </Body>
                </div>
              )}
            </div>
          </div>
        </div>

        {/* Footer */}
        <div className="p-4 border-t border-border-primary bg-background-secondary">
          <Body size="sm" color="tertiary" className="text-center">
            Press <kbd className="px-1.5 py-0.5 bg-gray-200 dark:bg-gray-700 rounded text-xs">?</kbd> or 
            <kbd className="px-1.5 py-0.5 bg-gray-200 dark:bg-gray-700 rounded text-xs mx-1">Esc</kbd> to close this guide
          </Body>
        </div>
      </div>
    </div>
  );
}

// Shortcuts list component
function ShortcutsList({ 
  shortcuts, 
  platform 
}: { 
  shortcuts: KeyboardShortcut[];
  platform: 'mac' | 'windows' | 'linux';
}) {
  const formatKeyCombo = (keys: string[]): string[] => {
    return keys.map(key => {
      const mapping = keyDisplayMap[key.toLowerCase()];
      return mapping ? mapping[platform] : key;
    });
  };

  return (
    <div className="space-y-3">
      {shortcuts.map((shortcut) => (
        <div
          key={shortcut.id}
          className="flex items-center justify-between p-3 rounded-lg hover:bg-background-secondary transition-colors"
        >
          <div className="flex-1">
            <Body className="mb-1">{shortcut.description}</Body>
            {shortcut.context && (
              <Body size="sm" color="tertiary">
                Context: {shortcut.context}
              </Body>
            )}
          </div>
          
          <div className="flex items-center gap-1">
            {formatKeyCombo(shortcut.keys).map((key, index, array) => (
              <React.Fragment key={index}>
                <kbd className="px-2 py-1 text-xs bg-gray-200 dark:bg-gray-700 rounded font-mono">
                  {key}
                </kbd>
                {index < array.length - 1 && (
                  <span className="text-text-tertiary mx-1">+</span>
                )}
              </React.Fragment>
            ))}
          </div>
        </div>
      ))}
    </div>
  );
}

// Hook for keyboard shortcuts guide
export function useKeyboardShortcutsGuide() {
  const [isOpen, setIsOpen] = useState(false);

  const open = () => setIsOpen(true);
  const close = () => setIsOpen(false);
  const toggle = () => setIsOpen(!isOpen);

  // Global shortcut to open guide
  useKeyboardShortcuts({
    '?': toggle,
  });

  return {
    isOpen,
    open,
    close,
    toggle,
  };
}

// Quick shortcut display component
export interface QuickShortcutProps {
  keys: string[];
  className?: string;
  size?: 'sm' | 'md';
}

export function QuickShortcut({ 
  keys, 
  className,
  size = 'sm' 
}: QuickShortcutProps) {
  const platform = getPlatform();
  
  const formatKeyCombo = (keys: string[]): string[] => {
    return keys.map(key => {
      const mapping = keyDisplayMap[key.toLowerCase()];
      return mapping ? mapping[platform] : key;
    });
  };

  const sizeClasses = {
    sm: 'text-xs px-1.5 py-0.5',
    md: 'text-sm px-2 py-1',
  };

  return (
    <div className={cn('inline-flex items-center gap-1', className)}>
      {formatKeyCombo(keys).map((key, index, array) => (
        <React.Fragment key={index}>
          <kbd className={cn(
            'bg-gray-200 dark:bg-gray-700 rounded font-mono',
            sizeClasses[size]
          )}>
            {key}
          </kbd>
          {index < array.length - 1 && (
            <span className="text-text-tertiary text-xs">+</span>
          )}
        </React.Fragment>
      ))}
    </div>
  );
}