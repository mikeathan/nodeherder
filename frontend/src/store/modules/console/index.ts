import { Module } from 'vuex';
import { RootState } from '../../state';
import { ConsoleModuleState } from './state';

import { key } from '@/store';
import { LogMessageType } from '@/types/event-logs.type';

export const ConsoleModule: Module<
  ConsoleModuleState,
  RootState
> = {
  namespaced: true,

  // TODO: add limit how much we can store.
 clear last if we exeed number
  state: () => ({
    messages: [],
    initialized: false,
  }),

  getters: {
    initialized:
      (state: ConsoleModuleState) => (): boolean =>
        state.initialized,

    messages:
      (state: ConsoleModuleState) => (): LogMessageType[] =>
        state.messages,
  },

  mutations: {
    add(
      state: ConsoleModuleState,
      message: LogMessageType,
    ) {
      state.messages.push(message);
    },

    clear(state: ConsoleModuleState) {
      state.messages = [];

      state.initialized = false;
    },
  },

  actions: {
    init({ state, commit }) {
      commit('clear', state);

      state.initialized = true;
    },
  },
};
