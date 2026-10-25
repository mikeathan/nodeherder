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
      state.drafts[key] = value;
    },
  },
  actions: {
    async load({ state, commit }, signal?: AbortSignal) {
      const generation = state.generation;
      try {
        const [catalog, instances] = await Promise.all([
          connectionsApi.catalog(signal),
          connectionsApi.instances(signal).catch((e) => {
            if (e instanceof ConnectionError && e.status === 503) return [];
            throw e;
          }),
        ]);
        if (generation === state.generation) {
          commit('catalog', catalog);
          commit('instances', instances);
        }
      } catch (e) {
        if (generation === state.generation && e instanceof ConnectionError && e.status === 401) {
          commit('clear');
          commit('auth/logout', undefined, { root: true });
        }
        throw e;
      }
    },
    async loadInstance({ state, commit }, { id, signal }: { id: string; signal?: AbortSignal }) {
      const generation = state.generation;
      try {
        const value = await connectionsApi.instance(id, signal);
        if (generation === state.generation) commit('current', value);
        return value;
      } catch (e) {
        if (generation === state.generation && e instanceof ConnectionError && e.status === 401) {
          commit('clear');
          commit('auth/logout', undefined, { root: true });
        }
        throw e;
      }
    },
  },
};
