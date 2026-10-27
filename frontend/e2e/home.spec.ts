import {
  test,
  expect,
  signIn,
  signInAutomationHub,
  AUTOMATION_HUB_API,
  HUB_API,
  dragBetween,
} from './support/fixtures';
import { ui } from './support/ui';
import { MOCK } from './support/mock';
import type { Page } from '@playwright/test';
import { readFile } from 'node:fs/promises';
import type { ConnectionInstance } from '../src/types/integrations.type';

/** Seed only the isolated Go fixture; production persistence and source admission remain real. */
async function mixedHomeFixture(page: Page) {
  await signInAutomationHub(page, '/integrations');
  expect((await page.request.post(`${AUTOMATION_HUB_API}/__test/connections/reset`)).status()).toBe(204);
  const current = (await (await page.request.get(`${AUTOMATION_HUB_API}/home/layout`)).json()).data;
  const saved = await page.request.put(`${AUTOMATION_HUB_API}/home/layout`, {
    headers: { 'Origin': 'http://localhost:4100', 'X-NodeHerder-Request': '1', 'If-Match': current.etag },
    data: {
      ...current,
      blocks: ['Living room', 'Garden'].map((name, index) => ({
        id: `browser-area-${index}`,
        kind: 'group',
        name,
        cards: [],
      })),
    },
  });
  expect(saved.status()).toBe(200);
  await page.reload();
  await page.getByRole('button', { name: 'Add integration', exact: true }).click();
  await page.getByRole('button', { name: 'Set up', exact: true }).click();
  await page.getByRole('textbox', { name: 'Postcode', exact: true }).fill('XX1 1XX');
  await page.getByRole('button', { name: 'Find address' }).click();
  await page
    .getByRole('combobox', { name: 'Choose your address' })
    .selectOption({ label: '1 Synthetic Street, Test Town' });
  await page.getByRole('button', { name: 'Save integration' }).click();
  await expect(page.getByRole('link', { name: 'Open Bin collections integration', exact: true })).toBeVisible();
  await page.goto('/groupdashboard');
  await expect(page.getByRole('button', { name: 'Edit layout' })).toBeEnabled();
}

async function selectHomeSource(page: Page, label: string) {
  await page.getByRole('combobox', { name: 'Device or integration' }).selectOption({ label });
}

/** Supply three local services to exercise summary readability without provider traffic. */
async function showThreeServiceSchedule(page: Page) {
  await page.route(`${HUB_API}/integrations/instances`, async (route) => {
    const response = await route.fetch({ url: route.request().url().replace(HUB_API, AUTOMATION_HUB_API) });
    const payload = (await response.json()) as { data: ConnectionInstance[]; error: null };
    for (const instance of payload.data) {
      const resource = instance.resources[0];
      const original = resource.services[0];
      resource.services = [
        {
          ...original,
          service_id: '00000000-0000-4000-8000-000000000001',
          name: 'Black bin',
          waste_type: 'refuse',
          bin_colour: 'black',
        },
        {
          ...original,
          service_id: '00000000-0000-4000-8000-000000000002',
          name: 'Brown bin',
          waste_type: 'garden',
          bin_colour: 'brown',
        },
        original,
      ];
    }
    await route.fulfill({ response, json: payload });
  });
}

/** Let responsive sidebar/theme transitions finish before measuring or capturing the view. */
async function settleHomeView(page: Page) {
  await page.evaluate(async () => {
    await new Promise<void>((resolve) => requestAnimationFrame(() => requestAnimationFrame(() => resolve())));
    const transitions = document
      .getAnimations()
      .filter((animation) => Number.isFinite(animation.effect?.getComputedTiming().endTime));
    await Promise.allSettled(transitions.map((animation) => animation.finished));
  });
}

/** Measure after observer/paint settlement at each required width in both themes. */
async function checkHomeGeometry(page: Page, panel = false) {
  const original = page.viewportSize()!;
  for (const theme of ['light', 'dark']) {
    const toggle = page.getByRole('button', { name: `Switch to ${theme} mode`, exact: true });
    if (await toggle.count()) await toggle.click();
    for (const width of [390, 1024, 1280, 1440, 1680]) {
      await page.setViewportSize({ width, height: 844 });
      await settleHomeView(page);
      await expect
        .poll(() =>
          page.evaluate(() => {
            const containers = [
              document.documentElement,
              // Remove badges intentionally extend into the page padding. Check scrolling
              // containers and actual card bounds without treating that decoration as clipping.
              ...Array.from(document.querySelectorAll<HTMLElement>('.nh-app-main')),
            ];
            const blocks = Array.from(document.querySelectorAll<HTMLElement>('.nh-home-block, .hp-buttons > article'));
            return {
              overflow: containers
                .filter((node) => node.scrollWidth > node.clientWidth + 1)
                .map((node) => ({
                  name: node.className || node.tagName,
                  width: node.clientWidth,
                  scroll: node.scrollWidth,
                })),
              clipped: blocks
                .filter((node) => {
                  const box = node.getBoundingClientRect();
                  return box.left < -1 || box.right > innerWidth + 1;
                })
                .map((node) => ({
                  name: node.className,
                  left: node.getBoundingClientRect().left,
                  right: node.getBoundingClientRect().right,
                  viewport: innerWidth,
                })),
            };
          })
        )
        .toEqual({ overflow: [], clipped: [] });
      if (panel) {
        const summary = page.locator('.nh-integration-tile.is-summary');
        const bounds = await summary.evaluate((element) => ({
          width: element.getBoundingClientRect().width,
          available: element.parentElement!.clientWidth,
          rem: parseFloat(getComputedStyle(document.documentElement).fontSize),
          rowFont: parseFloat(getComputedStyle(element.querySelector('.nh-integration-rows')!).fontSize),
        }));
        expect(bounds.width).toBeLessThanOrEqual(24 * bounds.rem + 1);
        expect(bounds.width).toBeGreaterThanOrEqual(Math.min(bounds.available, 24 * bounds.rem) - 1);
        expect(bounds.rowFont).toBeGreaterThanOrEqual(bounds.rem);
      }
      if (!panel)
        await expect
          .poll(() =>
            page.locator('.nh-home-block').evaluateAll((nodes) => {
              const boxes = nodes.map((node) => node.getBoundingClientRect());
              return boxes.some((a, index) =>
                boxes
                  .slice(index + 1)
                  .some(
                    (b) =>
                      Math.min(a.right, b.right) - Math.max(a.left, b.left) > 1 &&
                      Math.min(a.bottom, b.bottom) - Math.max(a.top, b.top) > 1
                  )
              );
            })
          )
          .toBe(false);
    }
  }
  await page.setViewportSize(original);
  await settleHomeView(page);
}

test('standalone cards align with the neighboring area tile rows', async ({ page }, info) => {
  await mixedHomeFixture(page);
  const instances = (await (await page.request.get(`${AUTOMATION_HUB_API}/integrations/instances`)).json()).data;
  const instance = instances[0];
  const resource = instance.resources[0];
  const current = (await (await page.request.get(`${AUTOMATION_HUB_API}/home/layout`)).json()).data;
  const deviceCard = (id: string, deviceId: string, expose: string) => ({
    id, kind: 'device', panel: true, device_id: deviceId, expose,
  });
  const light = '0x0000000000000002';
  const dial = '0x0000000000000001';
  const blocks = [
    { id: 'living', kind: 'group', name: 'Living room', cards: [
      deviceCard('state', light, 'state'), deviceCard('brightness', light, 'brightness'), deviceCard('action', dial, 'action'),
    ] },
    { id: 'garden', kind: 'group', name: 'Garden', cards: [deviceCard('garden-light', light, 'state')] },
    { id: 'kitchen', kind: 'group', name: 'Kitchen', cards: [deviceCard('kitchen-light', light, 'state')] },
    { id: 'bins-block', kind: 'card', name: '', cards: [{
      id: 'bins-card', kind: 'waste_summary', panel: true, instance_id: instance.id, resource_id: resource.resource_id,
    }] },
  ];
  const saved = await page.request.put(`${AUTOMATION_HUB_API}/home/layout`, {
    headers: { Origin: 'http://localhost:4100', 'X-NodeHerder-Request': '1', 'If-Match': current.etag },
    data: { ...current, blocks },
  });
  expect(saved.status()).toBe(200);
  await page.reload();
  await expect(page.locator('[data-card-id="bins-card"]')).toBeVisible();

  for (const density of ['comfortable', 'compact', 'spacious']) {
    await page.evaluate((value) => document.documentElement.setAttribute('data-density', value), density);
    for (const width of [1280, 1440]) {
      await page.setViewportSize({ width, height: 900 });
      await settleHomeView(page);
      const spacing = await page.evaluate(() => {
        const living = document.querySelectorAll<HTMLElement>('[data-area="Living room"] .nh-home-card');
        const garden = document.querySelector<HTMLElement>('[data-area="Garden"] .nh-home-card')!;
        const summary = document.querySelector<HTMLElement>('[data-card-id="bins-card"]')!;
        const [first, second, third] = Array.from(living, card => card.getBoundingClientRect());
        const adjacent = garden.getBoundingClientRect();
        const standalone = summary.getBoundingClientRect();
        return {
          tileGap: second.left - first.right,
          areaGap: adjacent.left - second.right,
          verticalGap: standalone.top - adjacent.bottom,
          rowDifference: standalone.top - third.top,
          columnDifference: standalone.left - adjacent.left,
        };
      });
      expect(Math.abs(spacing.areaGap - spacing.tileGap)).toBeLessThanOrEqual(1);
      expect(Math.abs(spacing.verticalGap - spacing.tileGap)).toBeLessThanOrEqual(2);
      expect(Math.abs(spacing.rowDifference)).toBeLessThanOrEqual(2);
      expect(Math.abs(spacing.columnDifference)).toBeLessThanOrEqual(1);
    }
  }
  await page.setViewportSize({ width: 1680, height: 900 });
  await settleHomeView(page);
  await expect.poll(() => page.evaluate(() => {
    const summary = document.querySelector<HTMLElement>('[data-card-id="bins-card"]')!.getBoundingClientRect();
    const tile = document.querySelector<HTMLElement>('[data-area="Living room"] .nh-home-card')!.getBoundingClientRect();
    return Math.abs(summary.top - tile.top);
  })).toBeLessThanOrEqual(2);
  await page.setViewportSize({ width: 1280, height: 900 });
  await settleHomeView(page);
  await expect.poll(() => page.evaluate(() => {
    const summary = document.querySelector<HTMLElement>('[data-card-id="bins-card"]')!.getBoundingClientRect();
    const tile = document.querySelectorAll<HTMLElement>('[data-area="Living room"] .nh-home-card')[2].getBoundingClientRect();
    return Math.abs(summary.top - tile.top);
  })).toBeLessThanOrEqual(2);
  await page.screenshot({ path: info.outputPath('aligned-home-cards.png'), fullPage: true });
  await page.evaluate(() => document.documentElement.setAttribute('data-density', 'compact'));
  await settleHomeView(page);
  await ui.home.editLayout(page);
  await expect(page.getByRole('button', { name: 'Rename Garden', exact: true })).toBeVisible();
  await page.screenshot({ path: info.outputPath('aligned-home-cards-edit.png'), fullPage: true });
  await ui.home.done(page);
  const layout = (await (await page.request.get(`${AUTOMATION_HUB_API}/home/layout`)).json()).data;
  const reordered = [layout.blocks.at(-1), ...layout.blocks.slice(0, -1)];
  const moved = await page.request.put(`${AUTOMATION_HUB_API}/home/layout`, {
    headers: { Origin: 'http://localhost:4100', 'X-NodeHerder-Request': '1', 'If-Match': layout.etag },
    data: { ...layout, blocks: reordered },
  });
  expect(moved.status()).toBe(200);
  await page.reload();
  for (const density of ['comfortable', 'compact', 'spacious']) {
    await page.evaluate((value) => document.documentElement.setAttribute('data-density', value), density);
    for (const width of [1280, 1440, 1680]) {
      await page.setViewportSize({ width, height: 900 });
      await settleHomeView(page);
      await expect.poll(() => page.evaluate(() => {
        const summary = document.querySelector<HTMLElement>('[data-card-id="bins-card"]')!.getBoundingClientRect();
        const tile = document.querySelector<HTMLElement>('[data-area="Living room"] .nh-home-card')!.getBoundingClientRect();
        return Math.abs(summary.top - tile.top);
      })).toBeLessThanOrEqual(2);
    }
  }
  await page.screenshot({ path: info.outputPath('aligned-home-cards-first-row.png'), fullPage: true });
});

test('individual integration cards move naturally between areas and Home with shared panel visibility', async ({
  page,
}, info) => {
  test.setTimeout(60_000);
  const errors: Error[] = [];
  page.on('pageerror', (error) => errors.push(error));
  await mixedHomeFixture(page);
  await expect(ui.home.groupTitles(page)).toHaveText(['Living room', 'Garden']);
  await ui.home.editLayout(page);
  await ui.home.area(page, 'Living room').getByRole('button', { name: 'Add tiles' }).click();
  await selectHomeSource(page, 'Test light');
  await page.getByRole('checkbox', { name: /^State/ }).check();
  await page.getByRole('button', { name: 'Add 1 tile', exact: true }).click();
  await expect(page.getByRole('dialog')).toHaveCount(0);
  await ui.home.area(page, 'Living room').getByRole('button', { name: 'Add tiles' }).click();
  await selectHomeSource(page, 'Integration results');
  await page.getByRole('checkbox', { name: /^Blue bin/ }).check();
  await page.getByRole('button', { name: 'Add 1 tile', exact: true }).click();
  const tile = page.locator('.nh-integration-tile').filter({ hasText: 'Blue bin' });
  await expect(tile).toHaveCount(1);
  const id = await tile.getAttribute('data-home-card');
  const placement = page.locator(`[data-card-id="${id}"]`);
  await placement.getByRole('button', { name: 'Move card', exact: true }).focus();
  await page.keyboard.press('Enter');
  await page.getByRole('combobox', { name: 'Place card' }).selectOption({ label: 'Garden' });
  await page.getByRole('button', { name: 'Save placement' }).click();
  await expect(ui.home.area(page, 'Garden').locator(`[data-card-id="${id}"]`)).toBeVisible();
  await placement.hover();
  await placement.getByRole('button', { name: 'Move card', exact: true }).click();
  await page.getByRole('combobox', { name: 'Place card' }).selectOption({ label: 'Directly on Home' });
  await page.getByRole('checkbox', { name: 'Show this card in panel mode' }).uncheck();
  await page.getByRole('button', { name: 'Save placement' }).click();
  await expect(page.locator('.nh-home-block:not(.is-area)').locator(`[data-card-id="${id}"]`)).toBeVisible();
  await page.getByRole('button', { name: 'Add cards to Home' }).click();
  await selectHomeSource(page, 'Integration results');
  await page.getByRole('checkbox', { name: /complete schedule/ }).check();
  await page.getByRole('button', { name: 'Add 1 tile', exact: true }).click();
  await expect(page.locator('.nh-home-block:not(.is-area)')).toHaveCount(2);
  const downloading = page.waitForEvent('download');
  await page.getByRole('button', { name: 'Export', exact: true }).click();
  const exported = JSON.parse(await readFile((await (await downloading).path())!, 'utf8'));
  expect(exported.format).toBe('nodeherder.home/v1');
  expect(exported.layout.blocks.flatMap((block: { cards: unknown[] }) => block.cards)).toHaveLength(3);
  expect(JSON.stringify(exported)).not.toContain('p:synthetic');
  expect(JSON.stringify(exported)).not.toContain('checked_at');
  await page.locator('.nh-home input[type=file]').setInputFiles({
    name: 'home-layout.json',
    mimeType: 'application/json',
    buffer: Buffer.from(JSON.stringify(exported)),
  });
  await page.getByRole('dialog').getByRole('button', { name: 'Import', exact: true }).click();
  await expect(page.getByRole('dialog')).toHaveCount(0);
  await expect(page.getByRole('button', { name: 'Done', exact: true })).toBeEnabled();
  await checkHomeGeometry(page);
  await page.screenshot({ path: info.outputPath('home-mixed-edit.png'), fullPage: true });
  await ui.home.done(page);
  await page.reload();
  await expect(page.locator(`[data-card-id="${id}"]`)).toBeVisible();
  await checkHomeGeometry(page);
  await page.screenshot({ path: info.outputPath('home-mixed.png'), fullPage: true });
  await showThreeServiceSchedule(page);
  await page.getByRole('main').getByRole('link', { name: 'Panel mode', exact: true }).click();
  await expect(ui.panel.roomTabs(page)).toHaveCount(2);
  await expect(ui.panel.roomTabs(page)).toHaveText(['Living room', 'Bin collections']);
  await ui.panel.roomTabs(page).nth(1).click();
  await expect(page.locator('.nh-integration-tile')).toHaveCount(1);
  await expect(page.locator('.nh-integration-rows li')).toHaveCount(3);
  await checkHomeGeometry(page, true);
  await page.screenshot({ path: info.outputPath('panel-integration.png'), fullPage: true });
  if (info.project.name === 'desktop') {
    const original = page.viewportSize()!;
    await page.setViewportSize({ width: 1680, height: 900 });
    await settleHomeView(page);
    await page.screenshot({ path: info.outputPath('panel-integration-wide.png'), fullPage: true });
    await page.setViewportSize(original);
    await settleHomeView(page);
  }
  await page.locator('.nh-integration-tile').getByRole('link', { name: 'Bin collections', exact: true }).click();
  await page.getByRole('link', { name: 'Setup', exact: true }).click();
  await page.getByRole('link', { name: 'Panel', exact: true }).click();
  await expect(page).toHaveURL(/\/panel$/);
  await expect(ui.panel.roomTabs(page).nth(1)).toHaveAttribute('aria-selected', 'true');
  await ui.panel.exit(page).click();
  await ui.home.editLayout(page);
  await page.getByRole('button', { name: 'Rename Garden', exact: true }).click();
  await page.getByRole('textbox', { name: 'Area name', exact: true }).fill('Outside');
  await page.getByRole('button', { name: 'Rename', exact: true }).click();
  await expect(ui.home.groupTitles(page)).toHaveText(['Living room', 'Outside']);
  await page.getByRole('button', { name: 'Delete Outside', exact: true }).click();
  await page.getByRole('button', { name: 'Delete area', exact: true }).click();
  await expect(ui.home.groupTitles(page)).toHaveText(['Living room']);
  expect(errors).toEqual([]);
});

test('a failed area reorder restores accepted order and exposes recovery without automatic retry', async ({ page }) => {
  await signIn(page, ui.home.path);
  await expect(page.getByRole('button', { name: 'Edit layout' })).toBeEnabled();
  const before = await ui.home.groupTitles(page).allTextContents();
  await ui.home.editLayout(page);
  let writes = 0;
  await page.route('**/api/home/layout', async (route) => {
    if (route.request().method() !== 'PUT') return route.continue();
    writes++;
    await route.fulfill({
      status: 409,
      contentType: 'application/json',
      body: JSON.stringify({
        data: null,
        error: {
          code: 'revision_conflict',
          message: 'Home changed in another session. Reload the layout and try again.',
        },
      }),
    });
  });
  await ui.home.moveLater(page, before[0]).click();
  await expect(page.getByRole('alert')).toContainText('Home changed in another session');
  await expect(ui.home.groupTitles(page)).toHaveText(before);
  expect(writes).toBe(1);
  await expect(page.getByRole('button', { name: 'Reload layout' })).toBeEnabled();
});

test('a standalone card has a visible handle that reorders and saves its position', async ({ page }) => {
  await page.setViewportSize({ width: 1440, height: 900 });
  await mixedHomeFixture(page);
  const instances = (await (await page.request.get(`${AUTOMATION_HUB_API}/integrations/instances`)).json()).data;
  const current = (await (await page.request.get(`${AUTOMATION_HUB_API}/home/layout`)).json()).data;
  const instance = instances[0];
  const saved = await page.request.put(`${AUTOMATION_HUB_API}/home/layout`, {
    headers: { Origin: 'http://localhost:4100', 'X-NodeHerder-Request': '1', 'If-Match': current.etag },
    data: { ...current, blocks: [...current.blocks, {
      id: 'standalone-bins', kind: 'card', name: '', cards: [{
        id: 'bins-summary', kind: 'waste_summary', panel: true,
        instance_id: instance.id, resource_id: instance.resources[0].resource_id,
      }],
    }] },
  });
  expect(saved.status()).toBe(200);
  await page.reload();
  await ui.home.editLayout(page);
  const block = page.locator('[data-home-block="standalone-bins"]');
  for (const theme of ['light', 'dark']) {
    const toggle = page.getByRole('button', { name: `Switch to ${theme} mode`, exact: true });
    if (await toggle.count()) await toggle.click();
    const appearance = await page.evaluate(() => {
      const area = document.querySelector<HTMLElement>('[data-area="Living room"]')!;
      const standalone = document.querySelector<HTMLElement>('[data-home-block="standalone-bins"] .nh-home-edit-surface')!;
      const style = (element: HTMLElement) => ({
        color: getComputedStyle(element.querySelector('h2')!).color,
        background: getComputedStyle(element).backgroundColor,
        outline: getComputedStyle(element).outlineColor,
      });
      return [style(area), style(standalone)];
    });
    expect(appearance[0]).toEqual(appearance[1]);
  }
  await expect(block.getByRole('button', { name: 'Delete Bin collections', exact: true })).toBeVisible();
  const handle = block.getByRole('button', { name: 'Drag Bin collections card', exact: true });
  await expect(handle).toBeVisible();
  await block.locator('[data-card-id="bins-summary"]').hover();
  await expect(block.getByRole('button', { name: 'Move card', exact: true })).toBeVisible();
  const from = (await handle.boundingBox())!;
  const card = (await block.locator('[data-card-id="bins-summary"]').boundingBox())!;
  expect(from.y + from.height).toBeLessThanOrEqual(card.y + 1);
  const target = (await ui.home.area(page, 'Living room').boundingBox())!;
  const saving = page.waitForRequest((request) => request.url().endsWith('/api/home/layout') && request.method() === 'PUT');
  await dragBetween(page,
    { x: from.x + from.width / 2, y: from.y + from.height / 2 },
    { x: target.x + target.width * 0.25, y: target.y + 16 });
  const savedOrder = (await saving).postDataJSON().blocks.map((item: { id: string }) => item.id);
  expect(savedOrder.indexOf('standalone-bins')).toBeLessThan(2);
  await expect(page.getByRole('button', { name: 'Done', exact: true })).toBeEnabled();
  await page.reload();
  await expect(page.locator('.nh-home-block')).toHaveCount(savedOrder.length);
  expect(await page.locator('.nh-home-block').evaluateAll((nodes) => nodes.map((node) => node.getAttribute('data-home-block')))).toEqual(savedOrder);
  await ui.home.editLayout(page);
  const earlier = block.getByRole('button', { name: 'Move Bin collections earlier', exact: true });
  if (await earlier.isEnabled()) {
    await earlier.click();
    await expect(page.locator('.nh-home-block').first()).toHaveAttribute('data-home-block', 'standalone-bins');
  }
  await expect(page.getByRole('button', { name: 'Done', exact: true })).toBeEnabled();
  const areaHandle = ui.home.dragHandle(page, 'Garden');
  const areaFrom = (await areaHandle.boundingBox())!;
  const firstBlock = (await page.locator('.nh-home-block').first().boundingBox())!;
  const areaSaving = page.waitForRequest((request) => request.url().endsWith('/api/home/layout') && request.method() === 'PUT');
  await dragBetween(page,
    { x: areaFrom.x + areaFrom.width / 2, y: areaFrom.y + areaFrom.height / 2 },
    { x: firstBlock.x + firstBlock.width * 0.25, y: firstBlock.y + 16 });
  const reordered = (await areaSaving).postDataJSON().blocks.map((item: { id: string }) => item.id);
  expect(reordered.indexOf('browser-area-1')).toBeLessThan(2);
  await expect(page.getByRole('button', { name: 'Done', exact: true })).toBeEnabled();
  await page.reload();
  await expect(page.locator('.nh-home-block')).toHaveCount(reordered.length);
  expect(await page.locator('.nh-home-block').evaluateAll((nodes) => nodes.map((node) => node.getAttribute('data-home-block')))).toEqual(reordered);
  await ui.home.editLayout(page);
  await block.getByRole('button', { name: 'Delete Bin collections', exact: true }).click();
  await page.getByRole('dialog').getByRole('button', { name: 'Cancel', exact: true }).click();
  await expect(block).toBeVisible();
  await block.getByRole('button', { name: 'Delete Bin collections', exact: true }).click();
  await page.getByRole('dialog').getByRole('button', { name: 'Delete card', exact: true }).click();
  await expect(block).toHaveCount(0);
  await page.reload();
  await expect(block).toHaveCount(0);
  expect((await (await page.request.get(`${AUTOMATION_HUB_API}/integrations/instances`)).json()).data).toHaveLength(1);
});

test('dragging a card between groups saves one atomic placement change', async ({ page }) => {
  await page.setViewportSize({ width: 1440, height: 1000 });
  await signIn(page, ui.home.path);
  await expect(page.getByRole('button', { name: 'Edit layout' })).toBeEnabled();
  const original = (await (await page.request.get(`${HUB_API}/home/layout`)).json()).data;
  const names = await ui.home.groupTitles(page).allTextContents();
  await ui.home.editLayout(page);
  const source = ui.home.area(page, names[0]);
  const target = ui.home.area(page, names[1]);
  const card = source.locator('.nh-home-card').first();
  const id = await card.getAttribute('data-card-id');
  await card.hover();
  const handle = card.getByRole('button', { name: 'Drag card', exact: true });
  await handle.scrollIntoViewIfNeeded();
  const from = (await handle.boundingBox())!;
  const to = (await target.locator('.nh-home-card').first().boundingBox())!;
  let writes = 0;
  page.on('request', (request) => {
    if (request.url().endsWith('/api/home/layout') && request.method() === 'PUT') writes++;
  });
  try {
    await dragBetween(
      page,
      { x: from.x + from.width / 2, y: from.y + from.height / 2 },
      { x: to.x + to.width / 2, y: to.y + to.height / 2 },
      20,
      // Native HTML dragover admission is asynchronous. Release only once the destination
      // actually owns the dragged DOM node, as a person pausing over the drop target would.
      async () => {
        await expect(target.locator(`[data-card-id="${id}"]`)).toHaveCount(1);
      }
    );
    await expect(target.locator(`[data-card-id="${id}"]`)).toBeVisible();
    await expect(source.locator(`[data-card-id="${id}"]`)).toHaveCount(0);
    await expect(page.getByRole('button', { name: 'Done', exact: true })).toBeEnabled();
    expect(writes).toBe(1);
    await page.reload();
    await expect(target.locator(`[data-card-id="${id}"]`)).toBeVisible();
  } finally {
    const current = (await (await page.request.get(`${HUB_API}/home/layout`)).json()).data;
    expect(
      (
        await page.request.put(`${HUB_API}/home/layout`, {
          headers: { 'Origin': 'http://localhost:4100', 'X-NodeHerder-Request': '1', 'If-Match': current.etag },
          data: { ...original, etag: current.etag },
        })
      ).status()
    ).toBe(200);
  }
});

// Spec 007 US-02: dashboards must render hub groups and switch devices with the same command.
test.describe('home dashboard', () => {
  test('shows all dashboard groups', async ({ page }) => {
    await signIn(page, ui.home.path);
    await expect(ui.home.groupTitles(page)).toHaveCount(MOCK.groups.length);
  });

  test('switching a light sends deviceSetValue for its state expose', async ({ page, ws }) => {
    await signIn(page, ui.home.path);
    // the mock brings the living room light online on its first update (~8 s)
    await ws.waitForReceived('deviceUpdated', (p) => p.id === MOCK.livingRoomLight.id, 15_000);
    const tile = ui.home.tile(page, 'living room group', 'Brightness');
    await ui.home.toggleControl(tile).click();
    const frame = await ws.waitForSent('deviceSetValue');
    expect(frame.payload).toMatchObject({ id: MOCK.livingRoomLight.id, name: 'state' });
    expect(['ON', 'OFF']).toContain(frame.payload.value);
  });

  // Spec 007 US-06 (AC-19…AC-22): areas can be rearranged and the order is saved on the hub.
  test('arrow buttons move an area and save its new position', async ({ page }) => {
    await signIn(page, ui.home.path);
    const titles = ui.home.groupTitles(page);
    await expect(titles).toHaveCount(MOCK.groups.length);
    const before = await titles.allTextContents();
    await ui.home.editLayout(page);
    const saving = page.waitForRequest(
      (request) => request.url().endsWith('/api/home/layout') && request.method() === 'PUT'
    );
    await ui.home.moveLater(page, before[0]).click();
    await expect(titles).toHaveText([before[1], before[0], ...before.slice(2)]);
    const saved = (await saving).postDataJSON();
    expect(saved.blocks[1].name).toBe(before[0]);
    // put it back so other specs see the original order
    await page.getByRole('button', { name: `Move ${before[0]} earlier` }).click();
    await expect(titles).toHaveText(before);
    await ui.home.done(page);
  });

  test('dragging an area by its handle reorders and saves', async ({ page }) => {
    await signIn(page, ui.home.path);
    const titles = ui.home.groupTitles(page);
    await expect(titles).toHaveCount(MOCK.groups.length);
    const before = await titles.allTextContents();
    await ui.home.editLayout(page);
    const handle = ui.home.dragHandle(page, before[0]);
    const target = ui.home.area(page, before[1]);
    const from = (await handle.boundingBox())!;
    const to = (await target.boundingBox())!;
    // areas collapse to their headers once one is held, so aim just past the next header
    const start = { x: from.x + from.width / 2, y: from.y + from.height / 2 };
    const end =
      to.y > from.y + from.height * 2 ? { x: start.x, y: start.y + 90 } : { x: to.x + to.width * 0.7, y: to.y + 16 };
    const saving = page.waitForRequest(
      (request) => request.url().endsWith('/api/home/layout') && request.method() === 'PUT'
    );
    await dragBetween(page, start, end);
    await expect(titles.first()).toHaveText(before[1]);
    const saved = (await saving).postDataJSON();
    expect(saved.blocks[0].name).toBe(before[1]);
    // Touch collapse can move past more than one area. Restore the dragged area from its
    // accepted position rather than assuming a single adjacent swap.
    const position = (await titles.allTextContents()).indexOf(before[0]);
    for (let index = 0; index < position; index++) {
      await page.getByRole('button', { name: `Move ${before[0]} earlier` }).click();
      await expect(page.getByRole('button', { name: 'Done', exact: true })).toBeEnabled();
    }
    await expect(titles).toHaveText(before);
  });
});
