import { Module } from 'vuex';
import { RootState } from '@/store/state';
import { AuthModuleState } from './state';
import { User, UserSession } from '@/types/auth.type';
import { restoreSession as restoreSessionService } from '@/services/auth.service';

export const AuthModule: Module<AuthModuleState, RootState> = {
  namespaced: true,

  state: () => ({
    authenticated: false,
    user: {} as User,
  }),

  getters: {
    user: (state: AuthModuleState) => (): User => state.user,
    isAuthenticated: (state: AuthModuleState) => (): boolean => state.authenticated,
  },

  mutations: {
    login(state: AuthModuleState, userSession: UserSession) {
      state.user = userSession.user;
      state.authenticated = true;
    },
    logout(state: AuthModuleState) {
      state.user = {} as User;
      state.authenticated = false;
    },
  },
  actions: {
    loginUser({ commit }, userSession: UserSession) {
      commit('login', userSession);
    },
    logoutUser({ commit }) {
      commit('logout');
    },
    async restoreSession({ commit }) {
      const session = await restoreSessionService();
      if (session) {
        commit('login', session);
      } else {
        commit('logout');
      }
    },
  },
};
