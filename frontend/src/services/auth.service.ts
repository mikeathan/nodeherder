import { get, post } from '@/contracts/api';
import { createAuthSession } from '@/contracts/auth';
import { UserSession } from '@/types/auth.type';


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
  try {
    const res = await get(`auth/me`);

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
