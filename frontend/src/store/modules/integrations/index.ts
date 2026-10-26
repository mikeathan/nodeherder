import type { Module } from 'vuex';
import type { RootState } from '../../state';
import { connectionsApi, ConnectionError } from '@/services/integrations.service';
import type {
  ConnectionBoard,
  ConnectionCatalog,
  ConnectionInstance,
  ConnectionSetup,
  ProviderDraft,
  ConnectionDefinition,
} from '@/types/integrations.type';
export interface ConnectionsState {
  board: ConnectionBoard | null;
  boardReturn: { id: string; scrollTop: number } | null;
  instanceReads: number;
  catalog: ConnectionCatalog | null;
  instances: ConnectionInstance[];
  current: ConnectionInstance | null;
  setup: ConnectionSetup | null;
  drafts: Record<string, ProviderDraft>;
  definitions: Record<string, ConnectionDefinition>;
  generation: number;
}
export const IntegrationsModule: Module<ConnectionsState, RootState> = {
  namespaced: true,
  state: () => ({
    board: null,
    boardReturn: null,
    instanceReads: 0,
    catalog: null,
    instances: [],
    current: null,
    setup: null,
    drafts: {},
    definitions: {},
    generation: 0,
  }),
  mutations: {
    beginInstanceRead(state) {
      state.instanceReads++;
    },
    /** Purge private presentation/draft state and invalidate replies from the previous session. */
    clear(state) {
      state.board = null;
      state.boardReturn = null;
      state.catalog = null;
      state.instances = [];
      state.current = null;
      state.setup = null;
      state.drafts = {};
      state.definitions = {};
      state.generation++;
    },
    catalog(state, value: ConnectionCatalog) {
      state.catalog = value;
    },
    board(state, value: ConnectionBoard | null) {
      state.board = value;
    },
    boardReturn(state, value: ConnectionsState['boardReturn']) {
      state.boardReturn = value;
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
    /** Retain at most 32 accepted definitions for immediate, key-scoped editor presentation. */
    definition(state, value: ConnectionDefinition) {
      // Accepted DTOs are session-only presentation caches, never authoritative editor buffers.
      delete state.definitions[value.key];
      state.definitions[value.key] = value;
      const oldest = Object.keys(state.definitions)[0];
      if (Object.keys(state.definitions).length > 32) delete state.definitions[oldest];
    },
    /** Evict a removed provider and its draft so navigation cannot resurrect its editor buffer. */
    forgetDefinition(state, key: string) {
      delete state.definitions[key];
      delete state.drafts[key];
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
    async loadBoard({ state, commit }, signal?: AbortSignal) {
      const generation = state.generation;
      commit('beginInstanceRead');
      const read = state.instanceReads;
      const catalog = await connectionsApi.catalog(signal);
      // Degraded installations can still explain availability without attempting private reads.
      const [instances, board] = catalog.available
        ? await Promise.all([connectionsApi.instances(signal), connectionsApi.board(signal)])
        : [[], null];
      if (generation === state.generation && !signal?.aborted) {
        commit('catalog', catalog);
        if (read === state.instanceReads) commit('instances', instances);
        commit('board', board);
      }
    },
    async readInstances({ state, commit }, signal?: AbortSignal) {
      const generation = state.generation;
      commit('beginInstanceRead');
      const read = state.instanceReads;
      const instances = await connectionsApi.instances(signal);
      // A slow local projection read must not replace a newer operation result.
      if (generation === state.generation && read === state.instanceReads && !signal?.aborted)
        commit('instances', instances);
    },
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
