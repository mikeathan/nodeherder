import { Module } from 'vuex';
import { RootState } from '../../state';
import { AlertModuleState } from './state';
import {
  createWarning,
  createSuccess,
  createInfo,
  createError,
} from '../../../contracts/alerts';
import { AlertMessage } from '../../../types/alerts.type';

const TIMEOUT = 3000;

export const AlertsModule: Module<
  AlertModuleState,
  RootState
> = {
  namespaced: true,

  state: () => ({
    messages: {},
  }),

  getters: {
    messages:
      (state: AlertModuleState) => (): AlertMessage[] =>
        Object.values(state.messages),
  },

  mutations: {
    showAlert(
      state: AlertModuleState,
      message: AlertMessage,
    ) {
      state.messages[message.id] = message;
    },

    removeAlert(state: AlertModuleState, id: string) {
      delete state.messages[id];
    },

    clear(state: AlertModuleState) {
      state.messages = {};
    },
  },

  actions: {
    init({ state, commit }) {
      commit('clear', state);
    },

    showWarning({ dispatch }, message: string) {
      const alert = createWarning(message, TIMEOUT);
      dispatch('showAlert', alert);
    },
    showError({ dispatch }, message: string) {
      const alert = createError(message, TIMEOUT);
      dispatch('showAlert', alert);
    },
    showInfo({ dispatch }, message: string) {
      const alert = createInfo(message, TIMEOUT);
      dispatch('showAlert', alert);
    },
    showSuccess({ dispatch }, message: string) {
      const alert = createSuccess(message, TIMEOUT);
      dispatch('showAlert', alert);
    },
    showAlert({ commit }, alert: AlertMessage) {
      commit('showAlert', alert);
      if (alert.timeout) {
        setTimeout(
          () => commit('removeAlert', alert.id),
          alert.timeout,
        );
      }
    },
  },
};
