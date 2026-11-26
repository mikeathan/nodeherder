import { afterEach, beforeEach, describe, expect, test, jest } from '@jest/globals';
import { effectScope, type EffectScope } from 'vue';
import { store } from '@/store/index';
import { OUTAGE_GRACE_MS, useNotificationProducers } from '@/composables/useNotificationProducers';

describe('connection outage notifications', () => {
  let scope: EffectScope;

  beforeEach(() => {
    jest.useFakeTimers();
    store.commit('notifications/clear');
    store.commit('ws/setConnectionStatus', 'connected');
    scope = effectScope();
    scope.run(() => useNotificationProducers());
  });

  afterEach(() => {
    scope.stop();
    jest.useRealTimers();
  });

  test('a restart shorter than the grace period produces no notification', () => {
    store.commit('ws/setConnectionStatus', 'reconnecting');
    jest.advanceTimersByTime(OUTAGE_GRACE_MS - 1_000);
    store.commit('ws/setConnectionStatus', 'connected');
    expect(store.getters['notifications/items']()).toEqual([]);
  });

  test('a sustained outage produces one item that resolves in place', () => {
    store.commit('ws/setConnectionStatus', 'reconnecting');
    jest.advanceTimersByTime(OUTAGE_GRACE_MS + 1_000);
    const lost = store.getters['notifications/items']();
    expect(lost).toHaveLength(1);
    expect(lost[0]).toMatchObject({ title: 'Hub connection lost', severity: 'warn' });

    jest.advanceTimersByTime(3 * 60_000);
    store.commit('ws/setConnectionStatus', 'connected');
    const resolved = store.getters['notifications/items']();
    expect(resolved).toHaveLength(1);
    expect(resolved[0]).toMatchObject({ id: lost[0].id, title: 'Hub connection restored', severity: 'success' });
    expect(resolved[0].body).toContain('4 min');
  });

  test('a second unread outage reuses the single lifecycle row instead of adding a pair', () => {
    store.commit('ws/setConnectionStatus', 'reconnecting');
    jest.advanceTimersByTime(OUTAGE_GRACE_MS + 1_000);
    store.commit('ws/setConnectionStatus', 'connected');
    store.commit('ws/setConnectionStatus', 'reconnecting');
    jest.advanceTimersByTime(OUTAGE_GRACE_MS + 1_000);
    store.commit('ws/setConnectionStatus', 'connected');
    const items = store.getters['notifications/items']();
    expect(items).toHaveLength(1);
    expect(items[0].title).toBe('Hub connection restored');
  });
});
