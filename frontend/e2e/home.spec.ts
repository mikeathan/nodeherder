import { test, expect, signIn, dragBetween } from './support/fixtures';
import { ui } from './support/ui';
import { MOCK } from './support/mock';

// Spec 007 US-02: dashboards must render hub groups and switch devices with the same command.
test.describe('home dashboard', () => {
  test('shows all dashboard groups', async ({ page }) => {
    await signIn(page, ui.home.path);
    await expect(ui.home.groupTitles(page)).toHaveCount(MOCK.groups.length);
  });

  test('switching a light sends deviceSetValue for its state expose', async ({ page, ws }) => {
    await signIn(page, ui.home.path);
    // the mock brings the living room light online on its first update (~8 s)
    await ws.waitForReceived('deviceUpdated', (p) => p.id === MOCK.livingRoomLight.id, 15_000);
    const tile = ui.home.tile(page, 'living room group', 'Brightness');
    await ui.home.toggleControl(tile).click();
    const frame = await ws.waitForSent('deviceSetValue');
    expect(frame.payload).toMatchObject({ id: MOCK.livingRoomLight.id, name: 'state' });
    expect(['ON', 'OFF']).toContain(frame.payload.value);
  });

  // Spec 007 US-06 (AC-19…AC-22): areas can be rearranged and the order is saved on the hub.
  test('arrow buttons move an area and save its new position', async ({ page, ws }) => {
    await signIn(page, ui.home.path);
    const titles = ui.home.groupTitles(page);
    await expect(titles).toHaveCount(MOCK.groups.length);
    const before = await titles.allTextContents();
    await ui.home.editLayout(page);
    await ui.home.moveLater(page, before[0]).click();
    await expect(titles).toHaveText([before[1], before[0], ...before.slice(2)]);
    const saved = await ws.waitForSent('saveDashboardGroup', (p) => p.name === before[0]);
    expect(saved.payload.order).toBe(1);
    // put it back so other specs see the original order
    await page.getByRole('button', { name: `Move ${before[0]} earlier` }).click();
    await expect(titles).toHaveText(before);
    await ui.home.done(page);
  });

  test('dragging an area by its handle reorders and saves', async ({ page, ws }) => {
    await signIn(page, ui.home.path);
    const titles = ui.home.groupTitles(page);
    await expect(titles).toHaveCount(MOCK.groups.length);
    const before = await titles.allTextContents();
    await ui.home.editLayout(page);
    const handle = ui.home.dragHandle(page, before[0]);
    const target = ui.home.area(page, before[1]);
    const from = (await handle.boundingBox())!;
    const to = (await target.boundingBox())!;
    // areas collapse to their headers once one is held, so aim just past the next header
    const start = { x: from.x + from.width / 2, y: from.y + from.height / 2 };
    const end = to.y > from.y + from.height * 2 ? { x: start.x, y: start.y + 90 } : { x: to.x + to.width * 0.7, y: to.y + 16 };
    await dragBetween(page, start, end);
    await expect(titles.first()).toHaveText(before[1]);
    const saved = await ws.waitForSent('saveDashboardGroup', (p) => typeof p.order === 'number');
    expect(saved.payload).toHaveProperty('deviceGroup');
    // restore
    await page.getByRole('button', { name: `Move ${before[0]} earlier` }).click();
    await expect(titles).toHaveText(before);
  });
});
