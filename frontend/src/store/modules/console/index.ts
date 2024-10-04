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

    removeItems(
      state: ConsoleModuleState,
      itemsToRemove: LogMessageType[],
    ) {
      state.messages = state.messages.filter(
        (item) => !itemsToRemove.includes(item),
      );
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

    deleteExpiredMessages({ state, commit }) {
      const currentTime = new Date().getTime();
      const cutoffTime = currentTime - 30 * 60 * 1000; // 30 minutes in milliseconds

      const itemsToRemove = state.messages.filter(
        (message) => {
          return (
            new Date(message.timestamp).getTime() >=
            cutoffTime
          );
        },
      );

      commit('remoteItems', itemsToRemove);
    },
  },
};

// Run the action every 30 seconds
setInterval(() => {
  store.dispatch('deleteExpiredMessages');
}, 30 * 1000);
