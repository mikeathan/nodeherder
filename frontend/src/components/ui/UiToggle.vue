<script setup lang="ts">
  /* On/off switch (role="switch"). `pending` shows a command in flight (FR-04). */
  defineProps<{ modelValue: boolean; label: string; disabled?: boolean; pending?: boolean }>();
  defineEmits<{ (e: 'update:modelValue', value: boolean): void }>();
</script>

<template>
  <button
    type="button"
    role="switch"
    class="nh-toggle"
    :class="{ 'is-pending': pending }"
    :aria-checked="modelValue"
    :aria-label="label"
    :aria-busy="pending || undefined"
    :disabled="disabled"
    @click="$emit('update:modelValue', !modelValue)">
    <span class="nh-toggle-knob" />
  </button>
</template>

<style scoped>
  .nh-toggle {
    position: relative;
    display: inline-flex;
    align-items: center;
    width: 2.9rem;
    height: 1.6rem;
    border-radius: 99px;
    border: var(--nh-border-w) solid var(--nh-border-strong);
    background: var(--nh-surface-2);
    cursor: pointer;
    padding: 0;
    flex: none;
    transition: background var(--nh-motion);
  }
  .nh-toggle-knob {
    position: absolute;
    left: 0.15rem;
    width: 1.2rem;
    height: 1.2rem;
    border-radius: 50%;
    background: #fff;
    box-shadow: 0 1px 3px rgb(0 0 0 / 0.3);
    transition: transform var(--nh-motion);
  }
  .nh-toggle[aria-checked='true'] {
    background: var(--nh-accent);
    border-color: var(--nh-accent);
  }
  .nh-toggle[aria-checked='true'] .nh-toggle-knob {
    transform: translateX(1.3rem);
  }
  .nh-toggle:disabled {
    opacity: 0.45;
    cursor: not-allowed;
  }
  .nh-toggle.is-pending .nh-toggle-knob {
    animation: nh-pulse 0.9s infinite;
  }
</style>
