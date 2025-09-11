// KubeChat Frontend Test Setup
// Vitest and Testing Library configuration

import { beforeAll, afterEach } from 'vitest';
import { cleanup } from '@testing-library/react';
import '@testing-library/jest-dom';

// Setup before all tests
beforeAll(() => {
  // Global test setup
});

// Cleanup after each test
afterEach(() => {
  cleanup();
});
