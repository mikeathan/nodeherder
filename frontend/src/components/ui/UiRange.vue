<script setup lang="ts">
  import { computed, ref } from 'vue';
  const props = withDefaults(
    defineProps<{
      value: number;
      min: number;
      max: number;
      step?: number;
      label: string;
      valueText: string;
      disabled?: boolean;
      vertical?: boolean;
      temperature?: boolean;
    }>(),
    { step: 1 }
  );
  const emit = defineEmits<{ preview: [value: number]; commit: [value: number]; start: []; cancel: [] }>();
  const cancelled = ref(false);
  const percent = computed(() =>
    Math.max(0, Math.min(100, ((props.value - props.min) / (props.max - props.min)) * 100))
  );
  const style = computed(() => ({
    '--range-percent': `${percent.value}%`,
    '--grip-position': `${Math.max(3, Math.min(97, percent.value - 3))}%`,
  }));
  function input(event: Event) {
    emit('preview', Number((event.target as HTMLInputElement).value));
  }
  function start() {
    cancelled.value = false;
    emit('start');
  }
  function cancel() {
    cancelled.value = true;
    emit('cancel');
  }
  function key(event: KeyboardEvent) {
    if (event.key === 'Escape') cancel();
    else cancelled.value = false;
  }
  function change(event: Event) {
    if (!cancelled.value) emit('commit', Number((event.target as HTMLInputElement).value));
  }
</script>
<template>
  <div
    class="nh-range"
    :class="{ 'is-vertical': vertical, 'is-temperature': temperature, 'is-disabled': disabled }"
    :style="style">
    <div class="nh-range-track" aria-hidden="true" />
    <span class="nh-range-grip" aria-hidden="true" />
    <input
      type="range"
      :value="value"
      :min="min"
      :max="max"
      :step="step"
      :disabled="disabled"
      :aria-label="label"
      :aria-valuetext="valueText"
      :aria-orientation="vertical ? 'vertical' : 'horizontal'"
      @input="input"
      @change="change"
      @pointerdown="start"
      @pointercancel="cancel"
      @keydown="key" />
  </div>
</template>
<style scoped>
  .nh-range {
    position: relative;
    width: 100%;
    height: 3rem;
    border-radius: var(--nh-radius-m);
    isolation: isolate;
  }
  .nh-range-track {
    position: absolute;
    inset: 0;
    border-radius: inherit;
    background: linear-gradient(
      to right,
      var(--range-color, var(--nh-accent)) var(--range-percent),
      var(--nh-surface-2) var(--range-percent)
    );
  }
  .nh-range-grip {
    position: absolute;
    width: 6px;
    height: 40%;
    border-radius: 99px;
    background: var(--nh-range-grip);
    left: var(--grip-position);
    top: 30%;
    box-shadow: 0 1px 4px rgb(0 0 0 / 0.2);
    pointer-events: none;
  }
  input {
    position: absolute;
    inset: 0;
    margin: 0;
    width: 100%;
    height: 100%;
    opacity: 0;
    cursor: pointer;
    touch-action: none;
  }
  .nh-range:focus-within {
    outline: 3px solid var(--nh-focus);
    outline-offset: 5px;
  }
  .nh-range.is-vertical {
    width: clamp(7.5rem, 20vw, 10rem);
    height: clamp(12rem, 43dvh, 26rem);
    border-radius: 2.5rem;
  }
  .is-vertical .nh-range-track {
    background: linear-gradient(
      to top,
      var(--range-color, var(--nh-accent)) var(--range-percent),
      var(--nh-range-empty) var(--range-percent)
    );
  }
  .is-vertical .nh-range-grip {
    width: 45%;
    height: 6px;
    left: 27.5%;
    top: auto;
    bottom: var(--grip-position);
  }
  .is-vertical input {
    writing-mode: vertical-lr;
    direction: rtl;
  }
  .nh-range.is-temperature .nh-range-track {
    background: linear-gradient(to top, var(--nh-range-warm), var(--nh-range-cool));
  }
  .is-disabled {
    opacity: 0.5;
  }
  .is-disabled input {
    cursor: not-allowed;
  }
</style>
