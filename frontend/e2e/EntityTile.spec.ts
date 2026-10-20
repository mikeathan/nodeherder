import { test, expect, signIn } from './support/fixtures';
import { ui } from './support/ui';
import { MOCK } from './support/mock';

test('Home tile retains its confirmed reading and appearance while a toggle awaits confirmation', async ({ page }) => {
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
  await signIn(page, ui.home.path);
  const tile = ui.home.tile(page, 'living room group', 'Brightness');
  const toggle = tile.locator('button.nh-tile-ic');
  const value = tile.locator('.nh-tile-value');
  const before = { text: await value.innerText(), color: await value.evaluate((el) => getComputedStyle(el).color) };
  await toggle.click();
  await expect(toggle).toHaveAttribute('aria-busy', 'true');
  await expect(toggle).toHaveAttribute('aria-pressed', 'true');
  await expect(value).toHaveText(before.text);
  await expect(value).toHaveCSS('color', before.color);
  await expect(tile.getByText(/Sending|Waiting for device confirmation/)).toHaveCount(0);
  await expect(tile.locator('.nh-spin')).toHaveCount(0);
  await toggle.click();
  await expect.poll(() => commands.length).toBe(1);
  expect(commands[0]).toEqual({ id: MOCK.livingRoomLight.id, name: 'state', value: 'OFF' });
  confirm();
  await expect(toggle).toHaveAttribute('aria-pressed', 'false');
  await expect(toggle).not.toHaveAttribute('aria-busy', 'true');
});
