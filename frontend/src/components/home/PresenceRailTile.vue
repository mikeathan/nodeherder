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
  <section ref="root" class="nh-rail-tile" :aria-label="`${device.friendly_name} presence`">
    <header>
      <strong>{{ device.friendly_name }}</strong>
      <span class="nh-muted">{{ statusOf(device) }}</span>
    </header>
    <div v-if="preview.hasFrame.value" class="nh-rail-preview">
      <SpatialCanvas :workspace="preview.workspace.value" :tracks="preview.tracks.value" :occupancy="new Map()"
        :scale="4000" :orientation="0" :online="preview.online.value" :mode="SpatialViewIds.Live"
        selected-id="" :trails="[]" :cells="[]" :show-grid="false" :show-room-context="false"
        :label="`${device.friendly_name} preview`" />
    </div>
    <p v-if="preview.hasFrame.value" class="nh-rail-state">
      <strong :class="{ 'is-occupied': values.occupied }">{{ state }}</strong>
      <span v-if="values.targetCount !== null"> · {{ values.targetCount }} target{{ values.targetCount === 1 ? '' : 's' }}</span>
      <span v-if="values.lastZone"> · {{ values.lastZone }}</span>
    </p>
    <p v-else class="nh-rail-state nh-muted">Waiting for a position…</p>
    <div class="nh-rail-actions">
      <RouterLink :to="{ name: RouteName.DevicePage, params: { id: device.id }, query: { tab: 'presence' } }">Live</RouterLink>
      <RouterLink :to="{ name: RouteName.DevicePage, params: { id: device.id }, query: { tab: 'history' } }">History</RouterLink>
    </div>
  </section>
</template>

<style scoped>
  .nh-rail-tile { display: grid; gap: 0.35rem; padding: 0.6rem; border: 1px solid var(--nh-border); border-radius: var(--nh-radius); background: var(--nh-surface); min-width: 0; }
  header { display: flex; justify-content: space-between; gap: 0.4rem; align-items: baseline; }
  .nh-rail-preview :deep(svg) { width: 100%; height: auto; max-height: 8rem; }
  .nh-rail-state { margin: 0; font-size: 0.85rem; }
  .nh-rail-state .is-occupied { color: var(--nh-accent); }
  .nh-rail-actions { display: flex; gap: 0.75rem; font-size: 0.85rem; }
</style>
