import type { ActionContext, Module } from 'vuex';
import type { RootState } from '@/store/state';
import type { HomeLayout } from '@/types/home.type';
import { homeApi } from '@/services/home.service';
import { homeDeviceGroups } from '@/domain/home-compatibility';

/** Accepted layout only; local drag/dialog state and source values belong to their own owners. */
export interface HomeState {
  layout: HomeLayout | null;
  generation: number;
  reads: number;
}

interface HomeRequest {
  generation: number;
  sequence: number;
}
type HomeContext = ActionContext<HomeState, RootState>;

function beginRequest({ state, commit }: HomeContext): HomeRequest {
  commit('begin');
  return { generation: state.generation, sequence: state.reads };
}

function publishReply({ state, commit }: HomeContext, request: HomeRequest, signal: AbortSignal, layout: HomeLayout) {
  const currentSession = request.generation === state.generation;
  const latestRequest = request.sequence === state.reads;
  if (signal.aborted || !currentSession || !latestRequest) return;
  commit('accepted', layout);
  commit('hub/setDashboardGroups', homeDeviceGroups(layout), { root: true });
}

/** Session-owned Home layout. Generation/read fences prevent late replies replacing newer state. */
export const HomeModule: Module<HomeState, RootState> = {
  namespaced: true,
  state: () => ({ layout: null, generation: 0, reads: 0 }),
  mutations: {
    /** Clear private state and invalidate all outstanding replies at logout/session change. */
    clear(state) {
      state.layout = null;
      state.generation++;
      state.reads++;
    },
    /** Reserve the next accepted read/save sequence before starting transport work. */
    begin(state) {
      state.reads++;
    },
    /** Publish a validated owned DTO; components must create proposals rather than mutate it. */
    accepted(state, layout: HomeLayout) {
      state.layout = layout;
    },
  },
  actions: {
    /** Load local layout with the caller's cancellation signal and current session/read fence. */
    async read(context, signal: AbortSignal) {
      const request = beginRequest(context);
      const layout = await homeApi.read(signal);
      publishReply(context, request, signal, layout);
      return layout;
    },
    /** Persist one revision-fenced proposal; only this session's newest live request may publish. */
    async save(context, { proposal, signal }: { proposal: HomeLayout; signal: AbortSignal }) {
      const request = beginRequest(context);
      const layout = await homeApi.save(proposal, signal);
      publishReply(context, request, signal, layout);
      return layout;
    },
  },
};
