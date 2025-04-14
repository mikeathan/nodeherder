<script setup lang="ts">
  import { ref } from 'vue';
  import Slider from 'primevue/slider';

  const brightness = ref(50); // Initialize with a default brightness value
</script>

<style scoped>
  /* .spotlight-brightness {
    display: flex;
    align-items: center;
  }

  .spotlight-slider {
    flex-grow: 1;
    margin-right: 10px;
  }

  .spotlight-number-input {
    width: 40px;
    margin-left: 8px;
    text-align: right;
  }

  .spotlight-unit {
    margin-left: 4px;
  } */

  .my-slider .p-slider-range {
    background-color: #42b983; /* your left color */
  }

  /* Right side (track after handle) */
  .my-slider.p-slider {
    height: 42px !important;
    background-color: #ffc107 !important;
    background: #ffc107 !important;
    color: #ffc107 !important;
    border: none;
    border-radius: 8px;
    position: relative;
    overflow: hidden;
  }

  .my-slider.p-slider::before {
    content: '';
    position: absolute;
    top: 0;
    bottom: 0;
    left: var(--range-end, 50%); /* adjust this dynamically if needed */
    right: 0;
    background-color: #ffc107 !important;
    z-index: 1;
  }
  /* .p-slider {
    height: 42px !important;
    background-color: #ffc107 !important;
    background: #ffc107 !important;
    color: #ffc107 !important;
    border: none;
    border-radius: 8px;
  } */
  /* .spotlight-slider .p-slider-range {
    background: #ffc107 !important;
  }
  .spotlight-slider .p-slider {
    height: 42px !important;
    background-color: #ffc107 !important;
    background: #ffc107 !important;
    border: none;
    border-radius: 8px;
  } */
  /* .p-slider-range {
    background-color: #ffc107 !important;
    background: #ffc107 !important;
    color: #ffc107 !important;
   
  } */

  /* .p-slider-handle {
    width: 14px;
    height: 14px;
    background-color: #fff;
    border: 1px solid #ffc107;
    border-radius: 50%;
    cursor: pointer;
    top: 50%;
    transform: translateY(-50%);
    box-shadow: none; 
  }

  .p-slider:focus {
    outline: none;
    box-shadow: 0 0 0 0.2rem rgba(255, 193, 7, 0.25); 
  } */
</style>
<template>
  <div class="spotlight-brightness">
    <input type="number" v-model.number="brightness" min="0" max="100" class="spotlight-number-input" />
    <span class="spotlight-unit">%</span>
    <Slider v-model="brightness" :min="0" :max="100" class="my-slider" />
  </div>
</template>
