import { test, expect, signIn } from './support/fixtures';
import { ui } from './support/ui';
import { MOCK } from './support/mock';

// Spec 007 US-07 (AC-23, AC-24): panel mode shows each Home area and exits back to Home.
test.describe('panel mode', () => {
  test('shows one tab per area and exits to Home', async ({ page }) => {
    await signIn(page, ui.panel.path);
    await expect(ui.panel.roomTabs(page)).toHaveCount(MOCK.groups.length);
    await ui.panel.roomTabs(page).nth(1).click();
    await expect(ui.panel.roomTabs(page).nth(1)).toHaveAttribute('aria-selected', 'true');
    await ui.panel.exit(page).click();
    await expect(page).toHaveURL(/\/groupdashboard$/);
  });
});
