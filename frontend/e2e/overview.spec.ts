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
    await expect(page.locator('.nh-activity').getByText('Activity recording is off')).toBeVisible();
    expect(await page.evaluate(() => localStorage.getItem('nodeherder_activity'))).toBe('off');
    await page.reload();
    await expect(ui.overview.activitySwitch(page)).toHaveAttribute('aria-checked', 'false');
    await ui.overview.activitySwitch(page).click();
    await expect(ui.overview.activitySwitch(page)).toHaveAttribute('aria-checked', 'true');
  });
});

// Spec 070: the verdict action must operate, not rely on a hash the app's scroll container ignores.
test('review in order moves focus to the ranked list', async ({ page }) => {
  await signIn(page, '/overview');
  const review = page.getByRole('button', { name: 'Review in order' });
  await expect(review).toBeVisible();
  await review.click();
  await expect(page.getByRole('heading', { name: 'Start here' })).toBeFocused();
});

// Spec 070: the health line is actionable; a status opens the device list already filtered.
test('offline health link opens the device list filtered to offline', async ({ page }) => {
  await signIn(page, '/overview');
  const link = page.getByRole('link', { name: /offline$/ }).first();
  await expect(link).toBeVisible();
  await link.click();
  await expect(page).toHaveURL(/\/devicelist\?status=offline$/);
  await expect(page.getByRole('combobox', { name: 'Status' }).locator('option:checked')).toHaveText('Offline');
});

test('device list reads status and search from the URL', async ({ page }) => {
  await signIn(page, '/devicelist?status=quiet');
  await expect(page.getByRole('combobox', { name: 'Status' }).locator('option:checked')).toHaveText('Quiet (not reporting)');
  await expect(page.getByText(/of 13/)).toBeVisible();
});
