<script setup lang="ts">
  import ChatMessage from '@/components/assistant/chat/ChatMessage.vue';
  import type { AssistantMessage } from '@/types/assistant.type';
  import UiIcon from '@/components/ui/UiIcon.vue';

  defineProps<{ messages: AssistantMessage[]; isLoading: boolean }>();
  defineEmits<{ (e: 'rerun', content: string): void }>();
</script>

<template>
  <div class="nh-chat-list">
    <ChatMessage v-for="message in messages" :key="message.id" :message="message" @rerun="$emit('rerun', $event)" />
    <div v-if="isLoading" class="nh-thinking" role="status">
      <UiIcon name="assistant" />
      <span>Thinking…</span>
    </div>
  </div>
</template>

<style scoped>
  .nh-chat-list {
    width: 100%;
    max-width: 48rem;
    display: flex;
    flex-direction: column;
    gap: 1.25rem;
    padding-bottom: var(--nh-space);
  }
  .nh-thinking {
    display: flex;
    align-items: center;
    gap: 0.45rem;
    color: var(--nh-text-muted);
    font-size: 0.85rem;
  }
  .nh-thinking .nh-ic {
    width: 1rem;
    height: 1rem;
    animation: nh-pulse 1.4s infinite;
  }
</style>
