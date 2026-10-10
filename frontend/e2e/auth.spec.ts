import { test, expect, signIn, HUB_API } from './support/fixtures';
import { ui } from './support/ui';
import { MOCK } from './support/mock';

// Spec 007 US-08 (AC-25, AC-26): OAuth redirect sign-in must keep working.
test.describe('sign in', () => {
  test('protected route shows sign-in; Google button completes the OAuth round trip', async ({ page }) => {
    await page.goto(ui.home.path);
    const button = ui.login.signInButton(page);
    await expect(button).toBeVisible();
    const loginCall = page.waitForRequest((r) => r.url() === `${HUB_API}/auth/login` && r.method() === 'POST');
    await button.click();
    await loginCall;
    await expect(page).toHaveURL(/\/groupdashboard$/);
    await expect(ui.home.groupTitles(page).first()).toBeVisible();
  });

  test('sign out returns to the sign-in page', async ({ page }) => {
    await signIn(page);
    await expect(ui.home.groupTitles(page).first()).toBeVisible();
    const logoutCall = page.waitForRequest((r) => r.url() === `${HUB_API}/auth/logout`);
    await ui.shell.signOut(page);
    await logoutCall;
    await expect(ui.login.signInButton(page)).toBeVisible();
  });

  test('every mock group name is reachable after sign-in', async ({ page }) => {
    await signIn(page);
    for (const g of MOCK.groups) await expect(ui.home.groupTitles(page).filter({ hasText: g }).first()).toBeVisible();
  });
});
