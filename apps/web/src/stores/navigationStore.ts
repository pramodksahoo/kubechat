import { create } from 'zustand';

interface NavigationItem {
  id: string;
  label: string;
  href?: string;
  iconName?: string; // Use string identifier instead of JSX
  onClick?: () => void;
  badge?: string | number;
  children?: NavigationItem[];
  disabled?: boolean;
  external?: boolean;
}

interface NavigationState {
  // Sidebar state
  sidebarOpen: boolean;
  sidebarCollapsed: boolean;
  
  // Navigation items
  navigationItems: NavigationItem[];
  activeItemId: string | null;
  
  // Breadcrumbs
  breadcrumbs: { label: string; href?: string }[];
  
  // Mobile state
  isMobile: boolean;
  
  // Actions
  setSidebarOpen: (open: boolean) => void;
  setSidebarCollapsed: (collapsed: boolean) => void;
  toggleSidebar: () => void;
  setActiveItem: (itemId: string) => void;
  setNavigationItems: (items: NavigationItem[]) => void;
  setBreadcrumbs: (breadcrumbs: { label: string; href?: string }[]) => void;
  setIsMobile: (isMobile: boolean) => void;
}

export const useNavigationStore = create<NavigationState>((set, get) => ({
  // Initial state
  sidebarOpen: true,
  sidebarCollapsed: false,
  navigationItems: [
    {
      id: 'dashboard',
      label: 'Dashboard',
      href: '/',
      iconName: 'home'
    },
    {
      id: 'chat',
      label: 'Chat Interface',
      href: '/chat',
      iconName: 'chat'
    },
    {
      id: 'clusters',
      label: 'Cluster Explorer',
      href: '/clusters',
      iconName: 'server'
    },
    {
      id: 'audit',
      label: 'Audit Trail',
      href: '/audit',
      iconName: 'clipboard'
    },
    {
      id: 'compliance',
      label: 'Compliance',
      href: '/compliance',
      iconName: 'shield'
    },
    {
      id: 'security',
      label: 'Security',
      href: '/security',
      iconName: 'lock'
    }
  ],
  activeItemId: 'dashboard',
  breadcrumbs: [{ label: 'Dashboard' }],
  isMobile: false,

  // Actions
  setSidebarOpen: (open: boolean) => set({ sidebarOpen: open }),
  setSidebarCollapsed: (collapsed: boolean) => set({ sidebarCollapsed: collapsed }),
  toggleSidebar: () => set((state) => ({ sidebarOpen: !state.sidebarOpen })),
  setActiveItem: (itemId: string) => {
    const { navigationItems } = get();
    const activeItem = navigationItems.find(item => item.id === itemId);
    if (activeItem) {
      set({ 
        activeItemId: itemId,
        breadcrumbs: [{ label: activeItem.label, href: activeItem.href }]
      });
    }
  },
  setNavigationItems: (items: NavigationItem[]) => set({ navigationItems: items }),
  setBreadcrumbs: (breadcrumbs: { label: string; href?: string }[]) => set({ breadcrumbs }),
  setIsMobile: (isMobile: boolean) => set({ isMobile })
}));