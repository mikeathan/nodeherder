import { test, expect, signIn, dragBetween } from './support/fixtures';
import { MOCK } from './support/mock';

test('device-page sliders use simple rails, aligned values and usable targets in both modes', async ({ page }, info) => {
  await page.emulateMedia({ reducedMotion: 'reduce' });
  await page.addInitScript(() => localStorage.setItem('nodeherder_theme', JSON.stringify({
    v: 1, theme: 'graphite', mode: 'dark',
  })));
  await page.route('**/api/hubstate', async (route) => {
    const response = await route.fetch();
    const state = await response.json();
    const light = state.devices.find((d: { id: string }) => d.id === MOCK.livingRoomLight.id);
    light.availability = 'online';
    light.exposes.brightness.data = 184;
    Object.assign(light.exposes.color_temp, { data: 370, unit: 'mired', attributes: { min: 153, max: 500 } });
    await route.fulfill({ response, json: state });
  });
  // Keep fixture values stable and prevent all test commands reaching the hub.
  await page.routeWebSocket('ws://localhost:4110/ws', (socket) => socket.onMessage(() => {}));
  await signIn(page, `/devicepage/${MOCK.livingRoomLight.id}`);
  const controls = page.getByRole('tabpanel', { name: 'Controls' });
  for (const mode of ['dark', 'light']) {
    await page.locator('html').evaluate((html, mode) => html.setAttribute('data-mode', mode), mode);
    for (const width of [320, 390, 768, 1280]) {
      await page.setViewportSize({ width, height: 900 });
      await page.evaluate(() => new Promise<void>((resolve) => requestAnimationFrame(() => requestAnimationFrame(() => resolve()))));
      expect(await page.evaluate(() => document.documentElement.scrollWidth)).toBeLessThanOrEqual(width);
      const ends: number[] = [];
      for (const name of ['Brightness', 'Colour temperature']) {
        ends.push(await controls.getByRole('slider', { name, exact: true }).evaluate((input) =>
          input.parentElement!.parentElement!.querySelector('.nh-xc-out')!.getBoundingClientRect().right));
      }
      expect(Math.max(...ends) - Math.min(...ends)).toBeLessThan(1);
      for (const slider of await controls.getByRole('slider').all()) {
        const geometry = await slider.evaluate((input) => {
          const range = input.parentElement!;
          const grip = range.querySelector('.nh-range-grip')?.getBoundingClientRect();
          return { height: input.getBoundingClientRect().height,
            rail: range.querySelector('.nh-range-visual')!.getBoundingClientRect().height,
            gripWidth: grip?.width ?? null, gripHeight: grip?.height ?? null,
            background: getComputedStyle(range.querySelector('.nh-range-track')!).backgroundImage };
        });
        expect(geometry.height).toBeGreaterThanOrEqual(44);
        expect(geometry.rail).toBeLessThanOrEqual(6);
        if (geometry.gripWidth !== null && geometry.gripHeight !== null) {
          expect(Math.abs(geometry.gripWidth - geometry.gripHeight)).toBeLessThan(1);
        }
        expect(geometry.background).toBe('none');
      }
      expect(await controls.evaluate((el) => el.scrollWidth <= el.clientWidth)).toBe(true);
      await controls.locator('.nh-card').first().screenshot({ path: info.outputPath(`device-controls-${mode}-${width}.png`) });
    }
  }
  const brightness = controls.getByRole('slider', { name: 'Brightness', exact: true });
  await brightness.focus();
  await page.keyboard.press('Tab');
  await page.keyboard.press('Shift+Tab');
  await expect(brightness).toBeFocused();
  expect(await brightness.evaluate((input) => getComputedStyle(input.parentElement!).outlineStyle)).toBe('solid');
});

// Characterizes the reported failure with controlled browser-side hub responses.
// No command reaches the mock hub or any physical device. The Go investigation separately
// reproduces why the production backend can suppress numeric echoes for 60 seconds.
for (const gesture of ['keyboard', 'drag'] as const) {
  test(`brightness ${gesture} sends again but times out when its confirmation is suppressed`, async ({ page }) => {
    const commands: { id: string; name: string; value: number }[] = [];
    await page.route('**/api/hubstate', async (route) => {
      const response = await route.fetch();
      const state = await response.json();
      const light = state.devices.find((d: { id: string }) => d.id === MOCK.livingRoomLight.id);
      light.availability = 'online';
      light.exposes.brightness.data = 50;
      await route.fulfill({ response, json: state });
    });
    await page.routeWebSocket('ws://localhost:4110/ws', (socket) => {
      socket.onMessage((raw) => {
        const frame = JSON.parse(String(raw));
        if (frame.type !== 'deviceSetValue') return;
        commands.push(frame.payload);
        if (commands.length === 2) return; // device echo filtered by backend sampling
        socket.send(
          JSON.stringify({
            type: 'deviceUpdated',
            payload: {
              id: frame.payload.id,
              data: { [frame.payload.name]: frame.payload.value },
            },
          })
        );
      });
    });
    await signIn(page, `/devicepage/${MOCK.livingRoomLight.id}`);
    const slider = page.getByRole('slider', { name: 'Brightness', exact: true });
    await expect(slider).toBeEnabled();

    async function adjust() {
      if (gesture === 'keyboard') {
        await slider.focus();
        await slider.press('ArrowRight');
        return;
      }
      const box = (await slider.boundingBox())!;
      const value = Number(await slider.inputValue());
      const min = Number(await slider.getAttribute('min'));
      const max = Number(await slider.getAttribute('max'));
      const x = (v: number) => box.x + 8 + ((v - min) / (max - min)) * (box.width - 16);
      const y = box.y + box.height / 2;
      await dragBetween(page, { x: x(value), y }, { x: x(value + 40), y });
    }

    await adjust();
    await expect.poll(() => commands.length).toBe(1);
    expect(commands[0]).toEqual({ id: MOCK.livingRoomLight.id, name: 'brightness', value: expect.any(Number) });
    expect(commands[0].value).toBeGreaterThan(50);
    await expect(slider).toBeEnabled();

    await adjust();
    await expect.poll(() => commands.length).toBe(2);
    expect(commands[1].value).toBeGreaterThan(commands[0].value);
    await expect(slider).toBeDisabled();
    await expect(slider).toHaveAttribute('aria-busy', 'true');
    await expect(page.getByText(/Sending|Waiting for device confirmation/)).toHaveCount(0);
    await expect(page.locator('.nh-xc .nh-spin, .nh-xc-pending, .nh-xc-card-wait')).toHaveCount(0);
    await expect(page.getByText(/did not confirm the change/)).toBeVisible({ timeout: 7000 });
    await expect(slider).toBeEnabled();
    // A timed-out draft must return to the device-confirmed value.
    await expect(slider).toHaveValue(String(commands[0].value));

    await slider.focus();
    await adjust();
    await expect.poll(() => commands.length).toBe(3);
    expect(commands[2].value).toBe(commands[1].value);
    await expect(slider).toBeEnabled();
    await expect(page.getByText('Sending…', { exact: true })).toHaveCount(0);
  });
}
