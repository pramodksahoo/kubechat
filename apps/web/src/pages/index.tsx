import React from 'react';
import Head from 'next/head';

export default function Home() {
  return (
    <>
      <Head>
        <title>KubeChat - Kubernetes Natural Language Interface</title>
        <meta name="description" content="AI-powered Kubernetes management through natural language" />
        <meta name="viewport" content="width=device-width, initial-scale=1" />
        <link rel="icon" href="/favicon.ico" />
      </Head>
      
      <main className="min-h-screen bg-gradient-to-br from-blue-50 to-indigo-100">
        <div className="container mx-auto px-4 py-16">
          <div className="text-center">
            <h1 className="text-5xl font-bold text-gray-900 mb-6">
              Welcome to KubeChat
            </h1>
            <p className="text-xl text-gray-600 mb-8 max-w-2xl mx-auto">
              Manage your Kubernetes clusters through natural language with AI-powered intelligence
            </p>
            <div className="bg-white rounded-lg shadow-lg p-8 max-w-md mx-auto">
              <div className="text-green-600 text-lg font-semibold">
                ✅ Container-First Development Environment
              </div>
              <div className="text-sm text-gray-500 mt-2">
                Running in Kubernetes with container-first architecture
              </div>
            </div>
          </div>
        </div>
      </main>
    </>
  );
}