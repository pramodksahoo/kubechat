import React, { useState, useEffect, useRef, useCallback, ReactNode } from 'react';
import { createPortal } from 'react-dom';
import { Body, Heading, Code } from '../../design-system';
import { cn } from '../../lib/utils';
import { AriaUtils } from '../../design-system/accessibility';
import { useKeyboardShortcuts, useClickOutside } from '../../design-system/hooks/useKeyboardNavigation';
import { trackEvent } from '../../lib/monitoring';

// Tooltip positioning
export type TooltipPosition = 
  | 'top' | 'top-start' | 'top-end'
  | 'bottom' | 'bottom-start' | 'bottom-end'
  | 'left' | 'left-start' | 'left-end'
  | 'right' | 'right-start' | 'right-end'
  | 'auto';

// Help content types
export interface HelpContent {
  title?: string;
  description: string;
  examples?: string[];
  links?: HelpLink[];
  shortcuts?: KeyboardShortcut[];
  relatedTopics?: string[];
  category?: 'basic' | 'advanced' | 'troubleshooting' | 'security';
}

export interface HelpLink {
  label: string;
  url: string;
  external?: boolean;
}

export interface KeyboardShortcut {
  keys: string[];
  description: string;
  context?: string;
}

// Tooltip component
export interface TooltipProps {
  content: string | ReactNode;
  position?: TooltipPosition;
  delay?: number;
  disabled?: boolean;
  children: ReactNode;
  className?: string;
  maxWidth?: string;
  showArrow?: boolean;
  interactive?: boolean;
  triggerOn?: 'hover' | 'click' | 'focus' | 'manual';
  onShow?: () => void;
  onHide?: () => void;
}

export function Tooltip({
  content,
  position = 'auto',
  delay = 500,
  disabled = false,
  children,
  className,
  maxWidth = '20rem',
  showArrow = true,
  interactive = false,
  triggerOn = 'hover',
  onShow,
  onHide,
}: TooltipProps) {
  const [isVisible, setIsVisible] = useState(false);
  const [actualPosition, setActualPosition] = useState<TooltipPosition>(position);
  const [coords, setCoords] = useState({ x: 0, y: 0 });
  
  const triggerRef = useRef<HTMLDivElement>(null);
  const tooltipRef = useRef<HTMLDivElement>(null);
  const timeoutRef = useRef<NodeJS.Timeout>();
  const tooltipId = useRef(AriaUtils.generateId('tooltip'));

  // Calculate optimal position
  const calculatePosition = useCallback(() => {
    if (!triggerRef.current || !tooltipRef.current) return;

    const triggerRect = triggerRef.current.getBoundingClientRect();
    const tooltipRect = tooltipRef.current.getBoundingClientRect();
    const viewport = {
      width: window.innerWidth,
      height: window.innerHeight,
    };

    let finalPosition = position;
    let x = 0;
    let y = 0;
    const offset = 8;

    // Auto-detect best position if needed
    if (position === 'auto') {
      const spaceTop = triggerRect.top;
      const spaceBottom = viewport.height - triggerRect.bottom;
      const spaceLeft = triggerRect.left;
      const spaceRight = viewport.width - triggerRect.right;

      if (spaceBottom >= tooltipRect.height + offset) {
        finalPosition = 'bottom';
      } else if (spaceTop >= tooltipRect.height + offset) {
        finalPosition = 'top';
      } else if (spaceRight >= tooltipRect.width + offset) {
        finalPosition = 'right';
      } else if (spaceLeft >= tooltipRect.width + offset) {
        finalPosition = 'left';
      } else {
        finalPosition = 'bottom'; // Fallback
      }
    }

    // Calculate coordinates based on position
    switch (finalPosition) {
      case 'top':
      case 'top-start':
      case 'top-end':
        y = triggerRect.top - tooltipRect.height - offset;
        x = finalPosition === 'top-start' ? triggerRect.left :
            finalPosition === 'top-end' ? triggerRect.right - tooltipRect.width :
            triggerRect.left + (triggerRect.width - tooltipRect.width) / 2;
        break;

      case 'bottom':
      case 'bottom-start':
      case 'bottom-end':
        y = triggerRect.bottom + offset;
        x = finalPosition === 'bottom-start' ? triggerRect.left :
            finalPosition === 'bottom-end' ? triggerRect.right - tooltipRect.width :
            triggerRect.left + (triggerRect.width - tooltipRect.width) / 2;
        break;

      case 'left':
      case 'left-start':
      case 'left-end':
        x = triggerRect.left - tooltipRect.width - offset;
        y = finalPosition === 'left-start' ? triggerRect.top :
            finalPosition === 'left-end' ? triggerRect.bottom - tooltipRect.height :
            triggerRect.top + (triggerRect.height - tooltipRect.height) / 2;
        break;

      case 'right':
      case 'right-start':
      case 'right-end':
        x = triggerRect.right + offset;
        y = finalPosition === 'right-start' ? triggerRect.top :
            finalPosition === 'right-end' ? triggerRect.bottom - tooltipRect.height :
            triggerRect.top + (triggerRect.height - tooltipRect.height) / 2;
        break;
    }

    // Keep tooltip in viewport
    x = Math.max(8, Math.min(x, viewport.width - tooltipRect.width - 8));
    y = Math.max(8, Math.min(y, viewport.height - tooltipRect.height - 8));

    setActualPosition(finalPosition);
    setCoords({ x, y });
  }, [position]);

  // Show tooltip
  const show = useCallback(() => {
    if (disabled) return;

    if (timeoutRef.current) {
      clearTimeout(timeoutRef.current);
    }

    timeoutRef.current = setTimeout(() => {
      setIsVisible(true);
      onShow?.();
      trackEvent('tooltip_shown', { trigger: triggerOn });
    }, triggerOn === 'hover' ? delay : 0);
  }, [disabled, delay, triggerOn, onShow]);

  // Hide tooltip
  const hide = useCallback(() => {
    if (timeoutRef.current) {
      clearTimeout(timeoutRef.current);
    }

    setIsVisible(false);
    onHide?.();
  }, [onHide]);

  // Click outside handler for interactive tooltips
  useClickOutside(tooltipRef, hide, interactive && isVisible);

  // Keyboard shortcuts
  useKeyboardShortcuts({
    'escape': hide,
  }, isVisible && interactive);

  // Position calculation effect
  useEffect(() => {
    if (isVisible) {
      // Small delay to ensure DOM is ready
      setTimeout(calculatePosition, 0);
    }
  }, [isVisible, calculatePosition]);

  // Event handlers
  const handleMouseEnter = () => {
    if (triggerOn === 'hover') show();
  };

  const handleMouseLeave = () => {
    if (triggerOn === 'hover' && !interactive) hide();
  };

  const handleClick = () => {
    if (triggerOn === 'click') {
      if (isVisible) {
        hide();
      } else {
        show();
      }
    }
  };

  const handleFocus = () => {
    if (triggerOn === 'focus') show();
  };

  const handleBlur = () => {
    if (triggerOn === 'focus') hide();
  };

  // Cleanup
  useEffect(() => {
    return () => {
      if (timeoutRef.current) {
        clearTimeout(timeoutRef.current);
      }
    };
  }, []);

  const tooltip = isVisible ? (
    <div
      ref={tooltipRef}
      id={tooltipId.current}
      role="tooltip"
      className={cn(
        'fixed z-tooltip bg-gray-900 text-white text-sm rounded-lg shadow-lg animate-scale-in',
        'px-3 py-2 pointer-events-none max-w-xs',
        interactive && 'pointer-events-auto',
        className
      )}
      style={{
        left: coords.x,
        top: coords.y,
        maxWidth,
      }}
    >
      {content}
      
      {/* Arrow */}
      {showArrow && (
        <div
          className={cn(
            'absolute w-2 h-2 bg-gray-900 transform rotate-45',
            actualPosition?.startsWith('top') && 'bottom-0 translate-y-1/2',
            actualPosition?.startsWith('bottom') && 'top-0 -translate-y-1/2',
            actualPosition?.startsWith('left') && 'right-0 translate-x-1/2',
            actualPosition?.startsWith('right') && 'left-0 -translate-x-1/2',
          )}
        />
      )}
    </div>
  ) : null;

  return (
    <>
      <div
        ref={triggerRef}
        onMouseEnter={handleMouseEnter}
        onMouseLeave={handleMouseLeave}
        onClick={handleClick}
        onFocus={handleFocus}
        onBlur={handleBlur}
        aria-describedby={isVisible ? tooltipId.current : undefined}
        className="inline-block"
      >
        {children}
      </div>
      
      {tooltip && createPortal(tooltip, document.body)}
    </>
  );
}

// Help popover component for rich content
export interface HelpPopoverProps {
  content: HelpContent;
  trigger: ReactNode;
  position?: TooltipPosition;
  className?: string;
  onInteraction?: (action: string, data?: unknown) => void;
}

export function HelpPopover({
  content,
  trigger,
  position = 'bottom',
  className,
  onInteraction,
}: HelpPopoverProps) {
  const [isOpen, setIsOpen] = useState(false);
  const popoverRef = useRef<HTMLDivElement>(null);
  const triggerId = useRef(AriaUtils.generateId('help-trigger'));
  const popoverId = useRef(AriaUtils.generateId('help-popover'));

  useClickOutside(popoverRef, () => setIsOpen(false), isOpen);
  
  useKeyboardShortcuts({
    'escape': () => setIsOpen(false),
  }, isOpen);

  const handleToggle = () => {
    setIsOpen(!isOpen);
    onInteraction?.('toggle', { open: !isOpen });
    trackEvent('help_popover_toggled', { open: !isOpen, category: content.category });
  };

  const handleLinkClick = (link: HelpLink) => {
    onInteraction?.('link_click', link);
    trackEvent('help_link_clicked', { url: link.url, external: link.external });
  };

  return (
    <div className="relative inline-block">
      <button
        id={triggerId.current}
        onClick={handleToggle}
        aria-expanded={isOpen}
        aria-controls={popoverId.current}
        aria-label="Show help information"
        className="inline-flex items-center justify-center w-5 h-5 text-text-tertiary hover:text-text-primary transition-colors rounded-full hover:bg-background-secondary"
      >
        {trigger}
      </button>

      {isOpen && (
        <div
          ref={popoverRef}
          id={popoverId.current}
          role="dialog"
          aria-labelledby={content.title ? `${popoverId.current}-title` : undefined}
          className={cn(
            'absolute z-dropdown mt-2 w-80 bg-background-primary border border-border-primary rounded-lg shadow-dropdown',
            'animate-scale-in origin-top',
            position === 'top' && 'bottom-full mb-2 mt-0',
            position === 'left' && 'right-full mr-2 mt-0',
            position === 'right' && 'left-full ml-2 mt-0',
            className
          )}
        >
          <div className="p-4">
            {content.title && (
              <Heading level={4} id={`${popoverId.current}-title`} className="mb-2">
                {content.title}
              </Heading>
            )}
            
            <Body size="sm" className="mb-3">
              {content.description}
            </Body>

            {/* Examples */}
            {content.examples && content.examples.length > 0 && (
              <div className="mb-3">
                <Body size="sm" className="font-medium mb-2">Examples:</Body>
                <div className="space-y-1">
                  {content.examples.map((example, index) => (
                    <Code key={index} size="sm" className="block">
                      {example}
                    </Code>
                  ))}
                </div>
              </div>
            )}

            {/* Keyboard shortcuts */}
            {content.shortcuts && content.shortcuts.length > 0 && (
              <div className="mb-3">
                <Body size="sm" className="font-medium mb-2">Shortcuts:</Body>
                <div className="space-y-1">
                  {content.shortcuts.map((shortcut, index) => (
                    <div key={index} className="flex items-center justify-between">
                      <Body size="sm">{shortcut.description}</Body>
                      <div className="flex gap-1">
                        {shortcut.keys.map((key, keyIndex) => (
                          <kbd key={keyIndex} className="px-1.5 py-0.5 text-xs bg-gray-200 dark:bg-gray-700 rounded">
                            {key}
                          </kbd>
                        ))}
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            )}

            {/* Links */}
            {content.links && content.links.length > 0 && (
              <div className="mb-3">
                <Body size="sm" className="font-medium mb-2">Learn more:</Body>
                <div className="space-y-1">
                  {content.links.map((link, index) => (
                    <a
                      key={index}
                      href={link.url}
                      target={link.external ? '_blank' : undefined}
                      rel={link.external ? 'noopener noreferrer' : undefined}
                      onClick={() => handleLinkClick(link)}
                      className="block text-sm text-primary-600 dark:text-primary-400 hover:text-primary-700 dark:hover:text-primary-300 transition-colors"
                    >
                      {link.label}
                      {link.external && (
                        <svg className="inline w-3 h-3 ml-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14" />
                        </svg>
                      )}
                    </a>
                  ))}
                </div>
              </div>
            )}

            {/* Related topics */}
            {content.relatedTopics && content.relatedTopics.length > 0 && (
              <div>
                <Body size="sm" className="font-medium mb-2">Related:</Body>
                <div className="flex flex-wrap gap-1">
                  {content.relatedTopics.map((topic, index) => (
                    <button
                      key={index}
                      onClick={() => onInteraction?.('related_topic_click', { topic })}
                      className="px-2 py-1 text-xs bg-background-secondary hover:bg-background-tertiary rounded-full transition-colors"
                    >
                      {topic}
                    </button>
                  ))}
                </div>
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  );
}

// Help icon component
export function HelpIcon({ 
  size = 'sm',
  className 
}: { 
  size?: 'sm' | 'md' | 'lg';
  className?: string;
}) {
  const sizeClasses = {
    sm: 'w-4 h-4',
    md: 'w-5 h-5',
    lg: 'w-6 h-6',
  };

  return (
    <svg 
      className={cn(sizeClasses[size], className)} 
      fill="none" 
      stroke="currentColor" 
      viewBox="0 0 24 24"
      aria-hidden="true"
    >
      <path 
        strokeLinecap="round" 
        strokeLinejoin="round" 
        strokeWidth={2} 
        d="M8.228 9c.549-1.165 2.03-2 3.772-2 2.21 0 4 1.343 4 3 0 1.4-1.278 2.575-3.006 2.907-.542.104-.994.54-.994 1.093m0 3h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" 
      />
    </svg>
  );
}

// Context-aware help hook
export function useContextualHelp(context: string) {
  const [helpContent, setHelpContent] = useState<HelpContent | null>(null);
  const [isLoading, setIsLoading] = useState(false);

  const loadHelp = useCallback(async (contextKey: string) => {
    setIsLoading(true);
    try {
      // In a real app, this would fetch from a help API or static content
      // For now, we'll use a simple mapping
      const helpData = await getHelpContent(contextKey);
      setHelpContent(helpData);
    } catch (error) {
      console.error('Failed to load help content:', error);
    } finally {
      setIsLoading(false);
    }
  }, []);

  useEffect(() => {
    if (context) {
      loadHelp(context);
    }
  }, [context, loadHelp]);

  return {
    helpContent,
    isLoading,
    reload: () => loadHelp(context),
  };
}

// Help content provider (mock implementation)
async function getHelpContent(context: string): Promise<HelpContent | null> {
  // This would typically fetch from your help system/CMS
  const helpDatabase: Record<string, HelpContent> = {
    'cluster-selection': {
      title: 'Cluster Selection',
      description: 'Choose the Kubernetes cluster you want to manage. Only clusters you have access to will be displayed.',
      examples: [
        'production-us-east-1',
        'staging-europe-west1',
        'development-local',
      ],
      shortcuts: [
        { keys: ['Ctrl', 'K'], description: 'Open cluster selector' },
      ],
      links: [
        { label: 'Cluster Configuration Guide', url: '/docs/clusters', external: false },
        { label: 'RBAC Setup', url: '/docs/rbac', external: false },
      ],
      category: 'basic',
    },
    'kubectl-commands': {
      title: 'kubectl Commands',
      description: 'Execute kubectl commands safely with built-in validation and approval workflows.',
      examples: [
        'kubectl get pods',
        'kubectl describe deployment nginx',
        'kubectl logs -f pod-name',
      ],
      shortcuts: [
        { keys: ['Ctrl', 'Enter'], description: 'Execute command' },
        { keys: ['Ctrl', 'Shift', 'C'], description: 'Clear command' },
      ],
      links: [
        { label: 'kubectl Cheat Sheet', url: 'https://kubernetes.io/docs/reference/kubectl/cheatsheet/', external: true },
        { label: 'Command Safety Guide', url: '/docs/command-safety', external: false },
      ],
      category: 'basic',
    },
    'permissions': {
      title: 'Permissions & Access Control',
      description: 'Your access level determines which actions you can perform. Contact your administrator to request additional permissions.',
      links: [
        { label: 'Request Access', url: '/access-request', external: false },
        { label: 'Permission Levels', url: '/docs/permissions', external: false },
      ],
      category: 'security',
    },
  };

  return helpDatabase[context] || null;
}

// Inline help component for forms and inputs
export interface InlineHelpProps {
  content: string | HelpContent;
  position?: TooltipPosition;
  compact?: boolean;
  className?: string;
}

export function InlineHelp({ 
  content, 
  position = 'right',
  compact = false,
  className 
}: InlineHelpProps) {
  if (typeof content === 'string') {
    return (
      <Tooltip content={content} position={position} triggerOn="hover">
        <HelpIcon size={compact ? 'sm' : 'md'} className={cn('text-text-tertiary', className)} />
      </Tooltip>
    );
  }

  return (
    <HelpPopover 
      content={content} 
      position={position}
      className={className}
      trigger={<HelpIcon size={compact ? 'sm' : 'md'} />}
    />
  );
}