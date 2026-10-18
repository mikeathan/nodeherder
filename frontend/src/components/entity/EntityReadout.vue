<script setup lang="ts">
  import { computed } from 'vue';
  import type { Expose } from '@/types/device';
  import { exposeKind } from '@/domain/exposes';
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
  const stateful = computed(() => props.expose.type !== 'numeric');
  const color = computed(() => `var(--nh-kind-${exposeKind(props.expose.name)})`);
</script>

<template>
  <div class="entity-readout" :style="{ '--readout-color': color }">
    <span v-if="stateful" class="readout-badge" aria-hidden="true">
      <UiIcon :path="exposeIcon(expose.name)" />
    </span>
    <div class="readout-text">
      <div class="readout-value" :class="{ 'is-unknown': unknown }">{{ reading }}</div>
      <p>Device seen <RelativeTime :value="seen" /></p>
    </div>
  </div>
</template>

<style scoped>
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
