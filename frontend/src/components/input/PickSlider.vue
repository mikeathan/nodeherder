<script setup lang="ts">
  import { ref, computed, watch } from 'vue';
  import { ControlDirection} from '@/types/controls.type';

  const props = withDefaults(
    defineProps<{
      value?: number;
      min?: number;
      max?: number;
      disabled?: boolean;
      color?: string;
      direction?: ControlDirection;
    }>(),
    {
      value: 50,
      min: 0,
      max: 100,
      disabled: false,
      color: '#ffc107',
      direction: 'horizontal',
    }
  );

  const emit = defineEmits<{
    (e: 'update', value: number): void;
  }>();

  const value = ref(props.value);
  const lastKnownValue = ref(props.value);


  const trackPercentRaw = computed(() => {
    const { min, max } = props;
    const clamped = Math.min(Math.max(value.value, min), max);
    return ((clamped - min) / (max - min)) * 100;
  });

  const trackFill = computed(() => {
    const color = props.disabled ? '#ccc' : props.color;
    const bgColor = props.disabled ? '#ccc' : '#fff2cc';

    const direction = props.direction === 'vertical' ? 'to top' : 'to right';

    return {
      background: `linear-gradient(${direction}, ${color}, ${bgColor})`,
      opacity: props.disabled ? 0.5 : 1,
    };
  });

  watch(
    () => props.value,
    (newVal) => {
      if (!props.disabled) {
        value.value = newVal;
        lastKnownValue.value = newVal;
      }
    }
  );

  watch(
    () => props.disabled,
    (isDisabled) => {
      if (isDisabled) {
        lastKnownValue.value = value.value;
        value.value = props.min;
      } else {
        value.value = lastKnownValue.value;
      }
    }
  );

  function updateValue(event: any): void {
    const newValue = parseInt(event.target.value);
    emit('update', newValue);
  }

  const generateDirectionalClass = (baseClass: string) => [
    baseClass,
    props.direction === 'vertical' ? 'vertical' : 'horizontal',
    { disabled: props.disabled },
  ];
</script>

<template>
  <div class="slider-container" :class="props.direction">
    <div :class="generateDirectionalClass('slider-wrapper')">
      <div :class="generateDirectionalClass('track-background')" :style="trackFill"></div>
      <div
        class="tap-indicator"
        :class="[props.direction, 'tap']"
        :style="{
          '--indicator-percent': `${trackPercentRaw}%`,
          '--indicator-color': props.color,
        }"
        v-show="true" />

      <input
        type="range"
        :min="props.min"
        :max="props.max"
        v-model="value"
        class="slider"
        @change.stop="updateValue"
        @click.stop
        @mousedown.stop
        @pointerdown.stop />
    </div>

    <!-- Vertical -->
    <div v-if="props.direction === 'vertical'" class="value-marker vertical" :style="{ bottom: `${trackPercentRaw}%` }">
      — {{ value }} —
    </div>
  </div>
</template>

<style scoped>
  .slider-container {
    position: relative;
    display: inline-block;
    width: 100%;
  }

  .slider-container.vertical {
    height: 320px;
    width: auto;
  }
  /* === Slider Wrapper === */
  .slider-wrapper {
    position: relative;
    height: 46px;
    border-radius: 12px;
    overflow: hidden;
  }

  .slider-wrapper.vertical {
    height: 320px;
    width: 130px;
    border-radius: 36px;
  }

  /* === Track Background === */
  .track-background {
    position: absolute;
    inset: 0;
    pointer-events: none;
  }

  .track-background.horizontal {
    border-radius: 12px;
  }

  .track-background.vertical {
    border-radius: 36px;
  }

  /* === Slider Input Layer === */
  .slider {
    appearance: none;
    width: 100%;
    height: 100%;
    background: transparent;
    position: relative;
    z-index: 1;
    cursor: pointer;
    outline: none;
  }

  /* Vertical slider positioning via rotation */
  .vertical .slider {
    position: absolute;
    top: 50%;
    left: 50%;
    width: var(--wrapper-height, 320px);
    height: var(--wrapper-width, 130px);
    transform: translate(-50%, -50%) rotate(-90deg);
    margin: 0;
    transform-origin: center center;
  }

  /* Hide default slider thumbs */
  .slider::-webkit-slider-thumb,
  .slider::-moz-range-thumb {
    visibility: hidden;
    width: 0;
    height: 0;
    border: none;
  }

  .slider:disabled {
    opacity: 0.6;
  }

  /* === Tap Indicator (Common Base) === */
  .tap-indicator {
    position: absolute;
    pointer-events: none;
    z-index: 2;
    background-color: var(--indicator-color, #ccc);
  }

  /* === Horizontal Tap Indicator Styles === */
  .tap-indicator.horizontal.tap {
    width: 40px;
    height: 100%;
    background: white;
    box-shadow: 4px 0 6px -4px;
    left: var(--indicator-percent);
    top: 50%;
    transform: translate(-50%, -50%);

    /* Edge-aware rounded corners */
    --at-left-edge: max(0, min(1, (10% - var(--indicator-percent)) / 10%));
    --at-right-edge: max(0, min(1, (var(--indicator-percent) - 90%) / 10%));
    border-top-left-radius: calc(8px + (var(--at-left-edge) * 28px));
    border-bottom-left-radius: calc(8px + (var(--at-left-edge) * 28px));
    border-top-right-radius: calc(8px + (var(--at-right-edge) * 28px));
    border-bottom-right-radius: calc(8px + (var(--at-right-edge) * 28px));
  }

  .tap-indicator.horizontal.tap::before {
    content: '';
    position: absolute;
    top: 25%;
    left: 50%;
    width: 4px;
    height: 50%;
    background-color: #555;
    border-radius: 2px;
    transform: translateX(-50%);
  }


  /* === Vertical Tap Indicator Styles === */
  .tap-indicator.vertical.tap {
    width: 100%;
    height: 40px;
    background: white;
    box-shadow: 0 4px 6px -4px;
    top: calc(100% - var(--indicator-percent));
    left: 50%;
    transform: translate(-50%, -50%);

    /* Edge-aware rounded corners */
    --at-top-edge: max(0, min(1, (10% - var(--indicator-percent)) / 10%));
    --at-bottom-edge: max(0, min(1, (var(--indicator-percent) - 90%) / 10%));
    border-top-left-radius: calc(8px + (var(--at-top-edge) * 28px));
    border-top-right-radius: calc(8px + (var(--at-top-edge) * 28px));
    border-bottom-left-radius: calc(8px + (var(--at-bottom-edge) * 28px));
    border-bottom-right-radius: calc(8px + (var(--at-bottom-edge) * 28px));
  }

  .tap-indicator.vertical.tap::before {
    content: '';
    position: absolute;
    top: 50%;
    left: 25%;
    width: 50%;
    height: 4px;
    background-color: #555;
    border-radius: 2px;
    transform: translateY(-50%);
  }

  .slider:disabled {
    opacity: 0.6;
  }

  /* TEST */
  .value-marker {
    position: absolute;
    color:white;
    font-size: 12px;
    font-weight: bold;
    pointer-events: none;
    z-index: 1000;
    white-space: nowrap;
  }

  /* Horizontal: show below or above the slider */
  .value-marker.horizontal {
    top: 100%; 
    margin-top: 6px;
    transform: translateX(-50%);
  }

  /* Vertical: show to the left of the slider */
  .value-marker.vertical {
    left: -50px;
    transform: translateY(50%);
    border: 2px solid red;
  }

  /* Marks Container */
  .slider-marks {
    position: absolute;
    pointer-events: none;
    font-size: 10px;
    color: #333;
  }

  .slider-marks.horizontal {
    width: 100%;
    bottom: -10px;
    display: flex;
    justify-content: space-between;
    position: absolute;
  }

  .slider-marks.vertical {
    height: 100%;
    right: -10px;
    display: flex;
    flex-direction: column;
    justify-content: space-between;
    position: absolute;
  }

  /* Individual mark */
  .mark {
    position: absolute;
    white-space: nowrap;
  }
</style>
