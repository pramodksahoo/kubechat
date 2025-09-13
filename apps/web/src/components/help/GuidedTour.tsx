import React, { useState, useEffect, useRef, useCallback, ReactNode } from 'react';
import { createPortal } from 'react-dom';
import { Button, Heading, Body, StatusBadge } from '../../design-system';
import { cn } from '../../lib/utils';
import { ScreenReaderUtils, AriaUtils } from '../../design-system/accessibility';
import { useKeyboardShortcuts } from '../../design-system/hooks/useKeyboardNavigation';
import { trackEvent } from '../../lib/monitoring';

// Types
type PopoverPosition = 'top' | 'bottom' | 'left' | 'right' | 'center';

// Tour step configuration
export interface TourStep {
  id: string;
  target: string; // CSS selector or element ID
  title: string;
  content: ReactNode;
  position?: 'top' | 'bottom' | 'left' | 'right' | 'center';
  spotlight?: boolean;
  action?: TourAction;
  waitFor?: WaitCondition;
  skippable?: boolean;
  optional?: boolean;
  category?: 'navigation' | 'feature' | 'workflow' | 'safety';
}

export interface TourAction {
  type: 'click' | 'input' | 'scroll' | 'wait' | 'custom';
  description?: string;
  element?: string;
  value?: string;
  timeout?: number;
  customAction?: () => Promise<void>;
}

export interface WaitCondition {
  type: 'element' | 'timeout' | 'event' | 'custom';
  selector?: string;
  timeout?: number;
  event?: string;
  condition?: () => boolean;
}

// Tour configuration
export interface TourConfig {
  id: string;
  title: string;
  description: string;
  steps: TourStep[];
  autoStart?: boolean;
  showProgress?: boolean;
  allowSkip?: boolean;
  restartable?: boolean;
  category?: 'onboarding' | 'feature_discovery' | 'troubleshooting';
  prerequisites?: string[];
  estimatedDuration?: number; // in minutes
}

// Tour state
export interface TourState {
  currentStep: number;
  isActive: boolean;
  completedSteps: string[];
  skippedSteps: string[];
  startTime: Date;
  pausedTime?: Date;
  isPaused: boolean;
}

// Spotlight overlay component
function SpotlightOverlay({ 
  targetElement, 
  onClose 
}: { 
  targetElement: Element | null;
  onClose: () => void;
}) {
  const [spotlightStyle, setSpotlightStyle] = useState<React.CSSProperties>({});

  useEffect(() => {
    if (!targetElement) return;

    const updateSpotlight = () => {
      const rect = targetElement.getBoundingClientRect();
      const padding = 8;
      
      setSpotlightStyle({
        clipPath: `polygon(
          0 0,
          0 100%,
          ${rect.left - padding}px 100%,
          ${rect.left - padding}px ${rect.top - padding}px,
          ${rect.right + padding}px ${rect.top - padding}px,
          ${rect.right + padding}px ${rect.bottom + padding}px,
          ${rect.left - padding}px ${rect.bottom + padding}px,
          ${rect.left - padding}px 100%,
          100% 100%,
          100% 0
        )`,
      });
    };

    updateSpotlight();
    window.addEventListener('resize', updateSpotlight);
    window.addEventListener('scroll', updateSpotlight);

    return () => {
      window.removeEventListener('resize', updateSpotlight);
      window.removeEventListener('scroll', updateSpotlight);
    };
  }, [targetElement]);

  return (
    <div
      className="fixed inset-0 z-[9998] bg-black/50 transition-all duration-300"
      style={spotlightStyle}
      onClick={onClose}
      aria-hidden="true"
    />
  );
}

// Tour step component
function TourStepPopover({
  step,
  stepIndex,
  totalSteps,
  targetElement,
  onNext,
  onPrevious,
  onSkip,
  onClose,
  isFirst,
  isLast,
}: {
  step: TourStep;
  stepIndex: number;
  totalSteps: number;
  targetElement: Element | null;
  onNext: () => void;
  onPrevious: () => void;
  onSkip: () => void;
  onClose: () => void;
  isFirst: boolean;
  isLast: boolean;
}) {
  const [position, setPosition] = useState({ x: 0, y: 0 });
  const [popoverPosition, setPopoverPosition] = useState<PopoverPosition>('bottom');
  const popoverRef = useRef<HTMLDivElement>(null);
  const popoverId = useRef(AriaUtils.generateId('tour-step'));

  // Calculate popover position
  useEffect(() => {
    if (!targetElement || !popoverRef.current) return;

    const targetRect = targetElement.getBoundingClientRect();
    const popoverRect = popoverRef.current.getBoundingClientRect();
    const viewport = { width: window.innerWidth, height: window.innerHeight };
    const offset = 16;

    let finalPosition = step.position || 'bottom';
    let x = 0;
    let y = 0;

    // Auto-detect position if not specified or if specified position doesn't fit
    if (step.position === 'center') {
      x = (viewport.width - popoverRect.width) / 2;
      y = (viewport.height - popoverRect.height) / 2;
      setPopoverPosition('center' as PopoverPosition);
    } else {
      // Calculate based on available space
      const spaceBottom = viewport.height - targetRect.bottom;
      const spaceTop = targetRect.top;
      const spaceRight = viewport.width - targetRect.right;
      const spaceLeft = targetRect.left;

      if (finalPosition === 'bottom' && spaceBottom < popoverRect.height + offset) {
        finalPosition = spaceTop > spaceBottom ? 'top' : 'right';
      } else if (finalPosition === 'top' && spaceTop < popoverRect.height + offset) {
        finalPosition = spaceBottom > spaceTop ? 'bottom' : 'right';
      } else if (finalPosition === 'right' && spaceRight < popoverRect.width + offset) {
        finalPosition = spaceLeft > spaceRight ? 'left' : 'bottom';
      } else if (finalPosition === 'left' && spaceLeft < popoverRect.width + offset) {
        finalPosition = spaceRight > spaceLeft ? 'right' : 'bottom';
      }

      // Calculate coordinates
      switch (finalPosition) {
        case 'top':
          x = targetRect.left + (targetRect.width - popoverRect.width) / 2;
          y = targetRect.top - popoverRect.height - offset;
          break;
        case 'bottom':
          x = targetRect.left + (targetRect.width - popoverRect.width) / 2;
          y = targetRect.bottom + offset;
          break;
        case 'left':
          x = targetRect.left - popoverRect.width - offset;
          y = targetRect.top + (targetRect.height - popoverRect.height) / 2;
          break;
        case 'right':
          x = targetRect.right + offset;
          y = targetRect.top + (targetRect.height - popoverRect.height) / 2;
          break;
      }

      // Keep in viewport
      x = Math.max(8, Math.min(x, viewport.width - popoverRect.width - 8));
      y = Math.max(8, Math.min(y, viewport.height - popoverRect.height - 8));

      setPopoverPosition(finalPosition as PopoverPosition);
    }

    setPosition({ x, y });
  }, [targetElement, step.position]);

  // Scroll target into view
  useEffect(() => {
    if (targetElement) {
      targetElement.scrollIntoView({
        behavior: 'smooth',
        block: 'center',
        inline: 'center',
      });
    }
  }, [targetElement]);

  return createPortal(
    <div
      ref={popoverRef}
      id={popoverId.current}
      role="dialog"
      aria-labelledby={`${popoverId.current}-title`}
      aria-describedby={`${popoverId.current}-content`}
      className="fixed z-[9999] bg-background-primary border border-border-primary rounded-lg shadow-modal max-w-sm animate-scale-in"
      style={{ left: position.x, top: position.y }}
    >
      {/* Arrow */}
      <div
        className={cn(
          'absolute w-3 h-3 bg-background-primary border transform rotate-45',
          popoverPosition === 'top' && 'bottom-0 left-1/2 -translate-x-1/2 translate-y-1/2 border-t-0 border-l-0',
          popoverPosition === 'bottom' && 'top-0 left-1/2 -translate-x-1/2 -translate-y-1/2 border-b-0 border-r-0',
          popoverPosition === 'left' && 'right-0 top-1/2 translate-x-1/2 -translate-y-1/2 border-l-0 border-b-0',
          popoverPosition === 'right' && 'left-0 top-1/2 -translate-x-1/2 -translate-y-1/2 border-r-0 border-t-0',
        )}
      />

      <div className="p-6">
        {/* Header */}
        <div className="flex items-start justify-between mb-4">
          <div className="flex-1">
            <div className="flex items-center gap-3 mb-2">
              <Heading level={4} id={`${popoverId.current}-title`}>
                {step.title}
              </Heading>
              <StatusBadge variant="info">
                {stepIndex + 1}/{totalSteps}
              </StatusBadge>
            </div>
            
            {step.category && (
              <StatusBadge variant="neutral" className="text-xs">
                {step.category.replace('_', ' ')}
              </StatusBadge>
            )}
          </div>

          <Button variant="ghost" size="sm" onClick={onClose} aria-label="Close tour">
            <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
            </svg>
          </Button>
        </div>

        {/* Content */}
        <div id={`${popoverId.current}-content`} className="mb-6">
          {typeof step.content === 'string' ? (
            <Body>{step.content}</Body>
          ) : (
            step.content
          )}
        </div>

        {/* Action hint */}
        {step.action && (
          <div className="mb-4 p-3 bg-info-50 dark:bg-info-900/20 border border-info-200 dark:border-info-800 rounded-lg">
            <Body size="sm" color="info">
              <strong>Try it:</strong> {step.action.description}
            </Body>
          </div>
        )}

        {/* Footer */}
        <div className="flex items-center justify-between">
          <div className="flex gap-2">
            <Button
              variant="outline"
              size="sm"
              onClick={onPrevious}
              disabled={isFirst}
            >
              Previous
            </Button>
            
            {step.skippable && (
              <Button variant="ghost" size="sm" onClick={onSkip}>
                Skip
              </Button>
            )}
          </div>

          <Button variant="primary" size="sm" onClick={onNext}>
            {isLast ? 'Finish' : 'Next'}
          </Button>
        </div>
      </div>
    </div>,
    document.body
  );
}

// Main guided tour component
export interface GuidedTourProps {
  config: TourConfig;
  isActive: boolean;
  onComplete: (state: TourState) => void;
  onExit: () => void;
  onPause?: () => void;
  onResume?: () => void;
  className?: string;
}

export function GuidedTour({
  config,
  isActive,
  onComplete,
  onExit,
  onPause,
  onResume,
}: GuidedTourProps) {
  const [state, setState] = useState<TourState>({
    currentStep: 0,
    isActive: false,
    completedSteps: [],
    skippedSteps: [],
    startTime: new Date(),
    isPaused: false,
  });

  const [targetElement, setTargetElement] = useState<Element | null>(null);
  const [waitingForCondition, setWaitingForCondition] = useState(false);
  const currentStep = config.steps[state.currentStep];

  // Keyboard shortcuts
  useKeyboardShortcuts({
    'escape': () => {
      if (!state.isPaused) {
        handleExit();
      }
    },
    'arrowleft': () => {
      if (!state.isPaused && state.currentStep > 0) {
        handlePrevious();
      }
    },
    'arrowright': () => {
      if (!state.isPaused) {
        handleNext();
      }
    },
    'space': (e) => {
      e.preventDefault();
      if (state.isPaused) {
        handleResume();
      } else {
        handlePause();
      }
    },
  }, isActive);

  // Find target element
  useEffect(() => {
    if (!isActive || !currentStep) return;

    const findElement = () => {
      const element = document.querySelector(currentStep.target);
      if (element) {
        setTargetElement(element);
        
        // Announce step to screen readers
        ScreenReaderUtils.announce(
          `Tour step ${state.currentStep + 1}: ${currentStep.title}`,
          'polite'
        );
        
        return true;
      }
      return false;
    };

    if (!findElement()) {
      // If element not found, try again after a short delay
      const timeout = setTimeout(findElement, 500);
      return () => clearTimeout(timeout);
    }
  }, [isActive, currentStep, state.currentStep]);

  // Wait for conditions
  useEffect(() => {
    if (!currentStep?.waitFor || !isActive) return;

    const { waitFor } = currentStep;
    setWaitingForCondition(true);

    const checkCondition = async () => {
      switch (waitFor.type) {
        case 'element':
          if (waitFor.selector) {
            const element = document.querySelector(waitFor.selector);
            if (element) {
              setWaitingForCondition(false);
            }
          }
          break;

        case 'timeout':
          if (waitFor.timeout) {
            setTimeout(() => {
              setWaitingForCondition(false);
            }, waitFor.timeout);
          }
          break;

        case 'event':
          if (waitFor.event) {
            const handler = () => {
              setWaitingForCondition(false);
              document.removeEventListener(waitFor.event!, handler);
            };
            document.addEventListener(waitFor.event, handler);
          }
          break;

        case 'custom':
          if (waitFor.condition) {
            const interval = setInterval(() => {
              if (waitFor.condition!()) {
                setWaitingForCondition(false);
                clearInterval(interval);
              }
            }, 100);
          }
          break;
      }
    };

    checkCondition();
  }, [currentStep, isActive]);

  // Initialize tour
  useEffect(() => {
    if (isActive && !state.isActive) {
      setState(prev => ({
        ...prev,
        isActive: true,
        startTime: new Date(),
      }));
      
      trackEvent('tour_started', {
        tourId: config.id,
        category: config.category,
        estimatedDuration: config.estimatedDuration,
      });
    }
  }, [isActive, state.isActive, config]);

  const handleNext = useCallback(async () => {
    if (!currentStep || waitingForCondition) return;

    // Execute step action if present
    if (currentStep.action) {
      try {
        await executeStepAction(currentStep.action);
      } catch (error) {
        console.error('Failed to execute step action:', error);
      }
    }

    // Track step completion
    trackEvent('tour_step_completed', {
      tourId: config.id,
      stepId: currentStep.id,
      stepIndex: state.currentStep,
    });

    setState(prev => ({
      ...prev,
      completedSteps: [...prev.completedSteps, currentStep.id],
    }));

    if (state.currentStep === config.steps.length - 1) {
      // Tour completed
      const finalState = {
        ...state,
        completedSteps: [...state.completedSteps, currentStep.id],
      };
      
      trackEvent('tour_completed', {
        tourId: config.id,
        completedSteps: finalState.completedSteps.length,
        skippedSteps: finalState.skippedSteps.length,
        duration: Date.now() - finalState.startTime.getTime(),
      });

      onComplete(finalState);
    } else {
      setState(prev => ({ ...prev, currentStep: prev.currentStep + 1 }));
    }
  }, [currentStep, waitingForCondition, state, config, onComplete]);

  const handlePrevious = useCallback(() => {
    if (state.currentStep > 0) {
      setState(prev => ({ ...prev, currentStep: prev.currentStep - 1 }));
      trackEvent('tour_step_previous', {
        tourId: config.id,
        stepIndex: state.currentStep,
      });
    }
  }, [state.currentStep, config.id]);

  const handleSkip = useCallback(() => {
    if (!currentStep) return;

    trackEvent('tour_step_skipped', {
      tourId: config.id,
      stepId: currentStep.id,
      stepIndex: state.currentStep,
    });

    setState(prev => ({
      ...prev,
      skippedSteps: [...prev.skippedSteps, currentStep.id],
    }));

    if (state.currentStep === config.steps.length - 1) {
      onComplete(state);
    } else {
      setState(prev => ({ ...prev, currentStep: prev.currentStep + 1 }));
    }
  }, [currentStep, state, config.id, config.steps.length, onComplete]);

  const handleExit = useCallback(() => {
    trackEvent('tour_exited', {
      tourId: config.id,
      stepIndex: state.currentStep,
      completedSteps: state.completedSteps.length,
      duration: Date.now() - state.startTime.getTime(),
    });
    onExit();
  }, [config.id, state, onExit]);

  const handlePause = useCallback(() => {
    setState(prev => ({ ...prev, isPaused: true, pausedTime: new Date() }));
    onPause?.();
    trackEvent('tour_paused', { tourId: config.id, stepIndex: state.currentStep });
  }, [config.id, state.currentStep, onPause]);

  const handleResume = useCallback(() => {
    setState(prev => ({ ...prev, isPaused: false, pausedTime: undefined }));
    onResume?.();
    trackEvent('tour_resumed', { tourId: config.id, stepIndex: state.currentStep });
  }, [config.id, state.currentStep, onResume]);

  if (!isActive || !currentStep || !targetElement) {
    return null;
  }

  return (
    <>
      {/* Spotlight overlay */}
      {currentStep.spotlight && (
        <SpotlightOverlay targetElement={targetElement} onClose={handleExit} />
      )}

      {/* Tour step popover */}
      <TourStepPopover
        step={currentStep}
        stepIndex={state.currentStep}
        totalSteps={config.steps.length}
        targetElement={targetElement}
        onNext={handleNext}
        onPrevious={handlePrevious}
        onSkip={handleSkip}
        onClose={handleExit}
        isFirst={state.currentStep === 0}
        isLast={state.currentStep === config.steps.length - 1}
      />

      {/* Pause overlay */}
      {state.isPaused && (
        <div className="fixed inset-0 z-[10000] flex items-center justify-center bg-black/75 backdrop-blur-sm">
          <div className="bg-background-primary p-6 rounded-lg shadow-modal max-w-sm text-center">
            <Heading level={3} className="mb-4">Tour Paused</Heading>
            <Body className="mb-6">
              Press space to resume the tour, or escape to exit.
            </Body>
            <div className="flex gap-3 justify-center">
              <Button variant="primary" onClick={handleResume}>
                Resume Tour
              </Button>
              <Button variant="outline" onClick={handleExit}>
                Exit Tour
              </Button>
            </div>
          </div>
        </div>
      )}
    </>
  );
}

// Execute step actions
async function executeStepAction(action: TourAction): Promise<void> {
  switch (action.type) {
    case 'click':
      if (action.element) {
        const element = document.querySelector(action.element) as HTMLElement;
        if (element) {
          element.click();
        }
      }
      break;

    case 'input':
      if (action.element && action.value) {
        const element = document.querySelector(action.element) as HTMLInputElement;
        if (element) {
          element.value = action.value;
          element.dispatchEvent(new Event('input', { bubbles: true }));
        }
      }
      break;

    case 'scroll':
      if (action.element) {
        const element = document.querySelector(action.element);
        if (element) {
          element.scrollIntoView({ behavior: 'smooth' });
        }
      }
      break;

    case 'wait':
      if (action.timeout) {
        await new Promise(resolve => setTimeout(resolve, action.timeout));
      }
      break;

    case 'custom':
      if (action.customAction) {
        await action.customAction();
      }
      break;
  }
}

// Hook for managing tours
export function useGuidedTour() {
  const [activeTour, setActiveTour] = useState<TourConfig | null>(null);
  const [isActive, setIsActive] = useState(false);

  const startTour = useCallback((config: TourConfig) => {
    setActiveTour(config);
    setIsActive(true);
  }, []);

  const completeTour = useCallback(() => {
    setIsActive(false);
    setActiveTour(null);
  }, []);

  const exitTour = useCallback(() => {
    setIsActive(false);
    setActiveTour(null);
  }, []);

  return {
    activeTour,
    isActive,
    startTour,
    completeTour,
    exitTour,
  };
}