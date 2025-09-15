import { useEffect, useRef, useState, useCallback } from 'react';

export interface WebSocketMessage {
  type: string;
  data: unknown;
  timestamp: string;
}

export interface WebSocketOptions {
  onConnect?: () => void;
  onDisconnect?: () => void;
  onMessage?: (message: WebSocketMessage) => void;
  onError?: (error: Event) => void;
  reconnectAttempts?: number;
  reconnectDelay?: number;
  protocols?: string[];
}

export function useWebSocket(url: string | null, options: WebSocketOptions = {}) {
  const {
    onConnect,
    onDisconnect,
    onMessage,
    onError,
    reconnectAttempts = 5,
    reconnectDelay = 1000,
    protocols
  } = options;

  const [isConnected, setIsConnected] = useState(false);
  const [connectionError, setConnectionError] = useState<string | null>(null);
  const [lastMessage, setLastMessage] = useState<WebSocketMessage | null>(null);
  
  const wsRef = useRef<WebSocket | null>(null);
  const reconnectTimeoutRef = useRef<NodeJS.Timeout>();
  const attemptsRef = useRef(0);

  const connect = useCallback(() => {
    if (!url || wsRef.current?.readyState === WebSocket.CONNECTING) {
      return;
    }

    try {
      const ws = protocols ? new WebSocket(url, protocols) : new WebSocket(url);
      wsRef.current = ws;

      ws.onopen = () => {
        setIsConnected(true);
        setConnectionError(null);
        attemptsRef.current = 0;
        onConnect?.();
      };

      ws.onmessage = (event) => {
        try {
          const message: WebSocketMessage = JSON.parse(event.data);
          setLastMessage(message);
          onMessage?.(message);
        } catch (error) {
          console.error('Failed to parse WebSocket message:', error);
        }
      };

      ws.onclose = (event) => {
        setIsConnected(false);
        wsRef.current = null;
        onDisconnect?.();

        // Attempt reconnection if not a clean close and we haven't exceeded attempts
        if (!event.wasClean && attemptsRef.current < reconnectAttempts) {
          const delay = reconnectDelay * Math.pow(2, attemptsRef.current); // Exponential backoff
          
          reconnectTimeoutRef.current = setTimeout(() => {
            attemptsRef.current++;
            connect();
          }, delay);
        } else if (attemptsRef.current >= reconnectAttempts) {
          setConnectionError('Failed to reconnect after maximum attempts');
        }
      };

      ws.onerror = (error) => {
        setConnectionError('WebSocket connection error');
        onError?.(error);
      };

    } catch (error) {
      setConnectionError(`Failed to create WebSocket connection: ${error instanceof Error ? error.message : 'Unknown error'}`);
    }
  }, [url, protocols, onConnect, onDisconnect, onMessage, onError, reconnectAttempts, reconnectDelay]);

  const disconnect = useCallback(() => {
    if (reconnectTimeoutRef.current) {
      clearTimeout(reconnectTimeoutRef.current);
    }
    
    if (wsRef.current) {
      wsRef.current.close(1000, 'Client disconnecting');
      wsRef.current = null;
    }
    
    setIsConnected(false);
    setConnectionError(null);
    attemptsRef.current = 0;
  }, []);

  const sendMessage = useCallback((message: unknown) => {
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      try {
        const messageStr = typeof message === 'string' ? message : JSON.stringify(message);
        wsRef.current.send(messageStr);
        return true;
      } catch (error) {
        console.error('Failed to send WebSocket message:', error);
        return false;
      }
    }
    return false;
  }, []);

  // Connect when URL is provided
  useEffect(() => {
    if (url) {
      connect();
    }

    return () => {
      disconnect();
    };
  }, [url, connect, disconnect]);

  // Cleanup on unmount
  useEffect(() => {
    return () => {
      disconnect();
    };
  }, [disconnect]);

  return {
    isConnected,
    connectionError,
    lastMessage,
    sendMessage,
    connect,
    disconnect,
  };
}

// Specialized hook for chat WebSocket
export function useChatWebSocket(sessionId: string | null, token: string | null) {
  const [messages, setMessages] = useState<WebSocketMessage[]>([]);
  const [commandUpdates, setCommandUpdates] = useState<WebSocketMessage[]>([]);

  const wsUrl = sessionId && token 
    ? `${process.env.NEXT_PUBLIC_WS_URL || 'ws://localhost:8080'}/api/v1/chat/sessions/${sessionId}/ws?token=${token}`
    : null;

  const handleMessage = useCallback((message: WebSocketMessage) => {
    switch (message.type) {
      case 'chat_message':
        setMessages(prev => [...prev, message]);
        break;
      case 'command_status':
      case 'command_approval':
      case 'command_execution':
        setCommandUpdates(prev => [...prev, message]);
        break;
      default:
        console.log('Unknown WebSocket message type:', message.type);
    }
  }, []);

  const {
    isConnected,
    connectionError,
    lastMessage,
    sendMessage,
    connect,
    disconnect,
  } = useWebSocket(wsUrl, {
    onMessage: handleMessage,
    onConnect: () => console.log('Chat WebSocket connected'),
    onDisconnect: () => console.log('Chat WebSocket disconnected'),
    onError: (error) => console.error('Chat WebSocket error:', error),
    reconnectAttempts: 3,
    reconnectDelay: 2000,
  });

  const sendChatMessage = useCallback((content: string) => {
    return sendMessage({
      type: 'chat_message',
      content,
      timestamp: new Date().toISOString(),
    });
  }, [sendMessage]);

  const sendCommandRequest = useCallback((command: string, natural_language: string) => {
    return sendMessage({
      type: 'command_request',
      command,
      natural_language,
      timestamp: new Date().toISOString(),
    });
  }, [sendMessage]);

  return {
    isConnected,
    connectionError,
    lastMessage,
    messages,
    commandUpdates,
    sendChatMessage,
    sendCommandRequest,
    connect,
    disconnect,
  };
}