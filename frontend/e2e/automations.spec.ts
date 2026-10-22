import { test, expect, signIn, signInAutomationHub, AUTOMATION_HUB_API } from './support/fixtures';
import { ui } from './support/ui';
import { MOCK } from './support/mock';

// Spec 007 US-04 / AC-14: automations list, and saving an unchanged automation must send
// exactly the payload the hub provided (no fields added, removed or reordered in arrays).
// Spec 025 AC-06 is the one exception: a zero trigger-action delay is omitted because the
// hub rejects it; the editor's zero delay means "no wait".
test.describe('automations', () => {
  test('binary conditions exclude commands and switching action properties retains typed inputs', async ({ page, ws, request }) => {
    const unauthorized = await request.post(`${AUTOMATION_HUB_API}/__test/dial`);
    expect(unauthorized.status()).toBe(401);
    await signInAutomationHub(page, ui.editor.path('0x0000000000000002'));
    await ui.editor.when(page).selectOption({ label: 'State' });
    await ui.editor.addCondition(page).click();
    await ui.editor.conditionProperty(page).selectOption({ label: 'State' });
    await expect(ui.editor.compareWith(page).getByRole('option')).toHaveText(['Choose a value', 'On', 'Off']);
    await ui.editor.compareWith(page).selectOption({ label: 'Off' });
    await ui.editor.addAction(page).click();
    await ui.editor.actionDevice(page).selectOption({ label: 'Test light' });
    await ui.editor.addValue(page).click();
    await ui.editor.propertyToSet(page).selectOption({ label: 'State' });
    await ui.editor.newValue(page).selectOption({ label: 'Toggle' });
    await ui.editor.propertyToSet(page).selectOption({ label: 'Brightness' });
    await expect(ui.editor.numericValue(page)).toHaveValue('');
    await ui.editor.numericValue(page).fill('0');
    await expect(ui.editor.summary(page)).toContainText('brightness to 0 %');
    await ui.editor.propertyToSet(page).selectOption({ label: 'State' });
    await expect(ui.editor.newValue(page).locator('option:checked')).toHaveText('Choose a value');
    expect(ws.sentOf('saveAutomation')).toHaveLength(0);
    expect(ws.sentOf('deviceSetValue')).toHaveLength(0);
  });

  test('a UI-created dial automation toggles a light on and off through the Go engine', async ({ page, ws }) => {
    const dial = '0x0000000000000001';
    const light = '0x0000000000000002';
    await signInAutomationHub(page, ui.editor.path(dial));
    // An empty saved recipe must reopen as an editable list, not crash on null.
    await ui.editor.removeTrigger(page);
    await ui.editor.save(page);
    await ws.waitForReceived('operationSuccess');
    await ui.automations.open(page, 'Test dial');
    await page.reload();
    await expect(ui.editor.enabled(page)).toBeVisible();
    await expect(ui.editor.when(page)).toHaveCount(0);
    await ui.editor.addTrigger(page).click();
    await ui.editor.enabled(page).check();
    await ui.editor.when(page).selectOption({ label: 'Action' });
    await ui.editor.addCondition(page).click();
    await ui.editor.conditionProperty(page).selectOption({ label: 'Action' });
    await ui.editor.compareWith(page).selectOption({ label: 'Button 1 press release' });
    await ui.editor.addAction(page).click();
    await ui.editor.actionDevice(page).selectOption({ label: 'Test light' });
    await ui.editor.addValue(page).click();
    await ui.editor.propertyToSet(page).selectOption({ label: 'State' });
    const value = ui.editor.newValue(page);
    await expect(value.getByRole('option')).toHaveText(['Choose a value', 'On', 'Off', 'Toggle']);
    // Exercise native keyboard selection as well as its accessible name/focus.
    await value.focus();
    await expect(value).toBeFocused();
    await value.press('t');
    await value.press('Tab');
    await expect(value.locator('option:checked')).toHaveText('Toggle');
    await expect(ui.editor.summary(page)).toContainText('set Test light state to Toggle');
    const beforeSave = ws.received.length;
    await ui.editor.save(page);
    const saved = await ws.waitForSent('saveAutomation', (a) => a.id === dial && a.enabled);
    expect(saved.payload.triggers).toHaveLength(1);
    expect(saved.payload.triggers[0].actions[0].exposes).toEqual([{ name: 'state', data: 'TOGGLE' }]);
    expect(saved.payload.triggers[0].actions[0]).not.toHaveProperty('delay');
    await expect.poll(() => ws.received.slice(beforeSave).filter((f) => f.type === 'operationSuccess' || f.type === 'operationFailed')
      .map((f) => f.type === 'operationFailed' ? f.payload : f.type)).toEqual(['operationSuccess']);

    // Reopen and reload from the backend, proving persistence through its real serializer.
    await expect(page).toHaveURL(/\/viewer$/);
    await ui.automations.open(page, 'Test dial');
    await page.reload();
    await expect(ui.editor.newValue(page).locator('option:checked')).toHaveText('Toggle');
    await expect(ui.editor.summary(page)).toContainText('set Test light state to Toggle');
    await page.goto(ui.devices.path(light));
    await expect(ui.devices.state(page)).not.toBeChecked();
    for (const state of ['ON', 'OFF']) {
      const response = await page.request.post(`${AUTOMATION_HUB_API}/__test/dial`);
      expect(response.ok(), await response.text()).toBeTruthy();
      expect(await response.json()).toEqual({ command: { state: 'TOGGLE' }, confirmedState: state });
      await expect(ui.devices.state(page)).toBeChecked({ checked: state === 'ON' });
      await ws.waitForReceived('deviceUpdated', (p) => p.id === light && p.data.state === state);
    }
    expect(ws.sentOf('deviceSetValue')).toHaveLength(0);
    // Reset must clear a light left ON, not fail the next case at sign-in.
    const leftover = await page.request.post(`${AUTOMATION_HUB_API}/__test/dial`);
    expect(leftover.ok(), await leftover.text()).toBeTruthy();
    expect((await leftover.json()).confirmedState).toBe('ON');
    expect((await page.request.post(`${AUTOMATION_HUB_API}/__test/reset`)).status()).toBe(204);
    await expect(ui.devices.state(page)).toBeChecked({ checked: false });
  });

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
