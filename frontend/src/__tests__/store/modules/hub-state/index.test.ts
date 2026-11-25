import { beforeEach, describe, expect, test } from '@jest/globals';
import { store } from '@/store/index';
import { default as hubState } from '../../../../../../docs/hub_state.json';

const snapshot = JSON.parse(JSON.stringify((hubState as { payload: { devices: unknown[]; config: unknown } }).payload));

describe('hub data-sync timestamp', () => {
  beforeEach(() => {
    store.commit('hub/clear');
  });

  test('is recorded by init and visible through the getter', () => {
    expect(store.getters['hub/lastSyncedAt']()).toBeNull();
    store.dispatch('hub/init', snapshot);
    expect(store.getters['hub/lastSyncedAt']()).toEqual(expect.any(Number));
  });

  test('is cleared with the snapshot so a logged-out shell cannot claim fresh data', () => {
    store.dispatch('hub/init', snapshot);
    store.commit('hub/resetSnapshot');
    expect(store.getters['hub/lastSyncedAt']()).toBeNull();
  });
});
