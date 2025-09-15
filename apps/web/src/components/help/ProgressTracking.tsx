import React, { useState, useEffect, useCallback, ReactNode } from 'react';
import { Button, Heading, Body, StatusBadge, MetricValue, MetricLabel } from '../../design-system';
import { cn } from '../../lib/utils';
import { ScreenReaderUtils } from '../../design-system/accessibility';
import { trackEvent } from '../../lib/monitoring';

// Achievement types
export interface Achievement {
  id: string;
  title: string;
  description: string;
  icon: ReactNode;
  category: 'onboarding' | 'expertise' | 'productivity' | 'collaboration' | 'security';
  rarity: 'common' | 'rare' | 'epic' | 'legendary';
  points: number;
  unlocked: boolean;
  unlockedAt?: Date;
  prerequisites?: string[];
  progress?: {
    current: number;
    total: number;
  };
}

// Learning path
export interface LearningPath {
  id: string;
  title: string;
  description: string;
  icon: ReactNode;
  difficulty: 'beginner' | 'intermediate' | 'advanced';
  estimatedTime: number; // in hours
  modules: LearningModule[];
  prerequisites?: string[];
  completedModules: string[];
  startedAt?: Date;
  completedAt?: Date;
}

export interface LearningModule {
  id: string;
  title: string;
  description: string;
  type: 'tutorial' | 'practice' | 'assessment' | 'documentation';
  estimatedTime: number; // in minutes
  content?: ReactNode;
  completed: boolean;
  completedAt?: Date;
  score?: number;
}

// User progress
export interface UserProgress {
  level: number;
  totalPoints: number;
  pointsToNextLevel: number;
  achievements: Achievement[];
  learningPaths: LearningPath[];
  completedOnboarding: boolean;
  lastActiveDate: Date;
  streakDays: number;
  totalTimeSpent: number; // in minutes
  commandsExecuted: number;
  clustersManaged: number;
  helpArticlesRead: number;
}

// Progress tracking component
export interface ProgressTrackingProps {
  userId: string;
  progress: UserProgress;
  onUpdateProgress?: (progress: UserProgress) => void;
  showDetails?: boolean;
  className?: string;
}

export function ProgressTracking({
  userId,
  progress,
  onUpdateProgress,
  showDetails = true,
  className,
}: ProgressTrackingProps) {
  const [selectedPath, setSelectedPath] = useState<LearningPath | null>(null);
  const [selectedAchievement, setSelectedAchievement] = useState<Achievement | null>(null);

  // Calculate level progress
  const levelProgress = Math.round(
    ((progress.totalPoints % 1000) / 1000) * 100
  );

  // Get recent achievements
  const recentAchievements = progress.achievements
    .filter(a => a.unlocked && a.unlockedAt)
    .sort((a, b) => (b.unlockedAt?.getTime() || 0) - (a.unlockedAt?.getTime() || 0))
    .slice(0, 3);

  // Get active learning paths
  const activePaths = progress.learningPaths.filter(
    path => path.startedAt && !path.completedAt
  );

  // Get completed paths
  const completedPaths = progress.learningPaths.filter(
    path => path.completedAt
  );

  const handleStartPath = useCallback((pathId: string) => {
    const updatedPaths = progress.learningPaths.map(path => 
      path.id === pathId
        ? { ...path, startedAt: new Date() }
        : path
    );

    const updatedProgress = {
      ...progress,
      learningPaths: updatedPaths,
    };

    onUpdateProgress?.(updatedProgress);
    trackEvent('learning_path_started', { pathId, userId });
    ScreenReaderUtils.announce(`Started learning path: ${progress.learningPaths.find(p => p.id === pathId)?.title}`);
  }, [progress, onUpdateProgress, userId]);

  const handleCompleteModule = useCallback((pathId: string, moduleId: string) => {
    const updatedPaths = progress.learningPaths.map(path => {
      if (path.id === pathId) {
        const updatedModules = path.modules.map(module =>
          module.id === moduleId
            ? { ...module, completed: true, completedAt: new Date() }
            : module
        );

        const completedModules = [...path.completedModules, moduleId];
        const isPathCompleted = completedModules.length === path.modules.length;

        return {
          ...path,
          modules: updatedModules,
          completedModules,
          completedAt: isPathCompleted ? new Date() : undefined,
        };
      }
      return path;
    });

    const updatedProgress = {
      ...progress,
      learningPaths: updatedPaths,
      totalPoints: progress.totalPoints + 50, // Award points for module completion
    };

    onUpdateProgress?.(updatedProgress);
    trackEvent('learning_module_completed', { pathId, moduleId, userId });
  }, [progress, onUpdateProgress, userId]);

  return (
    <div className={cn('space-y-6', className)}>
      {/* Overview */}
      <div className="enterprise-card p-6">
        <div className="flex items-center justify-between mb-6">
          <Heading level={3}>Your Progress</Heading>
          <StatusBadge variant="info">Level {progress.level}</StatusBadge>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-4 gap-6 mb-6">
          <div className="text-center">
            <MetricValue>{progress.totalPoints.toLocaleString()}</MetricValue>
            <MetricLabel>Total Points</MetricLabel>
          </div>
          <div className="text-center">
            <MetricValue>{progress.achievements.filter(a => a.unlocked).length}</MetricValue>
            <MetricLabel>Achievements</MetricLabel>
          </div>
          <div className="text-center">
            <MetricValue>{progress.streakDays}</MetricValue>
            <MetricLabel>Day Streak</MetricLabel>
          </div>
          <div className="text-center">
            <MetricValue>{Math.round(progress.totalTimeSpent / 60)}</MetricValue>
            <MetricLabel>Hours Learned</MetricLabel>
          </div>
        </div>

        {/* Level Progress */}
        <div>
          <div className="flex justify-between items-center mb-2">
            <Body size="sm" className="font-medium">
              Level {progress.level} Progress
            </Body>
            <Body size="sm" color="tertiary">
              {progress.pointsToNextLevel} points to next level
            </Body>
          </div>
          <div className="w-full bg-gray-200 dark:bg-gray-700 rounded-full h-3">
            <div
              className="bg-primary-600 h-3 rounded-full transition-all duration-500"
              style={{ width: `${levelProgress}%` }}
              role="progressbar"
              aria-valuenow={levelProgress}
              aria-valuemin={0}
              aria-valuemax={100}
              aria-label={`Level progress: ${levelProgress}%`}
            />
          </div>
        </div>
      </div>

      {showDetails && (
        <>
          {/* Recent Achievements */}
          {recentAchievements.length > 0 && (
            <div className="enterprise-card p-6">
              <Heading level={4} className="mb-4">Recent Achievements</Heading>
              <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                {recentAchievements.map((achievement) => (
                  <AchievementCard
                    key={achievement.id}
                    achievement={achievement}
                    onClick={() => setSelectedAchievement(achievement)}
                  />
                ))}
              </div>
            </div>
          )}

          {/* Active Learning Paths */}
          {activePaths.length > 0 && (
            <div className="enterprise-card p-6">
              <Heading level={4} className="mb-4">Continue Learning</Heading>
              <div className="space-y-4">
                {activePaths.map((path) => (
                  <LearningPathCard
                    key={path.id}
                    path={path}
                    onSelect={() => setSelectedPath(path)}
                    onCompleteModule={handleCompleteModule}
                  />
                ))}
              </div>
            </div>
          )}

          {/* Available Learning Paths */}
          {progress.learningPaths.filter(p => !p.startedAt).length > 0 && (
            <div className="enterprise-card p-6">
              <Heading level={4} className="mb-4">Start Learning</Heading>
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                {progress.learningPaths
                  .filter(p => !p.startedAt)
                  .map((path) => (
                    <LearningPathPreview
                      key={path.id}
                      path={path}
                      onStart={() => handleStartPath(path.id)}
                    />
                  ))}
              </div>
            </div>
          )}

          {/* Completed Paths */}
          {completedPaths.length > 0 && (
            <div className="enterprise-card p-6">
              <Heading level={4} className="mb-4">Completed Learning Paths</Heading>
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                {completedPaths.map((path) => (
                  <div
                    key={path.id}
                    className="p-4 border border-success-200 dark:border-success-800 rounded-lg bg-success-50 dark:bg-success-900/20"
                  >
                    <div className="flex items-start gap-3">
                      <div className="w-8 h-8 text-success-600 dark:text-success-400">
                        {path.icon}
                      </div>
                      <div className="flex-1">
                        <Body className="font-medium mb-1">{path.title}</Body>
                        <Body size="sm" color="secondary">
                          Completed on {path.completedAt?.toLocaleDateString()}
                        </Body>
                      </div>
                      <StatusBadge variant="success">Complete</StatusBadge>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          )}
        </>
      )}

      {/* Achievement Detail Modal */}
      {selectedAchievement && (
        <AchievementModal
          achievement={selectedAchievement}
          onClose={() => setSelectedAchievement(null)}
        />
      )}

      {/* Learning Path Detail Modal */}
      {selectedPath && (
        <LearningPathModal
          path={selectedPath}
          onClose={() => setSelectedPath(null)}
          onCompleteModule={handleCompleteModule}
        />
      )}
    </div>
  );
}

// Achievement card component
function AchievementCard({
  achievement,
  onClick,
}: {
  achievement: Achievement;
  onClick: () => void;
}) {
  const rarityColors = {
    common: 'border-gray-300 bg-gray-50 dark:bg-gray-800',
    rare: 'border-blue-300 bg-blue-50 dark:bg-blue-900/20',
    epic: 'border-purple-300 bg-purple-50 dark:bg-purple-900/20',
    legendary: 'border-yellow-300 bg-yellow-50 dark:bg-yellow-900/20',
  };

  return (
    <button
      onClick={onClick}
      className={cn(
        'p-4 rounded-lg border-2 transition-all duration-200 hover:scale-105 text-left w-full',
        rarityColors[achievement.rarity]
      )}
    >
      <div className="flex items-center gap-3 mb-2">
        <div className="w-8 h-8">{achievement.icon}</div>
        <div className="flex-1">
          <Body className="font-medium">{achievement.title}</Body>
          <Body size="sm" color="tertiary">
            {achievement.points} points
          </Body>
        </div>
      </div>
      {achievement.unlockedAt && (
        <Body size="sm" color="secondary">
          Unlocked {achievement.unlockedAt.toLocaleDateString()}
        </Body>
      )}
    </button>
  );
}

// Learning path card component
function LearningPathCard({
  path,
  onSelect,
  onCompleteModule,
}: {
  path: LearningPath;
  onSelect: () => void;
  onCompleteModule: (pathId: string, moduleId: string) => void;
}) {
  const progress = Math.round(
    (path.completedModules.length / path.modules.length) * 100
  );

  const nextModule = path.modules.find(m => !m.completed);

  return (
    <div className="border border-border-primary rounded-lg p-4">
      <div className="flex items-start justify-between mb-4">
        <div className="flex items-start gap-3">
          <div className="w-8 h-8 text-primary-600 dark:text-primary-400">
            {path.icon}
          </div>
          <div>
            <Heading level={5} className="mb-1">{path.title}</Heading>
            <Body size="sm" color="secondary" className="mb-2">
              {path.description}
            </Body>
            <StatusBadge variant="info">{path.difficulty}</StatusBadge>
          </div>
        </div>
        <Button variant="outline" size="sm" onClick={onSelect}>
          View Details
        </Button>
      </div>

      <div className="mb-4">
        <div className="flex justify-between items-center mb-2">
          <Body size="sm" className="font-medium">Progress</Body>
          <Body size="sm" color="tertiary">
            {path.completedModules.length} of {path.modules.length} modules
          </Body>
        </div>
        <div className="w-full bg-gray-200 dark:bg-gray-700 rounded-full h-2">
          <div
            className="bg-primary-600 h-2 rounded-full transition-all duration-500"
            style={{ width: `${progress}%` }}
          />
        </div>
      </div>

      {nextModule && (
        <div className="flex items-center justify-between">
          <div>
            <Body size="sm" className="font-medium">Next: {nextModule.title}</Body>
            <Body size="sm" color="tertiary">
              ~{nextModule.estimatedTime} minutes
            </Body>
          </div>
          <Button
            size="sm"
            onClick={() => onCompleteModule(path.id, nextModule.id)}
          >
            Continue
          </Button>
        </div>
      )}
    </div>
  );
}

// Learning path preview component
function LearningPathPreview({
  path,
  onStart,
}: {
  path: LearningPath;
  onStart: () => void;
}) {
  const difficultyColors = {
    beginner: 'bg-success-100 text-success-800 dark:bg-success-900/20 dark:text-success-200',
    intermediate: 'bg-warning-100 text-warning-800 dark:bg-warning-900/20 dark:text-warning-200',
    advanced: 'bg-error-100 text-error-800 dark:bg-error-900/20 dark:text-error-200',
  };

  return (
    <div className="border border-border-primary rounded-lg p-4 hover:shadow-hover transition-all duration-200">
      <div className="flex items-start gap-3 mb-4">
        <div className="w-8 h-8 text-text-secondary">{path.icon}</div>
        <div className="flex-1">
          <Heading level={5} className="mb-1">{path.title}</Heading>
          <Body size="sm" color="secondary" className="mb-3">
            {path.description}
          </Body>
          <div className="flex items-center gap-2 mb-3">
            <StatusBadge className={difficultyColors[path.difficulty]}>
              {path.difficulty}
            </StatusBadge>
            <Body size="sm" color="tertiary">
              ~{path.estimatedTime} hours
            </Body>
          </div>
          <Body size="sm" color="tertiary">
            {path.modules.length} modules
          </Body>
        </div>
      </div>
      <Button variant="primary" size="sm" onClick={onStart} fullWidth>
        Start Learning Path
      </Button>
    </div>
  );
}

// Achievement modal
function AchievementModal({
  achievement,
  onClose,
}: {
  achievement: Achievement;
  onClose: () => void;
}) {
  return (
    <div className="fixed inset-0 z-modal flex items-center justify-center p-4 bg-black/50 backdrop-blur-sm">
      <div className="bg-background-primary rounded-modal shadow-modal max-w-md w-full p-6">
        <div className="flex items-center justify-between mb-4">
          <div className="flex items-center gap-3">
            <div className="w-10 h-10">{achievement.icon}</div>
            <div>
              <Heading level={4}>{achievement.title}</Heading>
              <StatusBadge variant="info">{achievement.rarity}</StatusBadge>
            </div>
          </div>
          <Button variant="ghost" size="sm" onClick={onClose}>
            <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
            </svg>
          </Button>
        </div>

        <Body className="mb-4">{achievement.description}</Body>

        <div className="flex items-center justify-between">
          <Body size="sm" color="tertiary">
            {achievement.points} points earned
          </Body>
          {achievement.unlockedAt && (
            <Body size="sm" color="secondary">
              Unlocked {achievement.unlockedAt.toLocaleDateString()}
            </Body>
          )}
        </div>
      </div>
    </div>
  );
}

// Learning path modal
function LearningPathModal({
  path,
  onClose,
  onCompleteModule,
}: {
  path: LearningPath;
  onClose: () => void;
  onCompleteModule: (pathId: string, moduleId: string) => void;
}) {
  return (
    <div className="fixed inset-0 z-modal flex items-center justify-center p-4 bg-black/50 backdrop-blur-sm">
      <div className="bg-background-primary rounded-modal shadow-modal max-w-2xl w-full max-h-[80vh] overflow-hidden">
        <div className="p-6 border-b border-border-primary">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <div className="w-8 h-8">{path.icon}</div>
              <Heading level={3}>{path.title}</Heading>
            </div>
            <Button variant="ghost" size="sm" onClick={onClose}>
              <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
              </svg>
            </Button>
          </div>
        </div>

        <div className="p-6 overflow-y-auto max-h-[60vh]">
          <Body className="mb-6">{path.description}</Body>

          <Heading level={4} className="mb-4">Modules</Heading>
          <div className="space-y-3">
            {path.modules.map((module, index) => (
              <div
                key={module.id}
                className={cn(
                  'p-4 rounded-lg border',
                  module.completed
                    ? 'border-success-200 bg-success-50 dark:border-success-800 dark:bg-success-900/20'
                    : 'border-border-primary'
                )}
              >
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-3">
                    <div className="w-8 h-8 flex items-center justify-center rounded-full bg-background-secondary text-sm font-medium">
                      {index + 1}
                    </div>
                    <div>
                      <Body className="font-medium">{module.title}</Body>
                      <Body size="sm" color="secondary">
                        {module.description}
                      </Body>
                      <Body size="sm" color="tertiary">
                        ~{module.estimatedTime} minutes • {module.type}
                      </Body>
                    </div>
                  </div>
                  {module.completed ? (
                    <StatusBadge variant="success">Complete</StatusBadge>
                  ) : (
                    <Button
                      size="sm"
                      onClick={() => onCompleteModule(path.id, module.id)}
                    >
                      Start
                    </Button>
                  )}
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>
    </div>
  );
}

// Hook for progress tracking
export function useProgressTracking(userId: string) {
  const [progress, setProgress] = useState<UserProgress | null>(null);
  const [loading, setLoading] = useState(true);

  // Load progress from storage
  useEffect(() => {
    const loadProgress = async () => {
      try {
        const stored = localStorage.getItem(`user_progress_${userId}`);
        if (stored) {
          const data = JSON.parse(stored);
          // Convert date strings back to Date objects
          const progress: UserProgress = {
            ...data,
            lastActiveDate: new Date(data.lastActiveDate),
            achievements: data.achievements.map((a: Achievement & { unlockedAt?: string }) => ({
              ...a,
              unlockedAt: a.unlockedAt ? new Date(a.unlockedAt) : undefined,
            })),
            learningPaths: data.learningPaths.map((p: LearningPath & { startedAt?: string; completedAt?: string; modules: (LearningModule & { completedAt?: string })[] }) => ({
              ...p,
              startedAt: p.startedAt ? new Date(p.startedAt) : undefined,
              completedAt: p.completedAt ? new Date(p.completedAt) : undefined,
              modules: p.modules.map((m: LearningModule & { completedAt?: string }) => ({
                ...m,
                completedAt: m.completedAt ? new Date(m.completedAt) : undefined,
              })),
            })),
          };
          setProgress(progress);
        }
      } catch (error) {
        console.error('Failed to load progress:', error);
      } finally {
        setLoading(false);
      }
    };

    loadProgress();
  }, [userId]);

  // Save progress to storage
  const updateProgress = useCallback((newProgress: UserProgress) => {
    setProgress(newProgress);
    try {
      localStorage.setItem(`user_progress_${userId}`, JSON.stringify(newProgress));
    } catch (error) {
      console.error('Failed to save progress:', error);
    }
  }, [userId]);

  // Award achievement
  const awardAchievement = useCallback((achievementId: string) => {
    if (!progress) return;

    const achievement = progress.achievements.find(a => a.id === achievementId);
    if (!achievement || achievement.unlocked) return;

    const updatedAchievements = progress.achievements.map(a =>
      a.id === achievementId
        ? { ...a, unlocked: true, unlockedAt: new Date() }
        : a
    );

    const updatedProgress = {
      ...progress,
      achievements: updatedAchievements,
      totalPoints: progress.totalPoints + achievement.points,
    };

    updateProgress(updatedProgress);
    
    ScreenReaderUtils.announce(`Achievement unlocked: ${achievement.title}`);
    trackEvent('achievement_unlocked', { achievementId, userId });
  }, [progress, updateProgress, userId]);

  return {
    progress,
    loading,
    updateProgress,
    awardAchievement,
  };
}