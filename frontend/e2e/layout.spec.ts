import { test, expect, signIn } from './support/fixtures';
import { MOCK } from './support/mock';

// Spec 007 NFR-01, NFR-06: no screen scrolls sideways at desktop or phone width.
const PATHS = ['/overview', '/groupdashboard', '/devicedashboard', '/devicelist', `/devicepage/${MOCK.atticSocket.id}`, '/viewer', `/editor/${MOCK.presenceAutomation.id}`, '/creator', '/settings', '/consoleviewer', '/panel'];

test.describe('responsive layout', () => {
  test('no horizontal overflow on any main screen', async ({ page }) => {
    await signIn(page);
    for (const path of PATHS) {
      await page.goto(path);
      await page.waitForLoadState('networkidle');
      const overflow = await page.evaluate(() => {
        const scrollers = [document.documentElement, ...Array.from(document.querySelectorAll<HTMLElement>('.nh-app-main, .hp-content'))];
        return scrollers.map((el) => el.scrollWidth - el.clientWidth).filter((d) => d > 1);
      });
      expect(overflow, `horizontal overflow on ${path}`).toEqual([]);
    }
  });
});

test('mobile hamburger keeps all destinations without a bottom bar or reserved strip', async ({ page }) => {
  await page.setViewportSize({ width: 390, height: 844 });
  await signIn(page, '/groupdashboard');
  const menu = page.getByRole('button', { name: 'Open navigation', exact: true });
  const nav = page.getByRole('navigation', { name: 'Main', includeHidden: true });
  const content = page.locator('.nh-content');
  await expect(menu).toBeVisible();
  expect(await content.evaluate((el) => parseFloat(getComputedStyle(el).paddingBottom))).toBeLessThanOrEqual(32);
  await expect(page.getByRole('navigation', { name: 'Primary', includeHidden: true })).toHaveCount(0);
  const destinations = [
    ['Home', '/groupdashboard'], ['Devices', '/devicedashboard'], ['Automations', '/viewer'],
    ['Assistant', '/assistant'], ['Settings', '/settings'],
  ];
  for (const width of [320, 390, 768]) {
    await page.setViewportSize({ width, height: 844 });
    const menuBox = (await menu.boundingBox())!;
    expect(menuBox.width).toBeGreaterThanOrEqual(44);
    expect(menuBox.height).toBeGreaterThanOrEqual(44);
    await menu.focus();
    await menu.press('Enter');
    await expect(menu).toHaveAttribute('aria-expanded', 'true');
    for (const label of ['Overview', 'Home', 'Devices', 'Device list', 'Automations', 'Assistant', 'Console', 'Settings', 'Panel mode']) {
      await expect(nav.getByRole('link', { name: label, exact: true })).toBeVisible();
    }
    await expect(nav.getByRole('button', { name: 'Permit join', exact: true })).toBeVisible();
    await expect(nav.getByRole('button', { name: 'Sign out', exact: true })).toBeVisible();
    await menu.press('Escape');
    await expect(nav).not.toBeVisible();
    if (width !== 390) continue;
    for (const [label, path] of destinations) {
      await menu.click();
      const link = nav.getByRole('link', { name: label, exact: true });
      const box = (await link.boundingBox())!;
      expect(box.width).toBeGreaterThanOrEqual(44);
      expect(box.height).toBeGreaterThanOrEqual(44);
      await link.click();
      await expect(page).toHaveURL(new RegExp(`${path}$`));
      await expect(menu).toHaveAttribute('aria-expanded', 'false');
      await menu.click();
      await expect(nav.getByRole('link', { name: label, exact: true })).toHaveAttribute('aria-current', 'page');
      await menu.press('Escape');
    }
  }
  await page.goto('/creator');
  await menu.click();
  await expect(nav.getByRole('link', { name: 'Automations', exact: true })).toHaveAttribute('aria-current', 'page');
  await page.evaluate(() => document.documentElement.style.setProperty('--nh-safe-bottom', '32px'));
  expect(await nav.evaluate((el) => parseFloat(getComputedStyle(el).paddingBottom))).toBeGreaterThanOrEqual(32);
  await menu.press('Escape');
  await page.setViewportSize({ width: 390, height: 844 });
  expect(await content.evaluate((el) => parseFloat(getComputedStyle(el).paddingBottom))).toBe(32);
  expect(await page.locator('meta[name="viewport"]').getAttribute('content')).toContain('viewport-fit=cover');
  expect(await (await page.request.get('/manifest.webmanifest')).json()).toMatchObject({ name: 'NodeHerder', icons: [{ type: 'image/svg+xml' }] });
  await page.route('**/manifest.webmanifest', (route) => route.fulfill({ status: 404 }));
  await page.reload();
  await expect(menu).toBeVisible();
  await menu.click();
  await expect(nav).toBeVisible();
  await menu.press('Escape');
  await page.setViewportSize({ width: 1280, height: 844 });
  await expect(nav).toBeVisible();
  await expect(menu).not.toBeVisible();
});

test('browser theme colour follows the active appearance on toggle and reload', async ({ page }) => {
  await signIn(page, '/groupdashboard');
  await expect(page.getByRole('button', { name: /Switch to (light|dark) mode/ })).toBeVisible();
  const meta = page.locator('meta[name="theme-color"]');
  for (const mode of ['light', 'dark']) {
    await expect(page.getByRole('button', { name: /Switch to (light|dark) mode/ })).toBeVisible();
    const toggle = page.getByRole('button', { name: `Switch to ${mode} mode` });
    if (await toggle.count()) await toggle.click();
    await expect(page.locator('html')).toHaveAttribute('data-mode', mode);
    const background = await page.evaluate(() => getComputedStyle(document.documentElement).getPropertyValue('--nh-bg').trim());
    await expect(meta).toHaveAttribute('content', background);
    let release!: () => void;
    const gate = new Promise<void>((resolve) => { release = resolve; });
    await page.route((url) => url.pathname === '/src/components/shell/AppShell.vue', async (route) => {
      await gate;
      await route.continue();
    });
    try {
      // Appearance is ready before the lazy shell; exercise that ordering on every reload.
      await page.reload({ waitUntil: 'domcontentloaded' });
      await expect(page.getByRole('status').filter({ hasText: 'Loading NodeHerder…' })).toBeVisible();
      await expect(meta).toHaveAttribute('content', background);
    } finally {
      release();
    }
  }
});
