<script setup lang="ts">
  import { computed, ref } from 'vue';
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
  // Pointer presentation follows the finger continuously. Only the command draft
  // snaps to native steps, so coarse device steps cannot make the handle stutter.
  const pointerPercent = ref<number | null>(null);
  let activePointer: number | null = null;
  const percent = computed(() => pointerPercent.value ?? rangePercent(props.value, props.min, props.max));
  const style = computed(() => ({
    '--range-percent': `${percent.value}%`,
    '--range-level': percent.value / 100,
  }));
  function pointerValue(event: PointerEvent) {
    const input = event.currentTarget as HTMLInputElement;
    const box = input.getBoundingClientRect();
    const fraction = props.vertical
      ? (box.bottom - event.clientY) / box.height
      : (event.clientX - box.left) / box.width;
    const bounded = Math.max(0, Math.min(1, fraction));
    pointerPercent.value = bounded * 100;
    const raw = props.min + bounded * (props.max - props.min);
    const stepped = props.min + Math.round((raw - props.min) / props.step) * props.step;
    const value = Math.max(props.min, Math.min(props.max, Number(stepped.toFixed(8))));
    input.value = String(value);
    emit('preview', value);
    return value;
  }
  function pointerDown(event: PointerEvent) {
    if (props.disabled || event.button > 0 || activePointer !== null) return;
    event.preventDefault();
    start();
    activePointer = event.pointerId;
    const input = event.currentTarget as HTMLInputElement;
    input.focus({ preventScroll: true });
    keyboard.value = false;
    if (event.isTrusted) input.setPointerCapture(event.pointerId);
    pointerValue(event);
  }
  function pointerMove(event: PointerEvent) {
    if (activePointer !== event.pointerId) return;
    pointerValue(event);
  }
  function pointerUp(event: PointerEvent) {
    if (activePointer !== event.pointerId) return;
    const value = pointerValue(event);
    activePointer = null;
    pointerPercent.value = null;
    emit('commit', value);
    end();
  }
  function lostCapture() {
    if (activePointer !== null) cancel();
  }
  function input(event: Event) {
    emit('preview', Number((event.target as HTMLInputElement).value));
  }
  function start() {
    cancelled.value = false;
    keyboard.value = false;
    emit('start');
  }
  function end() {
    emit('end');
  }
  function cancel() {
    cancelled.value = true;
    activePointer = null;
    pointerPercent.value = null;
    emit('cancel');
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
    if (!cancelled.value) emit('commit', Number((event.target as HTMLInputElement).value));
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
      :value="value"
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
      @pointerdown="pointerDown"
      @pointermove="pointerMove"
      @pointerup="pointerUp"
      @lostpointercapture="lostCapture"
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
    cursor: pointer;
    touch-action: none;
  }
  .nh-range.is-keyboard {
    outline: 3px solid var(--nh-focus);
    outline-offset: 5px;
  }
  .nh-range.is-vertical {
    width: clamp(7.5rem, 20vw, 10rem);
    height: clamp(10rem, min(40dvh, calc(100dvh - 23rem)), 22rem);
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
    writing-mode: vertical-lr;
    direction: rtl;
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
