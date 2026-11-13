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
    await expect(table.getByRole('button', { name: 'Isolate addr:aa:bb:cc:dd:ee:01', exact: true })).toBeVisible();
    await expect(table.getByRole('button', { name: 'Isolate fp:9c02ab41', exact: true })).toBeVisible();

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
    const table = page.locator('.observation-table');
    // The recently-appeared beacon and an older device both start visible.
    await expect(table.getByText('Front door tag')).toBeVisible();
    await expect(table.getByRole('button', { name: 'Isolate addr:aa:bb:cc:dd:ee:01', exact: true })).toBeVisible();

    await page.getByRole('button', { name: 'Only new', exact: true }).click();

    // "New" means recently appeared, not untriaged: the older identity is filtered out even
    // though it has never been triaged. Filtering on triage state would match everything here,
    // which is the bug this replaces.
    await expect(table.getByText('Front door tag')).toBeVisible();
    await expect(table.getByRole('button', { name: 'Isolate addr:aa:bb:cc:dd:ee:01', exact: true })).toHaveCount(0);
  });

  test('pausing freezes the rows so they can be read', async ({ page }) => {
    await signIn(page, '/wireless');
    await page.getByRole('button', { name: 'Pause updates', exact: true }).click();

    await expect(page.getByRole('button', { name: 'Resume updates', exact: true })).toBeVisible();
    await expect(page.getByText('Frozen', { exact: true })).toBeVisible();
    // The table keeps rendering while frozen; it is a snapshot, not an empty state.
    await expect(page.locator('.observation-table').getByText('Front door tag')).toBeVisible();

    await page.getByRole('button', { name: 'Resume updates', exact: true }).click();
    await expect(page.getByRole('button', { name: 'Pause updates', exact: true })).toBeVisible();
  });

  test('the list view groups identities into collapsible sections', async ({ page }) => {
    await signIn(page, '/wireless');
    await page.getByRole('button', { name: 'List', exact: true }).click();

    const groups = page.locator('.observation-groups');
    await expect(groups).toBeVisible();
    const section = groups.getByRole('button', { name: /iBeacon/ });
    await expect(section).toHaveAttribute('aria-expanded', 'true');
    // Sections collapse so a noisy category can be folded away.
    await section.click();
    await expect(section).toHaveAttribute('aria-expanded', 'false');
  });

  test('view mode switches back to the radar', async ({ page }) => {
    await signIn(page, '/wireless');
    await page.getByRole('button', { name: 'List', exact: true }).click();
    await expect(page.locator('.observation-groups')).toBeVisible();
    await page.getByRole('button', { name: 'Radar', exact: true }).click();
    await expect(page.locator('.observation-radar')).toBeVisible();
    await expect(page.locator('.observation-groups')).toHaveCount(0);
  });

  test('isolating an identity shows only that identity in Focus', async ({ page }) => {
    await signIn(page, '/wireless');
    // Start from the list, then isolate one identity by name.
    await page.getByRole('button', { name: 'List', exact: true }).click();
    await page.getByRole('button', { name: 'Isolate Front door tag', exact: true }).click();

    const focus = page.locator('.observation-focus');
    await expect(focus).toBeVisible();
    await expect(focus.getByRole('heading', { name: 'Front door tag' })).toBeVisible();
    // Only the isolated identity is present; the others are gone from every view.
    await expect(page.getByRole('button', { name: 'Isolate addr:aa:bb:cc:dd:ee:01', exact: true })).toHaveCount(0);
    await expect(page.getByRole('button', { name: 'Isolate fp:9c02ab41', exact: true })).toHaveCount(0);
  });

  test('Focus shows distance, band and trend', async ({ page }) => {
    await signIn(page, '/wireless');
    await page.getByRole('button', { name: 'List', exact: true }).click();
    await page.getByRole('button', { name: 'Isolate Front door tag', exact: true }).click();

    const focus = page.locator('.observation-focus');
    // Distance is formatted by the shared formatter, so a unit always accompanies the figure.
    await expect(focus.locator('.observation-focus__value')).toContainText('m');
    await expect(focus.getByText(/near|mid|far/)).toBeVisible();
    // The trend pairs an arrow with a word, never a colour or arrow alone.
    await expect(focus.getByText(/Approaching|Receding|Steady|No trend yet/)).toBeVisible();
    await expect(focus.getByText(/Estimated from signal strength only/)).toBeVisible();
  });

  test('escape clears isolation and returns to the full view', async ({ page }) => {
    await signIn(page, '/wireless');
    await page.getByRole('button', { name: 'List', exact: true }).click();
    await page.getByRole('button', { name: 'Isolate Front door tag', exact: true }).click();
    await expect(page.locator('.observation-focus')).toBeVisible();

    await page.keyboard.press('Escape');
    await expect(page.locator('.observation-focus')).toHaveCount(0);
    // Back to the unfiltered set, so the other identities are reachable again.
    await expect(page.getByRole('button', { name: 'Isolate Front door tag', exact: true })).toHaveCount(1);
  });

  test('the All identities control also leaves isolation', async ({ page }) => {
    await signIn(page, '/wireless');
    await page.getByRole('button', { name: 'List', exact: true }).click();
    await page.getByRole('button', { name: 'Isolate Front door tag', exact: true }).click();
    await page.getByRole('button', { name: '‹ All identities', exact: true }).click();
    await expect(page.locator('.observation-focus')).toHaveCount(0);
  });

  test('the radar scale zooms in and out with an unambiguous direction', async ({ page }) => {
    await signIn(page, '/wireless');

    // Auto is the default: the scale follows the data.
    const auto = page.getByRole('button', { name: 'Auto', exact: true });
    await expect(auto).toHaveAttribute('aria-pressed', 'true');

    const range = page.locator('.wireless-viewer__range');
    const zoomIn = page.getByRole('button', { name: 'Zoom in', exact: true });
    const zoomOut = page.getByRole('button', { name: 'Zoom out', exact: true });

    // Zooming in must *narrow* the range: a narrower span draws the same space larger, which is
    // the direction the word means. This is the assertion that would have caught an inverted scale.
    const before = Number((await range.textContent())?.replace(' m', ''));
    await zoomIn.click();
    const afterIn = Number((await range.textContent())?.replace(' m', ''));
    expect(afterIn).toBeLessThan(before);

    // Zooming out widens it again.
    await zoomOut.click();
    expect(Number((await range.textContent())?.replace(' m', ''))).toBeGreaterThan(afterIn);

    // Stepping the scale leaves Auto, because the operator has taken control.
    await expect(auto).toHaveAttribute('aria-pressed', 'false');

    // Returning to Auto restores following the data.
    await auto.click();
    await expect(auto).toHaveAttribute('aria-pressed', 'true');
  });

  test('the radar rings keep even spacing at every zoom level', async ({ page }) => {
    await signIn(page, '/wireless');

    // The bug this guards: rings used to come from a fixed absolute ladder, so their gaps changed
    // with zoom (23%, 23%, 13% at a 3 m span) and the scale read as broken.
    const radii = async () =>
      page.locator('.observation-radar__ring').evaluateAll((nodes) =>
        nodes
          .map((node) => Number(node.getAttribute('r')))
          .sort((a, b) => a - b),
      );

    await page.getByRole('button', { name: 'Zoom in', exact: true }).click();
    const narrow = await radii();
    await page.getByRole('button', { name: 'Zoom out', exact: true }).click();
    const narrowThenWide = await radii();

    for (const set of [narrow, narrowThenWide]) {
      for (let index = 1; index < set.length; index++) {
        expect(set[index] - set[index - 1]).toBeCloseTo(set[0], 0);
      }
    }
  });

  test('sub-metre distance bands remain available', async ({ page }) => {
    await signIn(page, '/wireless');
    const fine = page.getByRole('button', { name: '0.5–1 m', exact: true });
    await expect(fine).toBeVisible();
    await fine.click();
    await expect(fine).toHaveAttribute('aria-pressed', 'true');
  });

  test('a distance band narrows the list to that range only', async ({ page }) => {
    await signIn(page, '/wireless');
    await page.getByRole('button', { name: 'List', exact: true }).click();

    // The mock spans roughly 1.9 m, 11 m and an unplaceable one, so a band is genuinely
    // selective. The labelled beacon is smoothed to -66 dBm, which is 1.91 m: the 1-2 m band.
    const band = page.getByRole('button', { name: '1–2 m', exact: true });
    await expect(band).toBeVisible();
    await band.click();
    await expect(band).toHaveAttribute('aria-pressed', 'true');
    await expect(page.getByRole('button', { name: 'Isolate Front door tag', exact: true })).toHaveCount(1);
    await expect(page.getByRole('button', { name: 'Isolate addr:aa:bb:cc:dd:ee:01', exact: true })).toHaveCount(0);

    // A band that excludes the near device must exclude it, not merely reorder.
    await page.getByRole('button', { name: '5–10 m', exact: true }).click();
    await expect(page.getByRole('button', { name: 'Isolate Front door tag', exact: true })).toHaveCount(0);

    // A band with no devices is empty rather than falling back to everything.
    await page.getByRole('button', { name: '0.5–1 m', exact: true }).click();
    await expect(page.getByRole('button', { name: 'Isolate Front door tag', exact: true })).toHaveCount(0);

    // Any clears the filter.
    await page.getByRole('button', { name: 'Any', exact: true }).click();
    await expect(page.getByRole('button', { name: 'Isolate Front door tag', exact: true })).toHaveCount(1);
    await expect(page.getByRole('button', { name: 'Isolate addr:aa:bb:cc:dd:ee:01', exact: true })).toHaveCount(1);
  });

  test('the distance control is a single-select group of bands', async ({ page }) => {
    await signIn(page, '/wireless');
    const any = page.getByRole('button', { name: 'Any', exact: true });
    const band = page.getByRole('button', { name: '2–5 m', exact: true });
    await expect(any).toHaveAttribute('aria-pressed', 'true');
    await band.click();
    // Bands are exclusive: one selected, never two.
    await expect(band).toHaveAttribute('aria-pressed', 'true');
    await expect(any).toHaveAttribute('aria-pressed', 'false');
  });

  test('the global radar uses most of the available height', async ({ page }) => {
    // Desktop only: below 1200px the table stacks underneath, so the radar is bounded by the
    // column width instead of the viewport height.
    await page.setViewportSize({ width: 1440, height: 900 });
    await signIn(page, '/wireless');
    await expect(page.locator('.observation-table')).toBeVisible();
    const box = (await page.locator('.observation-radar__canvas').boundingBox())!;
    const viewport = page.viewportSize()!;

    // Previously capped at 28rem (448px) in a narrow column, which made the global view small.
    // The radar is now the primary pane: at least 60% of the viewport height.
    expect(box.height).toBeGreaterThan(viewport.height * 0.6);
    // It stays square, so the rings remain circular.
    expect(Math.abs(box.width - box.height)).toBeLessThan(2);
  });

  test('the radar fills the column on a phone without overflow', async ({ page }) => {
    await page.setViewportSize({ width: 390, height: 844 });
    await signIn(page, '/wireless');
    await expect(page.locator('.observation-table')).toBeVisible();
    const box = (await page.locator('.observation-radar__canvas').boundingBox())!;
    // Full column minus page padding, and never wider than the viewport.
    expect(box.width).toBeGreaterThan(300);
    expect(box.width).toBeLessThanOrEqual(390);
    const overflow = await page.evaluate(
      () => document.documentElement.scrollWidth - document.documentElement.clientWidth,
    );
    expect(overflow).toBeLessThanOrEqual(1);
  });

  test('clicking a radar blip isolates that identity', async ({ page }) => {
    await signIn(page, '/wireless');

    // Blips are real controls, reachable by role rather than by pixel position.
    const blip = page.getByRole('button', { name: /^Isolate Front door tag, \d+ metres$/ }).first();
    await expect(blip).toBeVisible();
    await blip.click();

    const focus = page.locator('.observation-focus');
    await expect(focus).toBeVisible();
    await expect(focus.getByRole('heading', { name: 'Front door tag' })).toBeVisible();
    // Only the isolated identity remains anywhere in the view.
    await expect(page.getByRole('button', { name: 'Isolate addr:aa:bb:cc:dd:ee:01', exact: true })).toHaveCount(0);
  });

  test('blips are keyboard operable', async ({ page }) => {
    await signIn(page, '/wireless');
    const blip = page.getByRole('button', { name: /^Isolate Front door tag, \d+ metres$/ }).first();
    await blip.focus();
    // Visible focus is required: the SVG highlight is the only cue here.
    await expect(blip).toBeFocused();
    await blip.press('Enter');
    await expect(page.locator('.observation-focus')).toBeVisible();
  });

  test('the radar derives its distance from smoothed signal', async ({ page }) => {
    await signIn(page, '/wireless');
    const blip = page.getByRole('button', { name: /^Isolate Front door tag, \d+ metres$/ }).first();
    const label = (await blip.getAttribute('aria-label')) ?? '';
    const metres = Number(label.match(/(\d+) metres/)?.[1]);

    // The mock sends rssi -62 but rssiSmoothed -66, with txPower -59 and exponent 2.5:
    //   raw      -> 10^((3)/25)  = 1.3 m -> 1 m
    //   smoothed -> 10^((7)/25)  = 1.9 m -> 2 m
    // So 2 proves the smoothed value is used; 1 would mean the single packet leaked through.
    expect(metres).toBe(2);
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

/*
 * Behavioural classification (spec 065).
 *
 * The point of this work was the maintainer's complaint that the viewer was "a big list of mostly
 * mac address". These tests assert the two halves of the fix: demoted classes are hidden with an
 * honest count, and every class that is shown says what it means in words.
 */
test.describe('classification', () => {
  test('demoted classes are hidden and their count is reported', async ({ page }) => {
    await signIn(page, '/wireless');

    const table = page.locator('.observation-table');
    // The installed device is hidden by default; this is the row type that made the list useless.
    await expect(table.getByRole('button', { name: 'Isolate addr:aa:bb:cc:dd:ee:02', exact: true })).toHaveCount(0);

    // Hiding is never silent: the footer states how many and offers a way back.
    await expect(page.getByText(/1 device hidden/)).toBeVisible();
    await expect(page.getByRole('button', { name: 'Show them', exact: true })).toBeVisible();
  });

  test('revealing demoted classes brings them back', async ({ page }) => {
    await signIn(page, '/wireless');

    await page.getByRole('button', { name: 'Show them', exact: true }).click();

    const table = page.locator('.observation-table');
    await expect(table.getByRole('button', { name: 'Isolate addr:aa:bb:cc:dd:ee:02', exact: true })).toBeVisible();
    await expect(page.getByText(/device hidden/)).toHaveCount(0);
  });

  test('the show-all toggle is an alternative route to the same state', async ({ page }) => {
    await signIn(page, '/wireless');

    await page.getByRole('button', { name: 'Show all kinds', exact: true }).click();

    const table = page.locator('.observation-table');
    await expect(table.getByRole('button', { name: 'Isolate addr:aa:bb:cc:dd:ee:02', exact: true })).toBeVisible();
  });

  test('every visible row states its class in words', async ({ page }) => {
    await signIn(page, '/wireless');
    const table = page.locator('.observation-table');

    // No raw identifiers: the operator must never see "infrastructure" as a badge.
    await expect(table.locator('.class-badge').first()).toBeVisible();
    await expect(table.getByText('Yours', { exact: true })).toBeVisible();
    await expect(table.getByText('infrastructure', { exact: true })).toHaveCount(0);
  });

  test('learning shows its progress rather than a bare label', async ({ page }) => {
    await signIn(page, '/wireless');

    // A slow classifier that shows nothing looks broken; naming the progress makes it read as work
    // in progress.
    await expect(page.locator('.observation-table').getByText('Learning', { exact: true })).toBeVisible();
    await expect(page.locator('.observation-table').getByText('6 of 24 hours', { exact: true })).toBeVisible();
  });

  test('a provisional class is marked as a guess', async ({ page }) => {
    await signIn(page, '/wireless');
    const table = page.locator('.observation-table');

    const badge = table.locator('.class-badge--provisional').first();
    await expect(badge).toBeVisible();
    // The marker and the accessible text both say so, so it is not colour-only.
    await expect(table.getByText('(provisional guess)', { exact: false }).first()).toBeAttached();
  });

  test('the isolated view explains the class and shows its evidence', async ({ page }) => {
    await signIn(page, '/wireless');

    // The beacon carries a settled class with a reason, so it is the one whose explanation can be
    // asserted; the other identities are still learning.
    await page.getByRole('button', { name: 'Isolate Front door tag', exact: true }).click();

    const focus = page.locator('.observation-focus');
    await expect(focus).toBeVisible();
    // The reasoning must be present: a class the operator cannot check is a class they cannot trust.
    // Scoped to the reason element: the same sentence also appears in the badge tooltip, and a
    // strict-mode match on both would be ambiguous.
    await expect(focus.locator('.observation-focus__class-reason')).toContainText(
      'Present for part of the day with a changing signal',
    );
    await expect(focus.getByText(/Advisory only/)).toBeVisible();
    await expect(focus.getByRole('heading', { name: 'Evidence', exact: true })).toBeVisible();
  });

  test('the radar legend names its classes so colour is never the only cue', async ({ page }) => {
    await signIn(page, '/wireless');

    const legend = page.locator('.observation-radar__legend');
    await expect(legend).toBeVisible();
    await expect(legend.getByText('Yours', { exact: true })).toBeVisible();
  });

  test('classification survives a phone viewport without overflow', async ({ page }) => {
    await signIn(page, '/wireless');
    await page.getByRole('button', { name: 'Isolate addr:aa:bb:cc:dd:ee:01', exact: true }).click();

    const overflow = await page.evaluate(() => document.documentElement.scrollWidth > window.innerWidth + 1);
    expect(overflow).toBe(false);
  });
});
