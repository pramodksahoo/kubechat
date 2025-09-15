import React, { useState, useEffect, useCallback, useRef } from 'react';
import { Button, Heading, Body, StatusBadge } from '../../design-system';
import { cn } from '../../lib/utils';
import { useFocusTrap, useKeyboardShortcuts } from '../../design-system/hooks/useKeyboardNavigation';
import { ScreenReaderUtils } from '../../design-system/accessibility';
import { trackEvent } from '../../lib/monitoring';

// Onboarding step types
export interface OnboardingStep {
  id: string;
  title: string;
  description: string;
  content: React.ReactNode;
  actions?: OnboardingAction[];
  validation?: () => boolean | Promise<boolean>;
  skippable?: boolean;
  required?: boolean;
  category?: 'setup' | 'introduction' | 'feature' | 'completion';
  estimatedTime?: number; // in minutes
  prerequisites?: string[];
}

export interface OnboardingAction {
  label: string;
  action: () => void | Promise<void>;
  variant?: 'primary' | 'secondary' | 'outline';
  type?: 'next' | 'skip' | 'custom';
  loading?: boolean;
  disabled?: boolean;
}

// Onboarding context
export interface OnboardingContext {
  userId: string;
  userRole: 'admin' | 'developer' | 'viewer';
  organizationSize: 'small' | 'medium' | 'large' | 'enterprise';
  experience: 'beginner' | 'intermediate' | 'expert';
  goals: string[];
  clusterAccess: boolean;
  preferences: {
    skipIntroductions: boolean;
    enableTours: boolean;
    showKeyboardShortcuts: boolean;
  };
}

// Progress tracking
export interface OnboardingProgress {
  currentStep: number;
  completedSteps: string[];
  skippedSteps: string[];
  startTime: Date;
  estimatedCompletion?: Date;
  timeSpent: number; // in seconds
}

// Main onboarding flow component
export interface OnboardingFlowProps {
  steps: OnboardingStep[];
  context: OnboardingContext;
  onComplete: (progress: OnboardingProgress) => void;
  onExit: () => void;
  autoSave?: boolean;
  className?: string;
}

export function OnboardingFlow({
  steps,
  context,
  onComplete,
  onExit,
  autoSave = true,
  className,
}: OnboardingFlowProps) {
  const [currentStepIndex, setCurrentStepIndex] = useState(0);
  const [progress, setProgress] = useState<OnboardingProgress>({
    currentStep: 0,
    completedSteps: [],
    skippedSteps: [],
    startTime: new Date(),
    timeSpent: 0,
  });
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const containerRef = useRef<HTMLDivElement>(null);
  const startTimeRef = useRef(Date.now());
  const timeTrackingRef = useRef<NodeJS.Timeout>();

  // Focus trap for modal-like behavior
  useFocusTrap(true, containerRef);

  // Keyboard shortcuts
  useKeyboardShortcuts({
    'escape': () => {
      trackEvent('onboarding_exit_keyboard', { step: currentStep.id, progress: currentStepIndex });
      onExit();
    },
    'enter': () => {
      if (!loading) {
        handleNext();
      }
    },
    'alt+s': () => {
      if (currentStep.skippable) {
        handleSkip();
      }
    },
  });

  const currentStep = steps[currentStepIndex];
  const isLastStep = currentStepIndex === steps.length - 1;
  const completionPercentage = Math.round(((currentStepIndex + 1) / steps.length) * 100);

  // Time tracking
  useEffect(() => {
    timeTrackingRef.current = setInterval(() => {
      setProgress(prev => ({
        ...prev,
        timeSpent: Math.floor((Date.now() - startTimeRef.current) / 1000),
      }));
    }, 1000);

    return () => {
      if (timeTrackingRef.current) {
        clearInterval(timeTrackingRef.current);
      }
    };
  }, []);

  // Auto-save progress
  useEffect(() => {
    if (autoSave) {
      const savedProgress = {
        ...progress,
        currentStep: currentStepIndex,
      };
      localStorage.setItem('onboarding_progress', JSON.stringify(savedProgress));
    }
  }, [currentStepIndex, progress, autoSave]);

  // Announce step changes to screen readers
  useEffect(() => {
    if (currentStep) {
      ScreenReaderUtils.announce(
        `Step ${currentStepIndex + 1} of ${steps.length}: ${currentStep.title}`,
        'polite'
      );
    }
  }, [currentStepIndex, currentStep, steps.length]);

  // Load saved progress
  useEffect(() => {
    if (autoSave) {
      try {
        const saved = localStorage.getItem('onboarding_progress');
        if (saved) {
          const savedProgress = JSON.parse(saved);
          setProgress(savedProgress);
          setCurrentStepIndex(savedProgress.currentStep || 0);
        }
      } catch (error) {
        console.warn('Failed to load onboarding progress:', error);
      }
    }
  }, [autoSave]);

  // Handle next step
  const handleNext = useCallback(async () => {
    if (!currentStep) return;

    setLoading(true);
    setError(null);

    try {
      // Validate current step if needed
      if (currentStep.validation) {
        const isValid = await currentStep.validation();
        if (!isValid) {
          setError('Please complete all required fields before continuing.');
          setLoading(false);
          return;
        }
      }

      // Track step completion
      trackEvent('onboarding_step_completed', {
        stepId: currentStep.id,
        stepIndex: currentStepIndex,
        timeSpent: progress.timeSpent,
        context,
      });

      // Update progress
      setProgress(prev => ({
        ...prev,
        completedSteps: [...prev.completedSteps, currentStep.id],
      }));

      if (isLastStep) {
        // Complete onboarding
        const finalProgress = {
          ...progress,
          completedSteps: [...progress.completedSteps, currentStep.id],
        };
        
        trackEvent('onboarding_completed', {
          totalTime: finalProgress.timeSpent,
          completedSteps: finalProgress.completedSteps.length,
          skippedSteps: finalProgress.skippedSteps.length,
          context,
        });

        onComplete(finalProgress);
      } else {
        setCurrentStepIndex(prev => prev + 1);
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'An error occurred');
    } finally {
      setLoading(false);
    }
  }, [currentStep, currentStepIndex, isLastStep, progress, context, onComplete]);

  // Handle skip step
  const handleSkip = useCallback(() => {
    if (!currentStep?.skippable) return;

    trackEvent('onboarding_step_skipped', {
      stepId: currentStep.id,
      stepIndex: currentStepIndex,
      context,
    });

    setProgress(prev => ({
      ...prev,
      skippedSteps: [...prev.skippedSteps, currentStep.id],
    }));

    if (isLastStep) {
      onComplete(progress);
    } else {
      setCurrentStepIndex(prev => prev + 1);
    }
  }, [currentStep, currentStepIndex, isLastStep, progress, context, onComplete]);

  // Handle previous step
  const handlePrevious = useCallback(() => {
    if (currentStepIndex > 0) {
      setCurrentStepIndex(prev => prev - 1);
      trackEvent('onboarding_step_previous', {
        stepId: currentStep.id,
        stepIndex: currentStepIndex,
      });
    }
  }, [currentStepIndex, currentStep]);

  // Handle exit
  const handleExit = useCallback(() => {
    trackEvent('onboarding_exited', {
      stepId: currentStep?.id,
      stepIndex: currentStepIndex,
      progress: completionPercentage,
      timeSpent: progress.timeSpent,
    });
    onExit();
  }, [currentStep, currentStepIndex, completionPercentage, progress.timeSpent, onExit]);

  if (!currentStep) {
    return null;
  }

  return (
    <div className={cn(
      'fixed inset-0 z-modal flex items-center justify-center p-4 bg-black/50 backdrop-blur-sm',
      className
    )}>
      <div 
        ref={containerRef}
        className="relative w-full max-w-2xl bg-background-primary rounded-modal shadow-modal max-h-[90vh] overflow-hidden"
        role="dialog"
        aria-labelledby="onboarding-title"
        aria-describedby="onboarding-description"
      >
        {/* Header */}
        <div className="flex items-center justify-between p-6 border-b border-border-primary">
          <div className="flex items-center gap-4">
            <Heading level={2} id="onboarding-title">
              {currentStep.title}
            </Heading>
            <StatusBadge variant="info">
              Step {currentStepIndex + 1} of {steps.length}
            </StatusBadge>
          </div>
          
          <Button variant="ghost" size="sm" onClick={handleExit} aria-label="Exit onboarding">
            <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
            </svg>
          </Button>
        </div>

        {/* Progress bar */}
        <div className="px-6 py-2 bg-background-secondary">
          <div className="flex items-center justify-between mb-2">
            <Body size="sm" color="secondary">
              Progress: {completionPercentage}%
            </Body>
            {currentStep.estimatedTime && (
              <Body size="sm" color="tertiary">
                ~{currentStep.estimatedTime} min
              </Body>
            )}
          </div>
          <div className="w-full bg-gray-200 dark:bg-gray-700 rounded-full h-2">
            <div 
              className="bg-primary-600 h-2 rounded-full transition-all duration-300"
              style={{ width: `${completionPercentage}%` }}
              role="progressbar"
              aria-valuenow={completionPercentage}
              aria-valuemin={0}
              aria-valuemax={100}
              aria-label={`Onboarding progress: ${completionPercentage}%`}
            />
          </div>
        </div>

        {/* Content */}
        <div className="p-6 overflow-y-auto max-h-[50vh]">
          <Body id="onboarding-description" className="mb-6">
            {currentStep.description}
          </Body>
          
          {error && (
            <div className="mb-4 p-3 bg-error-50 dark:bg-error-900/20 border border-error-200 dark:border-error-800 rounded-lg">
              <Body size="sm" color="error">
                {error}
              </Body>
            </div>
          )}

          <div className="onboarding-step-content">
            {currentStep.content}
          </div>
        </div>

        {/* Footer */}
        <div className="flex items-center justify-between p-6 border-t border-border-primary bg-background-secondary">
          <div className="flex items-center gap-3">
            <Button
              variant="outline"
              onClick={handlePrevious}
              disabled={currentStepIndex === 0}
            >
              Previous
            </Button>
            
            {currentStep.skippable && (
              <Button variant="ghost" onClick={handleSkip}>
                Skip{' '}
                <kbd className="ml-1 px-1.5 py-0.5 text-xs bg-gray-200 dark:bg-gray-700 rounded">
                  Alt+S
                </kbd>
              </Button>
            )}
          </div>

          <div className="flex items-center gap-3">
            {/* Custom actions */}
            {currentStep.actions?.map((action, index) => (
              <Button
                key={index}
                variant={action.variant || 'secondary'}
                onClick={action.action}
                loading={action.loading}
                disabled={action.disabled}
              >
                {action.label}
              </Button>
            ))}

            {/* Next/Complete button */}
            <Button
              variant="primary"
              onClick={handleNext}
              loading={loading}
              disabled={currentStep.required && error !== null}
            >
              {isLastStep ? 'Complete Setup' : 'Next'}
              {!isLastStep && (
                <kbd className="ml-2 px-1.5 py-0.5 text-xs bg-primary-700 rounded">
                  Enter
                </kbd>
              )}
            </Button>
          </div>
        </div>

        {/* Keyboard shortcuts hint */}
        <div className="px-6 py-2 bg-background-tertiary border-t border-border-primary">
          <Body size="sm" color="tertiary" className="text-center">
            Press <kbd className="px-1 bg-gray-200 dark:bg-gray-700 rounded text-xs">Esc</kbd> to exit • 
            <kbd className="px-1 bg-gray-200 dark:bg-gray-700 rounded text-xs mx-1">Enter</kbd> to continue
            {currentStep.skippable && (
              <> • <kbd className="px-1 bg-gray-200 dark:bg-gray-700 rounded text-xs">Alt+S</kbd> to skip</>
            )}
          </Body>
        </div>
      </div>
    </div>
  );
}

// Onboarding step builder utility
export class OnboardingStepBuilder {
  private step: Partial<OnboardingStep> = {};

  static create(id: string): OnboardingStepBuilder {
    return new OnboardingStepBuilder().setId(id);
  }

  setId(id: string): OnboardingStepBuilder {
    this.step.id = id;
    return this;
  }

  setTitle(title: string): OnboardingStepBuilder {
    this.step.title = title;
    return this;
  }

  setDescription(description: string): OnboardingStepBuilder {
    this.step.description = description;
    return this;
  }

  setContent(content: React.ReactNode): OnboardingStepBuilder {
    this.step.content = content;
    return this;
  }

  setCategory(category: OnboardingStep['category']): OnboardingStepBuilder {
    this.step.category = category;
    return this;
  }

  setEstimatedTime(minutes: number): OnboardingStepBuilder {
    this.step.estimatedTime = minutes;
    return this;
  }

  setSkippable(skippable: boolean = true): OnboardingStepBuilder {
    this.step.skippable = skippable;
    return this;
  }

  setRequired(required: boolean = true): OnboardingStepBuilder {
    this.step.required = required;
    return this;
  }

  setValidation(validation: () => boolean | Promise<boolean>): OnboardingStepBuilder {
    this.step.validation = validation;
    return this;
  }

  addAction(action: OnboardingAction): OnboardingStepBuilder {
    if (!this.step.actions) {
      this.step.actions = [];
    }
    this.step.actions.push(action);
    return this;
  }

  setPrerequisites(prerequisites: string[]): OnboardingStepBuilder {
    this.step.prerequisites = prerequisites;
    return this;
  }

  build(): OnboardingStep {
    if (!this.step.id || !this.step.title || !this.step.description || !this.step.content) {
      throw new Error('Missing required fields: id, title, description, and content are required');
    }
    return this.step as OnboardingStep;
  }
}

// Hook for managing onboarding state
export function useOnboarding() {
  const [isActive, setIsActive] = useState(false);
  const [context, setContext] = useState<OnboardingContext | null>(null);
  const [progress, setProgress] = useState<OnboardingProgress | null>(null);

  const startOnboarding = useCallback((userContext: OnboardingContext) => {
    setContext(userContext);
    setIsActive(true);
    trackEvent('onboarding_started', { context: userContext });
  }, []);

  const completeOnboarding = useCallback((finalProgress: OnboardingProgress) => {
    setProgress(finalProgress);
    setIsActive(false);
    localStorage.removeItem('onboarding_progress');
    trackEvent('onboarding_completed', { progress: finalProgress });
  }, []);

  const exitOnboarding = useCallback(() => {
    setIsActive(false);
    trackEvent('onboarding_exited');
  }, []);

  const resetOnboarding = useCallback(() => {
    setProgress(null);
    setContext(null);
    setIsActive(false);
    localStorage.removeItem('onboarding_progress');
  }, []);

  return {
    isActive,
    context,
    progress,
    startOnboarding,
    completeOnboarding,
    exitOnboarding,
    resetOnboarding,
  };
}