import { test, expect, signIn } from './support/fixtures';

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
  await dialog.screenshot({ path: `/private/tmp/nodeherder-temperature-${info.project.name}.png` });
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
  await history.getByLabel('Entity', { exact: true }).selectOption('temperature');
  await expect(history.locator('.nh-history-canvas[data-ready="true"]')).toBeVisible();
  await history.getByLabel('Entity', { exact: true }).selectOption('contact');
  await expect(history.getByText('Transitions', { exact: true })).toBeVisible();
  await expect(history.getByRole('group', { name: 'State history intervals' })).toBeVisible();
  await history.getByLabel('Entity', { exact: true }).selectOption('action');
  await expect(history.getByRole('heading', { name: 'Recent events' })).toBeVisible();
  await expect(history.locator('.history-recent').getByText('Single press', { exact: true })).toHaveCount(6);
  await history.getByLabel('Period', { exact: true }).selectOption('custom');
  await expect(history.getByText('Choose a valid period of at most 31 days.')).toBeVisible();
  await history.getByLabel('From', { exact: true }).fill('2026-10-17T10:00');
  await history.getByLabel('To', { exact: true }).fill('2026-10-17T14:00');
  await expect(history.getByRole('heading', { name: 'Recent events' })).toBeVisible();
});
