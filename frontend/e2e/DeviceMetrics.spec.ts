import { test, expect, signIn } from './support/fixtures';
import { readFile } from 'node:fs/promises';

const fixtureId = 'design-sensor';
const readings = [11.2, 10.8, 10.7, 10.5, 10.2, 10.1, 9.8, 9.7];
const exposes = {
  temperature: {
    name: 'temperature',
    type: 'numeric',
    category: 'measurement',
    access_mode: 'read',
    unit: '°C',
    data: 9.7,
    attributes: null,
    values: null,
  },
  humidity: {
    name: 'humidity',
    type: 'numeric',
    category: 'measurement',
    access_mode: 'read',
    unit: '%',
    data: 58,
    attributes: null,
    values: null,
  },
  contact: {
    name: 'contact',
    type: 'binary',
    category: 'measurement',
    access_mode: 'read',
    unit: '',
    data: true,
    attributes: null,
    values: { on: false, off: true },
  },
  action: {
    name: 'action',
    type: 'enum',
    category: 'measurement',
    access_mode: 'read',
    unit: '',
    data: 'single_press',
    attributes: null,
    values: null,
  },
  energy: {
    name: 'energy',
    type: 'numeric',
    category: 'measurement',
    access_mode: 'read',
    unit: 'kWh',
    data: 286.2,
    attributes: null,
    values: null,
  },
};

test.beforeEach(async ({ page }) => {
  await page.emulateMedia({ reducedMotion: 'reduce' });
  await page.addInitScript(() =>
    localStorage.setItem('nodeherder_theme', JSON.stringify({ v: 1, mode: 'dark', theme: 'graphite' }))
  );
  await page.route('**/api/hubstate', async (route) => {
    const response = await route.fetch();
    const body = await response.json();
    body.devices.push({
      ...body.devices[0],
      id: fixtureId,
      friendly_name: 'Design sensor',
      availability: 'online',
      exposes,
    });
    body.config.hub.dashboardGroups['Design checks'] = {
      name: 'Design checks',
      order: 99,
      deviceGroup: { [fixtureId]: { deviceId: fixtureId, exposes: Object.keys(exposes) } },
    };
    await route.fulfill({ response, json: body });
  });
  await page.route('**/api/metrics/query', async (route) => {
    const request = route.request().postDataJSON();
    if (request.deviceIds[0] !== fixtureId) {
      await route.continue();
      return;
    }
    const from = Date.parse(request.time.from),
      to = Date.parse(request.time.to);
    const result = request.exposes.map((name: keyof typeof exposes) => {
      const expose = exposes[name];
      const data = readings.map((value, i) => ({
        timestamp: from + Math.round(((to - from) * i) / (readings.length - 1)),
        value: expose.type === 'binary' ? i % 3 !== 1 : expose.type === 'enum' ? 'single_press' : value,
      }));
      return {
        expose: name,
        query_start_timestamp: from,
        query_end_timestamp: to,
        values: [
          {
            device_id: fixtureId,
            unit: expose.unit,
            value: { name, type: expose.type, from, to, partial: false, data },
          },
        ],
      };
    });
    await route.fulfill({ json: result });
  });
});

test('read-only environmental view has approved chart treatment and no adjustment control', async ({ page }, info) => {
  await signIn(page, '/groupdashboard');
  await page.locator(`[data-device="${fixtureId}"][data-expose="temperature"] .nh-tile-body`).click();
  const dialog = page.getByRole('dialog', { name: 'Temperature', exact: true });
  await expect(dialog.locator('.nh-history-canvas[data-ready="true"]')).toBeVisible();
  await expect(dialog.getByRole('slider')).toHaveCount(0);
  await expect(dialog.getByText('Sample average', { exact: true })).toBeVisible();
  await dialog.screenshot({ path: info.outputPath('temperature.png') });
  await dialog.getByRole('button', { name: 'Close', exact: true }).click();
  for (const width of [320, 360, 768, 1280, 1920]) {
    await page.setViewportSize({ width, height: width === 768 ? 500 : 844 });
    await page.locator(`[data-device="${fixtureId}"][data-expose="temperature"] .nh-tile-body`).click();
    const view = page.getByRole('dialog');
    await expect(view.locator('.nh-history-canvas[data-ready="true"]')).toBeVisible();
    expect(await view.evaluate((el) => el.scrollWidth <= el.clientWidth)).toBe(true);
    await view.getByRole('button', { name: 'Close', exact: true }).click();
  }
});

test('full history shares typed numeric, state and repeated-event views', async ({ page }) => {
  await signIn(page, `/devicepage/${fixtureId}?tab=history`);
  const history = page.getByRole('region', { name: 'Device history' });
  await history.getByRole('group', { name: 'Entity', exact: true }).getByRole('button', { name: 'Temperature', exact: true }).click();
  await expect(history.locator('.nh-history-canvas[data-ready="true"]')).toBeVisible();
  await history.getByRole('group', { name: 'Entity', exact: true }).getByRole('button', { name: 'Contact', exact: true }).click();
  await expect(history.getByText('Transitions', { exact: true })).toBeVisible();
  await expect(history.getByRole('group', { name: 'State history intervals' })).toBeVisible();
  await history.getByRole('group', { name: 'Entity', exact: true }).getByRole('button', { name: 'Action', exact: true }).click();
  await expect(history.getByRole('heading', { name: 'Recent events' })).toBeVisible();
  await expect(history.locator('.history-recent').getByText('Single press', { exact: true })).toHaveCount(6);
  await history.getByRole('group', { name: 'Period', exact: true }).getByRole('button', { name: 'Custom', exact: true }).click();
  await expect(history.getByText('Choose a valid period of at most 31 days.')).toBeVisible();
  await history.getByLabel('From', { exact: true }).fill('2026-10-17T10:00');
  await history.getByLabel('To', { exact: true }).fill('2026-10-17T14:00');
  await expect(history.getByRole('heading', { name: 'Recent events' })).toBeVisible();
});

test('history zoom uses mouse, wheel, keyboard and touch without additional requests or commands', async ({ page, ws }, info) => {
  let requests = 0;
  page.on('request', (r) => { if (r.url().endsWith('/api/metrics/query')) requests++; });
  await signIn(page, `/devicepage/${fixtureId}?tab=history&expose=temperature`);
  const chart = page.locator('[data-history-chart]');
  const canvas = chart.locator('canvas[data-ready="true"]');
  await expect(canvas).toBeVisible();
  const from = Number(await chart.getAttribute('data-from')), to = Number(await chart.getAttribute('data-to'));
  const initialRequests = requests, commands = ws.sentOf('deviceSetValue').length;
  const span = async () => Number(await chart.getAttribute('data-to')) - Number(await chart.getAttribute('data-from'));
  const reset = chart.getByRole('button', { name: 'Reset', exact: true });
  await chart.getByRole('button', { name: 'Zoom in', exact: true }).click();
  await expect.poll(span).toBe((to - from) / 2);
  await chart.focus();
  await chart.press('ArrowRight');
  await expect.poll(async () => Number(await chart.getAttribute('data-from'))).toBeGreaterThan(from + (to - from) / 4);
  await chart.press('0');
  await expect.poll(span).toBe(to - from);
  const box = (await canvas.boundingBox())!;
  await page.mouse.move(box.x + box.width / 2, box.y + box.height / 2);
  await expect(chart.locator('.chart-readout')).toContainText('°C');
  await page.mouse.move(box.x + box.width / 2, box.y - 2);
  await expect(chart.locator('.chart-readout')).toContainText('Drag to inspect samples');
  await page.mouse.move(box.x + box.width / 2, box.y + box.height / 2);
  await page.mouse.wheel(0, -120);
  await expect.poll(span).toBeLessThan(to - from);
  await reset.click();
  await page.mouse.move(box.x + box.width * 0.25, box.y + box.height / 2);
  await page.mouse.down();
  await page.mouse.move(box.x + box.width * 0.75, box.y + box.height / 2, { steps: 8 });
  await page.mouse.up();
  await expect.poll(span).toBeLessThan(to - from);
  await reset.click();
  if (info.project.name === 'mobile') {
    const cdp = await page.context().newCDPSession(page);
    const y = box.y + box.height / 2, x = box.x + box.width / 2;
    const point = (id: number, x: number) => ({ id, x, y });
    await cdp.send('Input.dispatchTouchEvent', { type: 'touchStart', touchPoints: [point(1, x)] });
    await cdp.send('Input.dispatchTouchEvent', { type: 'touchMove', touchPoints: [point(1, x + 25)] });
    await cdp.send('Input.dispatchTouchEvent', { type: 'touchEnd', touchPoints: [] });
    await expect(chart.locator('.chart-readout')).toContainText('°C');
    expect(await span()).toBe(to - from);
    const scroll = await page.locator('.nh-app-main').evaluate((el) => el.scrollTop);
    await cdp.send('Input.dispatchTouchEvent', { type: 'touchStart', touchPoints: [point(1, x - 20), point(2, x + 20)] });
    for (const distance of [30, 45, 65]) {
      await cdp.send('Input.dispatchTouchEvent', { type: 'touchMove', touchPoints: [point(1, x - distance), point(2, x + distance)] });
    }
    await cdp.send('Input.dispatchTouchEvent', { type: 'touchEnd', touchPoints: [] });
    await expect.poll(span).toBeLessThan(to - from);
    expect(await page.locator('.nh-app-main').evaluate((el) => el.scrollTop)).toBe(scroll);
    await reset.click();
    await cdp.send('Input.dispatchTouchEvent', { type: 'touchStart', touchPoints: [{ id: 1, x, y: y - 20 }, { id: 2, x, y: y + 20 }] });
    await cdp.send('Input.dispatchTouchEvent', { type: 'touchMove', touchPoints: [{ id: 1, x, y: y - 60 }, { id: 2, x, y: y + 60 }] });
    await cdp.send('Input.dispatchTouchEvent', { type: 'touchEnd', touchPoints: [] });
    await expect.poll(span).toBeLessThan(to - from);
    await cdp.detach();
  }
  expect(requests).toBe(initialRequests);
  expect(ws.sentOf('deviceSetValue')).toHaveLength(commands);
  expect(box.height).toBeGreaterThanOrEqual(info.project.name === 'desktop' ? 360 : 220);
  await chart.screenshot({ path: info.outputPath('history-interactions.png') });
});

test('daily statistics and both CSV schemas use raw samples and the selected period', async ({ page }, info) => {
  await signIn(page, `/devicepage/${fixtureId}?tab=history&expose=temperature`);
  const history = page.getByRole('region', { name: 'Device history' });
  await expect(history.locator('canvas[data-ready="true"]')).toBeVisible();
  await history.getByRole('button', { name: 'Statistics', exact: true }).click();
  await expect(history.getByRole('table').getByText('Samples', { exact: true })).toBeVisible();
  await expect(history.getByText('Based on 8 numeric samples', { exact: true })).toBeVisible();
  const legend = history.getByRole('list', { name: 'Statistics legend', exact: true });
  await expect(legend.getByRole('listitem')).toHaveCount(3);
  await expect(legend).toContainText('Minimum–Maximum range');
  await expect(history.getByText('Statistics CSV uses the displayed values and units.')).toBeVisible();
  await expect(history.locator('canvas[data-ready="true"]')).toBeVisible();
  await history.locator('.history-statistics').screenshot({ path: info.outputPath('daily-statistics.png') });
  const download = page.waitForEvent('download');
  await history.getByRole('button', { name: 'Export CSV' }).click();
  const csv = await readFile((await (await download).path())!, 'utf8');
  expect(csv.split('\r\n')[0]).toBe('bucket_start,bucket_end,min,average,max,last,samples,unit');
  expect(csv).toContain('°C');
  await history.getByRole('button', { name: 'History', exact: true }).click();
  await expect(history.getByText('History CSV uses native sensor values and units.')).toBeVisible();
  const rawDownload = page.waitForEvent('download');
  await history.getByRole('button', { name: 'Export CSV' }).click();
  const rawCSV = await readFile((await (await rawDownload).path())!, 'utf8');
  expect(rawCSV.trim().split('\r\n')).toHaveLength(9);
  expect(rawCSV.split('\r\n')[0]).toBe('timestamp,value,unit');
});

test('Yesterday statistics omit an empty midnight day and preserve an actual end sample on refresh', async ({ page }) => {
  let endpoint = false;
  await page.route('**/api/metrics/query', async (route) => {
    const request = route.request().postDataJSON();
    const from = Date.parse(request.time.from), to = Date.parse(request.time.to);
    const data = [{ timestamp: from + 3600000, value: 18 }, { timestamp: to - 3600000, value: 24 }];
    if (endpoint) data.push({ timestamp: to, value: 20 });
    await route.fulfill({ json: [{ expose: 'temperature', values: [{ device_id: fixtureId, unit: '°C',
      value: { name: 'temperature', type: 'numeric', from, to, data, partial: false } }] }] });
  });
  await signIn(page, `/devicepage/${fixtureId}?tab=history&expose=temperature`);
  const history = page.getByRole('region', { name: 'Device history' });
  await history.getByRole('button', { name: 'Yesterday', exact: true }).click();
  await history.getByRole('button', { name: 'Statistics', exact: true }).click();
  await expect(history.locator('.statistics-table tbody tr')).toHaveCount(1);
  await expect(history.getByText('Based on 2 numeric samples', { exact: true })).toBeVisible();
  endpoint = true;
  await history.getByRole('button', { name: 'Refresh history', exact: true }).click();
  await expect(history.getByRole('button', { name: 'Statistics', exact: true })).toHaveAttribute('aria-pressed', 'true');
  await expect(history.locator('.statistics-table tbody tr')).toHaveCount(2);
  await expect(history.getByText(/end-boundary sample/)).toBeVisible();
  await expect(history.getByText('Based on 3 numeric samples', { exact: true })).toBeVisible();
});

test('Yesterday, expose and custom bounds survive reload; invalid saved bounds fall back', async ({ page }) => {
  await signIn(page, `/devicepage/${fixtureId}?tab=history`);
  const history = page.getByRole('region', { name: 'Device history' });
  await history.getByRole('group', { name: 'Entity', exact: true }).getByRole('button', { name: 'Temperature', exact: true }).click();
  await history.getByRole('button', { name: 'Yesterday', exact: true }).click();
  await expect(history.locator('canvas[data-ready="true"]')).toBeVisible();
  await page.reload();
  await expect(history.getByRole('button', { name: 'Yesterday', exact: true })).toHaveAttribute('aria-pressed', 'true');
  await history.getByRole('button', { name: 'Custom', exact: true }).click();
  await history.getByLabel('From', { exact: true }).fill('2026-10-17T10:00');
  await history.getByLabel('To', { exact: true }).fill('2026-10-17T14:00');
  await expect(history.locator('canvas[data-ready="true"]')).toBeVisible();
  await page.reload();
  await expect(history.getByLabel('From', { exact: true })).toHaveValue('2026-10-17T10:00');
  await expect(history.getByRole('button', { name: 'Temperature', exact: true })).toHaveAttribute('aria-pressed', 'true');
  await page.evaluate((id) => localStorage.setItem(`nh-history-pref:${id}`, JSON.stringify({ version: 1, preset: 'custom', expose: 'temperature', custom: { from: 2, to: 1 } })), fixtureId);
  await page.reload();
  await expect(history.getByRole('button', { name: '24h', exact: true })).toHaveAttribute('aria-pressed', 'true');
});

test('partial history names its coverage, disables export and narrows only on request', async ({ page }) => {
  await page.route('**/api/metrics/query', async (route) => {
    const req = route.request().postDataJSON();
    const from = Date.parse(req.time.from), to = Date.parse(req.time.to);
    const data = Array.from({ length: 10001 }, (_, i) => ({ timestamp: from + (to - from) * i / 20000, value: i % 100 }));
    await route.fulfill({ json: [{ expose: 'temperature', values: [{ device_id: fixtureId, unit: '°C', value: { name: 'temperature', type: 'numeric', from, to, partial: false, data } }] }] });
  });
  await signIn(page, `/devicepage/${fixtureId}?tab=history&expose=temperature`);
  const history = page.getByRole('region', { name: 'Device history' });
  await expect(history.getByText(/Covered samples:/)).toBeVisible();
  await expect(history.getByRole('button', { name: 'Export CSV' })).toBeDisabled();
  expect(Number(await history.locator('canvas[data-ready="true"]').getAttribute('data-points'))).toBeLessThanOrEqual(1000);
  await history.getByText('Inspect recorded samples', { exact: true }).click();
  await expect(history.getByText(/Page 1 of 100/)).toBeVisible();
  await history.getByRole('button', { name: 'Older samples', exact: true }).click();
  await expect(history.getByText(/Page 2 of 100/)).toBeVisible();
  await expect(history.getByRole('button', { name: '24h', exact: true })).toHaveAttribute('aria-pressed', 'true');
  await history.getByRole('button', { name: 'Narrow this range' }).click();
  await expect(history.getByRole('button', { name: 'Custom', exact: true })).toHaveAttribute('aria-pressed', 'true');
});

test('full history stays readable at desktop, tablet and phone widths in both themes', async ({ page }, info) => {
  await signIn(page, `/devicepage/${fixtureId}?tab=history&expose=temperature`);
  const chart = page.locator('[data-history-chart]');
  for (const theme of ['dark', 'light']) {
    const toggle = page.getByRole('button', { name: `Switch to ${theme} mode` });
    if (await toggle.count()) await toggle.click();
    for (const width of [1280, 768, 390]) {
      await page.setViewportSize({ width, height: width === 768 ? 1024 : 844 });
      const canvas = chart.locator('canvas[data-ready="true"]');
      await expect(canvas).toBeVisible();
      await expect.poll(async () => (await canvas.boundingBox())!.height).toBeGreaterThanOrEqual(width >= 1024 ? 360 : width >= 720 ? 300 : 220);
      await expect.poll(() => page.locator('.nh-app-main').evaluate((el) => el.scrollWidth - el.clientWidth)).toBeLessThanOrEqual(1);
      await chart.focus();
      await chart.press('+');
      expect(await chart.evaluate((el) => getComputedStyle(el).outlineStyle)).not.toBe('none');
      await chart.press('0');
      await page.screenshot({ path: info.outputPath(`history-${theme}-${width}.png`) });
    }
  }
});
