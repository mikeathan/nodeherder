<script setup lang="ts" generic="T">
  /* Segmented choice for 2–6 short options. */
  defineProps<{ modelValue: T; options: readonly { value: T; label: string }[]; label: string; disabled?: boolean }>();
  defineEmits<{ (e: 'update:modelValue', value: T): void }>();
</script>

<template>
  <div class="nh-seg" role="group" :aria-label="label">
    <button
      v-for="opt in options"
      :key="String(opt.value)"
      type="button"
      :aria-pressed="opt.value === modelValue"
      :disabled="disabled"
      @click="$emit('update:modelValue', opt.value)">
      {{ opt.label }}
    </button>
  </div>
</template>

<style scoped>
  .nh-seg {
    display: inline-flex;
    flex-wrap: wrap;
    border: var(--nh-border-w) solid var(--nh-border-strong);
    border-radius: var(--nh-radius-m);
    overflow: hidden;
    background: var(--nh-surface);
    max-width: 100%;
  }
  .nh-seg button {
    border: 0;
    background: none;
    color: var(--nh-text);
    font-size: 0.82rem;
    padding: 0 0.75rem;
    min-height: calc(var(--nh-ctl-h) * 0.85);
    cursor: pointer;
    border-right: var(--nh-border-w) solid var(--nh-border);
  }
  .nh-seg button:last-child {
    border-right: 0;
  }
  .nh-seg button[aria-pressed='true'] {
    background: var(--nh-accent);
    color: var(--nh-accent-contrast);
    font-weight: 600;
  }
  .nh-seg button:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }
</style>
