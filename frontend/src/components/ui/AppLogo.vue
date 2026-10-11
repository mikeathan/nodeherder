<script setup lang="ts">
  /* NodeHerder mark with optional wordmark (spec 007 FR-13). */
  import mark from '@/assets/images/nodeherder-mark.webp';

  withDefaults(defineProps<{ size?: number; wordmark?: boolean; version?: string }>(), { size: 40, wordmark: true });
</script>

<template>
  <span class="nh-logo">
    <img :src="mark" alt="" :height="size" :width="Math.round(size * 0.68)" class="nh-logo-img" />
    <span v-if="wordmark" class="nh-logo-word">
      <span>Node<b>Herder</b></span>
      <small v-if="version">v{{ version }}</small>
    </span>
    <span v-else class="sr-only">NodeHerder</span>
  </span>
</template>

<style scoped>
  .nh-logo {
    display: inline-flex;
    align-items: center;
    gap: 0.6rem;
    min-width: 0;
  }
  .nh-logo-img {
    display: block;
    object-fit: contain;
    flex: none;
  }
  .nh-logo-word {
    display: flex;
    flex-direction: column;
    line-height: 1.1;
    font-family: var(--nh-font-display);
    font-size: 1.15rem;
    letter-spacing: -0.01em;
    white-space: nowrap;
  }
  .nh-logo-word b {
    color: var(--nh-accent);
  }
  .nh-logo-word small {
    font-size: 0.68rem;
    color: var(--nh-text-muted);
    font-weight: 500;
  }
</style>
