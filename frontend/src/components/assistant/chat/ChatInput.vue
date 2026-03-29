<script setup lang="ts">
  import { computed } from 'vue';

  const props = defineProps<{
    modelValue: string;
    isLoading?: boolean;
  }>();

  const emit = defineEmits<{
    (e: 'update:modelValue', value: string): void;
    (e: 'send'): void;
  }>();

  const trimmedValue = computed(() => props.modelValue.trim());
  const canSend = computed(() => !!trimmedValue.value && !props.isLoading);

  const handleInput = (event: Event) => {
    const target = event.target as HTMLTextAreaElement;
    emit('update:modelValue', target.value);
  };

  const handleSend = () => {
    if (canSend.value) {
      emit('send');
    }
  };
</script>

<template>
  <div
    class="chat-input-container w-full relative shadow-1 border-1 border-solid surface-border border-round-xl py-1 px-2 flex align-items-end transition-colors surface-overlay">
    <Textarea
      :value="modelValue"
      @input="handleInput"
      autoResize
      rows="1"
      class="flex-1 bg-transparent border-none outline-none shadow-none text-color text-base custom-textarea px-3 py-2 m-1"
      placeholder="Ask anything..."
      @keydown.enter.prevent="handleSend" />

    <div class="pb-2 pr-1 flex align-items-center justify-content-center">
      <Button
        rounded
        :disabled="!canSend"
        :class="[
          'p-0 flex justify-content-center align-items-center border-none transition-colors chat-send-button',
          canSend ? 'active' : 'surface-400',
        ]"
        style="width: 32px; height: 32px"
        @click="handleSend">
        <i
          :class="['pi pi-arrow-up text-sm font-bold chat-send-icon', canSend ? 'active' : 'text-color-secondary']"></i>
      </Button>
    </div>
  </div>
</template>

<style scoped>
  .chat-input-container {
    max-width: 48rem;
    margin: 0 auto;
  }

  .chat-send-button.active {
    background-color: var(--text-color) !important;
  }

  .chat-send-icon.active {
    color: var(--surface-overlay) !important;
  }
  :deep(.custom-textarea) {
    min-height: 24px;
    max-height: 300px;
    resize: none;
    background-color: transparent !important;
    border: none !important;
    box-shadow: none !important;
  }
  :deep(.custom-textarea:focus) {
    outline: none !important;
    box-shadow: none !important;
  }

  :deep(.p-button.p-button-icon-only) {
    padding: 0;
  }
  :deep(.p-button.surface-400:disabled) {
    background-color: var(--surface-00) !important;
    color: var(--surface-600) !important;
    opacity: 1;
  }
</style>
