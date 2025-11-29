import { Module } from 'vuex';
import { RootState } from '../../state';
import { ConsoleModuleState } from './state';
import { LogMessageType } from '@/types/console.type';

const MAX_MESSAGE_SIZE = 1000;
const MESSAGE_EXPIRATION_TIME = 24 * 60 * 60 * 1000; // 1 day in milliseconds

export const ConsoleModule: Module<
  ConsoleModuleState,
  RootState
> = {
  namespaced: true,

  state: () => ({
    messages: [],
    isEnabled: false,
  }),

  getters: {
    isEnabled: (state: ConsoleModuleState) => (): boolean =>
      state.isEnabled,

    messages:
      (state: ConsoleModuleState) => (): LogMessageType[] =>
        state.messages,
  },

  mutations: {
    setEnabled(
      state: ConsoleModuleState,
      enabled: boolean
    ) {
      state.isEnabled = enabled;
    },

    add(
      state: ConsoleModuleState,
      message: LogMessageType
    ) {
      if (state.messages.length >= MAX_MESSAGE_SIZE) {
        state.messages.shift(); //? or clear all
      }
      state.messages.push(message);
    },

    removeExpiredMessages(
      state: ConsoleModuleState,
      expirationTimeInMs?: number
    ) {
      if (state.messages.length === 0) {
        return;
      }

      const currentTime = new Date().getTime();
      const expirationTime =
        currentTime -
        (expirationTimeInMs == null
          ? MESSAGE_EXPIRATION_TIME
          : expirationTimeInMs);

      state.messages = state.messages.filter((message) => {
        return (
          new Date(message.timestamp).getTime() >=
          expirationTime
        );
      });

    
    },

    clear(state: ConsoleModuleState) {
      state.messages = [];
    },
  },

  actions: {
    init({ state, commit }) {
      commit('clear', state);
    },

    addMessage({ commit }, message: LogMessageType) {
      commit('add', message);
    },

    enableRemoteLogging(
      { state, commit, dispatch },
      enabled: boolean
    ) {
      if (enabled == state.isEnabled) {
        return;
      }

      commit('setEnabled', enabled);
      dispatch(
        'ws/emit',
        {
          event: 'enableRemoteLogger',
          message: {
            enabled: enabled,
          },
        },
        { root: true }
      );
    },
  },
};
