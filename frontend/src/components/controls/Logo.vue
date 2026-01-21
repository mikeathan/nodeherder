<script setup lang="ts">
  import Status from '@/components/controls/Status.vue';
</script>
<style scoped>
  .logo-link {
    display: inline-flex;
    align-items: center;
    text-decoration: none;
    gap: 0.5rem;
  }

  .logo-text {
    font-family: 'Roboto', sans-serif;
    color: var(--text-color);
    font-size: 24px;
    font-weight: 700;
    white-space: nowrap;
  }

  .status-wrapper {
    display: flex;
    align-items: center;
    /* Adjust position to be vertically centered or slightly offset */
    padding-top: 4px;
  }
</style>

<template>
  <div class="logo-container">
    <RouterLink :to="`/`" class="logo-link">
      <img src="../../assets/images/nodeherder_logo.png" width="40" height="40" alt="Node-Herder" class="logo-image" />
      <span class="logo-text">Node-Herder</span>
      <div class="status-wrapper">
        <Status />
      </div>
    </RouterLink>
  </div>
</template>
