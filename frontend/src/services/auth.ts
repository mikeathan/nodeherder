import { createAuthSession, createNotAuthenticatedSession } from '@/contracts/auth';
import { UserSession } from '@/types/auth.type';

const baseUrl = import.meta.env.VITE_API_BASE_URL;

wip
async function login(): Promise<UserSession> {
  const username = 'testuser';
  const res = await fetch(`${baseUrl}/api/auth/login?username=${username}`);
  if (!res.ok) {
    console.error('Login failed');
    return createNotAuthenticatedSession();
  }
  const data = await res.json();
  return createAuthSession(data, data.user);
}
