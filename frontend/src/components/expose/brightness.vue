<script setup lang="ts">
import { ref, computed } from 'vue';
import { prop } from 'vue-class-component';
const props = defineProps({
  value: {
    type: Number,
    default: 50,
  },
});
const emit = defineEmits<{
  (e: 'update', value: number): void;
}>();

const value = ref(props.value);

const trackFill = computed(() => {
  const percent = value.value;
  return {
    background: `linear-gradient(to right, #ffc107 ${percent}%, #fff4cc ${percent}%)`,
  };
});

function updateValue(event: any): void {
  const newValue = parseInt(event.target.value);
  emit('update', newValue);
}
</script>

<template>
  <div class="slider-wrapper">
    <div class="track-background" :style="trackFill"></div>
    <input type="range" min="0" max="100" v-model="value" class="slider" @change="updateValue" />
  </div>
</template>

<style scoped>
.slider-wrapper {
  position: relative;
  width: 280px;
  height: 50px;
  overflow: hidden;
}

.track-background {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  border-radius: 10px / 10px;
  pointer-events: none;
}

/* input layer */
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

/* Webkit thumb */
/* .slider::-webkit-slider-thumb {
  appearance: none;
  width: 32px;
  height: 6px;
  background: white;
  border-radius: 3px;
  margin-top: calc((42px - 6px) / 2);
  cursor: pointer;
  position: relative;
  z-index: 2;
  box-shadow: 0 0 1px rgba(0, 0, 0, 0.2);
} */

/* Firefox track & thumb */
.slider::-moz-range-track {
  height: 100%;
  background: transparent;
  border: none;
}

.slider::-moz-range-thumb {
  width: 6px;
  height: 30px;
  background: white;
  border: none;
  border-radius: 3px;
  cursor: pointer;
  transform: translateX(-10px);
  /* Shift left by 2 pixels */
}
</style>
