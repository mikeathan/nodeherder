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
  await page.goto('/integrations');
  await expect(page.getByRole('heading', { name: 'Integrations', exact: true })).toBeVisible();
}

test('Integrations naming keeps legacy list, detail and provider bookmarks usable', async ({ page }) => {
  await signInConnections(page);
  await page.goto('/connections');
  await expect(page.getByRole('heading', { name: 'Integrations', exact: true })).toBeVisible();
  await addConnection(page);
  const detail = new URL(page.url()).pathname;
  expect(detail).toMatch(/^\/integrations\//);
  await page.goto(detail.replace('/integrations/', '/connections/'));
  await expect(page.getByRole('heading', { name: 'Blue bin', exact: true })).toBeVisible();
  await page.getByRole('link', { name: 'Integrations', exact: true }).last().click();
  await expect(page).toHaveURL(/\/integrations$/);
  await page.goto('/connections/providers/waste_collection/whenisbins');
  await expect(page.getByRole('textbox', { name: 'Provider URL', exact: true })).toBeVisible();
  await page.getByRole('link', { name: 'Help', exact: true }).click();
  await expect(page).toHaveURL(/\/integrations\/providers\/waste_collection\/whenisbins\/help$/);
  await expect(page.getByRole('heading', { name: 'Integrations help', exact: true })).toBeVisible();
});
async function addConnection(page: Page, openDetails = true) {
  await page.getByRole('button', { name: 'Add integration', exact: true }).click();
  await page.getByRole('button', { name: 'Set up', exact: true }).click();
  await page.getByRole('textbox', { name: 'Postcode', exact: true }).fill('XX1 1XX');
  await page.getByRole('button', { name: 'Find address' }).click();
  await page
    .getByRole('combobox', { name: 'Choose your address' })
    .selectOption({ label: '1 Synthetic Street, Test Town' });
  await page.getByRole('button', { name: 'Save integration' }).click();
  await expect(page.getByRole('link', { name: 'Open Bin collections integration', exact: true })).toBeVisible();
  if (openDetails) {
    await page.getByRole('link', { name: 'Open Bin collections integration', exact: true }).click();
    await expect(page.getByRole('heading', { name: 'Blue bin', exact: true })).toBeVisible();
  }
}

test('selected section has clear space above the saved schedule', async ({ page }) => {
  await signInConnections(page);
  await addConnection(page);
  const tabs = page.getByRole('navigation', { name: 'Integration sections' });
  const gap = await tabs.evaluate((element) => {
    const schedule = document.querySelector('.waste-schedule-status')!;
    return schedule.getBoundingClientRect().top - element.getBoundingClientRect().bottom;
  });
  expect(gap).toBeGreaterThanOrEqual(12);
});

test('operations and Help retain header, tabs and cached draft geometry during delayed reads', async ({ page }) => {
  await signInConnections(page);
  await addConnection(page);
  await expect(page.locator('.connection-detail')).toHaveAttribute('aria-busy', 'false');
  const tabs = page.getByRole('navigation', { name: 'Integration sections' });
  const geometry = () =>
    page.locator('.connections .nh-page-head').evaluate((header) => {
      const bounds = header.getBoundingClientRect();
      const navigation = document.querySelector('.connection-tabs')!.getBoundingClientRect();
      return { y: bounds.y, height: bounds.height, tabsY: navigation.y, tabsHeight: navigation.height };
    });
  const detailGeometry = await geometry();
  let release!: () => void;
  let pending = true;
  const gate = new Promise<void>((resolve) => {
    release = resolve;
  });
  await page.route('**/api/integrations/definitions/waste_collection/whenisbins', async (route) => {
    if (pending && route.request().method() === 'GET') await gate;
    await route.fallback();
  });
  try {
    await tabs.getByRole('link', { name: 'Configure operations', exact: true }).click();
    await expect(tabs.getByRole('link', { name: 'Configure operations', exact: true })).toHaveAttribute(
      'aria-current',
      'page'
    );
    expect(await geometry()).toEqual(detailGeometry);
    const field = page.getByRole('textbox', { name: 'Display name', exact: true });
    await expect(field).toHaveValue('WhenIsBins');
    const pendingField = await field.boundingBox();
    pending = false;
    release();
    await expect(field).toBeEnabled();
    expect(await geometry()).toEqual(detailGeometry);
    expect(await field.boundingBox()).toEqual(pendingField);
    await field.fill('A draft kept while switching sections');
    await tabs.getByRole('link', { name: 'Help', exact: true }).click();
    await expect(tabs.getByRole('link', { name: 'Help', exact: true })).toHaveAttribute('aria-current', 'page');
    expect(await geometry()).toEqual(detailGeometry);
    await tabs.getByRole('link', { name: 'Configure operations', exact: true }).click();
    await expect(field).toHaveValue('A draft kept while switching sections');
    expect(await geometry()).toEqual(detailGeometry);
  } finally {
    pending = false;
    release();
  }
});

test('integration sections persist through operations, Help, reload and history', async ({ page }, testInfo) => {
  await signInConnections(page);
  await addConnection(page);
  const tabs = page.getByRole('navigation', { name: 'Integration sections' });
  await tabs.getByRole('link', { name: 'Configure operations', exact: true }).click();
  await expect(tabs).toBeVisible();
  await expect(tabs.getByRole('link', { name: 'Configure operations', exact: true })).toHaveAttribute(
    'aria-current',
    'page'
  );
  await expect(tabs.getByRole('link')).toHaveCount(4);
  await page.getByRole('textbox', { name: 'Display name', exact: true }).fill('My provider draft');
  await tabs.getByRole('link', { name: 'Help', exact: true }).click();
  await expect(tabs.getByRole('link', { name: 'Help', exact: true })).toHaveAttribute('aria-current', 'page');
  await expect(tabs.getByRole('link')).toHaveCount(4);
  await tabs.getByRole('link', { name: 'Configure operations', exact: true }).click();
  await expect(page.getByRole('textbox', { name: 'Display name', exact: true })).toHaveValue('My provider draft');
  // A data-section link retains the existing draft leave confirmation, including cancellation.
  await tabs.getByRole('link', { name: 'Setup', exact: true }).click();
  await expect(page.getByRole('dialog', { name: 'Leave provider editor?' })).toBeVisible();
  await page.getByRole('button', { name: 'Cancel', exact: true }).click();
  await expect(tabs.getByRole('link', { name: 'Configure operations', exact: true })).toHaveAttribute(
    'aria-current',
    'page'
  );
  await tabs.getByRole('link', { name: 'Help', exact: true }).click();
  await page.screenshot({
    path: `/private/tmp/nodeherder-032-sections-help-${testInfo.project.name}.png`,
    fullPage: true,
  });
  await page.reload();
  await expect(tabs.getByRole('link')).toHaveCount(4);
  await tabs.getByRole('link', { name: 'Configure operations', exact: true }).click();
  await expect(page.getByRole('textbox', { name: 'Display name', exact: true })).toHaveValue('WhenIsBins');
  await tabs.getByRole('link', { name: 'Setup', exact: true }).click();
  await expect(page.getByRole('textbox', { name: 'Integration name', exact: true })).toBeVisible();
  await expect(tabs.getByRole('link', { name: 'Setup', exact: true })).toHaveAttribute('aria-current', 'page');
  await page.reload();
  await expect(page.getByRole('textbox', { name: 'Integration name', exact: true })).toBeVisible();
  await tabs.getByRole('link', { name: 'Results', exact: true }).focus();
  await page.keyboard.press('Enter');
  await expect(page.getByRole('heading', { name: 'Blue bin', exact: true })).toBeVisible();
  await expect(tabs.locator('[aria-current="page"]')).toHaveCount(1);
  await page.goBack();
  await expect(tabs.getByRole('link', { name: 'Setup', exact: true })).toHaveAttribute('aria-current', 'page');
  expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true);
});

test('direct provider entry exposes shared provider sections without guessing an integration', async ({ page }) => {
  await signInConnections(page);
  await addConnection(page, false);
  await page.goto('/integrations/providers/waste_collection/whenisbins');
  const tabs = page.getByRole('navigation', { name: 'Integration sections' });
  await expect(tabs.getByRole('link')).toHaveCount(2);
  await expect(tabs.getByRole('link', { name: 'Results', exact: true })).toHaveCount(0);
  await tabs.getByRole('link', { name: 'Help', exact: true }).click();
  await expect(tabs.getByRole('link')).toHaveCount(2);
  await expect(tabs.getByRole('link', { name: 'Help', exact: true })).toHaveAttribute('aria-current', 'page');
});
test('manual setup, results and settings stay inside Integrations', async ({ page }, testInfo) => {
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
  await expect(page.locator('main').getByRole('link', { name: 'Integrations', exact: true })).toHaveCount(0);
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
  await expect(ui.shell.nav(page).getByRole('link', { name: 'Integrations', exact: true })).toHaveAttribute(
    'aria-current',
    'page'
  );
});
test('signed-out deep links require sign in', async ({ page }) => {
  await page.goto('/integrations');
  await expect(page.getByRole('button', { name: /Sign in with Google/i })).toBeVisible();
  await expect(page.getByRole('button', { name: 'Add integration', exact: true })).toHaveCount(0);
});

test('invalid YAML and a feature network failure keep the editable draft', async ({ page }) => {
  await signInConnections(page);
  await page.getByRole('button', { name: 'Add integration', exact: true }).click();
  await page.getByLabel('Bin collections', { exact: true })
    .getByRole('link', { name: 'Configure operations', exact: true }).click();
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

test('integration settings preserve saved results when renamed or disabled', async ({ page }) => {
  await signInConnections(page);
  await addConnection(page);
  await page.getByRole('link', { name: 'Setup', exact: true }).click();
  await page.getByRole('textbox', { name: 'Integration name', exact: true }).fill('My collection dates');
  await page.getByRole('button', { name: 'Save name', exact: true }).click();
  await expect(page.getByRole('heading', { name: 'My collection dates', exact: true })).toBeVisible();
  await page.getByRole('button', { name: 'Disable integration', exact: true }).click();
  await expect(page.getByRole('button', { name: 'Enable integration', exact: true })).toBeVisible();
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
  await expect(page.getByRole('button', { name: 'Add integration', exact: true })).toHaveCount(0);
});

test('an expired editor session discards private drafts without a leave confirmation', async ({ page }) => {
  await signInConnections(page);
  await page.getByRole('button', { name: 'Add integration', exact: true }).click();
  await page.getByLabel('Bin collections', { exact: true })
    .getByRole('link', { name: 'Configure operations', exact: true }).click();
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
    if (!resource || resource.contract_version !== 'waste_schedule/v1') throw new Error('Expected a bin schedule');
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

test('board shows saved results and shared bin artwork, with direct card navigation', async ({ page }, testInfo) => {
  await signInConnections(page);
  await addConnection(page, false);
  const widget = page.locator('.connection-widget');
  await expect(widget).toHaveCount(1);
  await expect(widget.getByText('Blue bin', { exact: true })).toBeVisible();
  await expect(widget.locator('.waste-bin-icon')).toHaveCount(1);
  await expect(widget.locator('header .waste-bin-icon')).toHaveCount(0);
  await expect(widget.locator('.connection-widget-badge .nh-ic')).toBeVisible();
  await expect(page.getByRole('link', { name: 'Details', exact: true })).toHaveCount(0);
  await expect(page.getByRole('combobox')).toHaveCount(0);
  await expect(page.getByRole('button', { name: 'Reload providers', exact: true })).toHaveCount(0);
  const bounds = await widget.boundingBox();
  expect(bounds!.width).toBeLessThan(500);
  await page.screenshot({ path: `/private/tmp/nodeherder-032-board-${testInfo.project.name}.png`, fullPage: true });
  await widget.getByRole('button', { name: 'Refresh Bin collections', exact: true }).click();
  await expect(page).toHaveURL(/\/integrations$/);
  const link = widget.getByRole('link', { name: 'Open Bin collections integration', exact: true });
  await link.focus();
  await link.press('Enter');
  await expect(page.getByRole('heading', { name: 'Blue bin', exact: true })).toBeVisible();
  await page.getByRole('link', { name: 'Integrations', exact: true }).last().click();
  await expect(link).toBeFocused();
  // The stretched route link covers the result content, not only its heading.
  const resultBounds = await widget.getByText('Blue bin', { exact: true }).boundingBox();
  await page.mouse.click(resultBounds!.x + resultBounds!.width / 2, resultBounds!.y + resultBounds!.height / 2);
  await expect(page).toHaveURL(/\/integrations\/[^/]+$/);
  await page.getByRole('link', { name: 'Setup', exact: true }).click();
  await expect(page.getByRole('textbox', { name: 'Integration name', exact: true })).toBeVisible();
});

test('hide/save survives reload and restore keeps the existing configured address', async ({ page }) => {
  await signInConnections(page);
  await addConnection(page, false);
  await page.getByRole('button', { name: 'Edit layout', exact: true }).click();
  await expect(page.getByRole('link', { name: 'Open Bin collections integration', exact: true })).toHaveCount(0);
  await page.getByRole('button', { name: 'Remove Bin collections from page', exact: true }).click();
  await page.getByRole('button', { name: 'Cancel', exact: true }).click();
  await expect(page.locator('.connection-widget')).toHaveCount(1);
  await page.getByRole('button', { name: 'Edit layout', exact: true }).click();
  await page.getByRole('button', { name: 'Remove Bin collections from page', exact: true }).click();
  await page.getByRole('button', { name: 'Save layout', exact: true }).click();
  await expect(page.locator('.connection-widget')).toHaveCount(0);
  await page.reload();
  await expect(page.locator('.connection-widget')).toHaveCount(0);
  await page.getByRole('button', { name: 'Add integration', exact: true }).click();
  await page.getByRole('button', { name: 'Show on page', exact: true }).click();
  await expect(page.getByRole('dialog')).toHaveCount(0);
  await expect(page.locator('.connection-widget').getByText('Blue bin', { exact: true })).toBeVisible();
  const persisted = await page.evaluate(() => JSON.stringify(localStorage));
  expect(persisted).not.toContain('XX1');
  expect(persisted).not.toContain('p:synthetic');
});

test('board layout conflict preserves the draft and narrow layouts fit', async ({ page }) => {
  await signInConnections(page);
  await addConnection(page, false);
  await page.route('**/api/integrations/board', async (route) => {
    if (route.request().method() !== 'PUT') return route.fallback();
    await route.fulfill({
      status: 409,
      contentType: 'application/json',
      body: JSON.stringify({
        data: null,
        error: { code: 'revision_conflict', message: 'This layout changed in another session. Reload before saving.' },
      }),
    });
  });
  await page.getByRole('button', { name: 'Edit layout', exact: true }).click();
  await page.getByRole('button', { name: 'Remove Bin collections from page', exact: true }).click();
  await page.getByRole('button', { name: 'Save layout', exact: true }).click();
  await expect(page.getByRole('alert')).toContainText('changed in another session');
  await expect(page.getByRole('button', { name: 'Save layout', exact: true })).toBeVisible();
  await expect(page.locator('.connection-widget')).toHaveCount(0);
  await page.getByRole('button', { name: 'Cancel', exact: true }).click();
  await expect(page.locator('.connection-widget')).toHaveCount(1);
  for (const width of [736, 390, 320]) {
    await page.setViewportSize({ width, height: 900 });
    expect(await page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth)).toBe(true);
    const widget = page.locator('.connection-widget');
    expect(await widget.evaluate((element) => element.scrollWidth <= element.clientWidth)).toBe(true);
  }
});

test('layout arrows and drag save the displayed order without exposing navigation', async ({ page }, testInfo) => {
  await signInConnections(page);
  await addConnection(page, false);
  // Isolated read fixtures exercise multi-widget presentation; Go companions cover atomic persistence.
  let entries = [
    { instance_id: 'board-a', hidden: false },
    { instance_id: 'board-b', hidden: false },
    { instance_id: 'board-c', hidden: false },
  ];
  await page.route('**/api/integrations/instances', async (route) => {
    const response = await route.fetch({ url: route.request().url().replace(HUB_API, AUTOMATION_HUB_API) });
    const payload = (await response.json()) as { data: ConnectionInstance[]; error: null };
    const original = payload.data[0];
    payload.data = entries.map((entry, index) => ({
      ...original,
      id: entry.instance_id,
      name: ['Bins A', 'Bins B', 'Bins C'][index],
    }));
    await route.fulfill({ response, json: payload });
  });
  await page.route('**/api/integrations/board', async (route) => {
    if (route.request().method() === 'PUT')
      entries = (route.request().postDataJSON() as { entries: typeof entries }).entries;
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ data: { etag: 'synthetic-layout', entries }, error: null }),
    });
  });
  await page.reload();
  await page.getByRole('button', { name: 'Edit layout', exact: true }).click();
  await page.getByRole('button', { name: 'Move Bins A later', exact: true }).click();
  await expect(page.locator('.connection-widget').first()).toHaveAttribute('data-instance-id', 'board-b');
  await expect(page.getByRole('link', { name: /Open Bins/ })).toHaveCount(0);
  if (testInfo.project.name === 'desktop') {
    const handle = page.getByRole('button', { name: 'Drag Bins C', exact: true });
    const from = await handle.boundingBox();
    const to = await page.locator('.connection-widget').first().boundingBox();
    await page.mouse.move(from!.x + from!.width / 2, from!.y + from!.height / 2);
    await page.mouse.down();
    await page.mouse.move(to!.x + to!.width / 2, to!.y + to!.height / 2, { steps: 20 });
    await page.mouse.up();
    // Dropping on the first card may insert before or after it, depending on the grid's
    // responsive column count. Both are real movement from the last position.
    await expect(page.locator('.connection-widget').last()).not.toHaveAttribute('data-instance-id', 'board-c');
  }
  await page.getByRole('button', { name: 'Save layout', exact: true }).click();
  await expect(page.getByRole('button', { name: 'Edit layout', exact: true })).toBeVisible();
  const displayed = await page
    .locator('.connection-widget')
    .evaluateAll((items) => items.map((item) => (item as HTMLElement).dataset.instanceId));
  expect(entries.map((entry) => entry.instance_id)).toEqual(displayed);
  if (testInfo.project.name === 'mobile') expect(entries[0].instance_id).toBe('board-b');
});

test('removing an integration requires confirmation and it can be configured again', async ({ page }) => {
  await signInConnections(page);
  await addConnection(page);
  await page.getByRole('link', { name: 'Setup', exact: true }).click();
  await page.getByRole('button', { name: 'Remove integration', exact: true }).click();
  const confirm = page.getByRole('dialog');
  await expect(confirm).toContainText('saved inputs, credentials and results');
  await confirm.getByRole('button', { name: 'Cancel', exact: true }).click();
  await expect(page.getByRole('textbox', { name: 'Integration name', exact: true })).toBeVisible();
  await page.getByRole('button', { name: 'Remove integration', exact: true }).click();
  await confirm.getByRole('button', { name: 'Remove integration', exact: true }).click();
  await expect(page).toHaveURL(/\/integrations$/);
  await expect(page.locator('.connection-widget')).toHaveCount(0);
  await addConnection(page, false);
  await expect(page.locator('.connection-widget')).toHaveCount(1);
});

test('three coloured bin summaries fit long names in both themes and restore board position', async ({
  page,
}, testInfo) => {
  await signInConnections(page);
  await addConnection(page, false);
  await page.route('**/api/integrations/instances', async (route) => {
    const response = await route.fetch({ url: route.request().url().replace(HUB_API, AUTOMATION_HUB_API) });
    const payload = (await response.json()) as { data: ConnectionInstance[]; error: null };
    const instance = payload.data[0];
    instance.name = 'Collection dates for the property with a longer integration name';
    const resource = instance.resources[0];
    if (!resource || resource.contract_version !== 'waste_schedule/v1') throw new Error('Expected a bin schedule');
    const original = resource.services[0];
    resource.services = ['black', 'blue', 'brown'].map((colour) => ({
      ...original,
      service_id: `service-${colour}`,
      bin_colour: colour,
      name: `${colour} service`,
    }));
    await route.fulfill({ response, json: payload });
  });
  await page.reload();
  const widget = page.locator('.connection-widget');
  await expect(widget.locator('.waste-bin-icon')).toHaveCount(3);
  await expect(widget.getByText('Black bin', { exact: true })).toBeVisible();
  await expect(widget.getByText('Brown bin', { exact: true })).toBeVisible();
  for (const width of [1280, 390, 320]) {
    await page.setViewportSize({ width, height: 844 });
    expect(await widget.evaluate((element) => element.scrollWidth <= element.clientWidth)).toBe(true);
  }
  await page.setViewportSize({ width: testInfo.project.name === 'desktop' ? 1280 : 390, height: 844 });
  for (const theme of ['dark', 'light']) {
    await page.getByRole('button', { name: `Switch to ${theme} mode`, exact: true }).click();
    await page.screenshot({
      path: `/private/tmp/nodeherder-032-board-three-${theme}-${testInfo.project.name}.png`,
      fullPage: true,
      animations: 'disabled',
    });
  }
  await page.setViewportSize({ width: 320, height: 420 });
  await expect(page.locator('.nh-app-nav')).toBeHidden();
  await page.locator('.nh-app-main').evaluate((element) => {
    element.scrollTop = 80;
  });
  const previous = await page.locator('.nh-app-main').evaluate((element) => element.scrollTop);
  expect(previous).toBeGreaterThan(0);
  await widget.getByRole('link').click();
  await expect(page.getByRole('heading', { name: 'Blue bin', exact: true })).toBeVisible();
  expect(await page.locator('.nh-app-main').evaluate((element) => element.scrollTop)).toBe(0);
  await page.getByRole('link', { name: 'Integrations', exact: true }).last().click();
  await expect(widget.getByRole('link')).toBeFocused();
  await expect.poll(() => page.locator('.nh-app-main').evaluate((element) => element.scrollTop)).toBe(previous);
});

test('cached navigation keeps the header and content stable while reads are pending', async ({ page }) => {
  await signInConnections(page);
  await addConnection(page, false);
  let releaseInstance!: () => void;
  let releaseProvider!: () => void;
  let releaseBoard!: () => void;
  const instanceGate = new Promise<void>((resolve) => {
    releaseInstance = resolve;
  });
  const providerGate = new Promise<void>((resolve) => {
    releaseProvider = resolve;
  });
  const boardGate = new Promise<void>((resolve) => {
    releaseBoard = resolve;
  });
  await page.route('**/api/integrations/instances/*', async (route) => {
    if (route.request().method() === 'GET') await instanceGate;
    await route.fallback();
  });
  await page.route('**/api/integrations/definitions/waste_collection/whenisbins', async (route) => {
    if (route.request().method() === 'GET') await providerGate;
    await route.fallback();
  });
  const bounds = () =>
    page.locator('.connection-detail .nh-page-head').evaluate((element) => {
      const rectangle = element.getBoundingClientRect();
      const tabs = document.querySelector('.connection-detail .connection-tabs')?.getBoundingClientRect();
      return { y: rectangle.y, height: rectangle.height, tabsY: tabs?.y };
    });
  try {
    await page.getByRole('link', { name: 'Open Bin collections integration', exact: true }).click();
    // The real saved identity and data are already available from the board; no placeholder flash.
    await expect(page.locator('.connection-detail h1')).toHaveText('Bin collections');
    await expect(page.locator('.connection-detail .nh-sub')).toHaveText('XX1 1XX');
    await expect(page.locator('.connection-detail')).toHaveAttribute('aria-busy', 'true');
    const pendingInstance = await bounds();
    await page.getByRole('link', { name: 'Setup', exact: true }).click();
    await page.getByRole('textbox', { name: 'Integration name', exact: true }).fill('My collection dates');
    releaseInstance();
    await expect(page.getByRole('link', { name: 'Setup', exact: true })).toBeVisible();
    const pendingProvider = await bounds();
    expect(pendingProvider).toEqual(pendingInstance);
    await expect(page.getByRole('textbox', { name: 'Integration name', exact: true })).toHaveValue(
      'My collection dates'
    );
    releaseProvider();
    await expect(page.locator('.connection-detail')).toHaveAttribute('aria-busy', 'false');
    await expect(page.getByRole('textbox', { name: 'Integration name', exact: true })).toHaveValue(
      'My collection dates'
    );
    expect(await bounds()).toEqual(pendingInstance);
    await page.route('**/api/integrations/board', async (route) => {
      if (route.request().method() === 'GET') await boardGate;
      await route.fallback();
    });
    await page.getByRole('link', { name: 'Integrations', exact: true }).last().click();
    await expect(page.locator('.connections-board-page')).toHaveAttribute('aria-busy', 'true');
    const pendingCard = await page.locator('.connection-widget').boundingBox();
    const pendingHeader = await page.locator('.connections-board-page .nh-page-head').boundingBox();
    releaseBoard();
    await expect(page.locator('.connections-board-page')).toHaveAttribute('aria-busy', 'false');
    expect(await page.locator('.connection-widget').boundingBox()).toEqual(pendingCard);
    expect(await page.locator('.connections-board-page .nh-page-head').boundingBox()).toEqual(pendingHeader);
  } finally {
    releaseInstance();
    releaseProvider();
    releaseBoard();
  }
});

test('a missing cached integration never reuses another detail or clears the lookup error', async ({ page }) => {
  await signInConnections(page);
  await addConnection(page);
  await page.route('**/api/integrations/instances', async (route) => {
    const response = await route.fetch({ url: route.request().url().replace(HUB_API, AUTOMATION_HUB_API) });
    const payload = (await response.json()) as { data: ConnectionInstance[]; error: null };
    payload.data.push({
      ...payload.data[0],
      id: 'missing-connection',
      name: 'Different property',
      postcode: 'YY2 2YY',
      resources: [],
    });
    await route.fulfill({ response, json: payload });
  });
  await page.getByRole('link', { name: 'Integrations', exact: true }).last().click();
  await expect(page.getByRole('link', { name: 'Open Different property integration', exact: true })).toBeVisible();
  let release!: () => void;
  const gate = new Promise<void>((resolve) => {
    release = resolve;
  });
  let providerReads = 0;
  await page.route('**/api/integrations/definitions/waste_collection/whenisbins', async (route) => {
    providerReads++;
    await route.fallback();
  });
  await page.route('**/api/integrations/instances/missing-connection', async (route) => {
    await gate;
    await route.fulfill({
      status: 404,
      contentType: 'application/json',
      body: JSON.stringify({ data: null, error: { code: 'not_found', message: 'This integration no longer exists.' } }),
    });
  });
  try {
    await page.getByRole('link', { name: 'Open Different property integration', exact: true }).click();
    await expect(page.locator('.connection-detail h1')).toHaveText('Different property');
    await expect(page.locator('.connection-detail .nh-sub')).toHaveText('YY2 2YY');
    await expect(page.getByRole('heading', { name: 'Blue bin', exact: true })).toHaveCount(0);
    release();
    await expect(page.getByRole('alert')).toContainText('no longer exists');
    await expect(page.locator('.connection-detail')).toHaveAttribute('aria-busy', 'false');
    await expect(page.locator('.connection-detail h1')).toHaveText('Integration');
    // The stable header keeps a hidden empty subtitle line, never the deleted property's value.
    await expect(page.locator('.connection-detail .nh-sub')).toHaveText('');
    await expect(page.locator('.connection-detail .nh-sub [aria-hidden="true"]')).toHaveCount(1);
    expect(providerReads).toBe(0);
  } finally {
    release();
  }
});

test('widget and full results show bins in upcoming collection-date order', async ({ page }) => {
  await signInConnections(page);
  await addConnection(page, false);
  await page.route(/\/api\/integrations\/instances(?:\/[^/?]+)?$/, async (route) => {
    if (route.request().method() !== 'GET') return route.fallback();
    const response = await route.fetch({ url: route.request().url().replace(HUB_API, AUTOMATION_HUB_API) });
    const payload = (await response.json()) as { data: ConnectionInstance | ConnectionInstance[]; error: null };
    for (const item of Array.isArray(payload.data) ? payload.data : [payload.data]) {
      const resource = item.resources[0];
      if (!resource || resource.contract_version !== 'waste_schedule/v1') continue;
      resource.evaluated_at = '2026-10-26T12:00:00Z';
      const original = resource.services[0];
      resource.services = [
        { ...original, service_id: 'service-blue', bin_colour: 'blue', dates: ['2026-11-17'] },
        { ...original, service_id: 'service-black', bin_colour: 'black', dates: ['2026-10-27'] },
        { ...original, service_id: 'service-brown', bin_colour: 'brown', dates: ['2026-11-03'] },
      ];
    }
    await route.fulfill({ response, json: payload });
  });
  await page.reload();
  const widget = page.locator('.connection-widget');
  await expect(widget.locator('.connection-widget-service > span:last-child')).toHaveText([
    'Black bin',
    'Brown bin',
    'Blue bin',
  ]);
  await expect(widget.locator('.connection-widget-date')).toContainText(['27 Oct', '3 Nov', '17 Nov']);
  await widget.getByRole('link').click();
  await expect(page.locator('.waste-collection-card .nh-card-title')).toHaveText([
    'Black bin',
    'Brown bin',
    'Blue bin',
  ]);
  await expect(page.locator('.waste-collection-card .connection-date')).toContainText([
    '27 October',
    '3 November',
    '17 November',
  ]);
});

test('weather details retain five-day mobile pages and desktop scrolling', async ({ page }) => {
  await signInConnections(page);
  await page.getByRole('button', { name: 'Add integration', exact: true }).click();
  await page.getByRole('combobox', { name: 'Weather provider' }).selectOption('weather/open_meteo');
  await page.getByRole('button', { name: 'Set up weather', exact: true }).click();
  await expect(page.getByRole('textbox', { name: 'API key / token' })).toHaveCount(0);
  await page.getByRole('textbox', { name: 'Town or postcode' }).fill('Bristol');
  await page.getByRole('button', { name: 'Find location' }).click();
  await page.getByRole('combobox', { name: 'Choose your location' }).selectOption({ label: 'Bristol' });
  await page.getByRole('button', { name: 'Save integration', exact: true }).click();
  await page.getByRole('link', { name: 'Open Weather integration', exact: true }).click();
  await expect(page.getByRole('heading', { name: 'Weather', exact: true })).toBeVisible();
  await expect(page.getByRole('heading', { name: '15-day forecast' })).toBeVisible();
  for (const width of [320, 375, 768, 1440]) {
    await page.setViewportSize({ width, height: 900 });
    const days = page.getByRole('region', { name: 'Daily forecast' });
    if (width <= 700) {
      await expect(days.locator('article:visible')).toHaveCount(5);
      await page.getByRole('button', { name: 'Next →', exact: true }).focus();
      await page.keyboard.press('Enter');
      await expect(page.getByRole('status').filter({ hasText: 'Days 6–10 of 15' })).toBeVisible();
      await page.getByRole('button', { name: '← Previous', exact: true }).click();
    } else await expect(days.locator('article:visible')).toHaveCount(15);
    expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBeTruthy();
    await page.screenshot({ path: `/private/tmp/nodeherder-weather-${width}.png`, fullPage: true });
  }
});

for (const provider of ['met_office', 'weatherapi'])
  test(`weather ${provider} setup offers a provider key link`, async ({ page }) => {
    await signInConnections(page);
    await page.getByRole('button', { name: 'Add integration', exact: true }).click();
    await page.getByRole('combobox', { name: 'Weather provider' }).selectOption(`weather/${provider}`);
    await page.getByRole('button', { name: 'Set up weather', exact: true }).click();
    await page.getByText('Provider credential').click();
    await expect(page.getByRole('textbox', { name: 'API key / token' })).toBeVisible();
    await expect(page.getByRole('link', { name: 'Get an API key from the provider' })).toHaveAttribute(
      'href',
      provider === 'met_office'
        ? 'https://datahub.metoffice.gov.uk/docs/getting-started'
        : 'https://www.weatherapi.com/signup.aspx'
    );
  });
