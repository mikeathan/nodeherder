import { test, expect, signIn } from './support/fixtures';

// Spec 070 US-02: operation failures survive the toast and are readable after reload.
test('operation failures stay in the inbox and read state persists', async ({ page }) => {
  let send: ((data: string) => void) | undefined;
  await page.routeWebSocket('ws://localhost:4110/ws', (route) => {
    route.connectToServer();
    send = (data) => route.send(data);
  });
  await signIn(page, '/overview');
  await expect.poll(() => typeof send).toBe('function');

  send!(JSON.stringify({ type: 'operationFailed', payload: 'Bridge rejected the request' }));

  const bell = page.getByRole('button', { name: /Notifications, 1 unread/ });
  await expect(bell).toBeVisible();
  await bell.click();
  const dialog = page.getByRole('dialog', { name: 'Notifications' });
  await expect(dialog.getByText('Operation failed')).toBeVisible();
  await expect(dialog.getByText('Bridge rejected the request')).toBeVisible();

  await dialog.getByRole('button', { name: 'Mark all read' }).click();
  await expect(page.getByRole('button', { name: 'Notifications', exact: true })).toBeVisible();
  await dialog.getByRole('button', { name: 'Done' }).click();

  await page.reload();
  await expect(page.getByRole('button', { name: 'Notifications', exact: true })).toBeVisible();
  await page.getByRole('button', { name: 'Notifications', exact: true }).click();
  const reopened = page.getByRole('dialog', { name: 'Notifications' });
  await expect(reopened.getByText('Operation failed')).toBeVisible();
  await expect(reopened.getByText('All caught up')).toBeVisible();
});
