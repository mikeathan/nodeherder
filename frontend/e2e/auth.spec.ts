import { test, expect, signIn, HUB_API } from './support/fixtures';
import { ui } from './support/ui';
import { MOCK } from './support/mock';

// Spec 007 US-08 (AC-25, AC-26): OAuth redirect sign-in must keep working.
test.describe('sign in', () => {
  for (const flow of ['OAuth', 'cookie']) {
    test(`${flow} handoff keeps one standalone login until the destination is ready`, async ({ page }, testInfo) => {
      let release!: () => void;
      const gate = new Promise<void>((resolve) => { release = resolve; });
      let destinationRequested = false;
      let modeChecks = 0;
      page.on('request', (request) => {
        if (request.url() === `${HUB_API}/auth/mode`) modeChecks += 1;
      });
      await page.addInitScript(() => {
        const logins = new Set<Element>();
        const observer = new MutationObserver(() => {
          document.querySelectorAll('.nh-login').forEach((login) => logins.add(login));
          document.documentElement.dataset.loginMounts = String(logins.size);
        });
        observer.observe(document, { childList: true, subtree: true });
      });
      await page.route('**/src/components/dashboards/GroupDashboard.vue', async (route) => {
        destinationRequested = true;
        await gate;
        await route.continue();
      });
      try {
        if (flow === 'OAuth') {
          await signIn(page);
        } else {
          const response = await page.request.post(`${HUB_API}/auth/login`);
          const { url } = await response.json();
          await page.request.get(url); // Set the real mock cookie without mounting the app.
          await page.goto('/');
        }
        await expect.poll(() => destinationRequested).toBe(true);
        await expect(page.locator('.nh-login')).toHaveCount(1);
        await expect(ui.shell.nav(page)).toHaveCount(0);
        await expect(ui.login.signInButton(page)).toHaveCount(0);
        await page.screenshot({ path: testInfo.outputPath('pending-handoff.png') });
        release();
        await expect(ui.home.groupTitles(page).first()).toBeVisible();
        await expect(page.locator('.nh-login')).toHaveCount(0);
        await expect(page.locator('html')).toHaveAttribute('data-login-mounts', '1');
        expect(modeChecks).toBe(1);
        await expect(page).toHaveURL(/\/groupdashboard$/);
      } finally {
        release();
      }
    });
  }

  test('local access has local wording and completes without Google', async ({ page }, testInfo) => {
    await page.route(`${HUB_API}/auth/mode`, (route) => route.fulfill({ json: { mode: 'local' } }));
    await page.route(`${HUB_API}/auth/login`, (route) => route.fulfill({ json: { url: `${HUB_API}/auth/offline/start` } }));
    await page.route(`${HUB_API}/auth/offline/start`, async (route) => {
      await page.context().addCookies([{ name: 'session', value: 'mock-local', url: HUB_API, httpOnly: true, sameSite: 'Lax', expires: Date.now() / 1000 + 86400 }]);
      await route.fulfill({ status: 302, headers: { location: 'http://localhost:4100/?auth=success' } });
    });
    await page.goto('/');
    const local = page.getByRole('button', { name: 'Continue on local network' });
    await expect(local).toBeVisible();
    await expect(ui.login.signInButton(page)).toHaveCount(0);
    await expect(page.getByText('No Google account is needed.', { exact: false })).toBeVisible();
    await page.screenshot({ path: testInfo.outputPath('local-login.png') });
    await local.focus();
    await expect(local).toBeFocused();
    await page.keyboard.press('Enter');
    await expect(page).toHaveURL(/\/groupdashboard$/);
    await expect(ui.home.groupTitles(page).first()).toBeVisible();
  });

  test('valid cookie restores sign-in without cached browser identity', async ({ page }) => {
    await signIn(page);
    await expect(ui.home.groupTitles(page).first()).toBeVisible();
    await page.evaluate(() => localStorage.removeItem('nodeherder_auth'));
    await page.goto('/');
    await expect(page).toHaveURL(/\/groupdashboard$/);
    await expect(ui.home.groupTitles(page).first()).toBeVisible();
    await expect(ui.login.signInButton(page)).toHaveCount(0);
  });

  test('initial session check waits before presenting sign-in', async ({ page }) => {
    let release!: () => void;
    const gate = new Promise<void>((resolve) => { release = resolve; });
    await page.route(`${HUB_API}/auth/me`, async (route) => {
      await gate;
      await route.fulfill({ status: 401, json: { error: 'missing token' } });
    });
    await page.goto('/');
    await expect(page.getByRole('status').filter({ hasText: 'Checking your session' })).toBeVisible();
    await expect(ui.login.signInButton(page)).toHaveCount(0);
    release();
    await expect(ui.login.signInButton(page)).toBeVisible();
  });

  test('mode failure offers retry rather than guessing Google', async ({ page }) => {
    await page.route(`${HUB_API}/auth/mode`, (route) => route.fulfill({ status: 503, json: {} }));
    await page.goto('/');
    const retry = page.getByRole('button', { name: 'Retry connection' });
    await expect(retry).toBeVisible();
    await expect(ui.login.signInButton(page)).toHaveCount(0);
    await page.unroute(`${HUB_API}/auth/mode`);
    await retry.click();
    await expect(ui.login.signInButton(page)).toBeVisible();
  });

  test('server outage preserves sign-in and retry recovers; 401 signs out', async ({ page }) => {
    await signIn(page);
    await expect(ui.home.groupTitles(page).first()).toBeVisible();
    await page.route(`${HUB_API}/auth/me`, (route) => route.fulfill({ status: 503, json: {} }));
    await page.reload();
    const retry = page.getByRole('button', { name: 'Retry session check' });
    await expect(retry).toBeVisible();
    await expect(ui.login.signInButton(page)).toHaveCount(0);
    await page.unroute(`${HUB_API}/auth/me`);
    await retry.click();
    await expect(retry).toHaveCount(0);
    await expect(ui.home.groupTitles(page).first()).toBeVisible();
    await page.route(`${HUB_API}/auth/me`, (route) => route.fulfill({ status: 401, json: {} }));
    await page.reload();
    await expect(ui.login.signInButton(page)).toBeVisible();
  });

  test('network recovery rechecks a retained session', async ({ page }) => {
    await signIn(page);
    await expect(ui.home.groupTitles(page).first()).toBeVisible();
    await page.route(`${HUB_API}/auth/me`, (route) => route.abort('failed'));
    await page.reload();
    await expect(page.getByRole('button', { name: 'Retry session check' })).toBeVisible();
    await page.unroute(`${HUB_API}/auth/me`);
    await page.evaluate(() => window.dispatchEvent(new Event('online')));
    await expect(page.getByRole('button', { name: 'Retry session check' })).toHaveCount(0);
    await expect(ui.home.groupTitles(page).first()).toBeVisible();
  });

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
