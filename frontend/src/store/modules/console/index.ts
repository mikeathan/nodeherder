import { Module } from 'vuex';
import { RootState } from '../../state';
import { ConsoleModuleState } from './state';

import { key } from '@/store';
import { LogMessageType } from '@/types/event-logs.type';
const MAX_MESSAGE_SIZE = 100;
const MESSAGE_EXPIRATION_TIME = 30 * 60 * 1000; // 30 minutes in milliseconds


export const ConsoleModule: Module<
  ConsoleModuleState,
  RootState
> = {
  namespaced: true,

  state: () => ({
    messages: [],
  }),

  getters: {
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
      console.log(
        'Store - removeItems before',
        state.messages.length,
      );

      TODO
      // state.messages = state.messages.filter(
      //   (item) => !itemsToRemove.includes(item),
      // );

      console.log(
        'Store removeItems after',
        state.messages.length,
      );
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

    deleteExpiredMessages({ state, commit }) {
      if (state.messages.length === 0) {
        console.log(
          'Store - deleteExpiredMessages no messages',
        );
        return;
      }

      const currentTime = new Date().getTime();
      const expirationTime =
        currentTime - MESSAGE_EXPIRATION_TIME;


        do the filtering here and add this function to mutators
      const itemsToRemove = state.messages.filter(
        (message) => {
          return (
            new Date(message.timestamp).getTime() >=
            expirationTime
          );
        },
      );

      console.log(
        'Store - deleteExpiredMessages found',
        itemsToRemove.length,
      );

      if (itemsToRemove.length > 0) {
        commit('removeItems', itemsToRemove);
      }
    },
  },
};
