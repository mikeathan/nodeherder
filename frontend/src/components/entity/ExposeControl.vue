<script setup lang="ts">
  /*
   * Input for one writable expose (switch, slider, choice or number), chosen from expose
   * metadata (domain/exposes.ts controlOf). Commands go through useDeviceCommand, so pending
   * state, timeouts and the "can control" rules are the same everywhere (FR-04).
   */
  import { computed } from 'vue';
  import { Device, Expose } from '@/types/device';
  import {
    binaryOffValue,
    binaryOnValue,
    controlOf,
    exposeLabel,
    exposePresets,
    exposeRange,
    formatExposeValue,
    isBinaryOn,
  } from '@/domain/exposes';
  import { useExposeEdit } from '@/composables/useExposeEdit';
  import { numericStep } from '@/domain/controls';
  import UiRange from '@/components/ui/UiRange.vue';
  import UiToggle from '@/components/ui/UiToggle.vue';
  import UiSelect from '@/components/ui/UiSelect.vue';

  const props = defineProps<{ device: Device; expose: Expose; label?: string }>();

  const control = computed(() => controlOf(props.expose));
  const name = computed(() => props.label ?? exposeLabel(props.expose.name));
  const range = computed(() => exposeRange(props.expose));
  const presets = computed(() => exposePresets(props.expose));

  const {
    value: numberValue,
    displayValue,
    pending,
    disabled,
    dragging,
    preview,
    commit,
    commitNumber: sendNumber,
    cancel,
  } = useExposeEdit(
    () => props.device,
    () => props.expose
  );
  const step = computed(() => numericStep(props.expose));
  function commitNumber(event: Event) {
    const input = event.target as HTMLInputElement;
    if (!input.value.trim()) {
      cancel();
      return;
    }
    sendNumber(Number(input.value));
  }
</script>

<template>
  <div class="nh-xc" :class="`is-${control}`">
    <UiToggle
      v-if="control === 'switch'"
      :model-value="isBinaryOn(expose)"
      :label="name"
      :disabled="disabled"
      :pending="pending"
      @update:model-value="commit($event ? binaryOnValue(expose) : binaryOffValue(expose))" />

    <template v-else-if="control === 'slider' && range">
      <UiRange
        class="nh-xc-range"
        :min="range.min"
        :max="range.max"
        :step="step"
        :value="numberValue"
        :label="name"
        :value-text="formatExposeValue(expose, displayValue)"
        :disabled="disabled"
        @start="dragging = true"
        @preview="preview"
        @commit="sendNumber"
        @cancel="cancel" />
      <output class="nh-xc-out">{{ formatExposeValue(expose, displayValue) }}</output>
      <UiSelect
        v-if="presets.length"
        class="nh-xc-presets"
        :model-value="null"
        :options="presets.map((p) => ({ value: p.value, label: p.label }))"
        placeholder="Preset"
        :label="`${name} preset`"
        :disabled="disabled"
        @update:model-value="commit" />
    </template>

    <UiSelect
      v-else-if="control === 'choice'"
      :model-value="expose.data"
      :options="presets.map((p) => ({ value: p.value, label: p.label }))"
      :label="name"
      :disabled="disabled"
      @update:model-value="commit" />

    <input
      v-else-if="control === 'number'"
      type="number"
      class="nh-input is-num"
      inputmode="decimal"
      :step="step"
      :value="displayValue"
      :aria-label="name"
      :disabled="disabled"
      @change="commitNumber" />

    <span v-else class="nh-xc-read">{{ formatExposeValue(expose) }}</span>
    <span v-if="pending" class="nh-xc-pending" role="status"><span class="nh-spin" aria-hidden="true" /> Sending…</span>
  </div>
</template>

<style scoped>
  .nh-xc {
    display: inline-flex;
    align-items: center;
    justify-content: flex-end;
    flex-wrap: wrap;
    gap: 0.5rem;
    min-width: 0;
  }
  .nh-xc.is-slider {
    flex: 1 1 14rem;
    flex-wrap: nowrap;
  }
  .nh-xc-range {
    flex: 1 1 8rem;
    min-width: 5rem;
  }
  .nh-xc-out {
    min-width: 3.5rem;
    text-align: right;
    font-weight: 600;
    font-variant-numeric: tabular-nums;
  }
  .nh-xc-presets {
    width: 6.5rem;
  }
  .nh-xc-read {
    font-weight: 600;
    font-variant-numeric: tabular-nums;
  }
  .nh-xc-pending {
    font-size: 0.75rem;
    color: var(--nh-accent);
    white-space: nowrap;
  }
  @media (max-width: 480px) {
    .nh-xc.is-slider {
      flex-wrap: wrap;
    }
    .nh-xc-presets {
      width: 100%;
    }
  }
</style>
