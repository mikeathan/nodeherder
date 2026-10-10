<script setup lang="ts">
  /* Arc gauge used by panel mode (temperature, health ratios). */
  import { computed } from 'vue';

  const props = defineProps<{ value: number; min: number; max: number; label: string; caption: string; kind?: 'climate' | 'ok' | 'signal' | 'lamp' }>();
  const R = 70;
  const C = 2 * Math.PI * R;
  const ARC = C * 0.75;
  const fill = computed(() => {
    const span = props.max - props.min || 1;
    const f = (Math.max(props.min, Math.min(props.max, props.value)) - props.min) / span;
    return ARC * (Number.isFinite(f) ? f : 0);
  });
</script>

<template>
  <figure class="pd" :class="`kind-${kind ?? 'climate'}`">
    <svg viewBox="0 0 180 180" role="img" :aria-label="`${label}, ${caption}`">
      <circle cx="90" cy="90" :r="R" class="pd-track" :stroke-dasharray="`${ARC} ${C}`" transform="rotate(135 90 90)" />
      <circle cx="90" cy="90" :r="R" class="pd-arc" :stroke-dasharray="`${fill} ${C}`" transform="rotate(135 90 90)" />
    </svg>
    <figcaption>
      <b>{{ label }}</b>
      <span>{{ caption }}</span>
    </figcaption>
  </figure>
</template>

<style scoped>
  .pd {
    position: relative;
    margin: 0;
    display: grid;
    place-items: center;
  }
  .pd svg {
    width: 100%;
    max-width: 17rem;
  }
  .pd circle {
    fill: none;
    stroke-width: 12;
    stroke-linecap: round;
  }
  .pd-track {
    stroke: var(--nh-surface-2);
  }
  .pd-arc {
    stroke: var(--pd-c);
    filter: drop-shadow(0 0 8px color-mix(in srgb, var(--pd-c) 50%, transparent));
    transition: stroke-dasharray 400ms;
  }
  .kind-climate {
    --pd-c: var(--dk-climate);
  }
  .kind-ok {
    --pd-c: var(--nh-ok);
  }
  .kind-signal {
    --pd-c: var(--nh-accent);
  }
  .kind-lamp {
    --pd-c: var(--dk-lamp);
  }
  figcaption {
    position: absolute;
    inset: 0;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    text-align: center;
    padding: 0 2rem;
  }
  figcaption b {
    font-size: 2.4rem;
    font-weight: 300;
    font-variant-numeric: tabular-nums;
    line-height: 1.1;
  }
  figcaption span {
    color: var(--nh-text-muted);
    font-size: 0.9rem;
  }
</style>
