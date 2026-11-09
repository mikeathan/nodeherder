import { test, expect, signIn } from './support/fixtures';

// Spec 059 US-02: retention and usage in one section, with real saves and inline errors.
test('retention edits save without rewriting the other field', async ({ page, ws }) => {
  await signIn(page, '/settings?section=storage');
  await expect(page.getByRole('heading', { level: 2, name: 'Storage & retention' })).toBeVisible();

  const cleanup = page.getByRole('spinbutton', { name: 'Cleanup interval' });
  const keep = page.getByRole('spinbutton', { name: 'Keep history for' });
  await expect(cleanup).toHaveValue('12');
  await expect(keep).toHaveValue('10');
  await expect(page.getByRole('combobox', { name: 'Cleanup interval unit' })).toHaveValue('1');
  await expect(page.getByRole('combobox', { name: 'Keep history for unit' })).toHaveValue('2');

  const before = ws.sentOf('saveHistoryConfig').length;
  await cleanup.fill('6');
  await cleanup.press('Tab');
  await expect.poll(() => ws.sentOf('saveHistoryConfig').at(-1)?.payload.expireAt.value).toBe(10);
  expect(ws.sentOf('saveHistoryConfig')).toHaveLength(before + 1);
  expect(ws.sentOf('saveHistoryConfig').at(-1)?.payload.sleepTimeout).toEqual({ value: 6, unit: 'hours' });

  await keep.fill('30');
  await keep.press('Tab');
  await expect.poll(() => ws.sentOf('saveHistoryConfig').at(-1)?.payload.expireAt).toEqual({ value: 30, unit: 'days' });
  expect(ws.sentOf('saveHistoryConfig').at(-1)?.payload.sleepTimeout).toEqual({ value: 6, unit: 'hours' });

  await page.reload();
  await expect(page.getByRole('spinbutton', { name: 'Cleanup interval' })).toHaveValue('6');
  await expect(page.getByRole('spinbutton', { name: 'Keep history for' })).toHaveValue('30');

  // Restore the fixture values for other tests.
  await page.getByRole('spinbutton', { name: 'Cleanup interval' }).fill('12');
  await page.getByRole('spinbutton', { name: 'Cleanup interval' }).press('Tab');
  await page.getByRole('spinbutton', { name: 'Keep history for' }).fill('10');
  await page.getByRole('spinbutton', { name: 'Keep history for' }).press('Tab');
});

test('changing the unit keeps the value and saved unit', async ({ page, ws }) => {
  await signIn(page, '/settings?section=storage');
  const unit = page.getByRole('combobox', { name: 'Keep history for unit' });
  await unit.selectOption('hours');
  await expect.poll(() => ws.sentOf('saveHistoryConfig').at(-1)?.payload.expireAt).toEqual({ value: 10, unit: 'hours' });
  await unit.selectOption('days');
  await expect.poll(() => ws.sentOf('saveHistoryConfig').at(-1)?.payload.expireAt).toEqual({ value: 10, unit: 'days' });
});

test('invalid retention values report an inline error and send nothing', async ({ page, ws }) => {
  await signIn(page, '/settings?section=storage');
  const cleanup = page.getByRole('spinbutton', { name: 'Cleanup interval' });
  const before = ws.sentOf('saveHistoryConfig').length;
  await cleanup.fill('0');
  await cleanup.press('Tab');
  await expect(page.getByRole('alert')).toContainText('at least one minute');
  expect(ws.sentOf('saveHistoryConfig')).toHaveLength(before);
  await cleanup.fill('1.5');
  await cleanup.press('Tab');
  await expect(page.getByRole('alert')).toContainText('whole number');
  expect(ws.sentOf('saveHistoryConfig')).toHaveLength(before);
});

test('usage is read-only, formatted with units and refreshable', async ({ page }) => {
  await signIn(page, '/settings?section=storage');
  const usage = page.getByRole('heading', { level: 3, name: 'Usage' });
  await expect(usage).toBeVisible();
  await expect(page.getByText('Metrics database', { exact: true })).toBeVisible();
  await expect(page.getByText('1 MB', { exact: false }).first()).toBeVisible();
  await expect(page.getByText('Data directory', { exact: true })).toBeVisible();
  await expect(page.getByText(/free of/)).toBeVisible();

  await page.getByRole('button', { name: 'Refresh', exact: true }).click();
  await expect(page.getByText('Metrics database', { exact: true })).toBeVisible();
});

test('retention controls are disabled while the hub is disconnected', async ({ page }) => {
  let disconnected = false;
  let disconnect: (() => Promise<void>) | undefined;
  await page.routeWebSocket('ws://localhost:4110/ws', (socket) => {
    if (disconnected) { void socket.close(); return; }
    socket.connectToServer(); disconnect = () => socket.close();
  });
  await signIn(page, '/settings?section=storage');
  await expect(page.getByRole('spinbutton', { name: 'Cleanup interval' })).toBeEnabled();
  disconnected = true;
  await disconnect?.();
  await expect(page.getByRole('spinbutton', { name: 'Cleanup interval' })).toBeDisabled();
  await expect(page.getByText('Connect to the hub to change retention.', { exact: false })).toBeVisible();
  expect(await page.evaluate(() => document.documentElement.scrollWidth > innerWidth)).toBe(false);
});
