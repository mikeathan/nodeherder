import { describe, expect, test, beforeEach, afterEach, jest } from '@jest/globals';
import { getAuthMode, getOAuthUrl, logout, restoreSession } from '@/services/auth.service';

// Sign-in contract (spec 007 FR-11): POST auth/login -> { url }; GET auth/me -> user; POST auth/logout.
type Call = { url: string; init: RequestInit };

describe('auth service', () => {
  const calls: Call[] = [];
  const g = globalThis as unknown as { fetch: unknown };
  let previous: unknown;
  const respond = (status: number, body: unknown) => ({ ok: status < 400, status, json: async () => body });

  beforeEach(() => {
    calls.length = 0;
    previous = g.fetch;
    jest.spyOn(console, 'error').mockImplementation(() => {});
  });
  afterEach(() => {
    g.fetch = previous;
    jest.restoreAllMocks();
    jest.useRealTimers();
  });
  const mockFetch = (status: number, body: unknown) => {
    g.fetch = jest.fn(async (url: string, init: RequestInit) => {
      calls.push({ url, init });
      return respond(status, body);
    });
  };

  test('getOAuthUrl posts to auth/login with credentials and returns the url', async () => {
    mockFetch(200, { url: 'https://accounts.example/auth' });
    await expect(getOAuthUrl()).resolves.toBe('https://accounts.example/auth');
    expect(calls[0].url).toBe('http://localhost:4001/auth/login');
    expect(calls[0].init.method).toBe('POST');
    expect(calls[0].init.credentials).toBe('include');
  });

  test('restoreSession returns an authenticated session for a valid user', async () => {
    mockFetch(200, { id: '1', username: 'mike' });
    await expect(restoreSession()).resolves.toEqual({ user: { id: '1', username: 'mike' }, isAuthenticated: true });
    expect(calls[0].url).toBe('http://localhost:4001/auth/me');
  });

  test('restoreSession returns null only when authentication is rejected', async () => {
    mockFetch(401, { error: 'no' });
    await expect(restoreSession()).resolves.toBeNull();
    mockFetch(200, { id: '1' });
    await expect(restoreSession()).rejects.toThrow();
  });

  test.each([403, 500, 503])('restoreSession reports HTTP %i as unavailable', async (status) => {
    mockFetch(status, {});
    await expect(restoreSession()).rejects.toThrow();
  });

  test('network failure is distinguishable from a rejected session', async () => {
    g.fetch = jest.fn(async () => { throw new TypeError('Network unavailable'); });
    await expect(restoreSession()).rejects.toThrow('Network unavailable');
  });

  test('session check aborts after eight seconds and clears its timer', async () => {
    jest.useFakeTimers();
    g.fetch = jest.fn(async (_url: string, init: RequestInit) => new Promise((_resolve, reject) => {
      init.signal?.addEventListener('abort', () => reject(new Error('Request aborted')), { once: true });
    }));
    const request = expect(restoreSession()).rejects.toThrow('Request aborted');
    await jest.advanceTimersByTimeAsync(8000);
    await request;
    expect(jest.getTimerCount()).toBe(0);
  });

  test.each(['local', 'google'])('mode is validated: %s', async (mode) => {
    mockFetch(200, { mode });
    await expect(getAuthMode()).resolves.toBe(mode);
    expect(calls[0].url).toBe('http://localhost:4001/auth/mode');
  });

  test('invalid mode is not guessed', async () => {
    mockFetch(200, { mode: 'unknown' });
    await expect(getAuthMode()).rejects.toThrow();
  });

  test('logout posts to auth/logout and reports success', async () => {
    mockFetch(200, { status: 'ok' });
    await expect(logout()).resolves.toBe(true);
    expect(calls[0].url).toBe('http://localhost:4001/auth/logout');
    mockFetch(500, {});
    await expect(logout()).resolves.toBe(false);
  });
});
