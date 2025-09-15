import React, { useState } from 'react';
import { BaseComponentProps } from '@kubechat/shared/types';
import { Header } from './Header';
import { Sidebar } from './Sidebar';

export interface MainLayoutProps extends BaseComponentProps {
  title?: string;
  showSidebar?: boolean;
}

export const MainLayout: React.FC<MainLayoutProps> = ({
  title = 'KubeChat',
  showSidebar = true,
  className = '',
  children,
  'data-testid': dataTestId = 'main-layout'
}) => {
  const [internalSidebarOpen, setInternalSidebarOpen] = useState(true);

  const sidebarOpen = internalSidebarOpen;

  const handleSidebarClose = () => {
    setInternalSidebarOpen(false);
  };

  const handleSidebarToggle = () => {
    setInternalSidebarOpen(!internalSidebarOpen);
  };

  return (
    <div
      className={`min-h-screen bg-gray-50 dark:bg-gray-950 flex flex-col ${className}`}
      data-testid={dataTestId}
    >
      {/* Header - Full Width */}
      <Header
        title={title}
        onMenuToggle={showSidebar ? handleSidebarToggle : undefined}
      />

      {/* Content Area with Sidebar and Main Content */}
      <div className="flex flex-1 overflow-hidden">
        {/* Sidebar - Below Header */}
        {showSidebar && (
          <Sidebar
            isOpen={sidebarOpen}
            onClose={handleSidebarClose}
          />
        )}

        {/* Main Content - Below Header, Right of Sidebar */}
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