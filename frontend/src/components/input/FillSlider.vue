<script setup lang="ts">
  import { computed } from 'vue';
  import { ControlDirection } from '@/types/controls.type';
  import BaseSlider from './BaseSlider.vue';

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

  const calculateGradient = (trackPercent: number) => {
    const color = props.disabled ? '#ccc' : props.color;
    const bgColor = props.disabled ? '#ccc' : '#fff2cc';
    const direction = props.direction === 'vertical' ? 'to top' : 'to right';
    const percent = `${trackPercent.toFixed(1)}%`;
    return {
      background: `linear-gradient(${direction}, ${color} ${percent}, ${bgColor} ${percent})`,
      opacity: props.disabled ? 0.5 : 1,
    };
  };

  function handleUpdate(value: number) {
    emit('update', value);
  }
</script>

<template>
  <BaseSlider
    :value="value"
    :min="min"
    :max="max"
    :disabled="disabled"
    :color="color"
    :direction="direction"
    @update="handleUpdate">
    <!-- Override the track background with fill-specific gradient -->
    <template #track="{ trackPercentRaw }">
      <div class="fill-track" :style="calculateGradient(trackPercentRaw)"></div>
    </template>

    <!-- Custom indicator for Fill type -->
    <template #indicator="{ trackPercentRaw, direction }">
      <div
        class="tap-indicator"
        :class="[direction, 'filled']"
        :style="{
          '--indicator-percent': `${trackPercentRaw}%`,
        }"></div>
    </template>

    <!-- Custom thumb -->
    <template #thumb="{ trackPercentRaw, direction }">
      <div
        class="slider-thumb"
        :class="[direction]"
        :style="{
          '--indicator-percent': `${trackPercentRaw}%`,
        }"></div>
    </template>
  </BaseSlider>
</template>

<style scoped>
  /* Track styling */
  .fill-track {
    width: 100%;
    height: 100%;
    border-radius: 4px;
  }

  /* === Thumb styling === */
  .slider-thumb {
    position: absolute;
    z-index: 3;
    background-color: white;
    border: 1px solid #ddd;
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.2);
  }

  /* Horizontal thumb */
  .slider-thumb.horizontal {
    width: 16px;
    height: 16px;
    border-radius: 50%;
    left: var(--indicator-percent);
    top: 50%;
    transform: translate(-50%, -50%);
  }

  /* Vertical thumb */
  .slider-thumb.vertical {
    width: 16px;
    height: 16px;
    border-radius: 50%;
    left: 50%;
    top: calc(100% - var(--indicator-percent));
    transform: translate(-50%, 50%);
  }

  /* === Tap Indicator (Common Base) === */
  .tap-indicator {
    position: absolute;
    pointer-events: none;
    z-index: 2;
  }

  /* === Horizontal Tap Indicator Styles === */
  .tap-indicator.horizontal.filled {
    height: 30px;
    width: 5px;
    border-radius: 20px;
    margin-left: -5px;
    border: 1px solid #ccc;
    background-color: white;
    left: var(--indicator-percent);
    top: 50%;
    transform: translate(-50%, -50%);
    box-shadow: 0px 0px 2px rgba(0, 0, 0, 0.3);
  }

  /* === Vertical Tap Indicator Styles === */
  .tap-indicator.vertical.filled {
    width: 70px;
    height: 5px;
    border-radius: 30px;
    background-color: white;
    margin-top: 8px;
    top: calc(100% - var(--indicator-percent));
    left: 50%;
    transform: translate(-50%, -50%);
    box-shadow: 0px 0px 2px rgba(0, 0, 0, 0.3);
  }
</style>
