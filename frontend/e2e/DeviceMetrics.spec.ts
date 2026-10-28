import type { Page } from '@playwright/test';
import { test, expect, signIn } from './support/fixtures';
import { readFile } from 'node:fs/promises';
import { ui } from './support/ui';

const fixtureId = 'design-sensor';
const readings = [11.2, 10.8, 10.7, 10.5, 10.2, 10.1, 9.8, 9.7];
const stressSampleCount = 10_000;
const sparseContactOffsetsSeconds = [0, 60, 75, 135, 150, 210, 225, 285, 290];
type ContactSample = { timestamp: number; value: boolean | null };
/** Per-test history override: no shared module state, so parallel sharding stays safe. */
async function serveContactHistory(page: Page, samples: (from: number, to: number) => ContactSample[]) {
  await page.route('**/api/metrics/query', async (route) => {
    const request = route.request().postDataJSON();
    if (request.deviceIds[0] !== fixtureId || request.exposes.length !== 1 || request.exposes[0] !== 'contact') {
      await route.fallback();
      return;
    }
    const from = Date.parse(request.time.from);
    const to = Date.parse(request.time.to);
    await route.fulfill({
      json: [{
        expose: 'contact',
        values: [{
          device_id: fixtureId,
          unit: '',
          value: { name: 'contact', type: 'binary', from, to, partial: false, data: samples(from, to) },
        }],
      }],
    });
  });
}
function denseContactSamples(from: number, to: number): ContactSample[] {
  return Array.from({ length: stressSampleCount }, (_, i) => ({
    timestamp: from + Math.round(((to - from) * i) / (stressSampleCount - 1)),
    value: denseContactValue(i),
  }));
}
function sparseContactSamples(from: number, to: number): ContactSample[] {
  const start = to - 5 * 60000;
  return sparseContactOffsetsSeconds.map((offset, i) => ({
    timestamp: start + offset * 1000,
    value: i % 2 === 0 ? false : true,
  }));
}
/** Mirrors the summary's humanised duration so tests assert rendered text, not internals. */
function summaryDuration(ms: number): string {
  const total = Math.max(1, Math.round(ms / 60000));
  const hours = Math.floor(total / 60);
  const minutes = total % 60;
  return minutes ? `${hours} h ${minutes} min` : `${hours} h`;
}
function denseContactValue(index: number): boolean | null {
  if (index >= 5490 && index < 5510) return null;
  return index % 2 === 0 ? false : true;
}
function expectedDenseContactTotals(from: number, to: number) {
  const span = to - from;
  let active = 0;
  let known = 0;
  let transitions = 0;
  let previous: boolean | null = null;
  for (let i = 0; i < stressSampleCount - 1; i++) {
    const start = from + Math.round((span * i) / (stressSampleCount - 1));
    const end = from + Math.round((span * (i + 1)) / (stressSampleCount - 1));
    const value = denseContactValue(i);
    if (value === null) {
      previous = null;
      continue;
    }
    known += end - start;
    if (value === false) active += end - start;
    const state = value === false;
    if (previous !== null && previous !== state) transitions++;
    previous = state;
  }
  return { active, known, transitions };
}
const exposes = {
  current: { name: 'current', type: 'numeric', category: 'measurement', access_mode: 'read', unit: 'A', data: 0.8, attributes: null, values: null },
  voltage: { name: 'voltage', type: 'numeric', category: 'measurement', access_mode: 'read', unit: 'V', data: 230, attributes: null, values: null },
  power: { name: 'power', type: 'numeric', category: 'measurement', access_mode: 'read', unit: 'W', data: 80, attributes: null, values: null },
  countdown: { name: 'countdown', type: 'numeric', category: 'config', access_mode: 'read', unit: 's', data: 0, attributes: null, values: null },
  color_temp: {
    name: 'color_temp', type: 'numeric', category: 'measurement', access_mode: 'read',
    unit: 'mired', data: 300, attributes: null, values: null,
  },
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
  await page.route('**/api/home/layout', async (route) => {
    const response = await route.fetch();
    const body = await response.json();
    body.data.blocks.push({
      id: 'design-checks', kind: 'group', name: 'Design checks',
      cards: Object.keys(exposes).map((expose) => ({
        id: `design-${expose}`, kind: 'device', panel: false, device_id: fixtureId, expose,
      })),
    });
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
        value: expose.type === 'binary' ? i % 3 !== 1 : expose.type === 'enum' ? 'single_press' : name === 'color_temp' ? 350 - i * 10 : value,
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
  await ui.history.choose(page, 'Entity', 'Temperature');
  await expect(history.locator('.nh-history-canvas[data-ready="true"]')).toBeVisible();
  await ui.history.choose(page, 'Entity', 'Contact');
  await expect(history.getByText('Transitions', { exact: true })).toBeVisible();
  await expect(history.locator('.state-chart')).toHaveAttribute('data-state-mode', 'steps');
  await expect(history.getByRole('group', { name: /Inspect Contact history/ })).toBeVisible();
  await ui.history.choose(page, 'Entity', 'Action');
  await expect(history.getByRole('heading', { name: 'Recent events' })).toBeVisible();
  await expect(history.locator('.history-recent').getByText('Single press', { exact: true })).toHaveCount(6);
  await ui.history.choose(page, 'Period', 'Custom');
  await expect(history.getByText('Choose a valid period of at most 31 days.')).toBeVisible();
  await history.getByLabel('From', { exact: true }).fill('2026-10-17T10:00');
  await history.getByLabel('To', { exact: true }).fill('2026-10-17T14:00');
  await expect(history.getByRole('heading', { name: 'Recent events' })).toBeVisible();
});

test('binary history renders labelled lanes at phone, tablet and desktop widths', async ({ page }, info) => {
  await signIn(page, `/devicepage/${fixtureId}?tab=history&expose=contact`);
  const history = page.getByRole('region', { name: 'Device history' });
  const chart = history.locator('.state-chart');
  await expect(chart).toBeVisible();
  await expect(chart).toHaveAttribute('data-state-mode', 'steps');
  await expect(chart.getByRole('group', { name: /Inspect Contact history/ })).toBeVisible();
  for (const theme of ['dark', 'light']) {
    const toggle = page.getByRole('button', { name: `Switch to ${theme} mode` });
    if (await toggle.count()) await toggle.click();
    for (const width of [320, 390, 768, 1280]) {
      await page.setViewportSize({ width, height: width === 768 ? 1024 : width === 1280 ? 800 : 844 });
      await expect(chart).toBeVisible();
      await expect.poll(async () => {
        const measured = Number(await chart.getAttribute('data-layout-width'));
        return Math.abs(measured - (await chart.evaluate((node) => node.clientWidth))) <= 1;
      }).toBe(true);
      await expect(chart.locator('.state-lane-labels text')).toHaveText(['Open', 'Closed']);
      const geometry = await chart.evaluate((node) => {
        const plotTop = Number(node.getAttribute('data-plot-top'));
        const plotBottom = Number(node.getAttribute('data-plot-bottom'));
        const range = plotBottom - plotTop;
        const guide = (name: string) => Number(node.querySelector(`[data-state-lane="${name}"]`)?.getAttribute('y1'));
        const labels = Array.from(node.querySelectorAll('.state-grid text')).map((text) => {
          const box = text.getBoundingClientRect();
          return { left: box.left, right: box.right };
        }).sort((a, b) => a.left - b.left);
        return {
          width: Number(node.getAttribute('data-layout-width')),
          active: (guide('active') - plotTop) / range,
          inactive: (guide('inactive') - plotTop) / range,
          overlap: labels.some((label, index) => index > 0 && label.left < labels[index - 1].right),
        };
      });
      expect(geometry.active).toBeCloseTo(geometry.width < 560 ? 0.34 : 0.3, 2);
      expect(geometry.inactive).toBeCloseTo(geometry.width < 560 ? 0.72 : 0.7, 2);
      expect(geometry.overlap).toBe(false);
      const colors = await chart.locator('.state-lane-labels circle').evaluateAll((nodes) => nodes.map((node) => getComputedStyle(node).fill));
      expect(colors).toEqual(['rgb(217, 138, 0)', 'rgb(38, 166, 154)']);
      expect(await page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth)).toBe(true);
      await chart.screenshot({ path: info.outputPath(`binary-${theme}-${width}.png`) });
    }
  }
  await page.setViewportSize({ width: 390, height: 844 });
  await expect.poll(async () => {
    const measured = Number(await chart.getAttribute('data-layout-width'));
    return Math.abs(measured - (await chart.evaluate((node) => node.clientWidth))) <= 1;
  }).toBe(true);
  const target = chart.getByRole('group', { name: /Inspect Contact history/ });
  await target.focus();
  await expect(target).toBeFocused();
  expect(await target.evaluate((node) => getComputedStyle(node).outlineStyle)).not.toBe('none');
  await expect(chart.locator('.state-tooltip')).toContainText('Closed');
  await target.press('ArrowRight');
  await expect(chart.locator('.state-tooltip')).toContainText('Open');
  if (info.project.name === 'mobile') {
    const box = (await target.boundingBox())!;
    await target.tap({ position: { x: box.width / 2, y: box.height / 2 } });
    await expect(chart.locator('.state-tooltip')).toBeVisible();
  } else {
    await target.hover({ position: { x: 20, y: 20 } });
    await expect(chart.locator('.state-tooltip')).toBeVisible();
  }
});

test('binary history aggregates 10,000 samples within the responsive pixel budget', async ({ page, ws }, info) => {
  let served: { from: number; to: number } | null = null;
  await serveContactHistory(page, (from, to) => {
    served = { from, to };
    return denseContactSamples(from, to);
  });
  let queryCount = 0;
  page.on('request', (request) => { if (request.url().endsWith('/api/metrics/query')) queryCount++; });
  await signIn(page, `/devicepage/${fixtureId}?tab=history&expose=contact`);
  const history = page.getByRole('region', { name: 'Device history' });
  const chart = history.locator('.state-chart');
  await expect(chart).toHaveAttribute('data-state-mode', 'aggregated');
  await expect.poll(() => served).not.toBeNull();
  const expected = expectedDenseContactTotals(served!.from, served!.to);
  const summary = history.locator('.history-summary');
  for (const width of [320, 390, 768, 1280]) {
    await page.setViewportSize({ width, height: width === 768 ? 1024 : width === 1280 ? 800 : 844 });
    await expect.poll(async () => {
      const measured = Number(await chart.getAttribute('data-layout-width'));
      return Math.abs(measured - (await chart.evaluate((node) => node.clientWidth))) <= 1;
    }).toBe(true);
    await expect(chart).toHaveAttribute('data-state-mode', 'aggregated');
    const budget = await chart.evaluate((node) => ({
      width: Number(node.getAttribute('data-plot-width')),
      bins: Number(node.getAttribute('data-bin-count')),
      rendered: node.querySelectorAll('[data-state-bin="true"]').length,
      unknown: node.querySelectorAll('[data-unknown-band="true"]').length,
    }));
    expect(budget.bins).toBeLessThanOrEqual(Math.min(Math.floor(budget.width), 1000));
    expect(budget.rendered).toBeLessThanOrEqual(budget.bins);
    expect(budget.unknown).toBeGreaterThan(0);
    // Summaries stay exact and are read from the rendered text, not from chart internals.
    await expect(summary.locator('dd').nth(0)).toHaveText(summaryDuration(expected.active));
    await expect(summary.locator('dd').nth(1)).toHaveText(String(expected.transitions));
    await expect(summary.locator('dd').nth(2)).toHaveText(`${Math.round((expected.known / (served!.to - served!.from)) * 100)}%`);
    expect(await page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth)).toBe(true);
    if (width === 390 || width === 1280) await chart.screenshot({ path: info.outputPath(`binary-dense-${width}.png`) });
  }
  const target = chart.locator('.state-inspect-target');
  await target.focus();
  await expect(chart.locator('.state-tooltip')).toContainText('Open');
  await target.press('ArrowRight');
  await expect(chart.locator('.state-tooltip')).toContainText('transition');
  await expect(chart.locator('.state-tooltip')).toContainText('known');
  if (info.project.name === 'mobile') {
    expect(await target.evaluate((node) => getComputedStyle(node).touchAction)).toBe('auto');
    const box = (await target.boundingBox())!;
    await target.tap({ position: { x: box.width / 2, y: box.height / 2 } });
    await expect(chart.locator('.state-tooltip')).toBeVisible();
  } else {
    await target.hover({ position: { x: 20, y: 20 } });
    await expect(chart.locator('.state-tooltip')).toBeVisible();
  }
  expect(queryCount).toBe(1);
  expect(ws.sentOf('deviceSetValue')).toHaveLength(0);
});


test('binary history focuses recent sparse activity in the same chart without changing the selected range', async ({ page, ws }, info) => {
  await serveContactHistory(page, sparseContactSamples);
  let queries = 0;
  page.on('request', (request) => { if (request.url().endsWith('/api/metrics/query')) queries++; });
  await signIn(page, `/devicepage/${fixtureId}?tab=history&expose=contact`);
  const history = page.getByRole('region', { name: 'Device history' });
  const chart = history.locator('.state-chart');
  const summary = history.locator('.history-summary');
  const zoom = history.getByRole('button', { name: /Zoom to activity/ });
  // A five-minute burst inside 24h is unreadable without a focus action.
  await expect(zoom).toContainText('8 transitions');
  await expect(zoom).toContainText('4 min');
  expect((await zoom.boundingBox())!.height).toBeGreaterThanOrEqual(44);
  await expect(history.locator('.state-coverage')).toContainText('recorded');
  await expect(history.locator('.state-coverage')).toContainText('23 h 55 min no data');
  await expect(summary.locator('dd').nth(0)).toHaveText('4 min');
  await expect(summary.locator('dd').nth(1)).toHaveText('8');
  await expect(summary.locator('dd').nth(2)).toHaveText('<1%');
  await expect(chart.locator('[data-unknown-band="true"]').first()).toBeVisible();
  const fullFrom = Number(await chart.getAttribute('data-view-from'));
  const fullTo = Number(await chart.getAttribute('data-view-to'));
  await history.locator('.state-toolbar').screenshot({ path: info.outputPath('binary-activity-overview-toolbar.png') });
  await history.locator('.state-coverage').screenshot({ path: info.outputPath('binary-activity-coverage.png') });

  await zoom.click();
  await expect(history.locator('.state-range strong')).toHaveText('Activity detail');
  await expect(history.getByRole('status').filter({ hasText: 'Activity detail shown' })).toHaveText(/Activity detail shown/);
  await expect(summary.locator('dd').nth(0)).toHaveText('4 min');
  await expect(summary.locator('dd').nth(1)).toHaveText('8');
  await expect(history.locator('.state-coverage')).toContainText('23 h 55 min no data');
  expect(Number(await chart.getAttribute('data-view-from'))).toBeGreaterThan(fullFrom);
  expect(Number(await chart.getAttribute('data-view-to'))).toBe(fullTo);
  const reset = history.getByRole('button', { name: 'Back to full range' });
  expect((await reset.boundingBox())!.height).toBeGreaterThanOrEqual(44);
  await expect(history.getByText(/Selected range:/)).toBeVisible();
  expect(await ui.history.selected(page, 'Period')).toBe('24h');
  await history.locator('.state-toolbar').screenshot({ path: info.outputPath('binary-activity-toolbar.png') });
  await chart.screenshot({ path: info.outputPath('binary-activity-focused.png') });

  await reset.click();
  await expect(history.locator('.state-range strong')).toHaveText('Full range');
  expect(Number(await chart.getAttribute('data-view-from'))).toBe(fullFrom);
  expect(Number(await chart.getAttribute('data-view-to'))).toBe(fullTo);

  await ui.history.choose(page, 'Period', '6h');
  await expect(zoom).toBeVisible();
  await expect(summary.locator('dd').nth(1)).toHaveText('8');
  await zoom.click();
  await expect(history.locator('.state-range strong')).toHaveText('Activity detail');
  await expect(summary.locator('dd').nth(0)).toHaveText('4 min');
  await expect(history.getByText(/Selected range:/)).toBeVisible();
  await history.getByRole('button', { name: 'Back to full range' }).click();
  await expect(history.locator('.state-range strong')).toHaveText('Full range');
  expect(await ui.history.selected(page, 'Period')).toBe('6h');

  await ui.history.choose(page, 'Period', '1h');
  // A wide one-hour overview keeps the burst in context; the phone may still offer focus.
  if (info.project.name === 'desktop') {
    await expect(zoom).toHaveCount(0);
  } else {
    await expect(zoom).toBeVisible();
  }
  await chart.screenshot({ path: info.outputPath('binary-activity-1h.png') });
  expect(queries).toBe(3);
  expect(ws.sentOf('deviceSetValue')).toHaveLength(0);
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
  const reset = () => ui.history.chartAction(page, 'Reset');
  await ui.history.chartAction(page, 'Zoom in');
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
  await expect(chart.locator('.chart-readout')).toContainText('inspect');
  await page.mouse.move(box.x + box.width / 2, box.y + box.height / 2);
  await page.mouse.wheel(0, -120);
  await expect.poll(span).toBeLessThan(to - from);
  await reset();
  await page.mouse.move(box.x + box.width * 0.25, box.y + box.height / 2);
  await page.mouse.down();
  await page.mouse.move(box.x + box.width * 0.75, box.y + box.height / 2, { steps: 8 });
  await page.mouse.up();
  await expect.poll(span).toBeLessThan(to - from);
  await reset();
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
    await reset();
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
  await ui.history.setMode(page, 'Statistics');
  await expect(history.getByRole('table').getByText('Samples', { exact: true })).toBeVisible();
  await expect(history.getByText('Based on 8 numeric samples', { exact: true })).toBeVisible();
  const legend = history.getByRole('list', { name: 'Statistics legend', exact: true });
  await expect(legend.getByRole('listitem')).toHaveCount(3);
  await expect(legend).toContainText('Minimum–Maximum range');
  await expect((await ui.history.openOptions(page)).getByText('Statistics CSV uses the displayed values and units.')).toBeVisible();
  await ui.history.closeOptions(page);
  await expect(history.locator('canvas[data-ready="true"]')).toBeVisible();
  await history.locator('.history-statistics').screenshot({ path: info.outputPath('daily-statistics.png') });
  const download = page.waitForEvent('download');
  await ui.history.action(page, 'Export CSV');
  const csv = await readFile((await (await download).path())!, 'utf8');
  expect(csv.split('\r\n')[0]).toBe('bucket_start,bucket_end,min,average,max,last,samples,unit');
  expect(csv).toContain('°C');
  await ui.history.setMode(page, 'History');
  await expect((await ui.history.openOptions(page)).getByText('History CSV uses native sensor values and units.')).toBeVisible();
  const rawDownload = page.waitForEvent('download');
  await ui.history.action(page, 'Export CSV');
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
  await ui.history.choose(page, 'Period', 'Yesterday');
  await ui.history.setMode(page, 'Statistics');
  await expect(history.locator('.statistics-table tbody tr')).toHaveCount(1);
  await expect(history.getByText('Based on 2 numeric samples', { exact: true })).toBeVisible();
  endpoint = true;
  await ui.history.action(page, 'Refresh history');
  await expect.poll(() => ui.history.currentMode(page)).toBe('Statistics');
  await expect(history.locator('.statistics-table tbody tr')).toHaveCount(2);
  await expect(history.getByText(/end-boundary sample/)).toBeVisible();
  await expect(history.getByText('Based on 3 numeric samples', { exact: true })).toBeVisible();
});

test('Yesterday, expose and custom bounds survive reload; invalid saved bounds fall back', async ({ page }) => {
  await signIn(page, `/devicepage/${fixtureId}?tab=history`);
  const history = page.getByRole('region', { name: 'Device history' });
  await ui.history.choose(page, 'Entity', 'Temperature');
  await ui.history.choose(page, 'Period', 'Yesterday');
  await expect(history.locator('canvas[data-ready="true"]')).toBeVisible();
  await page.reload();
  await expect.poll(() => ui.history.selected(page, 'Period')).toBe('Yesterday');
  await ui.history.choose(page, 'Period', 'Custom');
  await history.getByLabel('From', { exact: true }).fill('2026-10-17T10:00');
  await history.getByLabel('To', { exact: true }).fill('2026-10-17T14:00');
  await expect(history.locator('canvas[data-ready="true"]')).toBeVisible();
  await page.reload();
  await expect(history.getByLabel('From', { exact: true })).toHaveValue('2026-10-17T10:00');
  await expect.poll(() => ui.history.selected(page, 'Entity')).toBe('Temperature');
  await page.evaluate((id) => localStorage.setItem(`nh-history-pref:${id}`, JSON.stringify({ version: 1, preset: 'custom', expose: 'temperature', custom: { from: 2, to: 1 } })), fixtureId);
  await page.reload();
  await expect.poll(() => ui.history.selected(page, 'Period')).toBe('24h');
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
  await expect((await ui.history.openOptions(page)).getByRole('button', { name: 'Export CSV' })).toBeDisabled();
  await ui.history.closeOptions(page);
  expect(Number(await history.locator('canvas[data-ready="true"]').getAttribute('data-points'))).toBeLessThanOrEqual(1000);
  await history.getByText('Inspect recorded samples', { exact: true }).click();
  await expect(history.getByText(/Page 1 of 100/)).toBeVisible();
  await history.getByRole('button', { name: 'Older samples', exact: true }).click();
  await expect(history.getByText(/Page 2 of 100/)).toBeVisible();
  await expect.poll(() => ui.history.selected(page, 'Period')).toBe('24h');
  await history.getByRole('button', { name: 'Narrow this range' }).click();
  await expect.poll(() => ui.history.selected(page, 'Period')).toBe('Custom');
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
      await expect.poll(() => ui.history.selected(page, 'Entity')).toBe('Temperature');
      await expect.poll(() => ui.history.selected(page, 'Period')).toBe('24h');
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

test('desktop history uses a quiet labelled toolbar for many entities and ranges', async ({ page }, info) => {
  await page.setViewportSize({ width: 1280, height: 900 });
  await signIn(page, `/devicepage/${fixtureId}?tab=history`);
  await ui.history.choose(page, 'Entity', 'Current');
  const history = page.getByRole('region', { name: 'Device history' });
  const toolbar = history.locator('.history-toolbar');
  await expect(history.locator('canvas[data-ready="true"]')).toBeVisible();
  expect((await toolbar.boundingBox())!.height).toBeLessThanOrEqual(112);
  const entity = toolbar.getByRole('combobox', { name: 'Entity', exact: true });
  const period = toolbar.getByRole('combobox', { name: 'Period', exact: true });
  await expect(entity.getByRole('option')).toHaveCount(10);
  await expect(period.getByRole('option')).toHaveCount(9);
  await expect(period.getByRole('option', { name: '3d', exact: true })).toHaveCount(1);
  await entity.selectOption('voltage');
  await period.selectOption('168');
  await expect(history.locator('canvas[data-ready="true"]')).toBeVisible();
  for (const theme of ['light', 'dark']) {
    const toggle = page.getByRole('button', { name: `Switch to ${theme} mode` });
    if (await toggle.isVisible()) await toggle.click();
    for (const width of [768, 1024, 1280, 1920]) {
      await page.setViewportSize({ width, height: 900 });
      await expect(entity).toHaveValue('voltage');
      await expect(period).toHaveValue('168');
      await expect(toolbar.getByRole('group', { name: 'Entity', exact: true }).getByRole('button')).toHaveCount(0);
      await expect(toolbar.getByRole('group', { name: 'Period', exact: true }).getByRole('button')).toHaveCount(0);
      const boxes = await Promise.all([entity, period, toolbar.getByRole('button', { name: 'History', exact: true }), toolbar.getByRole('button', { name: 'Export CSV', exact: true })].map(async (control) => (await control.boundingBox())!));
      const centres = boxes.map((box) => box.y + box.height / 2);
      expect(Math.max(...centres) - Math.min(...centres)).toBeLessThanOrEqual(2);
      for (let i = 1; i < boxes.length; i++) {
        expect(boxes[i - 1].x + boxes[i - 1].width).toBeLessThanOrEqual(boxes[i].x + 1);
      }
      expect(await page.locator('.nh-app-main').evaluate((el) => el.scrollWidth - el.clientWidth)).toBeLessThanOrEqual(1);
      await toolbar.screenshot({ path: info.outputPath(`desktop-toolbar-${theme}-${width}.png`) });
    }
  }
  await page.reload();
  await expect(entity).toHaveValue('voltage');
  await expect(period).toHaveValue('168');
  await expect(history.locator('canvas[data-ready="true"]')).toBeVisible();
  await page.setViewportSize({ width: 384, height: 854 });
  await expect(entity).toHaveValue('voltage');
  await expect(period).toHaveValue('168');
  expect((await toolbar.boundingBox())!.height).toBe(44);
  await expect(toolbar.getByRole('button', { name: 'History options', exact: true })).toBeVisible();
});

test('phone metrics prioritise a large plot with options and chart tools on demand', async ({ page }, info) => {
  await page.setViewportSize({ width: 384, height: 854 });
  await signIn(page, `/devicepage/${fixtureId}?tab=history&expose=color_temp`);
  const history = page.getByRole('region', { name: 'Device history' });
  const toolbar = history.locator('.history-toolbar');
  const chart = history.locator('[data-history-chart]');
  const canvas = chart.locator('canvas[data-ready="true"]');
  await expect(canvas).toBeVisible();
  for (const mode of ['light', 'dark']) {
    const toggle = page.getByRole('button', { name: `Switch to ${mode} mode` });
    if (await toggle.isVisible()) await toggle.click();
    for (const width of [320, 384, 390]) {
      await page.setViewportSize({ width, height: 854 });
      const controls = (await toolbar.boundingBox())!;
      expect(controls.height).toBeLessThanOrEqual(48);
      await expect.poll(async () => (await canvas.boundingBox())!.height).toBeGreaterThanOrEqual(400);
      const plot = (await canvas.boundingBox())!;
      expect(plot.y - controls.y - controls.height).toBeLessThanOrEqual(12);
      for (const control of await toolbar.locator('button:visible, select:visible').all()) {
        const box = (await control.boundingBox())!;
        expect(box.height).toBeGreaterThanOrEqual(44);
        expect(box.width).toBeGreaterThanOrEqual(44);
      }
      await expect(chart.getByRole('button', { name: 'Zoom in', exact: true })).not.toBeVisible();
      const tools = chart.getByText('Chart tools', { exact: true });
      expect((await tools.boundingBox())!.y).toBeGreaterThanOrEqual(plot.y + plot.height);
      expect(await history.evaluate((el) => el.scrollWidth - el.clientWidth)).toBeLessThanOrEqual(1);
      await page.screenshot({ path: info.outputPath(`chart-first-${mode}-${width}.png`) });
    }
  }
  await page.setViewportSize({ width: 390, height: 568 });
  await expect.poll(async () => (await canvas.boundingBox())!.height).toBeGreaterThanOrEqual(280);
  await page.screenshot({ path: info.outputPath('chart-first-short.png') });
  await page.setViewportSize({ width: 384, height: 854 });
  const period = toolbar.getByRole('combobox', { name: 'Period', exact: true });
  await period.selectOption({ label: 'Yesterday' });
  await expect(canvas).toBeVisible();
  await page.reload();
  await expect(period).toHaveValue('yesterday');
  const entity = toolbar.getByRole('combobox', { name: 'Entity', exact: true });
  await expect(entity).toHaveValue('color_temp');
  await expect(canvas).toBeVisible();
  await entity.focus();
  expect(await entity.evaluate((el) => getComputedStyle(el).outlineStyle)).not.toBe('none');
  await entity.press('Tab');
  await expect(period).toBeFocused();
  await period.press('Tab');
  const options = toolbar.getByRole('button', { name: 'History options', exact: true });
  await expect(options).toBeFocused();
  await page.keyboard.press('Space');
  const sheet = page.getByRole('dialog', { name: 'History options', exact: true });
  await expect(sheet).toBeVisible();
  for (const control of await sheet.locator('button:visible, select:visible').all()) {
    const box = (await control.boundingBox())!;
    expect(box.width).toBeGreaterThanOrEqual(44);
    expect(box.height).toBeGreaterThanOrEqual(44);
  }
  await sheet.screenshot({ path: info.outputPath('history-options.png') });
  await expect(sheet.getByText('History CSV uses native sensor values and units.')).toBeVisible();
  const browserZone = await page.evaluate(() => Intl.DateTimeFormat().resolvedOptions().timeZone);
  await expect(sheet.getByText(browserZone, { exact: false })).toBeVisible();
  await page.keyboard.press('Escape');
  await expect(options).toBeFocused();
  const tools = chart.locator('.mobile-chart-tools');
  await tools.locator('summary').focus();
  await tools.locator('summary').press('Enter');
  await expect(chart.getByRole('button', { name: 'Zoom in', exact: true })).toBeVisible();
  await tools.screenshot({ path: info.outputPath('chart-tools.png') });
  await tools.locator('summary').press('Escape');
  await expect(chart.getByRole('button', { name: 'Zoom in', exact: true })).not.toBeVisible();
  await options.click();
  await sheet.getByRole('combobox', { name: 'View', exact: true }).selectOption('statistics');
  await expect(sheet).not.toBeVisible();
  await expect(options).toBeFocused();
  await expect(options).toHaveAttribute('aria-description', 'Current view: Statistics');
  const stats = history.getByRole('region', { name: 'Sample statistics' });
  const statsCanvas = stats.locator('canvas[data-ready="true"]');
  await expect(statsCanvas).toBeVisible();
  expect((await statsCanvas.boundingBox())!.y).toBeLessThan((await stats.locator('.history-summary').boundingBox())!.y);
  await page.screenshot({ path: info.outputPath('chart-first-statistics.png') });
  await period.selectOption({ label: 'Custom' });
  await expect(history.getByLabel('From', { exact: true })).toBeVisible();
  await options.click();
  await expect(sheet.getByRole('button', { name: 'Export CSV', exact: true })).toBeDisabled();
});
