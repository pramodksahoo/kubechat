# KubeChat AI Frontend Development Prompts

## Overview

This document contains comprehensive, production-ready prompts for AI-driven frontend development tools (Vercel v0, Lovable.ai, etc.) to generate KubeChat's enterprise-grade web interface components. Each prompt follows the structured four-part framework for optimal results.

## Tech Stack Context

**Primary Stack:** Next.js 15, TypeScript, Tailwind CSS, React Server Components
**Component Library:** Custom enterprise design system with Lucide React icons
**State Management:** Zustand for client state, React Query for server state
**Real-time:** WebSocket connection for collaborative features and live updates
**Authentication:** NextAuth.js with enterprise SSO integration

---

## 1. Dashboard Component Prompt

### High-Level Goal
Create a responsive enterprise dashboard for KubeChat that serves as the central command hub, showcasing natural language Kubernetes management capabilities with professional aesthetics and immediate value demonstration.

### Detailed Step-by-Step Instructions

1. **Create Main Dashboard Component (`components/Dashboard.tsx`)**
   - Use Next.js 15 App Router server component architecture
   - Implement responsive grid layout using Tailwind CSS Grid
   - Add TypeScript interfaces for all props and data structures

2. **Build Header Section**
   - Create KubeChat logo with enterprise branding (use placeholder for now)
   - Add horizontal navigation: Dashboard, Chat, Monitor, Audit, Profile
   - Include user account dropdown and cluster connection status indicator
   - Implement responsive hamburger menu for mobile breakpoints

3. **Implement Central Command Bar**
   - Design prominent natural language input with auto-suggest functionality  
   - Add placeholder text: "Ask KubeChat: 'Show me pods with high memory usage in production'"
   - Include microphone icon (future voice input) and submit button
   - Implement loading states with AI processing indicators

4. **Create Stats Cards Grid**
   - Build 4 stat cards: Cluster Health, Active Conversations, Recent Alerts, Compliance Status
   - Each card should be clickable with hover effects and navigation
   - Use color-coded indicators (green=healthy, yellow=warning, red=error)
   - Include trend indicators and summary numbers

5. **Add Recent Activity Panel**
   - Create scrollable list of last 5 kubectl operations
   - Show user avatars, timestamps, operation descriptions
   - Add "Re-execute" quick action buttons
   - Implement skeleton loading states

6. **Build Team Activity Stream**
   - Display real-time collaborative activity feed
   - Show team member actions with proper attribution
   - Include "Join Session" buttons for active collaborations
   - Add presence indicators (online/offline status)

7. **Implement Responsive Behavior**
   - Mobile: Single column, collapsible sections, bottom navigation
   - Tablet: Two-column grid, touch-optimized interactions
   - Desktop: Full multi-column layout with persistent sidebars

### Code Examples, Data Structures & Constraints

**Required Data Interfaces:**
```typescript
interface DashboardData {
  clusterHealth: {
    status: 'healthy' | 'warning' | 'critical';
    nodeCount: number;
    cpuUsage: number;
    memoryUsage: number;
  };
  recentActivity: {
    id: string;
    user: string;
    action: string;
    timestamp: Date;
    status: 'success' | 'error' | 'pending';
  }[];
  teamActivity: {
    userId: string;
    userName: string;
    activity: string;
    sessionId?: string;
    timestamp: Date;
  }[];
}
```

**Color Scheme (from specification):**
- Primary: #2563EB (trustworthy blue)
- Secondary: #1E293B (enterprise dark gray)  
- Success: #10B981, Warning: #F59E0B, Error: #EF4444
- Neutral grays: #64748B, #94A3B8, #CBD5E1, #F1F5F9

**Styling Constraints:**
- Use Inter font family for all text
- JetBrains Mono for any code/command text
- 8px base spacing unit (space-2, space-4, space-6, etc.)
- Minimum 44px touch targets for interactive elements
- WCAG 2.1 AA contrast ratios (4.5:1 for text)

**DO NOT:**
- Use any external UI libraries (no shadcn/ui, Ant Design, etc.)
- Implement actual API calls (use mock data)
- Add complex animations (simple transitions only)
- Include actual authentication logic

### Define Strict Scope

**Files to Create:**
- `components/Dashboard.tsx` (main component)
- `components/ui/StatCard.tsx` (reusable stat card)
- `components/ui/ActivityItem.tsx` (activity list item)
- `types/dashboard.ts` (TypeScript interfaces)

**Files to Reference but NOT Modify:**
- Any existing layout or authentication components
- Global CSS or configuration files
- API route handlers

**Integration Points:**
- Component should accept props for all dynamic data
- Export default Dashboard component for easy importing
- Include proper TypeScript exports for interfaces

---

## 2. Chat Interface Component Prompt  

### High-Level Goal
Build KubeChat's signature natural language chat interface that demonstrates the core competitive advantage - converting conversational queries into safe kubectl commands with professional enterprise UI patterns and real-time collaborative features.

### Detailed Step-by-Step Instructions

1. **Create Chat Layout Component (`components/ChatInterface.tsx`)**
   - Implement three-panel layout: sidebar, main chat, context panel
   - Use CSS Grid for responsive layout management
   - Add collapsible sidebar for mobile/small screens

2. **Build Conversation Sidebar**
   - Create conversation history list with search functionality
   - Add "New Conversation" button with prominent styling
   - Include shared session indicators and "Join" buttons
   - Implement conversation templates/favorites section
   - Add export conversation option

3. **Design Main Chat Area**
   - Create ChatGPT-style message interface with user/assistant bubbles
   - Implement auto-scrolling to latest messages
   - Add message timestamp and user attribution
   - Include copy message and share message actions

4. **Build Natural Language Input**
   - Create expandable textarea with placeholder suggestions
   - Add send button with loading states
   - Implement auto-suggest dropdown with common queries
   - Include voice input button (placeholder for future)

5. **Create Command Preview Cards**
   - Design safety-classified cards (safe=green, warning=yellow, dangerous=red)
   - Include kubectl command with syntax highlighting
   - Add plain-English explanation of command impact
   - Implement approve/deny buttons with confirmation modals

6. **Build Result Display Components**
   - Create structured result viewer (not raw terminal output)
   - Implement expandable JSON tree viewer
   - Add table view for list results (pods, services, etc.)
   - Include export results functionality

7. **Add Collaboration Features**
   - Show active collaborators with presence indicators
   - Add "Share Session" button with invite link generation
   - Implement real-time cursor sharing (placeholder UI)
   - Include session recording indicator

8. **Implement Context Panel**
   - Display current cluster/namespace context
   - Add quick cluster switching dropdown  
   - Show relevant resources and quick actions
   - Include help documentation links

### Code Examples, Data Structures & Constraints

**Required Message Interface:**
```typescript
interface ChatMessage {
  id: string;
  type: 'user' | 'assistant' | 'system';
  content: string;
  timestamp: Date;
  userId?: string;
  commandPreview?: {
    command: string;
    explanation: string;
    safety: 'safe' | 'warning' | 'dangerous';
    approved?: boolean;
  };
  result?: {
    status: 'success' | 'error';
    data: any;
    formattedOutput: string;
  };
}

interface ChatSession {
  id: string;
  name: string;
  participants: string[];
  messages: ChatMessage[];
  isShared: boolean;
  cluster: string;
  namespace: string;
}
```

**Component Architecture:**
```typescript
// Main chat interface with proper TypeScript
interface ChatInterfaceProps {
  sessions: ChatSession[];
  activeSessionId: string;
  currentUser: string;
  onSendMessage: (message: string) => void;
  onApproveCommand: (messageId: string) => void;
}
```

**Styling Requirements:**
- Message bubbles with proper spacing and typography
- Syntax highlighting for kubectl commands (use Prism.js classes)
- Color-coded safety classifications matching design system
- Responsive breakpoints: mobile stacked, tablet split, desktop three-panel

**DO NOT:**
- Implement actual WebSocket connections (mock real-time updates)
- Add actual command execution logic
- Include complex syntax highlighting libraries
- Create actual file upload/download functionality

### Define Strict Scope

**Files to Create:**
- `components/ChatInterface.tsx` (main layout)
- `components/chat/MessageBubble.tsx` (individual messages)
- `components/chat/CommandPreview.tsx` (command preview cards)
- `components/chat/ResultViewer.tsx` (structured results)
- `components/chat/ConversationSidebar.tsx` (sidebar component)
- `types/chat.ts` (TypeScript interfaces)

**Mock Data Requirements:**
- Sample conversations with various message types
- Example kubectl commands with safety classifications
- Mock collaboration data (active users, shared sessions)

---

## 3. Cluster Monitor Dashboard Prompt

### High-Level Goal
Create an enterprise-grade real-time Kubernetes cluster monitoring interface that showcases web-based collaborative advantages over CLI tools, featuring interactive resource exploration, live metrics, and natural language query integration.

### Detailed Step-by-Step Instructions

1. **Build Monitor Layout Component (`components/ClusterMonitor.tsx`)**
   - Create responsive multi-panel dashboard layout
   - Implement resizable panels using CSS Grid and drag handles
   - Add panel collapse/expand functionality
   - Include full-screen mode for individual panels

2. **Create Cluster Health Overview**
   - Build real-time status cards for CPU, Memory, Storage, Network
   - Add cluster-wide metrics with trend indicators
   - Include alert summary with severity indicators
   - Implement quick health score calculation and display

3. **Build Interactive Resource Tree**
   - Create hierarchical tree view: Cluster → Namespaces → Resources
   - Add expandable/collapsible nodes with resource counts
   - Include health status indicators at each level
   - Implement click-to-query functionality (generates natural language)

4. **Design Live Metrics Dashboard**
   - Create real-time charts for key metrics (CPU, Memory, Network)
   - Use simple SVG charts or Canvas for performance
   - Add time range selectors (1h, 6h, 24h, 7d)
   - Include zoom and pan functionality for detailed analysis

5. **Implement Log Viewer Panel**
   - Create scrollable log stream with auto-scroll toggle
   - Add log level filtering (Error, Warning, Info, Debug)
   - Include search functionality with highlighting
   - Implement "Query logs with AI" integration point

6. **Build Resource Detail Modals**
   - Create expandable detail views for pods, services, deployments
   - Include YAML/JSON viewers with syntax highlighting
   - Add resource actions (describe, edit, delete with confirmations)
   - Implement resource relationship visualization

7. **Add Collaborative Features**
   - Show shared cursor positions when in collaborative mode
   - Add "Share current view" functionality
   - Include annotation system for marking issues
   - Implement session recording for troubleshooting

8. **Create Alert Management**
   - Build alert list with filtering and prioritization
   - Add alert acknowledgment and assignment
   - Include alert history and pattern analysis
   - Implement escalation workflows

### Code Examples, Data Structures & Constraints

**Core Data Structures:**
```typescript
interface ClusterMetrics {
  cpu: {
    used: number;
    total: number;
    percentage: number;
    trend: number[];
  };
  memory: {
    used: number;
    total: number;
    percentage: number;
    trend: number[];
  };
  nodes: {
    total: number;
    ready: number;
    notReady: number;
  };
  pods: {
    total: number;
    running: number;
    pending: number;
    failed: number;
  };
}

interface ResourceTreeNode {
  id: string;
  name: string;
  type: 'cluster' | 'namespace' | 'deployment' | 'service' | 'pod';
  status: 'healthy' | 'warning' | 'error';
  children?: ResourceTreeNode[];
  metrics?: {
    cpu: number;
    memory: number;
  };
}
```

**Chart Configuration:**
- Use simple line/area charts for metrics
- Color scheme: Primary blue for normal, warning amber, error red
- Responsive sizing with minimum dimensions
- Update frequency: 30-second intervals (simulated)

**Performance Constraints:**
- Virtual scrolling for large resource lists
- Debounced search inputs
- Lazy loading for resource details
- Maximum 1000 visible log entries

**DO NOT:**
- Implement actual Kubernetes API calls
- Add complex charting libraries (keep it simple)
- Include real WebSocket connections
- Create actual resource modification capabilities

### Define Strict Scope

**Files to Create:**
- `components/ClusterMonitor.tsx` (main dashboard)
- `components/monitoring/MetricsPanel.tsx` (metrics display)
- `components/monitoring/ResourceTree.tsx` (hierarchical tree)
- `components/monitoring/LogViewer.tsx` (log stream)
- `components/monitoring/AlertPanel.tsx` (alert management)
- `types/monitoring.ts` (interfaces)

**Mock Data Requirements:**
- Simulated real-time metrics with trend data
- Sample Kubernetes resource hierarchy
- Example log entries with various levels
- Mock alert data with different priorities

---

## 4. Audit & Compliance Interface Prompt

### High-Level Goal
Build a professional enterprise audit interface that demonstrates KubeChat's compliance capabilities - comprehensive audit trail management, regulatory reporting, and investigation tools that CLI-based competitors cannot provide.

### Detailed Step-by-Step Instructions

1. **Create Audit Layout Component (`components/AuditInterface.tsx`)**
   - Build professional three-panel layout: filters, results, details
   - Use enterprise-standard table layouts with advanced functionality
   - Add responsive design for mobile audit review

2. **Design Advanced Search Panel**
   - Create comprehensive filtering interface (date range, user, operation, cluster)
   - Add saved search profiles with naming and sharing
   - Include quick filter buttons for common searches
   - Implement advanced boolean search capabilities

3. **Build Audit Trail Table**
   - Create sortable, filterable data table with virtualization
   - Include columns: Timestamp, User, Operation, Resource, Result, Risk Level
   - Add row selection for bulk operations
   - Implement expandable row details with full context

4. **Create Detail Panel**
   - Build comprehensive audit entry viewer
   - Include full command context, before/after states
   - Add user session timeline view
   - Implement related operations grouping

5. **Design Compliance Dashboard**
   - Create compliance status overview (SOX, HIPAA, SOC 2)
   - Add risk scoring and trend analysis
   - Include compliance report generation interface
   - Implement alert threshold configuration

6. **Build Export & Reporting**
   - Create professional export interface with format selection
   - Add digital signature capabilities for audit integrity
   - Include batch export with progress tracking
   - Implement secure delivery options (email, secure download)

7. **Add Investigation Tools**
   - Create timeline visualization for incident investigation
   - Add user behavior pattern analysis
   - Include correlation analysis for related events
   - Implement investigation notes and collaboration

8. **Create Compliance Templates**
   - Build template-based report generator
   - Add regulatory framework selection (SOX, HIPAA, etc.)
   - Include custom report builder interface
   - Implement scheduled reporting functionality

### Code Examples, Data Structures & Constraints

**Audit Data Structures:**
```typescript
interface AuditEntry {
  id: string;
  timestamp: Date;
  userId: string;
  userName: string;
  operation: string;
  resource: {
    type: string;
    namespace: string;
    name: string;
  };
  command: string;
  result: 'success' | 'error' | 'denied';
  riskLevel: 'low' | 'medium' | 'high' | 'critical';
  context: {
    sessionId: string;
    ipAddress: string;
    userAgent: string;
    clusterId: string;
  };
  impact?: {
    resourcesAffected: number;
    dataModified: boolean;
    configurationChanged: boolean;
  };
}

interface ComplianceReport {
  id: string;
  framework: 'SOX' | 'HIPAA' | 'SOC2' | 'PCI-DSS';
  dateRange: {
    start: Date;
    end: Date;
  };
  status: 'compliant' | 'warning' | 'non-compliant';
  findings: ComplianceFinding[];
  recommendations: string[];
}
```

**Table Requirements:**
- Virtual scrolling for performance with large datasets
- Server-side sorting and filtering (simulated)
- Column visibility controls and resizing
- Export selected rows functionality

**Professional Styling:**
- Enterprise data table with alternating row colors
- Professional typography with clear hierarchy
- Status indicators with consistent color coding
- Print-friendly styles for compliance reports

**Security Considerations:**
- Mock data should not include actual sensitive information
- Implement proper access control UI indicators
- Add data retention policy displays
- Include audit trail integrity verification UI

**DO NOT:**
- Include actual sensitive audit data
- Implement real export functionality (mock the UI)
- Add actual digital signature capabilities
- Create real email sending functionality

### Define Strict Scope

**Files to Create:**
- `components/AuditInterface.tsx` (main layout)
- `components/audit/AuditTable.tsx` (main data table)
- `components/audit/SearchPanel.tsx` (filtering interface)
- `components/audit/ComplianceDashboard.tsx` (compliance overview)
- `components/audit/ExportPanel.tsx` (export interface)
- `components/audit/InvestigationTools.tsx` (analysis tools)
- `types/audit.ts` (TypeScript interfaces)

**Mock Data Requirements:**
- Comprehensive audit entries spanning multiple timeframes
- Sample compliance reports for different frameworks
- Example investigation scenarios
- Mock export formats and templates

---

## Usage Instructions

1. **Select appropriate prompt** based on component needed
2. **Copy entire prompt** including context, instructions, and constraints  
3. **Paste into AI tool** (v0, Lovable, etc.) without modification
4. **Iterate with refinements** using follow-up prompts for specific adjustments
5. **Test responsive behavior** across breakpoints before integration

## Important Notes

- All prompts designed for **production-ready code**
- **TypeScript and accessibility** requirements built into each prompt
- **Mock data patterns** provided for realistic development
- **Enterprise design standards** enforced throughout
- **Component isolation** ensures safe integration into larger application

---

*🤖 Generated with [Claude Code](https://claude.ai/code)*

*Co-Authored-By: Claude <hellow@kubechat.dev>*