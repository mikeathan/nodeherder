import { test, expect, signIn, resetMockState, HUB_API } from './support/fixtures';
import type { Page } from '@playwright/test';

/**
 * Open the filter fields on a phone. On desktop the fields are always visible and the disclosure
 * does not exist, so this is a no-op there.
 */
async function revealFilters(page: Page): Promise<void> {
  // The viewer renders its filters only after the first snapshot, so wait for the element to be
  // attached before deciding whether the disclosure is on screen.
  const disclosure = page.locator('.observation-filters__disclosure');
  await disclosure.waitFor({ state: 'attached' });
  if (await disclosure.isVisible()) await disclosure.click();
}

/** Switch to the radar view. List is the default surface, so radar tests must ask for it. */
async function showRadar(page: Page): Promise<void> {
  await page.getByRole('button', { name: 'Radar', exact: true }).click();
  await expect(page.locator('.observation-radar')).toBeVisible();
}

/*
 * Wireless observation viewer (spec 064).
 *
 * This is the only suite that renders the page. Unit suites cover the pure logic, but a
 * mount-time failure (a wrong store import, a bad prop) is invisible to them: this spec exists
 * because exactly that class of bug reached the branch once already.
 */

test.describe('wireless viewer', () => {
  /**
   * The mock outlives a single spec, so turning observation off here would leave every later test
   * looking at an empty radar. Restored unconditionally, including when this test itself failed.
   */
  test.afterEach(async ({ page }) => {
    await page.request.post(`${HUB_API}/__test/observation`, { data: { enabled: true } }).catch(() => undefined);
  });

  test('turned off observation is explained and links to the fix', async ({ page }) => {
    await signIn(page, '/wireless');
    // Nothing to assert until the socket has delivered the populated fixture.
    await expect(page.locator('.observation-groups').getByText('Front door tag')).toBeVisible();

    await page.request.post(`${HUB_API}/__test/observation`, { data: { enabled: false } });

    // An empty radar is what a broken sensor looks like too, so the page has to name which one this
    // is rather than leaving the operator to guess.
    const notice = page.getByRole('status').filter({ hasText: 'Observation is turned off' });
    await expect(notice).toBeVisible();
    const fix = page.getByRole('link', { name: 'Turn it on in the integration settings', exact: true });
    await expect(fix).toBeVisible();

    // The header link is always there, not only in this state.
    await expect(page.getByRole('link', { name: 'Integration settings', exact: true })).toBeVisible();

    // This fixture has no integrations service at all, so the link falls back to Integrations rather
    // than naming an instance that cannot exist. The direct, instance-specific target is asserted
    // against the Go-backed fixture in ConnectionsPage.spec.ts, where instances are real.
    await expect(fix).toHaveAttribute('href', /^\/integrations/);
    await fix.click();
    await expect(page).toHaveURL(/\/integrations$/);
  });

  test('renders observed identities with source diagnostics', async ({ page }) => {
    await signIn(page, '/wireless');

    // The page is an integration detail now, so its heading names the integration rather than the
    // medium, and the header carries a way back to Integrations.
    await expect(page.getByRole('heading', { name: 'Wireless observation', exact: true })).toBeVisible();
    await expect(page.getByRole('link', { name: 'Integrations', exact: true }).first()).toBeVisible();

    // Assertions are scoped to the table: the radar draws the same identity text in SVG
    // <title> tooltips, so an unscoped text match is ambiguous by design.
    const table = page.locator('.observation-groups');
    await expect(table.getByText('Front door tag')).toBeVisible();
    await expect(table.getByRole('button', { name: /^Isolate addr:aa:bb:cc:dd:ee:01, /, exact: false })).toBeVisible();
    await expect(table.getByRole('button', { name: /^Isolate fp:9c02ab41, /, exact: false })).toBeVisible();

    // Source health sits behind a disclosure now: it is a diagnostic, and keeping it expanded
    // pushed the radar and device list below the fold on every visit. A silent source still forces
    // its notice into view, so a fault cannot hide here.
    await page.locator('.wireless-viewer__sources-details summary').click();
    // Exact match: each row also says "via Mock hall proxy", so a loose text match is ambiguous.
    await expect(page.getByText('Mock hall proxy', { exact: true })).toBeVisible();
    await expect(page.getByText(/Scanning · 128 advertisements/)).toBeVisible();
    await expect(page.getByText(/no advertisements received/i)).toBeVisible();
    await expect(page.getByText(/One source is connected but not sending/i)).toBeVisible();
  });

  test('the range display is explicitly bearing-less', async ({ page }) => {
    await signIn(page, '/wireless');
    await showRadar(page);
    // The viewer must never imply a bearing it cannot compute from a single source.
    await expect(page.getByText(/Range only · no bearing/)).toBeVisible();
  });

  test('facet filters appear because a source declares them', async ({ page }) => {
    await signIn(page, '/wireless');
    await revealFilters(page);
    const filters = page.locator('.observation-filters');
    await expect(filters).toBeVisible();
    // Empty state must not appear: this is the D-1 regression guard.
    await expect(page.getByText('No source declares filterable facets yet.')).toHaveCount(0);
    await expect(filters.getByRole('button', { name: 'Needs triage', exact: true })).toBeVisible();
  });

  test('needs triage narrows to untriaged devices and drops them once triaged', async ({ page }) => {
    // This spec writes a triage decision into the shared mock, so it starts from a clean fixture
    // rather than inheriting whatever a previous spec decided.
    await resetMockState(page);
    await signIn(page, '/wireless');
    await revealFilters(page);
    const list = page.locator('.observation-groups');

    await page.getByRole('button', { name: 'Needs triage', exact: true }).click();

    // Only devices whose decision is still unknown remain. The beacon is already marked known in
    // the fixture, so it is gone; the untriaged identities stay.
    await expect(list.getByText('Front door tag')).toHaveCount(0);
    await expect(list.getByRole('button', { name: /^Isolate addr:aa:bb:cc:dd:ee:01, /, exact: false })).toBeVisible();

    // Triaging a device removes it from the filtered view. This is the behaviour that was broken:
    // the filter used to consider only arrival time, so marking a device known left it on screen
    // and the control looked inert. The row is targeted by name rather than by position: row order
    // is deliberately independent of arrival, so "the first row" is not a stable subject.
    const row = list
      .locator('.observation-row')
      .filter({ has: page.getByRole('button', { name: /^Isolate addr:aa:bb:cc:dd:ee:01, /, exact: false }) });
    await row.getByRole('button', { name: 'Known', exact: true }).click();
    await expect(list.getByRole('button', { name: /^Isolate addr:aa:bb:cc:dd:ee:01, /, exact: false })).toHaveCount(0);
  });

  test('pausing freezes the cards so they can be read', async ({ page }) => {
    await signIn(page, '/wireless');
    await page.getByRole('button', { name: 'Pause updates', exact: true }).click();

    await expect(page.getByRole('button', { name: 'Resume updates', exact: true })).toBeVisible();
    await expect(page.getByText('Frozen', { exact: true })).toBeVisible();
    // The table keeps rendering while frozen; it is a snapshot, not an empty state.
    await expect(page.locator('.observation-row').getByText('Front door tag')).toBeVisible();

    await page.getByRole('button', { name: 'Resume updates', exact: true }).click();
    await expect(page.getByRole('button', { name: 'Pause updates', exact: true })).toBeVisible();
  });

  test('the list view groups identities into collapsible class sections', async ({ page }) => {
    await signIn(page, '/wireless');
    await page.getByRole('button', { name: 'List', exact: true }).click();

    const groups = page.locator('.observation-groups');
    await expect(groups).toBeVisible();
    // Sections are behaviour classes, in the same fixed order as the Class presets. Vendor
    // grouping was opaque: "Vendor 0x004C" answers nothing the operator is asking.
    const section = groups.getByRole('button', { name: /^Yours/ });
    await expect(section).toHaveAttribute('aria-expanded', 'true');
    // Sections collapse so a noisy category can be folded away, and the count stays visible.
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
    // Start from the grouped list, then isolate one identity by name.
    await page.getByRole('button', { name: 'List', exact: true }).click();
    await page.locator('.observation-groups').getByRole('button', { name: /Front door tag/ }).first().click();

    const focus = page.locator('.observation-focus');
    await expect(focus).toBeVisible();
    await expect(focus.getByRole('heading', { name: 'Front door tag' })).toBeVisible();
    // Only the isolated identity is present in the list; the others are gone from it.
    await expect(page.locator('.observation-row')).toHaveCount(0);

    // The radar stays visible with the selection. It used to be replaced by a separate Focus
    // screen with its own mini radar, so selecting a device threw away the view you selected it
    // from and the two radars could disagree about distance.
    await expect(page.locator('.observation-radar')).toBeVisible();
    await expect(page.locator('.observation-radar__blip-group.is-selected')).toHaveCount(1);
  });

  test('Focus shows distance, band and trend', async ({ page }) => {
    await signIn(page, '/wireless');
    await page.getByRole('button', { name: 'List', exact: true }).click();
    await page.locator('.observation-groups').getByRole('button', { name: /Front door tag/ }).first().click();

    const focus = page.locator('.observation-focus');
    // Distance and motion are presented as one statement, because together they answer "where is
    // it and is it coming closer" — the question the panel exists for.
    await expect(focus.locator('.observation-focus__value')).toContainText('m');
    await expect(focus.locator('.observation-focus__band-line')).toContainText(/near|mid|far/);
    // The trend pairs an arrow with a word, never a colour or arrow alone.
    await expect(focus.getByText(/Approaching|Receding|Steady|No trend yet/)).toBeVisible();
    // Measurement detail lives here rather than on the card, where it was one of eight columns.
    await expect(focus.locator('.observation-focus__facts')).toBeVisible();
  });

  test('escape clears isolation and returns to the full view', async ({ page }) => {
    await signIn(page, '/wireless');
    await page.getByRole('button', { name: 'List', exact: true }).click();
    await page.locator('.observation-groups').getByRole('button', { name: /Front door tag/ }).first().click();
    await expect(page.locator('.observation-focus')).toBeVisible();

    await page.keyboard.press('Escape');
    await expect(page.locator('.observation-focus')).toHaveCount(0);
    // Back to the unfiltered set, so the other identities are reachable again. After Escape the
    // view returns to the radar layout, where the card carries the distance in its label.
    await expect(page.locator('.observation-groups').getByText('Front door tag')).toHaveCount(1);
  });

  test('the All identities control also leaves isolation', async ({ page }) => {
    await signIn(page, '/wireless');
    await page.getByRole('button', { name: 'List', exact: true }).click();
    await page.locator('.observation-groups').getByRole('button', { name: /Front door tag/ }).first().click();
    await page.getByRole('button', { name: '‹ All devices', exact: true }).click();
    await expect(page.locator('.observation-focus')).toHaveCount(0);
  });

  test('the radar scale zooms in and out with an unambiguous direction', async ({ page }) => {
    await signIn(page, '/wireless');
    await showRadar(page);

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
    await showRadar(page);

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
    await revealFilters(page);
    const fine = page.getByRole('button', { name: '0.5–1 m', exact: true });
    await expect(fine).toBeVisible();
    await fine.click();
    await expect(fine).toHaveAttribute('aria-pressed', 'true');
  });

  test('a distance band narrows the list to that range only', async ({ page }) => {
    await signIn(page, '/wireless');
    await revealFilters(page);
    await page.getByRole('button', { name: 'List', exact: true }).click();

    // The mock spans roughly 1.9 m, 11 m and an unplaceable one, so a band is genuinely
    // selective. The labelled beacon is smoothed to -66 dBm, which is 1.91 m: the 1-2 m band.
    const band = page.getByRole('button', { name: '1–2 m', exact: true });
    await expect(band).toBeVisible();
    await band.click();
    await expect(band).toHaveAttribute('aria-pressed', 'true');
    const groups = page.locator('.observation-groups');
    await expect(groups.getByText('Front door tag')).toHaveCount(1);
    await expect(groups.getByText('addr:aa:bb:cc:dd:ee:01')).toHaveCount(0);

    // A band that excludes the near device must exclude it, not merely reorder.
    await page.getByRole('button', { name: '5–10 m', exact: true }).click();
    await expect(groups.getByText('Front door tag')).toHaveCount(0);

    // A band with no devices is empty rather than falling back to everything.
    await page.getByRole('button', { name: '0.5–1 m', exact: true }).click();
    await expect(groups.getByText('Front door tag')).toHaveCount(0);

    // Any clears the filter. The list view is still active, so its own controls are asserted.
    await page.getByRole('button', { name: 'Any', exact: true }).click();
    await expect(groups.getByText('Front door tag')).toHaveCount(1);
    await expect(groups.getByText('addr:aa:bb:cc:dd:ee:01')).toHaveCount(1);
  });

  test('the distance control is a single-select group of bands', async ({ page }) => {
    await signIn(page, '/wireless');
    await revealFilters(page);
    const any = page.getByRole('button', { name: 'Any', exact: true });
    const band = page.getByRole('button', { name: '2–5 m', exact: true });
    await expect(any).toHaveAttribute('aria-pressed', 'true');
    await band.click();
    // Bands are exclusive: one selected, never two.
    await expect(band).toHaveAttribute('aria-pressed', 'true');
    await expect(any).toHaveAttribute('aria-pressed', 'false');
  });

  test('the list uses the page width and the radar has room in its own view', async ({ page }) => {
    // The rail beside the radar was capped at 24rem, which truncated every name. The list is now
    // the primary full-width surface, and the radar is a deliberate view of its own.
    await page.setViewportSize({ width: 1440, height: 900 });
    await signIn(page, '/wireless');

    const listBox = (await page.locator('.observation-groups').boundingBox())!;
    expect(listBox.width).toBeGreaterThan(900);

    await showRadar(page);
    const box = (await page.locator('.observation-radar__canvas').boundingBox())!;
    const viewport = page.viewportSize()!;

    // Still substantial: at least a third of the viewport, and large in absolute terms.
    expect(box.height).toBeGreaterThan(viewport.height * 0.33);
    expect(box.height).toBeGreaterThan(200);
    // It stays square, so the rings remain circular.
    expect(Math.abs(box.width - box.height)).toBeLessThan(2);
  });

  test('the radar fills the column on a phone without overflow', async ({ page }) => {
    await page.setViewportSize({ width: 390, height: 844 });
    await signIn(page, '/wireless');
    await showRadar(page);
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
    await showRadar(page);

    // Blips are real controls, reachable by role rather than by pixel position.
    const blip = page.getByRole('button', { name: /^Isolate Front door tag, \d+ m$/ }).first();
    await expect(blip).toBeVisible();
    await blip.click();

    const focus = page.locator('.observation-focus');
    await expect(focus).toBeVisible();
    await expect(focus.getByRole('heading', { name: 'Front door tag' })).toBeVisible();
    // Only the isolated identity remains anywhere in the view.
    await expect(page.getByRole('button', { name: /^Isolate addr:aa:bb:cc:dd:ee:01, /, exact: false })).toHaveCount(0);
  });

  test('blips are keyboard operable', async ({ page }) => {
    await signIn(page, '/wireless');
    await showRadar(page);
    const blip = page.getByRole('button', { name: /^Isolate Front door tag, \d+ m$/ }).first();
    await blip.focus();
    // Visible focus is required: the SVG highlight is the only cue here.
    await expect(blip).toBeFocused();
    await blip.press('Enter');
    await expect(page.locator('.observation-focus')).toBeVisible();
  });

  test('the radar derives its distance from smoothed signal', async ({ page }) => {
    await signIn(page, '/wireless');
    await showRadar(page);
    const blip = page.getByRole('button', { name: /^Isolate Front door tag, \d+ m$/ }).first();
    const label = (await blip.getAttribute('aria-label')) ?? '';
    const metres = Number(label.match(/(\d+) m$/)?.[1]);

    // The mock sends rssi -62 but rssiSmoothed -66, with txPower -59 and exponent 2.5:
    //   raw      -> 10^((3)/25)  = 1.3 m -> 1 m
    //   smoothed -> 10^((7)/25)  = 1.9 m -> 2 m
    // So 2 proves the smoothed value is used; 1 would mean the single packet leaked through.
    expect(metres).toBe(2);
  });

  test('triage controls are keyboard reachable and report a decision', async ({ page }) => {
    await resetMockState(page);
    await signIn(page, '/wireless');
    const known = page.getByRole('button', { name: 'Known', exact: true }).first();
    await known.focus();
    // Visible focus is an FE-04 requirement, not decoration.
    await expect(known).toBeFocused();
    await known.press('Enter');

    // The decision is applied and rebroadcast, so the triaged device leaves an untriaged filter and
    // the button disappears for it. Asserting "still enabled" would pass on a mock that only
    // acknowledged and never applied anything — which is exactly what this suite used to do, and
    // why a real triage defect went unnoticed.
    await expect(page.getByRole('button', { name: 'Known', exact: true }).first()).not.toBeFocused();
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

    const table = page.locator('.observation-groups');
    // The installed device is hidden by default; this is the row type that made the list useless.
    await expect(table.getByRole('button', { name: /^Isolate addr:aa:bb:cc:dd:ee:02, /, exact: false })).toHaveCount(0);

    // Hiding is never silent: the footer states how many and offers a way back.
    await expect(page.getByText(/1 of \d+ devices\s+hidden/)).toBeVisible();
    await expect(page.getByRole('button', { name: 'Show them', exact: true })).toBeVisible();
  });

  test('revealing demoted classes brings them back', async ({ page }) => {
    await signIn(page, '/wireless');

    await page.getByRole('button', { name: 'Show them', exact: true }).click();

    const table = page.locator('.observation-groups');
    await expect(table.getByRole('button', { name: /^Isolate addr:aa:bb:cc:dd:ee:02, /, exact: false })).toBeVisible();
    await expect(page.getByText(/devices\s+hidden/)).toHaveCount(0);
  });

  test('the show-all toggle is an alternative route to the same state', async ({ page }) => {
    await signIn(page, '/wireless');
    await revealFilters(page);

    await page.getByRole('button', { name: 'Show all kinds', exact: true }).click();

    const table = page.locator('.observation-groups');
    await expect(table.getByRole('button', { name: /^Isolate addr:aa:bb:cc:dd:ee:02, /, exact: false })).toBeVisible();
  });

  test('every visible row states its class in words', async ({ page }) => {
    await signIn(page, '/wireless');
    const table = page.locator('.observation-groups');

    // No raw identifiers: the operator must never see "infrastructure" as a badge.
    await expect(table.locator('.class-badge').first()).toBeVisible();
    await expect(table.getByText('Yours', { exact: true })).toBeVisible();
    await expect(table.getByText('infrastructure', { exact: true })).toHaveCount(0);
  });

  test('learning shows its progress rather than a bare label', async ({ page }) => {
    await signIn(page, '/wireless');

    // A slow classifier that shows nothing looks broken; naming the progress makes it read as work
    // in progress.
    await expect(page.locator('.observation-groups').getByText('Learning', { exact: true })).toBeVisible();
    await expect(page.locator('.observation-groups').getByText('6 of 24 hours', { exact: true })).toBeVisible();
  });

  test('a provisional class is marked as a guess', async ({ page }) => {
    await signIn(page, '/wireless');
    const table = page.locator('.observation-groups');

    const badge = table.locator('.class-badge--provisional').first();
    await expect(badge).toBeVisible();
    // The marker and the accessible text both say so, so it is not colour-only.
    await expect(table.getByText('(provisional guess)', { exact: false }).first()).toBeAttached();
  });

  test('the isolated view explains the class and shows its evidence', async ({ page }) => {
    await signIn(page, '/wireless');

    // The beacon carries a settled class with a reason, so it is the one whose explanation can be
    // asserted; the other identities are still learning. Scoped to the card list: the radar blip
    // carries the same accessible name.
    await page.locator('.observation-groups').getByRole('button', { name: /^Isolate Front door tag, /, exact: false }).click();

    const focus = page.locator('.observation-focus');
    await expect(focus).toBeVisible();
    // The reasoning must be present: a class the operator cannot check is a class they cannot trust.
    // Scoped to the reason element: the same sentence also appears in the badge tooltip, and a
    // strict-mode match on both would be ambiguous.
    await expect(focus.locator('.observation-focus__reason')).toContainText(
      'Present for part of the day with a changing signal',
    );
    await expect(focus.getByText(/Advisory only/)).toBeVisible();
    await expect(focus.getByRole('heading', { name: 'Evidence', exact: true })).toBeVisible();
  });

  test('the radar legend names its classes so colour is never the only cue', async ({ page }) => {
    await signIn(page, '/wireless');
    await showRadar(page);

    const legend = page.locator('.observation-radar__legend');
    await expect(legend).toBeVisible();
    await expect(legend.getByText('Yours', { exact: true })).toBeVisible();
  });

  test('classification survives a phone viewport without overflow', async ({ page }) => {
    await signIn(page, '/wireless');
    await page.locator('.observation-groups').getByRole('button', { name: /^Isolate addr:aa:bb:cc:dd:ee:01, /, exact: false }).click();

    const overflow = await page.evaluate(() => document.documentElement.scrollWidth > window.innerWidth + 1);
    expect(overflow).toBe(false);
  });
});

/*
 * Class presets and search (spec 066).
 *
 * The classification existed but could not be asked about. These tests assert that a preset
 * actually isolates its group, that an explicitly chosen demoted class is revealed rather than
 * filtered into nothing, and that search narrows to a known identifier.
 */
test.describe('class presets and search', () => {
  test('a class preset isolates its group and Any restores the view', async ({ page }) => {
    await signIn(page, '/wireless');
    await revealFilters(page);
    const list = page.locator('.observation-groups');

    // "Yours" is the settled personal class: the front door tag only.
    const yours = page.getByRole('button', { name: /^Yours, / });
    await expect(yours).toBeVisible();
    await yours.click();
    await expect(yours).toHaveAttribute('aria-pressed', 'true');
    await expect(list.getByText('Front door tag')).toBeVisible();
    await expect(list.getByRole('button', { name: /^Isolate addr:aa:bb:cc:dd:ee:01, /, exact: false })).toHaveCount(0);

    // A zero-count group stays in place instead of disappearing, so the row cannot shuffle while
    // the operator is reading or clicking it.
    await expect(page.getByRole('button', { name: 'Passing, 0 devices', exact: true })).toBeVisible();

    await page.getByRole('button', { name: 'Any class', exact: true }).click();
    await expect(list.getByRole('button', { name: /^Isolate addr:aa:bb:cc:dd:ee:01, /, exact: false })).toBeVisible();
  });

  test('choosing a demoted class reveals it without the reveal toggle', async ({ page }) => {
    await signIn(page, '/wireless');
    await revealFilters(page);
    const list = page.locator('.observation-groups');

    // Installed equipment is hidden by default; an explicit preset must still show it, or the
    // preset would look broken.
    await expect(list.getByRole('button', { name: /^Isolate addr:aa:bb:cc:dd:ee:02, /, exact: false })).toHaveCount(0);
    await page.getByRole('button', { name: /^Installed, / }).click();

    await expect(list.getByRole('button', { name: /^Isolate addr:aa:bb:cc:dd:ee:02, /, exact: false })).toBeVisible();
    // Nothing is hidden by the default rule now, so the footer does not claim otherwise.
    await expect(page.getByText(/devices\s+hidden/)).toHaveCount(0);
  });

  test('search narrows to a matching name and clearing restores the view', async ({ page }) => {
    await signIn(page, '/wireless');
    const list = page.locator('.observation-groups');
    const search = page.getByPlaceholder('Name, address or vendor');

    await search.fill('Front door');
    await expect(list.getByText('Front door tag')).toBeVisible();
    await expect(list.getByRole('button', { name: /^Isolate addr:aa:bb:cc:dd:ee:01, /, exact: false })).toHaveCount(0);

    await page.getByRole('button', { name: 'Clear search', exact: true }).click();
    await expect(list.getByRole('button', { name: /^Isolate addr:aa:bb:cc:dd:ee:01, /, exact: false })).toBeVisible();
  });

  test('a query with no matches says so instead of looking empty', async ({ page }) => {
    await signIn(page, '/wireless');
    await page.getByPlaceholder('Name, address or vendor').fill('no-such-device');

    await expect(page.getByText('No devices match “no-such-device”.', { exact: true })).toBeVisible();
  });
});

/*
 * Filter layout (spec 066).
 *
 * The reported defects were visual: a bare toggle floated out of line with the labelled fields,
 * the joined distance control clipped its last bands on a phone, and class chips moved as their
 * counts changed. These assert geometry, not just presence, because presence alone did not catch
 * any of them.
 */
test.describe('filter layout', () => {
  test('labelled filter controls share one label-to-control rhythm', async ({ page }) => {
    await page.setViewportSize({ width: 1280, height: 900 });
    await signIn(page, '/wireless');
    await page.locator('.observation-filters__chip').first().waitFor();

    // Compare the gap between each label and its control, not the absolute row: the fields may
    // wrap onto a second line, and wrapping is fine as long as every control sits at the same
    // distance below its own label.
    const gap = async (label: ReturnType<Page['locator']>, control: ReturnType<Page['locator']>) => {
      const labelBox = (await label.boundingBox())!;
      const controlBox = (await control.boundingBox())!;
      return controlBox.y - (labelBox.y + labelBox.height);
    };

    const gaps = await Promise.all([
      gap(page.locator('.observation-filters__label'), page.locator('.observation-filters__search-box')),
      gap(
        page.locator('.observation-filters__field').nth(0).locator('legend'),
        page.locator('.observation-filters__field').nth(0).locator('.observation-filters__chips'),
      ),
      gap(
        page.locator('.observation-filters__field').nth(1).locator('legend'),
        page.locator('.observation-filters__field').nth(1).locator('.observation-filters__chips'),
      ),
      gap(
        page.locator('.observation-filters__field').nth(2).locator('legend'),
        page.locator('.observation-filters__field').nth(2).locator('.observation-filters__presets'),
      ),
    ]);
    expect(Math.max(...gaps) - Math.min(...gaps)).toBeLessThan(1);
  });

  test('the phone filter panel opens and its bands wrap without clipping', async ({ page }) => {
    await page.setViewportSize({ width: 390, height: 844 });
    await signIn(page, '/wireless');
    await page.locator('.observation-filters__disclosure').waitFor();

    // Collapsed, the fields are not in the accessibility tree; the disclosure is.
    await expect(page.getByRole('button', { name: 'Needs triage', exact: true })).toHaveCount(0);
    await revealFilters(page);
    await expect(page.getByRole('button', { name: 'Needs triage', exact: true })).toBeVisible();

    // Every band is reachable and the row wraps instead of overflowing its column.
    await expect(page.getByRole('button', { name: '≥ 10 m', exact: true })).toBeVisible();
    const overflow = await page
      .locator('.observation-filters__presets')
      .evaluate((element) => element.scrollWidth - element.clientWidth);
    expect(overflow).toBeLessThanOrEqual(1);
  });

  test('class chips keep their width and order as their counts change', async ({ page }) => {
    await page.setViewportSize({ width: 1280, height: 900 });
    await signIn(page, '/wireless');
    const classChips = page.locator('.observation-filters__chips').first();
    await classChips.locator('.observation-filters__chip').first().waitFor();

    // "Someone's device" is the widest label and has a two-digit count in the fixture.
    const chip = classChips.locator('.observation-filters__chip').nth(2);
    const before = (await chip.boundingBox())!;
    const groupsBefore = await classChips
      .locator('.observation-filters__chip')
      .evaluateAll((nodes) => nodes.map((node) => node.getAttribute('data-group')));

    await page.getByPlaceholder('Name, address or vendor').fill('Front door');

    const after = (await chip.boundingBox())!;
    const groupsAfter = await classChips
      .locator('.observation-filters__chip')
      .evaluateAll((nodes) => nodes.map((node) => node.getAttribute('data-group')));

    // Same chips, same place: a count going to zero must not remove or resize one.
    expect(groupsAfter).toEqual(groupsBefore);
    expect(Math.abs(after.x - before.x)).toBeLessThan(1);
    expect(Math.abs(after.width - before.width)).toBeLessThan(1);
  });
});
