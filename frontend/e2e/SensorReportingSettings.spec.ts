import { test, expect, signIn } from './support/fixtures';
import { MOCK } from './support/mock';
import type { Page } from '@playwright/test';

async function addSensor(page: Page, name: string, history = false, global = false) {
  const group = page.getByRole('group', { name: history ? 'History sensor adjustments' : 'Sensor adjustments', exact: true });
  const picker = group.getByLabel(history ? 'History sensor' : 'Sensor to adjust', { exact: true });
  if (global) await picker.fill(name); else await picker.selectOption(name);
  await group.getByRole('button', { name: 'Add adjustment', exact: true }).click();
  return group;
}

async function customInterval(page: Page, label: string, seconds: string) {
  await page.getByRole('combobox', { name: label, exact: true }).selectOption('custom');
  const input = page.getByRole('spinbutton', { name: `${label} seconds`, exact: true });
  await input.fill(seconds);
  await input.press('Tab');
  return input;
}

const openHistory = (page: Page) => page.locator('summary').filter({ hasText: /^History recording/ }).click();
const openAdvanced = (page: Page) => page.locator('summary').filter({ hasText: /^Advanced Command confirmation/ }).click();

test('sensor adjustment is simple, saves native-unit thresholds and resets live independently of history', async ({ page, ws }, testInfo) => {
  await signIn(page, `/devicepage/${MOCK.presenceAutomation.id}?tab=settings`);
  await expect(page.getByRole('heading', { name: 'Dashboard & activity updates', exact: true })).toBeVisible();
  let input = page.getByRole('spinbutton', { name: 'Ignore changes smaller than for target_distance' });
  const beforeAdd = ws.sentOf('saveDeviceConfigOverride').length;
  if (!await input.count()) await addSensor(page, 'target_distance');
  expect(ws.sentOf('saveDeviceConfigOverride')).toHaveLength(beforeAdd);
  await expect(page.getByRole('combobox', { name: 'Update timing for target_distance' })).toBeHidden();
  await expect(page.getByRole('combobox', { name: 'History for target_distance' })).toBeHidden();
  await input.fill('0.2'); await input.press('Tab');
  await expect.poll(() => ws.sentOf('saveDeviceConfigOverride').at(-1)?.payload.reportingThresholds?.target_distance).toBe(.2);
  await openHistory(page);
  if (!await page.getByRole('combobox', { name: 'History for target_distance' }).count()) await addSensor(page, 'target_distance', true);
  await page.getByRole('combobox', { name: 'History for target_distance' }).selectOption('true');
  await openAdvanced(page);
  const timeout = page.getByRole('spinbutton', { name: 'Command confirmation timeout (milliseconds)' });
  await timeout.fill('12000'); await timeout.press('Tab');
  await expect.poll(() => ws.sentOf('saveDeviceConfigOverride').at(-1)?.payload.commandConfirmationTimeoutMs).toBe(12000);
  await page.reload();
  await page.getByRole('heading', { name: 'Dashboard & activity updates', exact: true }).scrollIntoViewIfNeeded();
  await page.screenshot({ path: testInfo.outputPath('reporting-main.png'), fullPage: true, scale: 'css' });
  input = page.getByRole('spinbutton', { name: 'Ignore changes smaller than for target_distance' });
  await expect(page.locator('summary').filter({ hasText: /^Advanced update timing/ })).toContainText('Every change');
  await page.getByRole('button', { name: 'Switch to dark mode', exact: true }).click();
  await input.scrollIntoViewIfNeeded();
  await page.screenshot({ path: testInfo.outputPath('sensor-adjustment.png'), fullPage: true, scale: 'css' });
  await expect(input).toHaveValue('0.2');
  await openHistory(page);
  await expect(page.getByRole('combobox', { name: 'History for target_distance' })).toHaveValue('true');
  await openAdvanced(page); await expect(timeout).toHaveValue('12000');
  await input.focus(); await expect(input).toBeFocused();
  await input.press('Tab'); // Native keyboard navigation leaves the number control.
  await expect(input).not.toBeFocused();
  await input.fill('0'); await input.press('Tab');
  await expect.poll(() => ws.sentOf('saveDeviceConfigOverride').at(-1)?.payload.reportingThresholds?.target_distance).toBe(0);
  await page.getByRole('button', { name: 'Reset Target distance adjustment', exact: true }).click();
  await expect(input).toHaveCount(0);
  await expect.poll(() => ws.sentOf('saveDeviceConfigOverride').at(-1)?.payload.reportingThresholds).toEqual({});
  await expect.poll(() => ws.sentOf('saveDeviceConfigOverride').at(-1)?.payload.metricOverrides?.target_distance).toBe(true);
  await page.screenshot({ path: testInfo.outputPath('reporting-simplified.png'), fullPage: true, scale: 'css' });
  expect(await page.evaluate(() => document.documentElement.scrollWidth > innerWidth)).toBe(false);
});

test('shared sensor rules and confirmation timing survive reload', async ({ page, ws }) => {
  await signIn(page, '/settings?section=device-defaults');
  await openAdvanced(page);
  const timeout = page.getByRole('spinbutton', { name: 'Command confirmation timeout (milliseconds)' });
  const originalTimeout = await timeout.inputValue();
  const beforeAdd = ws.sentOf('saveDeviceConfigDefaults').length;
  await addSensor(page, 'new_sensor_reading', false, true);
  expect(ws.sentOf('saveDeviceConfigDefaults')).toHaveLength(beforeAdd);
  const input = page.getByRole('spinbutton', { name: 'Ignore changes smaller than for new_sensor_reading' });
  await input.fill('7.5'); await input.press('Tab');
  await openHistory(page); await addSensor(page, 'new_sensor_reading', true, true);
  await page.getByRole('combobox', { name: 'History for new_sensor_reading' }).selectOption('true');
  await timeout.fill('2500'); await timeout.press('Tab');
  await expect.poll(() => ws.sentOf('saveDeviceConfigDefaults').at(-1)?.payload.commandConfirmationTimeoutMs).toBe(2500);
  await page.reload(); await expect(input).toHaveValue('7.5');
  await openAdvanced(page); await expect(timeout).toHaveValue('2500');
  await openHistory(page);
  await expect(page.getByRole('combobox', { name: 'History for new_sensor_reading' })).toHaveValue('true');
  await page.getByRole('button', { name: 'Reset New sensor reading adjustment', exact: true }).click();
  await page.getByRole('button', { name: 'Reset New sensor reading history', exact: true }).click();
  await timeout.fill(originalTimeout); await timeout.press('Tab');
  expect(await page.evaluate(() => document.documentElement.scrollWidth > innerWidth)).toBe(false);
});

test('shared category updates and advanced sensor timing save independently of history', async ({ page, ws }, testInfo) => {
  await signIn(page, '/settings?section=device-defaults');
  const diagnostics = await customInterval(page, 'Diagnostics updates', '15');
  await expect.poll(() => ws.sentOf('saveDeviceConfigDefaults').at(-1)?.payload.liveIntervalByCategory?.diagnostic?.value).toBe(15000);
  await openHistory(page);
  const historyMode = page.getByRole('combobox', { name: 'Diagnostics history timing', exact: true });
  await expect(historyMode).toHaveValue('custom');
  await expect(page.getByRole('spinbutton', { name: 'Diagnostics history timing seconds', exact: true })).toHaveValue('300');
  await page.reload(); await expect(diagnostics).toHaveValue('15');
  await page.goto(`/devicepage/${MOCK.presenceAutomation.id}?tab=settings`);
  await expect(page.getByText('At most every 15 seconds · Shared default', { exact: true })).toBeVisible();
  await addSensor(page, 'linkquality');
  await page.locator('summary').filter({ hasText: /^Advanced update timing/ }).click();
  await page.getByRole('combobox', { name: 'Update timing for linkquality', exact: true }).selectOption('every');
  await expect.poll(() => ws.sentOf('saveDeviceConfigOverride').at(-1)?.payload.liveIntervalOverrides?.linkquality?.value).toBe(0);
  await page.reload();
  await page.locator('summary').filter({ hasText: /^Advanced update timing/ }).click();
  await expect(page.getByRole('combobox', { name: 'Update timing for linkquality', exact: true })).toHaveValue('every');
  await page.getByRole('button', { name: 'Reset Link quality adjustment', exact: true }).click();
  await page.screenshot({ path: testInfo.outputPath('device-category-inheritance.png'), fullPage: true, scale: 'css' });
  expect(await page.evaluate(() => document.documentElement.scrollWidth > innerWidth)).toBe(false);
  await page.goto('/settings?section=device-defaults');
  await customInterval(page, 'Diagnostics updates', '30');
});

test('device categories support custom, zero and reset without changing shared defaults or history', async ({ page, ws }, testInfo) => {
  if (testInfo.project.name === 'desktop') await page.setViewportSize({ width: 1440, height: 900 });
  await signIn(page, `/devicepage/${MOCK.presenceAutomation.id}?tab=settings`);
  const mode = page.getByRole('combobox', { name: 'Diagnostics updates', exact: true });
  await mode.selectOption('inherit');
  const beforeCustom = ws.sentOf('saveDeviceConfigOverride').length;
  await mode.selectOption('custom');
  expect(ws.sentOf('saveDeviceConfigOverride')).toHaveLength(beforeCustom);
  await customInterval(page, 'Diagnostics updates', '10');
  await customInterval(page, 'Measurements updates', '5');
  await expect.poll(() => ws.sentOf('saveDeviceConfigOverride').at(-1)?.payload.liveIntervalByCategory?.measurement?.value).toBe(5000);
  await openHistory(page); await customInterval(page, 'Diagnostics history timing', '20');
  await expect.poll(() => ws.sentOf('saveDeviceConfigOverride').at(-1)?.payload.defaultDebounceByCategory?.diagnostic?.value).toBe(20000);
  await page.reload();
  await expect(page.getByRole('spinbutton', { name: 'Diagnostics updates seconds', exact: true })).toHaveValue('10');
  await expect(page.getByRole('spinbutton', { name: 'Measurements updates seconds', exact: true })).toHaveValue('5');
  await mode.focus(); await expect(mode).toBeFocused();
  await mode.selectOption('every');
  await expect.poll(() => ws.sentOf('saveDeviceConfigOverride').at(-1)?.payload.liveIntervalByCategory?.diagnostic?.value).toBe(0);
  await page.reload(); await expect(mode).toHaveValue('every');
  await mode.selectOption('inherit');
  await expect.poll(() => ws.sentOf('saveDeviceConfigOverride').at(-1)?.payload.liveIntervalByCategory?.diagnostic).toBeUndefined();
  await page.getByRole('combobox', { name: 'Measurements updates', exact: true }).selectOption('inherit');
  await openHistory(page);
  await expect(page.getByRole('spinbutton', { name: 'Diagnostics history timing seconds', exact: true })).toHaveValue('20');
  await page.getByRole('combobox', { name: 'Diagnostics history timing', exact: true }).selectOption('inherit');
  await openHistory(page);
  await page.screenshot({ path: testInfo.outputPath('category-overrides-clean.png'), fullPage: true, scale: 'css' });
  expect(await page.evaluate(() => document.documentElement.scrollWidth > innerWidth)).toBe(false);
  await page.goto('/settings?section=device-defaults');
  await expect(page.getByRole('spinbutton', { name: 'Diagnostics updates seconds', exact: true })).toHaveValue('30');
  await expect(page.getByRole('combobox', { name: 'Measurements updates', exact: true })).toHaveValue('every');
});

test('invalid reporting intervals show feedback without saving', async ({ page, ws }) => {
  await signIn(page, '/settings?section=device-defaults');
  const before = ws.sentOf('saveDeviceConfigDefaults').length;
  await customInterval(page, 'Diagnostics updates', '-1');
  await expect(page.getByRole('alert')).toContainText('Enter zero or a positive interval');
  expect(ws.sentOf('saveDeviceConfigDefaults')).toHaveLength(before);
});

test('reporting controls explain and block edits while disconnected', async ({ page }) => {
  let disconnected = false;
  let disconnect: (() => Promise<void>) | undefined;
  await page.routeWebSocket('ws://localhost:4110/ws', (socket) => {
    if (disconnected) { void socket.close(); return; }
    socket.connectToServer(); disconnect = () => socket.close();
  });
  await signIn(page, '/settings?section=device-defaults');
  const mode = page.getByRole('combobox', { name: 'Diagnostics updates', exact: true });
  await expect(mode).toBeEnabled();
  expect(disconnect).toBeDefined(); disconnected = true; await disconnect?.();
  await expect(mode).toBeDisabled();
  await expect(page.getByText('Connect to the hub to change reporting settings.', { exact: true })).toBeVisible();
});
