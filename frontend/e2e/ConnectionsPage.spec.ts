import { test, expect, HUB_API, AUTOMATION_HUB_API } from './support/fixtures';
import { ui } from './support/ui';
import type { Page } from '@playwright/test';
import type { ConnectionInstance } from '../src/types/integrations.type';

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
  await expect(page.getByRole('heading', { name: 'Blue bin', exact: true })).toBeVisible();
}
test('manual setup, results and settings stay inside Connections', async ({ page }, testInfo) => {
  await signInConnections(page);
  await addConnection(page);
  await page.screenshot({ path: `/private/tmp/nodeherder-032-results-${testInfo.project.name}.png`, fullPage: true });
  await page.locator('.waste-schedule-status > summary').click();
  await expect(page.getByText(/Refresh is manual/)).toBeVisible();
  await page.locator('.waste-schedule-status > summary').click();
  const persisted = await page.evaluate(() => JSON.stringify(localStorage));
  expect(persisted).not.toContain('XX1');
  expect(persisted).not.toContain('p:synthetic');
  await page.getByRole('button', { name: 'Refresh', exact: true }).click();
  await expect(page.getByRole('heading', { name: 'Blue bin', exact: true })).toBeVisible();
  await page.goto('/groupdashboard');
  await expect(page.locator('.nh-tile')).toHaveCount(18);
  await expect(page.getByRole('heading', { name: 'Blue bin' })).toHaveCount(0);
  await expect(page.getByRole('button', { name: /manage addons|manage connections/i })).toHaveCount(0);
  await page.goto('/devicedashboard');
  await expect(page.locator('main').getByRole('link', { name: 'Connections', exact: true })).toHaveCount(0);
  await page.goto('/overview');
  await expect(page.getByRole('heading', { name: 'Bin collections', exact: true })).toHaveCount(0);
  await page.goto('/panel');
  await expect(page.getByRole('heading', { name: 'Blue bin' })).toHaveCount(0);
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
  await expect(page.getByRole('heading', { name: 'Blue bin', exact: true })).toBeVisible();
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

test('three collection headers and dates align with colour-matched bin illustrations', async ({ page }, testInfo) => {
  await signInConnections(page);
  await addConnection(page);
  // Change only the isolated read DTO: no extra remote lookup or household data is involved.
  await page.route('**/api/integrations/instances/*', async (route) => {
    const response = await route.fetch({ url: route.request().url().replace(HUB_API, AUTOMATION_HUB_API) });
    const payload = (await response.json()) as { data: ConnectionInstance; error: null };
    const resource = payload.data.resources[0];
    const original = resource.services[0];
    resource.services = [
      {
        ...original,
        service_id: '00000000-0000-4000-8000-000000000001',
        name: 'Black Bin - Non-recyclable waste',
        waste_type: 'refuse',
        bin_colour: 'black',
      },
      {
        ...original,
        service_id: '00000000-0000-4000-8000-000000000002',
        name: 'Blue Bin - Paper and card',
        bin_colour: 'blue',
      },
      {
        ...original,
        service_id: '00000000-0000-4000-8000-000000000003',
        name: 'Brown Bin - Glass bottles, jars, tins, cans, foil, plastic and cartons',
        waste_type: 'glass',
        bin_colour: 'brown',
      },
    ];
    await route.fulfill({ response, json: payload });
  });
  await page.reload();
  const cards = page.locator('.waste-collection-card');
  await expect(cards).toHaveCount(3);
  for (const [index, name] of ['Black bin', 'Blue bin', 'Brown bin'].entries())
    await expect(cards.nth(index).getByRole('heading', { name, exact: true })).toBeVisible();
  await expect(cards.getByText(/Limited dates/)).toHaveCount(0);
  for (const [index, colour] of ['#343e48', '#307fbd', '#95633e'].entries()) {
    await expect(cards.nth(index).locator('.waste-bin-icon')).toBeVisible();
    await expect(cards.nth(index).locator('.waste-bin-body')).toHaveAttribute('fill', colour);
    await expect(cards.nth(index).locator('.waste-bin-icon')).toHaveAttribute('aria-hidden', 'true');
  }
  const bounds = await cards.evaluateAll((elements) =>
    elements.map((card) => {
      const box = card.getBoundingClientRect();
      const header = card.querySelector('header')!.getBoundingClientRect();
      const date = card.querySelector('.connection-date')!.getBoundingClientRect();
      return {
        top: box.top,
        left: box.left,
        width: box.width,
        headerHeight: header.height,
        headerBottom: header.bottom,
        dateTop: date.top,
      };
    })
  );
  const status = page.locator('.waste-schedule-status');
  await expect(status).not.toHaveAttribute('open');
  expect((await status.boundingBox())!.height).toBeLessThanOrEqual(48);
  for (const box of bounds) {
    expect(box.headerHeight).toBeGreaterThanOrEqual(70);
    expect(box.headerHeight).toBeLessThanOrEqual(88);
  }
  await status.locator('summary').focus();
  await page.keyboard.press('Enter');
  await expect(page.getByText(/Refresh is manual/)).toBeVisible();
  await expect(
    status.getByText('Brown Bin - Glass bottles, jars, tins, cans, foil, plastic and cartons', { exact: true })
  ).toBeVisible();
  await page.keyboard.press('Enter');
  await expect(status).not.toHaveAttribute('open');
  if (testInfo.project.name === 'desktop') {
    for (const box of bounds.slice(1)) {
      expect(Math.abs(box.top - bounds[0].top)).toBeLessThan(1);
      expect(Math.abs(box.headerBottom - bounds[0].headerBottom)).toBeLessThan(1);
      expect(Math.abs(box.dateTop - bounds[0].dateTop)).toBeLessThan(1);
    }
  } else {
    for (const [index, box] of bounds.entries()) {
      expect(Math.abs(box.left - bounds[0].left)).toBeLessThan(1);
      expect(Math.abs(box.width - bounds[0].width)).toBeLessThan(1);
      if (index) expect(box.top).toBeGreaterThan(bounds[index - 1].top);
    }
  }
  expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true);
  await page.screenshot({
    path: `/private/tmp/nodeherder-032-colour-bins-${testInfo.project.name}.png`,
    fullPage: true,
  });
  await page.getByRole('button', { name: 'Switch to dark mode', exact: true }).click();
  await expect(page.locator('html')).toHaveAttribute('data-mode', 'dark');
  await page.screenshot({
    path: `/private/tmp/nodeherder-032-colour-bins-dark-${testInfo.project.name}.png`,
    fullPage: true,
  });
});
