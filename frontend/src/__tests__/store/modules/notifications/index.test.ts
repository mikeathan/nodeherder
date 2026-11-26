import { beforeEach, describe, expect, test } from '@jest/globals';
import { store } from '@/store/index';
import type { NotificationInput } from '@/store/modules/notifications';

const input = (overrides: Partial<NotificationInput> = {}): NotificationInput => ({
  severity: 'error',
  category: 'connection',
  title: 'Hub connection lost',
  ...overrides,
});

describe('notification inbox', () => {
  beforeEach(() => {
    store.dispatch('notifications/clear');
  });

  test('records an unread item with an id and timestamp', () => {
    const item = store.dispatch('notifications/notify', input());
    const items = store.getters['notifications/items']();
    expect(items).toHaveLength(1);
    expect(items[0]).toMatchObject({ title: 'Hub connection lost', read: false });
    expect(items[0].id).toEqual(expect.any(String));
  });

  test('deduplicates an unread condition but allows it to reopen after being read', async () => {
    await store.dispatch('notifications/notify', input({ dedupeKey: 'ws-lost' }));
    await store.dispatch('notifications/notify', input({ dedupeKey: 'ws-lost' }));
    const items = store.getters['notifications/items']();
    expect(items).toHaveLength(1);

    store.dispatch('notifications/markAllRead');
    await store.dispatch('notifications/notify', input({ dedupeKey: 'ws-lost' }));
    expect(store.getters['notifications/items']()).toHaveLength(2);
  });

  test('marks one or all read and clears', () => {
    store.dispatch('notifications/notify', input({ dedupeKey: 'a' }));
    store.dispatch('notifications/notify', input({ dedupeKey: 'b' }));
    expect(store.getters['notifications/unreadCount']()).toBe(2);

    const [newest] = store.getters['notifications/items']();
    store.dispatch('notifications/markRead', newest.id);
    expect(store.getters['notifications/unreadCount']()).toBe(1);

    store.dispatch('notifications/markAllRead');
    expect(store.getters['notifications/unreadCount']()).toBe(0);

    store.dispatch('notifications/clear');
    expect(store.getters['notifications/items']()).toEqual([]);
  });

  test('resolving a connection outage updates one item and preserves read state', async () => {
    const item = await store.dispatch('notifications/notify', input({ dedupeKey: 'hub-connection', title: 'Hub connection lost' }));
    store.dispatch('notifications/markRead', item.id);
    store.dispatch('notifications/resolveConnection', 4 * 60_000);
    const items = store.getters['notifications/items']();
    expect(items).toHaveLength(1);
    expect(items[0]).toMatchObject({ title: 'Hub connection restored', read: true, severity: 'success' });
    expect(items[0].body).toContain('4 min');
  });

  test('bounds the inbox and returns newest first', () => {
    for (let index = 0; index < 205; index += 1) {
      store.dispatch('notifications/notify', input({ title: `event ${index}`, at: Date.now() + index }));
    }
    const items = store.getters['notifications/items']();
    expect(items).toHaveLength(200);
    expect(items[0].title).toBe('event 204');
  });
});
