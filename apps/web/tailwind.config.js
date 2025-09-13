const { colors, spacing, typography, borderRadius, boxShadow, animation, breakpoints, zIndex } = require('./src/design-system/tokens');

/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    './src/pages/**/*.{js,ts,jsx,tsx,mdx}',
    './src/components/**/*.{js,ts,jsx,tsx,mdx}',
    './src/app/**/*.{js,ts,jsx,tsx,mdx}',
    './src/design-system/**/*.{js,ts,jsx,tsx,mdx}',
    // Include shared UI package
    '../../../packages/ui/src/**/*.{js,ts,jsx,tsx}',
  ],
  theme: {
    extend: {
      // Import colors from design tokens
      colors: {
        ...colors,
        // Theme-aware color utilities
        background: {
          primary: 'var(--color-background-primary)',
          secondary: 'var(--color-background-secondary)',
          tertiary: 'var(--color-background-tertiary)',
          inverse: 'var(--color-background-inverse)',
        },
        text: {
          primary: 'var(--color-text-primary)',
          secondary: 'var(--color-text-secondary)',
          tertiary: 'var(--color-text-tertiary)',
          inverse: 'var(--color-text-inverse)',
          disabled: 'var(--color-text-disabled)',
        },
        border: {
          primary: 'var(--color-border-primary)',
          secondary: 'var(--color-border-secondary)',
          focus: 'var(--color-border-focus)',
          error: 'var(--color-border-error)',
        },
      },
      
      // Import spacing from design tokens
      spacing: {
        ...spacing,
        // Component-specific spacing
        'button-sm-x': spacing[3],
        'button-sm-y': spacing[1],
        'button-md-x': spacing[4],
        'button-md-y': spacing[2],
        'button-lg-x': spacing[6],
        'button-lg-y': spacing[3],
        'input-sm-x': spacing[3],
        'input-sm-y': spacing[1],
        'input-md-x': spacing[4],
        'input-md-y': spacing[2],
        'input-lg-x': spacing[4],
        'input-lg-y': spacing[3],
        'card-sm': spacing[4],
        'card-md': spacing[6],
        'card-lg': spacing[8],
      },

      // Import typography from design tokens
      fontFamily: typography.fontFamily,
      fontSize: typography.fontSize,
      fontWeight: typography.fontWeight,
      letterSpacing: typography.letterSpacing,
      lineHeight: typography.lineHeight,

      // Import border radius from design tokens
      borderRadius: {
        ...borderRadius,
        'card': borderRadius.lg,
        'modal': borderRadius.xl,
      },

      // Import box shadows from design tokens
      boxShadow: {
        ...boxShadow,
        // Component-specific shadows
        'card': boxShadow.enterprise.card,
        'modal': boxShadow.enterprise.modal,
        'dropdown': boxShadow.enterprise.dropdown,
        'hover': boxShadow.enterprise.hover,
      },

      // Import breakpoints from design tokens
      screens: {
        ...breakpoints,
        'dashboard-compact': breakpoints.dashboard.compact,
        'dashboard-comfortable': breakpoints.dashboard.comfortable,
        'dashboard-wide': breakpoints.dashboard.wide,
      },

      // Import z-index from design tokens
      zIndex: {
        ...zIndex,
        'dropdown': zIndex.dropdown,
        'tooltip': zIndex.tooltip,
        'modal': zIndex.modal,
        'notification': zIndex.notification,
        'overlay': zIndex.overlay,
      },

      // Import animations from design tokens with DevOps-specific animations
      animation: {
        'fade-in': `fadeIn ${animation.duration.normal} ${animation.easing.out}`,
        'slide-up': `slideUp ${animation.duration.normal} ${animation.easing.enterprise.smooth}`,
        'slide-down': `slideDown ${animation.duration.normal} ${animation.easing.enterprise.smooth}`,
        'slide-left': `slideLeft ${animation.duration.normal} ${animation.easing.enterprise.smooth}`,
        'slide-right': `slideRight ${animation.duration.normal} ${animation.easing.enterprise.smooth}`,
        'scale-in': `scaleIn ${animation.duration.fast} ${animation.easing.enterprise.bounce}`,
        'pulse-subtle': `pulseSubtle 2s ${animation.easing.inOut} infinite`,
        'ping-slow': `ping 2s ${animation.easing.inOut} infinite`,
        'spin-slow': `spin 3s linear infinite`,
        'bounce-subtle': `bounceSubtle ${animation.duration.slow} ${animation.easing.enterprise.bounce}`,
        // Status indicator animations
        'status-success': `statusSuccess 1s ${animation.easing.out}`,
        'status-warning': `statusWarning 0.5s ${animation.easing.enterprise.sharp}`,
        'status-error': `statusError 0.3s ${animation.easing.enterprise.sharp}`,
        // Command execution animations
        'command-executing': `commandExecuting 1.5s ${animation.easing.inOut} infinite`,
        'terminal-cursor': `terminalCursor 1s step-end infinite`,
      },

      keyframes: {
        fadeIn: {
          '0%': { opacity: '0' },
          '100%': { opacity: '1' },
        },
        slideUp: {
          '0%': { transform: 'translateY(10px)', opacity: '0' },
          '100%': { transform: 'translateY(0)', opacity: '1' },
        },
        slideDown: {
          '0%': { transform: 'translateY(-10px)', opacity: '0' },
          '100%': { transform: 'translateY(0)', opacity: '1' },
        },
        slideLeft: {
          '0%': { transform: 'translateX(10px)', opacity: '0' },
          '100%': { transform: 'translateX(0)', opacity: '1' },
        },
        slideRight: {
          '0%': { transform: 'translateX(-10px)', opacity: '0' },
          '100%': { transform: 'translateX(0)', opacity: '1' },
        },
        scaleIn: {
          '0%': { transform: 'scale(0.95)', opacity: '0' },
          '100%': { transform: 'scale(1)', opacity: '1' },
        },
        pulseSubtle: {
          '0%, 100%': { opacity: '1' },
          '50%': { opacity: '0.8' },
        },
        bounceSubtle: {
          '0%, 100%': { transform: 'translateY(0)' },
          '50%': { transform: 'translateY(-2px)' },
        },
        statusSuccess: {
          '0%': { transform: 'scale(0.8)', backgroundColor: colors.success[200] },
          '50%': { transform: 'scale(1.1)', backgroundColor: colors.success[400] },
          '100%': { transform: 'scale(1)', backgroundColor: colors.success[500] },
        },
        statusWarning: {
          '0%': { backgroundColor: colors.warning[200] },
          '100%': { backgroundColor: colors.warning[500] },
        },
        statusError: {
          '0%': { backgroundColor: colors.error[200], transform: 'scale(1)' },
          '50%': { backgroundColor: colors.error[400], transform: 'scale(1.05)' },
          '100%': { backgroundColor: colors.error[500], transform: 'scale(1)' },
        },
        commandExecuting: {
          '0%, 100%': { opacity: '1', transform: 'scale(1)' },
          '50%': { opacity: '0.7', transform: 'scale(0.98)' },
        },
        terminalCursor: {
          '0%, 50%': { opacity: '1' },
          '51%, 100%': { opacity: '0' },
        },
      },

      // Enterprise-specific utilities
      backgroundImage: {
        'gradient-enterprise': `linear-gradient(135deg, ${colors.primary[500]} 0%, ${colors.secondary[600]} 100%)`,
        'gradient-kubernetes': `linear-gradient(135deg, ${colors.kubernetes.blue} 0%, ${colors.kubernetes.navy} 100%)`,
        'gradient-success': `linear-gradient(135deg, ${colors.success[400]} 0%, ${colors.success[600]} 100%)`,
        'gradient-warning': `linear-gradient(135deg, ${colors.warning[400]} 0%, ${colors.warning[600]} 100%)`,
        'gradient-error': `linear-gradient(135deg, ${colors.error[400]} 0%, ${colors.error[600]} 100%)`,
      },

      // DevOps tool specific gradients and patterns
      backdropBlur: {
        'xs': '2px',
        'modal': '8px',
      },

      // Professional spacing for enterprise layouts
      maxWidth: {
        'modal-sm': '24rem',
        'modal-md': '32rem',
        'modal-lg': '48rem',
        'modal-xl': '64rem',
        'dashboard': '1440px',
      },

      // Component heights following design tokens
      height: {
        'button-sm': '2rem',
        'button-md': '2.5rem',
        'button-lg': '3rem',
        'input-sm': '2rem',
        'input-md': '2.5rem',
        'input-lg': '3rem',
        'header': '4rem',
        'sidebar': 'calc(100vh - 4rem)',
      },
    },
  },
  plugins: [
    require('@tailwindcss/forms')({
      strategy: 'class', // Use class strategy for better control
    }),
    require('@tailwindcss/typography'),
    // Custom plugin for enterprise utilities
    function({ addUtilities, theme }) {
      const newUtilities = {
        // Status indicators
        '.status-indicator': {
          '@apply inline-flex items-center justify-center w-2 h-2 rounded-full': {},
        },
        '.status-success': {
          '@apply bg-enterprise-safe text-white': {},
        },
        '.status-warning': {
          '@apply bg-enterprise-risky text-white': {},
        },
        '.status-danger': {
          '@apply bg-enterprise-dangerous text-white': {},
        },
        '.status-unknown': {
          '@apply bg-enterprise-unknown text-white': {},
        },
        
        // Command safety levels
        '.command-safe': {
          '@apply border-l-4 border-enterprise-safe bg-success-50 dark:bg-success-900/20': {},
        },
        '.command-risky': {
          '@apply border-l-4 border-enterprise-risky bg-warning-50 dark:bg-warning-900/20': {},
        },
        '.command-dangerous': {
          '@apply border-l-4 border-enterprise-dangerous bg-error-50 dark:bg-error-900/20': {},
        },
        
        // Professional card styles
        '.card-enterprise': {
          '@apply bg-background-primary border border-border-primary rounded-card shadow-card': {},
        },
        '.card-hover': {
          '@apply transition-all duration-200 hover:shadow-hover hover:scale-[1.02]': {},
        },
        
        // Focus styles for accessibility
        '.focus-enterprise': {
          '@apply focus:outline-none focus:ring-2 focus:ring-border-focus focus:ring-offset-2 focus:ring-offset-background-primary': {},
        },
        
        // Terminal/code styles
        '.terminal': {
          '@apply font-mono text-sm bg-gray-900 text-gray-100 p-4 rounded-lg': {},
        },
        '.code-inline': {
          '@apply font-mono text-sm bg-gray-100 dark:bg-gray-800 px-1.5 py-0.5 rounded': {},
        },
        
        // Loading states
        '.loading-pulse': {
          '@apply animate-pulse bg-gray-200 dark:bg-gray-700': {},
        },
        '.loading-spinner': {
          '@apply animate-spin rounded-full border-2 border-gray-300 border-t-primary-500': {},
        },
      }
      
      addUtilities(newUtilities)
    },
  ],
  // Dark mode configuration
  darkMode: 'class',
};