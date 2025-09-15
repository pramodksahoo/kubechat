import React, { forwardRef, HTMLAttributes, ElementType } from 'react';
import { cn } from '../../lib/utils';
import { textColorClasses, TextColorClass } from '../typography';

// Base text component props
interface BaseTextProps extends HTMLAttributes<HTMLElement> {
  as?: ElementType;
  color?: TextColorClass;
  className?: string;
  children: React.ReactNode;
}

// Typography variant props
interface TypographyProps extends BaseTextProps {
  variant?: 
    | 'display-2xl' | 'display-xl' | 'display-lg' | 'display-md' | 'display-sm'
    | 'h1' | 'h2' | 'h3' | 'h4' | 'h5' | 'h6'
    | 'body-xl' | 'body-lg' | 'body-md' | 'body-sm' | 'body-xs'
    | 'label-lg' | 'label-md' | 'label-sm'
    | 'code-lg' | 'code-md' | 'code-sm'
    | 'caption-lg' | 'caption-md'
    | 'overline-lg' | 'overline-md';
}

// Variant to class mapping following our typography system
const variantClasses = {
  // Display variants
  'display-2xl': 'text-6xl font-extrabold leading-none tracking-tight font-display',
  'display-xl': 'text-5xl font-bold leading-tight tracking-tight font-display',
  'display-lg': 'text-4xl font-bold leading-tight font-display',
  'display-md': 'text-3xl font-semibold leading-snug font-display',
  'display-sm': 'text-2xl font-semibold leading-snug font-display',
  
  // Heading variants
  'h1': 'text-3xl font-bold leading-tight tracking-tight',
  'h2': 'text-2xl font-semibold leading-snug',
  'h3': 'text-xl font-semibold leading-snug',
  'h4': 'text-lg font-medium leading-normal',
  'h5': 'text-base font-medium leading-normal',
  'h6': 'text-sm font-medium leading-normal tracking-wide uppercase',
  
  // Body text variants
  'body-xl': 'text-xl font-normal leading-relaxed',
  'body-lg': 'text-lg font-normal leading-relaxed',
  'body-md': 'text-base font-normal leading-normal',
  'body-sm': 'text-sm font-normal leading-normal',
  'body-xs': 'text-xs font-normal leading-tight',
  
  // Label variants
  'label-lg': 'text-base font-medium leading-normal',
  'label-md': 'text-sm font-medium leading-normal',
  'label-sm': 'text-xs font-medium leading-tight tracking-wide uppercase',
  
  // Code variants
  'code-lg': 'text-base font-normal leading-relaxed font-mono',
  'code-md': 'text-sm font-normal leading-normal font-mono',
  'code-sm': 'text-xs font-normal leading-tight font-mono',
  
  // Caption variants
  'caption-lg': 'text-sm font-normal leading-normal',
  'caption-md': 'text-xs font-normal leading-tight',
  
  // Overline variants
  'overline-lg': 'text-sm font-semibold leading-tight tracking-wider uppercase',
  'overline-md': 'text-xs font-semibold leading-tight tracking-widest uppercase',
} as const;

// Default elements for variants
const variantElements = {
  'display-2xl': 'h1',
  'display-xl': 'h1',
  'display-lg': 'h1',
  'display-md': 'h1',
  'display-sm': 'h2',
  'h1': 'h1',
  'h2': 'h2',
  'h3': 'h3',
  'h4': 'h4',
  'h5': 'h5',
  'h6': 'h6',
  'body-xl': 'p',
  'body-lg': 'p',
  'body-md': 'p',
  'body-sm': 'p',
  'body-xs': 'p',
  'label-lg': 'label',
  'label-md': 'label',
  'label-sm': 'label',
  'code-lg': 'code',
  'code-md': 'code',
  'code-sm': 'code',
  'caption-lg': 'span',
  'caption-md': 'span',
  'overline-lg': 'span',
  'overline-md': 'span',
} as const;

// Generic Text component
export const Text = forwardRef<HTMLElement, TypographyProps>(
  ({ as, variant = 'body-md', color = 'primary', className, children, ...props }, ref) => {
    const Component = as || (variant ? variantElements[variant] : 'span');
    
    const classes = cn(
      variant && variantClasses[variant],
      color && textColorClasses[color],
      'transition-colors duration-200',
      className
    );
    
    return (
      <Component ref={ref} className={classes} {...props}>
        {children}
      </Component>
    );
  }
);

Text.displayName = 'Text';

// Specialized typography components with context-specific styling

// Display component for hero sections
export const Display = forwardRef<HTMLHeadingElement, Omit<TypographyProps, 'variant'> & { 
  size?: 'sm' | 'md' | 'lg' | 'xl' | '2xl' 
}>(
  ({ size = 'md', ...props }, ref) => (
    <Text ref={ref} variant={`display-${size}` as TypographyProps['variant']} {...props} />
  )
);

Display.displayName = 'Display';

// Heading component
export const Heading = forwardRef<HTMLHeadingElement, Omit<TypographyProps, 'variant'> & { 
  level?: 1 | 2 | 3 | 4 | 5 | 6 
}>(
  ({ level = 1, ...props }, ref) => (
    <Text ref={ref} variant={`h${level}` as TypographyProps['variant']} {...props} />
  )
);

Heading.displayName = 'Heading';

// Body text component
export const Body = forwardRef<HTMLParagraphElement, Omit<TypographyProps, 'variant'> & { 
  size?: 'xs' | 'sm' | 'md' | 'lg' | 'xl' 
}>(
  ({ size = 'md', ...props }, ref) => (
    <Text ref={ref} variant={`body-${size}` as TypographyProps['variant']} {...props} />
  )
);

Body.displayName = 'Body';

// Label component for forms
export const Label = forwardRef<HTMLLabelElement, Omit<TypographyProps, 'variant'> & { 
  size?: 'sm' | 'md' | 'lg' 
}>(
  ({ size = 'md', as = 'label', ...props }, ref) => (
    <Text ref={ref} as={as} variant={`label-${size}` as TypographyProps['variant']} {...props} />
  )
);

Label.displayName = 'Label';

// Code component for inline code
export const Code = forwardRef<HTMLElement, Omit<TypographyProps, 'variant'> & { 
  size?: 'sm' | 'md' | 'lg' 
}>(
  ({ size = 'md', className, ...props }, ref) => (
    <Text 
      ref={ref} 
      variant={`code-${size}` as TypographyProps['variant']} 
      className={cn('code-inline', className)}
      {...props} 
    />
  )
);

Code.displayName = 'Code';

// Caption component for helper text
export const Caption = forwardRef<HTMLSpanElement, Omit<TypographyProps, 'variant'> & { 
  size?: 'md' | 'lg' 
}>(
  ({ size = 'md', ...props }, ref) => (
    <Text ref={ref} variant={`caption-${size}` as TypographyProps['variant']} {...props} />
  )
);

Caption.displayName = 'Caption';

// Overline component for categories
export const Overline = forwardRef<HTMLSpanElement, Omit<TypographyProps, 'variant'> & { 
  size?: 'md' | 'lg' 
}>(
  ({ size = 'md', ...props }, ref) => (
    <Text ref={ref} variant={`overline-${size}` as TypographyProps['variant']} {...props} />
  )
);

Overline.displayName = 'Overline';

// Enterprise-specific typography components

// Dashboard metric display
export const MetricValue = forwardRef<HTMLDivElement, BaseTextProps>(
  ({ className, color = 'primary', ...props }, ref) => (
    <Text 
      ref={ref}
      as="div"
      className={cn('text-4xl font-bold leading-none tracking-tight', className)}
      color={color}
      {...props}
    />
  )
);

MetricValue.displayName = 'MetricValue';

// Dashboard metric label
export const MetricLabel = forwardRef<HTMLDivElement, BaseTextProps>(
  ({ className, color = 'secondary', ...props }, ref) => (
    <Text 
      ref={ref}
      as="div"
      className={cn('text-sm font-medium leading-tight tracking-wide uppercase', className)}
      color={color}
      {...props}
    />
  )
);

MetricLabel.displayName = 'MetricLabel';

// Terminal/command text
export const Terminal = forwardRef<HTMLPreElement, BaseTextProps>(
  ({ className, color = 'code', as = 'pre', ...props }, ref) => (
    <Text 
      ref={ref}
      as={as}
      className={cn('font-mono text-sm leading-relaxed whitespace-pre-wrap', className)}
      color={color}
      {...props}
    />
  )
);

Terminal.displayName = 'Terminal';

// Status badge text
export const StatusBadge = forwardRef<HTMLSpanElement, BaseTextProps & {
  variant?: 'success' | 'warning' | 'error' | 'info' | 'neutral';
}>(
  ({ variant = 'neutral', className, ...props }, ref) => {
    const variantClasses = {
      success: 'text-success-700 dark:text-success-200',
      warning: 'text-warning-700 dark:text-warning-200', 
      error: 'text-error-700 dark:text-error-200',
      info: 'text-info-700 dark:text-info-200',
      neutral: 'text-text-secondary',
    };
    
    return (
      <Text 
        ref={ref}
        as="span"
        className={cn(
          'text-xs font-medium leading-tight tracking-wide uppercase',
          variantClasses[variant],
          className
        )}
        {...props}
      />
    );
  }
);

StatusBadge.displayName = 'StatusBadge';

// Page title with responsive sizing
export const PageTitle = forwardRef<HTMLHeadingElement, BaseTextProps>(
  ({ className, ...props }, ref) => (
    <Text 
      ref={ref}
      as="h1"
      className={cn(
        'text-2xl font-semibold leading-snug md:text-3xl md:font-bold md:leading-tight lg:text-4xl',
        className
      )}
      {...props}
    />
  )
);

PageTitle.displayName = 'PageTitle';

// Section title with responsive sizing
export const SectionTitle = forwardRef<HTMLHeadingElement, BaseTextProps>(
  ({ className, ...props }, ref) => (
    <Text 
      ref={ref}
      as="h2"
      className={cn(
        'text-lg font-medium leading-normal md:text-xl md:font-semibold md:leading-snug lg:text-2xl',
        className
      )}
      {...props}
    />
  )
);

SectionTitle.displayName = 'SectionTitle';

// Table header text
export const TableHeader = forwardRef<HTMLElement, BaseTextProps>(
  ({ className, color = 'secondary', as = 'th', ...props }, ref) => (
    <Text 
      ref={ref}
      as={as}
      className={cn(
        'text-xs font-semibold leading-tight tracking-wider uppercase',
        className
      )}
      color={color}
      {...props}
    />
  )
);

TableHeader.displayName = 'TableHeader';

// Table cell text with optional numeric styling
export const TableCell = forwardRef<HTMLElement, BaseTextProps & {
  numeric?: boolean;
}>(
  ({ numeric, className, as = 'td', ...props }, ref) => (
    <Text 
      ref={ref}
      as={as}
      className={cn(
        'text-sm font-normal leading-normal',
        numeric && 'font-mono font-medium tabular-nums',
        className
      )}
      {...props}
    />
  )
);

TableCell.displayName = 'TableCell';