import { ref, nextTick, computed } from 'vue';
import type { Conversation } from '@/types/assistant.type';
import { sendMessageToLLM } from '@/services/assistant.service';
import { generateConversationId } from '@/utils/unique';
import { createAssistantErrorResponse, createAssistantResponse, createUserRequest } from '@/contracts/assistant';
import { store } from '@/store';

export function useAssistant() {
  const activeConversation = ref<Conversation>({
    id: generateConversationId(),
    messages: [],
  });

  const inputText = ref('');
  const isLoading = ref(false);
  const chatContainer = ref<HTMLElement | null>(null);

  const isConfigured = computed(() => {
    const config = store.getters['hub/assistant']();
    return !!config?.url && config.url !== '';
  });

  const scrollToBottom = async () => {
    await nextTick();
    if (chatContainer.value) {
      chatContainer.value.scrollTop = chatContainer.value.scrollHeight;
    }
  };

  const sendMessage = async () => {
    if (!isConfigured.value) return;

    const config = store.getters['hub/assistant']();
    const url = config.url;

    const input = inputText.value.trim();
    if (!input || isLoading.value) return;

    const userMessage = input;
    inputText.value = '';

    activeConversation.value.messages.push(createUserRequest(userMessage));

    await scrollToBottom();
    isLoading.value = true;

    try {
      const responseText = await sendMessageToLLM(url, activeConversation.value.id, userMessage);
      activeConversation.value.messages.push(createAssistantResponse(responseText));
    } catch (error) {
      console.error('Failed to send message:', error);
      activeConversation.value.messages.push(createAssistantErrorResponse(error as Error));
    } finally {
      isLoading.value = false;
      await scrollToBottom();
    }
  };

  const startNewConversation = () => {
    activeConversation.value = {
      id: generateConversationId(),
      messages: [],
    };
    inputText.value = '';
    isLoading.value = false;
  };

  return {
    activeConversation,
    inputText,
    isLoading,
    chatContainer,
    sendMessage,
    startNewConversation,
    isConfigured,
  };
}
