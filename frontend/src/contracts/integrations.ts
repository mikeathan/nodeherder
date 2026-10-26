import { WASTE_SCHEDULE_TIMEZONE } from '@/config/integrations';
import type {
  ConnectionBoard,
  ConnectionCatalog,
  ConnectionDefinition,
  ConnectionFailure,
  ConnectionInstance,
  ConnectionOperation,
  ConnectionSetup,
  DefinitionValidation,
  ProviderDefinition,
  WasteSchedule,
} from '@/types/integrations.type';

/** Reject non-object API data before field-level parsing; this is not a deep validator. */
export function object(value: unknown): Record<string, unknown> {
  if (!value || typeof value !== 'object' || Array.isArray(value)) throw new Error('Invalid Integrations response.');
  return value as Record<string, unknown>;
}
function text(value: unknown): string {
  if (typeof value !== 'string') throw new Error('Invalid Integrations text field.');
  return value;
}
function flag(value: unknown): boolean {
  if (typeof value !== 'boolean') throw new Error('Invalid Integrations flag.');
  return value;
}
function number(value: unknown): number {
  if (typeof value !== 'number' || !Number.isSafeInteger(value) || value < 0)
    throw new Error('Invalid Integrations revision.');
  return value;
}
function list<T>(value: unknown, parse: (v: unknown) => T): T[] {
  if (!Array.isArray(value)) throw new Error('Invalid Integrations list.');
  return value.map(parse);
}
function member(value: unknown, values: string[]): string {
  const s = text(value);
  if (!values.includes(s)) throw new Error('Unsupported Integrations value.');
  return s;
}
function nullableText(value: unknown): string | null {
  return value === null ? null : text(value);
}
function timestamp(value: unknown): string {
  const s = text(value);
  if (!s.endsWith('Z') || !Number.isFinite(Date.parse(s))) throw new Error('Invalid Integrations timestamp.');
  return s;
}
function nullableTime(value: unknown): string | null {
  return value === null ? null : timestamp(value);
}
function date(value: unknown): string {
  const s = text(value);
  if (!/^\d{4}-\d{2}-\d{2}$/.test(s) || new Date(`${s}T00:00:00Z`).toISOString().slice(0, 10) !== s)
    throw new Error('Invalid collection date.');
  return s;
}
/** Validate a nullable, sanitized failure envelope without retaining a raw upstream body. */
export function parseFailure(value: unknown): ConnectionFailure | null {
  if (value === null) return null;
  const v = object(value);
  const result: ConnectionFailure = { code: text(v.code), message: text(v.message) };
  if (v.retry_at !== undefined) result.retry_at = timestamp(v.retry_at);
  if (v.field_errors !== undefined)
    result.field_errors = list(v.field_errors, (f) => {
      const x = object(f);
      return { path: text(x.path), message: text(x.message) };
    });
  return result;
}
/**
 * Validate the versioned exact-property resource before it enters private frontend state.
 * Rejects unsupported source calendars, provisional evidence and inconsistent date states.
 * @throws Error for malformed/unsupported data; TypeScript types alone do not validate API input.
 */
export function parseWasteSchedule(value: unknown): WasteSchedule {
  const v = object(value);
  member(v.contract_version, ['waste_schedule/v1']);
  member(v.timezone, [WASTE_SCHEDULE_TIMEZONE]);
  text(v.instance_id);
  text(v.resource_id);
  number(v.revision);
  member(v.kind, ['exact', 'partial', 'confirmed_empty']);
  ['captured_at', 'last_checked_at', 'evaluated_at'].forEach((k) => timestamp(v[k]));
  nullableTime(v.source_retrieved_at);
  nullableTime(v.fresh_until);
  member(v.freshness, ['fresh', 'stale', 'unknown']);
  const quality = object(v.quality);
  if (quality.address_match !== 'exact' || quality.evidence_granularity !== 'property' || flag(quality.provisional))
    throw new Error('An approximate schedule cannot be saved.');
  member(quality.date_confidence, ['published_calendar', 'council_projection', 'next_collection_only', 'unknown']);
  nullableText(quality.source_confidence_label);
  const provenance = object(v.provenance);
  text(provenance.provider_id);
  text(provenance.provider_version);
  number(provenance.profile_revision);
  number(provenance.binding_revision);
  nullableText(provenance.council_gss);
  nullableText(provenance.source_url);
  list(v.notices, text);
  text(v.representative_consent);
  const seen = new Set<string>();
  list(v.services, (s) => {
    const service = object(s);
    const id = text(service.service_id);
    if (seen.has(id)) throw new Error('Duplicate collection service.');
    seen.add(id);
    text(service.name);
    nullableText(service.source_service_key);
    nullableText(service.source_waste_type);
    nullableText(service.bin_colour);
    nullableText(service.container);
    member(service.waste_type, [
      'refuse',
      'recycling',
      'paper_card',
      'glass',
      'garden',
      'food',
      'mixed',
      'other',
      'unknown',
    ]);
    member(service.subscription, ['required', 'not_required', 'unknown']);
    const state = member(service.date_state, ['known', 'confirmed_empty', 'unavailable']);
    const dates = list(service.dates, date);
    if ((state === 'known') !== dates.length > 0) throw new Error('Inconsistent collection dates.');
    const coverage = object(service.coverage);
    member(coverage.state, ['complete', 'limited', 'next_only', 'unknown', 'not_applicable']);
    if (coverage.from !== null) date(coverage.from);
    if (coverage.through !== null) date(coverage.through);
    return service;
  });
  return value as WasteSchedule;
}
/** Validate discovery/availability data; catalog reads do not contain private credentials. */
export function parseCatalog(value: unknown): ConnectionCatalog {
  const v = object(value);
  flag(v.available);
  parseFailure(v.availability_error);
  list(v.types, (x) => {
    const t = object(x);
    text(t.id);
    text(t.name);
    text(t.description);
    return t;
  });
  list(v.providers, (x) => {
    const p = object(x);
    text(p.key);
    text(p.name);
    flag(p.available);
    return p;
  });
  return value as ConnectionCatalog;
}
/** Validate shared wizard identity, choices, revisions and optional transient preview data. */
export function parseSetup(value: unknown): ConnectionSetup {
  const v = object(value);
  text(v.id);
  number(v.revision);
  text(v.etag);
  text(v.provider_key);
  text(v.status);
  timestamp(v.expires_at);
  if (v.preview !== undefined) parseWasteSchedule(v.preview);
  flag(v.credential_configured);
  parseFailure(v.last_error);
  list(v.choices, (x) => {
    const c = object(x);
    text(c.id);
    text(c.label);
    return c;
  });
  return value as ConnectionSetup;
}
/** Validate a redacted saved instance, including every resource, before consumers use it. */
export function parseInstance(value: unknown): ConnectionInstance {
  const v = object(value);
  list(v.credential_slots, text);
  text(v.id);
  text(v.name);
  text(v.provider_key);
  text(v.postcode);
  number(v.revision);
  text(v.etag);
  text(v.status);
  flag(v.enabled);
  flag(v.credential_configured);
  timestamp(v.updated_at);
  parseFailure(v.last_error);
  list(v.resources, parseWasteSchedule);
  list(v.runs, (r) => {
    const run = object(r);
    text(run.operation_id);
    text(run.kind);
    text(run.outcome);
    text(run.message);
    timestamp(run.at);
    return run;
  });
  return value as ConnectionInstance;
}
/** Validate each instance in a shared installation listing; one malformed item rejects the reply. */
export function parseInstances(value: unknown): ConnectionInstance[] {
  return list(value, parseInstance);
}

/** Reject malformed/duplicate placements before they can become shared layout state. */
export function parseBoard(value: unknown): ConnectionBoard {
  const v = object(value);
  const etag = text(v.etag);
  if (!etag) throw new Error('Missing Integrations board revision.');
  const seen = new Set<string>();
  const entries = list(v.entries, (entry) => {
    const item = object(entry);
    const id = text(item.instance_id);
    if (!id || id.length > 256 || seen.has(id)) throw new Error('Invalid Integrations board membership.');
    seen.add(id);
    return { instance_id: id, hidden: flag(item.hidden) };
  });
  if (entries.length > 50) throw new Error('Integrations board limit exceeded.');
  return { etag, entries };
}
/** Validate local operation progress/check timing; reading it does not authorize provider work. */
export function parseOperation(value: unknown): ConnectionOperation {
  const v = object(value);
  text(v.id);
  text(v.status);
  text(v.kind);
  timestamp(v.retry_at);
  timestamp(v.expires_at);
  parseFailure(v.last_error);
  return value as ConnectionOperation;
}
function optionalText(v: unknown): void {
  if (v !== undefined) text(v);
}
function optionalFlag(v: unknown): void {
  if (v !== undefined) flag(v);
}
function parseValue(value: unknown): void {
  if (value === null || typeof value === 'string' || typeof value === 'boolean') return;
  if (typeof value === 'number' && Number.isFinite(value)) return;
  const ref = object(value);
  if (Object.keys(ref).length !== 1 || !/^(input|state)\.[a-z][a-z0-9_]*$/.test(text(ref.ref)))
    throw new Error('Invalid request value reference.');
}
/** Validate supported Form/YAML document fields at ingress; server execution validation still applies. */
export function parseProviderDefinition(value: unknown): ProviderDefinition {
  const v = object(value);
  member(v.schema_version, ['nodeherder.integration/v1']);
  member(v.capability, ['waste_schedule/v1']);
  member(v.integration_type, ['waste_collection']);
  text(v.id);
  text(v.name);
  text(v.base_url);
  const inputs = object(v.inputs);
  Object.values(inputs).forEach((i) => {
    const x = object(i);
    member(x.type, ['postcode', 'text', 'select']);
    text(x.label);
    flag(x.required);
  });
  const auth = object(v.auth);
  member(auth.mode, ['none', 'bearer', 'optional_bearer']);
  optionalText(auth.secret_slot);
  const policy = object(v.policy);
  number(policy.poll_min_seconds);
  number(policy.renewal_min_seconds);
  number(policy.idempotency_window_seconds);
  const errors = object(v.errors);
  text(errors.code);
  list(errors.rules, (r) => {
    const rule = object(r);
    number(rule.status);
    text(rule.code);
    member(rule.action, ['renewal', 'restart_setup', 'reconnect', 'unsupported']);
    return rule;
  });
  const requests = object(v.requests);
  Object.values(requests).forEach((r) => {
    const x = object(r);
    member(x.method, ['GET', 'POST']);
    list(x.path, (p) => {
      parseValue(p);
      return p;
    });
    for (const key of ['query', 'body']) if (x[key] !== undefined) Object.values(object(x[key])).forEach(parseValue);
    optionalFlag(x.idempotency);
    optionalFlag(x.conditional);
    optionalText(x.response_root);
    if (x.captures !== undefined)
      Object.values(object(x.captures)).forEach((c) => {
        const capture = object(c);
        text(capture.pointer);
        member(capture.type, ['string', 'boolean', 'number', 'local-date', 'utc-timestamp']);
        if (capture.when !== undefined) member(capture.when, ['', 'complete']);
        optionalText(capture.strip_prefix);
        optionalFlag(capture.optional);
      });
  });
  const setup = object(v.setup);
  member(setup.type, ['search_select']);
  const search = object(setup.search);
  text(search.input);
  text(search.request);
  const options = object(search.options);
  text(options.items);
  text(options.value);
  text(options.label);
  if (search.expect !== undefined) {
    const expect = object(search.expect);
    text(expect.pointer);
    text(expect.value);
  }
  const select = object(setup.select);
  text(select.input);
  text(select.request);
  const status = object(setup.status);
  text(status.request);
  text(status.pointer);
  for (const key of ['pending', 'done', 'failed']) list(status[key], text);
  const refresh = object(v.refresh);
  text(refresh.request);
  optionalText(refresh.renewal_request);
  const mapping = object(v.mapping);
  Object.values(object(mapping.metadata)).forEach(text);
  const collections = object(mapping.collections);
  text(collections.items);
  Object.values(object(collections.fields)).forEach(text);
  return value as ProviderDefinition;
}
/** Validate host-definition metadata; inactive/error documents may remain editable recovery drafts. */
export function parseDefinition(value: unknown): ConnectionDefinition {
  const v = object(value);
  optionalText(v.shipped_yaml);
  optionalText(v.disk_error);
  object(v.document);
  text(object(v.document).name);
  text(v.key);
  text(v.yaml);
  text(v.content_hash);
  text(v.active_hash);
  number(v.revision);
  text(v.etag);
  text(v.status);
  flag(v.update_available);
  if (v.status === 'active') parseProviderDefinition(v.document);
  return value as ConnectionDefinition;
}
/** Validate the reviewed document and affected credential revisions used before activation. */
export function parseValidation(value: unknown): DefinitionValidation {
  const v = object(value);
  text(v.yaml);
  text(v.content_hash);
  text(v.origin);
  text(v.previous_origin);
  parseProviderDefinition(v.document);
  list(v.affected_instances, (x) => {
    const r = object(x);
    text(r.credential_slot);
    text(r.instance_id);
    number(r.revision);
    number(r.credential_revision);
    return r;
  });
  return value as DefinitionValidation;
}
