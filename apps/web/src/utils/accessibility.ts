// Accessibility utilities for WCAG AA compliance
// Comprehensive accessibility support for KubeChat Enterprise UI

export type AriaRole =
  | 'alert' | 'alertdialog' | 'application' | 'article' | 'banner' | 'button'
  | 'cell' | 'checkbox' | 'columnheader' | 'combobox' | 'complementary'
  | 'contentinfo' | 'definition' | 'dialog' | 'directory' | 'document'
  | 'form' | 'grid' | 'gridcell' | 'group' | 'heading' | 'img' | 'link'
  | 'list' | 'listbox' | 'listitem' | 'log' | 'main' | 'marquee' | 'math'
  | 'menu' | 'menubar' | 'menuitem' | 'menuitemcheckbox' | 'menuitemradio'
  | 'navigation' | 'note' | 'option' | 'presentation' | 'progressbar'
  | 'radio' | 'radiogroup' | 'region' | 'row' | 'rowgroup' | 'rowheader'
  | 'scrollbar' | 'search' | 'separator' | 'slider' | 'spinbutton'
  | 'status' | 'tab' | 'tablist' | 'tabpanel' | 'textbox' | 'timer'
  | 'toolbar' | 'tooltip' | 'tree' | 'treegrid' | 'treeitem';

export type AriaLevel = 1 | 2 | 3 | 4 | 5 | 6;

export type AriaLive = 'off' | 'polite' | 'assertive';

export type AriaPressed = boolean | 'mixed';

export type AriaExpanded = boolean | 'undefined';

export type AriaSelected = boolean | 'undefined';

export type AriaChecked = boolean | 'mixed' | 'undefined';

export interface AriaAttributes {
  role?: AriaRole;
  'aria-label'?: string;
  'aria-labelledby'?: string;
  'aria-describedby'?: string;
  'aria-expanded'?: AriaExpanded;
  'aria-selected'?: AriaSelected;
  'aria-checked'?: AriaChecked;
  'aria-pressed'?: AriaPressed;
  'aria-disabled'?: boolean;
  'aria-hidden'?: boolean;
  'aria-live'?: AriaLive;
  'aria-atomic'?: boolean;
  'aria-busy'?: boolean;
  'aria-controls'?: string;
  'aria-current'?: boolean | 'page' | 'step' | 'location' | 'date' | 'time';
  'aria-details'?: string;
  'aria-errormessage'?: string;
  'aria-flowto'?: string;
  'aria-haspopup'?: boolean | 'menu' | 'listbox' | 'tree' | 'grid' | 'dialog';
  'aria-invalid'?: boolean | 'grammar' | 'spelling';
  'aria-keyshortcuts'?: string;
  'aria-level'?: AriaLevel;
  'aria-modal'?: boolean;
  'aria-multiline'?: boolean;
  'aria-multiselectable'?: boolean;
  'aria-orientation'?: 'horizontal' | 'vertical';
  'aria-owns'?: string;
  'aria-placeholder'?: string;
  'aria-posinset'?: number;
  'aria-readonly'?: boolean;
  'aria-relevant'?: 'additions' | 'removals' | 'text' | 'all';
  'aria-required'?: boolean;
  'aria-roledescription'?: string;
  'aria-rowcount'?: number;
  'aria-rowindex'?: number;
  'aria-rowspan'?: number;
  'aria-setsize'?: number;
  'aria-sort'?: 'none' | 'ascending' | 'descending' | 'other';
  'aria-valuemax'?: number;
  'aria-valuemin'?: number;
  'aria-valuenow'?: number;
  'aria-valuetext'?: string;
}

// Screen reader utilities
export class ScreenReaderUtils {
  private static announceElement: HTMLElement | null = null;

  // Initialize announcement element
  static initialize(): void {
    if (typeof document === 'undefined') return;

    if (!this.announceElement) {
      this.announceElement = document.createElement('div');
      this.announceElement.setAttribute('aria-live', 'polite');
      this.announceElement.setAttribute('aria-atomic', 'true');
      this.announceElement.className = 'sr-only';
      this.announceElement.style.cssText = `
        position: absolute !important;
        width: 1px !important;
        height: 1px !important;
        padding: 0 !important;
        margin: -1px !important;
        overflow: hidden !important;
        clip: rect(0, 0, 0, 0) !important;
        white-space: nowrap !important;
        border: 0 !important;
      `;
      document.body.appendChild(this.announceElement);
    }
  }

  // Announce message to screen readers
  static announce(message: string, priority: AriaLive = 'polite'): void {
    if (!this.announceElement) {
      this.initialize();
    }

    if (this.announceElement) {
      this.announceElement.setAttribute('aria-live', priority);
      this.announceElement.textContent = message;

      // Clear after announcement to allow re-announcement of same message
      setTimeout(() => {
        if (this.announceElement) {
          this.announceElement.textContent = '';
        }
      }, 1000);
    }
  }

  // Announce error message
  static announceError(message: string): void {
    this.announce(`Error: ${message}`, 'assertive');
  }

  // Announce success message
  static announceSuccess(message: string): void {
    this.announce(`Success: ${message}`, 'polite');
  }

  // Announce loading state
  static announceLoading(message: string = 'Loading'): void {
    this.announce(message, 'polite');
  }

  // Announce command execution
  static announceCommandExecution(command: string): void {
    this.announce(`Executing command: ${command}`, 'polite');
  }

  // Announce command completion
  static announceCommandComplete(result: 'success' | 'error', details?: string): void {
    const message = result === 'success'
      ? `Command completed successfully${details ? ': ' + details : ''}`
      : `Command failed${details ? ': ' + details : ''}`;

    this.announce(message, result === 'error' ? 'assertive' : 'polite');
  }
}

// Color contrast utilities for WCAG AA compliance
export class ColorContrastUtils {
  // Convert hex to RGB
  static hexToRgb(hex: string): { r: number; g: number; b: number } | null {
    const result = /^#?([a-f\d]{2})([a-f\d]{2})([a-f\d]{2})$/i.exec(hex);
    return result ? {
      r: parseInt(result[1], 16),
      g: parseInt(result[2], 16),
      b: parseInt(result[3], 16)
    } : null;
  }

  // Calculate relative luminance
  static getRelativeLuminance(rgb: { r: number; g: number; b: number }): number {
    const { r, g, b } = rgb;

    const [rs, gs, bs] = [r, g, b].map(c => {
      c = c / 255;
      return c <= 0.03928 ? c / 12.92 : Math.pow((c + 0.055) / 1.055, 2.4);
    });

    return 0.2126 * rs + 0.7152 * gs + 0.0722 * bs;
  }

  // Calculate contrast ratio between two colors
  static getContrastRatio(color1: string, color2: string): number {
    const rgb1 = this.hexToRgb(color1);
    const rgb2 = this.hexToRgb(color2);

    if (!rgb1 || !rgb2) return 0;

    const lum1 = this.getRelativeLuminance(rgb1);
    const lum2 = this.getRelativeLuminance(rgb2);

    const brightest = Math.max(lum1, lum2);
    const darkest = Math.min(lum1, lum2);

    return (brightest + 0.05) / (darkest + 0.05);
  }

  // Check if color combination meets WCAG AA standards
  static meetsWCAGAA(foreground: string, background: string, isLargeText = false): boolean {
    const ratio = this.getContrastRatio(foreground, background);
    return isLargeText ? ratio >= 3 : ratio >= 4.5;
  }

  // Check if color combination meets WCAG AAA standards
  static meetsWCAGAAA(foreground: string, background: string, isLargeText = false): boolean {
    const ratio = this.getContrastRatio(foreground, background);
    return isLargeText ? ratio >= 4.5 : ratio >= 7;
  }

  // Get accessible text color for a background
  static getAccessibleTextColor(backgroundColor: string): string {
    const whiteRatio = this.getContrastRatio('#ffffff', backgroundColor);
    const blackRatio = this.getContrastRatio('#000000', backgroundColor);

    return whiteRatio > blackRatio ? '#ffffff' : '#000000';
  }
}

// Focus management utilities
export class FocusManagement {
  private static focusableSelectors = [
    'a[href]',
    'button:not([disabled])',
    'input:not([disabled])',
    'select:not([disabled])',
    'textarea:not([disabled])',
    '[tabindex]:not([tabindex="-1"])',
    '[contenteditable="true"]'
  ].join(', ');

  // Get all focusable elements within a container
  static getFocusableElements(container: HTMLElement): HTMLElement[] {
    return Array.from(container.querySelectorAll(this.focusableSelectors));
  }

  // Get first focusable element
  static getFirstFocusableElement(container: HTMLElement): HTMLElement | null {
    const elements = this.getFocusableElements(container);
    return elements.length > 0 ? elements[0] : null;
  }

  // Get last focusable element
  static getLastFocusableElement(container: HTMLElement): HTMLElement | null {
    const elements = this.getFocusableElements(container);
    return elements.length > 0 ? elements[elements.length - 1] : null;
  }

  // Trap focus within a container (for modals, dropdowns)
  static trapFocus(container: HTMLElement, event: KeyboardEvent): void {
    if (event.key !== 'Tab') return;

    const focusableElements = this.getFocusableElements(container);
    if (focusableElements.length === 0) return;

    const firstElement = focusableElements[0];
    const lastElement = focusableElements[focusableElements.length - 1];

    if (event.shiftKey) {
      // Shift + Tab: going backwards
      if (document.activeElement === firstElement) {
        event.preventDefault();
        lastElement.focus();
      }
    } else {
      // Tab: going forwards
      if (document.activeElement === lastElement) {
        event.preventDefault();
        firstElement.focus();
      }
    }
  }

  // Create focus trap for modal dialogs
  static createFocusTrap(container: HTMLElement): () => void {
    const handleKeyDown = (event: KeyboardEvent) => {
      this.trapFocus(container, event);
    };

    // Focus first focusable element when trap is created
    const firstFocusable = this.getFirstFocusableElement(container);
    if (firstFocusable) {
      firstFocusable.focus();
    }

    document.addEventListener('keydown', handleKeyDown);

    // Return cleanup function
    return () => {
      document.removeEventListener('keydown', handleKeyDown);
    };
  }

  // Restore focus to previously focused element
  static restoreFocus(previouslyFocusedElement: HTMLElement | null): void {
    if (previouslyFocusedElement && previouslyFocusedElement.focus) {
      previouslyFocusedElement.focus();
    }
  }
}

// Keyboard navigation utilities
export class KeyboardNavigation {
  // Standard keyboard event keys
  static readonly KEYS = {
    ENTER: 'Enter',
    SPACE: ' ',
    TAB: 'Tab',
    ESCAPE: 'Escape',
    ARROW_UP: 'ArrowUp',
    ARROW_DOWN: 'ArrowDown',
    ARROW_LEFT: 'ArrowLeft',
    ARROW_RIGHT: 'ArrowRight',
    HOME: 'Home',
    END: 'End',
    PAGE_UP: 'PageUp',
    PAGE_DOWN: 'PageDown'
  } as const;

  // Check if key is an arrow key
  static isArrowKey(key: string): boolean {
    return [this.KEYS.ARROW_UP, this.KEYS.ARROW_DOWN, this.KEYS.ARROW_LEFT, this.KEYS.ARROW_RIGHT].includes(key as any);
  }

  // Handle arrow key navigation in lists
  static handleListNavigation(
    event: KeyboardEvent,
    items: HTMLElement[],
    currentIndex: number,
    onSelect?: (index: number) => void
  ): number {
    let newIndex = currentIndex;

    switch (event.key) {
      case this.KEYS.ARROW_UP:
        event.preventDefault();
        newIndex = currentIndex > 0 ? currentIndex - 1 : items.length - 1;
        break;
      case this.KEYS.ARROW_DOWN:
        event.preventDefault();
        newIndex = currentIndex < items.length - 1 ? currentIndex + 1 : 0;
        break;
      case this.KEYS.HOME:
        event.preventDefault();
        newIndex = 0;
        break;
      case this.KEYS.END:
        event.preventDefault();
        newIndex = items.length - 1;
        break;
      case this.KEYS.ENTER:
      case this.KEYS.SPACE:
        event.preventDefault();
        onSelect?.(currentIndex);
        return currentIndex;
    }

    if (newIndex !== currentIndex && items[newIndex]) {
      items[newIndex].focus();
    }

    return newIndex;
  }

  // Handle horizontal navigation (tabs, breadcrumbs)
  static handleHorizontalNavigation(
    event: KeyboardEvent,
    items: HTMLElement[],
    currentIndex: number
  ): number {
    let newIndex = currentIndex;

    switch (event.key) {
      case this.KEYS.ARROW_LEFT:
        event.preventDefault();
        newIndex = currentIndex > 0 ? currentIndex - 1 : items.length - 1;
        break;
      case this.KEYS.ARROW_RIGHT:
        event.preventDefault();
        newIndex = currentIndex < items.length - 1 ? currentIndex + 1 : 0;
        break;
      case this.KEYS.HOME:
        event.preventDefault();
        newIndex = 0;
        break;
      case this.KEYS.END:
        event.preventDefault();
        newIndex = items.length - 1;
        break;
    }

    if (newIndex !== currentIndex && items[newIndex]) {
      items[newIndex].focus();
    }

    return newIndex;
  }
}

// Accessible component builders
export class AccessibleComponentBuilder {
  // Create accessible button attributes
  static button(options: {
    label?: string;
    labelledBy?: string;
    describedBy?: string;
    pressed?: AriaPressed;
    expanded?: AriaExpanded;
    disabled?: boolean;
    hasPopup?: boolean | 'menu' | 'listbox' | 'tree' | 'grid' | 'dialog';
    controls?: string;
  } = {}): AriaAttributes & { type: 'button' } {
    return {
      type: 'button',
      role: 'button',
      'aria-label': options.label,
      'aria-labelledby': options.labelledBy,
      'aria-describedby': options.describedBy,
      'aria-pressed': options.pressed,
      'aria-expanded': options.expanded,
      'aria-disabled': options.disabled,
      'aria-haspopup': options.hasPopup,
      'aria-controls': options.controls,
    };
  }

  // Create accessible input attributes
  static input(options: {
    label?: string;
    labelledBy?: string;
    describedBy?: string;
    placeholder?: string;
    required?: boolean;
    invalid?: boolean;
    errorMessage?: string;
    readonly?: boolean;
  } = {}): AriaAttributes {
    return {
      'aria-label': options.label,
      'aria-labelledby': options.labelledBy,
      'aria-describedby': options.describedBy,
      'aria-placeholder': options.placeholder,
      'aria-required': options.required,
      'aria-invalid': options.invalid,
      'aria-errormessage': options.errorMessage,
      'aria-readonly': options.readonly,
    };
  }

  // Create accessible dialog attributes
  static dialog(options: {
    label?: string;
    labelledBy?: string;
    describedBy?: string;
    modal?: boolean;
  } = {}): AriaAttributes {
    return {
      role: 'dialog',
      'aria-label': options.label,
      'aria-labelledby': options.labelledBy,
      'aria-describedby': options.describedBy,
      'aria-modal': options.modal,
    };
  }

  // Create accessible menu attributes
  static menu(options: {
    label?: string;
    labelledBy?: string;
    orientation?: 'horizontal' | 'vertical';
  } = {}): AriaAttributes {
    return {
      role: 'menu',
      'aria-label': options.label,
      'aria-labelledby': options.labelledBy,
      'aria-orientation': options.orientation || 'vertical',
    };
  }

  // Create accessible menuitem attributes
  static menuitem(options: {
    label?: string;
    disabled?: boolean;
    hasPopup?: boolean | 'menu' | 'listbox' | 'tree' | 'grid' | 'dialog';
    expanded?: AriaExpanded;
  } = {}): AriaAttributes {
    return {
      role: 'menuitem',
      'aria-label': options.label,
      'aria-disabled': options.disabled,
      'aria-haspopup': options.hasPopup,
      'aria-expanded': options.expanded,
    };
  }

  // Create accessible tab attributes
  static tab(options: {
    label?: string;
    selected?: boolean;
    controls?: string;
    disabled?: boolean;
  } = {}): AriaAttributes {
    return {
      role: 'tab',
      'aria-label': options.label,
      'aria-selected': options.selected,
      'aria-controls': options.controls,
      'aria-disabled': options.disabled,
    };
  }

  // Create accessible tabpanel attributes
  static tabpanel(options: {
    label?: string;
    labelledBy?: string;
  } = {}): AriaAttributes {
    return {
      role: 'tabpanel',
      'aria-label': options.label,
      'aria-labelledby': options.labelledBy,
    };
  }

  // Create accessible status attributes for real-time updates
  static status(options: {
    label?: string;
    live?: AriaLive;
    atomic?: boolean;
  } = {}): AriaAttributes {
    return {
      role: 'status',
      'aria-label': options.label,
      'aria-live': options.live || 'polite',
      'aria-atomic': options.atomic !== false,
    };
  }

  // Create accessible alert attributes
  static alert(options: {
    label?: string;
    live?: AriaLive;
  } = {}): AriaAttributes {
    return {
      role: 'alert',
      'aria-label': options.label,
      'aria-live': options.live || 'assertive',
    };
  }
}

// Accessibility testing utilities
export class AccessibilityTesting {
  // Test color contrast
  static testColorContrast(foreground: string, background: string): {
    ratio: number;
    meetsAA: boolean;
    meetsAAA: boolean;
    aaLargeText: boolean;
    aaaLargeText: boolean;
  } {
    const ratio = ColorContrastUtils.getContrastRatio(foreground, background);

    return {
      ratio,
      meetsAA: ColorContrastUtils.meetsWCAGAA(foreground, background),
      meetsAAA: ColorContrastUtils.meetsWCAGAAA(foreground, background),
      aaLargeText: ColorContrastUtils.meetsWCAGAA(foreground, background, true),
      aaaLargeText: ColorContrastUtils.meetsWCAGAAA(foreground, background, true),
    };
  }

  // Check if element has accessible name
  static hasAccessibleName(element: HTMLElement): boolean {
    return !!(
      element.getAttribute('aria-label') ||
      element.getAttribute('aria-labelledby') ||
      element.getAttribute('title') ||
      (element as HTMLInputElement).placeholder ||
      element.textContent?.trim()
    );
  }

  // Check if interactive element is keyboard accessible
  static isKeyboardAccessible(element: HTMLElement): boolean {
    const tabIndex = element.getAttribute('tabindex');
    const tagName = element.tagName.toLowerCase();

    // Natively focusable elements
    const nativelyFocusable = ['a', 'button', 'input', 'select', 'textarea'].includes(tagName);

    // Elements with explicit tabindex
    const explicitlyFocusable = tabIndex !== null && tabIndex !== '-1';

    return nativelyFocusable || explicitlyFocusable;
  }

  // Basic accessibility audit for an element
  static auditElement(element: HTMLElement): {
    hasAccessibleName: boolean;
    isKeyboardAccessible: boolean;
    hasFocusIndicator: boolean;
    hasValidRole: boolean;
    issues: string[];
  } {
    const issues: string[] = [];

    const hasName = this.hasAccessibleName(element);
    if (!hasName) {
      issues.push('Element lacks accessible name');
    }

    const isKeyboardAccessible = this.isKeyboardAccessible(element);
    if (!isKeyboardAccessible && element.onclick) {
      issues.push('Interactive element is not keyboard accessible');
    }

    // Check for focus indicators (simplified check)
    const computedStyle = window.getComputedStyle(element);
    const hasFocusIndicator = computedStyle.outline !== 'none' ||
                             computedStyle.boxShadow.includes('inset') ||
                             element.classList.contains('focus:');

    if (isKeyboardAccessible && !hasFocusIndicator) {
      issues.push('Focusable element lacks visible focus indicator');
    }

    const role = element.getAttribute('role');
    const hasValidRole = !role || this.isValidAriaRole(role);
    if (!hasValidRole) {
      issues.push('Element has invalid ARIA role');
    }

    return {
      hasAccessibleName: hasName,
      isKeyboardAccessible,
      hasFocusIndicator,
      hasValidRole,
      issues,
    };
  }

  // Check if ARIA role is valid
  private static isValidAriaRole(role: string): boolean {
    const validRoles = [
      'alert', 'alertdialog', 'application', 'article', 'banner', 'button',
      'cell', 'checkbox', 'columnheader', 'combobox', 'complementary',
      'contentinfo', 'definition', 'dialog', 'directory', 'document',
      'form', 'grid', 'gridcell', 'group', 'heading', 'img', 'link',
      'list', 'listbox', 'listitem', 'log', 'main', 'marquee', 'math',
      'menu', 'menubar', 'menuitem', 'menuitemcheckbox', 'menuitemradio',
      'navigation', 'note', 'option', 'presentation', 'progressbar',
      'radio', 'radiogroup', 'region', 'row', 'rowgroup', 'rowheader',
      'scrollbar', 'search', 'separator', 'slider', 'spinbutton',
      'status', 'tab', 'tablist', 'tabpanel', 'textbox', 'timer',
      'toolbar', 'tooltip', 'tree', 'treegrid', 'treeitem'
    ];

    return validRoles.includes(role);
  }
}

// Initialize screen reader utilities when module loads
if (typeof window !== 'undefined') {
  ScreenReaderUtils.initialize();
}

