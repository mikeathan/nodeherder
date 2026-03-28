<script setup lang="ts">
  import { ref, computed } from 'vue';
  import ChatInput from '@/components/assistant/chat/ChatInput.vue';
  import AssistantMessageList from '@/components/assistant/chat/ChatMessageList.vue';
  import HistorySidebar from '@/components/assistant/history/HistorySidebar.vue';
  import AssistantEmptyState from '@/components/assistant/states/EmptyState.vue';
  import AssistantNotConfigured from '@/components/assistant/states/NotConfiguredState.vue';
  import { useAssistant } from '@/composables/useAssistant';

  const {
    activeConversation,
    conversations,
    inputText,
    isLoading,
    isLoadingHistory,
    setChatContainer,
    sendMessage,
    startNewConversation,
    selectConversation,
    deleteConversation,
    isConfigured,
    notConfiguredReason,
  } = useAssistant();

  const handleSelect = (id: string) => {
    selectConversation(id);
  };

  const handleRerun = (newText: string) => {
    inputText.value = newText;
    sendMessage();
  };
</script>

<template>
  <HistorySidebar
    :conversations="conversations"
    :activeId="activeConversation?.id"
    @new-chat="startNewConversation"
    @select="handleSelect"
    @delete="deleteConversation"
  >
    <div class="flex flex-column h-full pt-4">
      <AssistantNotConfigured v-if="!isConfigured" :reason="notConfiguredReason" />

      <template v-else>
        <div v-if="isLoadingHistory" class="flex-1 flex align-items-center justify-content-center">
          <i class="pi pi-spin pi-spinner text-4xl text-color-secondary"></i>
        </div>

        <template v-else>
          <AssistantEmptyState
            v-if="activeConversation.messages.length === 0"
            v-model="inputText"
            :isLoading="isLoading"
            @send="sendMessage" />

          <template v-else>
            <div
              :ref="setChatContainer"
              class="flex-1 overflow-y-auto px-4 py-4 flex flex-column align-items-center w-full min-h-0 messages-container">
              <AssistantMessageList
                :messages="activeConversation.messages"
                :isLoading="isLoading"
                @rerun="handleRerun" />
            </div>

            <div class="w-full flex justify-content-center pb-6 px-4 surface-ground pt-3 relative z-2 flex-shrink-0">
              <div class="w-full">
                <ChatInput v-model="inputText" :isLoading="isLoading" @send="sendMessage" />
              </div>
            </div>
          </template>
        </template>
      </template>
    </div>
  </HistorySidebar>
</template>

<style scoped>
  .messages-container {
    scroll-behavior: smooth;
  }
</style>
