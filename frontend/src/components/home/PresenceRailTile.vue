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
  let observer: IntersectionObserver | null = null;
  onMounted(() => {
    if (typeof IntersectionObserver === 'undefined' || !root.value) return;
    observer = new IntersectionObserver(([entry]) => { visible.value = entry.isIntersecting; }, { rootMargin: '120px' });
    observer.observe(root.value);
  });
  onBeforeUnmount(() => observer?.disconnect());
</script>

<template>
  <section ref="root" class="nh-rail-tile" :class="{ 'has-preview': preview.hasFrame.value }" :aria-label="`${device.friendly_name} presence`">
    <header>
      <strong>{{ device.friendly_name }}</strong>
      <span class="nh-muted">{{ statusOf(device) }}</span>
    </header>
    <div v-if="preview.hasFrame.value" class="nh-rail-preview">
      <SpatialCanvas :workspace="preview.workspace.value" :tracks="preview.tracks.value" :occupancy="new Map()"
        :view-box="preview.viewBox.value" :scale="preview.scale.value" :orientation="0" :online="preview.online.value" :mode="SpatialViewIds.Live"
        selected-id="" :trails="[]" :cells="[]" :show-grid="true" :show-room-context="true" :show-target-labels="false"
        :label="`${device.friendly_name} preview`" />
    </div>
    <p v-if="preview.hasFrame.value" class="nh-rail-state">
      <strong :class="{ 'is-occupied': values.occupied }">{{ state }}</strong>
      <span v-if="values.targetCount !== null"> · {{ values.targetCount }} target{{ values.targetCount === 1 ? '' : 's' }}</span>
      <span v-if="values.lastZone"> · {{ values.lastZone }}</span>
      <span v-if="preview.targetDistanceMm.value !== null"> · {{ (preview.targetDistanceMm.value / 1000).toFixed(1) }} m</span>
    </p>
    <p v-else class="nh-rail-state nh-muted">Waiting for a position…</p>
    <div class="nh-rail-actions">
      <RouterLink :to="{ name: RouteName.DevicePage, params: { id: device.id }, query: { tab: 'presence' } }">Live</RouterLink>
      <RouterLink :to="{ name: RouteName.DevicePage, params: { id: device.id }, query: { tab: 'history' } }">History</RouterLink>
    </div>
  </section>
</template>

<style scoped>
  .nh-rail-tile {
    display: grid;
    grid-template-columns: minmax(0, 1fr);
    grid-template-areas:
      'header'
      'state'
      'actions';
    align-items: center;
    row-gap: 0.1rem;
    min-width: 0;
    padding: 0.45rem 0;
    border-radius: 0;
    background: transparent;
  }
  .nh-rail-tile + .nh-rail-tile {
    border-top: 1px solid var(--nh-border);
  }
  .nh-rail-tile.has-preview {
    grid-template-columns: 4.5rem minmax(0, 1fr);
    grid-template-areas:
      'preview header'
      'preview state'
      'preview actions';
    column-gap: 0.6rem;
  }
  header { grid-area: header; display: flex; justify-content: space-between; gap: 0.4rem; align-items: baseline; min-width: 0; }
  header strong { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
  header .nh-muted { flex: none; font-size: 0.75rem; text-transform: capitalize; }
  .nh-rail-preview { grid-area: preview; width: 4.5rem; height: 4.5rem; overflow: hidden; border-radius: var(--nh-radius-s); background: var(--nh-surface); }
  .nh-rail-preview :deep(svg) { display: block; width: 100%; height: 100%; aspect-ratio: auto; max-height: none; }
  .nh-rail-state { grid-area: state; margin: 0; font-size: 0.85rem; }
  .nh-rail-state .is-occupied { color: var(--nh-accent); }
  .nh-rail-actions { grid-area: actions; display: flex; gap: 0.75rem; font-size: 0.8rem; }
  @container (min-width: 34rem) {
    .nh-rail-tile,
    .nh-rail-tile.has-preview {
      grid-template-columns: 1fr;
      grid-template-areas: 'header' 'preview' 'state' 'actions';
      align-items: stretch;
      gap: 0.4rem;
      padding: 0;
      border-radius: 0;
      background: transparent;
    }
    .nh-rail-tile + .nh-rail-tile { border-top: 0; }
    .nh-rail-preview { width: 100%; height: 9rem; background: var(--nh-surface-2); }
    .nh-rail-preview :deep(svg) { width: 100%; height: 9rem; }
  }
</style>
