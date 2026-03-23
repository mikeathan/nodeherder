<script setup lang="ts">
  import type { AssistantMessage } from '@/types/assistant.type';
  import AssistantMessageBubble from '@/components/assistant/chat/AgentBubble.vue';
  import UserMessageBubble from '@/components/assistant/chat/UserBubble.vue';

  defineProps<{
    message: AssistantMessage;
  }>();

  defineEmits<{
    (e: 'rerun', content: string): void;
  }>();
</script>

<template>
  <div class="message-wrapper flex w-full" style="max-width: 48rem">
    <AssistantMessageBubble v-if="message.role === 'assistant'" :message="message" />
    <UserMessageBubble v-else :message="message" @rerun="$emit('rerun', $event)" />
  </div>
</template>

<style scoped></style>
