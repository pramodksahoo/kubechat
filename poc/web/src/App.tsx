import React, { useState, useRef, useEffect } from 'react';
import axios from 'axios';
import './App.css';

interface Message {
  id: string;
  type: 'user' | 'assistant';
  content: string;
  command?: string;
  explanation?: string;
  safety?: string;
  result?: any;
  error?: string;
  timestamp: Date;
}

interface QueryResponse {
  command: string;
  explanation: string;
  safety: string;
  preview: boolean;
  result?: any;
  error?: string;
  timestamp: string;
}

function App() {
  const [messages, setMessages] = useState<Message[]>([
    {
      id: '1',
      type: 'assistant',
      content: 'Hello! I\'m KubeChat, your AI-powered Kubernetes assistant. Ask me anything about your cluster using natural language.',
      timestamp: new Date(),
    }
  ]);
  const [input, setInput] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const messagesEndRef = useRef<HTMLDivElement>(null);

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  };

  useEffect(() => {
    scrollToBottom();
  }, [messages]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!input.trim()) return;

    const userMessage: Message = {
      id: Date.now().toString(),
      type: 'user',
      content: input,
      timestamp: new Date(),
    };

    setMessages(prev => [...prev, userMessage]);
    setInput('');
    setIsLoading(true);

    try {
      const response = await axios.post<QueryResponse>('/api/query', {
        query: input
      });

      const assistantMessage: Message = {
        id: (Date.now() + 1).toString(),
        type: 'assistant',
        content: response.data.explanation || 'Command generated successfully',
        command: response.data.command,
        explanation: response.data.explanation,
        safety: response.data.safety,
        result: response.data.result,
        error: response.data.error,
        timestamp: new Date(response.data.timestamp),
      };

      setMessages(prev => [...prev, assistantMessage]);
    } catch (error) {
      const errorMessage: Message = {
        id: (Date.now() + 1).toString(),
        type: 'assistant',
        content: 'Sorry, I encountered an error processing your request.',
        error: axios.isAxiosError(error) ? error.message : 'Unknown error',
        timestamp: new Date(),
      };
      setMessages(prev => [...prev, errorMessage]);
    } finally {
      setIsLoading(false);
    }
  };

  const executeCommand = async (command: string) => {
    setIsLoading(true);
    try {
      const response = await axios.post('/api/query', {
        query: `execute: ${command}`
      });

      const resultMessage: Message = {
        id: Date.now().toString(),
        type: 'assistant',
        content: 'Command executed successfully',
        command: command,
        result: response.data.result,
        error: response.data.error,
        timestamp: new Date(),
      };

      setMessages(prev => [...prev, resultMessage]);
    } catch (error) {
      console.error('Execute error:', error);
    } finally {
      setIsLoading(false);
    }
  };

  const getSafetyColor = (safety?: string) => {
    switch (safety) {
      case 'safe':
        return 'text-green-600 bg-green-100';
      case 'warning':
        return 'text-yellow-600 bg-yellow-100';
      case 'dangerous':
        return 'text-red-600 bg-red-100';
      default:
        return 'text-gray-600 bg-gray-100';
    }
  };

  const formatResult = (result: any) => {
    if (!result) return null;
    
    if (Array.isArray(result)) {
      return (
        <div className="mt-2">
          <div className="text-sm text-gray-600 mb-2">Results ({result.length} items):</div>
          <div className="space-y-1 max-h-64 overflow-y-auto">
            {result.map((item, index) => (
              <div key={index} className="text-xs bg-gray-50 p-2 rounded border">
                {typeof item === 'object' ? (
                  <pre className="whitespace-pre-wrap">{JSON.stringify(item, null, 2)}</pre>
                ) : (
                  <span>{String(item)}</span>
                )}
              </div>
            ))}
          </div>
        </div>
      );
    }

    return (
      <div className="mt-2">
        <div className="text-sm text-gray-600 mb-2">Result:</div>
        <div className="text-xs bg-gray-50 p-3 rounded border">
          <pre className="whitespace-pre-wrap">{JSON.stringify(result, null, 2)}</pre>
        </div>
      </div>
    );
  };

  return (
    <div className="flex flex-col h-screen bg-gray-50">
      {/* Header */}
      <header className="bg-blue-600 text-white p-4">
        <h1 className="text-2xl font-bold">KubeChat</h1>
        <p className="text-blue-100">Natural Language Kubernetes Management</p>
      </header>

      {/* Messages */}
      <div className="flex-1 overflow-y-auto p-4 space-y-4">
        {messages.map(message => (
          <div
            key={message.id}
            className={`flex ${
              message.type === 'user' ? 'justify-end' : 'justify-start'
            }`}
          >
            <div
              className={`max-w-2xl p-4 rounded-lg ${
                message.type === 'user'
                  ? 'bg-blue-500 text-white'
                  : 'bg-white text-gray-800 shadow-md'
              }`}
            >
              <div>{message.content}</div>
              
              {message.command && (
                <div className="mt-3 p-3 bg-gray-900 text-green-400 rounded text-sm font-mono">
                  <div className="flex items-center justify-between mb-2">
                    <span>Generated Command:</span>
                    {message.safety && (
                      <span className={`px-2 py-1 rounded-full text-xs ${getSafetyColor(message.safety)}`}>
                        {message.safety}
                      </span>
                    )}
                  </div>
                  <code>{message.command}</code>
                  {message.safety === 'warning' || message.safety === 'dangerous' ? (
                    <div className="mt-2">
                      <button
                        onClick={() => executeCommand(message.command!)}
                        className="bg-yellow-600 hover:bg-yellow-700 text-white px-3 py-1 rounded text-xs"
                        disabled={isLoading}
                      >
                        Execute Command
                      </button>
                    </div>
                  ) : null}
                </div>
              )}

              {message.result && formatResult(message.result)}

              {message.error && (
                <div className="mt-2 p-2 bg-red-100 text-red-700 rounded text-sm">
                  Error: {message.error}
                </div>
              )}

              <div className="text-xs text-gray-500 mt-2">
                {message.timestamp.toLocaleTimeString()}
              </div>
            </div>
          </div>
        ))}
        
        {isLoading && (
          <div className="flex justify-start">
            <div className="bg-white p-4 rounded-lg shadow-md">
              <div className="flex items-center space-x-2">
                <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-blue-500"></div>
                <span>Processing your query...</span>
              </div>
            </div>
          </div>
        )}
        
        <div ref={messagesEndRef} />
      </div>

      {/* Input */}
      <form onSubmit={handleSubmit} className="p-4 bg-white border-t">
        <div className="flex space-x-4">
          <input
            type="text"
            value={input}
            onChange={(e) => setInput(e.target.value)}
            placeholder="Ask me about your Kubernetes cluster... (e.g., 'show me all pods in default namespace')"
            className="flex-1 p-3 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
            disabled={isLoading}
          />
          <button
            type="submit"
            disabled={isLoading}
            className="px-6 py-3 bg-blue-500 text-white rounded-lg hover:bg-blue-600 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            Send
          </button>
        </div>
        <div className="mt-2 text-xs text-gray-500">
          Try: "show me all pods", "get services in kube-system", "which nodes are running"
        </div>
      </form>
    </div>
  );
}

export default App;