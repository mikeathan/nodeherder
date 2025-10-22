import { ApiResponse } from '@/types/api.type';

export function isApiResponse<T>(res: any): res is ApiResponse<T> {
  return typeof res === 'object' && res !== null;
}

const baseUrl = import.meta.env.VITE_API_BASE_URL;

export interface FetchOptions extends RequestInit {
  headers?: HeadersInit;
  includeAuth?: boolean;
}

export async function fetchWithAuth(endpoint: string, options: FetchOptions = {}): Promise<Response> {
  const { includeAuth = true, headers = {}, ...restOptions } = options;

  // Build headers
  const defaultHeaders: HeadersInit = {
    'Content-Type': 'application/json',
  };

  // // Add authorization token if required
  // if (includeAuth) {
  //   const token = store.getters['auth/token']();
  //   if (token) {
  //     Object.assign(defaultHeaders, {
  //       Authorization: `Bearer ${token}`,
  //     });
  //   }
  // }

  // Merge custom headers with defaults
  const mergedHeaders = {
    ...defaultHeaders,
    ...headers,
  };

  const url = endpoint.startsWith('http') ? endpoint : `${baseUrl}${endpoint}`;

  return fetch(url, {
    ...restOptions,
    credentials: 'include',
    headers: mergedHeaders,
  });
}

export async function get(endpoint: string, options?: FetchOptions): Promise<Response> {
  return fetchWithAuth(endpoint, { ...options, method: 'GET' });
}

export async function post(endpoint: string, body?: any, options?: FetchOptions): Promise<Response> {
  return fetchWithAuth(endpoint, {
    ...options,
    method: 'POST',
    body: body ? JSON.stringify(body) : undefined,
  });
}
