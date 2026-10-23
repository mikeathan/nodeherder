import { test, expect, signIn } from './support/fixtures';
import { ui } from './support/ui';
import { MOCK } from './support/mock';

// Spec 007 US-07 (AC-23, AC-24): panel mode shows each Home area and exits back to Home.
test.describe('panel mode', () => {
  test('shows one tab per area and exits to Home', async ({ page }) => {
    await signIn(page, ui.panel.path);
    await expect(ui.panel.roomTabs(page)).toHaveCount(MOCK.groups.length);
    await ui.panel.roomTabs(page).nth(1).click();
    await expect(ui.panel.roomTabs(page).nth(1)).toHaveAttribute('aria-selected', 'true');
    await ui.panel.exit(page).click();
    await expect(page).toHaveURL(/\/groupdashboard$/);
  });
});

// Spec 011: the Panel view owns this pin-and-return flow.
test.describe('browser default view', () => {
  const pinName = 'Pin view as the default on this browser';
  const unpinName = 'Unpin view as the default on this browser';

  test('pins explicitly, preserves direct links, replaces and clears the default', async ({ page }, testInfo) => {
    await signIn(page, '/panel');
    const pin = page.getByRole('button', { name: pinName });
    await pin.focus();
    await expect(pin).toBeFocused();
    await page.keyboard.press('Enter');
    await expect(page.getByRole('button', { name: unpinName })).toHaveAttribute('aria-pressed', 'true');
    const pinned = page.getByRole('button', { name: unpinName });
    await pinned.hover();
    const roomColors = await page.getByRole('button', { name: 'Rooms', exact: true }).evaluate((el) => {
      const style = getComputedStyle(el);
      return { color: style.color, background: style.backgroundColor };
    });
    const pinColors = await pinned.evaluate((el) => {
      const style = getComputedStyle(el);
      return { color: style.color, background: style.backgroundColor, outline: style.boxShadow };
    });
    // Saved preference must not look like a second selected panel page.
    expect(pinColors.background).not.toBe(roomColors.background);
    expect(pinColors.outline).not.toBe('none');
    await expect(page.getByRole('button', { name: 'Rooms', exact: true })).toHaveAttribute('aria-pressed', 'true');
    await page.screenshot({ path: testInfo.outputPath('panel-pinned.png'), fullPage: true, scale: 'css' });
    await page.reload();
    await expect(page.getByRole('button', { name: unpinName })).toBeVisible();
    await page.goto('/');
    await expect(page).toHaveURL(/\/panel$/);

    await ui.panel.exit(page).click();
    await expect(page).toHaveURL(/\/groupdashboard$/);
    await expect(page.getByRole('button', { name: pinName })).toHaveAttribute('aria-pressed', 'false');
    await page.goto('/?redirect=/settings');
    await expect(page).toHaveURL(/\/settings$/);
    await page.goto('/');
    await expect(page).toHaveURL(/\/panel$/);

    await page.goto('/overview');
    await page.getByRole('button', { name: pinName }).click();
    await page.screenshot({ path: testInfo.outputPath('overview-pinned.png'), fullPage: true, scale: 'css' });
    await page.goto('/');
    await expect(page).toHaveURL(/\/overview$/);
    await ui.shell.signOut(page);
    await signIn(page);
    await expect(page).toHaveURL(/\/overview$/);
    await page.getByRole('button', { name: unpinName }).click();
    await page.goto('/');
    await expect(page).toHaveURL(/\/groupdashboard$/);
    await page.goto(`/devicepage/${MOCK.atticSocket.id}`);
    await expect(page.getByRole('button', { name: pinName })).toHaveCount(0);
    await page.goto('/creator');
    await expect(page.getByRole('button', { name: pinName })).toHaveCount(0);
  });

  test('sign-out waits for a deferred shell before opening navigation', async ({ page }) => {
    await signIn(page, '/overview');
    await expect(page.getByRole('navigation', { name: 'Main', includeHidden: true })).toBeAttached();
    let release!: () => void;
    const gate = new Promise<void>((resolve) => { release = resolve; });
    let gated = false;
    // Vite can append an HMR timestamp; hold the module regardless of its query string.
    await page.route((url) => url.pathname === '/src/components/shell/AppShell.vue', async (route) => {
      gated = true;
      await gate;
      await route.continue();
    });
    try {
      await page.goto('/overview', { waitUntil: 'domcontentloaded' });
      await expect.poll(() => gated).toBe(true);
      await expect(page.getByRole('status').filter({ hasText: 'Loading NodeHerder…' })).toBeVisible();
      const skeleton = page.locator('.shell-loading');
      const box = (await skeleton.boundingBox())!;
      expect(box.width).toBe(page.viewportSize()!.width);
      expect(box.height).toBe(page.viewportSize()!.height);
      expect(await skeleton.evaluate((el) => getComputedStyle(el).backgroundColor)).not.toBe('rgba(0, 0, 0, 0)');
      await expect(skeleton.getByRole('button')).toHaveCount(0);
      const signingOut = ui.shell.signOut(page);
      // Complete an ordered browser read while the module gate is still closed. The old
      // helper's one-shot visibility check skips the phone drawer in this state.
      expect(await page.getByRole('button', { name: 'Open navigation' }).isVisible()).toBe(false);
      release();
      await signingOut;
      await expect(ui.login.signInButton(page)).toBeVisible();
    } finally {
      release();
    }
  });

  test('invalid storage falls back; blocked preference storage still allows navigation', async ({ page }) => {
    await signIn(page);
    await page.evaluate(() => localStorage.setItem('nodeherder_default_view', 'unknown-view'));
    await page.goto('/');
    await expect(page).toHaveURL(/\/groupdashboard$/);
    await page.addInitScript(() => {
      const guard = (key: string) => {
        if (key === 'nodeherder_default_view') throw new Error('Preference storage blocked');
      };
      const getItem = Storage.prototype.getItem;
      Storage.prototype.getItem = function (this: Storage, key: string) {
        guard(key);
        return getItem.call(this, key);
      };
      const setItem = Storage.prototype.setItem;
      Storage.prototype.setItem = function (this: Storage, key: string, value: string) {
        guard(key);
        setItem.call(this, key, value);
      };
      const removeItem = Storage.prototype.removeItem;
      Storage.prototype.removeItem = function (this: Storage, key: string) {
        guard(key);
        removeItem.call(this, key);
      };
    });
    await page.goto('/panel');
    await page.getByRole('button', { name: pinName }).click();
    await expect(page.getByRole('button', { name: unpinName })).toHaveAttribute('aria-pressed', 'true');
    await ui.panel.exit(page).click();
    await page.getByRole('button', { name: pinName }).click();
    await expect(page.getByRole('button', { name: unpinName })).toHaveAttribute('aria-pressed', 'true');
    await page.getByRole('button', { name: unpinName }).click();
    await page.reload();
    await page.goto('/');
    await expect(page).toHaveURL(/\/groupdashboard$/);
  });
});


test('panel dock stays compact with accessible touch targets and keyboard navigation', async ({ page }, testInfo) => {
  await signIn(page, '/panel');
  const dock = page.getByRole('navigation', { name: 'Panel', exact: true });
  const pin = dock.getByRole('button', { name: 'Pin view as the default on this browser' });
  const status = dock.getByRole('button', { name: 'Status', exact: true });
  const rooms = dock.getByRole('button', { name: 'Rooms', exact: true });
  await status.focus(); await status.press('Enter');
  await expect(status).toHaveAttribute('aria-pressed', 'true');
  await expect(rooms).toHaveAttribute('aria-pressed', 'false');
  await rooms.focus(); await rooms.press('Enter');
  await expect(rooms).toHaveAttribute('aria-pressed', 'true');
  const widths = testInfo.project.name === 'mobile' ? [390, 320] : [1280];
  for (const width of widths) {
    await page.setViewportSize({ width, height: 844 });
    const bounds = await dock.boundingBox();
    expect(bounds).not.toBeNull();
    expect(bounds!.height).toBeLessThanOrEqual(60);
    expect(bounds!.x).toBeGreaterThanOrEqual(0);
    expect(bounds!.x + bounds!.width).toBeLessThanOrEqual(width);
    for (const control of await dock.locator('button, a').all()) {
      const box = (await control.boundingBox())!;
      expect(box.width).toBeGreaterThanOrEqual(44);
      expect(box.height).toBeGreaterThanOrEqual(44);
    }
    if (width <= 760) await expect(pin.locator('span')).toBeHidden();
    else await expect(pin.locator('span')).toBeVisible();
    expect(await page.evaluate(() => document.documentElement.scrollWidth > innerWidth)).toBe(false);
    await page.screenshot({ path: testInfo.outputPath(`compact-panel-${width}.png`), fullPage: true, scale: 'css' });
  }
  await pin.press('Enter');
  await expect(dock.getByRole('button', { name: 'Unpin view as the default on this browser' })).toHaveAttribute('aria-pressed', 'true');
  await page.getByRole('button', { name: 'Switch to dark mode', exact: true }).click();
  await page.screenshot({ path: testInfo.outputPath('compact-panel-dark-pinned.png'), fullPage: true, scale: 'css' });
  await dock.getByRole('link', { name: 'Exit panel', exact: true }).click();
  await expect(page).toHaveURL(/\/groupdashboard$/);
});

for (const stored of ['Infinity', '1.5', 'NaN']) {
  test(`invalid saved room ${stored} falls back without breaking the panel`, async ({ page }) => {
    await page.addInitScript((value) => localStorage.setItem('nodeherder_panel_room', value), stored);
    const errors: string[] = [];
    page.on('pageerror', (error) => errors.push(error.message));
    await signIn(page, '/panel');
    const tabs = page.getByRole('tablist', { name: 'Rooms' }).getByRole('tab');
    await expect(tabs.first()).toHaveAttribute('aria-selected', 'true');
    await expect(page.getByRole('navigation', { name: 'Panel', exact: true })).toBeVisible();
    expect(errors).toEqual([]);
  });
}

test('room keyboard focus follows selection and modal controls never change rooms', async ({ page }) => {
  await signIn(page, '/panel');
  const tabs = page.getByRole('tablist', { name: 'Rooms' }).getByRole('tab');
  await tabs.first().focus();
  await tabs.first().press('ArrowRight');
  await expect(tabs.nth(1)).toBeFocused();
  await expect(tabs.nth(1)).toHaveAttribute('aria-selected', 'true');
  await expect(tabs.first()).toHaveAttribute('tabindex', '-1');
  await tabs.nth(1).press('End');
  await expect(tabs.last()).toBeFocused();
  await expect(page.getByRole('tabpanel')).toHaveAttribute('aria-labelledby', await tabs.last().getAttribute('id') as string);
  await tabs.last().press('ArrowRight');
  await expect(tabs.first()).toBeFocused();
  await tabs.first().press('Home');
  await expect(tabs.first()).toHaveAttribute('aria-selected', 'true');
  await page.getByRole('button', { name: 'Appearance', exact: true }).click();
  const dialog = page.getByRole('dialog');
  await dialog.getByRole('button', { name: 'Dark', exact: true }).focus();
  await page.keyboard.press('ArrowRight');
  await expect(tabs.first()).toHaveAttribute('aria-selected', 'true');
  await page.keyboard.press('Escape');
  await page.locator('#nh-content').focus();
  await page.keyboard.press('Control+ArrowRight');
  await expect(tabs.first()).toHaveAttribute('aria-selected', 'true');
  await page.keyboard.press('ArrowRight');
  await expect(tabs.nth(1)).toHaveAttribute('aria-selected', 'true');
});

test('panel disconnect notices leave the dock usable in narrow and short viewports', async ({ page }, testInfo) => {
  let disconnected = false;
  let disconnect: (() => Promise<void>) | undefined;
  await page.routeWebSocket('ws://localhost:4110/ws', (socket) => {
    if (disconnected) { void socket.close(); return; }
    socket.connectToServer(); disconnect = () => socket.close();
  });
  await signIn(page, '/panel');
  disconnected = true; await disconnect?.();
  const notice = page.getByRole('complementary', { name: 'Connection notices' });
  await expect(notice).toBeVisible();
  for (const viewport of [{ width: 320, height: 844 }, { width: 844, height: 390 }]) {
    await page.setViewportSize(viewport);
    const dock = (await page.getByRole('navigation', { name: 'Panel', exact: true }).boundingBox())!;
    const bounds = (await notice.boundingBox())!;
    expect(bounds.y + bounds.height).toBeLessThanOrEqual(dock.y);
    expect(dock.y + dock.height).toBeLessThanOrEqual(viewport.height);
    await page.getByRole('button', { name: 'Status', exact: true }).click();
    await expect(page.getByRole('button', { name: 'Status', exact: true })).toHaveAttribute('aria-pressed', 'true');
    await page.getByRole('button', { name: 'Rooms', exact: true }).click();
  }
  await page.setViewportSize({ width: 390, height: 844 });
  await page.screenshot({ path: testInfo.outputPath('panel-notice-clear-of-dock.png'), fullPage: true, scale: 'css' });
  await notice.getByRole('button', { name: 'Dismiss connection notices' }).click();
  await expect(notice).toBeHidden();
  await page.getByRole('navigation', { name: 'Panel', exact: true }).getByRole('link', { name: 'Exit panel', exact: true }).click();
  await expect(page).toHaveURL(/\/groupdashboard$/);
});

test('12-hour clock remains on one line at 320px', async ({ page }) => {
  await page.setViewportSize({ width: 320, height: 844 });
  await signIn(page, '/panel');
  await page.locator('.hp-clock b').evaluate((el) => { el.textContent = '12:59 PM'; });
  const size = await page.locator('.hp-clock b').evaluate((el) => {
    const range = document.createRange(); range.selectNodeContents(el);
    return { lines: range.getClientRects().length, width: el.scrollWidth, available: el.clientWidth };
  });
  expect(size.lines).toBe(1);
  expect(size.width).toBeLessThanOrEqual(size.available + 1);
  expect(await page.evaluate(() => document.documentElement.scrollWidth > innerWidth)).toBe(false);
});
