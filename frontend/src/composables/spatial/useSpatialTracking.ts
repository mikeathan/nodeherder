import { computed, onBeforeUnmount, ref, shallowRef, watch } from 'vue';
import { store } from '@/store';
import type { SpatialFrame } from '@/types/spatial';
import { hasCompleteSpatialMapping, type SpatialTarget } from '@/domain/spatial';
import { frameSamples } from '@/domain/spatial/frames';
import { createSpatialFilter, type SpatialFilterOutput } from '@/domain/spatial/filter';
import {
  createTrackAssociator,
  SPATIAL_HOLD_MS,
  SPATIAL_SILENCE_MS,
  type SpatialDisplayTrack,
  type SpatialTrackInput,
  type SpatialTrackState,
} from '@/domain/spatial/tracks';
import { appendReplayFrame, type ReplayFrame } from '@/domain/spatial/replay';
import { SPATIAL_DEFAULT_SMOOTHING, type SpatialSmoothing } from '@/domain/spatial/smoothing';

/** Recent raw samples replayed when the smoothing level changes, so the effect is immediate. */
const HISTORY_WINDOW_MS = 3000;
const HISTORY_MAX_SAMPLES = 40;

/** Which source produced the current positions. */
export type SpatialTrackingSource = 'frames' | 'sampled' | 'none';

/** Immutable per-update tracking state; display age is derived from `now` by the caller. */
export type SpatialTrackingSnapshot = {
  /** Tracks observed in the current update, before age and hold presentation. */
  liveTracks: SpatialDisplayTrack[];
  /** Filtered slot positions for trails and session sampling. */
  slots: SpatialTarget[];
  source: SpatialTrackingSource;
  /** Whether any position data (frame or sampled coordinate) has arrived since the last reset. */
  received: boolean;
  /** Monotonic counter of accepted tracking updates; zone hysteresis runs once per value. */
  revision: number;
  /** Highest per-slot jitter estimate, in millimetres. */
  jitterMm: number;
  duplicates: number;
};

const EMPTY_SNAPSHOT: SpatialTrackingSnapshot = {
  liveTracks: [],
  slots: [],
  source: 'none',
  received: false,
  revision: 0,
  jitterMm: 0,
  duplicates: 0,
};

type SlotPoint = { x: number; y: number; rawX: number; rawY: number };

/**
 * Turns mapped slot readings into stable, filtered tracks for one device view. Frames are
 * preferred when they are at least as fresh as sampled coordinates; the sampled path keeps the
 * view working against a server without frames. All filter and association state is owned
 * here and reset on device, mapping, source and connection changes.
 */
export function useSpatialTracking(
  device: () => { id: string },
  mappings: () => readonly { x: string; y: string; active: string }[],
  online: () => boolean,
  sampledTargets: () => readonly SpatialTarget[]
) {
  const smoothing = ref<SpatialSmoothing>(SPATIAL_DEFAULT_SMOOTHING);
  const now = ref(Date.now());
  const snapshot = shallowRef<SpatialTrackingSnapshot>(EMPTY_SNAPSHOT);

  /** Filtered position and raw source position per slot, for tracks and the raw view. */
  const slots = new Map<number, SlotPoint>();
  const filters = new Map<number, ReturnType<typeof createSpatialFilter>>();
  const clearedAt = new Map<number, { track: SpatialDisplayTrack; at: number }>();
  const associator = createTrackAssociator();
  // Per-slot recent raw samples. A level change replays them, because a still target produces
  // no new frames and the control would otherwise appear to do nothing.
  const history = new Map<number, { x: number; y: number; at: number }[]>();
  // Session-only replay recording; bounded by domain helpers and never persisted.
  const replayFrames = ref<ReplayFrame[]>([]);
  let lastSource: SpatialTrackingSource | null = null;
  // Arrival time in the browser's clock. Frame timestamps come from the hub's clock; using
  // them for age would misjudge staleness on any machine with clock skew.
  let lastFrameRef: SpatialFrame | undefined;
  let lastFrameSeenAt = 0;

  const frame = computed<SpatialFrame | undefined>(() => store.getters['hub/spatialFrame'](device().id));
  /**
   * Sampled values change only when a coordinate arrives. This signature ignores the
   * service's one-second freshness tick, so the pipeline does not recompute while idle.
   */
  const sampledSignature = computed(() => sampledTargets().map((target) => `${target.id}:${target.x}:${target.y}`).join('|'));

  /** Drops filters, identities and hold state so a new source or connection starts clean. */
  function reset(): void {
    slots.clear();
    filters.clear();
    history.clear();
    replayFrames.value = [];
    clearedAt.clear();
    associator.reset();
    lastSource = null;
    snapshot.value = EMPTY_SNAPSHOT;
  }

  function filterFor(slot: number) {
    const existing = filters.get(slot);
    if (existing) return existing;
    const created = createSpatialFilter();
    filters.set(slot, created);
    return created;
  }

  /** Records the last known track per slot so a cleared slot can be shown as held briefly. */
  function rememberCleared(slot: number, at: number): void {
    const previous = snapshot.value.liveTracks.find((track) => track.slot === slot);
    if (previous) clearedAt.set(slot, { track: { ...previous }, at });
    else clearedAt.delete(slot);
    filters.delete(slot);
    slots.delete(slot);
    history.delete(slot);
  }

  function ingest(): void {
    if (!online() || !hasCompleteSpatialMapping(mappings())) {
      reset();
      return;
    }

    const observedAt = Date.now();
    const currentFrame = frame.value;
    if (currentFrame !== lastFrameRef) {
      lastFrameRef = currentFrame;
      lastFrameSeenAt = observedAt;
    }
    // Frames are a synchronized, complete view and the server sends them only while a mapping
    // exists. Once one arrives, a sampled coordinate must not flip the whole view (and wipe its
    // identities); sampled data is the fallback for a server without frames.
    const useFrames = lastFrameSeenAt > 0;
    // Frames and sampled coordinates arrive with different timing and values; carrying an
    // identity or filter state across a source switch would show a phantom jump.
    const nextSource: SpatialTrackingSource = useFrames ? 'frames' : 'sampled';
    if (lastSource && lastSource !== nextSource) reset();
    lastSource = nextSource;
    const samples: { slot: number; x: number; y: number; at: number }[] = [];

    if (useFrames && currentFrame) {
      const reported = new Set<number>();
      for (const sample of frameSamples(currentFrame)) {
        reported.add(sample.slot);
        if (!sample.active || sample.x === null || sample.y === null) {
          rememberCleared(sample.slot, observedAt);
          continue;
        }
        clearedAt.delete(sample.slot);
        samples.push({ slot: sample.slot, x: sample.x, y: sample.y, at: observedAt });
      }
      // A slot missing from the frame no longer reports; treat it as cleared without inventing
      // a position, and let the hold window decide how long it stays visible.
      for (const slot of [...filters.keys()]) {
        if (!reported.has(slot)) rememberCleared(slot, observedAt);
      }
    } else {
      const at = observedAt;
      const reported = new Set<number>();
      for (const target of sampledTargets()) {
        reported.add(target.id);
        clearedAt.delete(target.id);
        samples.push({ slot: target.id, x: target.x, y: target.y, at });
      }
      for (const slot of [...filters.keys()]) {
        if (!reported.has(slot)) rememberCleared(slot, at);
      }
    }

    if (!samples.length) {
      snapshot.value = {
        liveTracks: [],
        slots: [],
        source: nextSource,
        received: snapshot.value.received || !!currentFrame || sampledTargets().length > 0,
        revision: snapshot.value.revision + 1,
        jitterMm: 0,
        duplicates: 0,
      };
      return;
    }

    if (useFrames && currentFrame) {
      replayFrames.value = appendReplayFrame(replayFrames.value, {
        at: observedAt,
        samples: samples.map((sample) => ({ slot: sample.slot, x: sample.x, y: sample.y })),
      });
    }
    rememberHistory(samples);
    applySamples(samples, observedAt, nextSource);
  }

  /** Keeps a short raw window per slot, bounded by age and count. */
  function rememberHistory(samples: readonly { slot: number; x: number; y: number; at: number }[]): void {
    for (const sample of samples) {
      const points = history.get(sample.slot) ?? [];
      points.push({ x: sample.x, y: sample.y, at: sample.at });
      const cutoff = sample.at - HISTORY_WINDOW_MS;
      history.set(sample.slot, points.filter((point) => point.at >= cutoff).slice(-HISTORY_MAX_SAMPLES));
    }
  }

  /** One filtered slot position ready for association. */
  type FilteredPoint = {
    slot: number;
    x: number;
    y: number;
    vx: number;
    vy: number;
    at: number;
    state: 'live' | 'predicted';
    jitterMm: number;
    rawX: number;
    rawY: number;
  };

  /** Associates filtered points and publishes one tracking snapshot. */
  function publishPoints(points: FilteredPoint[], associationAt: number, source: SpatialTrackingSource): void {
    const inputs: SpatialTrackInput[] = [];
    let jitterMm = 0;
    for (const point of points) {
      slots.set(point.slot, { x: point.x, y: point.y, rawX: point.rawX, rawY: point.rawY });
      jitterMm = Math.max(jitterMm, point.jitterMm);
      inputs.push({ slot: point.slot, x: point.x, y: point.y, vx: point.vx, vy: point.vy, at: point.at, state: point.state });
    }
    const result = associator.update(inputs, associationAt);
    const liveTracks = result.tracks.map((track) => {
      const point = slots.get(track.slot);
      return { ...track, rawX: point?.rawX ?? track.x, rawY: point?.rawY ?? track.y };
    });
    snapshot.value = {
      liveTracks,
      slots: [...slots.entries()].map(([slot, point]) => ({
        id: slot,
        x: point.x,
        y: point.y,
        distance: Math.hypot(point.x, point.y),
      })),
      source,
      received: true,
      revision: snapshot.value.revision + 1,
      jitterMm,
      duplicates: result.duplicates,
    };
  }

  /** Filters one batch of fresh samples, then associates them. */
  function applySamples(
    samples: readonly { slot: number; x: number; y: number; at: number }[],
    associationAt: number,
    source: SpatialTrackingSource
  ): void {
    const points: FilteredPoint[] = [];
    for (const sample of samples) {
      const filtered = filterFor(sample.slot).apply({ x: sample.x, y: sample.y, at: sample.at }, smoothing.value);
      points.push({ ...filtered, slot: sample.slot, rawX: sample.x, rawY: sample.y, at: sample.at });
    }
    publishPoints(points, associationAt, source);
  }

  /**
   * Replays the recent raw window through a fresh filter when the level changes. The window is
   * shifted so its last sample lands now, preserving sample spacing and producing an
   * immediately different position instead of waiting for new motion.
   */
  function replaySmoothing(): void {
    if (!online() || !hasCompleteSpatialMapping(mappings()) || history.size === 0) {
      ingest();
      return;
    }
    const now = Date.now();
    const points: FilteredPoint[] = [];
    for (const [slot, samples] of history) {
      if (!samples.length) continue;
      const offset = now - samples[samples.length - 1].at;
      filters.delete(slot);
      let filtered: SpatialFilterOutput = { x: samples[0].x, y: samples[0].y, vx: 0, vy: 0, state: 'live', jitterMm: 0 };
      for (const sample of samples) {
        filtered = filterFor(slot).apply(
          { x: sample.x, y: sample.y, at: sample.at + offset },
          smoothing.value
        );
      }
      points.push({ ...filtered, slot, rawX: samples[samples.length - 1].x, rawY: samples[samples.length - 1].y, at: now });
    }
    if (!points.length) {
      ingest();
      return;
    }
    publishPoints(points, now, lastSource ?? 'sampled');
  }

  watch(
    [frame, () => device().id, () => online(), () => hasCompleteSpatialMapping(mappings()), sampledSignature],
    ingest,
    { immediate: true }
  );
  watch(smoothing, replaySmoothing);
  // A quarter-second tick keeps idle/hold transitions within the documented bounds. It stops
  // while the document is hidden, so a background tab does no work; environments without a real
  // document (unit tests) simply keep ticking.
  const watchVisibility =
    typeof document !== 'undefined' && typeof document.addEventListener === 'function' ? document : null;
  const tick = () => {
    if (watchVisibility?.hidden) return;
    now.value = Date.now();
  };
  const timer = setInterval(tick, 250);
  const onVisibility = () => {
    if (!watchVisibility?.hidden) tick();
  };
  watchVisibility?.addEventListener('visibilitychange', onVisibility);
  onBeforeUnmount(() => {
    clearInterval(timer);
    watchVisibility?.removeEventListener('visibilitychange', onVisibility);
  });

  const tracks = computed<SpatialDisplayTrack[]>(() => {
    const at = now.value;
    const live = snapshot.value.liveTracks.map((track) => ({
      ...track,
      state: track.state === 'predicted' ? track.state : at - track.at > SPATIAL_SILENCE_MS ? ('idle' as SpatialTrackState) : track.state,
    }));
    const liveSlots = new Set(live.map((track) => track.slot));
    const held = [...clearedAt.values()]
      .filter((entry) => !liveSlots.has(entry.track.slot) && at - entry.at < SPATIAL_HOLD_MS)
      .map((entry) => ({ ...entry.track, state: 'held' as SpatialTrackState }));
    return [...live, ...held];
  });

  return {
    tracks,
    slots: computed(() => snapshot.value.slots),
    source: computed(() => snapshot.value.source),
    hasData: computed(() => snapshot.value.received),
    revision: computed(() => snapshot.value.revision),
    jitterMm: computed(() => Math.round(snapshot.value.jitterMm)),
    duplicates: computed(() => snapshot.value.duplicates),
    replayFrames,
    smoothing,
  };
}
