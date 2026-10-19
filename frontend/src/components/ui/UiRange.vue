<script setup lang="ts">
  import { computed, ref, watch } from 'vue';
  import { rangePercent } from '@/domain/controls';
  const props = withDefaults(
    defineProps<{
      value: number;
      min: number;
      max: number;
      step?: number;
      label: string;
      valueText: string;
      disabled?: boolean;
      busy?: boolean;
      vertical?: boolean;
      temperature?: boolean;
      /** No confirmed reading or draft: show a muted rail, never a fake minimum fill. */
      unknown?: boolean;
    }>(),
    { step: 1, disabled: false, busy: false, vertical: false, temperature: false, unknown: false }
  );
  const emit = defineEmits<{
    preview: [value: number];
    commit: [value: number];
    start: [];
    end: [];
    cancel: [];
  }>();
  const cancelled = ref(false);
  const keyboard = ref(false);
  const dragging = ref(false);
  // The browser owns pointer movement. Keep its local value separate from hub
  // reports so no second pointer writer or live update can fight a native drag.
  const localValue = ref(props.value);
  const percent = computed(() => rangePercent(localValue.value, props.min, props.max));
  const style = computed(() => ({
    '--range-percent': `${percent.value}%`,
    '--range-level': percent.value / 100,
  }));
  watch(() => props.value, (value) => {
    if (!dragging.value) localValue.value = value;
  });
  watch(() => props.disabled, (disabled) => {
    if (disabled && dragging.value) cancel();
  });
  function input(event: Event) {
    emit('preview', Number((event.target as HTMLInputElement).value));
  }
  function start() {
    if (props.disabled) return;
    cancelled.value = false;
    keyboard.value = false;
    dragging.value = true;
    emit('start');
  }
  function end() {
    dragging.value = false;
    emit('end');
  }
  function cancel() {
    cancelled.value = true;
    dragging.value = false;
    emit('cancel');
    localValue.value = props.value;
  }
  function key(event: KeyboardEvent) {
    keyboard.value = true;
    if (event.key === 'Escape') cancel();
    else cancelled.value = false;
  }
  function focus(event: FocusEvent) {
    keyboard.value = (event.target as HTMLInputElement).matches(':focus-visible');
  }
  function blur() {
    keyboard.value = false;
  }
  function change(event: Event) {
    if (cancelled.value) return;
    end();
    // Native change commits on release; input events above only update the draft.
    emit('commit', Number((event.target as HTMLInputElement).value));
  }
</script>
<template>
  <div
    class="nh-range"
    :class="{
      'is-vertical': vertical,
      'is-temperature': temperature,
      'is-disabled': disabled,
      'is-busy': busy,
      'is-unknown': unknown,
      'is-keyboard': keyboard,
    }"
    :style="style">
    <div class="nh-range-visual" aria-hidden="true">
      <div class="nh-range-track" />
      <div v-if="!temperature && !unknown" class="nh-range-fill" />
    </div>
    <span v-if="!unknown" class="nh-range-grip" aria-hidden="true" />
    <input
      type="range"
      v-model.number="localValue"
      :min="min"
      :max="max"
      :step="step"
      :disabled="disabled"
      :aria-busy="busy"
      :aria-label="label"
      :aria-valuetext="valueText"
      :aria-orientation="vertical ? 'vertical' : 'horizontal'"
      @input="input"
      @change="change"
      @pointerdown="start"
      @pointerup="end"
      @pointercancel="cancel"
      @focus="focus"
      @blur="blur"
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
    touch-action: none;
  }
  /* Clip only the painted surface, so compact row sliders keep their 44px hit area. */
  .nh-range-visual {
    position: absolute;
    inset: 0;
    overflow: hidden;
    border-radius: inherit;
    pointer-events: none;
  }
  .nh-range-track,
  .nh-range-fill {
    position: absolute;
    inset: 0;
    border-radius: inherit;
    pointer-events: none;
  }
  .nh-range-track {
    background: var(--nh-surface-2);
  }
  .nh-range-fill {
    background: var(--range-color, var(--nh-accent));
    transform: scaleX(var(--range-level));
    transform-origin: left;
    border-radius: 0;
  }
  .nh-range-grip {
    position: absolute;
    width: 6px;
    height: 40%;
    border-radius: 99px;
    background: var(--nh-range-grip);
    left: clamp(3px, var(--range-percent), calc(100% - 3px));
    top: 50%;
    transform: translate(-50%, -50%);
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
    appearance: none;
    padding: 0;
    border: 0;
    cursor: pointer;
    touch-action: none;
  }
  /* Zero-size native thumbs keep native travel aligned with the painted rail.
     The custom grip is decorative; it never writes pointer positions back. */
  input::-webkit-slider-thumb {
    appearance: none;
    width: 0;
    height: 0;
    border: 0;
  }
  input::-moz-range-thumb {
    width: 0;
    height: 0;
    border: 0;
  }
  .nh-range.is-keyboard {
    outline: 3px solid var(--nh-focus);
    outline-offset: 5px;
  }
  .nh-range.is-vertical {
    --range-width: clamp(7.5rem, 20vw, 10rem);
    --range-height: clamp(10rem, min(40dvh, calc(100dvh - 23rem)), 22rem);
    width: var(--range-width);
    height: var(--range-height);
    border-radius: 2.5rem;
  }
  .is-vertical .nh-range-track {
    background: var(--nh-range-empty);
  }
  .is-vertical .nh-range-fill {
    transform: scaleY(var(--range-level));
    transform-origin: bottom;
  }
  .is-vertical .nh-range-grip {
    width: 45%;
    height: 6px;
    left: 50%;
    top: auto;
    bottom: clamp(12px, var(--range-percent), calc(100% - 12px));
    transform: translate(-50%, 50%);
  }
  .is-vertical input {
    /* Match the previous BaseSlider: rotate one native horizontal range instead
       of combining browser-specific vertical behaviour with a pointer engine. */
    inset: auto;
    left: 50%;
    top: 50%;
    width: var(--range-height);
    height: var(--range-width);
    transform: translate(-50%, -50%) rotate(-90deg);
  }
  .nh-range.is-temperature .nh-range-track {
    background: linear-gradient(to right, var(--nh-range-warm), var(--nh-range-cool));
  }
  .nh-range.is-vertical.is-temperature .nh-range-track {
    background: linear-gradient(to top, var(--nh-range-warm), var(--nh-range-cool));
  }
  .nh-range.is-unknown .nh-range-track {
    background: var(--nh-surface-2);
  }
  .is-disabled:not(.is-busy) {
    opacity: 0.5;
  }
  .is-disabled input {
    cursor: not-allowed;
  }
</style>
