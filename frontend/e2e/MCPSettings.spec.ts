import { test, expect, signIn } from './support/fixtures';

// Spec 059 US-03: MCP status, identity rows and process controls that wait for the hub.
// The mock hub keeps MCP state across tests, so each test normalises to stopped first.
async function stopIfRunning(page: import('@playwright/test').Page, ws: { sentOf: (type: string) => unknown[] }) {
  const stop = page.getByRole('button', { name: 'Stop', exact: true });
  if (await stop.count()) {
    const before = ws.sentOf('stopMCP').length;
    await stop.click();
    await expect.poll(() => ws.sentOf('stopMCP').length).toBe(before + 1);
    await expect(page.getByText('Disabled', { exact: true })).toBeVisible();
  }
}

test('mcp section reports status and controls the process', async ({ page, ws }) => {
  await signIn(page, '/settings?section=mcp');
  await expect(page.getByRole('heading', { level: 2, name: 'MCP server' })).toBeVisible();
  await stopIfRunning(page, ws);

  await expect(page.getByText('Disabled', { exact: true })).toBeVisible();

  const start = page.getByRole('button', { name: 'Start', exact: true });
  await expect(start).toBeEnabled();
  await start.click();
  await expect.poll(() => ws.sentOf('startMCP').length).toBe(1);
  await expect.poll(() => ws.sentOf('loadMCPStatus').length).toBeGreaterThan(0);
  await expect(page.getByText('Running', { exact: true })).toBeVisible();
  await expect(page.getByText('Mock MCP Server', { exact: true })).toBeVisible();
  await expect(page.getByText('1 client', { exact: true })).toBeVisible();

  await page.getByRole('button', { name: 'Restart', exact: true }).click();
  await expect.poll(() => ws.sentOf('restartMCP').length).toBe(1);
  await expect(page.getByText('Running', { exact: true })).toBeVisible();

  await page.getByRole('button', { name: 'Stop', exact: true }).click();
  await expect.poll(() => ws.sentOf('stopMCP').length).toBeGreaterThan(0);
  await expect(page.getByText('Disabled', { exact: true })).toBeVisible();
});

test('controls follow the reported process state', async ({ page, ws }) => {
  await signIn(page, '/settings?section=mcp');
  await stopIfRunning(page, ws);

  // Stopped: only Start is meaningful; stopping an already-stopped server is not offered.
  await expect(page.getByRole('button', { name: 'Stop', exact: true })).toHaveCount(0);
  await expect(page.getByRole('button', { name: 'Restart', exact: true })).toHaveCount(0);

  await page.getByRole('button', { name: 'Start', exact: true }).click();
  await expect.poll(() => ws.sentOf('startMCP').length).toBe(1);
  await expect(page.getByText('Running', { exact: true })).toBeVisible();
  await expect(page.getByRole('button', { name: 'Stop', exact: true })).toBeVisible();
  await expect(page.getByRole('button', { name: 'Restart', exact: true })).toBeVisible();

  await page.getByRole('button', { name: 'Stop', exact: true }).click();
  await expect(page.getByText('Disabled', { exact: true })).toBeVisible();
});
