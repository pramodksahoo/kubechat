// Chat Services Index for Story 2.2
// Exports all authenticated chat-related services

export { ChatSessionService, chatSessionService } from './chatSessionService';
export { NLPService, nlpService } from './nlpService';
export { CommandService, commandService } from './commandService';
export { AuthenticatedWebSocketService, authenticatedWebSocketService } from './websocketService';
export { CommandHistoryService, commandHistoryService } from './commandHistoryService';

// Re-export types for convenience
// Note: These types are currently defined inline in the service files
// In a real implementation, these would be extracted to separate type files