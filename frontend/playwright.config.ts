import { defineConfig, devices } from '@playwright/test';
import { randomUUID } from 'node:crypto';

// Browser tests use the Node mock hub plus an isolated Go hub with in-memory MQTT.
// Never point these tests at a real hub: they toggle devices and save automations.
const CI = !!process.env.CI;
// Set E2E_EXTERNAL_SERVERS=1 with E2E_BASE_URL/E2E_HUB_API to target a manually
// started isolated servers (including the Go fixture on 4120).
const baseURL = process.env.E2E_BASE_URL ?? 'http://localhost:4100';
const externalServers = process.env.E2E_EXTERNAL_SERVERS === '1';
if (externalServers && !process.env.NODEHERDER_E2E_RUN_ID) throw new Error('External servers require a matching NODEHERDER_E2E_RUN_ID fixture handshake.');
process.env.NODEHERDER_E2E_RUN_ID ??= randomUUID();

export default defineConfig({
  globalSetup: './e2e/support/global-setup.ts',
  testDir: './e2e',
  // Shard by individual test, not by file: file-level sharding left one job with most of
  // EntityViewDialog/DeviceMetrics while another finished in half the time. Workers stay
  // at one because the in-memory hubs are shared by all tests in a job.
  fullyParallel: true,
  workers: 1,
  forbidOnly: CI,
  // One retry on CI, none locally.
  //
  // These tests drive real servers, sockets and animations, so a small number of them can lose a
  // race that has nothing to do with the product: a frame that had not settled, a socket that
  // closed a millisecond early. With `retries: 0` any one of those fails the build, which trains
  // people to re-run rather than read the failure, and that is how a genuine regression gets waved
  // through. A retry keeps the signal honest: a real defect fails both attempts, and a flake is
  // reported as "flaky" rather than hiding. Reproduce locally by running the file directly.
  retries: CI ? 1 : 0,
  timeout: 30_000,
  expect: { timeout: 7_000 },
  reporter: CI ? [['list'], ['html', { open: 'never' }]] : 'list',
  use: {
    baseURL,
    trace: 'retain-on-failure',
    screenshot: 'only-on-failure',
  },
  projects: [
    { name: 'desktop', use: { ...devices['Desktop Chrome'], viewport: { width: 1280, height: 800 } } },
    { name: 'mobile', use: { ...devices['Pixel 7'], viewport: { width: 390, height: 844 } } },
  ],
  webServer: externalServers
    ? undefined
    : [
        { command: 'node tools/server/index.mjs', env: {NODEHERDER_E2E_RUN_ID:process.env.NODEHERDER_E2E_RUN_ID}, url: 'http://localhost:4110/api/__test/health', reuseExistingServer: false, stdout: 'ignore' },
        {
          command: 'go test -count=1 -timeout=20m -run "^TestAutomationBrowserHub$" ./internal',
          cwd: '../backend',
          env: { NODEHERDER_BROWSER_HUB: '1',NODEHERDER_E2E_RUN_ID:process.env.NODEHERDER_E2E_RUN_ID },
          url: 'http://localhost:4120/api/__test/health',
          reuseExistingServer: false,
          // CI warms the Go build cache first; local cold runs may need this window.
          timeout: 300_000,
          stdout: 'pipe',
        },
        { command: 'npx vite --mode e2e --port 4100 --strictPort', url: 'http://localhost:4100', reuseExistingServer: false, timeout: 120_000 },
      ],
});
