import { get, post } from '@/contracts/api';
import { createAuthSession, createNotAuthenticatedSession } from '@/contracts/auth';
import { store } from '@/store';
import { UserSession } from '@/types/auth.type';

const baseUrl = import.meta.env.VITE_API_BASE_URL;

export async function login(): Promise<UserSession> {
  // Get OAuth login URL from backend
  const res = await get(`api/auth/login`);
  const { url } = await res.json();

  // Open OAuth popup
  const popup = window.open(url, 'oauth', 'width=500,height=600');
  if (!popup) {
    console.error('Popup blocked');
    return createNotAuthenticatedSession();
  }

  // Wait for popup message
  return new Promise<UserSession>((resolve) => {
    const handler = async (event: MessageEvent) => {
      const allowedOrigins = ['http://localhost:4100', 'http://localhost:4110']; // local dev
      if (!allowedOrigins.includes(event.origin)) return;

      if (event.data.status === 'success') {
        console.log('Received auth success from popup', event.data);
        window.removeEventListener('message', handler);
        popup.close();

        try {
          // Token is now in cookie - just fetch user info
          const meRes = await get(`api/auth/me`);

          if (!meRes.ok) {
            console.error('/me fetch failed', meRes.status);
            resolve(createNotAuthenticatedSession());
            return;
          }

          const userData = await meRes.json();
          console.log('Fetched user info from /me', userData);
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
      }
    };

    window.addEventListener('message', handler);
  });
}

export async function logout(): Promise<boolean> {
  try {
    const res = await post(`$api/auth/logout`);
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
