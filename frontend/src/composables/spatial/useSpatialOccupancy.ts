import { computed, shallowRef, watch } from 'vue';
import type { SpatialZone } from '@/types/spatial';
import { createZoneOccupancyTracker, type ZoneOccupancy } from '@/domain/spatial/occupancy';

/**
 * Tracks zone occupancy with hysteresis for one view. Positions come from filtered tracks, so
 * a boundary hover cannot flicker; the map is cleared while the device is offline instead of
 * reporting an empty room as measured.
 */
export function useSpatialOccupancy(
  zones: () => readonly SpatialZone[],
  points: () => readonly { x: number; y: number }[],
  revision: () => number,
  online: () => boolean
) {
  const occupancy = shallowRef<Map<string, ZoneOccupancy>>(new Map());
  const tracker = createZoneOccupancyTracker();

  // Zone hysteresis runs once per accepted tracking update, not on every wall-clock tick:
  // "two consecutive samples" must mean two position updates, not two timer ticks.
  watch(
    [zones, revision, online],
    () => {
      if (!online()) {
        tracker.reset();
        occupancy.value = new Map();
        return;
      }
      occupancy.value = tracker.update(zones(), points(), Date.now());
    },
    { immediate: true, deep: true }
  );

  /** Names of occupied zones in workspace order, for compact status text. */
  const occupiedZoneNames = computed(() =>
    zones()
      .filter((zone) => occupancy.value.get(zone.id)?.occupied)
      .map((zone) => zone.name)
  );
  return { occupancy, occupiedZoneNames };
}
