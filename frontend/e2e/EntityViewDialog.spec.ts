import { test, expect, signIn, dragBetween } from './support/fixtures';
import { MOCK } from './support/mock';
import { ui } from './support/ui';

// EntityViewDialog.vue acceptance flows against the isolated mock hub only.
test.beforeEach(async ({ page }) => {
  await page.emulateMedia({ reducedMotion: 'reduce' });
  await page.addInitScript(() =>
    localStorage.setItem('nodeherder_theme', JSON.stringify({ v: 1, mode: 'dark', theme: 'graphite' }))
  );
  await page.route('**/api/hubstate', async (route) => {
    const response = await route.fetch();
    const body = await response.json();
    const light = body.devices.find((d: { id: string }) => d.id === MOCK.livingRoomLight.id);
    light.availability = 'online';
    light.exposes.brightness.data = 184;
    Object.assign(light.exposes.color_temp, { data: 370, unit: 'mired', attributes: { min: 153, max: 500 } });
    await route.fulfill({ response, json: body });
  });
});

test('brightness uses native values, fullscreen mobile, shared confirmation and minimal history', async ({
  page,
  ws,
}, info) => {
  const errors: string[] = [];
  page.on('pageerror', (error) => errors.push(error.message));
  await signIn(page, ui.home.path);
  const tile = ui.home.tile(page, 'living room group', 'Brightness');
  const opener = tile.locator('.nh-tile-body');
  await opener.click();
  const dialog = page.getByRole('dialog', { name: 'Brightness', exact: true });
  await expect(dialog).toBeVisible();
  await expect(dialog.getByLabel('Brightness preview')).toHaveText('72 %');
  await expect(dialog.locator('.nh-history-canvas[data-ready="true"]')).toBeVisible();
  await dialog.screenshot({ path: `/private/tmp/nodeherder-brightness-${info.project.name}.png` });
  const slider = dialog.getByRole('slider', { name: 'Brightness', exact: true });
  await slider.focus();
  await slider.press('ArrowUp');
  await ws.waitForSent(
    'deviceSetValue',
    (p) => p.id === MOCK.livingRoomLight.id && p.name === 'brightness' && p.value === 185
  );
  if (info.project.name === 'mobile') {
    const box = (await dialog.boundingBox())!;
    expect(box.y).toBe(0);
    expect(box.height).toBe(page.viewportSize()!.height);
    expect((await slider.boundingBox())!.height).toBeGreaterThan(250);
  }
  await page.keyboard.press('Escape');
  await expect(dialog).toHaveCount(0);
  await expect(opener).toBeFocused();
  expect(errors).toEqual([]);
});

test('history failures are explicit and refresh retries a read', async ({ page }) => {
  let fail = true;
  await page.route('**/api/metrics/query', async (route) => {
    if (fail) await route.fulfill({ status: 500, json: { error: 'read failed' } });
    else await route.continue();
  });
  await signIn(page, ui.home.path);
  await ui.home.tile(page, 'living room group', 'Brightness').locator('.nh-tile-body').click();
  const dialog = page.getByRole('dialog');
  await expect(dialog.getByRole('alert')).toContainText('Could not load history');
  fail = false;
  await dialog.getByRole('button', { name: 'Refresh history' }).click();
  await expect(dialog.locator('.nh-history-canvas[data-ready="true"]')).toBeVisible();
});

test('closing 50 completed views releases chart instances, keyboard listeners and body scroll', async ({ page }) => {
  test.setTimeout(90000);
  await page.addInitScript(() => {
    const add = document.addEventListener.bind(document);
    const remove = document.removeEventListener.bind(document);
    const listeners = new Set<EventListenerOrEventListenerObject>();
    (window as unknown as { modalListenerCount: () => number }).modalListenerCount = () => listeners.size;
    document.addEventListener = ((
      type: string,
      listener: EventListenerOrEventListenerObject,
      options?: boolean | AddEventListenerOptions
    ) => {
      if (type === 'keydown' && options === true) listeners.add(listener);
      add(type, listener, options);
    }) as typeof document.addEventListener;
    document.removeEventListener = ((
      type: string,
      listener: EventListenerOrEventListenerObject,
      options?: boolean | EventListenerOptions
    ) => {
      if (type === 'keydown' && options === true) listeners.delete(listener);
      remove(type, listener, options);
    }) as typeof document.removeEventListener;
  });
  await signIn(page, ui.home.path);
  const opener = ui.home.tile(page, 'living room group', 'Brightness').locator('.nh-tile-body');
  const pending = new Set<string>();
  page.on('request', (request) => {
    if (request.url().includes('/metrics/query')) pending.add(request.url());
  });
  page.on('requestfinished', (request) => pending.delete(request.url()));
  page.on('requestfailed', (request) => pending.delete(request.url()));
  const cdp = await page.context().newCDPSession(page);
  const heaps: { cycles: number; usedBytes: number; nodes: number; listeners: number }[] = [];
  for (let i = 0; i < 50; i++) {
    await opener.click();
    const dialog = page.getByRole('dialog');
    await expect(dialog.locator('.nh-history-canvas[data-ready="true"]')).toBeVisible();
    await dialog.getByRole('button', { name: 'Close', exact: true }).click();
    await expect(dialog).toHaveCount(0);
    if ([4, 24, 49].includes(i)) {
      await cdp.send('HeapProfiler.collectGarbage');
      const heap = await cdp.send('Runtime.getHeapUsage');
      const dom = await cdp.send('Memory.getDOMCounters');
      heaps.push({ cycles: i + 1, usedBytes: heap.usedSize, nodes: dom.nodes, listeners: dom.jsEventListeners });
    }
  }
  await cdp.detach();
  await expect.poll(() => pending.size).toBe(0);
  await test.info().attach('retained-heap-trend', { body: JSON.stringify(heaps), contentType: 'application/json' });
  console.log(`Entity retained heap (${test.info().project.name}): ${JSON.stringify(heaps)}`);
  expect(await page.evaluate(() => document.body.style.overflow)).toBe('');
  expect(
    await page.evaluate(() => (window as unknown as { modalListenerCount: () => number }).modalListenerCount())
  ).toBe(0);
  expect(await page.locator('canvas.nh-history-canvas').count()).toBe(0);
  const warm = heaps[1],
    final = heaps[2];
  expect(final.nodes).toBeLessThanOrEqual(warm.nodes + 100);
  expect(final.listeners).toBeLessThanOrEqual(warm.listeners + 5);
  expect(final.usedBytes - warm.usedBytes).toBeLessThan(2 * 1024 * 1024);
});

test('vertical adjustment commits once after a mouse or touch gesture and cancellation sends nothing', async ({
  page,
  ws,
}) => {
  await signIn(page, ui.home.path);
  await ui.home.tile(page, 'living room group', 'Brightness').locator('.nh-tile-body').click();
  const slider = page.getByRole('dialog').getByRole('slider', { name: 'Brightness', exact: true });
  const box = (await slider.boundingBox())!;
  const before = ws.sentOf('deviceSetValue').length;
  await slider.dispatchEvent('pointerdown');
  await slider.evaluate((element: HTMLInputElement) => {
    element.value = '120';
    element.dispatchEvent(new Event('input', { bubbles: true }));
  });
  await slider.dispatchEvent('pointercancel');
  await slider.dispatchEvent('change');
  expect(ws.sentOf('deviceSetValue')).toHaveLength(before);
  await dragBetween(
    page,
    { x: box.x + box.width / 2, y: box.y + box.height * 0.7 },
    { x: box.x + box.width / 2, y: box.y + box.height * 0.2 }
  );
  await ws.waitForSent('deviceSetValue', (p) => p.name === 'brightness');
  expect(ws.sentOf('deviceSetValue').slice(before)).toHaveLength(1);
});

test('colour temperature mode displays Kelvin but sends native mired values', async ({ page, ws }) => {
  await signIn(page, ui.home.path);
  await ui.home.tile(page, 'living room group', 'Brightness').locator('.nh-tile-body').click();
  const dialog = page.getByRole('dialog');
  await dialog.getByRole('button', { name: 'Colour temperature', exact: true }).click();
  await expect(dialog.getByLabel('Colour temperature preview')).toHaveText('2703 K');
  const slider = dialog.getByRole('slider', { name: 'Colour temperature', exact: true });
  await slider.press('ArrowUp');
  await ws.waitForSent('deviceSetValue', (p) => p.name === 'color_temp' && p.value === 369);
});
