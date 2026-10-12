import { test, expect, signIn, dragBetween } from './support/fixtures';
import { MOCK } from './support/mock';

// Characterizes the reported failure with controlled browser-side hub responses.
// No command reaches the mock hub or any physical device. The Go investigation separately
// reproduces why the production backend can suppress numeric echoes for 60 seconds.
for (const gesture of ['keyboard', 'drag'] as const) {
  test(`brightness ${gesture} sends again but times out when its confirmation is suppressed`, async ({ page }) => {
    const commands: { id: string; name: string; value: number }[] = [];
    await page.route('**/api/hubstate', async (route) => {
      const response = await route.fetch();
      const state = await response.json();
      const light = state.devices.find((d: { id: string }) => d.id === MOCK.livingRoomLight.id);
      light.availability = 'online';
      light.exposes.brightness.data = 50;
      await route.fulfill({ response, json: state });
    });
    await page.routeWebSocket('ws://localhost:4110/ws', (socket) => {
      socket.onMessage((raw) => {
        const frame = JSON.parse(String(raw));
        if (frame.type !== 'deviceSetValue') return;
        commands.push(frame.payload);
        if (commands.length === 2) return; // device echo filtered by backend sampling
        socket.send(JSON.stringify({ type: 'deviceUpdated', payload: {
          id: frame.payload.id, data: { [frame.payload.name]: frame.payload.value },
        } }));
      });
    });
    await signIn(page, `/devicepage/${MOCK.livingRoomLight.id}`);
    const slider = page.getByRole('slider', { name: 'Brightness', exact: true });
    await expect(slider).toBeEnabled();

    async function adjust() {
      if (gesture === 'keyboard') {
        await slider.focus();
        await slider.press('ArrowRight');
        return;
      }
      const box = (await slider.boundingBox())!;
      const value = Number(await slider.inputValue());
      const min = Number(await slider.getAttribute('min'));
      const max = Number(await slider.getAttribute('max'));
      const x = (v: number) => box.x + 8 + ((v - min) / (max - min)) * (box.width - 16);
      const y = box.y + box.height / 2;
      await dragBetween(page, { x: x(value), y }, { x: x(value + 40), y });
    }

    await adjust();
    await expect.poll(() => commands.length).toBe(1);
    expect(commands[0]).toEqual({ id: MOCK.livingRoomLight.id, name: 'brightness', value: expect.any(Number) });
    expect(commands[0].value).toBeGreaterThan(50);
    await expect(slider).toBeEnabled();

    await adjust();
    await expect.poll(() => commands.length).toBe(2);
    expect(commands[1].value).toBeGreaterThan(commands[0].value);
    await expect(slider).toBeDisabled();
    await expect(page.getByText('Sending…', { exact: true })).toBeVisible();
    await expect(page.getByText(/did not confirm the change/)).toBeVisible({ timeout: 7000 });
    await expect(slider).toBeEnabled();
    // Additional UI defect: the unconfirmed draft remains displayed after the timeout.
    await expect(slider).toHaveValue(String(commands[1].value));

    await slider.focus();
    await adjust();
    await expect.poll(() => commands.length).toBe(3);
    expect(commands[2].value).toBeGreaterThan(commands[1].value);
    await expect(slider).toBeEnabled();
    await expect(page.getByText('Sending…', { exact: true })).toHaveCount(0);
  });
}
