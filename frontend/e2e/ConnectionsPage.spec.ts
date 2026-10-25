import { test, expect, HUB_API, AUTOMATION_HUB_API } from './support/fixtures';
import { ui } from './support/ui';
import type { Page } from '@playwright/test';

async function signInConnections(page: Page) {
  // Keep Node device fixtures unchanged; only Connections uses the real isolated Go API.
  await page.route(`${HUB_API}/integrations/**`, async (route) => {
    const response = await route.fetch({ url: route.request().url().replace(HUB_API, AUTOMATION_HUB_API) });
    await route.fulfill({ response });
  });
  const response = await page.request.post(`${AUTOMATION_HUB_API}/auth/login`);
  expect(response.ok()).toBeTruthy();
  const { url } = await response.json();
  await page.goto(url);
  await expect(page).not.toHaveURL(/auth=success/);
  expect((await page.request.post(`${AUTOMATION_HUB_API}/__test/connections/reset`)).status()).toBe(204);
  await page.goto('/connections');
  await expect(page.getByRole('heading', { name: 'Connections', exact: true })).toBeVisible();
}
async function addConnection(page: Page) {
  await page.getByRole('button', { name: 'Add connection', exact: true }).click();
  await page.getByRole('textbox', { name: 'Postcode', exact: true }).fill('XX1 1XX');
  await page.getByRole('button', { name: 'Find address' }).click();
  await page
    .getByRole('combobox', { name: 'Choose your address' })
    .selectOption({ label: '1 Synthetic Street, Test Town' });
  await page.getByRole('button', { name: 'Save connection' }).click();
  await expect(page.getByRole('heading', { name: 'Paper and card', exact: true })).toBeVisible();
}
test('manual setup, results and settings stay inside Connections', async ({ page }, testInfo) => {
  await signInConnections(page);
  await addConnection(page);
  await page.screenshot({ path: `/private/tmp/nodeherder-032-results-${testInfo.project.name}.png`, fullPage: true });
  await expect(page.getByText(/Refresh is manual/)).toBeVisible();
  const persisted = await page.evaluate(() => JSON.stringify(localStorage));
  expect(persisted).not.toContain('XX1');
  expect(persisted).not.toContain('p:synthetic');
  await page.getByRole('button', { name: 'Refresh', exact: true }).click();
  await expect(page.getByRole('heading', { name: 'Paper and card', exact: true })).toBeVisible();
  await page.goto('/groupdashboard');
  await expect(page.locator('.nh-tile')).toHaveCount(18);
  await expect(page.getByRole('heading', { name: 'Paper and card' })).toHaveCount(0);
  await expect(page.getByRole('button', { name: /manage addons|manage connections/i })).toHaveCount(0);
  await page.goto('/devicedashboard');
  await expect(page.locator('main').getByRole('link', { name: 'Connections', exact: true })).toHaveCount(0);
  await page.goto('/overview');
  await expect(page.getByRole('heading', { name: 'Bin collections', exact: true })).toHaveCount(0);
  await page.goto('/panel');
  await expect(page.getByRole('heading', { name: 'Paper and card' })).toHaveCount(0);
});
test('Form and YAML repair share a draft across Help and work without horizontal overflow', async ({
  page,
}, testInfo) => {
  await signInConnections(page);
  await addConnection(page);
  await page.getByRole('link', { name: 'Configure operations', exact: true }).click();
  await expect(page.getByRole('textbox', { name: 'Display name', exact: true })).toBeVisible();
  await page.screenshot({ path: `/private/tmp/nodeherder-032-editor-${testInfo.project.name}.png`, fullPage: true });
  await page.getByRole('textbox', { name: 'Display name', exact: true }).fill('Waste dates provider');
  await page.getByRole('textbox', { name: 'Display name', exact: true }).press('Tab');
  await expect(page.getByRole('textbox', { name: 'Provider URL', exact: true })).toBeFocused();
  await page.getByRole('link', { name: 'Help', exact: true }).click();
  await page.getByRole('searchbox', { name: 'Find help' }).fill('schema');
  await expect(page.getByRole('heading', { name: 'Map the response fields' })).toBeVisible();
  // Deliberately slow instance loading: focus must be restored after the disabled fieldset reopens.
  await page.route('**/api/integrations/instances', async (route) => {
    await new Promise((resolve) => setTimeout(resolve, 200));
    await route.fallback();
  });
  await page.getByRole('link', { name: 'Configure operations', exact: true }).click();
  await expect(page.getByRole('textbox', { name: 'Display name', exact: true })).toHaveValue('Waste dates provider');
  await expect(page.getByRole('textbox', { name: 'Provider URL', exact: true })).toBeFocused();
  await page.getByRole('button', { name: 'YAML', exact: true }).click();
  const yaml = page.getByRole('textbox', { name: 'Provider YAML' });
  await expect(yaml).toHaveValue(/Waste dates provider/);
  const source = await yaml.inputValue();
  await yaml.fill(source.replace('items: /collections', 'items: /bins'));
  await page.getByRole('button', { name: 'Save changes', exact: true }).click();
  await expect(page.getByText('Saved on the host and applied. Existing collection dates were kept.')).toBeVisible();
  expect(await page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth)).toBe(true);
  await ui.shell.openNav(page);
  await expect(ui.shell.nav(page).getByRole('link', { name: 'Connections', exact: true })).toHaveAttribute(
    'aria-current',
    'page'
  );
});
test('signed-out deep links require sign in', async ({ page }) => {
  await page.goto('/connections');
  await expect(page.getByRole('button', { name: /Sign in with Google/i })).toBeVisible();
  await expect(page.getByRole('button', { name: 'Add connection', exact: true })).toHaveCount(0);
});

test('invalid YAML and a feature network failure keep the editable draft', async ({ page }) => {
  await signInConnections(page);
  await page.getByRole('link', { name: 'Configure operations', exact: true }).click();
  await page.getByRole('button', { name: 'YAML', exact: true }).click();
  const editor = page.getByRole('textbox', { name: 'Provider YAML' });
  const valid = await editor.inputValue();
  await editor.fill(valid + '\nunknown_field: true\n');
  await page.getByRole('button', { name: 'Validate', exact: true }).click();
  await expect(page.getByRole('alert')).toContainText('Check the provider configuration');
  await expect(editor).toHaveValue(valid + '\nunknown_field: true\n');
  await editor.fill(valid);
  await page.route('**/api/integrations/**', (route) => route.abort());
  await page.getByRole('button', { name: 'Validate', exact: true }).click();
  await expect(page.getByRole('alert')).toContainText('NodeHerder could not be reached');
  await expect(editor).toHaveValue(valid);
});

test('connection settings preserve saved results when renamed or disabled', async ({ page }) => {
  await signInConnections(page);
  await addConnection(page);
  await page.getByRole('link', { name: 'Setup', exact: true }).click();
  await page.getByRole('textbox', { name: 'Connection name', exact: true }).fill('My collection dates');
  await page.getByRole('button', { name: 'Save name', exact: true }).click();
  await expect(page.getByRole('heading', { name: 'My collection dates', exact: true })).toBeVisible();
  await page.getByRole('button', { name: 'Disable connection', exact: true }).click();
  await expect(page.getByRole('button', { name: 'Enable connection', exact: true })).toBeVisible();
  await expect(page.getByRole('button', { name: 'Refresh', exact: true })).toBeDisabled();
  await page.getByRole('link', { name: 'Results', exact: true }).click();
  await expect(page.getByRole('heading', { name: 'Paper and card', exact: true })).toBeVisible();
});

test('an expired login during catalog loading returns to sign in', async ({ page }) => {
  await signInConnections(page);
  await page.context().clearCookies();
  await page.route('**/api/integrations/**', (route) =>
    route.fulfill({
      status: 401,
      contentType: 'application/json',
      body: JSON.stringify({ data: null, error: { code: 'unauthorized', message: 'Expired' } }),
    })
  );
  await page.reload();
  await expect(page.getByRole('button', { name: /Sign in with Google/i })).toBeVisible();
  await expect(page.getByRole('button', { name: 'Add connection', exact: true })).toHaveCount(0);
});

test('an expired editor session discards private drafts without a leave confirmation', async ({ page }) => {
  await signInConnections(page);
  await page.getByRole('link', { name: 'Configure operations', exact: true }).click();
  await page.getByRole('textbox', { name: 'Display name', exact: true }).fill('Unsaved editor draft');
  // Expire the session in both the real feature API and the separate shell auth fixture.
  await page.context().clearCookies();
  await page.route('**/api/integrations/**', (route) =>
    route.fulfill({
      status: 401,
      contentType: 'application/json',
      body: JSON.stringify({ data: null, error: { code: 'unauthorized', message: 'Expired' } }),
    })
  );
  await page.getByRole('button', { name: 'Validate', exact: true }).click();
  await expect(page.getByRole('button', { name: /Sign in with Google/i })).toBeVisible();
  await expect(page.getByRole('dialog')).toHaveCount(0);
});
