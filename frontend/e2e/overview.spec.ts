import { test, expect, signIn } from './support/fixtures';
import { ui } from './support/ui';

// Spec 007 US-09 (AC-27): recent activity can be switched off and stays off after reload.
test.describe('overview', () => {
  test('recent activity switch turns recording off and is remembered', async ({ page }) => {
    await signIn(page, ui.overview.path);
    const sw = ui.overview.activitySwitch(page);
    await expect(sw).toHaveAttribute('aria-checked', 'true');
    await sw.click();
    await expect(sw).toHaveAttribute('aria-checked', 'false');
    await expect(page.getByText('Activity recording is off')).toBeVisible();
    expect(await page.evaluate(() => localStorage.getItem('nodeherder_activity'))).toBe('off');
    await page.reload();
    await expect(ui.overview.activitySwitch(page)).toHaveAttribute('aria-checked', 'false');
    await ui.overview.activitySwitch(page).click();
    await expect(ui.overview.activitySwitch(page)).toHaveAttribute('aria-checked', 'true');
  });
});
