<!-- PickSlider.vue -->
<script setup lang="ts">
import { ref, watch, computed } from 'vue';
import { ControlDirection } from '@/types/controls.type';
import BaseSlider from './BaseSlider.vue';

const props = withDefaults(
  defineProps<{
    value?: number;
    unit?: string;
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

const showMarker = ref(false);
let markerTimeout: ReturnType<typeof setTimeout> | null = null;
// Add a ref to track the percent for the marker
const currentTrackPercent = ref(0);

function showMarkerWithTimeout() {
  showMarker.value = true;

  if (markerTimeout) {
    clearTimeout(markerTimeout);
  }

  markerTimeout = setTimeout(() => {
    showMarker.value = false;
  }, 500);
}

function handleUpdate(value: number) {
  emit('update', value);
  showMarkerWithTimeout();
}

// Calculate the track percent locally
function calculateTrackPercent(val: number) {
  const range = props.max - props.min;
  if (range === 0) return 0;
  const clamped = Math.min(Math.max(val, props.min), props.max);
  return ((clamped - props.min) / range) * 100;
}

// Initialize with proper value
currentTrackPercent.value = calculateTrackPercent(props.value);
</script>

<template>
  <div class="slider-container" :class="props.direction">
    <BaseSlider
      :value="value"
      :min="min"
      :max="max"
      :disabled="disabled"
      :color="color"
      :direction="direction"
      @click="showMarkerWithTimeout"
      @update="handleUpdate">
      
      <!-- Custom tap indicator for Pick type -->
      <template #indicator="{ trackPercentRaw, direction, color }">
        <div
          class="tap-indicator"
          :class="[direction, 'tap']"
          :style="{
            '--indicator-percent': `${trackPercentRaw}%`,
            '--indicator-color': color,
          }">
        </div>
      </template>
      
      <!-- Add the input event handler -->
      <template #extra="{ trackPercentRaw, value: sliderValue }">
        <input @input.stop="showMarkerWithTimeout" />
        <div v-show="false">{{ currentTrackPercent = trackPercentRaw }}</div>
      </template>
    </BaseSlider>

    <!-- Value marker -->
    <div
      v-show="showMarker"
      :class="['value-marker', props.direction]"
      :style="props.direction === 'vertical' ? 
        { bottom: `${currentTrackPercent}%` } : 
        { left: `${currentTrackPercent}%` }">
      {{ value }} {{ props.unit }}
    </div>
  </div>
</template>

<style scoped>
.slider-container {
  position: relative;
  
}

.slider-container.vertical {
  height: 320px;
  width: auto;
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

/* === Value Marker === */
.value-marker {
  position: absolute;
  color: black;
  font-size: 14px;
  font-weight: bold;
  pointer-events: none;
  z-index: 1000;
  white-space: nowrap;
  display: flex;
  align-items: center;
  justify-content: center;
  opacity: 1;
  transition: opacity 0.4s ease;
}

/* Horizontal Value Marker */
.value-marker.horizontal {
  top: 100%;
  transform: translateX(-50%);
  margin-top: 6px;
  background: white;
  box-shadow: 1px 4px 6px #333;
  border-radius: 999px;
  padding: 4px 12px;
}

/* Vertical Value Marker  */
.value-marker.vertical {
  left: -70%;
  transform: translateY(50%);
  height: 40px;
  width: 70%;
  background: white;
  box-shadow: 1px 4px 6px #333;
  border-radius: 999px;
  padding: 4px 12px;
}
</style>