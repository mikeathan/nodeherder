import type { Page, WebSocketRoute } from '@playwright/test';
import { test, expect, signIn, dragBetween } from './support/fixtures';
import { MOCK } from './support/mock';

// DeviceDashboard.vue: card presentation and mixed-state grid regressions.
// All requests target the isolated mock hub; device commands are intercepted here.
type Command = { id: string; name: string; value: unknown };

async function cardFixture(page: Page) {
  const commands: Command[] = [];
  let unknown = false;
  let connection: WebSocketRoute;
  await page.routeWebSocket('ws://localhost:4110/ws', (socket) => {
    connection = socket;
    socket.onMessage((raw) => {
      const frame = JSON.parse(String(raw));
      if (frame.type === 'deviceSetValue') commands.push(frame.payload);
    });
  });
  await page.route('**/api/hubstate', async (route) => {
    const response = await route.fetch();
    const state = await response.json();
    for (const device of state.devices) {
      if (/air sensor|attic alarm|attic light/i.test(device.friendly_name)) {
        device.availability = 'online';
        device.last_seen = Date.now();
        const values = { co2: 352, formaldehyd: 1, humidity: 83, pm25: 311, temperature: 20.1,
          voc: 11, battery: 100, brightness: 184, color_temp: 370, state: 'ON' };
        for (const [name, value] of Object.entries(values)) {
          if (device.exposes[name]) device.exposes[name].data = value;
        }
      }
    }
    const light = state.devices.find((device: { id: string }) => device.id === MOCK.livingRoomLight.id);
    light.availability = 'online';
    light.last_seen = Date.now();
    light.exposes.brightness.data = unknown ? null : 184;
    Object.assign(light.exposes.color_temp, { data: 370, unit: 'mired', attributes: { min: 153, max: 500 } });
    await route.fulfill({ response, json: state });
  });
  return {
    commands,
    setUnknown: () => { unknown = true; },
    confirm: (command: Command) => connection.send(JSON.stringify({
      type: 'deviceUpdated', payload: { id: command.id, data: { [command.name]: command.value } },
    })),
  };
}

test('mixed online and offline cards pack at natural heights with compact footers at all widths', async ({ page }, info) => {
  await page.emulateMedia({ reducedMotion: 'reduce' });
  await page.addInitScript(() => localStorage.setItem('nodeherder_theme', JSON.stringify({
    v: 1, theme: 'graphite', mode: 'dark',
  })));
  await cardFixture(page);
  await signIn(page, '/devicedashboard');
  const cards = page.locator('.nh-dcard');
  await expect(cards).toHaveCount(MOCK.deviceCount);
  await expect(cards.filter({ has: page.locator('.nh-overlay') }).first()).toBeVisible();

  for (const mode of ['dark', 'light']) {
    await page.locator('html').evaluate((html, mode) => html.setAttribute('data-mode', mode), mode);
    for (const width of [320, 390, 768, 1280, 1920]) {
      await page.setViewportSize({ width, height: 1000 });
      await page.evaluate(() => new Promise<void>((resolve) => requestAnimationFrame(() => requestAnimationFrame(() => resolve()))));
      const geometry = await cards.evaluateAll((elements) => elements.map((card) => {
        const rect = card.getBoundingClientRect();
        const footer = card.querySelector('footer')!.getBoundingClientRect();
        return { x: rect.x, y: rect.y, bottom: rect.bottom, width: rect.width, height: rect.height,
          footerHeight: footer.height, title: card.querySelector('.nh-dcard-t b')!.textContent,
          fragments: card.getClientRects().length, overflow: card.scrollWidth - card.clientWidth };
      }));
      expect(await page.evaluate(() => document.documentElement.scrollWidth)).toBeLessThanOrEqual(width);
      expect(geometry.every((card) => card.overflow <= 1)).toBe(true);
      const air = geometry.find((card) => card.title === 'Attic air sensor')!;
      const alarm = geometry.find((card) => card.title === 'Attic alarm')!;
      expect(air.footerHeight).toBeLessThan(56);
      expect(air.height - alarm.height).toBeGreaterThan(100);
      expect(geometry.every((card) => card.fragments === 1)).toBe(true);
      const gap = await page.locator('.nh-dcards').evaluate((grid) => parseFloat(getComputedStyle(grid).columnGap));
      const columns = [...new Set(geometry.map((card) => Math.round(card.x)))];
      const columnTops: number[] = [];
      for (const left of columns) {
        const column = geometry.filter((card) => Math.abs(card.x - left) < 1).sort((a, b) => a.y - b.y);
        columnTops.push(column[0].y);
        for (let i = 1; i < column.length; i++) {
          expect(Math.abs(column[i].y - column[i - 1].bottom - gap)).toBeLessThan(1);
        }
      }
      expect(Math.max(...columnTops) - Math.min(...columnTops)).toBeLessThan(1);
      expect(Math.max(...geometry.map((card) => card.width)) - Math.min(...geometry.map((card) => card.width))).toBeLessThan(1);
      const sliders = page.locator('.nh-dcard input[type="range"]');
      for (const slider of await sliders.all()) {
        const box = (await slider.boundingBox())!;
        expect(box.height).toBeGreaterThanOrEqual(44);
        const railHeight = await slider.evaluate((input) =>
          input.parentElement!.querySelector('.nh-range-visual')!.getBoundingClientRect().height);
        expect(railHeight).toBeLessThanOrEqual(6);
      }
      if ([390, 1280, 1920].includes(width)) {
        await page.screenshot({ path: info.outputPath(`cards-${mode}-${width}.png`), fullPage: true });
      }
    }
  }
});

test('friendly-name headers navigate to online and offline device pages without sending commands', async ({ page }) => {
  const fixture = await cardFixture(page);
  await signIn(page, '/devicedashboard');
  const card = page.locator(`.nh-dcard[data-device="${MOCK.livingRoomLight.id}"]`);
  const link = card.getByRole('link', { name: `Open ${MOCK.livingRoomLight.name} device page`, exact: true });
  await expect(link).toHaveAttribute('href', `/devicepage/${MOCK.livingRoomLight.id}`);
  await card.locator('.nh-dcard-t b').click();
  await expect(page).toHaveURL(new RegExp(`/devicepage/${MOCK.livingRoomLight.id}$`));
  await page.goBack();
  const offline = page.locator('.nh-dcard.is-offline').first();
  const offlineId = await offline.getAttribute('data-device');
  const offlineLink = offline.locator('.nh-dcard-head');
  await offlineLink.focus();
  await page.keyboard.press('Tab');
  await page.keyboard.press('Shift+Tab');
  await expect(offlineLink).toBeFocused();
  expect(await offlineLink.evaluate((el) => getComputedStyle(el).outlineStyle)).toBe('solid');
  await page.keyboard.press('Enter');
  await expect(page).toHaveURL(new RegExp(`/devicepage/${offlineId}$`));
  expect(fixture.commands).toHaveLength(0);
});

test('card sliders keep native commands, pending confirmation, keyboard focus and unknown readings', async ({ page }, info) => {
  await page.addInitScript(() => localStorage.setItem('nodeherder_theme', JSON.stringify({
    v: 1, theme: 'graphite', mode: 'dark',
  })));
  const fixture = await cardFixture(page);
  await signIn(page, '/devicedashboard');
  const card = page.locator(`.nh-dcard[data-device="${MOCK.livingRoomLight.id}"]`);
  const brightness = card.getByRole('slider', { name: 'Brightness', exact: true });
  await expect(brightness).toBeEnabled();
  const layout = () => page.locator('.nh-dcard, .nh-dcard footer, .nh-dcard .nh-xc-range').evaluateAll((elements) =>
    elements.map((element) => {
      const rect = element.getBoundingClientRect();
      return [rect.x + scrollX, rect.y + scrollY, rect.width, rect.height].map((value) => Math.round(value * 100) / 100);
    }));
  for (const mode of ['dark', 'light']) {
    await page.locator('html').evaluate((html, mode) => html.setAttribute('data-mode', mode), mode);
    await card.screenshot({ path: info.outputPath(`card-controls-${mode}.png`) });
  }
  await brightness.focus();
  const beforePending = await layout();
  await brightness.press('ArrowRight');
  await expect.poll(() => fixture.commands.length).toBe(1);
  expect(fixture.commands[0]).toEqual({ id: MOCK.livingRoomLight.id, name: 'brightness', value: 185 });
  await expect(brightness).toBeDisabled();
  await expect(brightness).toHaveAttribute('aria-busy', 'true');
  await expect(card.locator('.nh-spin, .nh-xc-pending, .nh-xc-card-wait')).toHaveCount(0);
  await expect(card.getByText(/Sending|Waiting for device confirmation/)).toHaveCount(0);
  expect(await layout()).toEqual(beforePending);
  fixture.confirm(fixture.commands[0]);
  await expect(brightness).toBeEnabled();
  expect(await layout()).toEqual(beforePending);
  await brightness.focus();
  await page.keyboard.press('Tab');
  await page.keyboard.press('Shift+Tab');
  await expect(brightness).toBeFocused();
  expect(await brightness.evaluate((input) => getComputedStyle(input.parentElement!).outlineStyle)).toBe('solid');

  const temperature = card.getByRole('slider', { name: 'Colour temperature', exact: true });
  const box = (await temperature.boundingBox())!;
  await dragBetween(page, { x: box.x + box.width * 0.4, y: box.y + box.height / 2 },
    { x: box.x + box.width * 0.7, y: box.y + box.height / 2 }, 14, async () => {
      expect(fixture.commands).toHaveLength(1);
    });
  await expect.poll(() => fixture.commands.length).toBe(2);
  expect(fixture.commands[1]).toMatchObject({ id: MOCK.livingRoomLight.id, name: 'color_temp' });
  expect(fixture.commands[1].value).toBeGreaterThanOrEqual(153);
  expect(fixture.commands[1].value).toBeLessThanOrEqual(500);
  fixture.confirm(fixture.commands[1]);
  await expect(temperature).toBeEnabled();

  fixture.setUnknown();
  await page.reload();
  await expect(card.locator('.nh-xc-out').first()).toHaveText('Unknown');
  const unknownRange = card.getByRole('slider', { name: 'Brightness', exact: true }).locator('..');
  await expect(unknownRange.locator('.nh-range-fill, .nh-range-grip')).toHaveCount(0);
  await expect(page.locator('.nh-dcard.is-offline input[type="range"]')).toHaveCount(0);
});

test('card commands show no progress indicators and never move cards, rails or footers', async ({ page }, info) => {
  await page.emulateMedia({ reducedMotion: 'reduce' });
  const fixture = await cardFixture(page);
  await signIn(page, '/devicedashboard');
  const card = page.locator(`.nh-dcard[data-device="${MOCK.livingRoomLight.id}"]`);
  const geometry = () => page.locator('.nh-dcard, .nh-dcard footer, .nh-dcard .nh-xc-range').evaluateAll((elements) =>
    elements.map((element) => {
      const rect = element.getBoundingClientRect();
      return [rect.x + scrollX, rect.y + scrollY, rect.width, rect.height].map((value) => Math.round(value * 100) / 100);
    }));
  for (const mode of ['light', 'dark']) {
    await page.locator('html').evaluate((html, mode) => html.setAttribute('data-mode', mode), mode);
    for (const width of [320, 390, 768, 1280]) {
      await page.setViewportSize({ width, height: 900 });
      for (const name of ['Brightness', 'Colour temperature']) {
        const slider = card.getByRole('slider', { name, exact: true });
        const control = slider.locator('../..');
        await slider.scrollIntoViewIfNeeded();
        await expect(slider).toBeEnabled();
        const before = await geometry();
        await slider.evaluate((element) => {
          const input = element as HTMLInputElement;
          input.value = String(Number(input.value) + Number(input.step));
          input.dispatchEvent(new Event('input', { bubbles: true }));
          input.dispatchEvent(new Event('change', { bubbles: true }));
        });
        await expect(slider).toBeDisabled();
        await expect(slider).toHaveAttribute('aria-busy', 'true');
        expect(await geometry()).toEqual(before);
        await expect(control.locator('.nh-spin, .nh-xc-pending, .nh-xc-card-wait')).toHaveCount(0);
        await expect(control.getByText(/Sending|Waiting for device confirmation/)).toHaveCount(0);
        const preset = control.getByRole('combobox');
        if (await preset.count()) await expect(preset).toHaveCSS('opacity', '1');
        if (width === 390 && name === 'Brightness') {
          await card.screenshot({ path: info.outputPath(`card-pending-${mode}.png`) });
        }
        fixture.confirm(fixture.commands.at(-1)!);
        await expect(slider).toBeEnabled();
        expect(await geometry()).toEqual(before);
      }
    }
  }
  const toggle = card.getByRole('switch');
  const confirmed = await toggle.getAttribute('aria-checked');
  await toggle.click();
  await expect(toggle).toHaveAttribute('aria-busy', 'true');
  await expect(toggle).toBeDisabled();
  await expect(toggle).toHaveAttribute('aria-checked', confirmed!);
  await expect(toggle).toHaveCSS('opacity', '1');
  await expect(toggle.locator('.nh-toggle-knob')).toHaveCSS('animation-name', 'none');
  await expect(card.getByText(/Sending|Waiting for device confirmation/)).toHaveCount(0);
  fixture.confirm(fixture.commands.at(-1)!);
  await expect(toggle).toBeEnabled();
  await expect(toggle).not.toHaveAttribute('aria-checked', confirmed!);
});
