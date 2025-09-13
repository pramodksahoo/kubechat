import React, { useEffect } from 'react';
import Head from 'next/head';
import { MainLayout } from '@/components/layout';
import { ChatInterface } from '@/components/chat';
import { useChatStore } from '@/stores/chatStore';

export default function ChatPage() {
  const {
    currentSession,
    messages,
    loading,
    createSession,
    sendMessage,
  } = useChatStore();

  // Create a default session if none exists
  useEffect(() => {
    if (!currentSession) {
      createSession().catch(console.error);
    }
  }, [currentSession, createSession]);

  const handleSendMessage = async (message: string) => {
    try {
      await sendMessage(message);
    } catch (error) {
      console.error('Failed to send message:', error);
    }
  };

  // Show loading state if no session yet
  if (!currentSession) {
    return (
      <>
        <Head>
          <title>Chat Interface - KubeChat</title>
          <meta name="description" content="Natural language interface for Kubernetes management" />
        </Head>

        <MainLayout>
          <div className="space-y-6">
            <div>
              <h1 className="text-2xl font-bold text-gray-900 dark:text-white">Chat Interface</h1>
              <p className="text-gray-600 dark:text-gray-400 mt-1">
                Interact with your Kubernetes clusters using natural language
              </p>
            </div>

            <div className="bg-white dark:bg-gray-900 rounded-lg border border-gray-200 dark:border-gray-700 p-8">
              <div className="flex items-center justify-center">
                <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary-600"></div>
                <span className="ml-3 text-gray-600 dark:text-gray-400">Initializing chat session...</span>
              </div>
            </div>
          </div>
        </MainLayout>
      </>
    );
  }

  return (
    <>
      <Head>
        <title>Chat Interface - KubeChat</title>
        <meta name="description" content="Natural language interface for Kubernetes management" />
      </Head>

      <MainLayout>
        <div className="space-y-6">
          <div>
            <h1 className="text-2xl font-bold text-gray-900 dark:text-white">Chat Interface</h1>
            <p className="text-gray-600 dark:text-gray-400 mt-1">
              Interact with your Kubernetes clusters using natural language
            </p>
          </div>

          <div className="bg-white dark:bg-gray-900 rounded-lg border border-gray-200 dark:border-gray-700 overflow-hidden">
            <ChatInterface
              session={currentSession as any}
              messages={messages as any}
              onSendMessage={handleSendMessage}
              loading={loading}
            />
          </div>
        </div>
      </MainLayout>
    </>
  );
}