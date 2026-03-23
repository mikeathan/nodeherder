<script setup lang="ts">
  import type { AssistantMessage } from '@/types/assistant.type';
  import BaseBubble from './BaseBubble.vue';

  defineProps<{
    message: AssistantMessage;
  }>();
</script>

<template>
  <BaseBubble alignment="left" :textToCopy="message.content">
    <template #avatar>
      <i class="pi pi-sparkles text-primary text-xl"></i>
    </template>
    <template #content>
      <div
        class="message-content text-base line-height-3 text-color w-full bg-transparent"
        style="word-break: break-word; white-space: pre-wrap">
        {{ message.content }}
      </div>
    </template>
  </BaseBubble>
</template>

<style scoped></style>
