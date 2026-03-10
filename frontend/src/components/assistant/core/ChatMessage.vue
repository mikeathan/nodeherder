<script setup lang="ts">
  import type { AssistantMessage } from '@/types/assistant.type';
  import AssistantMessageBubble from '@/components/assistant/bubbles/AssistantMessageBubble.vue';
  import UserMessageBubble from '@/components/assistant/bubbles/UserMessageBubble.vue';

  defineProps<{
    message: AssistantMessage;
  }>();
</script>

<template>
  <div class="message-wrapper flex w-full" style="max-width: 48rem">
    <AssistantMessageBubble v-if="message.role === 'assistant'" :message="message" />
    <UserMessageBubble v-else :message="message" />
  </div>
</template>

<style scoped></style>
