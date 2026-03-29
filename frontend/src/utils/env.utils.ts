import { getEnv } from '@/config/env';

export function getAllowedOrigins(): string[] {
  const envOrigins = getEnv('VITE_API_BASE_URL');

  let allowed: string[] = [];
  if (envOrigins) {
    allowed = envOrigins
      .split(',')
      .map((s: string) => s.trim())
      .filter(Boolean)
      .map((s: string) => {
        try {
          return new URL(s, window.location.href).origin;
        } catch {
          return null;
        }
      })
      .filter((o): o is string => !!o);
  }

  // Include current page origin as safe default
  if (typeof window !== 'undefined' && !allowed.includes(window.location.origin)) {
    allowed.push(window.location.origin);
  }

  return Array.from(new Set(allowed));
}
