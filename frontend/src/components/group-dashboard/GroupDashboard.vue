<script setup lang="ts">
  import { computed } from 'vue';
  import SensorView from '../expose/SensorView.vue';

</script>
<style scoped>
  /* .dashboard {
  background-color: var(--p-card-background);
  height: 100vh;
  overflow: auto;
} */
</style>

<template>
  <div class="grid">
    <div class="col-12 md:col-6 lg:col-3 xg:col-2">
      <SensorView :value="14"/>
    </div>
  </div>
</template>
