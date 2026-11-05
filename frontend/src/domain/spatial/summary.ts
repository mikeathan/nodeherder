/*
 * One-line spatial status for compact surfaces. Precedence is explicit so a widget can never
 * claim a live position while the device is offline, unmapped or silent.
 */
import type { SpatialTrackState } from './tracks';

/** Summary tone; callers map it to a shared status style. */
export type SpatialSummaryTone = 'offline' | 'unmapped' | 'idle' | 'clear' | 'tracked';

/** Compact status line built from live view state. */
export type SpatialSummary = { tone: SpatialSummaryTone; label: string };

/** View state a summary needs; tracks carry only the presentation state. */
export type SpatialSummaryInput = {
  online: boolean;
  mapped: boolean;
  tracks: readonly { state: SpatialTrackState }[];
  /** Names of zones currently occupied by a track, in workspace order. */
  occupiedZones?: readonly string[];
};

/** Chip tone for the compact headline; never the only carrier of meaning. */
export type SpatialHeadlineTone = 'danger' | 'muted' | 'ok' | 'accent';

/** One-line presence headline used outside the presence workspace. */
export type SpatialHeadline = { tone: SpatialHeadlineTone; label: string };

/** Input for the headline; slots are mapped rows, not tracks or people. */
export type SpatialHeadlineInput = { online: boolean; configured: boolean; activeSlots: number };

/**
 * Builds the device-page presence headline from a frame snapshot only. It deliberately says
 * "target", not "track": association, hold and hysteresis live in the presence workspace, so
 * a chip must not imply more certainty than it has.
 */
export function spatialHeadline(input: SpatialHeadlineInput): SpatialHeadline {
  if (!input.online) return { tone: 'danger', label: 'Offline' };
  if (!input.configured) return { tone: 'muted', label: 'Not set up' };
  if (input.activeSlots <= 0) return { tone: 'ok', label: 'Clear' };
  return { tone: 'accent', label: `${input.activeSlots} ${input.activeSlots === 1 ? 'target' : 'targets'}` };
}

/** Builds the compact status line. Track counts are display counts, not people. */
export function spatialSummary(input: SpatialSummaryInput): SpatialSummary {
  if (!input.online) return { tone: 'offline', label: 'Offline' };
  if (!input.mapped) return { tone: 'unmapped', label: 'Target mapping not saved' };
  const tracks = input.tracks;
  if (!tracks.length) return { tone: 'clear', label: 'Clear' };
  if (tracks.every((track) => track.state === 'idle')) return { tone: 'idle', label: 'No new position' };
  const count = `${tracks.length} ${tracks.length === 1 ? 'track' : 'tracks'}`;
  const zone = input.occupiedZones?.[0];
  return { tone: 'tracked', label: zone ? `${count} · ${zone}` : count };
}
