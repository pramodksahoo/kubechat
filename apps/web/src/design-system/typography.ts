// Typography System for KubeChat Enterprise UI
// Semantic typography scales with proper hierarchy and responsive behavior

import { typography as baseTypography } from './tokens';

// Semantic text styles with context-specific variants
export const textStyles = {
  // Display text for hero sections and major headings
  display: {
    '2xl': {
      fontSize: baseTypography.fontSize['6xl'],
      fontWeight: baseTypography.fontWeight.extrabold,
      lineHeight: baseTypography.lineHeight.none,
      letterSpacing: baseTypography.letterSpacing.tight,
      fontFamily: baseTypography.fontFamily.display,
    },
    xl: {
      fontSize: baseTypography.fontSize['5xl'],
      fontWeight: baseTypography.fontWeight.bold,
      lineHeight: baseTypography.lineHeight.tight,
      letterSpacing: baseTypography.letterSpacing.tight,
      fontFamily: baseTypography.fontFamily.display,
    },
    lg: {
      fontSize: baseTypography.fontSize['4xl'],
      fontWeight: baseTypography.fontWeight.bold,
      lineHeight: baseTypography.lineHeight.tight,
      letterSpacing: baseTypography.letterSpacing.normal,
      fontFamily: baseTypography.fontFamily.display,
    },
    md: {
      fontSize: baseTypography.fontSize['3xl'],
      fontWeight: baseTypography.fontWeight.semibold,
      lineHeight: baseTypography.lineHeight.snug,
      letterSpacing: baseTypography.letterSpacing.normal,
      fontFamily: baseTypography.fontFamily.display,
    },
    sm: {
      fontSize: baseTypography.fontSize['2xl'],
      fontWeight: baseTypography.fontWeight.semibold,
      lineHeight: baseTypography.lineHeight.snug,
      letterSpacing: baseTypography.letterSpacing.normal,
      fontFamily: baseTypography.fontFamily.display,
    },
  },

  // Headings for sections and components
  heading: {
    h1: {
      fontSize: baseTypography.fontSize['3xl'],
      fontWeight: baseTypography.fontWeight.bold,
      lineHeight: baseTypography.lineHeight.tight,
      letterSpacing: baseTypography.letterSpacing.tight,
      fontFamily: baseTypography.fontFamily.sans,
    },
    h2: {
      fontSize: baseTypography.fontSize['2xl'],
      fontWeight: baseTypography.fontWeight.semibold,
      lineHeight: baseTypography.lineHeight.snug,
      letterSpacing: baseTypography.letterSpacing.normal,
      fontFamily: baseTypography.fontFamily.sans,
    },
    h3: {
      fontSize: baseTypography.fontSize.xl,
      fontWeight: baseTypography.fontWeight.semibold,
      lineHeight: baseTypography.lineHeight.snug,
      letterSpacing: baseTypography.letterSpacing.normal,
      fontFamily: baseTypography.fontFamily.sans,
    },
    h4: {
      fontSize: baseTypography.fontSize.lg,
      fontWeight: baseTypography.fontWeight.medium,
      lineHeight: baseTypography.lineHeight.normal,
      letterSpacing: baseTypography.letterSpacing.normal,
      fontFamily: baseTypography.fontFamily.sans,
    },
    h5: {
      fontSize: baseTypography.fontSize.base,
      fontWeight: baseTypography.fontWeight.medium,
      lineHeight: baseTypography.lineHeight.normal,
      letterSpacing: baseTypography.letterSpacing.normal,
      fontFamily: baseTypography.fontFamily.sans,
    },
    h6: {
      fontSize: baseTypography.fontSize.sm,
      fontWeight: baseTypography.fontWeight.medium,
      lineHeight: baseTypography.lineHeight.normal,
      letterSpacing: baseTypography.letterSpacing.wide,
      fontFamily: baseTypography.fontFamily.sans,
      textTransform: 'uppercase' as const,
    },
  },

  // Body text for content
  body: {
    xl: {
      fontSize: baseTypography.fontSize.xl,
      fontWeight: baseTypography.fontWeight.normal,
      lineHeight: baseTypography.lineHeight.relaxed,
      letterSpacing: baseTypography.letterSpacing.normal,
      fontFamily: baseTypography.fontFamily.sans,
    },
    lg: {
      fontSize: baseTypography.fontSize.lg,
      fontWeight: baseTypography.fontWeight.normal,
      lineHeight: baseTypography.lineHeight.relaxed,
      letterSpacing: baseTypography.letterSpacing.normal,
      fontFamily: baseTypography.fontFamily.sans,
    },
    md: {
      fontSize: baseTypography.fontSize.base,
      fontWeight: baseTypography.fontWeight.normal,
      lineHeight: baseTypography.lineHeight.normal,
      letterSpacing: baseTypography.letterSpacing.normal,
      fontFamily: baseTypography.fontFamily.sans,
    },
    sm: {
      fontSize: baseTypography.fontSize.sm,
      fontWeight: baseTypography.fontWeight.normal,
      lineHeight: baseTypography.lineHeight.normal,
      letterSpacing: baseTypography.letterSpacing.normal,
      fontFamily: baseTypography.fontFamily.sans,
    },
    xs: {
      fontSize: baseTypography.fontSize.xs,
      fontWeight: baseTypography.fontWeight.normal,
      lineHeight: baseTypography.lineHeight.tight,
      letterSpacing: baseTypography.letterSpacing.normal,
      fontFamily: baseTypography.fontFamily.sans,
    },
  },

  // Labels and form elements
  label: {
    lg: {
      fontSize: baseTypography.fontSize.base,
      fontWeight: baseTypography.fontWeight.medium,
      lineHeight: baseTypography.lineHeight.normal,
      letterSpacing: baseTypography.letterSpacing.normal,
      fontFamily: baseTypography.fontFamily.sans,
    },
    md: {
      fontSize: baseTypography.fontSize.sm,
      fontWeight: baseTypography.fontWeight.medium,
      lineHeight: baseTypography.lineHeight.normal,
      letterSpacing: baseTypography.letterSpacing.normal,
      fontFamily: baseTypography.fontFamily.sans,
    },
    sm: {
      fontSize: baseTypography.fontSize.xs,
      fontWeight: baseTypography.fontWeight.medium,
      lineHeight: baseTypography.lineHeight.tight,
      letterSpacing: baseTypography.letterSpacing.wide,
      fontFamily: baseTypography.fontFamily.sans,
      textTransform: 'uppercase' as const,
    },
  },

  // Code and terminal text
  code: {
    lg: {
      fontSize: baseTypography.fontSize.base,
      fontWeight: baseTypography.fontWeight.normal,
      lineHeight: baseTypography.lineHeight.relaxed,
      letterSpacing: baseTypography.letterSpacing.normal,
      fontFamily: baseTypography.fontFamily.mono,
    },
    md: {
      fontSize: baseTypography.fontSize.sm,
      fontWeight: baseTypography.fontWeight.normal,
      lineHeight: baseTypography.lineHeight.normal,
      letterSpacing: baseTypography.letterSpacing.normal,
      fontFamily: baseTypography.fontFamily.mono,
    },
    sm: {
      fontSize: baseTypography.fontSize.xs,
      fontWeight: baseTypography.fontWeight.normal,
      lineHeight: baseTypography.lineHeight.tight,
      letterSpacing: baseTypography.letterSpacing.normal,
      fontFamily: baseTypography.fontFamily.mono,
    },
  },

  // Captions and helper text
  caption: {
    lg: {
      fontSize: baseTypography.fontSize.sm,
      fontWeight: baseTypography.fontWeight.normal,
      lineHeight: baseTypography.lineHeight.normal,
      letterSpacing: baseTypography.letterSpacing.normal,
      fontFamily: baseTypography.fontFamily.sans,
    },
    md: {
      fontSize: baseTypography.fontSize.xs,
      fontWeight: baseTypography.fontWeight.normal,
      lineHeight: baseTypography.lineHeight.tight,
      letterSpacing: baseTypography.letterSpacing.normal,
      fontFamily: baseTypography.fontFamily.sans,
    },
  },

  // Overline text for categories and metadata
  overline: {
    lg: {
      fontSize: baseTypography.fontSize.sm,
      fontWeight: baseTypography.fontWeight.semibold,
      lineHeight: baseTypography.lineHeight.tight,
      letterSpacing: baseTypography.letterSpacing.wider,
      fontFamily: baseTypography.fontFamily.sans,
      textTransform: 'uppercase' as const,
    },
    md: {
      fontSize: baseTypography.fontSize.xs,
      fontWeight: baseTypography.fontWeight.semibold,
      lineHeight: baseTypography.lineHeight.tight,
      letterSpacing: baseTypography.letterSpacing.widest,
      fontFamily: baseTypography.fontFamily.sans,
      textTransform: 'uppercase' as const,
    },
  },
} as const;

// Context-specific typography styles for enterprise applications
export const enterpriseTextStyles = {
  // Dashboard and metrics
  dashboard: {
    metric: {
      value: {
        fontSize: baseTypography.fontSize['4xl'],
        fontWeight: baseTypography.fontWeight.bold,
        lineHeight: baseTypography.lineHeight.none,
        letterSpacing: baseTypography.letterSpacing.tight,
        fontFamily: baseTypography.fontFamily.sans,
      },
      label: {
        fontSize: baseTypography.fontSize.sm,
        fontWeight: baseTypography.fontWeight.medium,
        lineHeight: baseTypography.lineHeight.tight,
        letterSpacing: baseTypography.letterSpacing.wide,
        fontFamily: baseTypography.fontFamily.sans,
        textTransform: 'uppercase' as const,
      },
      trend: {
        fontSize: baseTypography.fontSize.xs,
        fontWeight: baseTypography.fontWeight.medium,
        lineHeight: baseTypography.lineHeight.tight,
        letterSpacing: baseTypography.letterSpacing.normal,
        fontFamily: baseTypography.fontFamily.sans,
      },
    },
    widget: {
      title: {
        fontSize: baseTypography.fontSize.lg,
        fontWeight: baseTypography.fontWeight.semibold,
        lineHeight: baseTypography.lineHeight.snug,
        letterSpacing: baseTypography.letterSpacing.normal,
        fontFamily: baseTypography.fontFamily.sans,
      },
      subtitle: {
        fontSize: baseTypography.fontSize.sm,
        fontWeight: baseTypography.fontWeight.normal,
        lineHeight: baseTypography.lineHeight.normal,
        letterSpacing: baseTypography.letterSpacing.normal,
        fontFamily: baseTypography.fontFamily.sans,
      },
    },
  },

  // Command interface and terminal
  command: {
    input: {
      fontSize: baseTypography.fontSize.sm,
      fontWeight: baseTypography.fontWeight.normal,
      lineHeight: baseTypography.lineHeight.normal,
      letterSpacing: baseTypography.letterSpacing.normal,
      fontFamily: baseTypography.fontFamily.mono,
    },
    output: {
      fontSize: baseTypography.fontSize.sm,
      fontWeight: baseTypography.fontWeight.normal,
      lineHeight: baseTypography.lineHeight.relaxed,
      letterSpacing: baseTypography.letterSpacing.normal,
      fontFamily: baseTypography.fontFamily.mono,
    },
    prompt: {
      fontSize: baseTypography.fontSize.sm,
      fontWeight: baseTypography.fontWeight.medium,
      lineHeight: baseTypography.lineHeight.normal,
      letterSpacing: baseTypography.letterSpacing.normal,
      fontFamily: baseTypography.fontFamily.mono,
    },
  },

  // Status and notifications
  status: {
    title: {
      fontSize: baseTypography.fontSize.sm,
      fontWeight: baseTypography.fontWeight.medium,
      lineHeight: baseTypography.lineHeight.normal,
      letterSpacing: baseTypography.letterSpacing.normal,
      fontFamily: baseTypography.fontFamily.sans,
    },
    message: {
      fontSize: baseTypography.fontSize.sm,
      fontWeight: baseTypography.fontWeight.normal,
      lineHeight: baseTypography.lineHeight.normal,
      letterSpacing: baseTypography.letterSpacing.normal,
      fontFamily: baseTypography.fontFamily.sans,
    },
    badge: {
      fontSize: baseTypography.fontSize.xs,
      fontWeight: baseTypography.fontWeight.medium,
      lineHeight: baseTypography.lineHeight.tight,
      letterSpacing: baseTypography.letterSpacing.wide,
      fontFamily: baseTypography.fontFamily.sans,
      textTransform: 'uppercase' as const,
    },
  },

  // Tables and data display
  table: {
    header: {
      fontSize: baseTypography.fontSize.xs,
      fontWeight: baseTypography.fontWeight.semibold,
      lineHeight: baseTypography.lineHeight.tight,
      letterSpacing: baseTypography.letterSpacing.wider,
      fontFamily: baseTypography.fontFamily.sans,
      textTransform: 'uppercase' as const,
    },
    cell: {
      fontSize: baseTypography.fontSize.sm,
      fontWeight: baseTypography.fontWeight.normal,
      lineHeight: baseTypography.lineHeight.normal,
      letterSpacing: baseTypography.letterSpacing.normal,
      fontFamily: baseTypography.fontFamily.sans,
    },
    numeric: {
      fontSize: baseTypography.fontSize.sm,
      fontWeight: baseTypography.fontWeight.medium,
      lineHeight: baseTypography.lineHeight.normal,
      letterSpacing: baseTypography.letterSpacing.normal,
      fontFamily: baseTypography.fontFamily.mono,
    },
  },

  // Navigation and UI chrome
  navigation: {
    primary: {
      fontSize: baseTypography.fontSize.sm,
      fontWeight: baseTypography.fontWeight.medium,
      lineHeight: baseTypography.lineHeight.normal,
      letterSpacing: baseTypography.letterSpacing.normal,
      fontFamily: baseTypography.fontFamily.sans,
    },
    secondary: {
      fontSize: baseTypography.fontSize.sm,
      fontWeight: baseTypography.fontWeight.normal,
      lineHeight: baseTypography.lineHeight.normal,
      letterSpacing: baseTypography.letterSpacing.normal,
      fontFamily: baseTypography.fontFamily.sans,
    },
    breadcrumb: {
      fontSize: baseTypography.fontSize.sm,
      fontWeight: baseTypography.fontWeight.normal,
      lineHeight: baseTypography.lineHeight.normal,
      letterSpacing: baseTypography.letterSpacing.normal,
      fontFamily: baseTypography.fontFamily.sans,
    },
  },

  // Buttons and interactive elements
  button: {
    lg: {
      fontSize: baseTypography.fontSize.base,
      fontWeight: baseTypography.fontWeight.medium,
      lineHeight: baseTypography.lineHeight.normal,
      letterSpacing: baseTypography.letterSpacing.normal,
      fontFamily: baseTypography.fontFamily.sans,
    },
    md: {
      fontSize: baseTypography.fontSize.sm,
      fontWeight: baseTypography.fontWeight.medium,
      lineHeight: baseTypography.lineHeight.normal,
      letterSpacing: baseTypography.letterSpacing.normal,
      fontFamily: baseTypography.fontFamily.sans,
    },
    sm: {
      fontSize: baseTypography.fontSize.xs,
      fontWeight: baseTypography.fontWeight.medium,
      lineHeight: baseTypography.lineHeight.tight,
      letterSpacing: baseTypography.letterSpacing.wide,
      fontFamily: baseTypography.fontFamily.sans,
    },
  },
} as const;

// Responsive typography utilities
export const responsiveTextStyles = {
  // Page titles that scale responsively
  pageTitle: {
    base: textStyles.heading.h2,
    md: textStyles.heading.h1,
    lg: textStyles.display.sm,
  },
  
  // Section titles
  sectionTitle: {
    base: textStyles.heading.h4,
    md: textStyles.heading.h3,
    lg: textStyles.heading.h2,
  },
  
  // Component titles
  componentTitle: {
    base: textStyles.heading.h5,
    md: textStyles.heading.h4,
  },
  
  // Body text
  bodyText: {
    base: textStyles.body.sm,
    md: textStyles.body.md,
    lg: textStyles.body.lg,
  },
  
  // Caption text
  captionText: {
    base: textStyles.caption.md,
    md: textStyles.caption.lg,
  },
} as const;

// Text color utilities for semantic context
export const textColorClasses = {
  // Primary text colors
  primary: 'text-text-primary',
  secondary: 'text-text-secondary',
  tertiary: 'text-text-tertiary',
  inverse: 'text-text-inverse',
  disabled: 'text-text-disabled',
  
  // Semantic text colors
  success: 'text-success-700 dark:text-success-400',
  warning: 'text-warning-700 dark:text-warning-400',
  error: 'text-error-700 dark:text-error-400',
  info: 'text-info-700 dark:text-info-400',
  
  // Brand colors
  brand: 'text-primary-600 dark:text-primary-400',
  kubernetes: 'text-kubernetes-blue',
  
  // Enterprise semantic colors
  compliant: 'text-enterprise-compliant',
  safe: 'text-enterprise-safe',
  risky: 'text-enterprise-risky',
  dangerous: 'text-enterprise-dangerous',
  
  // Code and terminal
  code: 'text-gray-800 dark:text-gray-200',
  terminal: 'text-green-400',
} as const;

// Export type definitions
export type TextStyle = typeof textStyles;
export type EnterpriseTextStyle = typeof enterpriseTextStyles;
export type ResponsiveTextStyle = typeof responsiveTextStyles;
export type TextColorClass = keyof typeof textColorClasses;