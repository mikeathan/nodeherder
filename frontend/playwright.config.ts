import { defineConfig, devices } from '@playwright/test';

// Browser end-to-end tests against the mock hub (tools/server). Spec 007 NFR-06, ADR-002.
// Never point these tests at a real hub: they toggle devices and save automations.
const CI = !!process.env.CI;

export default defineConfig({
  testDir: './e2e',
  fullyParallel: false,
  workers: 1, // the mock hub keeps in-memory state shared by all tests
  forbidOnly: CI,
  retries: 0,
  timeout: 30_000,
  expect: { timeout: 7_000 },
  reporter: CI ? [['list'], ['html', { open: 'never' }]] : 'list',
  use: {
    baseURL: 'http://localhost:4100',
    trace: 'retain-on-failure',
    screenshot: 'only-on-failure',
  },
  projects: [
    { name: 'desktop', use: { ...devices['Desktop Chrome'], viewport: { width: 1280, height: 800 } } },
    { name: 'mobile', use: { ...devices['Pixel 7'], viewport: { width: 390, height: 844 } } },
  ],
  webServer: [
    { command: 'node tools/server/index.mjs', port: 4110, reuseExistingServer: !CI, stdout: 'ignore' },
    { command: 'npx vite --mode e2e --port 4100 --strictPort', url: 'http://localhost:4100', reuseExistingServer: !CI, timeout: 120_000 },
  ],
});
