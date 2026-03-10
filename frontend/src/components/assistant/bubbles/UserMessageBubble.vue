<script setup lang="ts">
  import type { AssistantMessage } from '@/types/assistant.type';
  import BaseMessageBubble from './BaseMessageBubble.vue';

  defineProps<{
    message: AssistantMessage;
  }>();
</script>

<template>
  <BaseMessageBubble alignment="right" :textToCopy="message.content">
    <template #content>
      <div
        class="message-content user-bubble p-3 px-4 text-base line-height-3 text-color border-round-3xl w-full"
        style="word-break: break-word; white-space: pre-wrap">
        {{ message.content }}
      </div>
    </template>
  </BaseMessageBubble>
</template>

<style>
  .user-bubble {
    background-color: rgba(0, 0, 0, 0.06);
  }

  html.dark .user-bubble {
    background-color: rgba(255, 255, 255, 0.08);
  }
</style>
