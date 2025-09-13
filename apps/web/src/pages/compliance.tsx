import React from 'react';
import Head from 'next/head';
import { MainLayout } from '@/components/layout';

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

          <div className="bg-white dark:bg-gray-900 rounded-lg border border-gray-200 dark:border-gray-700 p-6">
            <div className="text-center py-12">
              <div className="h-16 w-16 mx-auto bg-primary-100 dark:bg-primary-900 rounded-full flex items-center justify-center">
                <svg className="h-8 w-8 text-primary-600 dark:text-primary-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12.75L11.25 15 15 9.75m-3-7.036A11.959 11.959 0 013.598 6 11.99 11.99 0 003 9.749c0 5.592 3.824 10.29 9 11.623 5.176-1.332 9-6.03 9-11.622 0-1.31-.21-2.571-.598-3.751h-.152c-3.196 0-6.1-1.248-8.25-3.285z" />
                </svg>
              </div>
              <h3 className="mt-4 text-lg font-medium text-gray-900 dark:text-white">Compliance Monitoring</h3>
              <p className="mt-2 text-gray-500 dark:text-gray-400">
                Compliance dashboard will be available here. Monitor security standards and regulatory requirements.
              </p>
            </div>
          </div>
        </div>
      </MainLayout>
    </>
  );
}