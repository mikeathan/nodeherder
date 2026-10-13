import { Module } from 'vuex';
import { RootState } from '@/store/state';
import { AuthModuleState } from './state';
import { User, UserSession } from '@/types/auth.type';
import { restoreSession as restoreSessionService } from '@/services/auth.service';

const pending = new WeakMap<AuthModuleState, Promise<void>>();

export const AuthModule: Module<AuthModuleState, RootState> = {
  namespaced: true,

  state: () => ({
    authenticated: false,
    user: {} as User,
    validation: 'idle',
    revision: 0,
  }),

  getters: {
    user: (state: AuthModuleState) => (): User => state.user,
    isAuthenticated: (state: AuthModuleState) => (): boolean => state.authenticated,
    validation: (state: AuthModuleState) => state.validation,
  },

  mutations: {
    login(state: AuthModuleState, userSession: UserSession) {
      state.user = userSession.user;
      state.authenticated = true;
      state.validation = 'authenticated';
      state.revision += 1;
    },
    logout(state: AuthModuleState) {
      state.user = {} as User;
      state.authenticated = false;
      state.validation = 'signedOut';
      state.revision += 1;
    },
    validation(state: AuthModuleState, value: AuthModuleState['validation']) {
      state.validation = value;
    },
  },
  actions: {
    loginUser({ commit }, userSession: UserSession) {
      commit('login', userSession);
    },
    logoutUser({ commit }) {
      commit('logout');
    },
    restoreSession({ commit, state }) {
      const existing = pending.get(state);
      if (existing) return existing;
      const revision = state.revision;
      commit('validation', 'checking');
      const check = (async () => {
        try {
          const session = await restoreSessionService();
          if (state.revision !== revision) return;
          if (session) commit('login', session);
          else commit('logout');
        } catch {
          if (state.revision === revision) commit('validation', 'unavailable');
        } finally {
          pending.delete(state);
        }
      })();
      pending.set(state, check);
      return check;
    },
  },
};
