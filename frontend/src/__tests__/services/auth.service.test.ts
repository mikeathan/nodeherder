import { describe, expect, test, beforeEach, afterEach, jest } from '@jest/globals';
import { getOAuthUrl, logout, restoreSession } from '@/services/auth.service';

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

  test('restoreSession returns null for 401 or incomplete user', async () => {
    mockFetch(401, { error: 'no' });
    await expect(restoreSession()).resolves.toBeNull();
    mockFetch(200, { id: '1' });
    await expect(restoreSession()).resolves.toBeNull();
  });

  test('logout posts to auth/logout and reports success', async () => {
    mockFetch(200, { status: 'ok' });
    await expect(logout()).resolves.toBe(true);
    expect(calls[0].url).toBe('http://localhost:4001/auth/logout');
    mockFetch(500, {});
    await expect(logout()).resolves.toBe(false);
  });
});
