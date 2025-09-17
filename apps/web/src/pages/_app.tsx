import '../styles/globals.css';
import '../styles/themes.css';
import type { AppProps } from 'next/app';
import Head from 'next/head';
import { useEffect } from 'react';
import { ErrorBoundary } from '../components/error/ErrorBoundary';
import { ThemeProvider } from '../providers/ThemeProvider';
import { initializeAuth, startTokenRefresh, stopTokenRefresh } from '../stores/authStore';

export default function App({ Component, pageProps }: AppProps) {
  useEffect(() => {
    // Initialize authentication on app start
    initializeAuth();

    // Start token refresh interval
    startTokenRefresh();

    // Cleanup on unmount
    return () => {
      stopTokenRefresh();
    };
  }, []);

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
        <Component {...pageProps} />
      </ThemeProvider>
    </ErrorBoundary>
  );
}