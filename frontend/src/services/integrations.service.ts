import { fetchWithAuth } from '@/contracts/api';
import {
  object,
  parseFailure,
  parseCatalog,
  parseDefinition,
  parseInstance,
  parseInstances,
  parseOperation,
  parseSetup,
  parseValidation,
} from '@/contracts/integrations';
import type {
  ConnectionFailure,
  ConnectionSetup,
  ConnectionInstance,
  DestinationReview,
  ProviderDefinition,
} from '@/types/integrations.type';
export class ConnectionError extends Error {
  constructor(
    public readonly failure: ConnectionFailure,
    public readonly status = 400
  ) {
    super(failure.message);
  }
}
const segment = (s: string) => encodeURIComponent(s);
const providerPath = (key: string) => `/definitions/${key.split('/').map(segment).join('/')}`;
async function call<T>(
  path: string,
  parse: (value: unknown) => T,
  method = 'GET',
  body?: unknown,
  etag?: string,
  signal?: AbortSignal,
  create = false
): Promise<T> {
  let response: Response;
  try {
    response = await fetchWithAuth(`/integrations${path}`, {
      method,
      signal,
      headers: {
        'X-NodeHerder-Request': '1',
        ...(etag ? { 'If-Match': etag } : {}),
        ...(create ? { 'If-None-Match': '*' } : {}),
      },
      ...(body === undefined ? {} : { body: JSON.stringify(body) }),
    });
  } catch (e) {
    if (e instanceof Error && e.name === 'AbortError') throw e;
    throw new ConnectionError(
      { code: 'offline', message: 'NodeHerder could not be reached. Your unsaved changes are still here.' },
      0
    );
  }
  if (response.status === 401)
    throw new ConnectionError({ code: 'unauthorized', message: 'Sign in again to use Connections.' }, 401);
  let envelope: Record<string, unknown>;
  try {
    envelope = object(await response.json());
  } catch {
    throw new ConnectionError({ code: 'invalid_response', message: 'NodeHerder returned an unreadable response.' });
  }
  if (!response.ok) {
    let failure: ConnectionFailure | null = null;
    try {
      failure = parseFailure(envelope.error);
    } catch {
      /* Existing transport limits may precede the feature envelope. */
    }
    throw new ConnectionError(
      failure ?? { code: 'request_failed', message: `The request failed (${response.status}).` },
      response.status
    );
  }
  if (envelope.error !== null)
    throw new ConnectionError({ code: 'invalid_response', message: 'NodeHerder returned an invalid response.' });
  return parse(envelope.data);
}
/** Binds every request (including mutations) to the owning page/session lifetime. */
export function createConnectionsApi(defaultSignal?: AbortSignal) {
  function request<T>(
    path: string,
    parse: (value: unknown) => T,
    method = 'GET',
    body?: unknown,
    etag?: string,
    signal?: AbortSignal,
    create = false
  ) {
    return call(path, parse, method, body, etag, signal ?? defaultSignal, create);
  }
  return {
    reloadProviders: () => request('/definitions/reload', parseCatalog, 'POST', {}),
    catalog: (signal?: AbortSignal) => request('/catalog', parseCatalog, 'GET', undefined, undefined, signal),
    instances: (signal?: AbortSignal) => request('/instances', parseInstances, 'GET', undefined, undefined, signal),
    instance: (id: string, signal?: AbortSignal) =>
      request(`/instances/${segment(id)}`, parseInstance, 'GET', undefined, undefined, signal),
    setup: (id: string, signal?: AbortSignal) =>
      request(`/setup/${segment(id)}`, parseSetup, 'GET', undefined, undefined, signal),
    createSetup: (provider: string, target?: { id: string; etag: string }) =>
      request('/setup', parseSetup, 'POST', {
        provider_key: provider,
        target_instance_id: target?.id,
        target_etag: target?.etag,
      }),
    step: (id: string, etag: string, body: unknown) =>
      request(`/setup/${segment(id)}/step`, parseSetup, 'POST', body, etag),
    cancel: (id: string, etag: string) => request(`/setup/${segment(id)}`, object, 'DELETE', undefined, etag),
    refresh: (id: string, etag: string) =>
      request(`/instances/${segment(id)}/refresh`, parseOperation, 'POST', {}, etag),
    check: (id: string, etag: string) =>
      request(`/operations/${segment(id)}/check-status`, parseOperation, 'POST', {}, etag),
    operation: (id: string, signal?: AbortSignal) =>
      request(`/operations/${segment(id)}`, parseOperation, 'GET', undefined, undefined, signal),
    update: (id: string, etag: string, name: string, enabled: boolean) =>
      request(`/instances/${segment(id)}`, parseInstance, 'PUT', { name, enabled }, etag),
    remove: (id: string, etag: string) => request(`/instances/${segment(id)}`, object, 'DELETE', undefined, etag),
    credential: (kind: 'setup' | 'instances', id: string, slot: string, etag: string, value: string, remove = false) =>
      request<ConnectionSetup | ConnectionInstance>(
        `/${kind}/${segment(id)}/credentials/${segment(slot)}`,
        kind === 'setup' ? (v) => parseSetup(v) : (v) => parseInstance(v),
        remove ? 'DELETE' : 'PUT',
        remove ? {} : { value },
        etag
      ),
    definition: (key: string, signal?: AbortSignal) =>
      request(providerPath(key), parseDefinition, 'GET', undefined, undefined, signal),
    validate: (key: string, body: { yaml: string } | { document: ProviderDefinition }) =>
      request(`${providerPath(key)}/validate`, parseValidation, 'POST', body),
    saveDefinition: (key: string, yaml: string, etag: string, review?: DestinationReview, create = false) =>
      request(providerPath(key), parseDefinition, 'PUT', { yaml, destination_review: review }, etag, undefined, create),
    reload: (key: string, etag: string, review?: DestinationReview) =>
      request(`${providerPath(key)}/reload`, parseDefinition, 'POST', { destination_review: review }, etag),
    restore: (key: string, etag: string, review?: DestinationReview) =>
      request(`${providerPath(key)}/restore`, parseDefinition, 'POST', { destination_review: review }, etag),
    test: (
      key: string,
      yaml: string,
      instance: string,
      mode: 'offline' | 'live',
      sample?: unknown,
      review?: DestinationReview
    ) =>
      request(`${providerPath(key)}/test`, parseSetup, 'POST', {
        yaml,
        instance_id: instance,
        mode,
        sample,
        destination_review: review,
      }),
    acknowledgeShipped: (key: string, etag: string) =>
      request(`${providerPath(key)}/shipped-reviewed`, parseDefinition, 'POST', {}, etag),
    deleteDefinition: (key: string, etag: string) => request(providerPath(key), object, 'DELETE', {}, etag),
  };
}

export const connectionsApi = createConnectionsApi();
