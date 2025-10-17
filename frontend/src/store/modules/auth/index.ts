import { Module } from 'vuex';
import { RootState } from '@/store/state';
import { AuthModuleState } from './state';
import { User } from '@/types/auth.type';
import { isTokenExpired } from '@/contracts/auth';

export const AuthModule: Module<AuthModuleState, RootState> = {
  namespaced: true,

  state: () => ({
    authenticated: false,
    user: {} as User,
    token: null,
  }),

  //// TODO:
  // we might have to add the user in the token
  ///
  getters: {
    user: (state: AuthModuleState) => (): User => state.user,
    isAuthenticated: (state: AuthModuleState) => (): boolean => state.authenticated,
    token: (state: AuthModuleState) => (): string | null => state.token,
  },

  mutations: {
    login(state: AuthModuleState, payload: { user: User; token: string }) {
      state.user = payload.user;
      state.token = payload.token;
      state.authenticated = true;
    },
    logout(state: AuthModuleState) {
      state.user = {} as User;
      state.token = null;
      state.authenticated = false;
    },
  },
  actions: {
    loginUser({ commit, rootState }, payload: { user: User; token: string }) {
      commit('login', payload);
    },
    logoutUser({ commit, rootState }) {
      commit('logout');
    },
    restoreSession({ commit, getters, rootState }) {
      const token = getters.token();
      if (token && !isTokenExpired(token)) {
        commit('login', { user: getters.user(), token });
      } else {
        commit('logout');
      }
    },
  },
};
