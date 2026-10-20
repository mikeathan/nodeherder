<script setup lang="ts">
  import { computed } from 'vue';
  import type { Device, Expose } from '@/types/device';
  import {
    colourTemperatureName,
    exposeKind,
    exposeLabel,
    exposeRange,
    formatExposeValue,
  } from '@/domain/exposes';
  import { displayNumber, numericStep, sliderValue } from '@/domain/controls';
  import { useExposeEdit } from '@/composables/useExposeEdit';
  import UiRange from '@/components/ui/UiRange.vue';
  import ExposeControl from './ExposeControl.vue';
  const props = defineProps<{ device: Device; expose: Expose }>();
  const edit = useExposeEdit(
    () => props.device,
    () => props.expose
  );
  const { value, displayValue, pending, disabled, dragging } = edit;
  const range = computed(() => exposeRange(props.expose));
  const position = computed(() => sliderValue(props.expose, value.value));
  const label = computed(() => exposeLabel(props.expose.name));
  const temperature = computed(() => props.expose.name === 'color_temp');
  /** Missing data is a state, not a value: never draw it as a minimum fill or a fake dash handle. */
  const unknown = computed(
    () => displayValue.value === null || displayValue.value === undefined || displayValue.value === ''
  );
  const text = computed(() => (unknown.value ? 'Unknown' : formatExposeValue(props.expose, displayValue.value)));
  const descriptor = computed(() => {
    if (!temperature.value || unknown.value || typeof displayValue.value !== 'number') return '';
    return colourTemperatureName(displayNumber(props.expose, displayValue.value));
  });
  const bottom = computed(() =>
    range.value ? formatExposeValue(props.expose, sliderValue(props.expose, range.value.min)) : ''
  );
  const top = computed(() =>
    range.value ? formatExposeValue(props.expose, sliderValue(props.expose, range.value.max)) : ''
  );
  const topText = computed(() => (temperature.value ? `Cool · ${top.value}` : top.value));
  const bottomText = computed(() => (temperature.value ? `Warm · ${bottom.value}` : bottom.value));
  const status = computed(() => {
    if (dragging.value) return 'Release to apply';
    return unknown.value ? 'No reading' : '';
  });
  function start() {
    dragging.value = true;
  }
  function end() {
    dragging.value = false;
  }
  function preview(value: number) {
    edit.preview(sliderValue(props.expose, value));
  }
  function commit(value: number) {
    edit.commitNumber(sliderValue(props.expose, value));
  }
</script>
<template>
  <div class="entity-adjustment" :style="{ '--range-color': `var(--nh-kind-${exposeKind(expose.name)})` }">
    <template v-if="range">
      <output
        class="entity-adjustment-value"
        :class="{ 'is-unknown': unknown }"
        :aria-label="`${label} preview`"
        >{{ text }}</output>
      <p class="entity-adjustment-detail">{{ descriptor }}</p>
      <span class="adjustment-status" role="status">{{ status }}</span>
      <div class="entity-adjustment-track">
        <UiRange
          :value="position"
          :min="range.min"
          :max="range.max"
          :step="numericStep(expose)"
          :label="label"
          :value-text="text"
          :disabled="disabled"
          :busy="pending"
          :unknown="unknown"
          vertical
          :temperature="temperature"
          @start="start"
          @end="end"
          @preview="preview"
          @commit="commit"
          @cancel="edit.cancel" />
        <div class="range-endpoints" aria-hidden="true">
          <span>{{ topText }}</span
          ><span>{{ bottomText }}</span>
        </div>
      </div>
    </template>
    <ExposeControl v-else :device="device" :expose="expose" />
  </div>
</template>
<style scoped>
  .entity-adjustment {
    display: flex;
    align-items: center;
    flex-direction: column;
    gap: 0.4rem;
  }
  .entity-adjustment-value {
    font-size: clamp(2.5rem, 7vw, 3.5rem);
    font-weight: 600;
    line-height: 1.1;
    letter-spacing: -0.04em;
    font-variant-numeric: tabular-nums;
  }
  .entity-adjustment-value.is-unknown {
    font-size: clamp(1.9rem, 5vw, 2.5rem);
    font-weight: 500;
    color: var(--nh-text-muted);
  }
  .entity-adjustment-detail {
    min-height: 1.25rem;
    font-size: 0.95rem;
    color: var(--nh-text-muted);
    margin: 0;
  }
  .adjustment-status {
    font-size: 0.8rem;
    min-height: 1.25rem;
    color: var(--nh-text-muted);
  }
  .entity-adjustment-track {
    position: relative;
    margin: 0.25rem 0 1rem;
  }
  .range-endpoints {
    position: absolute;
    left: calc(100% + 0.75rem);
    top: 0.4rem;
    bottom: 0.4rem;
    display: flex;
    flex-direction: column;
    justify-content: space-between;
    font-size: 0.75rem;
    white-space: nowrap;
    color: var(--nh-text-muted);
  }
</style>
