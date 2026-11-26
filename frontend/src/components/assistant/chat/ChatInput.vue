<script setup lang="ts">
  /* Composer with the shell's input/button tokens; Enter sends, Shift+Enter adds a line. */
  import { computed, nextTick, ref, watch } from 'vue';
  import UiButton from '@/components/ui/UiButton.vue';

  const props = defineProps<{ modelValue: string; isLoading?: boolean }>();
  const emit = defineEmits<{ (e: 'update:modelValue', value: string): void; (e: 'send'): void }>();

  const field = ref<HTMLTextAreaElement | null>(null);
  const canSend = computed(() => !!props.modelValue.trim() && !props.isLoading);

  function resize() {
    const element = field.value;
    if (!element) return;
    element.style.height = 'auto';
    element.style.height = `${Math.min(element.scrollHeight, 180)}px`;
  }

  watch(() => props.modelValue, () => void nextTick(resize));

  function onInput(event: Event) {
    emit('update:modelValue', (event.target as HTMLTextAreaElement).value);
    resize();
  }
  function send() {
    if (canSend.value) emit('send');
  }
  function onKeydown(event: KeyboardEvent) {
    if (event.key !== 'Enter' || event.shiftKey) return;
    event.preventDefault();
    send();
  }
</script>

<template>
  <div class="nh-chat-input">
    <textarea
      ref="field"
      :value="modelValue"
      rows="1"
      class="nh-chat-field"
      placeholder="Ask about your home…"
      aria-label="Message the assistant"
      @input="onInput"
      @keydown="onKeydown" />
    <UiButton variant="primary" icon="send" label="Send message" :disabled="!canSend" @click="send">
      <span class="nh-chat-send-label">Send</span>
    </UiButton>
  </div>
</template>

<style scoped>
  .nh-chat-input {
    width: 100%;
    max-width: 48rem;
    margin: 0 auto;
    display: flex;
    align-items: flex-end;
    gap: 0.45rem;
    padding: 0.35rem;
    border: var(--nh-border-w) solid var(--nh-border-strong);
    border-radius: var(--nh-radius-l);
    background: var(--nh-surface-2);
  }
  .nh-chat-field {
    flex: 1;
    min-width: 0;
    min-height: 2.35rem;
    max-height: 11rem;
    resize: none;
    border: 0;
    background: transparent;
    color: var(--nh-text);
    font: inherit;
    line-height: 1.45;
    padding: 0.55rem 0.6rem;
  }
  .nh-chat-field:focus {
    outline: none;
  }
  .nh-chat-input:focus-within {
    border-color: var(--nh-focus);
    box-shadow: 0 0 0 2px var(--nh-accent-soft);
  }
  @media (max-width: 640px) {
    .nh-chat-send-label {
      display: none;
    }
  }
</style>
