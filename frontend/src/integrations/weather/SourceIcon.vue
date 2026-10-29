<script setup lang="ts">
  import { weatherIcon } from '@/components/ui/icons';
  defineProps<{ condition?: string }>();
</script>

<template><span class="weather-icon" aria-hidden="true">{{ weatherIcon(condition) }}</span></template>

<style scoped>
  .weather-icon {
    display: inline-block;
    font-size: 1.6rem;
    line-height: 1;
  }
</style>
