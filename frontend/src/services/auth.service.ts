import { createAuthSession, createNotAuthenticatedSession } from '@/contracts/auth';
import { store } from '@/store';
import { AuthResponse, UserSession } from '@/types/auth.type';

const baseUrl = import.meta.env.VITE_API_BASE_URL;

export async function login(): Promise<UserSession> {
  const res = await fetch(`${baseUrl}/api/auth/login`, {
    method: 'POST',
    credentials: 'include',
  });

  const data =  await res.json();
  console.log('Opening popup for oauth url: ', data);
  const popup = window.open(data.url, 'oauth', 'width=500,height=600');

  window.addEventListener('message', async (event) => {
    if (event.origin !== window.origin) return;
    if (event.data.status === 'success') {
      const res = await fetch(`${baseUrl}/api/auth/me`, {
        method: 'GET',
        credentials: 'include',
      });
      console.log('Fetched oauthcallback after login ', res.ok);
      popup?.close();

      if (!res.ok) {
        console.error('Failed to fetch user info');
        return createNotAuthenticatedSession();
      }

      const data = await res.json();
      console.log('Fetched user info from me ', data);

      if (!data.user || !data.token) {
        console.error('Invalid login response', data);
        return createNotAuthenticatedSession();
      }
    }
  });


  return createNotAuthenticatedSession();
}


// const handler = async (event: MessageEvent) => {
//     if (event.origin !== window.origin) return;
//     if (event.data.status === 'success') {
//         const res = await fetch(`${baseUrl}/api/auth/me`, {
//             method: 'GET',
//             credentials: 'include',
//         });
//         popup?.close();
//         window.removeEventListener('message', handler); // remove after first use
//         if (!res.ok) return createNotAuthenticatedSession();
//         const data = await res.json();
//         console.log('User info', data);
//     }
// };

// window.addEventListener('message', handler);
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
