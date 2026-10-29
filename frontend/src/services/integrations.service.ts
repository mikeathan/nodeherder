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
  parseBoard,
} from '@/contracts/integrations';
import type {
  ConnectionFailure,
  ConnectionSetup,
  ConnectionInstance,
  DestinationReview,
  ProviderDefinition,
  ProviderTestMode,
  ConnectionBoardEntry,
} from '@/types/integrations.type';
/** Sanitized API/session failure exposed to UI feedback; never contains a raw provider response. */
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
/** Central auth/envelope/ingress boundary; page aborts stay cancellations rather than offline errors. */
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
    throw new ConnectionError({ code: 'unauthorized', message: 'Sign in again to use Integrations.' }, 401);
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
    /** Read shared placements without addresses or provider I/O. */
    board: (signal?: AbortSignal) => request('/board', parseBoard, 'GET', undefined, undefined, signal),
    /** Save full order/visibility atomically against a membership/preference revision. */
    saveBoard: (entries: ConnectionBoardEntry[], etag: string) =>
      request('/board', parseBoard, 'PUT', { entries }, etag),
    /** Explicitly reload host definitions for discovery and repair. */
    reloadProviders: () => request('/definitions/reload', parseCatalog, 'POST', {}),
    /** Read available types/providers and feature health without starting setup. */
    catalog: (signal?: AbortSignal) => request('/catalog', parseCatalog, 'GET', undefined, undefined, signal),
    /** Read redacted instances and current calendar projections without provider I/O. */
    instances: (signal?: AbortSignal) => request('/instances', parseInstances, 'GET', undefined, undefined, signal),
    /** Read one private instance/resource projection, using the supplied page signal. */
    instance: (id: string, signal?: AbortSignal) =>
      request(`/instances/${segment(id)}`, parseInstance, 'GET', undefined, undefined, signal),
    /** Read shared setup progress; only explicit step/check mutations admit provider work. */
    setup: (id: string, signal?: AbortSignal) =>
      request(`/setup/${segment(id)}`, parseSetup, 'GET', undefined, undefined, signal),
    /** Create a new wizard or revision-fenced replacement for an existing instance. */
    createSetup: (provider: string, target?: { id: string; etag: string }) =>
      request('/setup', parseSetup, 'POST', {
        provider_key: provider,
        target_instance_id: target?.id,
        target_etag: target?.etag,
      }),
    /** Submit an explicit search/save action against the current setup revision. */
    step: (id: string, etag: string, body: unknown) =>
      request(`/setup/${segment(id)}/step`, parseSetup, 'POST', body, etag),
    /** Delete an unfinished setup; remote requests already submitted cannot be recalled. */
    cancel: (id: string, etag: string) => request(`/setup/${segment(id)}`, object, 'DELETE', undefined, etag),
    /** Explicitly admit a provider refresh using the current instance revision. */
    refresh: (id: string, etag: string) =>
      request(`/instances/${segment(id)}/refresh`, parseOperation, 'POST', {}, etag),
    /** Explicitly continue the admitted operation, respecting its revision/backoff rules. */
    check: (id: string, etag: string) =>
      request(`/operations/${segment(id)}/check-status`, parseOperation, 'POST', {}, etag),
    /** Read local progress only; this GET never polls the upstream provider. */
    operation: (id: string, signal?: AbortSignal) =>
      request(`/operations/${segment(id)}`, parseOperation, 'GET', undefined, undefined, signal),
    /** Save name/enabled state without changing provider configuration. */
    update: (id: string, etag: string, name: string, enabled: boolean) =>
      request(`/instances/${segment(id)}`, parseInstance, 'PUT', { name, enabled }, etag),
    /** Delete a configured connection and private data; keep its shared provider definition. */
    remove: (id: string, etag: string) => request(`/instances/${segment(id)}`, object, 'DELETE', undefined, etag),
    /** Write/remove a secret slot; returned DTOs expose only configured status, never its value. */
    credential: (kind: 'setup' | 'instances', id: string, slot: string, etag: string, value: string, remove = false) =>
      request<ConnectionSetup | ConnectionInstance>(
        `/${kind}/${segment(id)}/credentials/${segment(slot)}`,
        kind === 'setup' ? (v) => parseSetup(v) : (v) => parseInstance(v),
        remove ? 'DELETE' : 'PUT',
        remove ? {} : { value },
        etag
      ),
    /** Read active/disk-error host configuration and editor revision metadata. */
    definition: (key: string, signal?: AbortSignal) =>
      request(providerPath(key), parseDefinition, 'GET', undefined, undefined, signal),
    /** Validate YAML or a form document without activating it. */
    validate: (key: string, body: { yaml: string } | { document: ProviderDefinition }) =>
      request(`${providerPath(key)}/validate`, parseValidation, 'POST', body),
    /** Save/activate reviewed host YAML with revision and credential-origin checks. */
    saveDefinition: (key: string, yaml: string, etag: string, review?: DestinationReview, create = false) =>
      request(providerPath(key), parseDefinition, 'PUT', { yaml, destination_review: review }, etag, undefined, create),
    /** Apply the host file after explicit destination review when required. */
    reload: (key: string, etag: string, review?: DestinationReview) =>
      request(`${providerPath(key)}/reload`, parseDefinition, 'POST', { destination_review: review }, etag),
    /** Restore the retained provider revision after required destination review. */
    restore: (key: string, etag: string, review?: DestinationReview) =>
      request(`${providerPath(key)}/restore`, parseDefinition, 'POST', { destination_review: review }, etag),
    /** Request an isolated offline/live preview; the UI confirms live work separately. */
    test: (
      key: string,
      yaml: string,
      instance: string,
      mode: ProviderTestMode,
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
    /** Record review of a shipped update without overwriting host customization. */
    acknowledgeShipped: (key: string, etag: string) =>
      request(`${providerPath(key)}/shipped-reviewed`, parseDefinition, 'POST', {}, etag),
    /** Delete a revision-fenced provider definition only when the server permits it. */
    deleteDefinition: (key: string, etag: string) => request(providerPath(key), object, 'DELETE', {}, etag),
  };
}

/** Stateless authenticated client for store actions; private responses live only in Vuex/page state. */
export const connectionsApi = createConnectionsApi();
