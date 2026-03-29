import { Module } from 'vuex';
import { RootState } from '../../state';
import { AssistantState } from './state';
import { Conversation, ConversationSummary } from '@/types/assistant.type';
import { generateConversationId } from '@/utils/unique';

export const AssistantModule: Module<AssistantState, RootState> = {
  namespaced: true,

  state: () => ({
    activeConversation: {
      id: generateConversationId(),
      messages: [],
    },
    conversations: [],
  }),

  getters: {
    activeConversation(state): Conversation {
      return state.activeConversation;
    },
    conversations(state): ConversationSummary[] {
      return state.conversations;
    },
  },

  mutations: {
    setActiveConversation(state, conversation: Conversation) {
      state.activeConversation = conversation;
    },
    clearActiveConversation(state) {
      state.activeConversation = {
        id: generateConversationId(),
        messages: [],
      };
    },
    setConversations(state, conversations: ConversationSummary[]) {
      state.conversations = conversations;
    },
    addMessage(state, message) {
      state.activeConversation.messages.push(message);
    },
    removeConversationFromList(state, conversationId: string) {
      state.conversations = state.conversations.filter((c) => c.id !== conversationId);
    },
  },

  actions: {
    setActiveConversation({ commit }, conversation: Conversation) {
      commit('setActiveConversation', conversation);
    },
    setConversations({ commit }, conversations: ConversationSummary[]) {
      commit('setConversations', conversations);
    },
    addMessage({ commit }, message: any) {
      commit('addMessage', message);
    },
    removeConversationFromList({ commit, state, dispatch }, conversationId: string) {
      commit('removeConversationFromList', conversationId);
      if (state.activeConversation.id === conversationId) {
        dispatch('startNewConversation');
      }
    },
    startNewConversation({ commit }) {
      commit('clearActiveConversation');
    },
  },
};
