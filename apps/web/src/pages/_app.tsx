import '../styles/globals.css';
import '../styles/themes.css';
import type { AppProps } from 'next/app';
import Head from 'next/head';
import { useEffect } from 'react';
import { ErrorBoundary } from '../components/error/ErrorBoundary';
import { ThemeProvider } from '../providers/ThemeProvider';
import { startTokenRefresh, stopTokenRefresh } from '../stores/authStore';
import { useAuthInitialization } from '../hooks/useAuthInitialization';

export default function App({ Component, pageProps }: AppProps) {
  const { isInitializing, initializationError } = useAuthInitialization();

  useEffect(() => {
    // Start token refresh interval for authenticated users
    startTokenRefresh();

    // Cleanup on unmount
    return () => {
      stopTokenRefresh();
    };
  }, []);

  if (isInitializing) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-slate-950 text-slate-200">
        <div className="space-y-4 text-center">
          <div className="mx-auto h-12 w-12 animate-spin rounded-full border-4 border-blue-500 border-t-transparent" />
          <p className="text-sm font-medium tracking-wide">Establishing secure session...</p>
        </div>
      </div>
    );
  }

  return (
    <ErrorBoundary>
      <ThemeProvider defaultTheme="system" enableSystem disableTransitionOnChange>
        <Head>
          <meta name="viewport" content="width=device-width, initial-scale=1, user-scalable=yes" />
          <meta httpEquiv="Cache-Control" content="no-cache, no-store, must-revalidate" />
          <meta httpEquiv="Pragma" content="no-cache" />
          <meta httpEquiv="Expires" content="0" />
          <title>KubeChat - Enterprise Kubernetes Management</title>
          <meta name="description" content="Professional Kubernetes cluster management with natural language interface" />
          <meta name="theme-color" content="#1e40af" />
          <link rel="preconnect" href="https://fonts.googleapis.com" />
          <link rel="preconnect" href="https://fonts.gstatic.com" crossOrigin="anonymous" />
          <link href="https://fonts.googleapis.com/css2?family=Inter:wght@300;400;500;600;700&family=JetBrains+Mono:wght@400;500;600&display=swap" rel="stylesheet" />
        </Head>
        {initializationError && (
          <div className="fixed bottom-6 right-6 rounded-lg border border-yellow-400 bg-yellow-50 px-4 py-3 text-sm text-yellow-900 shadow-lg">
            <strong className="block font-semibold">Authentication Warning</strong>
            <span>{initializationError}</span>
          </div>
        )}
        <Component {...pageProps} />
      </ThemeProvider>
    </ErrorBoundary>
  );
}