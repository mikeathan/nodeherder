import { computed } from 'vue';
import type { Device } from '@/types/device';
import { hasCompleteSpatialMapping, spatialTargetStatuses, type SpatialMapping } from '@/domain/spatial';
import type { SpatialWorkspace } from '@/types/spatial';
import { spatialSummary } from '@/domain/spatial/summary';
import { useHub } from '../useHub';
import { useSpatialTargets } from './useSpatialTargets';
import { useSpatialTracking } from './useSpatialTracking';
import { useSpatialOccupancy } from './useSpatialOccupancy';

/**
 * Live presence for one device: sampled fallback, tracking, zone occupancy and the compact
 * status line. Callers supply the mapping and workspace they are displaying, so the spatial
 * tab can pass unsaved drafts while the device summary passes saved settings; both share one
 * pipeline implementation instead of duplicating it.
 */
export function useSpatialPresence(
  device: () => Device,
  mappings: () => readonly SpatialMapping[],
  workspace: () => SpatialWorkspace
) {
  const { isConnected, statusOf } = useHub();
  const online = computed(() => isConnected.value && statusOf(device()) === 'online');
  const configured = computed(() => hasCompleteSpatialMapping(mappings()));

  const { targets: sampledTargets, seen, now } = useSpatialTargets(device, mappings, () => online.value);
  const tracking = useSpatialTracking(
    device,
    mappings,
    () => online.value,
    () => sampledTargets.value
  );
  /** Tracks still observed; held tracks are cleared slots and must not count as presence. */
  const activeTracks = computed(() => tracking.tracks.value.filter((track) => track.state !== 'held'));
  /** Slots a synchronized frame currently draws; they keep mapping statuses truthful. */
  const liveSlots = computed(() => new Set(activeTracks.value.map((track) => track.slot)));
  const statuses = computed(() =>
    spatialTargetStatuses(device().exposes, mappings(), seen.value, now.value, online.value, liveSlots.value)
  );
  const { occupancy, occupiedZoneNames } = useSpatialOccupancy(
    () => workspace().zones,
    () => activeTracks.value,
    () => tracking.revision.value,
    () => online.value
  );
  const summary = computed(() =>
    spatialSummary({
      online: online.value,
      mapped: configured.value,
      tracks: activeTracks.value,
      occupiedZones: occupiedZoneNames.value,
    })
  );

  return {
    online,
    configured,
    sampledTargets,
    statuses,
    activeTracks,
    occupancy,
    occupiedZoneNames,
    summary,
    ...tracking,
  };
}
