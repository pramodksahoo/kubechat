// WCAG AA Accessibility Utilities for KubeChat Enterprise Design System
// Ensuring compliance with Web Content Accessibility Guidelines 2.1 Level AA

import React, { createContext, useContext, useEffect, ReactNode } from 'react';
import { colors } from './tokens';

// WCAG AA Color Contrast Requirements
export const WCAG_AA_CONTRAST = {
  NORMAL_TEXT: 4.5,
  LARGE_TEXT: 3.0,
  NON_TEXT: 3.0,
} as const;

// Convert hex color to RGB
export function hexToRgb(hex: string): { r: number; g: number; b: number } | null {
  const result = /^#?([a-f\d]{2})([a-f\d]{2})([a-f\d]{2})$/i.exec(hex);
  return result ? {
    r: parseInt(result[1], 16),
    g: parseInt(result[2], 16),
    b: parseInt(result[3], 16)
  } : null;
}

// Calculate relative luminance
export function getRelativeLuminance(rgb: { r: number; g: number; b: number }): number {
  const { r, g, b } = rgb;
  
  const sRGB = [r, g, b].map(c => {
    const value = c / 255;
    return value <= 0.03928 ? value / 12.92 : Math.pow((value + 0.055) / 1.055, 2.4);
  });
  
  return 0.2126 * sRGB[0] + 0.7152 * sRGB[1] + 0.0722 * sRGB[2];
}

// Calculate contrast ratio between two colors
export function getContrastRatio(color1: string, color2: string): number {
  const rgb1 = hexToRgb(color1);
  const rgb2 = hexToRgb(color2);
  
  if (!rgb1 || !rgb2) return 0;
  
  const lum1 = getRelativeLuminance(rgb1);
  const lum2 = getRelativeLuminance(rgb2);
  
  const brightest = Math.max(lum1, lum2);
  const darkest = Math.min(lum1, lum2);
  
  return (brightest + 0.05) / (darkest + 0.05);
}

// Check if color combination meets WCAG AA standards
export function meetsWCAGAA(
  foregroundColor: string, 
  backgroundColor: string, 
  isLargeText: boolean = false
): boolean {
  const contrastRatio = getContrastRatio(foregroundColor, backgroundColor);
  const requiredRatio = isLargeText ? WCAG_AA_CONTRAST.LARGE_TEXT : WCAG_AA_CONTRAST.NORMAL_TEXT;
  
  return contrastRatio >= requiredRatio;
}

// Get accessible color pairs from our design tokens
export const accessibleColorPairs = {
  light: {
    // Background and text combinations that meet WCAG AA
    primaryText: {
      background: colors.gray[50],
      foreground: colors.gray[900],
      contrast: getContrastRatio(colors.gray[900], colors.gray[50]),
    },
    secondaryText: {
      background: colors.gray[50],
      foreground: colors.gray[700],
      contrast: getContrastRatio(colors.gray[700], colors.gray[50]),
    },
    tertiaryText: {
      background: colors.gray[50],
      foreground: colors.gray[500],
      contrast: getContrastRatio(colors.gray[500], colors.gray[50]),
    },
    primaryButton: {
      background: colors.primary[600],
      foreground: colors.gray[50],
      contrast: getContrastRatio(colors.gray[50], colors.primary[600]),
    },
    successButton: {
      background: colors.success[600],
      foreground: colors.gray[50],
      contrast: getContrastRatio(colors.gray[50], colors.success[600]),
    },
    warningButton: {
      background: colors.warning[600],
      foreground: colors.gray[50],
      contrast: getContrastRatio(colors.gray[50], colors.warning[600]),
    },
    errorButton: {
      background: colors.error[600],
      foreground: colors.gray[50],
      contrast: getContrastRatio(colors.gray[50], colors.error[600]),
    },
  },
  dark: {
    primaryText: {
      background: colors.gray[900],
      foreground: colors.gray[50],
      contrast: getContrastRatio(colors.gray[50], colors.gray[900]),
    },
    secondaryText: {
      background: colors.gray[900],
      foreground: colors.gray[300],
      contrast: getContrastRatio(colors.gray[300], colors.gray[900]),
    },
    tertiaryText: {
      background: colors.gray[900],
      foreground: colors.gray[500],
      contrast: getContrastRatio(colors.gray[500], colors.gray[900]),
    },
    primaryButton: {
      background: colors.primary[500],
      foreground: colors.gray[900],
      contrast: getContrastRatio(colors.gray[900], colors.primary[500]),
    },
    successButton: {
      background: colors.success[500],
      foreground: colors.gray[900],
      contrast: getContrastRatio(colors.gray[900], colors.success[500]),
    },
    warningButton: {
      background: colors.warning[500],
      foreground: colors.gray[900],
      contrast: getContrastRatio(colors.gray[900], colors.warning[500]),
    },
    errorButton: {
      background: colors.error[500],
      foreground: colors.gray[900],
      contrast: getContrastRatio(colors.gray[900], colors.error[500]),
    },
  },
} as const;

// Focus management utilities
export class FocusManager {
  private static focusableSelectors = [
    'a[href]',
    'area[href]',
    'input:not([disabled]):not([type="hidden"])',
    'select:not([disabled])',
    'textarea:not([disabled])',
    'button:not([disabled])',
    'iframe',
    'object',
    'embed',
    '[contenteditable]',
    '[tabindex]:not([tabindex^="-"])',
  ].join(',');

  // Get all focusable elements within a container
  static getFocusableElements(container: HTMLElement): HTMLElement[] {
    return Array.from(container.querySelectorAll(this.focusableSelectors))
      .filter(element => {
        return element instanceof HTMLElement && this.isVisible(element);
      }) as HTMLElement[];
  }

  // Check if element is visible
  static isVisible(element: HTMLElement): boolean {
    if (element.offsetParent === null) return false;
    
    const style = window.getComputedStyle(element);
    return style.visibility !== 'hidden' && style.display !== 'none';
  }

  // Create focus trap for modals
  static createFocusTrap(container: HTMLElement): () => void {
    const focusableElements = this.getFocusableElements(container);
    const firstFocusable = focusableElements[0];
    const lastFocusable = focusableElements[focusableElements.length - 1];

    const handleTabKey = (e: KeyboardEvent) => {
      if (e.key !== 'Tab') return;

      if (e.shiftKey) {
        if (document.activeElement === firstFocusable) {
          e.preventDefault();
          lastFocusable?.focus();
        }
      } else {
        if (document.activeElement === lastFocusable) {
          e.preventDefault();
          firstFocusable?.focus();
        }
      }
    };

    container.addEventListener('keydown', handleTabKey);
    
    // Focus first element
    firstFocusable?.focus();

    // Return cleanup function
    return () => {
      container.removeEventListener('keydown', handleTabKey);
    };
  }

  // Restore focus to previously focused element
  static restoreFocus(previouslyFocusedElement: HTMLElement | null) {
    if (previouslyFocusedElement && this.isVisible(previouslyFocusedElement)) {
      previouslyFocusedElement.focus();
    }
  }
}

// Screen reader utilities
export class ScreenReaderUtils {
  // Announce message to screen readers
  static announce(message: string, priority: 'polite' | 'assertive' = 'polite') {
    const announcement = document.createElement('div');
    announcement.setAttribute('aria-live', priority);
    announcement.setAttribute('aria-atomic', 'true');
    announcement.className = 'sr-only';
    announcement.textContent = message;
    
    document.body.appendChild(announcement);
    
    // Remove after announcement
    setTimeout(() => {
      document.body.removeChild(announcement);
    }, 1000);
  }

  // Create live region for dynamic content updates
  static createLiveRegion(id: string, priority: 'polite' | 'assertive' = 'polite'): HTMLElement {
    let liveRegion = document.getElementById(id);
    
    if (!liveRegion) {
      liveRegion = document.createElement('div');
      liveRegion.id = id;
      liveRegion.setAttribute('aria-live', priority);
      liveRegion.setAttribute('aria-atomic', 'true');
      liveRegion.className = 'sr-only';
      document.body.appendChild(liveRegion);
    }
    
    return liveRegion;
  }

  // Update live region content
  static updateLiveRegion(id: string, message: string) {
    const liveRegion = document.getElementById(id);
    if (liveRegion) {
      liveRegion.textContent = message;
    }
  }
}

// Keyboard navigation utilities
export class KeyboardNavigation {
  // Handle arrow key navigation for lists and grids
  static handleArrowNavigation(
    event: KeyboardEvent,
    items: HTMLElement[],
    currentIndex: number,
    orientation: 'horizontal' | 'vertical' | 'grid' = 'vertical',
    gridColumns?: number
  ): number {
    const { key } = event;
    let newIndex = currentIndex;

    switch (key) {
      case 'ArrowUp':
        if (orientation === 'vertical') {
          newIndex = currentIndex > 0 ? currentIndex - 1 : items.length - 1;
        } else if (orientation === 'grid' && gridColumns) {
          newIndex = currentIndex - gridColumns;
          if (newIndex < 0) newIndex = currentIndex;
        }
        break;
        
      case 'ArrowDown':
        if (orientation === 'vertical') {
          newIndex = currentIndex < items.length - 1 ? currentIndex + 1 : 0;
        } else if (orientation === 'grid' && gridColumns) {
          newIndex = currentIndex + gridColumns;
          if (newIndex >= items.length) newIndex = currentIndex;
        }
        break;
        
      case 'ArrowLeft':
        if (orientation === 'horizontal') {
          newIndex = currentIndex > 0 ? currentIndex - 1 : items.length - 1;
        } else if (orientation === 'grid') {
          newIndex = currentIndex > 0 ? currentIndex - 1 : currentIndex;
        }
        break;
        
      case 'ArrowRight':
        if (orientation === 'horizontal') {
          newIndex = currentIndex < items.length - 1 ? currentIndex + 1 : 0;
        } else if (orientation === 'grid') {
          newIndex = currentIndex < items.length - 1 ? currentIndex + 1 : currentIndex;
        }
        break;
        
      case 'Home':
        newIndex = 0;
        break;
        
      case 'End':
        newIndex = items.length - 1;
        break;
        
      default:
        return currentIndex;
    }

    if (newIndex !== currentIndex) {
      event.preventDefault();
      items[newIndex]?.focus();
    }

    return newIndex;
  }

  // Handle typeahead search in lists
  static handleTypeahead(
    event: KeyboardEvent,
    items: HTMLElement[],
    currentIndex: number,
    getItemText: (item: HTMLElement) => string
  ): number {
    const { key } = event;
    
    if (key.length !== 1 || event.altKey || event.ctrlKey || event.metaKey) {
      return currentIndex;
    }

    const searchText = key.toLowerCase();
    const startIndex = (currentIndex + 1) % items.length;
    
    // Search from current position forward
    for (let i = 0; i < items.length; i++) {
      const index = (startIndex + i) % items.length;
      const item = items[index];
      const itemText = getItemText(item).toLowerCase();
      
      if (itemText.startsWith(searchText)) {
        event.preventDefault();
        item.focus();
        return index;
      }
    }

    return currentIndex;
  }
}

// ARIA utilities
export class AriaUtils {
  // Generate unique IDs for aria relationships
  static generateId(prefix: string = 'aria'): string {
    return `${prefix}-${Math.random().toString(36).substr(2, 9)}`;
  }

  // Set up describedby relationship
  static setDescribedBy(element: HTMLElement, describerId: string) {
    const currentDescribedBy = element.getAttribute('aria-describedby');
    const newDescribedBy = currentDescribedBy 
      ? `${currentDescribedBy} ${describerId}`
      : describerId;
    
    element.setAttribute('aria-describedby', newDescribedBy);
  }

  // Remove from describedby relationship
  static removeDescribedBy(element: HTMLElement, describerId: string) {
    const currentDescribedBy = element.getAttribute('aria-describedby');
    if (!currentDescribedBy) return;
    
    const newDescribedBy = currentDescribedBy
      .split(' ')
      .filter(id => id !== describerId)
      .join(' ');
    
    if (newDescribedBy) {
      element.setAttribute('aria-describedby', newDescribedBy);
    } else {
      element.removeAttribute('aria-describedby');
    }
  }

  // Set up labelledby relationship
  static setLabelledBy(element: HTMLElement, labelId: string) {
    element.setAttribute('aria-labelledby', labelId);
  }

  // Toggle aria-expanded for collapsible content
  static toggleExpanded(trigger: HTMLElement, expanded?: boolean) {
    const currentExpanded = trigger.getAttribute('aria-expanded') === 'true';
    const newExpanded = expanded !== undefined ? expanded : !currentExpanded;
    trigger.setAttribute('aria-expanded', newExpanded.toString());
    return newExpanded;
  }
}

// Reduced motion utilities
export class ReducedMotionUtils {
  // Check if user prefers reduced motion
  static prefersReducedMotion(): boolean {
    return window.matchMedia('(prefers-reduced-motion: reduce)').matches;
  }

  // Apply reduced motion class conditionally
  static applyReducedMotion(element: HTMLElement, className: string = 'reduce-motion') {
    if (this.prefersReducedMotion()) {
      element.classList.add(className);
    } else {
      element.classList.remove(className);
    }
  }

  // Create media query listener for reduced motion changes
  static watchReducedMotion(callback: (prefersReduced: boolean) => void): () => void {
    const mediaQuery = window.matchMedia('(prefers-reduced-motion: reduce)');
    const handleChange = (e: MediaQueryListEvent) => callback(e.matches);
    
    mediaQuery.addEventListener('change', handleChange);
    
    // Call initially
    callback(mediaQuery.matches);
    
    // Return cleanup function
    return () => mediaQuery.removeEventListener('change', handleChange);
  }
}

// High contrast mode utilities
export class HighContrastUtils {
  // Check if high contrast mode is enabled
  static prefersHighContrast(): boolean {
    return window.matchMedia('(prefers-contrast: high)').matches;
  }

  // Apply high contrast styles
  static applyHighContrast(element: HTMLElement, className: string = 'high-contrast') {
    if (this.prefersHighContrast()) {
      element.classList.add(className);
    } else {
      element.classList.remove(className);
    }
  }

  // Watch for high contrast preference changes
  static watchHighContrast(callback: (prefersHigh: boolean) => void): () => void {
    const mediaQuery = window.matchMedia('(prefers-contrast: high)');
    const handleChange = (e: MediaQueryListEvent) => callback(e.matches);
    
    mediaQuery.addEventListener('change', handleChange);
    callback(mediaQuery.matches);
    
    return () => mediaQuery.removeEventListener('change', handleChange);
  }
}

// Export accessibility testing utilities
export const AccessibilityTesting = {
  // Test color contrast ratios
  testColorContrast: (foreground: string, background: string, isLargeText = false) => ({
    ratio: getContrastRatio(foreground, background),
    meetsAA: meetsWCAGAA(foreground, background, isLargeText),
    meetsAAA: getContrastRatio(foreground, background) >= (isLargeText ? 4.5 : 7),
  }),

  // Test focus management
  testFocusManagement: (container: HTMLElement) => ({
    focusableElements: FocusManager.getFocusableElements(container),
    hasFocusableElements: FocusManager.getFocusableElements(container).length > 0,
    firstFocusable: FocusManager.getFocusableElements(container)[0],
    lastFocusable: FocusManager.getFocusableElements(container).slice(-1)[0],
  }),

  // Test ARIA attributes
  testAriaAttributes: (element: HTMLElement) => ({
    hasLabel: !!(element.getAttribute('aria-label') || element.getAttribute('aria-labelledby')),
    hasDescription: !!element.getAttribute('aria-describedby'),
    hasRole: !!element.getAttribute('role'),
    isExpandable: element.hasAttribute('aria-expanded'),
    isPressed: element.hasAttribute('aria-pressed'),
    isSelected: element.hasAttribute('aria-selected'),
  }),
} as const;

// React AccessibilityProvider Component
interface AccessibilityContextType {
  prefersReducedMotion: boolean;
  prefersHighContrast: boolean;
  announceToScreenReader: (message: string, priority?: 'polite' | 'assertive') => void;
}

const AccessibilityContext = createContext<AccessibilityContextType | null>(null);

export function AccessibilityProvider({ children }: { children: ReactNode }) {
  const [prefersReducedMotion, setPrefersReducedMotion] = React.useState(false);
  const [prefersHighContrast, setPrefersHighContrast] = React.useState(false);

  useEffect(() => {
    // Initialize values
    setPrefersReducedMotion(ReducedMotionUtils.prefersReducedMotion());
    setPrefersHighContrast(HighContrastUtils.prefersHighContrast());

    // Set up watchers
    const cleanupReducedMotion = ReducedMotionUtils.watchReducedMotion(setPrefersReducedMotion);
    const cleanupHighContrast = HighContrastUtils.watchHighContrast(setPrefersHighContrast);

    // Apply initial classes to document
    if (prefersReducedMotion) {
      document.documentElement.classList.add('reduce-motion');
    }
    if (prefersHighContrast) {
      document.documentElement.classList.add('high-contrast');
    }

    // Create global live regions
    ScreenReaderUtils.createLiveRegion('global-announcements', 'polite');
    ScreenReaderUtils.createLiveRegion('global-alerts', 'assertive');

    return () => {
      cleanupReducedMotion();
      cleanupHighContrast();
    };
  }, [prefersReducedMotion, prefersHighContrast]);

  const announceToScreenReader = (message: string, priority: 'polite' | 'assertive' = 'polite') => {
    ScreenReaderUtils.announce(message, priority);
  };

  const contextValue: AccessibilityContextType = {
    prefersReducedMotion,
    prefersHighContrast,
    announceToScreenReader,
  };

  return React.createElement(
    AccessibilityContext.Provider,
    { value: contextValue },
    children
  );
}

export function useAccessibility(): AccessibilityContextType {
  const context = useContext(AccessibilityContext);
  if (!context) {
    throw new Error('useAccessibility must be used within an AccessibilityProvider');
  }
  return context;
}