<script setup lang="ts">
  import { computed, onBeforeUnmount, onMounted, ref } from 'vue';
  import type { Device } from '@/types/device';
  import { RouteName } from '@/types/router';
  import { statusOf } from '@/composables/useHub';
  import { useDerivedPresence } from '@/composables/spatial/useDerivedPresence';
  import { useSpatialPreview } from '@/composables/spatial/useSpatialPreview';
  import { SpatialViewIds } from '@/domain/spatial/view';
  import SpatialCanvas from '@/components/device/spatial/SpatialCanvas.vue';

  const props = defineProps<{ device: Device }>();
  const derived = useDerivedPresence(() => props.device);
  const root = ref<HTMLElement | null>(null);
  const visible = ref(true);
  const preview = useSpatialPreview(() => props.device.id, () => visible.value && !document.hidden);
  const values = computed(() => derived.values.value);
  const state = computed(() => (values.value.occupied === null ? 'Unknown' : values.value.occupied ? 'Occupied' : 'Clear'));
  const distance = computed(() =>
    preview.targetDistanceMm.value === null ? '' : `${(preview.targetDistanceMm.value / 1000).toFixed(1)} m from sensor`
  );
  let observer: IntersectionObserver | null = null;
  onMounted(() => {
    if (typeof IntersectionObserver === 'undefined' || !root.value) return;
    observer = new IntersectionObserver(([entry]) => { visible.value = entry.isIntersecting; }, { rootMargin: '120px' });
    observer.observe(root.value);
  });
  onBeforeUnmount(() => observer?.disconnect());
</script>

<template>
  <section ref="root" class="nh-sensor-panel" :aria-label="`${device.friendly_name} presence`">
    <div class="nh-sensor-preview">
      <SpatialCanvas v-if="preview.hasFrame.value" :workspace="preview.workspace.value" :tracks="preview.tracks.value"
        :occupancy="new Map()" :view-box="preview.viewBox.value" :scale="preview.scale.value" :orientation="0"
        :online="preview.online.value" :mode="SpatialViewIds.Live" selected-id="" :trails="[]" :cells="[]"
        :show-grid="true" :show-room-context="true" :show-target-labels="false" :target-scale="2.5"
        :label="`${device.friendly_name} preview`" />
      <p v-else class="nh-sensor-waiting">Waiting for a position…</p>
    </div>
    <div class="nh-sensor-details">
      <header>
        <strong>{{ device.friendly_name }}</strong>
        <span class="nh-muted">{{ statusOf(device) }}</span>
      </header>
      <p class="nh-sensor-state">
        <strong :class="{ 'is-occupied': values.occupied }">{{ state }}</strong>
        <span v-if="values.targetCount !== null"> · {{ values.targetCount }} target{{ values.targetCount === 1 ? '' : 's' }}</span>
        <span v-if="values.lastZone"> · {{ values.lastZone }}</span>
      </p>
      <p class="nh-muted nh-sensor-reference">
        <span v-if="distance">{{ distance }}</span>
        <span v-else>No position yet</span>
      </p>
      <div class="nh-sensor-actions">
        <RouterLink :to="{ name: RouteName.DevicePage, params: { id: device.id }, query: { tab: 'presence' } }">Open live</RouterLink>
        <RouterLink :to="{ name: RouteName.DevicePage, params: { id: device.id }, query: { tab: 'history' } }">History</RouterLink>
      </div>
    </div>
  </section>
</template>

<style scoped>
  .nh-sensor-panel {
    display: grid;
    grid-template-columns: minmax(0, 1fr);
    gap: 0.55rem;
    min-width: 0;
    padding: 0.35rem 0;
  }
  .nh-sensor-panel + .nh-sensor-panel {
    border-top: 1px solid var(--nh-border);
    margin-top: 0.5rem;
    padding-top: 0.85rem;
  }
  .nh-sensor-preview {
    width: 100%;
    aspect-ratio: 5 / 3;
    height: auto;
    min-width: 0;
    border-radius: var(--nh-radius);
    overflow: hidden;
    background: var(--nh-surface-2);
  }
  .nh-sensor-preview :deep(svg) {
    display: block;
    width: 100%;
    height: 100%;
    aspect-ratio: auto;
    max-height: none;
  }
  .nh-sensor-waiting {
    display: grid;
    place-items: center;
    height: 100%;
    margin: 0;
    padding: 0.75rem;
    text-align: center;
    color: var(--nh-text-muted);
    font-size: 0.85rem;
  }
  .nh-sensor-details { display: grid; align-content: center; gap: 0.3rem; min-width: 0; }
  header { display: flex; align-items: baseline; gap: 0.5rem; flex-wrap: wrap; }
  header strong { font-size: 1.05rem; }
  .nh-sensor-state { margin: 0; }
  .nh-sensor-state .is-occupied { color: var(--nh-accent); }
  .nh-sensor-reference { margin: 0; }
  .nh-sensor-actions { display: flex; gap: 0.75rem; margin-top: 0.2rem; }
  @media (max-width: 640px) {
    .nh-sensor-preview { aspect-ratio: 5 / 3; height: auto; }
  }
</style>
