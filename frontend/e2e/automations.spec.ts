import { test, expect, signIn } from './support/fixtures';
import { ui } from './support/ui';
import { MOCK } from './support/mock';

// Spec 007 US-04 / AC-14: automations list, and saving an unchanged automation must send
// exactly the payload the hub provided (no fields added, removed or reordered in arrays).
test.describe('automations', () => {
  test('list shows hub automations', async ({ page, ws }) => {
    await signIn(page, ui.automations.path);
    await ws.waitForReceived('automations');
    await expect(ui.automations.row(page, MOCK.presenceAutomation.name)).toBeVisible();
  });

  test('saving an unchanged automation round-trips the hub payload', async ({ page, ws }) => {
    await signIn(page, ui.automations.path);
    const list = await ws.waitForReceived('automations');
    const original = list.payload.find((a: any) => a.id === MOCK.presenceAutomation.id);
    expect(original).toBeTruthy();
    await ui.automations.open(page, MOCK.presenceAutomation.name);
    await expect(page).toHaveURL(new RegExp(ui.editor.path(MOCK.presenceAutomation.id)));
    await ui.editor.save(page);
    const saved = await ws.waitForSent('saveAutomation');
    expect(saved.payload).toEqual(original);
  });
});
