import React from 'react';
import Head from 'next/head';
import { MainLayout } from '@/components/layout';

export default function SecurityPage() {
  return (
    <>
      <Head>
        <title>Security - KubeChat</title>
        <meta name="description" content="Security monitoring and threat detection" />
      </Head>

      <MainLayout>
        <div className="space-y-6">
          <div>
            <h1 className="text-2xl font-bold text-gray-900 dark:text-white">Security</h1>
            <p className="text-gray-600 dark:text-gray-400 mt-1">
              Monitor security threats, vulnerabilities, and access controls
            </p>
          </div>

          <div className="bg-white dark:bg-gray-900 rounded-lg border border-gray-200 dark:border-gray-700 p-6">
            <div className="text-center py-12">
              <div className="h-16 w-16 mx-auto bg-primary-100 dark:bg-primary-900 rounded-full flex items-center justify-center">
                <svg className="h-8 w-8 text-primary-600 dark:text-primary-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M16.5 10.5V6.75a4.5 4.5 0 10-9 0v3.75m-.75 11.25h10.5a2.25 2.25 0 002.25-2.25v-6.75a2.25 2.25 0 00-2.25-2.25H6.75a2.25 2.25 0 00-2.25 2.25v6.75a2.25 2.25 0 002.25 2.25z" />
                </svg>
              </div>
              <h3 className="mt-4 text-lg font-medium text-gray-900 dark:text-white">Security Dashboard</h3>
              <p className="mt-2 text-gray-500 dark:text-gray-400">
                Advanced security monitoring will be available here. Track threats, vulnerabilities, and access patterns.
              </p>
            </div>
          </div>
        </div>
      </MainLayout>
    </>
  );
}