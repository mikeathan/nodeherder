<script setup lang="ts">
  defineProps<{ reading: string }>();
</script>

<template>
  <div class="entity-meter" role="group" aria-label="Energy meter reading">
    <!-- A cumulative counter has no natural maximum. This arc frames the reading;
         it deliberately encodes no percentage, target, or invented safety bands. -->
    <svg viewBox="0 0 280 156" aria-hidden="true">
      <path class="meter-track" d="M 26 130 A 114 114 0 0 1 254 130" />
      <path class="meter-accent" d="M 26 130 A 114 114 0 0 1 254 130" />
    </svg>
    <div class="meter-reading">
      <output>{{ reading }}</output>
      <span>Meter total</span>
    </div>
  </div>
</template>

<style scoped>
  .entity-meter {
    position: relative;
    width: min(100%, 18rem);
    margin: 1rem auto 0.5rem;
    color: var(--nh-kind-energy);
  }
  svg {
    display: block;
    width: 100%;
    overflow: visible;
  }
  path {
    fill: none;
    stroke-width: 10;
    stroke-linecap: round;
  }
  .meter-track {
    stroke: var(--nh-border);
  }
  .meter-accent {
    stroke: currentColor;
    opacity: 0.7;
  }
  .meter-reading {
    position: absolute;
    inset: auto 0 0.75rem;
    display: grid;
    justify-items: center;
    gap: 0.4rem;
  }
  output {
    color: var(--nh-text);
    font-size: clamp(1.8rem, 6vw, 2.4rem);
    font-weight: 600;
    letter-spacing: -0.04em;
    font-variant-numeric: tabular-nums;
  }
  span {
    color: var(--nh-text-muted);
    font-size: 0.8rem;
  }
</style>
