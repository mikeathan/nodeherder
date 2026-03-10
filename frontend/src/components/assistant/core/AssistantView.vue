<script setup lang="ts">
  import ChatInput from '@/components/assistant/core/ChatInput.vue';
  import AssistantMessageList from '@/components/assistant/core/AssistantMessageList.vue';
  import AssistantEmptyState from '@/components/assistant/state/AssistantEmptyState.vue';
  import AssistantNotConfigured from '@/components/assistant/state/AssistantNotConfigured.vue';
  import { useAssistant } from '@/composables/useAssistant';

  const { activeConversation, inputText, isLoading, chatContainer, sendMessage, startNewConversation, isConfigured } =
    useAssistant();
</script>

<template>
  <div class="assistant-container flex flex-column h-full surface-ground pt-4 relative">
    <AssistantNotConfigured v-if="!isConfigured" />

    <template v-else>
      <div class="absolute top-0 right-0 p-4 z-5" v-if="activeConversation.messages.length > 0">
        <Button
          icon="pi pi-pencil"
          label="New Chat"
          class="p-button-text p-button-sm text-color-secondary hover:text-color transition-colors"
          @click="startNewConversation" />
      </div>

      <AssistantEmptyState
        v-if="activeConversation.messages.length === 0"
        v-model="inputText"
        :isLoading="isLoading"
        @send="sendMessage" />

      <template v-else>
        <div
          ref="chatContainer"
          class="flex-1 overflow-y-auto px-4 py-4 flex flex-column align-items-center w-full min-h-0 messages-container">
          <AssistantMessageList :messages="activeConversation.messages" :isLoading="isLoading" />
        </div>

        <div class="w-full flex justify-content-center pb-6 px-4 surface-ground pt-3 relative z-2">
          <div class="w-full">
            <ChatInput v-model="inputText" :isLoading="isLoading" @send="sendMessage" />
          </div>
        </div>
      </template>
    </template>
  </div>
</template>

<style scoped>
  .assistant-container {
    height: 100%;
  }
  .messages-container {
    scroll-behavior: smooth;
  }
</style>
