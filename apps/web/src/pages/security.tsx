import React from 'react';
import Head from 'next/head';
import { MainLayout } from '@/components/layout';
import { SecurityDashboard } from '@/components/security';

export default function SecurityPage() {
  return (
    <>
      <Head>
        <title>Security Dashboard - KubeChat</title>
        <meta name="description" content="Security monitoring, user permissions, and threat detection" />
      </Head>

      <MainLayout>
        <div className="space-y-6">
          <div>
            <h1 className="text-2xl font-bold text-gray-900 dark:text-white">Security Dashboard</h1>
            <p className="text-gray-600 dark:text-gray-400 mt-1">
              Monitor user permissions, active sessions, and security events across your Kubernetes clusters
            </p>
          </div>

          <SecurityDashboard
            user={{
              id: 'user-1',
              email: 'admin@kubechat.example.com',
              username: 'administrator',
              firstName: 'System',
              lastName: 'Administrator',
              role: 'admin' as const,
              roles: [
                {
                  id: 'admin-role',
                  name: 'Administrator',
                  description: 'Full system administrator access',
                  permissions: [
                    {
                      id: 'system-admin',
                      name: 'System Administration',
                      description: 'Full administrative access to all system resources',
                      resource: '*',
                      action: '*',
                      category: 'system' as const
                    }
                  ],
                  isSystem: true
                }
              ],
              permissions: [
                {
                  id: 'system-admin',
                  name: 'System Administration',
                  description: 'Full administrative access to all system resources',
                  resource: '*',
                  action: '*',
                  category: 'system' as const
                }
              ],
              preferences: {
                theme: {
                  mode: 'system' as const,
                  primaryColor: '#3b82f6',
                  fontSize: 'medium' as const
                },
                notifications: {
                  email: true,
                  desktop: true,
                  sound: false,
                  webhooks: true
                },
                security: {
                  sessionTimeout: 3600,
                  requireTwoFactor: true,
                  allowRememberMe: false,
                  logoutOnClose: true
                },
                dashboard: {
                  defaultView: 'grid' as const,
                  refreshInterval: 30,
                  showMetrics: true,
                  compactMode: false
                },
                accessibility: {
                  highContrast: false,
                  reducedMotion: false,
                  screenReader: false,
                  keyboardNavigation: true
                },
                language: 'en',
                timezone: 'UTC'
              },
              createdAt: new Date(),
              updatedAt: new Date(),
              lastLoginAt: new Date(),
              isActive: true,
              clusters: []
            }}
          />
        </div>
      </MainLayout>
    </>
  );
}