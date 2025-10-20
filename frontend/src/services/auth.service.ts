import { createAuthSession, createNotAuthenticatedSession } from '@/contracts/auth';
import { store } from '@/store';
import { UserSession } from '@/types/auth.type';

const baseUrl = import.meta.env.VITE_API_BASE_URL;


export async function login(): Promise<UserSession> {
  // Get OAuth login URL from backend
  const res = await fetch(`${baseUrl}/api/auth/login`, { method: 'POST', credentials: 'include' });
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
        const token: string | undefined = event.data.token;

        window.removeEventListener('message', handler);
        popup.close();

        try {
          // Give browser a moment to persist the auth cookie
          // await wait(200);

          const headers: HeadersInit = token ? { Authorization: `Bearer ${token}` } : {};
          const meRes = await fetch(`${baseUrl}/api/auth/me`, {
            method: 'GET',
            headers,
          });

          if (!meRes.ok) {
            console.error('/me fetch failed', meRes.status);
            resolve(createNotAuthenticatedSession());
            return;
          }

          const userData = await meRes.json();
          we dont get thetoken and te object is not mapped correctly
          console.log('Fetched user info from me', userData);
          if (!userData.id || !userData.username) {
            resolve(createNotAuthenticatedSession());
            return;
          }

          resolve(createAuthSession(userData));
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
    const res = await fetch(`${baseUrl}/api/auth/logout`, {
      method: 'POST',
      credentials: 'include',
    });
    if (!res.ok) console.error('Logout failed', res.status);
    return res.ok;
  } catch (err) {
    console.error('Logout request error', err);
    return false;
  }
}

// export async function authFetch(input: RequestInfo | URL, init: RequestInit = {}) {
//   const headers = new Headers(init.headers || {});
//   store.
//   const token = localStorage.getItem('auth_token');
//   if (token) headers.set('Authorization', `Bearer ${token}`);
//   return fetch(input, { ...init, headers });
// }
