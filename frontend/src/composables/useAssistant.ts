import { ref, computed, onMounted, onScopeDispose, nextTick } from 'vue';
import { store } from '@/store';
import {
  sendMessageToLLM,
  fetchConversations,
  fetchConversationHistory,
  deleteConversation as deleteConversationApi,
} from '@/services/assistant.service';
import { createAssistantErrorResponse, createAssistantResponse, createUserRequest } from '@/contracts/assistant';
import type { Conversation, ConversationSummary } from '@/types/assistant.type';

export function useAssistant() {
  const inputText = ref('');
  const chatContainer = ref<HTMLElement | null>(null);

  const isLoading = ref(false);
  const isLoadingHistory = ref(false);

  const activeConversation = computed<Conversation>(() => store.getters['assistant/activeConversation']);
  const conversations = computed<ConversationSummary[]>(() => store.getters['assistant/conversations']);

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

  type RequestKind = 'message' | 'history' | 'list' | `delete:${string}`;
  const requests = new Map<RequestKind, AbortController>();
  let disposed = false;
  let generation = 0;
  let scrollTimer: ReturnType<typeof setTimeout> | undefined;

  function cancel(kind: RequestKind) {
    requests.get(kind)?.abort();
    requests.delete(kind);
    if (kind === 'message') isLoading.value = false;
    if (kind === 'history') isLoadingHistory.value = false;
  }
  function begin(kind: RequestKind) {
    cancel(kind);
    const controller = new AbortController();
    requests.set(kind, controller);
    return controller;
  }
  function current(kind: RequestKind, controller: AbortController) {
    return !disposed && !controller.signal.aborted && requests.get(kind) === controller;
  }
  function finish(kind: RequestKind, controller: AbortController) {
    if (requests.get(kind) === controller) cancel(kind);
  }
  function invalidate() {
    generation += 1;
    for (const kind of requests.keys()) cancel(kind);
    clearTimeout(scrollTimer);
    inputText.value = '';
  }
  const unsubscribe = store.subscribe((mutation) => {
    if (mutation.type === 'auth/logout') invalidate();
  });
  onScopeDispose(() => {
    disposed = true;
    invalidate();
    unsubscribe();
    chatContainer.value = null;
  });

  const setChatContainer = (el: unknown) => {
    if (el === null || el instanceof HTMLElement) chatContainer.value = el;
  };
  const scrollToBottom = async () => {
    const owner = generation;
    await nextTick();
    if (disposed || owner !== generation) return;
    const scroll = () => {
      if (chatContainer.value) chatContainer.value.scrollTop = chatContainer.value.scrollHeight;
    };
    scroll();
    clearTimeout(scrollTimer);
    scrollTimer = setTimeout(scroll, 150);
  };

  const sendMessage = async () => {
    const input = inputText.value.trim();
    if (disposed || !isConfigured.value || !input || isLoading.value) return;
    const id = activeConversation.value.id;
    const controller = begin('message');
    inputText.value = '';
    isLoading.value = true;
    await store.dispatch('assistant/addMessage', createUserRequest(input));
    void scrollToBottom();
    try {
      if (!current('message', controller)) return;
      const reply = await sendMessageToLLM(id, input, controller.signal);
      if (current('message', controller) && activeConversation.value.id === id) {
        await store.dispatch('assistant/addMessage', createAssistantResponse(reply));
      }
    } catch (error) {
      if (current('message', controller) && activeConversation.value.id === id) {
        await store.dispatch('assistant/addMessage', createAssistantErrorResponse(error as Error));
      }
    } finally {
      if (current('message', controller)) void scrollToBottom();
      finish('message', controller);
    }
  };

  const startNewConversation = () => {
    cancel('message');
    cancel('history');
    void store.dispatch('assistant/startNewConversation');
    inputText.value = '';
  };

  const loadConversations = async () => {
    if (disposed) return;
    const controller = begin('list');
    try {
      const response = await fetchConversations(controller.signal);
      if (current('list', controller)) {
        await store.dispatch('assistant/setConversations', response.conversations ?? []);
      }
    } catch {
      if (current('list', controller)) {
        console.error('Failed to load conversations');
        await store.dispatch('assistant/setConversations', []);
      }
    } finally {
      finish('list', controller);
    }
  };

  const loadHistory = async (id: string, quiet = false) => {
    if (disposed) return;
    const controller = begin('history');
    isLoadingHistory.value = true;
    try {
      const history = await fetchConversationHistory(id, controller.signal);
      if (current('history', controller)) {
        await store.dispatch('assistant/setActiveConversation', {
          id: history.conversation_id,
          messages: history.messages ?? [],
        });
        void scrollToBottom();
      }
    } catch {
      // Retain the current conversation on failures; aborted/stale work is ignored.
      if (!quiet && current('history', controller)) console.error('Failed to load conversation history');
    } finally {
      finish('history', controller);
    }
  };
  const selectConversation = (id: string) => {
    cancel('message');
    return loadHistory(id);
  };

  const deleteConversation = async (id: string) => {
    if (disposed) return;
    const kind = `delete:${id}` as const;
    const controller = begin(kind);
    try {
      await deleteConversationApi(id, controller.signal);
      if (!current(kind, controller)) return;
      cancel('list');
      if (activeConversation.value.id === id) {
        cancel('message');
        cancel('history');
      }
      await store.dispatch('assistant/removeConversationFromList', id);
    } catch {
      // A failed deletion leaves the conversation available for a manual retry.
      if (current(kind, controller)) console.error('Failed to delete conversation');
    } finally {
      finish(kind, controller);
    }
  };

  onMounted(async () => {
    const owner = generation;
    const id = activeConversation.value?.id;
    if (id) await loadHistory(id, true);
    if (!disposed && owner === generation) await loadConversations();
  });

  return {
    activeConversation,
    conversations,
    inputText,
    isLoading,
    isLoadingHistory,
    chatContainer,
    setChatContainer,
    sendMessage,
    startNewConversation,
    loadConversations,
    selectConversation,
    deleteConversation,
    isConfigured,
    notConfiguredReason,
  };
}
