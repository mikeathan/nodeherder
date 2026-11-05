/*
 * Zone occupancy with hysteresis. A raw point-in-zone test flickers when a target sits on a
 * boundary, so occupancy requires two consecutive inside samples to turn on and stays on for
 * a short grace period when the last target leaves. Detection previews remain overlays: they
 * are never presented as firmware occupancy.
 */
import type { SpatialPoint, SpatialZone } from '@/types/spatial';
import { targetInsideZone } from './workspace';

/** Consecutive inside samples required before a zone reports occupied. */
export const SPATIAL_ZONE_ENTER_SAMPLES = 2;
/** How long a zone stays occupied after its last inside sample, in milliseconds. */
export const SPATIAL_ZONE_EXIT_MS = 800;

/** Occupancy of one zone overlay. */
export type ZoneOccupancy = { occupied: boolean };

/** One position considered for zone occupancy; tracks supply filtered positions. */
export type ZoneOccupancyPoint = SpatialPoint;

type ZoneState = { occupied: boolean; insideHits: number; outsideSince: number };

/**
 * Creates a per-zone hysteresis tracker for one view. Call `update` with the current filtered
 * track positions whenever they change; zones removed from the workspace are forgotten.
 */
export function createZoneOccupancyTracker(): {
  update: (zones: readonly SpatialZone[], points: readonly ZoneOccupancyPoint[], at: number) => Map<string, ZoneOccupancy>;
  reset: () => void;
} {
  const states = new Map<string, ZoneState>();
  return {
    update(zones, points, at) {
      const result = new Map<string, ZoneOccupancy>();
      const known = new Set<string>();
      for (const zone of zones) {
        known.add(zone.id);
        const state = states.get(zone.id) ?? { occupied: false, insideHits: 0, outsideSince: 0 };
        const inside = points.some((point) => targetInsideZone(zone, zones, point));
        if (inside) {
          state.outsideSince = 0;
          if (!state.occupied) {
            state.insideHits += 1;
            if (state.insideHits >= SPATIAL_ZONE_ENTER_SAMPLES) state.occupied = true;
          }
        } else {
          state.insideHits = 0;
          if (state.occupied) {
            if (!state.outsideSince) {
              state.outsideSince = at;
            } else if (at - state.outsideSince >= SPATIAL_ZONE_EXIT_MS) {
              state.occupied = false;
              state.outsideSince = 0;
            }
          }
        }
        states.set(zone.id, state);
        result.set(zone.id, { occupied: state.occupied });
      }
      for (const id of states.keys()) {
        if (!known.has(id)) states.delete(id);
      }
      return result;
    },
    reset() {
      states.clear();
    },
  };
}

/**
 * Preview text for one zone overlay, stating the source instead of implying firmware truth.
 * `hasData` is false until any position arrives, so an empty view is never presented as a
 * measured clear room.
 */
export function zonePreviewText(occupancy: ZoneOccupancy | undefined, online: boolean, hasData: boolean): string {
  if (!online) return 'Unknown — device offline';
  if (!hasData || !occupancy) return 'Unknown — waiting for position data';
  return occupancy.occupied ? 'Target inside' : 'No target inside';
}
