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
    await expect.poll(() => pinned.evaluate((el) => {
      const style = getComputedStyle(el);
      return { color: style.color, background: style.backgroundColor };
    })).toEqual(roomColors);
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

  test('invalid storage falls back; blocked preference storage still allows navigation', async ({ page }) => {
    await signIn(page);
    await page.evaluate(() => localStorage.setItem('nodeherder_default_view', 'unknown-view'));
    await page.goto('/');
    await expect(page).toHaveURL(/\/groupdashboard$/);
    await page.addInitScript(() => {
      for (const method of ['getItem', 'setItem', 'removeItem'] as const) {
        const original = Storage.prototype[method];
        Storage.prototype[method] = function (key: string, value?: string) {
          if (key === 'nodeherder_default_view') throw new Error('Preference storage blocked');
          return original.call(this, key, value as string);
        } as typeof original;
      }
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
