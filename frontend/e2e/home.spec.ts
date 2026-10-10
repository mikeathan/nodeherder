import { test, expect, signIn } from './support/fixtures';
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
});
