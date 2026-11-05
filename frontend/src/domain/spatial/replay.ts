/*
 * Session-only replay of accepted raw positions. Recording is bounded in memory and time;
 * nothing is persisted and replay never feeds tracking, history or automations.
 */
import type { SpatialDisplayTrack } from '@/domain/spatial/tracks';

/** How far back the session replay reaches. */
export const REPLAY_WINDOW_MS = 5 * 60 * 1000;
/** Hard frame cap, so a long session cannot grow memory without bound. */
export const REPLAY_MAX_FRAMES = 1500;
/** Playback tick in milliseconds. */
export const REPLAY_TICK_MS = 200;

export type ReplaySample = { slot: number; x: number; y: number };
export type ReplayFrame = { at: number; samples: ReplaySample[] };

/** Appends one frame and enforces the time window and frame cap. */
export function appendReplayFrame(
  frames: readonly ReplayFrame[],
  frame: ReplayFrame,
  windowMs = REPLAY_WINDOW_MS,
  maxFrames = REPLAY_MAX_FRAMES
): ReplayFrame[] {
  const cutoff = frame.at - windowMs;
  return [...frames.filter((entry) => entry.at > cutoff), frame].slice(-maxFrames);
}

/** First frame that actually contains a target position, skipping empty/cleared frames. */
export function firstRecordedFrame(frames: readonly ReplayFrame[]): ReplayFrame | undefined {
  return frames.find((frame) => frame.samples.length > 0);
}

/** Newest frame that contains a target position. */
export function lastRecordedFrame(frames: readonly ReplayFrame[]): ReplayFrame | undefined {
  for (let index = frames.length - 1; index >= 0; index--) {
    if (frames[index].samples.length > 0) return frames[index];
  }
  return undefined;
}

/** Whether the buffer holds any recorded target position at all. */
export function hasRecordedPositions(frames: readonly ReplayFrame[]): boolean {
  return frames.some((frame) => frame.samples.length > 0);
}

/** The newest frame at or before `at`, else the first frame; undefined when empty. */
export function frameAt(frames: readonly ReplayFrame[], at: number): ReplayFrame | undefined {
  if (!frames.length) return undefined;
  let found = frames[0];
  for (const frame of frames) {
    if (frame.at > at) break;
    found = frame;
  }
  return found;
}

/** Synthetic display tracks for one recorded frame; no filtering or association is implied. */
export function replayTracks(frame: ReplayFrame | undefined): SpatialDisplayTrack[] {
  if (!frame) return [];
  return frame.samples.map((sample) => ({
    id: sample.slot,
    label: `Slot ${sample.slot}`,
    slot: sample.slot,
    x: sample.x,
    y: sample.y,
    rawX: sample.x,
    rawY: sample.y,
    vx: 0,
    vy: 0,
    at: frame.at,
    colorIndex: (Math.max(1, sample.slot) - 1) % 4,
    state: 'live' as const,
  }));
}
