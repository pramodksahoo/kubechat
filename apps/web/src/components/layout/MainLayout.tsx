import React, { useState } from 'react';
import { BaseComponentProps } from '@kubechat/shared/types';
import { Header } from './Header';
import { Sidebar } from './Sidebar';

export interface MainLayoutProps extends BaseComponentProps {
  title?: string;
  showSidebar?: boolean;
  sidebarCollapsed?: boolean;
  onSidebarToggle?: (collapsed: boolean) => void;
}

export const MainLayout: React.FC<MainLayoutProps> = ({
  title = 'KubeChat',
  showSidebar = true,
  sidebarCollapsed: externalSidebarCollapsed,
  onSidebarToggle,
  className = '',
  children,
  'data-testid': dataTestId = 'main-layout'
}) => {
  const [internalSidebarOpen, setInternalSidebarOpen] = useState(true);
  const [internalSidebarCollapsed, setInternalSidebarCollapsed] = useState(false);

  // Use external state if provided, otherwise use internal state
  const sidebarOpen = externalSidebarCollapsed !== undefined ? !externalSidebarCollapsed : internalSidebarOpen;
  const sidebarCollapsed = externalSidebarCollapsed !== undefined ? externalSidebarCollapsed : internalSidebarCollapsed;

  const handleSidebarToggle = () => {
    if (onSidebarToggle && externalSidebarCollapsed !== undefined) {
      onSidebarToggle(!externalSidebarCollapsed);
    } else {
      setInternalSidebarOpen(!internalSidebarOpen);
    }
  };

  const handleSidebarClose = () => {
    if (onSidebarToggle && externalSidebarCollapsed !== undefined) {
      onSidebarToggle(true);
    } else {
      setInternalSidebarOpen(false);
    }
  };

  const handleSidebarCollapseToggle = () => {
    if (onSidebarToggle) {
      // If external control is provided, use it
      onSidebarToggle(!sidebarCollapsed);
    } else {
      // Otherwise use internal state
      setInternalSidebarCollapsed(!internalSidebarCollapsed);
    }
  };

  return (
    <div
      className={`min-h-screen bg-gray-50 dark:bg-gray-950 flex ${className}`}
      data-testid={dataTestId}
    >
      {/* Sidebar - Full Height */}
      {showSidebar && (
        <Sidebar
          isOpen={sidebarOpen}
          isCollapsed={sidebarCollapsed}
          onClose={handleSidebarClose}
          onToggleCollapse={handleSidebarCollapseToggle}
        />
      )}

      {/* Main Content Area with Header */}
      <div className="flex flex-col flex-1 overflow-hidden">
        {/* Header - Only spans main content area */}
        <Header
          title={title}
          onMenuToggle={showSidebar ? handleSidebarToggle : undefined}
        />

        {/* Main Content */}
        <main className="flex-1 overflow-y-auto bg-gray-50 dark:bg-gray-950">
          <div className="py-6">
            <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
              {children}
            </div>
          </div>
        </main>
      </div>
    </div>
  );
};

export default MainLayout;