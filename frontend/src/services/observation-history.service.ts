import { get } from '@/contracts/api';

/**
 * Stored evidence for one identity, as served by /api/observations/history.
 *
 * These are aggregates, not observations: a day says how often, how strong and how mobile a device
 * was. The backend deliberately stores no packet contents, so there is nothing here that could
 * reconstruct what any device transmitted.
 */
export type ObservationDaySummary = {
  key: string;
  /** UTC day, YYYY-MM-DD. */
  day: string;
  packets: number;
  activeHours: number;
  /** Bit N set when hour N (UTC) contained an observation. */
  hourMask: number;
  rssiMean: number;
  strongest: number;
  weakest: number;
  crossings: number;
  family?: string;
};

/** One recorded class change. */
export type ObservationTransition = {
  key: string;
  day: string;
  /** RFC 3339 timestamp of the change. */
  at: string;
  from: string;
  to: string;
  reason: string;
};

export type ObservationHistory = {
  summaries: ObservationDaySummary[];
  transitions: ObservationTransition[];
};

/**
 * Load one identity's stored evidence.
 *
 * The identity key is a query parameter because keys are opaque and contain characters that would
 * need escaping in a path. A 503 means the store is unavailable rather than that the device has no
 * history, so the two are reported differently: "no history yet" is normal on a first day, while
 * "unavailable" is a fault the operator should know about.
 */
export async function loadObservationHistory(key: string, limit = 30): Promise<ObservationHistory> {
  const search = new URLSearchParams({ key, limit: String(limit) });
  const response = await get(`/observations/history?${search}`);
  if (!response.ok) {
    throw new Error(
      response.status === 503
        ? 'Observation history is unavailable.'
        : 'Observation history could not be loaded.',
    );
  }
  const payload = (await response.json()) as Partial<ObservationHistory>;
  return {
    summaries: Array.isArray(payload.summaries) ? payload.summaries : [],
    transitions: Array.isArray(payload.transitions) ? payload.transitions : [],
  };
}
