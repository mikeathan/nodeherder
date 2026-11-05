import { computed, onBeforeUnmount, ref, watch } from 'vue';
import { store } from '@/store';
import type { Device, DeviceUpdate } from '@/types/device';
import { SPATIAL_FRESHNESS_TICK_MS, spatialTargets, type SpatialMapping } from '@/domain/spatial';

/**
 * Tracks arrival freshness for the mapped coordinate entities of one device and exposes the
 * sampled fallback targets. Frames are the preferred source; this path keeps the view usable
 * with a server that does not send frames and supplies per-row mapping diagnostics.
 */
export function useSpatialTargets(device: () => Device, mappings: () => readonly SpatialMapping[], online: () => boolean) {
  const seen = ref(new Map<string, number>());
  const now = ref(Date.now());
  const coordinates = computed(() => new Set(mappings().flatMap((mapping) => [mapping.x, mapping.y])));

  function observe(update: DeviceUpdate) {
    if (!update || update.id !== device().id || !update.data || typeof update.data !== 'object') return;
    const current = Date.now();
    for (const key of Object.keys(update.data)) {
      if (coordinates.value.has(key)) seen.value.set(key, current);
    }
    now.value = current;
  }
  const unsubscribe = store.subscribe((mutation) => {
    if (mutation.type === 'hub/updateDevice') observe(mutation.payload as DeviceUpdate);
  });
  const timer = setInterval(() => { now.value = Date.now(); }, SPATIAL_FRESHNESS_TICK_MS);
  watch(() => device().id, () => { seen.value = new Map(); });
  watch(online, (value) => { if (!value) seen.value = new Map(); });
  onBeforeUnmount(() => { unsubscribe(); clearInterval(timer); });

  const targets = computed(() => spatialTargets(device().exposes, mappings(), seen.value, now.value, online()));
  return { targets, seen, now };
}
