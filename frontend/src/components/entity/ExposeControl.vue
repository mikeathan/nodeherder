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
    exposeKind,
    exposeLabel,
    exposePresets,
    exposeRange,
    formatExposeValue,
    isBinaryOn,
  } from '@/domain/exposes';
  import { useExposeEdit } from '@/composables/useExposeEdit';
  import { numericStep, sliderValue } from '@/domain/controls';
  import UiRange from '@/components/ui/UiRange.vue';
  import UiToggle from '@/components/ui/UiToggle.vue';
  import UiSelect from '@/components/ui/UiSelect.vue';

  const props = defineProps<{ device: Device; expose: Expose; label?: string; expanded?: boolean; card?: boolean }>();

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
  const unknown = computed(
    () => displayValue.value === null || displayValue.value === undefined || displayValue.value === ''
  );
  const sliderText = computed(() =>
    props.card && unknown.value ? 'Unknown' : formatExposeValue(props.expose, displayValue.value)
  );
  const position = computed(() => sliderValue(props.expose, numberValue.value));
  function previewSlider(value: number) {
    preview(sliderValue(props.expose, value));
  }
  function commitSlider(value: number) {
    sendNumber(sliderValue(props.expose, value));
  }
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
  <div
    class="nh-xc"
    :class="[`is-${control}`, { 'is-card': card, 'has-presets': presets.length > 0 }]"
    :aria-busy="pending || undefined"
    :style="{ '--range-color': `var(--nh-kind-${exposeKind(expose.name)})` }">
    <UiToggle
      v-if="control === 'switch'"
      :model-value="isBinaryOn(expose)"
      :label="name"
      :disabled="disabled"
      :pending="pending"
      @update:model-value="commit($event ? binaryOnValue(expose) : binaryOffValue(expose))" />

    <template v-else-if="control === 'slider' && range">
      <span v-if="card" class="nh-xc-label">{{ name }}</span>
      <UiRange
        class="nh-xc-range"
        :min="range.min"
        :max="range.max"
        :step="step"
        :value="position"
        :label="name"
        :value-text="sliderText"
        :disabled="disabled"
        :busy="pending"
        :unknown="unknown"
        :temperature="expose.name === 'color_temp'"
        @start="dragging = true"
        @end="dragging = false"
        @preview="previewSlider"
        @commit="commitSlider"
        @cancel="cancel" />
      <output class="nh-xc-out">{{ sliderText }}</output>
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

    <!-- Entity views favour directly reachable choices; rows and long lists keep
         the native select. Pressed state always comes from the confirmed expose. -->
    <div
      v-else-if="control === 'choice' && expanded && presets.length <= 5"
      class="choice-options"
      role="group"
      :aria-label="name">
      <button
        v-for="preset in presets"
        :key="preset.label"
        type="button"
        :aria-pressed="expose.data === preset.value"
        :disabled="disabled"
        @click="commit(preset.value)">
        {{ preset.label }}
      </button>
    </div>
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
    flex: 0 1 auto;
    flex-wrap: nowrap;
  }
  .choice-options {
    display: flex;
    flex-wrap: wrap;
    gap: 0.5rem;
  }
  .choice-options button {
    min-height: 44px;
    padding: 0.6rem 0.9rem;
    border-radius: var(--nh-radius-s);
    border: 1px solid var(--nh-border-strong);
    background: var(--nh-surface-2);
    color: var(--nh-text);
    font: inherit;
    cursor: pointer;
  }
  .choice-options button[aria-pressed='true'] {
    border-color: var(--nh-accent);
    background: var(--nh-accent-soft);
    color: var(--nh-accent);
  }
  .choice-options button:focus-visible {
    outline: 3px solid var(--nh-focus);
    outline-offset: 3px;
  }
  .choice-options button:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }
  .nh-xc[aria-busy='true'] :deep(.nh-input:disabled),
  .nh-xc[aria-busy='true'] :deep(.nh-select:disabled),
  .nh-xc[aria-busy='true'] .choice-options button:disabled {
    opacity: 1;
  }
  .nh-xc-range {
    flex: 1 1 9rem;
    min-width: 7rem;
    max-width: 7.5rem;
  }
  .nh-xc.is-slider .nh-xc-range {
    height: 44px;
    border-radius: var(--nh-radius-s);
  }
  /* Horizontal row controls share a quiet rail; the entity view renders its
     large vertical UiRange directly and keeps its independent presentation. */
  .nh-xc.is-slider .nh-xc-range :deep(.nh-range-visual) {
    inset: calc(50% - 2px) 0 auto;
    height: 4px;
    border-radius: 99px;
  }
  .nh-xc.is-slider .nh-xc-range :deep(.nh-range-track) {
    background: color-mix(in srgb, var(--nh-text-muted) 30%, var(--nh-surface));
  }
  .nh-xc.is-slider .nh-xc-range :deep(.nh-range-fill) {
    background: var(--nh-text-muted);
    border-radius: inherit;
  }
  .nh-xc.is-slider .nh-xc-range :deep(.nh-range-grip) {
    width: 16px;
    height: 16px;
    left: clamp(8px, var(--range-percent), calc(100% - 8px));
    border: 2px solid var(--nh-text-muted);
    border-radius: 50%;
    background: var(--nh-surface);
    box-shadow: var(--nh-shadow-1);
  }
  .nh-xc.is-slider .nh-xc-range :deep(input) {
    inset: 0;
    height: 100%;
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
  @media (max-width: 480px) {
    .nh-xc.is-slider {
      flex-wrap: wrap;
    }
    .nh-xc-presets {
      width: 100%;
    }
  }
  .nh-xc.is-slider:not(.is-card) {
    display: grid;
    grid-template-columns: minmax(0, 1fr) auto 4.5rem;
    flex: 0 1 20rem;
    width: 20rem;
    max-width: 100%;
  }
  .nh-xc.is-slider:not(.is-card) .nh-xc-range {
    grid-column: 1 / 3;
    grid-row: 1;
    min-width: 0;
    max-width: none;
  }
  .nh-xc.is-slider.has-presets:not(.is-card) .nh-xc-range {
    grid-column: 1;
  }
  .nh-xc.is-slider:not(.is-card) .nh-xc-out {
    grid-column: 3;
    grid-row: 1;
    min-width: 0;
    white-space: nowrap;
  }
  .nh-xc.is-slider:not(.is-card) .nh-xc-presets {
    grid-column: 2;
    grid-row: 1;
    width: 6rem;
    font-size: 0.8rem;
  }
  @media (max-width: 640px) {
    .nh-xc.is-slider:not(.is-card) {
      grid-template-columns: minmax(0, 1fr) 4.5rem;
      flex-basis: 100%;
      width: 100%;
    }
    .nh-xc.is-slider:not(.is-card) .nh-xc-range,
    .nh-xc.is-slider.has-presets:not(.is-card) .nh-xc-range {
      grid-column: 1;
    }
    .nh-xc.is-slider:not(.is-card) .nh-xc-out {
      grid-column: 2;
    }
    .nh-xc.is-slider:not(.is-card) .nh-xc-presets {
      grid-column: 1 / -1;
      grid-row: 2;
      justify-self: end;
    }
  }
  /* Cards retain their existing label/value header and control row. */
  .nh-xc.is-card.is-slider {
    display: grid;
    grid-template-columns: minmax(0, 1fr) auto;
    align-items: center;
    gap: 0 0.75rem;
    width: 100%;
  }
  .is-card .nh-xc-label {
    grid-column: 1;
    grid-row: 1;
    min-width: 0;
    font-size: 0.9rem;
    line-height: 1.4;
  }
  .is-card .nh-xc-out {
    grid-column: 2;
    grid-row: 1;
    min-width: 0;
    font-size: 0.85rem;
    white-space: nowrap;
  }
  .nh-xc.is-card.is-slider .nh-xc-range {
    grid-column: 1 / -1;
    grid-row: 2;
    min-width: 0;
    max-width: none;
    height: 44px;
    border-radius: var(--nh-radius-s);
  }
  .nh-xc.is-card.has-presets .nh-xc-range {
    grid-column: 1;
  }
  .nh-xc.is-card .nh-xc-presets {
    grid-column: 2;
    grid-row: 2;
    width: 5.25rem;
    min-height: 2rem;
    padding-block: 0.25rem;
    padding-left: 0.5rem;
    font-size: 0.75rem;
  }
</style>
