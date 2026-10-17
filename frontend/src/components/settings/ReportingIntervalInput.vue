<script setup lang="ts">
  import { intervalSeconds } from '@/domain/reporting-settings';
  import type { TimeInterval } from '@/types/types.type';
  defineProps<{ label: string; ariaLabel?: string; value?: TimeInterval; inherited?: string; inheritedSeconds?: number; disabled?: boolean; error?: string }>();
  defineEmits<{ (event: 'change', value: string): void }>();
</script>

<template>
  <label class="reporting-input">
    <span>{{ label }}</span>
    <div class="reporting-number">
      <input
        type="number"
        min="0"
        step="0.1"
        :aria-label="ariaLabel || label"
        :value="value ? intervalSeconds(value) : ''"
        :placeholder="String(inheritedSeconds ?? 0)"
        :disabled="disabled"
        :aria-invalid="!!error || undefined"
        @change="$emit('change', ($event.target as HTMLInputElement).value)"
        />
      <span aria-hidden="true">sec</span>
    </div>
    <small v-if="error" role="alert" class="reporting-error">{{ error }}</small>
    <small v-if="inherited">{{ inherited }}</small>
  </label>
</template>

<style scoped>
  .reporting-input { display: flex; flex-direction: column; gap: .4rem; min-width: 0; font-size: .85rem; }
  .reporting-number { display: flex; align-items: center; border: 1px solid var(--nh-border); border-radius: var(--nh-radius-s); background: var(--nh-surface); }
  input { width: 100%; min-width: 0; min-height: 2.75rem; padding: .65rem; color: var(--nh-text); background: transparent; border: 0; border-radius: inherit; font: inherit; }
  .reporting-number > span { padding-right: .75rem; color: var(--nh-text-muted); }
  .reporting-error { color: var(--nh-danger); }
  small { color: var(--nh-text-muted); line-height: 1.5; }
</style>
