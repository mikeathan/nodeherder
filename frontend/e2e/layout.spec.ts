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
