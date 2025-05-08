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

  //   const trackFill = computed(() => {
  //   const colorStart = props.disabled ? '#ccc' : props.color;
  //   const colorEnd = props.disabled ? '#ccc' : '#fff2cc';
  //   const direction = props.direction === 'vertical' ? 'to top' : 'to right';

  //   return {
  //     background: `linear-gradient(${direction}, ${colorStart}, ${colorEnd})`,
  //     opacity: props.disabled ? 0.5 : 1,
  //   };
  // });
  const trackFill = computed(() => {
    const percent = ((value.value - props.min) / (props.max - props.min)) * 100;
    const color = props.disabled ? '#ccc' : props.color;
    const bgColor = props.disabled ? '#ccc' : '#fff2cc';
    const direction = props.direction === 'vertical' ? 'top' : 'right';

    // //     background: `linear-gradient(${direction}, ${colorStart}, ${colorEnd})`,

    return {
      background: `linear-gradient(to ${direction}, ${color} ${percent}%, ${bgColor} ${percent}%)`,
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

  .horizontal .slider::-webkit-slider-thumb {
    margin-top: -5px;
    transform: none;
  }
  .horizontal .slider::-moz-range-thumb {
    transform: none;
  }

  .vertical .slider::-webkit-slider-thumb {
    margin-top: 0;
    height: 60px;
  }

  .vertical .slider::-moz-range-thumb {
    margin-top: 0;
    height: 60px;
  }

  /* Webkit thumb styling */
  .slider::-webkit-slider-thumb {
    appearance: none;
    width: 6px;
    height: 30px;
    background: white;
    border: 1px solid #ccc;
    border-radius: 4px;
    cursor: pointer;
    transform: translateX(-2px);
  }

  /* Firefox track & thumb */
  .slider::-moz-range-track {
    height: 100%;
    background: transparent;
    border: none;
  }

  .slider::-moz-range-thumb {
    appearance: none;
    width: 6px;
    height: 30px;
    background: white;
    border: 1px solid #ccc;
    border-radius: 4px;
    cursor: pointer;
    transform: translateX(-2px);
  }

  .slider:disabled {
    opacity: 0.6;
  }
</style>
