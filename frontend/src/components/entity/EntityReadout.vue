<script setup lang="ts">
  import { computed } from 'vue';
  import type { Expose } from '@/types/device';
  import { exposeKind, exposeValueColor } from '@/domain/exposes';
  import { exposeIcon } from '@/components/ui/icons';
  import UiIcon from '@/components/ui/UiIcon.vue';
  import RelativeTime from '@/components/ui/RelativeTime.vue';
  const props = defineProps<{
    expose: Expose;
    reading: string;
    unknown: boolean;
    seen: string | number | null | undefined;
  }>();
  // Activity/alert/event views have a state badge; environmental readings retain
  // the quieter typography-first layout from the accepted environmental design.
  const stateful = computed(
    () => props.expose.type !== 'numeric' || ['battery', 'signal'].includes(exposeKind(props.expose.name))
  );
  const color = computed(() => exposeValueColor(props.expose));
  const percentageBattery = computed(() => exposeKind(props.expose.name) === 'battery' && props.expose.unit === '%');
  const battery = computed(() => {
    if (!percentageBattery.value) return null;
    const value = props.expose.data;
    if (typeof value !== 'number') return null;
    // Comparisons reject NaN/infinity as well as invalid percentages. Missing data
    // must not become an empty (zero-percent) battery indicator.
    return value >= 0 && value <= 100 ? value : null;
  });
</script>

<template>
  <div class="entity-readout-block">
    <div class="entity-readout" :style="{ '--readout-color': color }">
      <span v-if="stateful" class="readout-badge" aria-hidden="true">
        <UiIcon :path="exposeIcon(expose.name)" />
      </span>
      <div class="readout-text">
        <div class="readout-value" :class="{ 'is-unknown': unknown }">{{ reading }}</div>
        <p>Device seen <RelativeTime :value="seen" /></p>
      </div>
    </div>
    <div
      v-if="battery !== null"
      class="battery-level"
      role="meter"
      aria-label="Battery level"
      aria-valuemin="0"
      aria-valuemax="100"
      :aria-valuenow="battery"
      :aria-valuetext="`${battery} %`">
      <span :style="{ width: `${battery}%` }" />
    </div>
  </div>
</template>

<style scoped>
  .battery-level {
    height: 0.65rem;
    margin: 0.25rem 0 1rem;
    border-radius: 99px;
    background: var(--nh-surface-2);
    overflow: hidden;
  }
  .battery-level > span {
    display: block;
    height: 100%;
    border-radius: inherit;
    background: var(--nh-kind-battery);
  }
  .entity-readout {
    display: flex;
    align-items: center;
    gap: 1rem;
    padding: 1rem 0;
  }
  .readout-text {
    min-width: 0;
  }
  .readout-value {
    font-size: clamp(2.2rem, 6vw, 3.5rem);
    font-weight: 600;
    line-height: 1.2;
    letter-spacing: -0.04em;
    overflow-wrap: anywhere;
  }
  .readout-value.is-unknown {
    font-size: 1.9rem;
    font-weight: 500;
    color: var(--nh-text-muted);
  }
  p {
    margin: 0.5rem 0 0;
    color: var(--nh-text-muted);
    font-size: 0.8rem;
  }
  .readout-badge {
    display: grid;
    place-items: center;
    flex: 0 0 3.5rem;
    height: 3.5rem;
    border-radius: 50%;
    color: var(--readout-color);
    background: color-mix(in srgb, var(--readout-color) 14%, transparent);
    border: 1px solid color-mix(in srgb, var(--readout-color) 35%, transparent);
    font-size: 1.8rem;
  }
</style>
