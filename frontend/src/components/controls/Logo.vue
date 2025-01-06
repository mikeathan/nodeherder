<script setup lang="ts">
  import { onBeforeMount, ref, h, resolveComponent, computed } from 'vue';
  import Status from '@/components/controls/Status.vue';
</script>
<style scoped>
 

  .logo-container {
    position: relative;
    display: inline-flex;
    align-items: center;
    white-space: nowrap;
  }

  .logo-text {
    font-family: 'Roboto';
    color: var(--bs-body-color);
    font-size: 24px;
    font-weight: 700;
    margin-left: 10px;
    white-space: nowrap;
  }

  .status-icon {
    position: absolute;
    top: 15px;
    right: -10px;
  }
</style>

<template>
  <div class="logo-container">
    <RouterLink :to="`/`" style="text-decoration: none">
      <img src="../../assets/images/nodeherder_logo.png" width="50" height="50" alt="Node-Herder" class="logo-image" />
      <Status class="status-icon" />
      <span class="logo-text">Node-Herder</span>
    </RouterLink>
  </div>
</template>
