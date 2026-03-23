import { ref, computed, onMounted, nextTick } from 'vue';
import { store } from '@/store';
import {
  sendMessageToLLM,
  fetchConversations,
  fetchConversationHistory,
  deleteConversation as deleteConversationApi
} from '@/services/assistant.service';
import { createAssistantErrorResponse, createAssistantResponse, createUserRequest } from '@/contracts/assistant';

export function useAssistant() {
  const inputText = ref('');
  const chatContainer = ref<HTMLElement | null>(null);

  const isLoading = ref(false);
  const isLoadingHistory = ref(false);

  const activeConversation = computed(() => store.getters['assistant/activeConversation']);
  const conversations = computed(() => store.getters['assistant/conversations']);

  const notConfiguredReason = computed(() => {
    const config = store.getters['hub/assistant']();
    if (!config?.url || config.url === '') {
      return 'assistant';
    }
    const mcpEnabled = store.getters['hub/mcpConfig']();
    const mcpStatus = store.getters['hub/mcpStatus']();
    
    if (!mcpEnabled) {
      return 'mcp';
    }

    // Wait for the websocket to hydrate the status
    if (mcpStatus === null) {
      return null;
    }

    if (mcpStatus.running !== true || mcpStatus.connectedClients === 0) {
      return 'mcp';
    }
    
    return null;
  });

  const isConfigured = computed(() => notConfiguredReason.value === null);

  const scrollToBottom = async () => {
    await nextTick();
    setTimeout(() => {
      if (chatContainer.value) {
        chatContainer.value.scrollTop = chatContainer.value.scrollHeight;
      }
    }, 10);
  };

  const sendMessage = async () => {
    if (!isConfigured.value) return;

    const input = inputText.value.trim();
    if (!input || isLoading.value) return;

    inputText.value = '';
    isLoading.value = true;
    
    await store.dispatch('assistant/addMessage', createUserRequest(input));
    scrollToBottom();

    try {
      const responseText = await sendMessageToLLM(activeConversation.value.id, input);
      await store.dispatch('assistant/addMessage', createAssistantResponse(responseText));
    } catch (error) {
      console.error('Failed to send message:', error);
      await store.dispatch('assistant/addMessage', createAssistantErrorResponse(error as Error));
    } finally {
      isLoading.value = false;
      scrollToBottom();
    }
  };

  const startNewConversation = () => {
    store.dispatch('assistant/startNewConversation');
    inputText.value = '';
  };

  const loadConversations = async () => {
    try {
      const response = await fetchConversations();
      await store.dispatch('assistant/setConversations', response.conversations ?? []);
    } catch (error) {
      console.error('Failed to load conversations:', error);
      await store.dispatch('assistant/setConversations', []);
    }
  };

  const selectConversation = async (conversationId: string) => {
    isLoadingHistory.value = true;
    try {
      const history = await fetchConversationHistory(conversationId);
      await store.dispatch('assistant/setActiveConversation', {
        id: history.conversation_id,
        messages: history.messages ?? [],
      });
    } catch (error) {
      console.error('Failed to load conversation history:', error);
    } finally {
      isLoadingHistory.value = false;
      scrollToBottom();
    }
  };

  const deleteConversation = async (conversationId: string) => {
    try {
      await deleteConversationApi(conversationId);
      await store.dispatch('assistant/removeConversationFromList', conversationId);
    } catch (error) {
      console.error('Failed to delete conversation:', error);
    }
  };

  const hydrateFromBackend = async () => {
    const id = activeConversation.value?.id;
    if (!id) return;

    isLoadingHistory.value = true;
    try {
      const history = await fetchConversationHistory(id);
      await store.dispatch('assistant/setActiveConversation', {
        id: history.conversation_id,
        messages: history.messages ?? [],
      });
    } catch {
      // Not found
    } finally {
      isLoadingHistory.value = false;
      scrollToBottom();
    }
  };

  onMounted(async () => {
    await hydrateFromBackend();
    await loadConversations();
  });

  return {
    activeConversation,
    conversations,
    inputText,
    isLoading,
    isLoadingHistory,
    chatContainer,
    sendMessage,
    startNewConversation,
    loadConversations,
    selectConversation,
    deleteConversation,
    isConfigured,
    notConfiguredReason,
  };
}
