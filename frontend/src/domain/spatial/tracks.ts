/*
 * Track association for mapped radar slots. Firmware reports fixed slots (1–3), not people,
 * so a slot can be reused by a different person between frames. This module keeps a stable
 * display identity per physical target by associating each filtered slot position with the
 * nearest previous track, and suppresses a slot that reports the same target twice. Both are
 * display heuristics: they never claim a person's identity and never feed automations.
 */

/** Association gate in millimetres at zero speed. */
export const SPATIAL_ASSOC_GATE_MM = 700;
/** Extra gate per 1 m/s of track speed, so a fast walker is not treated as a new track. */
export const SPATIAL_ASSOC_SPEED_MM_S = 1500;
/** Two tracks closer than this for long enough are treated as one visible target. */
export const SPATIAL_MERGE_MM = 350;
/** Consecutive frames required before a duplicate is merged. */
export const SPATIAL_MERGE_SAMPLES = 3;
/** How long a disappeared track keeps its identity for re-acquisition. */
export const SPATIAL_TRACK_RETAIN_MS = 8000;
/** Number of theme colours used to distinguish tracks. */
export const SPATIAL_TRACK_COLORS = 4;

/** How long a cleared slot keeps its last position visible, marked as held. */
export const SPATIAL_HOLD_MS = 2000;
/** Silence after which a still-reported track is shown as aged instead of current. */
export const SPATIAL_SILENCE_MS = 3000;

/** Presentation state of a track. */
export type SpatialTrackState = 'live' | 'predicted' | 'held' | 'idle';

/** One filtered slot position offered to the associator. */
export type SpatialTrackInput = {
  /** Mapping row number that produced the reading. */
  slot: number;
  x: number;
  y: number;
  vx: number;
  vy: number;
  /** Sample time in epoch milliseconds. */
  at: number;
  /** Whether the filter accepted the sample or held a position after rejecting a jump. */
  state: 'live' | 'predicted';
};

/** A track with a stable identity for display. */
export type SpatialTrack = {
  id: number;
  label: string;
  slot: number;
  x: number;
  y: number;
  vx: number;
  vy: number;
  at: number;
  colorIndex: number;
  state: SpatialTrackState;
};

/** A track with its unfiltered source position, so a raw view needs no second pipeline. */
export type SpatialDisplayTrack = SpatialTrack & { rawX: number; rawY: number };

/** Tracks to draw plus how many duplicates were suppressed this frame. */
export type SpatialTrackResult = { tracks: SpatialTrack[]; duplicates: number };

type InternalTrack = Omit<SpatialTrack, 'label'>;

/** Track labels never imply a person; the number is a display identity only. */
export function spatialTrackLabel(id: number): string {
  return `Track ${id}`;
}

function distance(track: { x: number; y: number }, input: { x: number; y: number }): number {
  return Math.hypot(track.x - input.x, track.y - input.y);
}

function mergeKey(left: number, right: number): string {
  return left < right ? `${left}:${right}` : `${right}:${left}`;
}

/**
 * Creates an associator for one device view. Call `update` with the current filtered slot
 * positions on every new frame; unmatched tracks are retained internally for
 * `SPATIAL_TRACK_RETAIN_MS` so a brief loss does not change a track's label.
 */
export function createTrackAssociator(): {
  update: (inputs: readonly SpatialTrackInput[], at: number) => SpatialTrackResult;
  reset: () => void;
} {
  let tracks: InternalTrack[] = [];
  let nextId = 1;
  const mergeCounts = new Map<string, number>();

  function reset(): void {
    tracks = [];
    nextId = 1;
    mergeCounts.clear();
  }

  function withinGate(track: InternalTrack, input: SpatialTrackInput, at: number): boolean {
    const elapsed = Math.max(0, (at - track.at) / 1000);
    const gate = SPATIAL_ASSOC_GATE_MM + Math.hypot(track.vx, track.vy) * elapsed;
    return distance(track, input) <= gate;
  }

  return {
    update(inputs, at) {
      // Drop identities that outlived their retention before matching, so an old track can
      // never steal a new target that appears nearby.
      tracks = tracks.filter((track) => at - track.at <= SPATIAL_TRACK_RETAIN_MS);

      const available = [...tracks];
      const matched: { track: InternalTrack; input: SpatialTrackInput }[] = [];
      const created: InternalTrack[] = [];

      for (const input of inputs) {
        // Slot stickiness first: the same slot is the strongest signal available.
        let index = available.findIndex((track) => track.slot === input.slot && withinGate(track, input, at));
        if (index < 0) {
          let best = Number.POSITIVE_INFINITY;
          available.forEach((track, candidate) => {
            if (!withinGate(track, input, at)) return;
            const candidateDistance = distance(track, input);
            if (candidateDistance < best) {
              best = candidateDistance;
              index = candidate;
            }
          });
        }
        if (index < 0) {
          const id = nextId++;
          created.push({
            id,
            slot: input.slot,
            x: input.x,
            y: input.y,
            vx: input.vx,
            vy: input.vy,
            at: input.at,
            colorIndex: (id - 1) % SPATIAL_TRACK_COLORS,
            state: input.state,
          });
          continue;
        }
        const [track] = available.splice(index, 1);
        matched.push({ track, input });
      }

      const updated: InternalTrack[] = [];
      for (const { track, input } of matched) {
        track.slot = input.slot;
        track.x = input.x;
        track.y = input.y;
        track.vx = input.vx;
        track.vy = input.vy;
        track.at = input.at;
        track.state = input.state;
        updated.push(track);
      }
      for (const track of created) updated.push(track);
      tracks = [...updated, ...available];
      updated.sort((left, right) => left.id - right.id);

      // Duplicate detection counts consecutive frames of closeness per track pair, so a
      // single crossed reading does not hide a second person.
      const seenKeys = new Set<string>();
      const suppressed = new Set<number>();
      for (let i = 0; i < updated.length; i++) {
        for (let j = i + 1; j < updated.length; j++) {
          if (distance(updated[i], updated[j]) > SPATIAL_MERGE_MM) continue;
          const key = mergeKey(updated[i].id, updated[j].id);
          seenKeys.add(key);
          const count = (mergeCounts.get(key) ?? 0) + 1;
          mergeCounts.set(key, count);
          if (count >= SPATIAL_MERGE_SAMPLES) suppressed.add(Math.max(updated[i].id, updated[j].id));
        }
      }
      for (const key of mergeCounts.keys()) {
        if (!seenKeys.has(key)) mergeCounts.delete(key);
      }

      const visible = updated.filter((track) => !suppressed.has(track.id)).map(toSpatialTrack);
      return { tracks: visible, duplicates: suppressed.size };
    },
    reset,
  };
}

function toSpatialTrack(track: InternalTrack): SpatialTrack {
  return {
    id: track.id,
    label: spatialTrackLabel(track.id),
    slot: track.slot,
    x: track.x,
    y: track.y,
    vx: track.vx,
    vy: track.vy,
    at: track.at,
    colorIndex: track.colorIndex,
    state: track.state,
  };
}
