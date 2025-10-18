import { createAuthSession, createNotAuthenticatedSession } from '@/contracts/auth';
import { AuthResponse, UserSession } from '@/types/auth.type';

const baseUrl = import.meta.env.VITE_API_BASE_URL;

export async function login(): Promise<UserSession> {
  const res = await fetch(`${baseUrl}/api/auth/login`, {
    method: 'POST',
  });
  if (!res.ok) {
    console.error('Login failed');
    return createNotAuthenticatedSession();
  }
  const data: AuthResponse = await res.json();
  if (!data.user || !data.token) {
    console.error('Invalid login response', data);
    return createNotAuthenticatedSession();
  }

  return createAuthSession(data);
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
