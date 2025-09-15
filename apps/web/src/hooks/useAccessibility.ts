import { useEffect, useRef, useState, useCallback } from 'react';
import {
  FocusManagement,
  KeyboardNavigation,
  ScreenReaderUtils,
  AriaAttributes,
  AccessibleComponentBuilder
} from '@/utils/accessibility';

// Hook for managing focus traps (modals, dropdowns)
export function useFocusTrap<T extends HTMLElement>() {
  const containerRef = useRef<T>(null);
  const previousFocusRef = useRef<HTMLElement | null>(null);
  const cleanupRef = useRef<(() => void) | null>(null);

  const enableFocusTrap = useCallback(() => {
    if (!containerRef.current) return;

    // Store previously focused element
    previousFocusRef.current = document.activeElement as HTMLElement;

    // Create focus trap
    cleanupRef.current = FocusManagement.createFocusTrap(containerRef.current);
  }, []);

  const disableFocusTrap = useCallback(() => {
    // Clean up focus trap
    if (cleanupRef.current) {
      cleanupRef.current();
      cleanupRef.current = null;
    }

    // Restore previous focus
    FocusManagement.restoreFocus(previousFocusRef.current);
    previousFocusRef.current = null;
  }, []);

  // Clean up on unmount
  useEffect(() => {
    return () => {
      disableFocusTrap();
    };
  }, [disableFocusTrap]);

  return {
    containerRef,
    enableFocusTrap,
    disableFocusTrap,
  };
}

// Hook for keyboard navigation in lists
export function useListKeyboardNavigation<T extends HTMLElement>(
  items: T[],
  onSelect?: (index: number, item: T) => void
) {
  const [currentIndex, setCurrentIndex] = useState(0);

  const handleKeyDown = useCallback((event: KeyboardEvent) => {
    const newIndex = KeyboardNavigation.handleListNavigation(
      event,
      items,
      currentIndex,
      (index) => onSelect?.(index, items[index])
    );
    setCurrentIndex(newIndex);
  }, [items, currentIndex, onSelect]);

  const setFocus = useCallback((index: number) => {
    if (items[index]) {
      items[index].focus();
      setCurrentIndex(index);
    }
  }, [items]);

  const moveFocus = useCallback((direction: 'up' | 'down' | 'first' | 'last') => {
    let newIndex = currentIndex;

    switch (direction) {
      case 'up':
        newIndex = currentIndex > 0 ? currentIndex - 1 : items.length - 1;
        break;
      case 'down':
        newIndex = currentIndex < items.length - 1 ? currentIndex + 1 : 0;
        break;
      case 'first':
        newIndex = 0;
        break;
      case 'last':
        newIndex = items.length - 1;
        break;
    }

    setFocus(newIndex);
  }, [currentIndex, items.length, setFocus]);

  return {
    currentIndex,
    handleKeyDown,
    setFocus,
    moveFocus,
  };
}

// Hook for managing screen reader announcements
export function useScreenReader() {
  const announce = useCallback((message: string, priority: 'polite' | 'assertive' = 'polite') => {
    ScreenReaderUtils.announce(message, priority);
  }, []);

  const announceError = useCallback((message: string) => {
    ScreenReaderUtils.announceError(message);
  }, []);

  const announceSuccess = useCallback((message: string) => {
    ScreenReaderUtils.announceSuccess(message);
  }, []);

  const announceLoading = useCallback((message?: string) => {
    ScreenReaderUtils.announceLoading(message);
  }, []);

  const announceCommand = useCallback((command: string, type: 'execution' | 'complete', result?: 'success' | 'error', details?: string) => {
    if (type === 'execution') {
      ScreenReaderUtils.announceCommandExecution(command);
    } else {
      ScreenReaderUtils.announceCommandComplete(result!, details);
    }
  }, []);

  return {
    announce,
    announceError,
    announceSuccess,
    announceLoading,
    announceCommand,
  };
}

// Hook for managing accessible dialogs/modals
export function useAccessibleDialog(isOpen: boolean) {
  const { containerRef, enableFocusTrap, disableFocusTrap } = useFocusTrap<HTMLDivElement>();
  const { announce } = useScreenReader();

  useEffect(() => {
    if (isOpen) {
      enableFocusTrap();
      announce('Dialog opened', 'polite');
    } else {
      disableFocusTrap();
    }
  }, [isOpen, enableFocusTrap, disableFocusTrap, announce]);

  const handleEscapeKey = useCallback((event: KeyboardEvent) => {
    if (event.key === 'Escape' && isOpen) {
      disableFocusTrap();
    }
  }, [isOpen, disableFocusTrap]);

  useEffect(() => {
    if (isOpen) {
      document.addEventListener('keydown', handleEscapeKey);
      return () => document.removeEventListener('keydown', handleEscapeKey);
    }
  }, [isOpen, handleEscapeKey]);

  const getDialogProps = useCallback((options: {
    label?: string;
    labelledBy?: string;
    describedBy?: string;
  } = {}) => ({
    ref: containerRef,
    ...AccessibleComponentBuilder.dialog({
      ...options,
      modal: true,
    }),
    onKeyDown: (event: React.KeyboardEvent) => {
      if (event.key === 'Escape') {
        disableFocusTrap();
      }
    },
  }), [containerRef, disableFocusTrap]);

  return {
    containerRef,
    getDialogProps,
  };
}

// Hook for accessible form validation
export function useAccessibleForm() {
  const [errors, setErrors] = useState<Record<string, string>>({});
  const { announceError } = useScreenReader();
  const errorRefs = useRef<Record<string, HTMLElement>>({});

  const setFieldError = useCallback((fieldName: string, errorMessage: string) => {
    setErrors(prev => ({
      ...prev,
      [fieldName]: errorMessage,
    }));
    announceError(`${fieldName}: ${errorMessage}`);

    // Focus on the field with error
    const fieldElement = document.querySelector(`[name="${fieldName}"]`) as HTMLElement;
    if (fieldElement) {
      fieldElement.focus();
    }
  }, [announceError]);

  const clearFieldError = useCallback((fieldName: string) => {
    setErrors(prev => {
      const newErrors = { ...prev };
      delete newErrors[fieldName];
      return newErrors;
    });
  }, []);

  const clearAllErrors = useCallback(() => {
    setErrors({});
  }, []);

  const getFieldProps = useCallback((fieldName: string, options: {
    label?: string;
    labelledBy?: string;
    required?: boolean;
    placeholder?: string;
  } = {}) => {
    const hasError = !!errors[fieldName];
    const errorId = `${fieldName}-error`;
    const describedBy = hasError ? errorId : undefined;

    return {
      ...AccessibleComponentBuilder.input({
        ...options,
        invalid: hasError,
        errorMessage: hasError ? errorId : undefined,
        describedBy,
      }),
      'aria-invalid': hasError,
      'aria-describedby': describedBy,
    };
  }, [errors]);

  const getErrorProps = useCallback((fieldName: string) => ({
    id: `${fieldName}-error`,
    role: 'alert' as const,
    'aria-live': 'assertive' as const,
    ref: (el: HTMLElement) => {
      if (el) {
        errorRefs.current[fieldName] = el;
      }
    },
  }), []);

  return {
    errors,
    setFieldError,
    clearFieldError,
    clearAllErrors,
    getFieldProps,
    getErrorProps,
    hasErrors: Object.keys(errors).length > 0,
  };
}

// Hook for accessible tabs
export function useAccessibleTabs(tabs: string[], defaultActiveTab = 0) {
  const [activeTab, setActiveTab] = useState(defaultActiveTab);
  const tabRefs = useRef<HTMLElement[]>([]);

  const handleTabKeyDown = useCallback((event: KeyboardEvent, index: number) => {
    const newIndex = KeyboardNavigation.handleHorizontalNavigation(event, tabRefs.current, index);
    if (newIndex !== index) {
      setActiveTab(newIndex);
    }
  }, []);

  const getTabProps = useCallback((index: number, options: {
    label?: string;
    controls?: string;
    disabled?: boolean;
  } = {}) => ({
    ref: (el: HTMLElement) => {
      if (el) {
        tabRefs.current[index] = el;
      }
    },
    ...AccessibleComponentBuilder.tab({
      ...options,
      selected: index === activeTab,
    }),
    tabIndex: index === activeTab ? 0 : -1,
    onClick: () => setActiveTab(index),
    onKeyDown: (event: React.KeyboardEvent) => handleTabKeyDown(event.nativeEvent, index),
  }), [activeTab, handleTabKeyDown]);

  const getTabPanelProps = useCallback((index: number, options: {
    label?: string;
    labelledBy?: string;
  } = {}) => ({
    ...AccessibleComponentBuilder.tabpanel(options),
    hidden: index !== activeTab,
    tabIndex: 0,
  }), [activeTab]);

  const getTabListProps = useCallback(() => ({
    role: 'tablist' as const,
    'aria-orientation': 'horizontal' as const,
  }), []);

  return {
    activeTab,
    setActiveTab,
    getTabProps,
    getTabPanelProps,
    getTabListProps,
  };
}

// Hook for accessible status updates (for real-time data)
export function useAccessibleStatus() {
  const { announce } = useScreenReader();
  const [status, setStatusState] = useState<string>('');

  const setStatus = useCallback((newStatus: string, priority: 'polite' | 'assertive' = 'polite') => {
    setStatusState(newStatus);
    announce(newStatus, priority);
  }, [announce]);

  const updateClusterStatus = useCallback((clusterName: string, status: 'healthy' | 'warning' | 'error' | 'offline') => {
    const statusMessages = {
      healthy: `${clusterName} cluster is healthy`,
      warning: `${clusterName} cluster has warnings`,
      error: `${clusterName} cluster has errors`,
      offline: `${clusterName} cluster is offline`,
    };

    const priority = status === 'error' || status === 'offline' ? 'assertive' : 'polite';
    setStatus(statusMessages[status], priority);
  }, [setStatus]);

  const updateCommandStatus = useCallback((command: string, status: 'executing' | 'completed' | 'failed') => {
    const statusMessages = {
      executing: `Executing command: ${command}`,
      completed: `Command completed: ${command}`,
      failed: `Command failed: ${command}`,
    };

    const priority = status === 'failed' ? 'assertive' : 'polite';
    setStatus(statusMessages[status], priority);
  }, [setStatus]);

  const getStatusProps = useCallback(() => ({
    ...AccessibleComponentBuilder.status({
      live: 'polite',
      atomic: true,
    }),
    'aria-live': 'polite' as const,
    'aria-atomic': true,
  }), []);

  return {
    status,
    setStatus,
    updateClusterStatus,
    updateCommandStatus,
    getStatusProps,
  };
}

// Hook for managing reduced motion preference
export function useReducedMotion() {
  const [prefersReducedMotion, setPrefersReducedMotion] = useState(false);

  useEffect(() => {
    if (typeof window === 'undefined') return;

    const mediaQuery = window.matchMedia('(prefers-reduced-motion: reduce)');
    setPrefersReducedMotion(mediaQuery.matches);

    const handleChange = (event: MediaQueryListEvent) => {
      setPrefersReducedMotion(event.matches);
    };

    if (mediaQuery.addEventListener) {
      mediaQuery.addEventListener('change', handleChange);
      return () => mediaQuery.removeEventListener('change', handleChange);
    } else {
      // Legacy browsers
      mediaQuery.addListener(handleChange);
      return () => mediaQuery.removeListener(handleChange);
    }
  }, []);

  return { prefersReducedMotion };
}

// Hook for accessible tooltips
export function useAccessibleTooltip() {
  const [isVisible, setIsVisible] = useState(false);
  const [tooltipId] = useState(() => `tooltip-${Math.random().toString(36).substr(2, 9)}`);

  const showTooltip = useCallback(() => setIsVisible(true), []);
  const hideTooltip = useCallback(() => setIsVisible(false), []);

  const getTriggerProps = useCallback((label?: string) => ({
    'aria-describedby': isVisible ? tooltipId : undefined,
    'aria-label': label,
    onMouseEnter: showTooltip,
    onMouseLeave: hideTooltip,
    onFocus: showTooltip,
    onBlur: hideTooltip,
  }), [isVisible, tooltipId, showTooltip, hideTooltip]);

  const getTooltipProps = useCallback(() => ({
    id: tooltipId,
    role: 'tooltip' as const,
    'aria-hidden': !isVisible,
  }), [tooltipId, isVisible]);

  return {
    isVisible,
    getTriggerProps,
    getTooltipProps,
    showTooltip,
    hideTooltip,
  };
}

