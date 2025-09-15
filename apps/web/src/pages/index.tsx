import React from 'react';
import Head from 'next/head';
import { MainLayout } from '@/components/layout';
import { DashboardView } from '@/components/dashboard';

export default function Home() {
  return (
    <>
      <Head>
        <title>KubeChat Dashboard - Kubernetes Natural Language Interface</title>
        <meta name="description" content="AI-powered Kubernetes management through natural language" />
        <meta name="viewport" content="width=device-width, initial-scale=1" />
        <link rel="icon" href="/favicon.ico" />
      </Head>
      
      <MainLayout>
        <DashboardView 
          userName="Administrator"
          onRefreshAll={() => {
            console.log('Refreshing all dashboard data...');
            // TODO: Implement refresh logic when backend is integrated
          }}
        />
      </MainLayout>
    </>
  );
}