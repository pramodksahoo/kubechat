import '../styles/globals.css';
import type { AppProps } from 'next/app';
import Head from 'next/head';
import { DesignSystemProvider } from '../design-system';
import { ErrorBoundary } from '../components/error/ErrorBoundary';
import { AccessibilityProvider } from '../design-system/accessibility';
import { GracefulDegradation } from '../components/resilience/GracefulDegradation';

export default function App({ Component, pageProps }: AppProps) {
  return (
    <ErrorBoundary>
      <GracefulDegradation>
        <DesignSystemProvider>
          <AccessibilityProvider>
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
            </Head>
            <Component {...pageProps} />
          </AccessibilityProvider>
        </DesignSystemProvider>
      </GracefulDegradation>
    </ErrorBoundary>
  );
}