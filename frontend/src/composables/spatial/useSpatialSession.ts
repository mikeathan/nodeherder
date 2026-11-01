// Owns the spatial tab's bounded, session-only observation history. Nothing is
// persisted or sent to a device; timers and subscriptions are released on scope exit.
import { computed, onBeforeUnmount, ref, watch } from 'vue';
import type { SpatialTarget } from '@/domain/spatial';
import {
  DEFAULT_OBSERVATION_MINUTES,
  MINUTE_MS,
  SPATIAL_SAMPLE_INTERVAL_MS,
  SPATIAL_SESSION_MS,
  activityByHour,
  appendFrame,
  heatCells,
  targetTrails,
  type SpatialFrame,
} from '@/domain/spatial/analytics';

/** Owns bounded session-only samples and its one-second timer; identity/mapping changes discard history. */
export function useSpatialSession(
  identity: () => string,
  targets: () => readonly SpatialTarget[],
  online: () => boolean
) {
  const frames = ref<SpatialFrame[]>([]);
  const now = ref(Date.now());
  const trailSince = ref(Date.now());
  watch(online, () => {
    trailSince.value = Date.now();
  });
  const started = ref(now.value);
  const minutes = ref(DEFAULT_OBSERVATION_MINUTES);
  const showTrails = ref(true);
  const clearRequested = ref(false);
  const timer = setInterval(() => {
    now.value = Date.now();
    if (online()) frames.value = appendFrame(frames.value, targets(), now.value);
    else frames.value = frames.value.filter((frame) => frame.at > now.value - SPATIAL_SESSION_MS);
  }, SPATIAL_SAMPLE_INTERVAL_MS);
  onBeforeUnmount(() => clearInterval(timer));
  /** Discards only this tab's observed positions and restarts coverage timing. */
  function clear() {
    frames.value = [];
    started.value = Date.now();
    clearRequested.value = false;
  }
  watch(identity, clear);
  const since = computed(() => now.value - minutes.value * MINUTE_MS);
  const cells = computed(() => heatCells(frames.value, since.value));
  const activity = computed(() => activityByHour(frames.value, since.value));
  const trails = computed(() =>
    online() && showTrails.value
      ? targetTrails(
          frames.value.filter((f) => f.at >= trailSince.value),
          now.value
        )
      : []
  );
  const observedFrames = computed(() => frames.value.filter((f) => f.at >= since.value && f.targets.length).length);
  return { minutes, showTrails, cells, activity, trails, started, observedFrames, clearRequested, clear };
}
