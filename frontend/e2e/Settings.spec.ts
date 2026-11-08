import { test, expect, signIn } from './support/fixtures';

// Spec 059 US-01: one rail, two scopes, stable ids and browser-back-safe links.
test('settings rail groups sections and keeps the open one in the URL', async ({ page }) => {
  await signIn(page, '/settings');
  const rail = page.getByRole('navigation', { name: 'Settings sections' });

  await expect(rail.getByRole('heading')).toHaveCount(0);
  await expect(rail.getByText('This browser', { exact: true })).toBeVisible();
  await expect(rail.getByText('Hub', { exact: true })).toBeVisible();
  const hubGroup = rail.getByRole('list', { name: 'Hub' });
  await expect(hubGroup.getByRole('link')).toHaveText([
    'Device defaults',
    'Storage & retention',
    'Logger',
    'MCP server',
    'Assistant',
  ]);

  await expect(rail.getByRole('link', { name: 'Appearance' })).toHaveAttribute('aria-current', 'page');
  await expect(page).toHaveURL(/\/settings$/);
  await expect(page.getByRole('heading', { level: 2, name: 'Appearance', exact: true })).toHaveCount(1);

  await rail.getByRole('link', { name: 'Logger' }).click();
  await expect(page).toHaveURL(/section=logger$/);
  await expect(rail.getByRole('link', { name: 'Logger' })).toHaveAttribute('aria-current', 'page');
  await expect(page.getByRole('heading', { level: 1, name: 'Settings' })).toBeVisible();
});

test('each section is reachable by URL and exposes exactly one section heading', async ({ page }) => {
  const sections: [string, string][] = [
    ['appearance', 'Appearance'],
    ['device-defaults', 'Device defaults'],
    ['storage', 'Storage & retention'],
    ['logger', 'Logger'],
    ['mcp', 'MCP server'],
    ['assistant', 'Assistant'],
  ];
  for (const [id, heading] of sections) {
    await signIn(page, `/settings?section=${id}`);
    await expect(page.getByRole('heading', { level: 1, name: 'Settings' })).toBeVisible();
    await expect(page.getByRole('heading', { level: 2, name: heading, exact: true })).toHaveCount(1);
    await expect(page.getByRole('navigation', { name: 'Settings sections' }).getByRole('link', { name: heading })).toHaveAttribute('aria-current', 'page');
  }
});

test('retired and unknown section ids resolve instead of failing', async ({ page }) => {
  await signIn(page, '/settings?section=history');
  await expect(page.getByRole('heading', { level: 2, name: 'Storage & retention' })).toBeVisible();
  await page.goto('/settings?section=system');
  await expect(page.getByRole('heading', { level: 2, name: 'Storage & retention' })).toBeVisible();
  await page.goto('/settings?section=does-not-exist');
  await expect(page.getByRole('heading', { level: 2, name: 'Appearance' })).toBeVisible();
});

test('rail is keyboard operable and scrolls horizontally on a phone', async ({ page }) => {
  await signIn(page, '/settings');
  const rail = page.getByRole('navigation', { name: 'Settings sections' });
  const assistant = rail.getByRole('link', { name: 'Assistant' });
  await assistant.focus();
  await expect(assistant).toBeFocused();
  await assistant.press('Enter');
  await expect(page).toHaveURL(/section=assistant$/);
  await expect(assistant).toHaveAttribute('aria-current', 'page');

  await page.setViewportSize({ width: 390, height: 844 });
  const box = await rail.boundingBox();
  expect(box!.height).toBeLessThan(140);
  expect(await page.evaluate(() => document.documentElement.scrollWidth > innerWidth)).toBe(false);
});

test('no settings section nests a card inside another card', async ({ page }) => {
  for (const id of ['appearance', 'device-defaults', 'storage', 'logger', 'mcp', 'assistant']) {
    await signIn(page, `/settings?section=${id}`);
    await page.waitForLoadState('networkidle');
    const nested = await page.evaluate(() => document.querySelectorAll('.nh-card .nh-card').length);
    expect(nested, `nested cards in ${id}`).toBe(0);
  }
});
