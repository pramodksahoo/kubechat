import React, { forwardRef, InputHTMLAttributes, TextareaHTMLAttributes } from 'react';
import { cn } from '../../lib/utils';

// Input size types
type InputSize = 'sm' | 'md' | 'lg';

// Input state types
type InputState = 'default' | 'error' | 'success';

// Base input props
interface BaseInputProps {
  size?: InputSize;
  state?: InputState;
  leftIcon?: React.ReactNode;
  rightIcon?: React.ReactNode;
  leftAddon?: React.ReactNode;
  rightAddon?: React.ReactNode;
  fullWidth?: boolean;
  label?: string;
  helperText?: string;
  errorText?: string;
  required?: boolean;
}

// Input component props
interface InputProps extends BaseInputProps, Omit<InputHTMLAttributes<HTMLInputElement>, 'size'> {}

// Textarea component props  
interface TextareaProps extends BaseInputProps, Omit<TextareaHTMLAttributes<HTMLTextAreaElement>, 'size'> {}

// Base input styles
const baseInputClasses = 
  'enterprise-input block w-full rounded-md border bg-background-primary text-text-primary ' +
  'placeholder-text-tertiary transition-all duration-200 ' +
  'disabled:opacity-50 disabled:cursor-not-allowed ' +
  'read-only:bg-background-secondary read-only:focus:ring-0';

// Size styles
const sizeClasses: Record<InputSize, string> = {
  sm: 'px-input-sm-x py-input-sm-y text-sm h-input-sm',
  md: 'px-input-md-x py-input-md-y text-base h-input-md',
  lg: 'px-input-lg-x py-input-lg-y text-lg h-input-lg',
};

// State styles
const stateClasses: Record<InputState, string> = {
  default: 'border-border-primary hover:border-border-secondary focus:border-border-focus focus:ring-2 focus:ring-border-focus focus:ring-opacity-20',
  error: 'border-border-error focus:border-border-error focus:ring-2 focus:ring-border-error focus:ring-opacity-20',
  success: 'border-success-500 focus:border-success-500 focus:ring-2 focus:ring-success-500 focus:ring-opacity-20',
};

// Icon sizes for different input sizes
const iconSizes: Record<InputSize, string> = {
  sm: 'w-4 h-4',
  md: 'w-5 h-5',
  lg: 'w-6 h-6',
};

// Label component
const InputLabel = forwardRef<HTMLLabelElement, {
  htmlFor?: string;
  required?: boolean;
  children: React.ReactNode;
  className?: string;
}>(({ htmlFor, required, children, className }, ref) => (
  <label
    ref={ref}
    htmlFor={htmlFor}
    className={cn(
      'form-label block text-sm font-medium text-text-primary mb-1.5',
      className
    )}
  >
    {children}
    {required && <span className="text-error-500 ml-1">*</span>}
  </label>
));

InputLabel.displayName = 'InputLabel';

// Helper text component
const InputHelperText = ({ 
  children, 
  state = 'default',
  className 
}: { 
  children: React.ReactNode;
  state?: InputState;
  className?: string;
}) => {
  const stateColors = {
    default: 'text-text-tertiary',
    error: 'text-error-600 dark:text-error-400',
    success: 'text-success-600 dark:text-success-400',
  };
  
  return (
    <div className={cn('form-help mt-1.5 text-sm', stateColors[state], className)}>
      {children}
    </div>
  );
};

// Input wrapper component for handling icons and addons
const InputWrapper = ({
  children,
  leftIcon,
  rightIcon,
  leftAddon,
  rightAddon,
  size = 'md',
  state = 'default',
  className,
}: {
  children: React.ReactNode;
  leftIcon?: React.ReactNode;
  rightIcon?: React.ReactNode;
  leftAddon?: React.ReactNode;
  rightAddon?: React.ReactNode;
  size?: InputSize;
  state?: InputState;
  className?: string;
}) => {
  const hasLeftIcon = Boolean(leftIcon);
  const hasRightIcon = Boolean(rightIcon);
  const hasLeftAddon = Boolean(leftAddon);
  const hasRightAddon = Boolean(rightAddon);
  
  if (!hasLeftIcon && !hasRightIcon && !hasLeftAddon && !hasRightAddon) {
    return <>{children}</>;
  }
  
  return (
    <div className={cn('relative flex items-center', className)}>
      {/* Left addon */}
      {hasLeftAddon && (
        <div className={cn(
          'flex items-center px-3 bg-background-secondary border border-r-0 border-border-primary rounded-l-md',
          sizeClasses[size].includes('h-input-sm') && 'h-input-sm',
          sizeClasses[size].includes('h-input-md') && 'h-input-md',
          sizeClasses[size].includes('h-input-lg') && 'h-input-lg',
          state === 'error' && 'border-border-error',
          state === 'success' && 'border-success-500',
        )}>
          {leftAddon}
        </div>
      )}
      
      {/* Left icon */}
      {hasLeftIcon && !hasLeftAddon && (
        <div className={cn(
          'absolute left-3 flex items-center pointer-events-none text-text-tertiary',
          iconSizes[size]
        )}>
          {leftIcon}
        </div>
      )}
      
      {/* Input with adjusted padding */}
      <div className={cn(
        'relative flex-1',
        hasLeftIcon && !hasLeftAddon && size === 'sm' && '[&>input]:pl-8 [&>textarea]:pl-8',
        hasLeftIcon && !hasLeftAddon && size === 'md' && '[&>input]:pl-10 [&>textarea]:pl-10',
        hasLeftIcon && !hasLeftAddon && size === 'lg' && '[&>input]:pl-12 [&>textarea]:pl-12',
        hasRightIcon && !hasRightAddon && size === 'sm' && '[&>input]:pr-8 [&>textarea]:pr-8',
        hasRightIcon && !hasRightAddon && size === 'md' && '[&>input]:pr-10 [&>textarea]:pr-10',
        hasRightIcon && !hasRightAddon && size === 'lg' && '[&>input]:pr-12 [&>textarea]:pr-12',
        hasLeftAddon && '[&>input]:rounded-l-none [&>textarea]:rounded-l-none',
        hasRightAddon && '[&>input]:rounded-r-none [&>textarea]:rounded-r-none',
      )}>
        {children}
      </div>
      
      {/* Right icon */}
      {hasRightIcon && !hasRightAddon && (
        <div className={cn(
          'absolute right-3 flex items-center pointer-events-none text-text-tertiary',
          iconSizes[size]
        )}>
          {rightIcon}
        </div>
      )}
      
      {/* Right addon */}
      {hasRightAddon && (
        <div className={cn(
          'flex items-center px-3 bg-background-secondary border border-l-0 border-border-primary rounded-r-md',
          sizeClasses[size].includes('h-input-sm') && 'h-input-sm',
          sizeClasses[size].includes('h-input-md') && 'h-input-md',
          sizeClasses[size].includes('h-input-lg') && 'h-input-lg',
          state === 'error' && 'border-border-error',
          state === 'success' && 'border-success-500',
        )}>
          {rightAddon}
        </div>
      )}
    </div>
  );
};

// Main Input component
export const Input = forwardRef<HTMLInputElement, InputProps>(
  ({ 
    size = 'md',
    state = 'default',
    leftIcon,
    rightIcon,
    leftAddon,
    rightAddon,
    fullWidth = true,
    label,
    helperText,
    errorText,
    required,
    className,
    id,
    ...props 
  }, ref) => {
    const inputId = id || `input-${Math.random().toString(36).substr(2, 9)}`;
    const finalState = errorText ? 'error' : state;
    const finalHelperText = errorText || helperText;
    
    return (
      <div className={cn('form-group', !fullWidth && 'inline-block')}>
        {label && (
          <InputLabel htmlFor={inputId} required={required}>
            {label}
          </InputLabel>
        )}
        
        <InputWrapper
          leftIcon={leftIcon}
          rightIcon={rightIcon}
          leftAddon={leftAddon}
          rightAddon={rightAddon}
          size={size}
          state={finalState}
        >
          <input
            ref={ref}
            id={inputId}
            className={cn(
              baseInputClasses,
              sizeClasses[size],
              stateClasses[finalState],
              className
            )}
            {...props}
          />
        </InputWrapper>
        
        {finalHelperText && (
          <InputHelperText state={finalState}>
            {finalHelperText}
          </InputHelperText>
        )}
      </div>
    );
  }
);

Input.displayName = 'Input';

// Textarea component
export const Textarea = forwardRef<HTMLTextAreaElement, TextareaProps>(
  ({ 
    size = 'md',
    state = 'default',
    leftIcon,
    rightIcon,
    leftAddon,
    rightAddon,
    fullWidth = true,
    label,
    helperText,
    errorText,
    required,
    className,
    id,
    rows = 4,
    ...props 
  }, ref) => {
    const inputId = id || `textarea-${Math.random().toString(36).substr(2, 9)}`;
    const finalState = errorText ? 'error' : state;
    const finalHelperText = errorText || helperText;
    
    // Textarea-specific size classes (no height constraint)
    const textareaSizeClasses: Record<InputSize, string> = {
      sm: 'px-input-sm-x py-input-sm-y text-sm',
      md: 'px-input-md-x py-input-md-y text-base',
      lg: 'px-input-lg-x py-input-lg-y text-lg',
    };
    
    return (
      <div className={cn('form-group', !fullWidth && 'inline-block')}>
        {label && (
          <InputLabel htmlFor={inputId} required={required}>
            {label}
          </InputLabel>
        )}
        
        <InputWrapper
          leftIcon={leftIcon}
          rightIcon={rightIcon}
          leftAddon={leftAddon}
          rightAddon={rightAddon}
          size={size}
          state={finalState}
        >
          <textarea
            ref={ref}
            id={inputId}
            rows={rows}
            className={cn(
              baseInputClasses,
              textareaSizeClasses[size],
              stateClasses[finalState],
              'resize-vertical min-h-[60px]',
              className
            )}
            {...props}
          />
        </InputWrapper>
        
        {finalHelperText && (
          <InputHelperText state={finalState}>
            {finalHelperText}
          </InputHelperText>
        )}
      </div>
    );
  }
);

Textarea.displayName = 'Textarea';

// Search input component
interface SearchInputProps extends Omit<InputProps, 'leftIcon' | 'type'> {
  onSearch?: (value: string) => void;
  onClear?: () => void;
  showClearButton?: boolean;
}

export const SearchInput = forwardRef<HTMLInputElement, SearchInputProps>(
  ({ 
    onSearch, 
    onClear, 
    showClearButton = true,
    rightIcon,
    ...props 
  }, ref) => {
    const handleKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
      if (e.key === 'Enter' && onSearch) {
        onSearch(e.currentTarget.value);
      }
      props.onKeyDown?.(e);
    };
    
    const handleClear = () => {
      if (onClear) {
        onClear();
      }
    };
    
    const SearchIcon = () => (
      <svg className="w-full h-full" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="m21 21-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
      </svg>
    );
    
    const ClearIcon = () => (
      <button
        type="button"
        onClick={handleClear}
        className="text-text-tertiary hover:text-text-primary transition-colors"
        tabIndex={-1}
      >
        <svg className="w-full h-full" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
        </svg>
      </button>
    );
    
    return (
      <Input
        ref={ref}
        type="search"
        leftIcon={<SearchIcon />}
        rightIcon={showClearButton && props.value ? <ClearIcon /> : rightIcon}
        onKeyDown={handleKeyDown}
        {...props}
      />
    );
  }
);

SearchInput.displayName = 'SearchInput';

// Password input component
interface PasswordInputProps extends Omit<InputProps, 'type' | 'rightIcon'> {
  showPasswordToggle?: boolean;
}

export const PasswordInput = forwardRef<HTMLInputElement, PasswordInputProps>(
  ({ showPasswordToggle = true, ...props }, ref) => {
    const [showPassword, setShowPassword] = React.useState(false);
    
    const togglePasswordVisibility = () => {
      setShowPassword(!showPassword);
    };
    
    const ToggleIcon = () => (
      <button
        type="button"
        onClick={togglePasswordVisibility}
        className="text-text-tertiary hover:text-text-primary transition-colors"
        tabIndex={-1}
        aria-label={showPassword ? 'Hide password' : 'Show password'}
      >
        {showPassword ? (
          <svg className="w-full h-full" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13.875 18.825A10.05 10.05 0 0112 19c-4.478 0-8.268-2.943-9.543-7a9.97 9.97 0 011.563-3.029m5.858.908a3 3 0 114.243 4.243M9.878 9.878l4.242 4.242M9.878 9.878L8.464 8.464m1.414 1.414L8.464 8.464m1.414 1.414l4.242 4.242M8.464 8.464L6.636 6.636m1.828 1.828l4.242 4.242M8.464 8.464l4.242 4.242m0 0L15.535 15.535" />
          </svg>
        ) : (
          <svg className="w-full h-full" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
          </svg>
        )}
      </button>
    );
    
    return (
      <Input
        ref={ref}
        type={showPassword ? 'text' : 'password'}
        rightIcon={showPasswordToggle ? <ToggleIcon /> : undefined}
        {...props}
      />
    );
  }
);

PasswordInput.displayName = 'PasswordInput';

// File input component
interface FileInputProps extends Omit<InputProps, 'type' | 'value' | 'onChange'> {
  accept?: string;
  multiple?: boolean;
  onChange?: (files: FileList | null) => void;
  dragAndDrop?: boolean;
}

export const FileInput = forwardRef<HTMLInputElement, FileInputProps>(
  ({
    accept,
    multiple,
    onChange,
    dragAndDrop = false,
    label,
    className,
    size: _size,
    state: _state,
    leftIcon: _leftIcon,
    rightIcon: _rightIcon,
    leftAddon: _leftAddon,
    rightAddon: _rightAddon,
    fullWidth: _fullWidth,
    helperText: _helperText,
    errorText: _errorText,
    required: _required,
    ...props
  }, ref) => {
    const [isDragOver, setIsDragOver] = React.useState(false);
    const inputRef = React.useRef<HTMLInputElement>(null);
    
    React.useImperativeHandle(ref, () => inputRef.current!);
    
    const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
      onChange?.(e.target.files);
    };
    
    const handleDragOver = (e: React.DragEvent) => {
      e.preventDefault();
      setIsDragOver(true);
    };
    
    const handleDragLeave = (e: React.DragEvent) => {
      e.preventDefault();
      setIsDragOver(false);
    };
    
    const handleDrop = (e: React.DragEvent) => {
      e.preventDefault();
      setIsDragOver(false);
      onChange?.(e.dataTransfer.files);
    };
    
    const handleClick = () => {
      inputRef.current?.click();
    };
    
    if (dragAndDrop) {
      return (
        <div className="form-group">
          {label && <InputLabel>{label}</InputLabel>}
          <div
            className={cn(
              'border-2 border-dashed border-border-primary rounded-lg p-6 text-center cursor-pointer',
              'hover:border-border-secondary transition-colors',
              isDragOver && 'border-primary-500 bg-primary-50 dark:bg-primary-950',
              className
            )}
            onDragOver={handleDragOver}
            onDragLeave={handleDragLeave}
            onDrop={handleDrop}
            onClick={handleClick}
          >
            <input
              ref={inputRef}
              type="file"
              accept={accept}
              multiple={multiple}
              onChange={handleChange}
              className="hidden"
              {...props}
            />
            <div className="text-text-secondary">
              <p>Drop files here or click to browse</p>
              <p className="text-sm text-text-tertiary mt-1">
                {accept && `Accepted formats: ${accept}`}
              </p>
            </div>
          </div>
        </div>
      );
    }
    
    return (
      <Input
        ref={inputRef}
        type="file"
        accept={accept}
        multiple={multiple}
        onChange={handleChange}
        label={label}
        className={className}
        {...props}
      />
    );
  }
);

FileInput.displayName = 'FileInput';