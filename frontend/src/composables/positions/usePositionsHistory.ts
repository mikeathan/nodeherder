import { computed, ref, watch } from 'vue';
import type {
  SpatialHeatCell,
  SpatialPoint,
  SpatialSession,
} from '@/services/spatial-history.service';
import { spatialHistoryApi } from '@/services/spatial-history.service';

/** The player's history range. */
export type PositionsPreset = '1h' | '24h' | '7d';

/** Range lengths, ms. One owner for the preset spans. */
export const POSITIONS_PRESET_SPANS: Readonly<Record<PositionsPreset, number>> = {
  '1h': 3_600_000,
  '24h': 86_400_000,
  '7d': 604_800_000,
};

export const POSITIONS_PRESETS: readonly PositionsPreset[] = ['1h', '24h', '7d'];

/**
 * Owns the read side of the Positions view: the filter, the recordings summary, the confirmed
 * recording's points and the heat cells.
 *
 * `selected` is the recording the operator asked for; `confirmed` is the one whose points are
 * rendered. Keeping them apart is what lets a slow or failed switch keep the verified content on
 * screen instead of presenting an unloaded recording as current. Every request carries a
 * generation counter, so a superseded reply can never overwrite newer state.
 *
 * `supersededAt` counts replies that arrived for a request that had been replaced. Playback
 * watches it to settle a pending continuous advance; a counter avoids a callback cycle between
 * the two composables and cannot be missed if a reply lands before playback exists.
 */
export function usePositionsHistory(deviceId: () => string) {
  const preset = ref<PositionsPreset>('24h');
  const to = ref(Date.now());
  const from = computed(() => to.value - POSITIONS_PRESET_SPANS[preset.value]);

  const sessions = ref<SpatialSession[]>([]);
  const points = ref<SpatialPoint[]>([]);
  const cells = ref<SpatialHeatCell[]>([]);
  const selected = ref('');
  const confirmed = ref('');

  const supersededAt = ref(0);
  const sessionsLoading = ref(false);
  const pointsLoading = ref(false);
  const heatmapLoading = ref(false);
  const sessionsError = ref('');
  const pointsError = ref('');
  const heatmapError = ref('');

  let pointsRequest = 0;
  let sessionsRequest = 0;
  let heatmapRequest = 0;

  /**
   * Loads one recording's points into the confirmed slot.
   *
   * @param sessionId Recording to confirm.
   * @param at Playhead to apply; defaults to the recording's first sample.
   * @returns true only when this request is the one that confirmed the recording. A stale or
   * failed request resolves false, so callers cannot start playback from an unloaded recording.
   */
  async function loadPoints(sessionId: string, at?: number): Promise<boolean> {
    const request = ++pointsRequest;
    pointsLoading.value = true;
    pointsError.value = '';
    try {
      const next = await spatialHistoryApi.points(deviceId(), sessionId, from.value, to.value);
      if (request !== pointsRequest) {
        supersededAt.value++;
        return false;
      }
      confirmed.value = sessionId;
      points.value = next;
      return true;
    } catch (cause) {
      if (request !== pointsRequest) {
        supersededAt.value++;
        return false;
      }
      pointsError.value = (cause as Error).message;
      selected.value = confirmed.value;
      return false;
    } finally {
      if (request === pointsRequest) pointsLoading.value = false;
    }
  }

  /**
   * Loads the recording list for the filter, then confirms one of them.
   *
   * Keeps the current selection when it is still in range, otherwise selects the newest recording
   * (spec 057 AC-04). A reply for a superseded filter is ignored.
   */
  async function loadSessions(): Promise<void> {
    const request = ++sessionsRequest;
    sessionsLoading.value = true;
    sessionsError.value = '';
    try {
      const next = await spatialHistoryApi.sessions(deviceId(), from.value, to.value);
      if (request !== sessionsRequest) return;
      sessions.value = next;
      if (!next.length) {
        selected.value = '';
        confirmed.value = '';
        points.value = [];
        return;
      }
      const keep = next.find((session) => session.id === selected.value) ?? next[next.length - 1];
      selected.value = keep.id;
      await loadPoints(keep.id);
    } catch (cause) {
      if (request !== sessionsRequest) return;
      sessionsError.value = (cause as Error).message;
    } finally {
      if (request === sessionsRequest) sessionsLoading.value = false;
    }
  }

  /** Loads the heat cells for the filter. A superseded reply is ignored. */
  async function loadHeatmap(): Promise<void> {
    const request = ++heatmapRequest;
    heatmapLoading.value = true;
    heatmapError.value = '';
    try {
      const next = await spatialHistoryApi.heatmap(deviceId(), from.value, to.value);
      if (request !== heatmapRequest) return;
      cells.value = next;
    } catch (cause) {
      if (request !== heatmapRequest) return;
      heatmapError.value = (cause as Error).message;
      cells.value = [];
    } finally {
      if (request === heatmapRequest) heatmapLoading.value = false;
    }
  }

  /** Re-reads the current filter without changing it. */
  function refresh(): void {
    void loadSessions();
    void loadHeatmap();
  }

  /**
   * Switches range. Re-selecting the active range refreshes it, which is what an operator expects
   * from pressing the range they are already on.
   */
  function choosePreset(value: PositionsPreset): void {
    if (preset.value === value) {
      to.value = Date.now();
      refresh();
      return;
    }
    preset.value = value;
  }

  // One owner for "the filter changed": no caller has to remember to reload and reset.
  watch([preset, deviceId], () => {
    to.value = Date.now();
    void loadSessions();
    void loadHeatmap();
  }, { immediate: true });

  return {
    preset,
    from,
    to,
    sessions,
    points,
    cells,
    selected,
    confirmed,
    supersededAt,
    sessionsLoading,
    pointsLoading,
    heatmapLoading,
    sessionsError,
    pointsError,
    heatmapError,
    loadPoints,
    loadSessions,
    loadHeatmap,
    refresh,
    choosePreset,
  };
}
