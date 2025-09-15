import React, { useState, useEffect } from 'react';
import Head from 'next/head';
import { MainLayout } from '@/components/layout';
import { ClusterHealthWidget } from '@/components/dashboard';
import { clusterService, ClusterInfo } from '@/services/clusterService';

export default function ClustersPage() {
  const [clusters, setClusters] = useState<ClusterInfo[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    loadClusters();
  }, []);

  const loadClusters = async () => {
    setLoading(true);
    setError(null);
    try {
      const clusterData = await clusterService.getClusters();
      setClusters(clusterData);
    } catch (error) {
      console.error('Failed to load clusters:', error);
      setError('Failed to load cluster data');
    } finally {
      setLoading(false);
    }
  };

  const handleRefresh = async () => {
    await loadClusters();
  };

  const handleClusterClick = (clusterId: string) => {
    console.log('Viewing cluster details for:', clusterId);
    // TODO: Navigate to cluster detail page
    // router.push(`/clusters/${clusterId}`);
  };

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

          {error && (
            <div className="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg p-4">
              <div className="flex">
                <div className="flex-shrink-0">
                  <svg className="h-5 w-5 text-red-400" viewBox="0 0 20 20" fill="currentColor">
                    <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clipRule="evenodd" />
                  </svg>
                </div>
                <div className="ml-3">
                  <h3 className="text-sm font-medium text-red-800 dark:text-red-200">
                    Unable to load cluster data
                  </h3>
                  <div className="mt-2 text-sm text-red-700 dark:text-red-300">
                    <p>{error}</p>
                  </div>
                  <div className="mt-4">
                    <div className="-mx-2 -my-1.5 flex">
                      <button
                        type="button"
                        onClick={handleRefresh}
                        className="bg-red-50 dark:bg-red-900/20 px-2 py-1.5 rounded-md text-sm font-medium text-red-800 dark:text-red-200 hover:bg-red-100 dark:hover:bg-red-900/30"
                      >
                        Retry
                      </button>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          )}

          <div className="grid grid-cols-1 lg:grid-cols-2 xl:grid-cols-3 gap-6">
            <div className="lg:col-span-2 xl:col-span-3">
              <ClusterHealthWidget
                clusters={clusters}
                isLoading={loading}
                onRefresh={handleRefresh}
                onClusterClick={handleClusterClick}
              />
            </div>
          </div>
        </div>
      </MainLayout>
    </>
  );
}