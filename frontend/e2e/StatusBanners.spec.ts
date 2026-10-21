import { test, expect, signIn, HUB_API } from './support/fixtures';
import { ui } from './support/ui';

// Production owner: src/components/shell/StatusBanners.vue.
async function observeWarnings(page: import('@playwright/test').Page) {
  await page.addInitScript(() => {
    const seen: string[] = [];
    Object.assign(window, { startupWarnings: seen });
    new MutationObserver(() => {
      document.querySelectorAll('.nh-banner.is-danger,.nh-banner.is-warn').forEach((element) => {
        seen.push(element.textContent ?? '');
      });
    }).observe(document, { subtree: true, childList: true, attributes: true });
  });
}

for (const flow of ['OAuth', 'cookie']) {
  test(`${flow} landing and authenticated reload never render an outage warning`, async ({ page }) => {
    await observeWarnings(page);
    if (flow === 'OAuth') await signIn(page);
    else {
      const response = await page.request.post(`${HUB_API}/auth/login`);
      const { url } = await response.json();
      await page.request.get(url);
      await page.goto('/');
    }
    await expect(ui.home.groupTitles(page).first()).toBeVisible();
    await expect(page.locator('.nh-conn.is-connected')).toBeVisible();
    expect(await page.evaluate(() => Reflect.get(window, 'startupWarnings'))).toEqual([]);
    await page.reload();
    await expect(page.locator('.nh-conn.is-connected')).toBeVisible();
    expect(await page.evaluate(() => Reflect.get(window, 'startupWarnings'))).toEqual([]);
  });
}

test('delayed successful session validation has no outage warning', async ({ page }) => {
  await signIn(page);
  await expect(page.locator('.nh-conn.is-connected')).toBeVisible();
  let release!: () => void;
  const gate = new Promise<void>((resolve) => { release = resolve; });
  await page.route(`${HUB_API}/auth/me`, async (route) => { await gate; await route.continue(); });
  try {
    await page.reload();
    await expect(page.locator('#nh-content')).toBeAttached();
    await expect(page.locator('.nh-banner.is-danger,.nh-banner.is-warn')).toHaveCount(0);
    release();
    await expect(page.locator('.nh-conn.is-connected')).toBeVisible();
  } finally { release(); }
});

test('snapshot failure and retry preserve socket truth and content position', async ({ page }, testInfo) => {
  let release!: () => void;
  const gate = new Promise<void>((resolve) => { release = resolve; });
  await page.route(`${HUB_API}/hubstate`, async (route) => {
    await gate;
    await route.fulfill({ status: 503, json: {} });
  });
  try {
    await signIn(page);
    await expect(page.locator('.nh-conn.is-connected')).toBeVisible();
    const before = await page.locator('#nh-content').boundingBox();
    release();
    const retry = page.getByRole('button', { name: 'Retry loading home' });
    await expect(retry).toBeVisible();
    await expect(page.locator('.nh-conn.is-connected')).toBeVisible();
    expect((await page.locator('#nh-content').boundingBox())?.y).toBe(before?.y);
    await page.screenshot({ path: testInfo.outputPath('snapshot-failure.png') });
    await retry.focus();
    await expect(retry).toBeFocused();
    await page.unroute(`${HUB_API}/hubstate`);
    await page.keyboard.press('Enter');
    await expect(retry).toHaveCount(0);
    await expect(ui.home.groupTitles(page).first()).toBeVisible();
    await expect(page.locator('.nh-conn.is-connected')).toBeVisible();
    expect((await page.locator('#nh-content').boundingBox())?.y).toBe(before?.y);
  } finally { release(); }
});

for (const path of ['/groupdashboard', '/panel']) {
  test(`actual connection loss on ${path} does not move content and can be dismissed`, async ({ page }, testInfo) => {
    let unavailable = false;
    let close!: () => void;
    await page.routeWebSocket('ws://localhost:4110/ws', (socket) => {
      if (unavailable) { socket.close({ code: 1012 }); return; }
      socket.connectToServer();
      close = () => socket.close({ code: 1012 });
    });
    await signIn(page, path);
    await expect(page.locator('.nh-conn.is-connected')).toBeVisible();
    const content = page.locator('#nh-content');
    const before = await content.boundingBox();
    unavailable = true;
    close();
    await expect(page.getByText('Reconnecting.', { exact: true })).toBeVisible();
    expect((await content.boundingBox())?.y).toBe(before?.y);
    await page.screenshot({ path: testInfo.outputPath('connection-loss.png') });
    const dismiss = page.getByRole('button', { name: 'Dismiss connection notices' });
    await dismiss.focus();
    await page.keyboard.press('Enter');
    await expect(page.getByText('Reconnecting.', { exact: true })).toHaveCount(0);
    await expect(page.locator('.nh-conn')).toBeFocused();
    await expect(page.locator('.nh-conn.is-reconnecting')).toBeVisible();
    unavailable = false;
    await expect(page.locator('.nh-conn.is-connected')).toBeVisible({ timeout: 10000 });
    unavailable = true;
    close();
    await expect(page.getByText('Reconnecting.', { exact: true })).toBeVisible();
    unavailable = false;
    expect((await content.boundingBox())?.y).toBe(before?.y);
  });
}

test('permit countdown survives panel navigation and resets for the next confirmed session', async ({ page }) => {
  let send!: (active: boolean) => void;
  await page.routeWebSocket('ws://localhost:4110/ws', (socket) => {
    socket.connectToServer();
    send = (permitJoin) => socket.send(JSON.stringify({ type: 'bridgeConfig', payload: { permitJoin, maxTimeAllowed: { value: 120, unit: 'seconds' } } }));
  });
  await signIn(page, '/groupdashboard');
  await expect(page.locator('.nh-conn.is-connected')).toBeVisible();
  send(true);
  const banner = page.getByRole('status').filter({ hasText: 'Permit join is open' });
  await expect(banner).toContainText('2:00 left');
  await ui.shell.openNav(page);
  await ui.shell.nav(page).getByRole('link', { name: /panel/i }).click();
  await expect(page).toHaveURL(/\/panel$/);
  send(false);
  send(true);
  await ui.panel.exit(page).click();
  await expect(banner).toContainText(/1:5\d left|2:00 left/);
  send(false);
  await expect(banner).toHaveCount(0);
  send(true);
  await expect(banner).toContainText('2:00 left');
  send(false);
});
