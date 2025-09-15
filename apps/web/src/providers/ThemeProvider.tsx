import React, { createContext, useContext, useEffect, useState, ReactNode } from 'react';

// Theme types
export type Theme = 'light' | 'dark' | 'system';
export type ResolvedTheme = 'light' | 'dark';

export interface ThemeConfig {
  theme: Theme;
  resolvedTheme: ResolvedTheme;
  setTheme: (theme: Theme) => void;
  toggleTheme: () => void;
  systemTheme: ResolvedTheme;
}

// Enterprise theme presets
export interface EnterpriseThemePreset {
  id: string;
  name: string;
  description: string;
  colors: {
    primary: string;
    secondary: string;
    accent: string;
  };
  darkMode: boolean;
}

export const enterpriseThemePresets: EnterpriseThemePreset[] = [
  {
    id: 'kubernetes',
    name: 'Kubernetes Blue',
    description: 'Official Kubernetes brand colors with professional styling',
    colors: {
      primary: '#326ce5',
      secondary: '#1a73e8',
      accent: '#4285f4'
    },
    darkMode: false
  },
  {
    id: 'enterprise-dark',
    name: 'Enterprise Dark',
    description: 'Professional dark theme optimized for extended DevOps work',
    colors: {
      primary: '#3b82f6',
      secondary: '#1e40af',
      accent: '#60a5fa'
    },
    darkMode: true
  },
  {
    id: 'devops-light',
    name: 'DevOps Light',
    description: 'Clean light theme with high contrast for better readability',
    colors: {
      primary: '#2563eb',
      secondary: '#1d4ed8',
      accent: '#3b82f6'
    },
    darkMode: false
  },
  {
    id: 'security-focused',
    name: 'Security Focused',
    description: 'High-contrast theme optimized for security monitoring',
    colors: {
      primary: '#dc2626',
      secondary: '#991b1b',
      accent: '#f87171'
    },
    darkMode: true
  }
];

// Theme context
const ThemeContext = createContext<ThemeConfig | undefined>(undefined);

// Storage keys
const STORAGE_KEY = 'kubechat-theme';
const PRESET_STORAGE_KEY = 'kubechat-theme-preset';

// Theme provider props
interface ThemeProviderProps {
  children: ReactNode;
  defaultTheme?: Theme;
  storageKey?: string;
  enableSystem?: boolean;
  disableTransitionOnChange?: boolean;
}

export function ThemeProvider({
  children,
  defaultTheme = 'system',
  storageKey = STORAGE_KEY,
  enableSystem = true,
  disableTransitionOnChange = false,
}: ThemeProviderProps) {
  const [theme, setThemeState] = useState<Theme>(defaultTheme);
  const [resolvedTheme, setResolvedTheme] = useState<ResolvedTheme>('light');
  const [systemTheme, setSystemTheme] = useState<ResolvedTheme>('light');

  // Get system theme preference
  const getSystemTheme = (): ResolvedTheme => {
    if (typeof window === 'undefined') return 'light';
    return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
  };

  // Apply theme to document
  const applyTheme = (newTheme: ResolvedTheme) => {
    if (typeof document === 'undefined') return;

    const root = document.documentElement;

    // Temporarily disable transitions to prevent flash
    if (disableTransitionOnChange) {
      const css = document.createElement('style');
      css.appendChild(
        document.createTextNode(
          '*, *::before, *::after { transition: none !important; animation-duration: 0.01ms !important; animation-delay: 0.01ms !important; }'
        )
      );
      document.head.appendChild(css);

      // Re-enable transitions after a short delay
      setTimeout(() => {
        document.head.removeChild(css);
      }, 1);
    }

    // Remove old theme classes
    root.classList.remove('light', 'dark');

    // Add new theme class
    root.classList.add(newTheme);

    // Set data attributes for CSS custom properties
    root.setAttribute('data-theme', newTheme);

    // Update meta theme-color for mobile browsers
    const metaThemeColor = document.querySelector('meta[name="theme-color"]');
    if (metaThemeColor) {
      metaThemeColor.setAttribute(
        'content',
        newTheme === 'dark' ? '#0f172a' : '#ffffff'
      );
    }
  };

  // Set theme and persist to storage
  const setTheme = (newTheme: Theme) => {
    setThemeState(newTheme);

    if (typeof window !== 'undefined') {
      localStorage.setItem(storageKey, newTheme);
    }

    // Resolve theme
    let resolved: ResolvedTheme;
    if (newTheme === 'system') {
      resolved = systemTheme;
    } else {
      resolved = newTheme;
    }

    setResolvedTheme(resolved);
    applyTheme(resolved);
  };

  // Toggle between light and dark (ignores system)
  const toggleTheme = () => {
    const newTheme = resolvedTheme === 'light' ? 'dark' : 'light';
    setTheme(newTheme);
  };

  // Initialize theme on mount
  useEffect(() => {
    if (typeof window === 'undefined') return;

    // Get system theme
    const currentSystemTheme = getSystemTheme();
    setSystemTheme(currentSystemTheme);

    // Get stored theme or use default
    const storedTheme = localStorage.getItem(storageKey) as Theme;
    const initialTheme = storedTheme || defaultTheme;

    // Resolve initial theme
    let resolved: ResolvedTheme;
    if (initialTheme === 'system') {
      resolved = currentSystemTheme;
    } else {
      resolved = initialTheme;
    }

    setThemeState(initialTheme);
    setResolvedTheme(resolved);
    applyTheme(resolved);

    // Listen for system theme changes
    if (enableSystem) {
      const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)');

      const handleSystemThemeChange = (e: MediaQueryListEvent) => {
        const newSystemTheme = e.matches ? 'dark' : 'light';
        setSystemTheme(newSystemTheme);

        // Update resolved theme if currently using system
        if (theme === 'system') {
          setResolvedTheme(newSystemTheme);
          applyTheme(newSystemTheme);
        }
      };

      // Modern browsers
      if (mediaQuery.addEventListener) {
        mediaQuery.addEventListener('change', handleSystemThemeChange);
        return () => mediaQuery.removeEventListener('change', handleSystemThemeChange);
      }
      // Legacy browsers
      else {
        mediaQuery.addListener(handleSystemThemeChange);
        return () => mediaQuery.removeListener(handleSystemThemeChange);
      }
    }
  }, [defaultTheme, enableSystem, storageKey, theme]);

  // Update resolved theme when theme or system theme changes
  useEffect(() => {
    const resolved = theme === 'system' ? systemTheme : theme;
    if (resolved !== resolvedTheme) {
      setResolvedTheme(resolved);
      applyTheme(resolved);
    }
  }, [theme, systemTheme, resolvedTheme]);

  const value: ThemeConfig = {
    theme,
    resolvedTheme,
    setTheme,
    toggleTheme,
    systemTheme,
  };

  return (
    <ThemeContext.Provider value={value}>
      {children}
    </ThemeContext.Provider>
  );
}

// Hook to use theme
export function useTheme() {
  const context = useContext(ThemeContext);

  if (context === undefined) {
    throw new Error('useTheme must be used within a ThemeProvider');
  }

  return context;
}

// Hook to manage theme presets
export function useThemePresets() {
  const [activePreset, setActivePresetState] = useState<string>('kubernetes');

  const setActivePreset = (presetId: string) => {
    const preset = enterpriseThemePresets.find(p => p.id === presetId);
    if (!preset) return;

    setActivePresetState(presetId);

    // Store preset preference
    if (typeof window !== 'undefined') {
      localStorage.setItem(PRESET_STORAGE_KEY, presetId);
    }

    // Apply preset colors to CSS custom properties
    if (typeof document !== 'undefined') {
      const root = document.documentElement;
      root.style.setProperty('--color-primary-500', preset.colors.primary);
      root.style.setProperty('--color-primary-600', preset.colors.secondary);
      root.style.setProperty('--color-primary-400', preset.colors.accent);
    }
  };

  // Initialize preset on mount
  useEffect(() => {
    if (typeof window === 'undefined') return;

    const storedPreset = localStorage.getItem(PRESET_STORAGE_KEY);
    if (storedPreset && enterpriseThemePresets.find(p => p.id === storedPreset)) {
      setActivePreset(storedPreset);
    }
  }, []);

  const getActivePreset = () => {
    return enterpriseThemePresets.find(p => p.id === activePreset);
  };

  return {
    presets: enterpriseThemePresets,
    activePreset,
    setActivePreset,
    getActivePreset,
  };
}

// Theme selector component
export function ThemeSelector({
  className = ''
}: {
  className?: string;
}) {
  const { theme, setTheme } = useTheme();

  const themes: { value: Theme; label: string; description: string }[] = [
    { value: 'light', label: 'Light', description: 'Clean light theme for daytime work' },
    { value: 'dark', label: 'Dark', description: 'Easy on the eyes for extended sessions' },
    { value: 'system', label: 'System', description: 'Follow system preference' },
  ];

  return (
    <div className={`space-y-2 ${className}`}>
      <label className="block text-sm font-medium text-gray-700 dark:text-gray-300">
        Theme Preference
      </label>

      <div className="grid grid-cols-3 gap-2">
        {themes.map((themeOption) => (
          <button
            key={themeOption.value}
            onClick={() => setTheme(themeOption.value)}
            className={`p-3 rounded-lg border text-center transition-all ${
              theme === themeOption.value
                ? 'border-blue-500 bg-blue-50 dark:bg-blue-900/20 text-blue-700 dark:text-blue-300'
                : 'border-gray-200 dark:border-gray-700 hover:border-gray-300 dark:hover:border-gray-600'
            }`}
            title={themeOption.description}
          >
            <div className="text-sm font-medium">{themeOption.label}</div>
          </button>
        ))}
      </div>
    </div>
  );
}

// Theme preset selector component
export function ThemePresetSelector({
  className = ''
}: {
  className?: string;
}) {
  const { presets, activePreset, setActivePreset } = useThemePresets();

  return (
    <div className={`space-y-3 ${className}`}>
      <label className="block text-sm font-medium text-gray-700 dark:text-gray-300">
        Color Scheme
      </label>

      <div className="grid grid-cols-1 gap-3">
        {presets.map((preset) => (
          <button
            key={preset.id}
            onClick={() => setActivePreset(preset.id)}
            className={`p-4 rounded-lg border text-left transition-all ${
              activePreset === preset.id
                ? 'border-blue-500 bg-blue-50 dark:bg-blue-900/20'
                : 'border-gray-200 dark:border-gray-700 hover:border-gray-300 dark:hover:border-gray-600'
            }`}
          >
            <div className="flex items-center justify-between">
              <div className="flex-1">
                <div className="font-medium text-gray-900 dark:text-white">
                  {preset.name}
                </div>
                <div className="text-sm text-gray-600 dark:text-gray-400 mt-1">
                  {preset.description}
                </div>
              </div>

              <div className="flex space-x-1 ml-3">
                {Object.values(preset.colors).map((color, index) => (
                  <div
                    key={index}
                    className="w-4 h-4 rounded-full border border-gray-300 dark:border-gray-600"
                    style={{ backgroundColor: color }}
                  />
                ))}
              </div>
            </div>
          </button>
        ))}
      </div>
    </div>
  );
}

export default ThemeProvider;