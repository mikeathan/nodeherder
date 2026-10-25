import type { Module } from 'vuex';
import type { RootState } from '../../state';
import { connectionsApi, ConnectionError } from '@/services/integrations.service';
import type { ConnectionCatalog, ConnectionInstance, ConnectionSetup, ProviderDraft } from '@/types/integrations.type';
export interface ConnectionsState {
  catalog: ConnectionCatalog | null;
  instances: ConnectionInstance[];
  current: ConnectionInstance | null;
  setup: ConnectionSetup | null;
  drafts: Record<string, ProviderDraft>;
  generation: number;
}
export const IntegrationsModule: Module<ConnectionsState, RootState> = {
  namespaced: true,
  state: () => ({ catalog: null, instances: [], current: null, setup: null, drafts: {}, generation: 0 }),
  mutations: {
    clear(state) {
      state.catalog = null;
      state.instances = [];
      state.current = null;
      state.setup = null;
      state.drafts = {};
      state.generation++;
    },
    catalog(state, value: ConnectionCatalog) {
      state.catalog = value;
    },
    instances(state, value: ConnectionInstance[]) {
      state.instances = value;
    },
    current(state, value: ConnectionInstance | null) {
      state.current = value;
    },
    setup(state, value: ConnectionSetup | null) {
      state.setup = value;
    },
    draft(state, { key, value }: { key: string; value: ProviderDraft }) {
      // Definitions are bounded server-side; keep at most 32 recently visited editor drafts.
      // Reinsert the current key so subsequent eviction never discards the active draft.
      delete state.drafts[key];
      state.drafts[key] = value;
      const oldest = Object.keys(state.drafts)[0];
      if (Object.keys(state.drafts).length > 32) delete state.drafts[oldest];
    },
  },
  actions: {
    // Transport/auth feedback belongs to the page scope. Store actions only own accepted data.
    async load({ state, commit }, signal?: AbortSignal) {
      const generation = state.generation;
      const [catalog, instances] = await Promise.all([
        connectionsApi.catalog(signal),
        connectionsApi.instances(signal).catch((error) => {
          if (error instanceof ConnectionError && error.status === 503) return [];
          throw error;
        }),
      ]);
      if (generation === state.generation && !signal?.aborted) {
        commit('catalog', catalog);
        commit('instances', instances);
      }
    },
    async loadInstance({ state, commit }, { id, signal }: { id: string; signal?: AbortSignal }) {
      const generation = state.generation;
      const value = await connectionsApi.instance(id, signal);
      if (generation === state.generation && !signal?.aborted) commit('current', value);
      return value;
    },
  },
};
