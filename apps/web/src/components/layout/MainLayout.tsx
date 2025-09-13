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
  
  // Use external state if provided, otherwise use internal state
  const sidebarOpen = externalSidebarCollapsed !== undefined ? !externalSidebarCollapsed : internalSidebarOpen;

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

  return (
    <div 
      className={`min-h-screen bg-gray-50 dark:bg-gray-950 ${className}`}
      data-testid={dataTestId}
    >
      {/* Sidebar */}
      {showSidebar && (
        <Sidebar 
          isOpen={sidebarOpen}
          onClose={handleSidebarClose}
        />
      )}

      {/* Main content */}
      <div className={showSidebar ? 'lg:pl-64' : ''}>
        {/* Header */}
        <Header 
          title={title}
          onMenuToggle={showSidebar ? handleSidebarToggle : undefined}
        />

        {/* Page content */}
        <main className="flex-1">
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