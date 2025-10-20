import { createAuthSession, createNotAuthenticatedSession } from '@/contracts/auth';
import { store } from '@/store';
import { UserSession } from '@/types/auth.type';

const baseUrl = import.meta.env.VITE_API_BASE_URL;

// Utility to wait a bit for the cookie to persist
const wait = (ms: number) => new Promise((resolve) => setTimeout(resolve, ms));

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
        window.removeEventListener('message', handler);
        popup.close();

        try {
          // Give browser a moment to persist the auth cookie
          await wait(200);

          const meRes = await fetch(`${baseUrl}/api/auth/me`, {
            method: 'GET',
            credentials: 'include',
          });

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
