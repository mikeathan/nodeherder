import { describe, test, expect, beforeEach, afterEach, jest } from '@jest/globals';
import { createStore } from 'vuex';
import { AuthModule } from '@/store/modules/auth';

describe('auth session validation', () => {
  const browser = globalThis as unknown as { fetch: unknown };
  let previous: unknown;
  const session = { user: { id: 'local', username: 'Offline User' }, isAuthenticated: true };
  const makeStore = () => createStore({ modules: { auth: AuthModule } });
  const respond = (status: number, body: unknown) => ({ status, ok: status === 200, json: async () => body });

  beforeEach(() => { previous = browser.fetch; });
  afterEach(() => { browser.fetch = previous; });

  test.each([500, 503])('HTTP %i preserves existing identity and reports unavailable', async (status) => {
    const store = makeStore();
    await store.dispatch('auth/loginUser', session);
    browser.fetch = jest.fn(async () => respond(status, {}));
    await store.dispatch('auth/restoreSession');
    expect(store.getters['auth/isAuthenticated']()).toBe(true);
    expect(store.getters['auth/user']()).toEqual(session.user);
    expect(store.getters['auth/validation']).toBe('unavailable');
  });

  test('network failure preserves identity; retry recovers; 401 clears it', async () => {
    const store = makeStore();
    await store.dispatch('auth/loginUser', session);
    browser.fetch = jest.fn(async () => { throw new TypeError('offline'); });
    await store.dispatch('auth/restoreSession');
    expect(store.getters['auth/isAuthenticated']()).toBe(true);
    expect(store.getters['auth/validation']).toBe('unavailable');
    browser.fetch = jest.fn(async () => respond(200, session.user));
    await store.dispatch('auth/restoreSession');
    expect(store.getters['auth/validation']).toBe('authenticated');
    browser.fetch = jest.fn(async () => respond(401, {}));
    await store.dispatch('auth/restoreSession');
    expect(store.getters['auth/isAuthenticated']()).toBe(false);
    expect(store.getters['auth/validation']).toBe('signedOut');
  });

  test('unavailable validation cannot sign in an anonymous browser', async () => {
    const store = makeStore();
    browser.fetch = jest.fn(async () => respond(503, {}));
    await store.dispatch('auth/restoreSession');
    expect(store.getters['auth/isAuthenticated']()).toBe(false);
    expect(store.getters['auth/validation']).toBe('unavailable');
  });

  test('concurrent checks share one request and a late success cannot undo sign-out', async () => {
    const store = makeStore();
    await store.dispatch('auth/loginUser', session);
    let finish!: (response: ReturnType<typeof respond>) => void;
    const fetch = jest.fn(() => new Promise<ReturnType<typeof respond>>((resolve) => { finish = resolve; }));
    browser.fetch = fetch;
    const first = store.dispatch('auth/restoreSession');
    const second = store.dispatch('auth/restoreSession');
    expect(fetch).toHaveBeenCalledTimes(1);
    expect(store.getters['auth/validation']).toBe('checking');
    await store.dispatch('auth/logoutUser');
    finish(respond(200, session.user));
    await Promise.all([first, second]);
    expect(store.getters['auth/isAuthenticated']()).toBe(false);
    expect(store.getters['auth/validation']).toBe('signedOut');
  });

  test('late rejection cannot clear a newer login', async () => {
    const store = makeStore();
    let finish!: (response: ReturnType<typeof respond>) => void;
    browser.fetch = jest.fn(() => new Promise<ReturnType<typeof respond>>((resolve) => { finish = resolve; }));
    const check = store.dispatch('auth/restoreSession');
    await store.dispatch('auth/loginUser', session);
    finish(respond(401, {}));
    await check;
    expect(store.getters['auth/isAuthenticated']()).toBe(true);
  });
});
