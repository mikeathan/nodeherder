<script setup lang="ts">
  /*
   * A tiny signal-strength sparkline for one identity.
   *
   * The points come from the observations domain; this component only renders them. A single
   * reading draws nothing, because one point is not a trend.
   */
  import { computed } from 'vue';
  import { sparklinePoints } from '@/domain/observations/presentation';

  const props = withDefaults(
    defineProps<{
      signal?: number[];
      width?: number;
      height?: number;
    }>(),
    { signal: undefined, width: 64, height: 18 },
  );

  const points = computed(() => sparklinePoints(props.signal, props.width, props.height));
</script>

<template>
  <svg
    v-if="points"
    class="signal-sparkline"
    :viewBox="`0 0 ${props.width} ${props.height}`"
    :width="props.width"
    :height="props.height"
    aria-hidden="true"
  >
    <polyline
      :points="points"
      fill="none"
      stroke="currentColor"
      stroke-width="1.5"
      stroke-linecap="round"
      stroke-linejoin="round"
    />
  </svg>
</template>

<style scoped>
  .signal-sparkline {
    flex: none;
    color: var(--nh-accent);
    opacity: 0.85;
  }
</style>
