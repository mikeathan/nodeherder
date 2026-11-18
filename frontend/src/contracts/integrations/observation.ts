/**
 * Contract parser for the published observation census.
 *
 * It exists because `parseResource` dispatches on `contract_version`, and a version it does not
 * recognise falls through to the waste schedule parser, which rejects the payload. That failure is
 * silent in the UI — the instance is created, the response fails validation, and the card never
 * appears — so every resource version must have a parser here.
 */
import type { ObservationCensus } from '@/types/integrations';


/**
 * Validates an observation census.
 *
 * `identities` and `hidden` are counts of the same set, so `hidden` exceeding `identities` means the
 * payload is internally inconsistent and must be rejected rather than rendered as a negative shown
 * count.
 */
export function parseObservationCensus(value: unknown): ObservationCensus {
  const v = object(value);
  if (v.contract_version !== 'observation_census/v1') throw new Error('Unexpected observation version.');
  const identities = count(v.identities);
  const hidden = count(v.hidden);
  if (hidden > identities) throw new Error('An observation census cannot hide more devices than it has.');
  const sources = Array.isArray(v.sources) ? v.sources.map(parseSource) : [];
  return {
    contract_version: 'observation_census/v1',
    resource_id: string(v.resource_id),
    revision: count(v.revision),
    identities,
    hidden,
    sources,
    observed_at: timestamp(v.observed_at),
  };
}

/** One reporting source; counters are lifetime-per-process and never history. */
function parseSource(value: unknown): ObservationCensus['sources'][number] {
  const s = object(value);
  return {
    id: string(s.id),
    name: string(s.name),
    advertisements: count(s.advertisements),
    decoded: count(s.decoded),
    dropped: count(s.dropped),
  };
}

/** Reject non-object API data before field-level parsing; this is not a deep validator. */
function object(value: unknown): Record<string, unknown> {
  if (!value || typeof value !== 'object' || Array.isArray(value)) throw new Error('Invalid observation payload.');
  return value as Record<string, unknown>;
}

function string(value: unknown): string {
  if (typeof value !== 'string') throw new Error('Invalid observation text field.');
  return value;
}

function count(value: unknown): number {
  if (typeof value !== 'number' || !Number.isFinite(value) || value < 0) {
    throw new Error('Invalid observation count.');
  }
  return value;
}

function timestamp(value: unknown): string {
  if (typeof value !== 'string' || Number.isNaN(Date.parse(value))) {
    throw new Error('Invalid observation timestamp.');
  }
  return value;
}
