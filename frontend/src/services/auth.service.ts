import { createAuthSession, createNotAuthenticatedSession } from '@/contracts/auth';
import { store } from '@/store';
import { AuthResponse, UserSession } from '@/types/auth.type';

const baseUrl = import.meta.env.VITE_API_BASE_URL;

export async function login(): Promise<UserSession> {
  const res = await fetch(`${baseUrl}/api/auth/login`, {
    method: 'POST',
    credentials: 'include',
  });

  const data = await res.json();
  console.log('Opening popup for oauth url: ', data);
  const popup = window.open(data.url, 'oauth', 'width=500,height=600');
  if (!popup) {
    console.error('Popup blocked');
    return createNotAuthenticatedSession();
  }

  return new Promise<UserSession>((resolve) => {
    const handler = async (event: MessageEvent) => {
      console.log('Received message from popup: ', event.origin, window.origin);
      const allowedOrigins = ['http://localhost:4100', 'http://localhost:4110'];
      if (!allowedOrigins.includes(event.origin)) return;

      //if (event.origin !== window.origin) return;

      if (event.data.status === 'success') {
        console.log('OAuth success, fetching user info');
        const res = await fetch(`${baseUrl}/api/auth/me`, {
          method: 'GET',
          credentials: 'include',
        });

        popup.close();
        window.removeEventListener('message', handler);

        console.log('Response from me: ', res);
        if (!res.ok) {
          console.error('Failed to fetch user info');
          resolve(createNotAuthenticatedSession());
          return;
        }

        const userData = await res.json();
        console.log('Fetched user info from me ', userData);

        if (!userData.id || !userData.username) {
          resolve(createNotAuthenticatedSession());
          return;
        }

        resolve(createAuthSession(userData));
      }
    };

    window.addEventListener('message', handler);
  });
}


export async function logout(): Promise<boolean> {
  const res = await fetch(`${baseUrl}/api/auth/logout`, {
    method: 'POST',
    credentials: 'include',
  });

  if (!res.ok) {
    console.error('Logout failed');
  }

  return res.ok;
}
