import { test, expect, signIn } from './support/fixtures';

// Spec 059 US-03: the assistant address validates before saving and never mutates state.
test('assistant URL saves only valid addresses', async ({ page, ws }) => {
  await signIn(page, '/settings?section=assistant');
  await expect(page.getByRole('heading', { level: 2, name: 'Assistant' })).toBeVisible();

  const url = page.getByLabel('Assistant URL', { exact: true });
  await expect(url).toHaveValue('http://localhost:4001/api/conversation');

  const before = ws.sentOf('saveAssistantConfig').length;
  await url.fill('ftp://not-http');
  await page.getByRole('button', { name: 'Save address' }).click();
  await expect(page.getByRole('alert')).toContainText('http:// or https://');
  expect(ws.sentOf('saveAssistantConfig')).toHaveLength(before);

  await url.fill('https://assistant.local/chat');
  await page.getByRole('button', { name: 'Save address' }).click();
  await expect.poll(() => ws.sentOf('saveAssistantConfig').at(-1)?.payload).toEqual({ url: 'https://assistant.local/chat' });
  await expect(page.getByText('Address sent to the hub.', { exact: false })).toBeVisible();

  await page.reload();
  await expect(page.getByLabel('Assistant URL', { exact: true })).toHaveValue('https://assistant.local/chat');

  // Restore the fixture value.
  await page.getByLabel('Assistant URL', { exact: true }).fill('http://localhost:4001/api/conversation');
  await page.getByRole('button', { name: 'Save address' }).click();
  await expect.poll(() => ws.sentOf('saveAssistantConfig').at(-1)?.payload).toEqual({ url: 'http://localhost:4001/api/conversation' });
});

test('assistant field is labelled, keyboard reachable and inert while disconnected', async ({ page }) => {
  let disconnected = false;
  let disconnect: (() => Promise<void>) | undefined;
  await page.routeWebSocket('ws://localhost:4110/ws', (socket) => {
    if (disconnected) { void socket.close(); return; }
    socket.connectToServer(); disconnect = () => socket.close();
  });
  await signIn(page, '/settings?section=assistant');
  const url = page.getByLabel('Assistant URL', { exact: true });
  await url.focus();
  await expect(url).toBeFocused();
  disconnected = true;
  await disconnect?.();
  await expect(url).toBeDisabled();
  await expect(page.getByRole('button', { name: 'Save address' })).toBeDisabled();
});
