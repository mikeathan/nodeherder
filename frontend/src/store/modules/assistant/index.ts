import { Module } from 'vuex';
import { RootState } from '../../state';
import { AssistantState } from './state';
import { AssistantMessage, Conversation, ConversationSummary } from '@/types/assistant.type';
import { generateConversationId } from '@/utils/unique';

const emptyConversation = (): Conversation => ({ id: generateConversationId(), messages: [] });

export const AssistantModule: Module<AssistantState, RootState> = {
  namespaced: true,

  state: () => ({
    activeConversation: emptyConversation(),
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
    clear(state) {
      state.activeConversation = emptyConversation();
      state.conversations = [];
    },
    setActiveConversation(state, conversation: Conversation) {
      state.activeConversation = conversation;
    },
    clearActiveConversation(state) {
      state.activeConversation = emptyConversation();
    },
    setConversations(state, conversations: ConversationSummary[]) {
      state.conversations = conversations;
    },
    addMessage(state, message: AssistantMessage) {
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
    addMessage({ commit }, message: AssistantMessage) {
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
