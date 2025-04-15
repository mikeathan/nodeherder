<script setup lang="ts">
import { ref, computed } from 'vue'

const brightness = ref(50)

const trackFill = computed(() => {
  const percent = brightness.value
  return {
    background: `linear-gradient(to right, #ffc107 ${percent}%, #fff4cc ${percent}%)`
  }
})
</script>

<template>
  <div class="slider-wrapper">
    <div class="track-background" :style="trackFill"></div>
    <input type="range" min="0" max="100" v-model="brightness" class="slider" />
  </div>
</template>

<style scoped>
.slider-wrapper {
  position: relative;
  width: 280px;
  height: 42px;
  border-radius: 21px;
  overflow: hidden;
}

/* background fill (independent layer) */
.track-background {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  border-radius: 21px;
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
  height: 25px;

  background: rgb(92, 105, 218);
  border: none;
  border-radius: 3px;
  cursor: pointer;
}
</style>
