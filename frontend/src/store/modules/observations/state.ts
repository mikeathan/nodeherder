import type { ObservationSnapshot } from '@/types/observations.type';

/** State owned by the wireless observation viewer (spec 064). */
export interface ObservationsState {
  /** Latest validated snapshot, or null before the first one arrives. */
  snapshot: ObservationSnapshot | null;
  /** True once a snapshot has been received, so empty and loading are distinguishable. */
  loaded: boolean;
}
