import React from 'react';
import Head from 'next/head';
import { MainLayout } from '@/components/layout';

export default function ClustersPage() {
  return (
    <>
      <Head>
        <title>Cluster Explorer - KubeChat</title>
        <meta name="description" content="Explore and manage your Kubernetes clusters" />
      </Head>

      <MainLayout>
        <div className="space-y-6">
          <div>
            <h1 className="text-2xl font-bold text-gray-900 dark:text-white">Cluster Explorer</h1>
            <p className="text-gray-600 dark:text-gray-400 mt-1">
              Explore and manage your Kubernetes clusters, nodes, and resources
            </p>
          </div>

          <div className="bg-white dark:bg-gray-900 rounded-lg border border-gray-200 dark:border-gray-700 p-6">
            <div className="text-center py-12">
              <div className="h-16 w-16 mx-auto bg-primary-100 dark:bg-primary-900 rounded-full flex items-center justify-center">
                <svg className="h-8 w-8 text-primary-600 dark:text-primary-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5.25 14.25h13.5m-13.5 0a3 3 0 01-3-3m3 3a3 3 0 100 6h13.5a3 3 0 100-6m-16.5-3a3 3 0 013-3h13.5a3 3 0 013 3m-19.5 0a4.5 4.5 0 01.9-2.7L5.737 5.1a3.375 3.375 0 012.7-1.35h7.126c1.062 0 2.062.5 2.7 1.35l2.587 3.45a4.5 4.5 0 01.9 2.7m0 0a3 3 0 01-3 3m0 3h.008v.008h-.008v-.008zm0-6h.008v.008h-.008v-.008zm-3 6h.008v.008h-.008v-.008zm0-6h.008v.008h-.008v-.008z" />
                </svg>
              </div>
              <h3 className="mt-4 text-lg font-medium text-gray-900 dark:text-white">Cluster Explorer</h3>
              <p className="mt-2 text-gray-500 dark:text-gray-400">
                Detailed cluster management interface coming soon. View cluster details, nodes, pods, and services.
              </p>
            </div>
          </div>
        </div>
      </MainLayout>
    </>
  );
}