<script setup lang="ts">
import { ref } from 'vue';
import Slider from 'primevue/slider';

const brightness = ref(50); // Initialize with a default brightness value
</script>

<style scoped>
.spotlight-brightness .my-slider.p-slider {

  background-color: #ffc107 !important;
}

/* The colored progress range (left side of slider) */
.spotlight-brightness .my-slider.p-slider .p-slider-range {
  background-color: #ffc107 !important;
}

/* Handle/thumb of the slider */
.spotlight-brightness .my-slider.p-slider .p-slider-handle {
  background: white;
  border-color: #ffc107;
}

:root {
  --primary-color: #ffc107;
  --primary-color-text: #212529;
}

.my-slider.p-slider {
  /* Target the main slider container */
  height: 42px !important;
  background-color: #ffc107 !important;
  /* Set the overall background */
  background: #ffc107 !important;
  border: none !important;
  border-radius: 8px !important;
  position: relative !important;
  overflow: hidden !important;
}

.my-slider .p-slider-range {
  /* Target the selected range (left side) */
  background-color: #ffc107 !important;
  background: #ffc107 !important;
}
</style>
<template>
  <div class="spotlight-brightness">
    <!-- <input type="number" v-model.number="brightness" min="0" max="100" class="spotlight-number-input" />
    <span class="spotlight-unit">%</span> -->
    <Slider v-model="brightness" :min="0" :max="100" class="my-slider" />
  </div>
</template>
