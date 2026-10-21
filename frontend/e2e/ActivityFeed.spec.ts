import { test, expect, signIn } from './support/fixtures';
import { MOCK } from './support/mock';

test('activity search filters devices without changing recording, links or clear-all', async ({ page, ws }, testInfo) => {
  // Controlled incoming reports only: no device command or live hub is involved.
  const deviceErrors: string[] = [];
  page.on('console', (message) => {
    if (message.type() === 'error' && /device .* not found/.test(message.text())) deviceErrors.push(message.text());
  });
  let releaseSnapshot!: () => void;
  const snapshotGate = new Promise<void>((resolve) => { releaseSnapshot = resolve; });
  await page.route('**/api/hubstate', async (route) => {
    const response = await route.fetch();
    await snapshotGate;
    await route.fulfill({ response });
  });
  let report: (id: string, data: Record<string, number>) => void;
  await page.routeWebSocket('ws://localhost:4110/ws', (socket) => {
    report = (id, data) => socket.send(JSON.stringify({ type: 'deviceUpdated', payload: { id, data } }));
  });
  await signIn(page, '/overview');
  const feed = page.locator('.nh-activity');
  const search = feed.getByRole('searchbox', { name: 'Filter activity by device name' });
  const rows = feed.locator('.nh-feed-row');
  await expect(search).toBeVisible();
  await expect(feed.getByText(/Waiting for device changes/)).toBeVisible();
  await expect.poll(() => typeof report).toBe('function');
  const inventory = page.getByRole('link', { name: /Devices online/ }).locator('.nh-stat-num');
  // Feed rendering and the socket can beat initial HTTP on a cold CI runner.
  // Hold that response deliberately, then wait for its device data to reach Vue.
  await expect(inventory).toHaveText('0/0');
  releaseSnapshot();
  await expect(inventory).toHaveText(new RegExp(`^\\d+/${MOCK.deviceCount}$`));
  report!(MOCK.livingRoomLight.id, { brightness: 101 });
  report!(MOCK.atticSocket.id, { power: 12345 });
  await expect(rows).toHaveCount(2);
  await search.fill('  LIVING ROOM LIGHT  ');
  await expect(rows).toHaveCount(1);
  await expect(rows.first()).toHaveAttribute('href', `/devicepage/${MOCK.livingRoomLight.id}`);
  await expect(rows.first().locator('.nh-feed-dev')).toHaveText(MOCK.livingRoomLight.name);
  report!(MOCK.atticSocket.id, { power: 12346 });
  report!(MOCK.livingRoomLight.id, { brightness: 102 });
  await expect(rows).toHaveCount(2);
  await expect(feed.getByRole('status')).toHaveText('2 of 2 matches');
  await search.fill('no-such-device');
  await expect(rows).toHaveCount(0);
  await expect(feed.getByText('No activity matches this device name.')).toBeVisible();
  await feed.getByRole('button', { name: 'Clear activity filter' }).click();
  await expect(search).toBeFocused();
  await expect(search).toHaveValue('');
  await expect(rows).toHaveCount(4);
  await search.fill('Attic'); await search.press('Escape');
  await expect(search).toHaveValue(''); await expect(rows).toHaveCount(4);
  await search.fill('Attic');
  await expect(rows).toHaveCount(2);
  await search.scrollIntoViewIfNeeded();
  await page.screenshot({ path: testInfo.outputPath('activity-device-filter.png'), fullPage: true, scale: 'css' });
  expect(await page.evaluate(() => document.documentElement.scrollWidth > innerWidth)).toBe(false);
  await feed.getByRole('button', { name: 'Clear all recent activity' }).click();
  await expect(rows).toHaveCount(0);
  await expect(feed.getByText(/Waiting for device changes/)).toBeVisible();
  // Clearing all events did not erase the filter or turn recording off.
  await expect(search).toHaveValue('Attic');
  report!(MOCK.livingRoomLight.id, { brightness: 103 });
  await expect(feed.getByText('No activity matches this device name.')).toBeVisible();
  report!(MOCK.atticSocket.id, { power: 12347 });
  await expect(rows).toHaveCount(1);
  await expect(rows.first().locator('.nh-feed-dev')).toHaveText(MOCK.atticSocket.name);
  await search.fill(''); await expect(rows).toHaveCount(2);
  await feed.getByRole('switch', { name: 'Record recent activity' }).click();
  await expect(feed.getByText(/Activity recording is off/)).toBeVisible();
  expect(ws.sentOf('deviceSetValue')).toHaveLength(0);
  expect(deviceErrors).toEqual([]);
});
