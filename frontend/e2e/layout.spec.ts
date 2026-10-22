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

test('mobile tabs and More drawer preserve navigation and platform metadata', async ({ page }) => {
  await page.setViewportSize({ width: 390, height: 844 });
  await signIn(page, '/groupdashboard');
  const nav = page.getByRole('navigation', { name: 'Primary', exact: true });
  await expect(nav.getByRole('link', { name: 'Home', exact: true })).toHaveAttribute('aria-current', 'page');
  for (const control of await nav.locator('a, button').all()) {
    const box = (await control.boundingBox())!;
    expect(box.width).toBeGreaterThanOrEqual(44);
    expect(box.height).toBeGreaterThanOrEqual(44);
  }
  const more = nav.getByRole('button', { name: 'More', exact: true });
  await more.click();
  await expect(more).toHaveAttribute('aria-expanded', 'true');
  await page.getByRole('navigation', { name: 'Main', exact: true }).getByRole('link', { name: 'Settings', exact: true }).click();
  await expect(more).toHaveAttribute('aria-current', 'page');
  await expect(more).toHaveAttribute('aria-expanded', 'false');
  await page.goto('/creator');
  await expect(nav.getByRole('link', { name: 'Automations', exact: true })).toHaveAttribute('aria-current', 'page');
  await page.evaluate(() => document.documentElement.style.setProperty('--nh-safe-bottom', '20px'));
  expect(await nav.evaluate((el) => getComputedStyle(el).paddingBottom)).toBe('20px');
  expect(await page.locator('meta[name="viewport"]').getAttribute('content')).toContain('viewport-fit=cover');
  expect(await (await page.request.get('/manifest.webmanifest')).json()).toMatchObject({ name: 'NodeHerder', icons: [{ type: 'image/svg+xml' }] });
  await page.route('**/manifest.webmanifest', (route) => route.fulfill({ status: 404 }));
  await page.reload();
  await expect(nav).toBeVisible();
});
