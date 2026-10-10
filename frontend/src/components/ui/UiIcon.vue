<script setup lang="ts">
  /* Inline SVG icon. Decorative unless `label` is given (then it is announced). */
  import { computed } from 'vue';
  import { ICONS, IconName } from './icons';

  const props = defineProps<{ name?: IconName; path?: string; label?: string }>();
  const d = computed(() => props.path ?? (props.name ? ICONS[props.name] : ''));
</script>

<template>
  <svg
    class="nh-ic"
    viewBox="0 0 24 24"
    :role="label ? 'img' : undefined"
    :aria-label="label"
    :aria-hidden="label ? undefined : 'true'"
    focusable="false">
    <path :d="d" fill="currentColor" />
  </svg>
</template>

<style>
  .nh-ic {
    width: 1.25em;
    height: 1.25em;
    flex: none;
  }
</style>
