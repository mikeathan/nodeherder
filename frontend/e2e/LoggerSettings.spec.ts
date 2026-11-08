import { test, expect, signIn } from './support/fixtures';

// Spec 059 US-03: labelled logger controls that save the complete settings object.
test('logger toggle and level save complete settings', async ({ page, ws }) => {
  await signIn(page, '/settings?section=logger');
  await expect(page.getByRole('heading', { level: 2, name: 'Logger' })).toBeVisible();

  const toggle = page.getByRole('switch', { name: 'Send hub logs to this browser' });
  await expect(toggle).toHaveAttribute('aria-checked', 'false');
  await toggle.click();
  await expect.poll(() => ws.sentOf('saveLoggerConfig').at(-1)?.payload).toEqual({ enableRemoteLogger: true, level: 'info' });

  const level = page.getByRole('combobox', { name: 'Minimum level' });
  await expect(level).toHaveValue('2');
  await level.selectOption('warn');
  await expect.poll(() => ws.sentOf('saveLoggerConfig').at(-1)?.payload).toEqual({ enableRemoteLogger: true, level: 'warn' });

  // Restore the fixture values.
  await toggle.click();
  await level.selectOption('info');
  await expect.poll(() => ws.sentOf('saveLoggerConfig').at(-1)?.payload).toEqual({ enableRemoteLogger: false, level: 'info' });
});

test('logger is usable by keyboard and blocks edits while disconnected', async ({ page }) => {
  let disconnected = false;
  let disconnect: (() => Promise<void>) | undefined;
  await page.routeWebSocket('ws://localhost:4110/ws', (socket) => {
    if (disconnected) { void socket.close(); return; }
    socket.connectToServer(); disconnect = () => socket.close();
  });
  await signIn(page, '/settings?section=logger');
  const level = page.getByRole('combobox', { name: 'Minimum level' });
  await level.focus();
  await expect(level).toBeFocused();
  disconnected = true;
  await disconnect?.();
  await expect(level).toBeDisabled();
  await expect(page.getByRole('switch', { name: 'Send hub logs to this browser' })).toBeDisabled();
  await expect(page.getByText('Connect to the hub to change logger settings.', { exact: false })).toBeVisible();
});
