<script setup lang="ts">
  import ChatInput from '../chat/ChatInput.vue';

  defineProps<{
    modelValue: string;
    isLoading: boolean;
  }>();

  defineEmits<{
    (e: 'update:modelValue', value: string): void;
    (e: 'send'): void;
  }>();
</script>

<template>
  <div
    class="absolute top-0 left-0 w-full h-full flex flex-column align-items-center justify-content-center px-4 pb-8 fadein z-1">
    <div class="w-full flex flex-column align-items-center">
      <div class="text-center mb-5">
        <h1 class="text-xl font-normal text-color-secondary m-0 mt-2">How can I help you today?</h1>
      </div>

      <div class="w-full">
        <ChatInput
          :modelValue="modelValue"
          @update:modelValue="$emit('update:modelValue', $event)"
          :isLoading="isLoading"
          @send="$emit('send')" />
      </div>
    </div>
  </div>
</template>
