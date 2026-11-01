/*
 * Spatial-tab view registry and small presentation helpers. Keeping the view
 * ids and labels here avoids string literals spread across panel components.
 */
import type { Device } from '@/types/device';
import type { SpatialTarget } from './index';
import type { SpatialZone } from '@/types/spatial';
import { formatExposeValue } from '@/domain/exposes';
import { zonePreview } from './workspace';

/** Context help topics embedded in older spatial components. */
export const SpatialHelpTopics = {
  Zones: 'zones',
  Room: 'room',
  Entry: 'entry',
  Heatmap: 'heatmap',
} as const;
export type SpatialHelpTopic = (typeof SpatialHelpTopics)[keyof typeof SpatialHelpTopics];

/** Summary labels for each context help topic. */
export const SPATIAL_CONTEXT_HELP_LABELS: Readonly<Record<SpatialHelpTopic, string>> = {
  zones: 'How to create zones',
  room: 'Help with Room Builder',
  entry: 'Help with Entry / Exit & Assumed Presence',
  heatmap: 'Help with Heatmap & Analytics',
};

/** Accessible descriptions for the small help illustrations. */
export const SPATIAL_HELP_ILLUSTRATIONS: Readonly<Record<SpatialHelpTopic, string>> = {
  zones: 'Diagram of detection, exclusion and entry zone shapes around a sensor',
  room: 'Diagram of a room outline with sensor, field of view and furniture',
  entry: 'Diagram of a doorway entry zone with an assumed-presence hold',
  heatmap: 'Diagram of a heat grid from cool blue to hot red activity cells',
};

/** Stable spatial tab view ids. */
export const SpatialViewIds = {
  Live: 'live',
  Zones: 'zones',
  Room: 'room',
  Entry: 'entry',
  Heatmap: 'heatmap',
} as const;

/** Spatial tab views in display order. */
export const SPATIAL_VIEWS = [
  { id: SpatialViewIds.Live, label: 'Live dashboard' },
  { id: SpatialViewIds.Zones, label: 'Zone editor' },
  { id: SpatialViewIds.Room, label: 'Room builder' },
  { id: SpatialViewIds.Entry, label: 'Entry / exit' },
  { id: SpatialViewIds.Heatmap, label: 'Heatmap & analytics' },
] as const;

export type SpatialViewId = (typeof SPATIAL_VIEWS)[number]['id'];

/** Default tab view. */
export const DEFAULT_SPATIAL_VIEW: SpatialViewId = SpatialViewIds.Live;
/** Views that expose pointer editing on the canvas. */
export const SPATIAL_EDIT_VIEWS: readonly SpatialViewId[] = [SpatialViewIds.Zones, SpatialViewIds.Room];
/** Views that show live target/zone context. */
export const SPATIAL_LIVE_VIEWS: readonly SpatialViewId[] = [SpatialViewIds.Live, SpatialViewIds.Zones];

/** Prompt shown before discarding session-only observations. */
export const SPATIAL_CLEAR_PROMPT = 'Discard this tab’s collected positions? Saved room layouts and zones remain.';
/** Observation-window option label. */
export function observationWindowLabel(minutes: number): string {
  return `Last ${minutes} minutes`;
}

/** Session start in the browser's local time. */
export function formatSessionStart(at: number): string {
  return new Date(at).toLocaleTimeString();
}

/** Count label for sampled seconds. */
export function sampleCountLabel(count: number): string {
  return `${count} target-second sample${count === 1 ? '' : 's'}`;
}

/** Count label for observed one-second frames. */
export function frameCountLabel(count: number): string {
  return `${count} one-second frame${count === 1 ? '' : 's'} with fresh targets in this window`;
}

/** Connection label for the spatial status strip. */
export function connectionLabel(online: boolean): string {
  return online ? 'Connected' : 'Offline';
}

/** Whether the view supports canvas pointer editing. */
export function isSpatialEditView(view: SpatialViewId): boolean {
  return SPATIAL_EDIT_VIEWS.includes(view);
}

/** Whether the view shows live target and zone previews. */
export function isSpatialLiveView(view: SpatialViewId): boolean {
  return SPATIAL_LIVE_VIEWS.includes(view);
}

/** Status text for the number of fresh targets. */
export function freshTargetLabel(targets: readonly SpatialTarget[]): string {
  return `${targets.length} fresh ${targets.length === 1 ? 'target' : 'targets'}`;
}

/** Status text for the number of saved zone overlays. */
export function zoneOverlayLabel(zoneCount: number): string {
  return `${zoneCount} zone overlay${zoneCount === 1 ? '' : 's'}`;
}

/** One zone preview row, including an optional confirmed device occupancy reading. */
export type ZonePreviewRow = { zone: SpatialZone; preview: string; reported: string | null; stale: boolean };

/** Builds all per-zone preview text for the panel in one pass. */
export function zonePreviewRows(
  device: Device,
  zones: readonly SpatialZone[],
  targets: readonly SpatialTarget[],
  online: boolean
): ZonePreviewRow[] {
  return zones.map((zone) => {
    const expose = device.exposes[zone.occupancyEntity ?? ''];
    return {
      zone,
      preview: zonePreview(zone, zones, targets, online),
      reported: expose ? formatExposeValue(expose, expose.data) : null,
      stale: !online,
    };
  });
}
