import { test, expect, signIn } from './support/fixtures';
import { ui } from './support/ui';
import { MOCK } from './support/mock';

// Spec 007 US-03: device list and device page.
test.describe('devices', () => {
  test('device list shows every hub device and opens a device page', async ({ page }) => {
    await signIn(page, ui.devices.listPath);
    await expect(ui.devices.listRows(page)).toHaveCount(MOCK.deviceCount);
    await ui.devices.openFromList(page, MOCK.atticSocket.name);
    await expect(page).toHaveURL(new RegExp(MOCK.atticSocket.id));
    await expect(ui.devices.pageTitle(page)).toContainText(MOCK.atticSocket.name);
  });
});
