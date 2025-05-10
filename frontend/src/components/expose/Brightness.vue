<script setup lang="ts">
  import { ref, computed, watch } from 'vue';

  const props = defineProps({
    value: {
      type: Number,
      default: 50,
    },
    min: {
      type: Number,
      default: 0,
    },
    max: {
      type: Number,
      default: 100,
    },
    disabled: {
      type: Boolean,
      default: false,
    },
    color: {
      type: String,
      default: '#ffc107',
    },
    trackFilled: {
      type: Boolean,
      default: false,
    },
    direction: {
      type: String as () => 'horizontal' | 'vertical',
      default: 'horizontal',
    },
  });

  const emit = defineEmits<{
    (e: 'update', value: number): void;
  }>();

  const value = ref(props.value);
  const lastKnownValue = ref(props.value);


  const calculateTrackPercent = () => {
    if (!props.trackFilled) return '';

    const { min, max } = props;
    const clampedValue = Math.min(Math.max(value.value, min), max);
    const range = max - min;

    if (range === 0) return '0%';

    const percent = ((clampedValue - min) / range) * 100;
    return `${percent.toFixed(1)}%`;
  };

  const trackPercentRaw = computed(() => {
    const { min, max } = props;
    const clamped = Math.min(Math.max(value.value, min), max);
    return ((clamped - min) / (max - min)) * 100;
  });
  const trackFill = computed(() => {
    const color = props.disabled ? '#ccc' : props.color;
    const bgColor = props.disabled ? '#ccc' : '#fff2cc';

    const direction = props.direction === 'vertical' ? 'to top' : 'to right';
    const percent = calculateTrackPercent();

    return {
      background: `linear-gradient(${direction}, ${color} ${percent}, ${bgColor} ${percent})`,
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
  <div :class="generateDirectionalClass('slider-wrapper')">
    <div :class="generateDirectionalClass('track-background')" :style="trackFill"></div>
    <div
      class="tap-indicator"
      :class="[props.direction, props.trackFilled ? 'filled' : 'tap']"
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
</template>

<style scoped>
  .slider-wrapper {
    position: relative;
    height: 46px;
    overflow: visible;
  }

  .slider-wrapper.vertical {
    height: 320px;
    width: 130px;
  }

  .track-background {
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    pointer-events: none;
  }

  .track-background.vertical {
    border-radius: 36px;
  }
  .track-background.horizontal {
    border-radius: 12px;
  }
  /* Input layer */
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

  .vertical .slider {
    position: absolute;
    top: 50%;
    left: 50%;
    width: var(--wrapper-height, 320px);
    height: var(--wrapper-width, 130px);
    margin: 0;
    transform-origin: center center;
    /* Center precisely, then rotate */
    transform: translate(-50%, -50%) rotate(-90deg);
  }

  /* Hide the default thumb */
  .slider::-webkit-slider-thumb,
  .slider::-moz-range-thumb {
    visibility: hidden;
    width: 0;
    height: 0;
    border: none;
  }

  /* Webkit thumb styling */
  /*.slider::-webkit-slider-thumb {
    /* appearance: none;
    width: 6px;
    height: 30px;
    background: white;
    border: 1px solid #ccc;
    border-radius: 4px;
    cursor: pointer;
    transform: translateX(-2px);
  } */

  /* Firefox track & thumb */
  /* .slider::-moz-range-track {
    height: 100%;
    background: transparent;
    border: none;
  } */

  /* .slider::-moz-range-thumb {
    appearance: none;
    width: 5px;
    height: 30px;
    background: white;
    border: 1px solid black;
    border-radius: 20px;
    cursor: pointer;
    transform: translateX(-2px);
  } */
  /* .horizontal .slider::-webkit-slider-thumb {
    margin-top: -5px;
    transform: none;
  }
  .horizontal .slider::-moz-range-thumb {
    transform: none;
  }

  .vertical .slider::-webkit-slider-thumb {
    margin-top: 0;
    height: 0%;
  }

  .vertical .slider::-moz-range-thumb {
    margin-top: 0;
    height: 0%;
  } */

  .tap-indicator {
    position: absolute;
    pointer-events: none;
    z-index: 2;
    background-color: var(--indicator-color, #ccc);
  }

  /* === Horizontal === */
  .tap-indicator.horizontal.tap {
    top: 0;
    bottom: 0;
    width: 2px;
    left: calc(var(--indicator-percent));
    transform: translateX(-1px);
  }

  .tap-indicator.horizontal.filled {
    height: 30px;
    width: 5px;
    border-radius: 20px;
    border: 1px solid #ccc;
    background-color: white;

    left: calc(var(--indicator-percent));
    top: 50%;
    transform: translate(-50%, -50%);
  }

  /* === Vertical === */
  .tap-indicator.vertical.tap {
    width: 100%;
    height: 40px;
    background: white;
    border-radius: 8px;
    margin-top: 5px;
    box-shadow: 0 4px 6px -4px rgba(0, 0, 0, 0.2), /* bottom shadow */ inset 0 10px 15px -10px rgba(255, 165, 0, 0.6),
      /* inner top orange */ inset 0 -10px 15px -10px rgba(255, 165, 0, 0.6); /* inner bottom orange */

    top: calc(100% - var(--indicator-percent));
    left: 50%;
    transform: translate(-50%, -50%);
  }

  .tap-indicator.vertical.tap::before {
    content: '';
    position: absolute;
    top: 50%;
    left: 25%;
    width: 50%;
    height: 4px;
    background-color: #555; /* or black */
    border-radius: 2px;
    transform: translateY(-50%);
  }


  .tap-indicator.vertical.filled {
    width: 60px;
    height: 5px;
    border-radius: 20px;
    border: 1px solid #ccc;
    background-color: white;
    margin-top: 5px;
    top: calc(100% - var(--indicator-percent));
    left: 50%;
    transform: translate(-50%, -50%);
  }
  .slider:disabled {
    opacity: 0.6;
  }
</style>
