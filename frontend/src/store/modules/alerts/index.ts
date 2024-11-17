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

    showWarning(
      { dispatch },
      message: string,
      timeout?: number,
    ) {
      const alert = createWarning(message, timeout);
      dispatch('showAlert', alert);
    },
    showError(
      { dispatch },
      message: string,
      timeout?: number,
    ) {
      console.log('store alert.showError:', message);
      const alert = createError(message, timeout);
      dispatch('showAlert', alert);
    },
    showInfo(
      { dispatch },
      message: string,
      timeout?: number,
    ) {
      const alert = createInfo(message, timeout);
      dispatch('showAlert', alert);
    },
    showSuccess(
      { dispatch },
      message: string,
      timeout?: number,
    ) {
      console.log('store alert.showSuccess', message);

      const alert = createSuccess(message, timeout);
      dispatch('showAlert', alert);
    },
    showAlert({ commit }, alert: AlertMessage) {
      console.log('store alert.showAlert', alert);

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
