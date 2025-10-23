import { get, post } from '@/contracts/api';
import { createAuthSession, createNotAuthenticatedSession } from '@/contracts/auth';
import { store } from '@/store';
import { UserSession } from '@/types/auth.type';

const baseUrl = import.meta.env.VITE_API_BASE_URL;

export async function getOAuthUrl(): Promise<string> {
  const res = await post(`api/auth/login`);
  const { url } = await res.json();
  return url;
}

export async function waitForOAuthCompletion(): Promise<UserSession> {
  return new Promise<UserSession>((resolve, reject) => {
    const handler = async (event: MessageEvent) => {
      const allowedOrigins = ['http://localhost:4100', 'http://localhost:4110']; // local dev
      if (!allowedOrigins.includes(event.origin)) {
        return;
      }

      if (event.data.status === 'success') {
        window.removeEventListener('message', handler);

        try {
          const meRes = await get(`api/auth/me`);

          if (!meRes.ok) {
            console.error('/me fetch failed', meRes.status);
            resolve(createNotAuthenticatedSession());
            return;
          }

          const userData = await meRes.json();
          if (!userData.id || !userData.username) {
            resolve(createNotAuthenticatedSession());
            return;
          }

          const userSession = createAuthSession(userData);
          store.dispatch('auth/loginUser', userSession);
          resolve(userSession);
        } catch (err) {
          console.error('Error fetching user info:', err);
          resolve(createNotAuthenticatedSession());
        }
      } else if (event.data.status === 'error') {
        window.removeEventListener('message', handler);
        reject(new Error(event.data.message || 'OAuth failed'));
      }
    };

    window.addEventListener('message', handler);
  });
}

export async function login(): Promise<UserSession> {
  // Keep the old method for backward compatibility but don't auto-open popup
  return waitForOAuthCompletion();
}

export async function logout(): Promise<boolean> {
  try {
    const res = await post(`api/auth/logout`);
    if (!res.ok) console.error('Logout failed', res.status);
    return res.ok;
  } catch (err) {
    console.error('Logout request error', err);
    return false;
  }
}

export async function restoreSession(): Promise<UserSession | null> {
  try {
    const res = await get(`api/auth/me`);

    if (res.ok) {
      const userData = await res.json();
      if (userData.id && userData.username) {
        return createAuthSession(userData);
      }
    }

    return null;
  } catch (err) {
    console.error('Failed to restore session:', err);
    return null;
  }
}
