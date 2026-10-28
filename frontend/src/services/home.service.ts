import { fetchWithAuth, type FetchOptions } from '@/contracts/api';
import { parseHomeLayout } from '@/contracts/home';
import type { HomeFailure, HomeLayout } from '@/types/home.type';

/** Safe transport failure with no raw provider response or imported payload. */
export class HomeError extends Error {
  constructor(
    public readonly failure: HomeFailure,
    public readonly status = 0
  ) {
    super(failure.message);
  }
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return value !== null && typeof value === 'object' && !Array.isArray(value);
}

function requestOptions(signal: AbortSignal, proposal?: HomeLayout): FetchOptions {
  const headers: Record<string, string> = { 'X-NodeHerder-Request': '1' };
  const options: FetchOptions = { signal, method: 'GET', headers };
  if (proposal) {
    options.method = 'PUT';
    options.body = JSON.stringify(proposal);
    headers['If-Match'] = proposal.etag;
  }
  return options;
}

function failureFromEnvelope(envelope: unknown): HomeFailure {
  const fallback = { code: 'request_failed', message: 'Home could not be loaded or saved.' };
  if (!isRecord(envelope) || !isRecord(envelope.error)) return fallback;
  const { code, message } = envelope.error;
  if (typeof code !== 'string' || typeof message !== 'string') return fallback;
  return { code, message };
}

async function decodeLayoutResponse(response: Response): Promise<HomeLayout> {
  let envelope: unknown;
  try {
    envelope = await response.json();
  } catch {
    throw new HomeError(
      { code: 'invalid_response', message: 'Home returned an unreadable response.' },
      response.status
    );
  }
  if (!response.ok) throw new HomeError(failureFromEnvelope(envelope), response.status);
  if (!isRecord(envelope) || envelope.error !== null) {
    throw new HomeError({ code: 'invalid_response', message: 'Home returned an invalid response.' });
  }
  try {
    return parseHomeLayout(envelope.data);
  } catch {
    throw new HomeError({
      code: 'invalid_response',
      message: 'Home returned an invalid layout. Existing areas are kept.',
    });
  }
}

async function requestLayout(signal: AbortSignal, proposal?: HomeLayout): Promise<HomeLayout> {
  let response: Response;
  try {
    response = await fetchWithAuth('/home/layout', requestOptions(signal, proposal));
  } catch (cause) {
    if (signal.aborted || (isRecord(cause) && cause.name === 'AbortError')) throw cause;
    throw new HomeError({ code: 'offline', message: 'NodeHerder could not be reached. The accepted layout is kept.' });
  }
  return decodeLayoutResponse(response);
}

/** Stateless authenticated layout transport. Callers own cancellation; nothing enters browser storage. */
export const homeApi = {
  /** Read local accepted placement state without initiating provider work. */
  read(signal: AbortSignal): Promise<HomeLayout> {
    return requestLayout(signal);
  },
  /** Save a complete proposal against its accepted ETag; never retry a conflict silently. */
  save(proposal: HomeLayout, signal: AbortSignal): Promise<HomeLayout> {
    return requestLayout(signal, proposal);
  },
};
