import { test, expect, signIn, dragBetween, HUB_API } from './support/fixtures';
import { MOCK } from './support/mock';
import { ui } from './support/ui';
const mockSocketURL = HUB_API.replace(/^http/, 'ws').replace(/\/api$/, '/ws');

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
    const socket = body.devices.find((d: { friendly_name: string }) => d.friendly_name === 'kitchen power socket');
    socket.availability = 'online';
    socket.exposes.energy.data = 286.62;
    socket.exposes.state.data = 'ON';
    await route.fulfill({ response, json: body });
  });
});

test('full history keeps the sheet expose ahead of saved preferences', async ({ page }) => {
  await signIn(page, ui.home.path);
  await page.evaluate((id) => localStorage.setItem(`nh-history-pref:${id}`, JSON.stringify({ version: 1, preset: '24', expose: 'state' })), MOCK.livingRoomLight.id);
  await ui.home.tile(page, 'living room group', 'Brightness').locator('.nh-tile-body').click();
  await page.getByRole('dialog').getByRole('link', { name: 'View full history' }).click();
  await expect.poll(() => ui.history.selected(page, 'Entity')).toBe('Brightness');
  await expect(page.getByRole('dialog')).toHaveCount(0);
});

test('phone sheet drag dismisses only from the header with motion enabled', async ({ page }) => {
  await page.setViewportSize({ width: 390, height: 844 });
  await page.emulateMedia({ reducedMotion: 'no-preference' });
  await signIn(page, ui.home.path);
  await ui.home.tile(page, 'living room group', 'Brightness').locator('.nh-tile-body').click();
  const dialog = page.getByRole('dialog');
  const header = (await dialog.locator('.nh-dialog-head').boundingBox())!;
  await page.mouse.move(header.x + 30, header.y + 25);
  await page.mouse.down();
  await page.mouse.move(header.x + 30, header.y + 180, { steps: 8 });
  await page.mouse.up();
  await expect(dialog).toHaveCount(0);
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
  await expect(dialog).toBeFocused();
  await expect(dialog.getByLabel('Brightness preview')).toHaveText('72 %');
  await expect(dialog.locator('.nh-history-canvas[data-ready="true"]')).toBeVisible();
  await expect(dialog.locator('.entity-power-state')).toHaveText(/^(On|Off|Unknown)$/);
  await dialog.screenshot({ path: info.outputPath('brightness.png') });
  const gripDelta = await page.evaluate(() => {
    const track = document.querySelector('.nh-range')!.getBoundingClientRect();
    const grip = document.querySelector('.nh-range-grip')!.getBoundingClientRect();
    const percent = parseFloat(
      getComputedStyle(document.querySelector('.nh-range')!).getPropertyValue('--range-percent')
    );
    return Math.abs(grip.top + grip.height / 2 - (track.bottom - (percent / 100) * track.height));
  });
  expect(gripDelta).toBeLessThan(2);
  if (info.project.name === 'desktop') {
    await expect(dialog.locator('.history-summary')).toContainText('Range');
  } else {
    await expect(dialog.locator('.history-summary')).toHaveCount(0);
  }
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

test('device-name shortcut opens the device page and releases the entity dialog', async ({ page, ws }) => {
  await signIn(page, ui.home.path);
  await ui.home.tile(page, 'living room group', 'Brightness').locator('.nh-tile-body').click();
  const dialog = page.getByRole('dialog', { name: 'Brightness', exact: true });
  const link = dialog.getByRole('link', { name: `Open ${MOCK.livingRoomLight.name} device page`, exact: true });
  await expect(link).toHaveAttribute('href', `/devicepage/${MOCK.livingRoomLight.id}`);
  const [otherTab] = await Promise.all([
    page.context().waitForEvent('page'),
    link.click({ modifiers: ['ControlOrMeta'] }),
  ]);
  await otherTab.waitForURL(new RegExp(`/devicepage/${MOCK.livingRoomLight.id}$`));
  await otherTab.close();
  await expect(dialog).toBeVisible();
  await link.focus();
  await page.keyboard.press('Enter');
  await expect(page).toHaveURL(new RegExp(`/devicepage/${MOCK.livingRoomLight.id}$`));
  await expect(dialog).toHaveCount(0);
  expect(await page.evaluate(() => document.body.style.overflow)).toBe('');
  await page.goBack();
  await expect(page).toHaveURL(new RegExp(ui.home.path));
  await expect(page.getByRole('dialog')).toHaveCount(0);
  const smoke = ui.home.tile(page, 'attic room', 'Smoke');
  const smokeId = await smoke.getAttribute('data-device');
  await smoke.locator('.nh-tile-body').click();
  await page.getByRole('dialog').getByRole('link', { name: /device page$/ }).click();
  await expect(page).toHaveURL(new RegExp(`/devicepage/${smokeId}$`));
  await expect(page.getByRole('dialog')).toHaveCount(0);
  expect(ws.sentOf('deviceSetValue')).toHaveLength(0);
});

test('unavailable entity shows no device shortcut', async ({ page }) => {
  await page.route('**/api/hubstate', async (route) => {
    const response = await route.fetch();
    const body = await response.json();
    const light = body.devices.find((d: { id: string }) => d.id === MOCK.livingRoomLight.id);
    delete light.exposes.brightness;
    await route.fulfill({ response, json: body });
  });
  await signIn(page, ui.home.path);
  await ui.home.tile(page, 'living room group', 'Brightness').locator('.nh-tile-body').click();
  const dialog = page.getByRole('dialog');
  await expect(dialog.getByText('This entity is no longer available.')).toBeVisible();
  await expect(dialog.getByRole('link', { name: /device page$/ })).toHaveCount(0);
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
  const sampleHeap = async (cycles: number) => {
    await cdp.send('HeapProfiler.collectGarbage');
    const heap = await cdp.send('Runtime.getHeapUsage');
    const dom = await cdp.send('Memory.getDOMCounters');
    return { cycles, usedBytes: heap.usedSize, nodes: dom.nodes, listeners: dom.jsEventListeners };
  };
  /**
   * A closed dialog can stay reachable for a frame or two, and the dashboard behind it
   * changes as mock reports arrive, so a single reading is noisy. Sample repeatedly and
   * keep the smallest reading: pending release only inflates the counters, while a real
   * leak stays reachable in every sample and is therefore still measured.
   */
  const settledHeap = async (cycles: number) => {
    let best = await sampleHeap(cycles);
    for (let attempt = 0; attempt < 6; attempt++) {
      await page.evaluate(() => new Promise<void>((resolve) => requestAnimationFrame(() => requestAnimationFrame(() => resolve()))));
      const next = await sampleHeap(cycles);
      if (next.nodes < best.nodes || (next.nodes === best.nodes && next.listeners < best.listeners)) best = next;
    }
    return best;
  };
  for (let i = 0; i < 50; i++) {
    await opener.click();
    const dialog = page.getByRole('dialog');
    await expect(dialog.locator('.nh-history-canvas[data-ready="true"]')).toBeVisible();
    await dialog.getByRole('button', { name: 'Close', exact: true }).click();
    await expect(dialog).toHaveCount(0);
    if ([4, 24].includes(i)) heaps.push(await settledHeap(i + 1));
  }
  // Measure the final state only after every history read has finished, so a late
  // response cannot be mistaken for retained memory.
  await expect.poll(() => pending.size).toBe(0);
  heaps.push(await settledHeap(50));
  await cdp.detach();
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

test('colour temperature mode displays Kelvin, warmth, endpoints and sends mired values', async ({ page, ws }) => {
  await signIn(page, ui.home.path);
  await ui.home.tile(page, 'living room group', 'Brightness').locator('.nh-tile-body').click();
  const dialog = page.getByRole('dialog');
  await dialog.getByRole('button', { name: 'Colour temperature', exact: true }).click();
  await expect(dialog.getByLabel('Colour temperature preview')).toHaveText('2703 K');
  await expect(dialog.locator('.entity-adjustment-detail')).toHaveText('Warm white');
  await expect(dialog.locator('.range-endpoints')).toContainText('Cool ·');
  await expect(dialog.locator('.range-endpoints')).toContainText('Warm ·');
  const slider = dialog.getByRole('slider', { name: 'Colour temperature', exact: true });
  await slider.press('ArrowUp');
  await ws.waitForSent('deviceSetValue', (p) => p.name === 'color_temp' && p.value === 369);
});

test('a missing adjustable reading is explicit, never a minimum fill', async ({ page }) => {
  await signIn(page, ui.home.path);
  await ui.home.tile(page, 'attic room', 'Brightness').locator('.nh-tile-body').click();
  const dialog = page.getByRole('dialog', { name: 'Brightness', exact: true });
  await expect(dialog.getByLabel('Brightness preview')).toHaveText('Unknown');
  await expect(dialog.getByRole('slider', { name: 'Brightness', exact: true })).toBeDisabled();
  await expect(dialog.locator('.nh-range')).toHaveClass(/is-unknown/);
  await expect(dialog.locator('.nh-range-grip')).toHaveCount(0);
});

test('read-only entity without history keeps its stable shell and explains the empty chart', async ({ page }, info) => {
  await page.route('**/api/metrics/query', async (route) => {
    const response = await route.fetch();
    const body = await response.json();
    for (const group of body) {
      if (group.expose !== 'illuminance') continue;
      group.values[0].value.data = [];
      group.values[0].value.partial = false;
    }
    await route.fulfill({ response, json: body });
  });
  await signIn(page, ui.home.path);
  await ui.home.tile(page, 'living room group', 'Illuminance').locator('.nh-tile-body').click();
  const dialog = page.getByRole('dialog', { name: 'Illuminance', exact: true });
  await expect(dialog).not.toHaveClass(/is-wide/);
  if (info.project.name === 'mobile') {
    expect((await dialog.boundingBox())!.y).toBe(0);
    expect((await dialog.boundingBox())!.height).toBe(page.viewportSize()!.height);
  }
  await expect(dialog.getByText('No recorded history for this period.')).toBeVisible();
  await expect(
    dialog.getByText('Charts appear when this device reports and metrics are recorded.')
  ).toBeVisible();
});

test('an alert view offers the labelled silence action', async ({ page }) => {
  await signIn(page, ui.home.path);
  await ui.home.tile(page, 'attic room', 'Smoke').locator('.nh-tile-body').click();
  const dialog = page.getByRole('dialog', { name: 'Smoke', exact: true });
  const action = dialog.locator('.entity-writable', { hasText: 'Silence' });
  await expect(action).toBeVisible();
  await expect(action.getByRole('switch', { name: 'Silence' })).toBeVisible();
});

test('missing current reading falls back to the latest recorded sample', async ({ page }) => {
  await signIn(page, ui.home.path);
  await ui.home.tile(page, 'attic room', 'Battery').locator('.nh-tile-body').click();
  const dialog = page.getByRole('dialog', { name: 'Battery', exact: true });
  await expect(dialog.getByText(/Latest recorded/)).toBeVisible();
});

test('a read-only entity exposes its writable control with a visible label', async ({ page }) => {
  await signIn(page, ui.home.path);
  await ui.home.tile(page, 'living room group', 'Alarm').locator('.nh-tile-body').click();
  const dialog = page.getByRole('dialog', { name: 'Alarm', exact: true });
  const writable = dialog.locator('.entity-writable');
  await expect(writable).toContainText('Alarm');
  await expect(writable.getByRole('switch', { name: 'Alarm' })).toBeVisible();
});

test('counter history without a complete interval explains the gap', async ({ page }) => {
  await page.route('**/api/metrics/query', async (route) => {
    const response = await route.fetch();
    const body = await response.json();
    for (const group of body) {
      if (group.expose !== 'energy') continue;
      const value = group.values[0].value;
      value.data = [{ timestamp: value.to - 1000, value: 286.62 }];
      value.partial = false;
    }
    await route.fulfill({ response, json: body });
  });
  await signIn(page, ui.home.path);
  await ui.home.tile(page, 'Kitchen', 'Energy').locator('.nh-tile-body').click();
  const dialog = page.getByRole('dialog', { name: 'Energy', exact: true });
  await expect(dialog.getByText('No complete counter interval in this period.')).toBeVisible();
  await expect(dialog.getByText('Not enough data')).toBeVisible();
  await expect(dialog.getByText('0 / 12')).toBeVisible();
  await expect(dialog.locator('canvas.nh-history-canvas')).toHaveCount(0);
});

test('primary controls stay reachable at 320x568', async ({ page }) => {
  await page.setViewportSize({ width: 320, height: 568 });
  await signIn(page, ui.home.path);
  await ui.home.tile(page, 'living room group', 'Brightness').locator('.nh-tile-body').click();
  const dialog = page.getByRole('dialog');
  const box = await dialog.locator('.entity-controls').boundingBox();
  expect(box).not.toBeNull();
  expect(box!.y + box!.height).toBeLessThanOrEqual(569);
  expect(await page.evaluate(() => document.documentElement.scrollWidth)).toBeLessThanOrEqual(320);
});

// Hold history until after the first geometry assertion: settled screenshots missed
// the original sheet-to-screen flash. All families use one stable phone shell.
for (const label of ['Temperature', 'Humidity', 'Presence', 'Illuminance', 'Energy', 'Contact', 'Smoke', 'Battery', 'Alarm']) {
  test(`${label} opens full-screen before empty history resolves`, async ({ page }, info) => {
    test.skip(info.project.name !== 'mobile', 'Phone-shell contract');
    let release!: () => void;
    const gate = new Promise<void>((resolve) => { release = resolve; });
    await page.route('**/api/metrics/query', async (route) => {
      await gate;
      const response = await route.fetch();
      const body = await response.json();
      for (const group of body) group.values[0].value.data = [];
      await route.fulfill({ response, json: body });
    });
    await signIn(page, ui.home.path);
    const tile = page.locator('.nh-tile').filter({
      has: page.locator('.nh-tile-label', { hasText: new RegExp(`^${label}$`, 'i') }),
    }).first();
    await tile.locator('.nh-tile-body').click();
    const dialog = page.getByRole('dialog');
    await expect(dialog.getByText('Loading history…')).toBeVisible();
    const initial = (await dialog.boundingBox())!;
    release();
    await expect(dialog.getByText('No recorded history for this period.')).toBeVisible();
    expect(initial.y).toBe(0);
    expect(initial.height).toBe(page.viewportSize()!.height);
    expect(await dialog.boundingBox()).toEqual(initial);
  });
}

test('energy meter reading stays visible without inventing a consumption percentage', async ({ page }) => {
  await signIn(page, ui.home.path);
  await ui.home.tile(page, 'Kitchen', 'Energy').locator('.nh-tile-body').click();
  const meter = page.getByRole('group', { name: 'Energy meter reading' });
  await expect(meter).toContainText('286.6 kWh');
  await expect(meter).toContainText('Meter total');
  await expect(meter.locator('svg')).toBeVisible();
});

test('continuous drag sends once and preserves the slider through confirmation', async ({ page, ws }, info) => {
  await page.emulateMedia({ reducedMotion: 'no-preference' });
  await signIn(page, ui.home.path);
  await ui.home.tile(page, 'living room group', 'Brightness').locator('.nh-tile-body').click();
  const dialog = page.getByRole('dialog');
  const input = dialog.getByRole('slider', { name: 'Brightness', exact: true });
  const box = (await input.boundingBox())!;
  const before = ws.sentOf('deviceSetValue').length;
  await input.evaluate((element: HTMLInputElement) => {
    const samples: { percent: number; error: number }[] = [];
    (element as HTMLInputElement & { dragSamples: typeof samples }).dragSamples = samples;
    const measure = () => requestAnimationFrame(() => {
      const rail = element.closest('.nh-range')!;
      const track = rail.getBoundingClientRect();
      const grip = rail.querySelector('.nh-range-grip')!.getBoundingClientRect();
      const percent = parseFloat(getComputedStyle(rail).getPropertyValue('--range-percent'));
      samples.push({ percent, error: Math.abs(grip.top + grip.height / 2 - (track.bottom - track.height * percent / 100)) });
    });
    element.addEventListener('pointermove', measure);
    element.addEventListener('pointerup', () => element.removeEventListener('pointermove', measure), { once: true });
  });
  await dragBetween(page,
    { x: box.x + box.width / 2, y: box.y + box.height * 0.75 },
    { x: box.x + box.width / 2, y: box.y + box.height * 0.33 }, 25);
  await ws.waitForSent('deviceSetValue', (p) => p.name === 'brightness');
  expect(ws.sentOf('deviceSetValue').slice(before)).toHaveLength(1);
  await expect(dialog.locator('.nh-range')).not.toHaveClass(/is-unknown/);
  const samples = await input.evaluate((element) =>
    (element as HTMLInputElement & { dragSamples: { percent: number; error: number }[] }).dragSamples);
  expect(new Set(samples.map((sample) => sample.percent)).size).toBeGreaterThan(12);
  expect(Math.max(...samples.map((sample) => sample.error))).toBeLessThan(2);
  await dialog.screenshot({ path: info.outputPath('repaired-brightness.png') });
});

for (const viewport of [{ width: 768, height: 1024 }, { width: 844, height: 390 }]) {
  test(`entity shell stays full-screen at ${viewport.width}x${viewport.height}`, async ({ page }) => {
    await page.setViewportSize(viewport);
    await signIn(page, ui.home.path);
    await ui.home.tile(page, 'living room group', 'Brightness').locator('.nh-tile-body').click();
    const dialog = page.getByRole('dialog');
    await expect(dialog).toBeVisible();
    expect(await dialog.boundingBox()).toEqual({ x: 0, y: 0, ...viewport });
    await dialog.getByRole('button', { name: 'Close', exact: true }).click();
    await expect(dialog).toHaveCount(0);
  });
}

for (const reading of [{ value: null, text: 'Unknown' }, { value: 0, text: '0 kWh' }]) {
  test(`energy meter distinguishes ${reading.text} from missing consumption history`, async ({ page }) => {
    await page.route('**/api/hubstate', async (route) => {
      const response = await route.fetch();
      const body = await response.json();
      const socket = body.devices.find((device: { friendly_name: string }) => device.friendly_name === 'kitchen power socket');
      socket.availability = 'online';
      socket.exposes.energy.data = reading.value;
      await route.fulfill({ response, json: body });
    });
    await page.route('**/api/metrics/query', (route) => route.fulfill({ json: [] }));
    await signIn(page, ui.home.path);
    await ui.home.tile(page, 'Kitchen', 'Energy').locator('.nh-tile-body').click();
    await expect(page.getByRole('group', { name: 'Energy meter reading' })).toContainText(reading.text);
  });
}


for (const mode of ['dark', 'light']) {
  test(`meter retains three distinct colours in ${mode} mode`, async ({ page }, info) => {
    await signIn(page, ui.home.path);
    if (mode === 'light') await page.getByRole('button', { name: 'Switch to light mode' }).click();
    await expect(page.locator('html')).toHaveAttribute('data-mode', mode);
    await ui.home.tile(page, 'Kitchen', 'Energy').locator('.nh-tile-body').click();
    const meter = page.getByRole('group', { name: 'Energy meter reading' });
    await expect(meter).toBeVisible();
    const colours = await meter.locator('stop').evaluateAll((stops) => stops.map((stop) => getComputedStyle(stop).stopColor));
    expect(new Set(colours).size).toBe(3);
    expect(await meter.locator('svg > path').last().evaluate((path) => getComputedStyle(path).stroke)).toContain('url(');
    await page.getByRole('dialog').screenshot({ path: info.outputPath(`meter-${mode}.png`), scale: 'css' });
  });
}

for (const level of [82, 0, null]) {
  test(`battery diagnostic keeps ${level} distinct from an invented reading`, async ({ page }, info) => {
    await page.route('**/api/hubstate', async (route) => {
      const response = await route.fetch();
      const body = await response.json();
      const device = body.devices.find((device: { exposes: Record<string, unknown> }) => device.exposes.battery);
      device.availability = 'online';
      Object.assign(device.exposes.battery, { data: level, unit: '%' });
      device.exposes.linkquality = { ...device.exposes.battery, name: 'linkquality', data: 116, unit: 'lqi' };
      await route.fulfill({ response, json: body });
    });
    await signIn(page, ui.home.path);
    await page.locator('.nh-tile').filter({ has: page.locator('.nh-tile-label', { hasText: /^Battery$/ }) }).first().locator('.nh-tile-body').click();
    const dialog = page.getByRole('dialog');
    const bar = dialog.getByRole('meter', { name: 'Battery level' });
    if (level === null) await expect(bar).toHaveCount(0);
    else await expect(bar).toHaveAttribute('aria-valuenow', String(level));
    await expect(dialog.getByRole('region', { name: 'Other readings' })).toContainText('116');
    await dialog.screenshot({ path: info.outputPath(`battery-${level}.png`), scale: 'css' });
  });
}

test('entity mode choices show direct metadata options and wait for confirmation', async ({ page }) => {
  await page.route('**/api/hubstate', async (route) => {
    const response = await route.fetch();
    const body = await response.json();
    const device = body.devices.find((device: { id: string }) => device.id === MOCK.livingRoomLight.id);
    device.availability = 'online';
    Object.assign(device.exposes.brightness, { type: 'enum', data: 'near', values: { near: 'near', medium: 'medium', far: 'far' } });
    await route.fulfill({ response, json: body });
  });
  const commands: { name: string; value: unknown }[] = [];
  let confirm: (() => void) | undefined;
  await page.routeWebSocket(mockSocketURL, (socket) => {
    socket.onMessage((raw) => {
      const frame = JSON.parse(String(raw));
      if (frame.type !== 'deviceSetValue') return;
      commands.push(frame.payload);
      confirm = () => socket.send(JSON.stringify({
        type: 'deviceUpdated',
        payload: { id: frame.payload.id, data: { [frame.payload.name]: frame.payload.value } },
      }));
    });
  });
  // Hold confirmation in the mock socket; clicking must never replace confirmed state.
  await signIn(page, ui.home.path);
  await ui.home.tile(page, 'living room group', 'Brightness').locator('.nh-tile-body').click();
  const choices = page.getByRole('dialog').getByRole('group', { name: 'Brightness', exact: true });
  await expect(choices.getByRole('button', { name: 'Near', exact: true })).toHaveAttribute('aria-pressed', 'true');
  await choices.getByRole('button', { name: 'Medium', exact: true }).click();
  await expect.poll(() => commands.length).toBe(1);
  expect(commands[0]).toMatchObject({ name: 'brightness', value: 'medium' });
  await expect(choices.getByRole('button', { name: 'Medium', exact: true })).toHaveAttribute('aria-pressed', 'false');
  await expect(choices.getByRole('button', { name: 'Near', exact: true })).toHaveAttribute('aria-pressed', 'true');
  await expect(choices.getByRole('button', { name: 'Medium', exact: true })).toBeDisabled();
  confirm!();
  await expect(choices.getByRole('button', { name: 'Medium', exact: true })).toHaveAttribute('aria-pressed', 'true');
  await expect(choices.getByRole('button', { name: 'Medium', exact: true })).toBeEnabled();
});


for (const control of [
  { name: 'brightness', label: 'Brightness', during: 20, pending: 30 },
  { name: 'color_temp', label: 'Colour temperature', during: 490, pending: 480 },
]) {
  test(`${control.name} ignores live reports during native drag and pending confirmation`, async ({ page }) => {
    const commands: { id: string; name: string; value: unknown }[] = [];
    let report!: (value: unknown, name?: string) => void;
    await page.routeWebSocket(mockSocketURL, (socket) => {
      report = (value, name = control.name) => socket.send(JSON.stringify({
        type: 'deviceUpdated', payload: { id: MOCK.livingRoomLight.id, data: { [name]: value } },
      }));
      socket.onMessage((raw) => {
        const frame = JSON.parse(String(raw));
        if (frame.type === 'deviceSetValue') commands.push(frame.payload);
      });
    });
    await signIn(page, ui.home.path);
    await ui.home.tile(page, 'living room group', 'Brightness').locator('.nh-tile-body').click();
    const dialog = page.getByRole('dialog');
    if (control.name === 'color_temp') await dialog.getByRole('button', { name: 'Colour temperature', exact: true }).click();
    const slider = dialog.getByRole('slider', { name: control.label, exact: true });
    const preview = dialog.getByLabel(`${control.label} preview`);
    const box = (await slider.boundingBox())!;
    let heldPreview = '';
    // Routed WebSocket frames are not visible to Playwright's network recorder.
    // Let delivery, Vue's update and a full rendered frame settle before checking.
    const paint = () => page.evaluate(() => new Promise<void>((resolve) => requestAnimationFrame(() => requestAnimationFrame(() => resolve()))));
    await dragBetween(page,
      { x: box.x + box.width / 2, y: box.y + box.height * 0.7 },
      { x: box.x + box.width / 2, y: box.y + box.height * 0.2 }, 20,
      async () => {
        expect(commands).toHaveLength(0);
        heldPreview = (await preview.textContent())!;
        report(control.during);
        await paint();
        await expect(preview).toHaveText(heldPreview);
        expect(commands).toHaveLength(0);
      });
    await expect.poll(() => commands.length).toBe(1);
    await expect(slider).toBeDisabled();
    await expect(dialog.getByText(/Sending|Waiting for device confirmation/)).toHaveCount(0);
    await expect(dialog.locator('.nh-spin')).toHaveCount(0);
    report(control.pending);
    await paint();
    await expect(preview).toHaveText(heldPreview);
    expect(commands).toHaveLength(1);
    report(commands[0].value);
    await expect(slider).toBeEnabled();
    await expect(preview).toHaveText(heldPreview);
    const power = dialog.getByRole('switch', { name: /^Power,/ });
    const state = dialog.locator('.entity-power-state');
    const confirmed = await state.innerText();
    await power.click();
    await expect(power).toHaveAttribute('aria-busy', 'true');
    await expect(power).toBeDisabled();
    await expect(power).toHaveCSS('opacity', '1');
    await expect(state).toHaveText(confirmed);
    await expect(dialog.getByText(/Sending|Waiting for device confirmation/)).toHaveCount(0);
    await expect.poll(() => commands.length).toBe(2);
    expect(commands[1].name).toBe('state');
    report(commands[1].value, 'state');
    await expect(power).toBeEnabled();
    await expect(state).not.toHaveText(confirmed);
  });
}
