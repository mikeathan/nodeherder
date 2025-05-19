<script setup lang="ts">
  import { ref, computed, watch } from 'vue';
  import { ControlDirection, SliderTick } from '@/types/controls.type';

  const props = withDefaults(
    defineProps<{
      value?: number;
      unit?: string;
      min?: number;
      max?: number;
      disabled?: boolean;
      color?: string;
      direction?: ControlDirection;
      ticks?: SliderTick[];
    }>(),
    {
      value: 50,
      min: 0,
      max: 100,
      disabled: false,
      color: '#ffc107',
      direction: 'horizontal',
      ticks: () => [],
    }
  );

  const emit = defineEmits<{
    (e: 'update', value: number): void;
    (e: 'mousedown', ev: MouseEvent): void;
    (e: 'touchstart', ev: TouchEvent): void;
    (e: 'click', ev: MouseEvent): void;
    (e: 'update', value: number): void;
  }>();

  const value = ref(props.value);
  const lastKnownValue = ref(props.value);

  const trackPercentRaw = computed(() => {
    const { min, max } = props;
    const clamped = Math.min(Math.max(value.value, min), max);
    return ((clamped - min) / (max - min)) * 100;
  });

  const calculateTrackPercent = () => {
    const { min, max } = props;
    const clampedValue = Math.min(Math.max(value.value, min), max);
    const range = max - min;

    if (range === 0) return '0%';

    const percent = ((clampedValue - min) / range) * 100;
    return `${percent.toFixed(1)}%`;
  };

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

  const showTicks = computed(() => {
    return props.ticks?.length > 0;
  });

  function updateValue(event: any): void {
    const newValue = parseInt(event.target.value);
    emit('update', newValue);
  }

  const generateDirectionalClass = (baseClass: string) => [
    baseClass,
    props.direction === 'vertical' ? 'vertical' : 'horizontal',
    { disabled: props.disabled },
  ];

  const handleTickClick = (tick: SliderTick) => {
    if (!props.disabled) {
      value.value = tick.value;
      emit('update', tick.value);
    }
  };

  const calculateTickPosition = (tick: SliderTick) => {
    const { min, max } = props;
    const percent = ((tick.value - min) / (max - min)) * 100;
    return percent;
  };
</script>

<template>
  <div
    @mousedown="$emit('mousedown', $event)"
    @touchstart="$emit('touchstart', $event)"
    @click="$emit('click', $event)">
    <!-- Ticks -->
    <div v-if="showTicks" :class="generateDirectionalClass('tick-container')">
      <div
        v-for="tick in props.ticks"
        :key="tick.value"
        class="tick"
        :class="props.direction"
        :style="
          props.direction === 'vertical'
            ? { bottom: `${calculateTickPosition(tick)}%` }
            : { left: `${calculateTickPosition(tick)}%` }
        "
        @click="handleTickClick(tick)">
        {{ tick.label ?? tick.value }}
      </div>
    </div>

    <div :class="generateDirectionalClass('slider-wrapper')">
      <div :class="generateDirectionalClass('track-background')" :style="trackFill">
        <slot name="track" :track-percent-raw="trackPercentRaw"></slot>
      </div>

      <slot name="indicator" :track-percent-raw="trackPercentRaw" :direction="props.direction" :color="props.color">
      </slot>

      <input
        type="range"
        :min="props.min"
        :max="props.max"
        v-model="value"
        :disabled="props.disabled"
        class="slider"
        @change.stop="updateValue"
        @click.stop />

      <slot name="extra" :value="value" :track-percent-raw="trackPercentRaw"></slot>
    </div>
  </div>
</template>

<style scoped>
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
  /* === Tick Container === */
  .tick-container {
    position: absolute;
    font-size: 12px;
    color: white;
    pointer-events: all;
    user-select: none;
    z-index: 2;
  }

  .tick-container.vertical {
    height: 100%;
    left: -50px;

  }

  .tick-container.horizontal {
    width: 100%;
    bottom: -20px;
  }

  .tick {
    position: absolute;
    cursor: pointer;
    white-space: nowrap;
  }

  .tick.vertical {
    left: 0;
  }

  .tick.horizontal {
    bottom: 0;
    transform: translateX(-50%);
  }
</style>
