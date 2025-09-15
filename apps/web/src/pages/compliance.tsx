import React from 'react';
import Head from 'next/head';
import { MainLayout } from '@/components/layout';
import { ComplianceDashboard } from '@/components/compliance';

export default function CompliancePage() {
  return (
    <>
      <Head>
        <title>Compliance - KubeChat</title>
        <meta name="description" content="Monitor compliance and security standards" />
      </Head>

      <MainLayout>
        <div className="space-y-6">
          <div>
            <h1 className="text-2xl font-bold text-gray-900 dark:text-white">Compliance</h1>
            <p className="text-gray-600 dark:text-gray-400 mt-1">
              Monitor compliance status and ensure adherence to security standards
            </p>
          </div>

          <ComplianceDashboard
            complianceStatuses={[]}
            auditSummary={{
              totalEvents: 0,
              successEvents: 0,
              failureEvents: 0,
              errorEvents: 0,
              criticalEvents: 0,
              topUsers: [],
              topActions: [],
              timeline: []
            }}
            recentFindings={[]}
            onGenerateReport={async () => {}}
            onUpdateFinding={async () => {}}
            onRefreshData={async () => {}}
          />
        </div>
      </MainLayout>
    </>
  );
}