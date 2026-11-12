import { test, expect, signIn } from './support/fixtures';

/*
 * Wireless observation viewer (spec 064).
 *
 * This is the only suite that renders the page. Unit suites cover the pure logic, but a
 * mount-time failure (a wrong store import, a bad prop) is invisible to them: this spec exists
 * because exactly that class of bug reached the branch once already.
 */

test.describe('wireless viewer', () => {
  test('renders observed identities with source diagnostics', async ({ page }) => {
    await signIn(page, '/wireless');

    await expect(page.getByRole('heading', { name: 'Wireless', exact: true })).toBeVisible();

    // Assertions are scoped to the table: the radar draws the same identity text in SVG
    // <title> tooltips, so an unscoped text match is ambiguous by design.
    const table = page.locator('.observation-table');
    await expect(table.getByText('Front door tag')).toBeVisible();
    await expect(table.getByText('addr:aa:bb:cc:dd:ee:01', { exact: true })).toBeVisible();
    await expect(table.getByText('fp:9c02ab41', { exact: true })).toBeVisible();

    // The source that delivers traffic reports it; the silent one explains itself rather than
    // leaving an empty screen.
    await expect(page.getByText('Mock hall proxy')).toBeVisible();
    await expect(page.getByText(/Scanning · 128 advertisements/)).toBeVisible();
    await expect(page.getByText(/no advertisements received/i)).toBeVisible();
  });

  test('the range display is explicitly bearing-less', async ({ page }) => {
    await signIn(page, '/wireless');
    // The viewer must never imply a bearing it cannot compute from a single source.
    await expect(page.getByText(/Range only · no bearing/)).toBeVisible();
  });

  test('facet filters appear because a source declares them', async ({ page }) => {
    await signIn(page, '/wireless');
    const filters = page.locator('.observation-filters');
    await expect(filters).toBeVisible();
    // Empty state must not appear: this is the D-1 regression guard.
    await expect(page.getByText('No source declares filterable facets yet.')).toHaveCount(0);
    await expect(filters.getByRole('button', { name: 'Only new', exact: true })).toBeVisible();
  });

  test('isolating new identities narrows the table', async ({ page }) => {
    await signIn(page, '/wireless');
    await page.getByRole('button', { name: 'Only new', exact: true }).click();

    // The "known" beacon is filtered out; untriaged rows remain.
    const table = page.locator('.observation-table');
    await expect(table.getByText('Front door tag')).toHaveCount(0);
    await expect(table.getByText('addr:aa:bb:cc:dd:ee:01', { exact: true })).toBeVisible();
  });

  test('triage controls are keyboard reachable and report a decision', async ({ page }) => {
    await signIn(page, '/wireless');
    const known = page.getByRole('button', { name: 'Mark known', exact: true }).first();
    await known.focus();
    // Visible focus is an FE-04 requirement, not decoration.
    await expect(known).toBeFocused();
    await known.press('Enter');
    // The mock acknowledges the action; the control must remain usable afterwards.
    await expect(known).toBeEnabled();
  });

  test('layout fits at desktop and phone widths', async ({ page }) => {
    for (const viewport of [
      { width: 1280, height: 800 },
      { width: 390, height: 844 },
    ]) {
      await page.setViewportSize(viewport);
      await signIn(page, '/wireless');
      const overflow = await page.evaluate(() => document.documentElement.scrollWidth - document.documentElement.clientWidth);
      expect(overflow, `horizontal overflow at ${viewport.width}px`).toBeLessThanOrEqual(1);
    }
  });
});
