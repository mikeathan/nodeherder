/**
 * Wireless observation state (spec 064, FE-01).
 *
 * The module holds a validated snapshot only. It never derives device truth and never
 * persists identities, which keeps household-adjacent identifiers out of browser storage.
 */
import type { Module } from 'vuex';
import type { RootState } from '../../state';
import type { ObservationsState } from './state';
import { parseObservationSnapshot } from '@/contracts/observations';
import type { ObservationSnapshot } from '@/types/observations.type';

export const ObservationsModule: Module<ObservationsState, RootState> = {
  namespaced: true,

  state: (): ObservationsState => ({
    snapshot: null,
    loaded: false,
    stale: false,
  }),

  getters: {
    /** Latest snapshot, or null before the first one. */
    snapshot: (state: ObservationsState): ObservationSnapshot | null => state.snapshot,
    /** Whether a snapshot has ever been received. */
    hasData: (state: ObservationsState): boolean => state.loaded,
    /** Whether the viewer should warn that the socket is disconnected. */
    stale: (state: ObservationsState): boolean => state.stale,
    /** Identities the operator has not triaged yet; this is the "isolate new" view. */
    untriaged: (state: ObservationsState) => (): ObservationSnapshot['identities'] =>
      (state.snapshot?.identities ?? []).filter((identity) => identity.decision === 'unknown'),
    /** Every source that contributed at least one identity. */
    sources: (state: ObservationsState) => () => state.snapshot?.sources ?? [],
  },

  mutations: {
    /**
     * Store a snapshot only when it validates. A malformed payload leaves the previous
     * snapshot untouched rather than applying a partial one.
     */
    setSnapshot(state: ObservationsState, payload: unknown) {
      const snapshot = parseObservationSnapshot(payload);
      if (!snapshot) {
        console.error('Rejected malformed observations snapshot');
        return;
      }
      state.snapshot = snapshot;
      state.loaded = true;
      state.stale = false;
    },

    setStale(state: ObservationsState, stale: boolean) {
      state.stale = stale;
    },

    clear(state: ObservationsState) {
      state.snapshot = null;
      state.loaded = false;
      state.stale = false;
    },
  },

  actions: {
    /**
     * Mark the retained snapshot stale on disconnect. Identity data is deliberately not
     * cleared: the operator keeps context while the socket retries.
     */
    markStale({ commit }) {
      commit('setStale', true);
    },
  },
};
