<script setup lang="ts">
  // Presentation only: choosing Custom opens an editor without copying/saving an
  // inherited value. Undefined and explicit zero stay distinct in the saved policy.
  import { computed, ref } from 'vue';
  import type { TimeInterval } from '@/types/types.type';
  import { intervalSeconds } from '@/domain/reporting-settings';
  import ReportingIntervalInput from './ReportingIntervalInput.vue';

  const props = defineProps<{ label: string; value?: TimeInterval; inherited: string; disabled?: boolean; error?: string }>();
  const emit = defineEmits<{ (event: 'change', input: string): void }>();
  const editingCustom = ref(false);
  const mode = computed(() => editingCustom.value || (props.value && intervalSeconds(props.value) > 0) ? 'custom' : props.value ? 'every' : 'inherit');
  function changeMode(input: string) {
    editingCustom.value = input === 'custom';
    if (input !== 'custom') emit('change', input === 'every' ? '0' : '');
  }
</script>

<template>
  <div class="rate-control">
    <label>
      <span>{{ label }}</span>
      <select :aria-label="label" :value="mode" :disabled="disabled" @change="changeMode(($event.target as HTMLSelectElement).value)">
        <option value="inherit">{{ inherited }}</option>
        <option value="every">Every change</option>
        <option value="custom">Custom interval…</option>
      </select>
    </label>
    <ReportingIntervalInput v-if="mode === 'custom'" label="Time between updates" :aria-label="`${label} seconds`" :value="value" :disabled="disabled" :error="error" @change="emit('change', $event)" />
    <small v-else-if="error" role="alert">{{ error }}</small>
  </div>
</template>

<style scoped>
  .rate-control, label { display: grid; gap: .4rem; min-width: 0; font-size: .85rem; }
  select { width: 100%; min-height: 2.75rem; padding: .65rem; border: 1px solid var(--nh-border); border-radius: var(--nh-radius-s); background: var(--nh-surface); color: var(--nh-text); font: inherit; }
  small { color: var(--nh-danger); }
</style>
