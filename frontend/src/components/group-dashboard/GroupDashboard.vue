<script setup lang="ts">
  import { computed } from 'vue';
  import Brightness from '../expose/brightness.vue';
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
      <Brightness />
    </div>
  </div>
</template>
