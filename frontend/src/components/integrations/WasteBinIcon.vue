<script setup lang="ts">
  import { computed } from 'vue';
  import { wasteBinColours } from '@/domain/waste-bin';

  // Compact, decorative wheelie-bin symbol; the adjacent title/metadata also identify the service.
  const props = defineProps<{ colour: string | null }>();
  const palette = computed(() => wasteBinColours(props.colour));
</script>

<template>
  <svg class="waste-bin-icon" viewBox="0 0 40 52" width="48" height="60" aria-hidden="true" focusable="false">
    <path d="M14 8V5h12v3" fill="none" :stroke="palette.lid" stroke-width="2.5" stroke-linejoin="round" />
    <path class="waste-bin-body" d="M8 13h24l-3 31a2 2 0 0 1-2 2H13a2 2 0 0 1-2-2Z" :fill="palette.body" />
    <path d="m27 16-2 25" fill="none" :stroke="palette.lid" stroke-width="2" stroke-linecap="round" />
    <path d="m13 17 2 24" fill="none" stroke="white" stroke-opacity="0.25" stroke-width="2" stroke-linecap="round" />
    <rect x="5" y="9" width="30" height="5" rx="2" :fill="palette.lid" />
    <circle cx="13" cy="47" r="3" fill="#71808b" />
    <circle cx="27" cy="47" r="3" fill="#71808b" />
  </svg>
</template>

<style scoped>
  .waste-bin-icon {
    display: block;
    flex: none;
  }
</style>
