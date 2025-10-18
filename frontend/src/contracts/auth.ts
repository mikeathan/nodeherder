import { AuthResponse, User, UserSession } from '@/types/auth.type';
import { jwtDecode } from 'jwt-decode';

export function isTokenExpired(token: string): boolean {
  try {
    if (!token) return true;
    const decoded: { exp?: number } = jwtDecode(token);
    if (!decoded.exp) return true;
    return decoded.exp * 1000 < Date.now();
  } catch (error) {
    console.error('Error decoding token:', error);
    return true;
  }
}

export function createAuthSession(authResponse: AuthResponse): UserSession {
  return {
    ...authResponse,
    isAuthenticated: true,
  };
}

export function createNotAuthenticatedSession(): UserSession {
  return {
    token: '',
    user: { id: '', username: '' },
    isAuthenticated: false,
  };
}
