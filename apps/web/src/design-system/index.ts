// KubeChat Enterprise Design System
// Complete design system with tokens, components, accessibility, and keyboard navigation

// Design Tokens
export * from './tokens';
export * from './typography';

// Components
export * from './components';

// Accessibility Utilities
export * from './accessibility';
import { AccessibilityTesting } from './accessibility';

// Keyboard Navigation Hooks
export * from './hooks/useKeyboardNavigation';

// Re-export commonly used utilities
export { cn } from '../lib/utils';

// Design System Version
export const DESIGN_SYSTEM_VERSION = '1.0.0';

// Theme Provider Context (for future theme switching)
import React, { createContext, useContext, useEffect, useState } from 'react';

type Theme = 'light' | 'dark' | 'auto';

interface ThemeContextType {
  theme: Theme;
  actualTheme: 'light' | 'dark';
  setTheme: (theme: Theme) => void;
}

const ThemeContext = createContext<ThemeContextType | undefined>(undefined);

export function ThemeProvider({ 
  children, 
  defaultTheme = 'auto' 
}: { 
  children: React.ReactNode;
  defaultTheme?: Theme;
}) {
  const [theme, setTheme] = useState<Theme>(defaultTheme);
  const [actualTheme, setActualTheme] = useState<'light' | 'dark'>('light');

  useEffect(() => {
    // Get initial theme from localStorage or system preference
    const stored = localStorage.getItem('kubechat-theme') as Theme;
    if (stored && ['light', 'dark', 'auto'].includes(stored)) {
      setTheme(stored);
    }
  }, []);

  useEffect(() => {
    // Update localStorage when theme changes
    localStorage.setItem('kubechat-theme', theme);

    // Determine actual theme
    let resolvedTheme: 'light' | 'dark' = 'light';
    
    if (theme === 'auto') {
      resolvedTheme = window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
    } else {
      resolvedTheme = theme;
    }

    setActualTheme(resolvedTheme);

    // Update document class
    document.documentElement.classList.toggle('dark', resolvedTheme === 'dark');
  }, [theme]);

  useEffect(() => {
    // Listen for system theme changes when in auto mode
    if (theme !== 'auto') return;

    const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)');
    const handleChange = (e: MediaQueryListEvent) => {
      setActualTheme(e.matches ? 'dark' : 'light');
      document.documentElement.classList.toggle('dark', e.matches);
    };

    mediaQuery.addEventListener('change', handleChange);
    return () => mediaQuery.removeEventListener('change', handleChange);
  }, [theme]);

  return React.createElement(
    ThemeContext.Provider,
    { value: { theme, actualTheme, setTheme } },
    children
  );
}

export function useTheme() {
  const context = useContext(ThemeContext);
  if (!context) {
    throw new Error('useTheme must be used within a ThemeProvider');
  }
  return context;
}

// Global Design System Setup Component
export function DesignSystemProvider({ 
  children 
}: { 
  children: React.ReactNode 
}) {
  useEffect(() => {
    // Add design system CSS custom properties to document
    const root = document.documentElement;
    
    // Set up reduced motion handling
    const handleReducedMotion = () => {
      const prefersReduced = window.matchMedia('(prefers-reduced-motion: reduce)').matches;
      root.classList.toggle('reduce-motion', prefersReduced);
    };

    // Set up high contrast handling
    const handleHighContrast = () => {
      const prefersHigh = window.matchMedia('(prefers-contrast: high)').matches;
      root.classList.toggle('high-contrast', prefersHigh);
    };

    // Initial setup
    handleReducedMotion();
    handleHighContrast();

    // Listen for changes
    const reducedMotionQuery = window.matchMedia('(prefers-reduced-motion: reduce)');
    const highContrastQuery = window.matchMedia('(prefers-contrast: high)');

    reducedMotionQuery.addEventListener('change', handleReducedMotion);
    highContrastQuery.addEventListener('change', handleHighContrast);

    return () => {
      reducedMotionQuery.removeEventListener('change', handleReducedMotion);
      highContrastQuery.removeEventListener('change', handleHighContrast);
    };
  }, []);

  return React.createElement(ThemeProvider, null, children);
}

// Component prop types for better TypeScript support
export interface BaseComponentProps {
  className?: string;
  children?: React.ReactNode;
}

export interface InteractiveComponentProps extends BaseComponentProps {
  disabled?: boolean;
  loading?: boolean;
  'aria-label'?: string;
  'aria-describedby'?: string;
  'aria-labelledby'?: string;
}

// Common size types used across components
export type ComponentSize = 'sm' | 'md' | 'lg';

// Common variant types for semantic components
export type SemanticVariant = 'default' | 'success' | 'warning' | 'error' | 'info';

// Status types for enterprise applications
export type StatusType = 'running' | 'stopped' | 'error' | 'warning' | 'pending' | 'unknown';

// Safety levels for command operations
export type SafetyLevel = 'safe' | 'risky' | 'dangerous';

// Export design system constants
export const DESIGN_SYSTEM_CONSTANTS = {
  // Minimum touch target size (44px) for accessibility
  MIN_TOUCH_TARGET: '44px',
  
  // Maximum line length for readability (75 characters)
  MAX_LINE_LENGTH: '75ch',
  
  // Focus ring thickness
  FOCUS_RING_WIDTH: '2px',
  
  // Animation duration presets
  ANIMATION_DURATION: {
    fast: '150ms',
    normal: '300ms',
    slow: '500ms',
  },
  
  // Z-index scale
  Z_INDEX: {
    dropdown: 1000,
    tooltip: 1100,
    modal: 1200,
    notification: 1300,
    overlay: 1400,
  },
} as const;

// Export validation utilities
export const ValidationUtils = {
  // Validate color contrast
  validateContrast: (foreground: string, background: string, isLargeText = false) => {
    const { testColorContrast } = AccessibilityTesting;
    return testColorContrast(foreground, background, isLargeText);
  },
  
  // Validate component accessibility
  validateAccessibility: (element: HTMLElement) => {
    const { testFocusManagement, testAriaAttributes } = AccessibilityTesting;
    return {
      focus: testFocusManagement(element),
      aria: testAriaAttributes(element),
    };
  },
} as const;

// Export commonly used patterns
export const CommonPatterns = {
  // Loading states
  LoadingStates: {
    IDLE: 'idle',
    LOADING: 'loading',
    SUCCESS: 'success',
    ERROR: 'error',
  },
  
  // Command execution states
  CommandStates: {
    DRAFT: 'draft',
    PENDING_APPROVAL: 'pending_approval',
    EXECUTING: 'executing',
    SUCCESS: 'success',
    FAILED: 'failed',
    CANCELLED: 'cancelled',
  },
  
  // Compliance states
  ComplianceStates: {
    COMPLIANT: 'compliant',
    WARNING: 'warning',
    NON_COMPLIANT: 'non_compliant',
    UNKNOWN: 'unknown',
  },
} as const;