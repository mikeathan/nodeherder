import { test, expect, signIn } from './support/fixtures';
import { MOCK } from './support/mock';

test('panel button retains confirmed state without pulsing or sending text', async ({ page }) => {
  const commands: { id: string; name: string; value: unknown }[] = [];
  let confirm!: () => void;
  await page.route('**/api/hubstate', async (route) => {
    const response = await route.fetch();
    const body = await response.json();
    const light = body.devices.find((device: { id: string }) => device.id === MOCK.livingRoomLight.id);
    light.availability = 'online';
    light.exposes.state.data = 'ON';
    light.exposes.brightness.data = 184;
    await route.fulfill({ response, json: body });
  });
  await page.routeWebSocket('ws://localhost:4110/ws', (socket) => {
    socket.onMessage((raw) => {
      const frame = JSON.parse(String(raw));
      if (frame.type === 'deviceSetValue') commands.push(frame.payload);
    });
    confirm = () => socket.send(JSON.stringify({ type: 'deviceUpdated', payload: {
      id: MOCK.livingRoomLight.id, data: { state: commands[0].value },
    } }));
  });
  await signIn(page, '/panel');
  await page.getByRole('tab', { name: 'living room group', exact: true }).click();
  const button = page.locator('button.pb').filter({ hasText: MOCK.livingRoomLight.name }).first();
  const before = await button.locator('.pb-txt').innerText();
  await button.click();
  await expect(button).toHaveAttribute('aria-busy', 'true');
  await expect(button).toHaveAttribute('aria-pressed', 'true');
  await expect(button.locator('.pb-txt')).toHaveText(before);
  await expect(button).toHaveCSS('animation-name', 'none');
  await expect(button.getByText(/Sending|Waiting for device confirmation/)).toHaveCount(0);
  await button.click();
  await expect.poll(() => commands.length).toBe(1);
  expect(commands[0]).toEqual({ id: MOCK.livingRoomLight.id, name: 'state', value: 'OFF' });
  confirm();
  await expect(button).toHaveAttribute('aria-pressed', 'false');
  await expect(button.locator('.pb-txt')).toHaveText('Off');
});
