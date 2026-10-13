import { get, post } from '@/contracts/api';
import { createAuthSession } from '@/contracts/auth';
import { UserSession } from '@/types/auth.type';

export type AuthMode = 'local' | 'google';

async function authGet(endpoint: string): Promise<{ status: number; ok: boolean; data: unknown }> {
  const controller = new AbortController();
  const timeout = setTimeout(() => controller.abort(), 8000);
  try {
    const response = await get(endpoint, { signal: controller.signal, cache: 'no-store' });
    const data: unknown = response.ok ? await response.json() : null;
    return { status: response.status, ok: response.ok, data };
  } finally {
    clearTimeout(timeout);
  }
}

export async function getAuthMode(): Promise<AuthMode> {
  const res = await authGet('auth/mode');
  if (!res.ok) throw new Error('Could not check sign-in mode');
  const data = res.data;
  if (typeof data === 'object' && data !== null && 'mode' in data && (data.mode === 'local' || data.mode === 'google')) {
    return data.mode;
  }
  throw new Error('Invalid sign-in mode response');
}

export async function getOAuthUrl(): Promise<string> {
  const res = await post(`auth/login`);
  const { url } = await res.json();
  return url;
}

export async function logout(): Promise<boolean> {
  try {
    const res = await post(`auth/logout`);
    if (!res.ok) console.error('Logout failed', res.status);
    return res.ok;
  } catch (err) {
    console.error('Logout request error', err);
    return false;
  }
}

export async function restoreSession(): Promise<UserSession | null> {
  const res = await authGet('auth/me');
  if (res.status === 401) return null;
  if (!res.ok) throw new Error(`Session check failed (${res.status})`);
  const data = res.data;
  if (typeof data === 'object' && data !== null && 'id' in data && 'username' in data &&
    typeof data.id === 'string' && data.id !== '' && typeof data.username === 'string' && data.username !== '') {
    return createAuthSession({ id: data.id, username: data.username });
  }
  throw new Error('Invalid session response');
}
