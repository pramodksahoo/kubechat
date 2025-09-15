import React, { forwardRef, ButtonHTMLAttributes } from 'react';
import { cn } from '../../lib/utils';

// Button variant types
type ButtonVariant = 
  | 'primary' 
  | 'secondary' 
  | 'success' 
  | 'warning' 
  | 'danger' 
  | 'ghost' 
  | 'outline'
  | 'kubernetes';

// Button size types
type ButtonSize = 'sm' | 'md' | 'lg';

// Button component props
interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: ButtonVariant;
  size?: ButtonSize;
  loading?: boolean;
  leftIcon?: React.ReactNode;
  rightIcon?: React.ReactNode;
  fullWidth?: boolean;
  children: React.ReactNode;
}

// Base button styles
const baseButtonClasses = 
  'enterprise-button inline-flex items-center justify-center font-medium rounded-md focus-enterprise ' +
  'transition-all duration-200 select-none cursor-pointer ' +
  'disabled:opacity-50 disabled:cursor-not-allowed disabled:pointer-events-none ' +
  'active:scale-[0.98]';

// Variant styles
const variantClasses: Record<ButtonVariant, string> = {
  primary: 
    'bg-primary-600 text-white shadow-sm hover:bg-primary-700 active:bg-primary-800 ' +
    'hover:shadow-md border border-primary-600 hover:border-primary-700',
  
  secondary: 
    'bg-background-primary text-text-primary border border-border-primary shadow-sm ' +
    'hover:bg-background-secondary hover:border-border-secondary active:bg-background-tertiary',
  
  success: 
    'bg-success-600 text-white shadow-sm hover:bg-success-700 active:bg-success-800 ' +
    'border border-success-600 hover:border-success-700',
  
  warning: 
    'bg-warning-600 text-white shadow-sm hover:bg-warning-700 active:bg-warning-800 ' +
    'border border-warning-600 hover:border-warning-700',
  
  danger: 
    'bg-error-600 text-white shadow-sm hover:bg-error-700 active:bg-error-800 ' +
    'border border-error-600 hover:border-error-700',
  
  ghost: 
    'text-text-primary hover:bg-background-secondary active:bg-background-tertiary ' +
    'border border-transparent',
  
  outline: 
    'bg-transparent text-text-primary border border-border-primary ' +
    'hover:bg-background-secondary hover:border-border-secondary active:bg-background-tertiary',
  
  kubernetes: 
    'bg-kubernetes-blue text-white shadow-sm hover:bg-kubernetes-navy active:bg-kubernetes-navy ' +
    'border border-kubernetes-blue hover:border-kubernetes-navy',
};

// Size styles
const sizeClasses: Record<ButtonSize, string> = {
  sm: 'px-button-sm-x py-button-sm-y text-sm h-button-sm gap-1.5',
  md: 'px-button-md-x py-button-md-y text-base h-button-md gap-2',
  lg: 'px-button-lg-x py-button-lg-y text-lg h-button-lg gap-2.5',
};

// Loading spinner component
const LoadingSpinner = ({ size }: { size: ButtonSize }) => {
  const spinnerSizes = {
    sm: 'w-3 h-3',
    md: 'w-4 h-4',
    lg: 'w-5 h-5',
  };
  
  return (
    <svg 
      className={cn('animate-spin', spinnerSizes[size])} 
      xmlns="http://www.w3.org/2000/svg" 
      fill="none" 
      viewBox="0 0 24 24"
    >
      <circle 
        className="opacity-25" 
        cx="12" 
        cy="12" 
        r="10" 
        stroke="currentColor" 
        strokeWidth="4"
      />
      <path 
        className="opacity-75" 
        fill="currentColor" 
        d="m4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
      />
    </svg>
  );
};

// Main Button component
export const Button = forwardRef<HTMLButtonElement, ButtonProps>(
  ({ 
    variant = 'primary', 
    size = 'md', 
    loading = false,
    leftIcon,
    rightIcon,
    fullWidth = false,
    className,
    children,
    disabled,
    ...props 
  }, ref) => {
    const isDisabled = disabled || loading;
    
    return (
      <button
        ref={ref}
        className={cn(
          baseButtonClasses,
          variantClasses[variant],
          sizeClasses[size],
          fullWidth && 'w-full',
          className
        )}
        disabled={isDisabled}
        {...props}
      >
        {loading && <LoadingSpinner size={size} />}
        {!loading && leftIcon && leftIcon}
        <span className={loading ? 'opacity-0' : ''}>{children}</span>
        {!loading && rightIcon && rightIcon}
      </button>
    );
  }
);

Button.displayName = 'Button';

// Specialized button variants

// Icon button for icon-only actions
interface IconButtonProps extends Omit<ButtonProps, 'leftIcon' | 'rightIcon' | 'children'> {
  icon: React.ReactNode;
  'aria-label': string;
}

export const IconButton = forwardRef<HTMLButtonElement, IconButtonProps>(
  ({ icon, size = 'md', className, ...props }, ref) => {
    const iconSizes = {
      sm: 'w-4 h-4',
      md: 'w-5 h-5', 
      lg: 'w-6 h-6',
    };
    
    const paddingClasses = {
      sm: 'p-1.5',
      md: 'p-2',
      lg: 'p-2.5',
    };
    
    return (
      <Button
        ref={ref}
        size={size}
        className={cn(
          paddingClasses[size],
          'aspect-square',
          className
        )}
        {...props}
      >
        <span className={iconSizes[size]}>{icon}</span>
      </Button>
    );
  }
);

IconButton.displayName = 'IconButton';

// Command button for dangerous operations
interface CommandButtonProps extends ButtonProps {
  safetyLevel?: 'safe' | 'risky' | 'dangerous';
  confirmText?: string;
}

export const CommandButton = forwardRef<HTMLButtonElement, CommandButtonProps>(
  ({ safetyLevel = 'safe', variant, className, ...props }, ref) => {
    // Override variant based on safety level if not explicitly set
    const safetyVariant = variant || {
      safe: 'success' as const,
      risky: 'warning' as const,
      dangerous: 'danger' as const,
    }[safetyLevel];
    
    const safetyClasses = {
      safe: 'command-safe',
      risky: 'command-risky', 
      dangerous: 'command-dangerous',
    };
    
    return (
      <Button
        ref={ref}
        variant={safetyVariant}
        className={cn(
          safetyClasses[safetyLevel],
          className
        )}
        {...props}
      />
    );
  }
);

CommandButton.displayName = 'CommandButton';

// Button group for related actions
interface ButtonGroupProps {
  children: React.ReactElement<ButtonProps>[];
  className?: string;
  orientation?: 'horizontal' | 'vertical';
}

export const ButtonGroup = ({ children, className, orientation = 'horizontal' }: ButtonGroupProps) => {
  const orientationClasses = {
    horizontal: 'flex-row',
    vertical: 'flex-col',
  };
  
  return (
    <div 
      className={cn(
        'inline-flex', 
        orientationClasses[orientation],
        className
      )}
      role="group"
    >
      {React.Children.map(children, (child, index) => {
        if (!React.isValidElement(child)) return child;
        
        const isFirst = index === 0;
        const isLast = index === children.length - 1;
        
        return React.cloneElement(child, {
          className: cn(
            child.props.className,
            orientation === 'horizontal' ? [
              !isFirst && '-ml-px',
              !isFirst && !isLast && 'rounded-none',
              isFirst && 'rounded-r-none',
              isLast && 'rounded-l-none',
            ] : [
              !isFirst && '-mt-px',
              !isFirst && !isLast && 'rounded-none',
              isFirst && 'rounded-b-none',
              isLast && 'rounded-t-none',
            ]
          ),
        });
      })}
    </div>
  );
};

ButtonGroup.displayName = 'ButtonGroup';

// Floating Action Button (FAB)
interface FABProps extends Omit<IconButtonProps, 'variant' | 'size'> {
  position?: 'bottom-right' | 'bottom-left' | 'top-right' | 'top-left';
  size?: 'md' | 'lg';
}

export const FAB = forwardRef<HTMLButtonElement, FABProps>(
  ({ position = 'bottom-right', size = 'lg', className, ...props }, ref) => {
    const positionClasses = {
      'bottom-right': 'fixed bottom-6 right-6',
      'bottom-left': 'fixed bottom-6 left-6',
      'top-right': 'fixed top-6 right-6',
      'top-left': 'fixed top-6 left-6',
    };
    
    const fabSizes = {
      md: 'w-14 h-14',
      lg: 'w-16 h-16',
    };
    
    return (
      <IconButton
        ref={ref}
        variant="primary"
        size={size}
        className={cn(
          positionClasses[position],
          fabSizes[size],
          'rounded-full shadow-lg hover:shadow-xl z-50',
          'transform hover:scale-110 active:scale-95',
          className
        )}
        {...props}
      />
    );
  }
);

FAB.displayName = 'FAB';

// Link button for navigation that looks like a button
interface LinkButtonProps extends Omit<ButtonProps, 'type'> {
  href: string;
  external?: boolean;
  target?: string;
  rel?: string;
}

export const LinkButton = forwardRef<HTMLAnchorElement, LinkButtonProps>(
  ({ href, external = false, target, rel, className, ...props }, ref) => {
    const linkProps = {
      href,
      target: target || (external ? '_blank' : undefined),
      rel: rel || (external ? 'noopener noreferrer' : undefined),
    };
    
    return (
      <a
        ref={ref}
        className={cn(
          baseButtonClasses,
          variantClasses[props.variant || 'primary'],
          sizeClasses[props.size || 'md'],
          props.fullWidth && 'w-full',
          'no-underline',
          className
        )}
        {...linkProps}
      >
        {props.leftIcon && props.leftIcon}
        <span>{props.children}</span>
        {props.rightIcon && props.rightIcon}
      </a>
    );
  }
);

LinkButton.displayName = 'LinkButton';