import { Module } from 'vuex';
import { RootState } from '../../state';
import { ConsoleModuleState } from './state';

import { key } from '@/store';
import { LogMessageType } from '@/types/event-logs.type';
const MAX_MESSAGE_SIZE = 100;
export const ConsoleModule: Module<
  ConsoleModuleState,
  RootState
> = {
  namespaced: true,

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
      if (state.messages.length >= MAX_MESSAGE_SIZE) {
        state.messages.shift(); //? or clear all
      }
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



clearOldMessages({ commit }) {
  const currentTime = new Date().getTime();
  const cutoffTime = currentTime - 30 * 60 * 1000; // 30 minutes in milliseconds

  const filteredMessages = state.messages.filter(
    (message) => {
      return message.timestamp >= cutoffTime;
    },
  );

  commit('SET_MESSAGES', filteredMessages);
},


setInterval(() => {
  store.dispatch('clearOldMessages');
}, 60 * 1000); // Run every minute
