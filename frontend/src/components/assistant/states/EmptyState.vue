<script setup lang="ts">
  /* First-run prompt inside the assistant surface; the composer owns send/enter behaviour. */
  import ChatInput from '../chat/ChatInput.vue';
  import UiIcon from '@/components/ui/UiIcon.vue';

  defineProps<{ modelValue: string; isLoading: boolean }>();
  defineEmits<{ (e: 'update:modelValue', value: string): void; (e: 'send'): void }>();
</script>

<template>
  <div class="nh-assistant-empty">
    <span class="nh-assistant-empty-ic"><UiIcon name="assistant" /></span>
    <h2>Ask about your home</h2>
    <p>Readings, history, devices and automations. Device changes still require confirmation in the app.</p>
    <div class="nh-assistant-empty-input">
      <ChatInput
        :modelValue="modelValue"
        :isLoading="isLoading"
        @update:modelValue="$emit('update:modelValue', $event)"
        @send="$emit('send')" />
    </div>
  </div>
</template>

<style scoped>
  .nh-assistant-empty {
    flex: 1;
    min-height: 0;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 0.5rem;
    padding: var(--nh-space);
    text-align: center;
  }
  .nh-assistant-empty-ic {
    width: 3rem;
    height: 3rem;
    border-radius: 1rem;
    display: grid;
    place-items: center;
    background: var(--nh-accent-soft);
    color: var(--nh-accent);
  }
  .nh-assistant-empty-ic .nh-ic {
    width: 1.6rem;
    height: 1.6rem;
  }
  .nh-assistant-empty h2 {
    font-size: 1.15rem;
  }
  .nh-assistant-empty p {
    max-width: 34rem;
    color: var(--nh-text-muted);
    font-size: 0.85rem;
  }
  .nh-assistant-empty-input {
    width: min(48rem, 100%);
    margin-top: 0.7rem;
  }
</style>
