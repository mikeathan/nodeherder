<script setup lang="ts">
  import ChatMessage from '@/components/assistant/core/ChatMessage.vue';
  import type { AssistantMessage } from '@/types/assistant.type';

  defineProps<{
    messages: AssistantMessage[];
    isLoading: boolean;
  }>();
</script>

<template>
  <div class="flex flex-column gap-5 w-full pb-4" style="max-width: 48rem">
    <ChatMessage v-for="msg in messages" :key="msg.id" :message="msg" />

    <div v-if="isLoading" class="message-wrapper flex w-full justify-content-start fadein" style="max-width: 48rem">
      <div class="flex align-items-start gap-4 w-full">
        <div class="mt-1">
          <i class="pi pi-sparkles text-primary text-xl pi-spin"></i>
        </div>
        <div class="message-content text-color-secondary flex-1 mt-1">
          <span class="text-base pi pi-ellipsis-h pi-spin text-xl"></span>
        </div>
      </div>
    </div>
  </div>
</template>
