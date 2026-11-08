<script setup lang="ts">
  import { computed, onBeforeUnmount, ref, watch } from 'vue';
  import { copyWorkspace, screenPoint } from '@/domain/spatial/workspace';
  import { SPATIAL_DEFAULT_ORIENTATION, SPATIAL_DEFAULT_SCALE_MM, SPATIAL_ILLUSTRATIVE_FOV } from '@/domain/spatial';
  import { sensorProjectedPoint, CANVAS_SENSOR_FONT_RATIO } from '@/domain/spatial/canvas';
  import { radarViewBox } from '@/domain/spatial/radar';
  import { SpatialViewIds } from '@/domain/spatial/view';
  import type { SpatialDisplayTrack } from '@/domain/spatial/tracks';
  import type { ZoneOccupancy } from '@/domain/spatial/occupancy';
  import { HEAT_CELL_MM, type HeatCell } from '@/domain/spatial/analytics';
  import type { SpatialHeatCell, SpatialPoint, SpatialSession } from '@/services/spatial-history.service';
  import { spatialHistoryApi } from '@/services/spatial-history.service';
  import { useDeviceSettings } from '@/composables/useDeviceSettings';
  import { useSpatialPlayback } from '@/composables/useSpatialPlayback';
  import {
    POSITIONS_HOP_MM,
    POSITIONS_IDLE_GAP_MS,
    POSITIONS_LAST_KNOWN_MS,
    clock,
    idleSpans,
    playedDistanceMm,
    sessionSegments,
    trailGroups,
  } from '@/domain/spatial/playback';
  import SpatialCanvas from './SpatialCanvas.vue';
  import SpatialViewer from './SpatialViewer.vue';
  import PositionsCoverageBar from './PositionsCoverageBar.vue';
  import PositionsSessionRail from './PositionsSessionRail.vue';
  import PositionsTransport from './PositionsTransport.vue';
  import UiButton from '@/components/ui/UiButton.vue';
  import UiDialog from '@/components/ui/UiDialog.vue';

  type Preset = '1h' | '24h' | '7d';

  const props = defineProps<{ deviceId: string }>();
  const { settings } = useDeviceSettings(() => props.deviceId);
  const workspace = computed(() => copyWorkspace(settings.value.spatial?.workspace));
  const noOccupancy: ReadonlyMap<string, ZoneOccupancy> = new Map();

  const preset = ref<Preset>('24h');
  const view = ref<'player' | 'heatmap'>('player');
  const to = ref(Date.now());
  const spans: Record<Preset, number> = {
    '1h': 3_600_000,
    '24h': 86_400_000,
    '7d': 604_800_000,
  };
  const from = computed(() => to.value - spans[preset.value]);

  const sessions = ref<SpatialSession[]>([]);
  const points = ref<SpatialPoint[]>([]);
  const cells = ref<SpatialHeatCell[]>([]);
  /**
   * `selected` is the recording the user asked for; `confirmed` is the one whose points are
   * rendered. Keeping them apart lets a slow or failed switch keep the verified content on
   * screen instead of presenting an unloaded session as current.
   */
  const selected = ref('');
  const confirmed = ref('');

  const sessionsLoading = ref(false);
  const pointsLoading = ref(false);
  const heatmapLoading = ref(false);
  const sessionsError = ref('');
  const pointsError = ref('');
  const heatmapError = ref('');
  const showPath = ref(true);
  const showGrid = ref(true);
  const sheetOpen = ref(false);
  const railOpen = ref(true);
  const zoom = ref(1);
  const zoomAnchor = ref<{ x: number; y: number } | null>(null);
  const showFov = ref(false);
  const configuredFov = computed(() => workspace.value.sensor.fov);
  const displayFov = computed(() =>
    configuredFov.value > 0 ? configuredFov.value : SPATIAL_ILLUSTRATIVE_FOV
  );
  const viewer = ref<HTMLElement | null>(null);
  const viewerWidth = ref(0);
  const viewerHeight = ref(0);
  let viewerObserver: ResizeObserver | null = null;

  const activeSession = computed(() =>
    sessions.value.find((session) => session.id === confirmed.value)
  );
  const bounds = computed(() =>
    activeSession.value
      ? { from: activeSession.value.from, to: activeSession.value.to }
      : null
  );
  const playback = useSpatialPlayback(() => bounds.value);
  const { time, playing, speed } = playback;

  const sessionsNewestFirst = computed(() =>
    [...sessions.value].sort((a, b) => b.from - a.from)
  );
  /**
   * The overview fits the recorded envelope instead of the whole preset: a four-minute
   * session inside a 24 h window would otherwise be a sub-pixel sliver. Sessions are
   * returned ascending, so the first start and last end bound every recording.
   */
  const overviewFrom = computed(() => sessions.value[0]?.from ?? from.value);
  const overviewTo = computed(() => sessions.value.at(-1)?.to ?? to.value);
  const segments = computed(() =>
    sessionSegments(sessions.value, overviewFrom.value, overviewTo.value)
  );
  const cellCount = computed(() => cells.value.length);
  const heat = computed<HeatCell[]>(() => {
    const max = Math.max(1, ...cells.value.map((cell) => cell.count));
    return cells.value.map((cell) => ({ ...cell, intensity: cell.count / max }));
  });

  /**
   * Standard view fits the room and the confirmed session's points (plus visible heat cells
   * in that mode). The zoom buttons shrink this box around its centre so points and the
   * room grow without changing the panel height.
   */
  const fittedBounds = computed(() => {
    const orientation = SPATIAL_DEFAULT_ORIENTATION;
    const { outline, sensor } = workspace.value;
    let minX = Number.POSITIVE_INFINITY;
    let minY = Number.POSITIVE_INFINITY;
    let maxX = Number.NEGATIVE_INFINITY;
    let maxY = Number.NEGATIVE_INFINITY;
    const add = (point: { x: number; y: number }) => {
      minX = Math.min(minX, point.x);
      minY = Math.min(minY, point.y);
      maxX = Math.max(maxX, point.x);
      maxY = Math.max(maxY, point.y);
    };
    add(screenPoint(sensor, orientation));
    for (const point of outline) add(screenPoint(point, orientation));
    for (const point of points.value) add(sensorProjectedPoint(point, sensor, orientation));
    if (view.value === 'heatmap') {
      for (const cell of cells.value) {
        add(sensorProjectedPoint(cell, sensor, orientation));
        add(
          sensorProjectedPoint(
            { x: cell.x + HEAT_CELL_MM, y: cell.y + HEAT_CELL_MM },
            sensor,
            orientation
          )
        );
      }
    }
    if (!Number.isFinite(minX)) return null;
    const width = Math.max(500, maxX - minX);
    const height = Math.max(500, maxY - minY);
    const pad = Math.max(width, height) * 0.08 + 150;
    return { minX: minX - pad, minY: minY - pad, maxX: maxX + pad, maxY: maxY + pad };
  });

  /** Panel aspect plus the extent that fits the content before the user's zoom. */
  const fittedExtent = computed(() => {
    const bounds = fittedBounds.value;
    const width = viewerWidth.value;
    const height = viewerHeight.value;
    if (!bounds || width <= 0 || height <= 0) return null;
    const aspect = width / height;
    const boundsWidth = bounds.maxX - bounds.minX;
    const boundsHeight = bounds.maxY - bounds.minY;
    return { aspect, extent: Math.max(boundsHeight, boundsWidth / aspect) };
  });

  /**
   * Screen point used to anchor a zoom level: the current sample (or the heat activity).
   * It is captured when zoom changes or the playhead is moved explicitly, not on every
   * playback tick, so the dot visibly travels across a fixed view.
   */
  function zoomAnchorPoint() {
    const orientation = SPATIAL_DEFAULT_ORIENTATION;
    const { sensor } = workspace.value;
    if (view.value === 'heatmap' && cells.value.length) {
      let minX = Number.POSITIVE_INFINITY;
      let minY = Number.POSITIVE_INFINITY;
      let maxX = Number.NEGATIVE_INFINITY;
      let maxY = Number.NEGATIVE_INFINITY;
      const add = (point: { x: number; y: number }) => {
        minX = Math.min(minX, point.x);
        minY = Math.min(minY, point.y);
        maxX = Math.max(maxX, point.x);
        maxY = Math.max(maxY, point.y);
      };
      for (const cell of cells.value) {
        add(sensorProjectedPoint(cell, sensor, orientation));
        add(
          sensorProjectedPoint(
            { x: cell.x + HEAT_CELL_MM, y: cell.y + HEAT_CELL_MM },
            sensor,
            orientation
          )
        );
      }
      return { x: (minX + maxX) / 2, y: (minY + maxY) / 2 };
    }
    const current = points.value.filter((point) => point.at <= time.value).at(-1);
    return current
      ? sensorProjectedPoint(current, sensor, orientation)
      : screenPoint(sensor, orientation);
  }

  /** Base fitted view; the shared viewer applies zoom, the active-sample anchor and pan. */
  const baseViewBox = computed(() => {
    const bounds = fittedBounds.value;
    const fitted = fittedExtent.value;
    if (!bounds || !fitted) return radarViewBox(SPATIAL_DEFAULT_SCALE_MM);
    const width = fitted.extent * fitted.aspect;
    const centerX = (bounds.minX + bounds.maxX) / 2;
    const centerY = (bounds.minY + bounds.maxY) / 2;
    const round = (value: number) => Number(value.toFixed(1));
    return [
      round(centerX - width / 2),
      round(centerY - fitted.extent / 2),
      round(width),
      round(fitted.extent),
    ].join(' ');
  });

  /** Session/preset/view switches clear the viewer's pan. */
  const resetKey = computed(() => `${confirmed.value}:${preset.value}:${view.value}`);

  /**
   * Keeps labels and markers at a readable pixel size regardless of the fitted extent or the
   * zoom level; only the map content (room, path, cells) scales.
   */
  const symbolScale = computed(() => {
    const fitted = fittedExtent.value;
    if (!fitted) return zoom.value;
    const fontMm = SPATIAL_DEFAULT_SCALE_MM * CANVAS_SENSOR_FONT_RATIO;
    const targetPx = 14;
    return (fontMm * viewerHeight.value * zoom.value) / (targetPx * fitted.extent);
  });

  /** Halo around labels, ~25% of their text size, so it never becomes a blob at high zoom. */
  const labelStroke = computed(() => {
    if (!fittedExtent.value) return 8;
    return (SPATIAL_DEFAULT_SCALE_MM * CANVAS_SENSOR_FONT_RATIO * 0.25) / symbolScale.value;
  });

  const groupedTrails = computed(() => trailGroups(points.value, time.value));
  const trails = computed(() =>
    groupedTrails.value
      .filter((group) => group.points.length)
      .map((group) => toTrail(group.slot, group.points))
  );
  const recentTrails = computed(() =>
    groupedTrails.value
      .filter((group) => group.recent.length > 1)
      .map((group) => toTrail(group.slot, group.recent))
  );

  const currentSampleAt = computed(() => {
    let latest = 0;
    for (const point of points.value) {
      if (point.at > time.value) break;
      latest = point.at;
    }
    return latest;
  });

  /**
   * Gaps without a recorded sample, including the head and tail of the session. These are the
   * same spans the transport draws as stripes, so the target indicator and the idle label
   * cover the whole gap instead of only the part past the 10 s threshold.
   */
  const idleGaps = computed(() =>
    activeSession.value
      ? idleSpans(points.value, activeSession.value.from, activeSession.value.to)
      : []
  );
  const activeIdleGap = computed(() => {
    const at = time.value;
    return idleGaps.value.find((gap) => at >= gap.from && at < gap.to) ?? null;
  });

  const tracks = computed<SpatialDisplayTrack[]>(() => {
    const at = time.value;
    if (!at) return [];
    const latest = new Map<number, SpatialPoint>();
    for (const point of points.value) {
      if (point.at > at) break;
      latest.set(point.slot, point);
    }
    const inGap = activeIdleGap.value !== null;
    return [...latest.entries()].map(([slot, point]): SpatialDisplayTrack => {
      const silentMs = at - point.at;
      const stale = silentMs > POSITIONS_LAST_KNOWN_MS;
      const idle = !stale && inGap;
      return {
        id: slot,
        label: stale ? `Slot ${slot} · last seen ${clock(point.at)}` : `Slot ${slot}`,
        slot,
        x: point.x,
        y: point.y,
        rawX: point.x,
        rawY: point.y,
        vx: 0,
        vy: 0,
        at: point.at,
        colorIndex: (slot - 1) % 4,
        state: stale ? 'held' : idle ? 'idle' : 'live',
      };
    });
  });

  /**
   * A connector between the previous and current sample. Drawn after a pause or a large
   * jump, so a position that appears somewhere else explains where it came from.
   */
  const hopLinks = computed(() => {
    if (!time.value) return [];
    const pairs = new Map<number, { current?: SpatialPoint; previous?: SpatialPoint }>();
    for (const point of points.value) {
      if (point.at > time.value) break;
      const entry = pairs.get(point.slot) ?? {};
      entry.previous = entry.current;
      entry.current = point;
      pairs.set(point.slot, entry);
    }
    const links: { from: SpatialPoint; to: SpatialPoint }[] = [];
    for (const { current, previous } of pairs.values()) {
      if (!current || !previous) continue;
      const jumped =
        Math.hypot(current.x - previous.x, current.y - previous.y) > POSITIONS_HOP_MM;
      if (current.at - previous.at > POSITIONS_IDLE_GAP_MS || jumped) {
        links.push({ from: previous, to: current });
      }
    }
    return links;
  });

  const distanceMm = computed(() => playedDistanceMm(points.value, time.value));
  const totalDistanceMm = computed(() =>
    playedDistanceMm(points.value, Number.POSITIVE_INFINITY)
  );
  const idleForMs = computed(() => {
    const gap = activeIdleGap.value;
    if (!gap || !currentSampleAt.value) return 0;
    return Math.max(0, time.value - gap.from);
  });

  function toTrail(slot: number, points: SpatialPoint[]) {
    return {
      id: slot,
      targets: points.map((point) => ({
        id: point.slot,
        x: point.x,
        y: point.y,
        distance: 0,
      })),
    };
  }

  let pointsRequest = 0;
  let sessionsRequest = 0;
  let heatmapRequest = 0;

  async function loadPoints(sessionId: string) {
    const request = ++pointsRequest;
    pointsLoading.value = true;
    pointsError.value = '';
    try {
      const next = await spatialHistoryApi.points(
        props.deviceId,
        sessionId,
        from.value,
        to.value
      );
      if (request !== pointsRequest) return;
      confirmed.value = sessionId;
      points.value = next;
      playback.pause();
      playback.seek(next[0]?.at ?? sessions.value.find((s) => s.id === sessionId)?.from ?? 0);
      if (zoom.value > 1) zoomAnchor.value = zoomAnchorPoint();
    } catch (cause) {
      if (request !== pointsRequest) return;
      pointsError.value = (cause as Error).message;
      selected.value = confirmed.value;
    } finally {
      if (request === pointsRequest) pointsLoading.value = false;
    }
  }

  async function loadSessions() {
    const request = ++sessionsRequest;
    sessionsLoading.value = true;
    sessionsError.value = '';
    try {
      const next = await spatialHistoryApi.sessions(props.deviceId, from.value, to.value);
      if (request !== sessionsRequest) return;
      sessions.value = next;
      if (confirmed.value && !next.some((session) => session.id === confirmed.value)) {
        confirmed.value = '';
        points.value = [];
        playback.pause();
        playback.seek(0);
        zoomAnchor.value = null;
      }
      if (!next.length) {
        selected.value = '';
        confirmed.value = '';
        points.value = [];
        playback.pause();
        playback.seek(0);
        zoomAnchor.value = null;
        return;
      }
      const keep =
        next.find((session) => session.id === selected.value) ?? next[next.length - 1];
      selected.value = keep.id;
      await loadPoints(keep.id);
    } catch (cause) {
      if (request !== sessionsRequest) return;
      sessionsError.value = (cause as Error).message;
    } finally {
      if (request === sessionsRequest) sessionsLoading.value = false;
    }
  }

  async function loadHeatmap() {
    const request = ++heatmapRequest;
    heatmapLoading.value = true;
    heatmapError.value = '';
    try {
      const next = await spatialHistoryApi.heatmap(props.deviceId, from.value, to.value);
      if (request !== heatmapRequest) return;
      cells.value = next;
      if (zoom.value > 1 && view.value === 'heatmap') zoomAnchor.value = zoomAnchorPoint();
    } catch (cause) {
      if (request !== heatmapRequest) return;
      heatmapError.value = (cause as Error).message;
      cells.value = [];
    } finally {
      if (request === heatmapRequest) heatmapLoading.value = false;
    }
  }

  function refresh() {
    void loadSessions();
    void loadHeatmap();
  }

  function selectSession(id: string) {
    if (!id) return;
    selected.value = id;
    sheetOpen.value = false;
    void loadPoints(id);
  }

  /** Explicit seeks re-anchor the zoom; playback ticks leave the view fixed. */
  function seekTo(value: number) {
    playback.seek(value);
    if (zoom.value > 1) zoomAnchor.value = zoomAnchorPoint();
  }

  // The panel owns the buttons; capture the anchor whenever the level changes.
  watch(zoom, (level) => {
    zoomAnchor.value = level > 1 ? zoomAnchorPoint() : null;
  });

  function choosePreset(value: Preset) {
    if (preset.value === value) {
      to.value = Date.now();
      refresh();
      return;
    }
    preset.value = value;
  }

  function chooseView(value: 'player' | 'heatmap') {
    view.value = value;
    zoomAnchor.value =
      value === 'player' && zoom.value > 1 ? zoomAnchorPoint() : null;
    if (value === 'heatmap') void loadHeatmap();
  }

  watch(
    [preset, () => props.deviceId],
    () => {
      to.value = Date.now();
      playback.pause();
      zoomAnchor.value = null;
      void loadSessions();
      void loadHeatmap();
    },
    { immediate: true }
  );

  function measureViewer() {
    const element = viewer.value;
    if (!element) return;
    const rect = element.getBoundingClientRect();
    viewerWidth.value = rect.width;
    viewerHeight.value = rect.height;
  }

  // The viewer aspect drives the fitted viewBox; re-observe when the view swaps the element.
  watch(
    viewer,
    (element) => {
      viewerObserver?.disconnect();
      viewerObserver = null;
      if (!element) return;
      viewerObserver = new ResizeObserver((entries) => {
        const box = entries[0]?.contentRect;
        if (!box) return;
        viewerWidth.value = box.width;
        viewerHeight.value = box.height;
      });
      viewerObserver.observe(element);
      measureViewer();
    },
    { flush: 'post' }
  );

  // Rail/view layout changes can precede the observer callback; measure after the DOM update.
  watch([railOpen, view], () => measureViewer(), { flush: 'post' });
  onBeforeUnmount(() => {
    viewerObserver?.disconnect();
    viewerObserver = null;
  });
</script>

<template>
  <section
    class="spatial-positions"
    aria-label="Recorded positions"
    :aria-busy="
      sessionsLoading || pointsLoading || (view === 'heatmap' && heatmapLoading)
    ">
    <div class="nh-positions-toolbar">
      <div role="group" aria-label="Positions view">
        <UiButton
          size="sm"
          :variant="view === 'player' ? 'primary' : 'ghost'"
          :pressed="view === 'player'"
          @click="chooseView('player')">
          Player
        </UiButton>
        <UiButton
          size="sm"
          :variant="view === 'heatmap' ? 'primary' : 'ghost'"
          :pressed="view === 'heatmap'"
          @click="chooseView('heatmap')">
          Heatmap
        </UiButton>
      </div>
      <div class="nh-positions-actions">
        <div role="group" aria-label="History range">
          <UiButton
            v-for="option in (['1h', '24h', '7d'] as const)"
            :key="option"
            size="sm"
            :variant="preset === option ? 'primary' : 'ghost'"
            :pressed="preset === option"
            @click="choosePreset(option)">
            {{ option }}
          </UiButton>
        </div>
        <UiButton
          v-if="view === 'player'"
          class="nh-recordings-open"
          size="sm"
          variant="ghost"
          @click="sheetOpen = true">
          Recordings ({{ sessions.length }})
        </UiButton>
        <UiButton
          v-if="view === 'player' && !railOpen"
          class="nh-recordings-show"
          size="sm"
          variant="ghost"
          @click="railOpen = true">
          Show recordings ({{ sessions.length }})
        </UiButton>
        <UiButton size="sm" variant="ghost" :loading="sessionsLoading" @click="refresh">
          Refresh
        </UiButton>
      </div>
    </div>

    <p v-if="sessionsError" class="nh-alert is-danger" role="alert">{{ sessionsError }}</p>
    <p v-else-if="pointsError" class="nh-alert is-danger" role="alert">{{ pointsError }}</p>
    <p v-if="sessionsLoading && !sessions.length" class="nh-muted" role="status">
      Loading recorded positions…
    </p>

    <div
      v-if="view === 'player'"
      class="nh-positions-layout"
      :class="{ 'is-rail-collapsed': !railOpen }">
      <div class="nh-positions-main">
        <SpatialViewer
          v-model:zoom="zoom"
          v-model:show-grid="showGrid"
          v-model:show-fov="showFov"
          :base-view-box="baseViewBox"
          :focus="zoom > 1 ? zoomAnchor : null"
          :fov-angle="displayFov"
          :reset-key="resetKey">
          <template #default="{ viewBox: canvasViewBox, panMode }">
            <div ref="viewer" class="nh-viewer">
              <SpatialCanvas
                :workspace="workspace"
                :tracks="tracks"
                :occupancy="noOccupancy"
                :scale="SPATIAL_DEFAULT_SCALE_MM"
                :orientation="SPATIAL_DEFAULT_ORIENTATION"
                :online="true"
                :mode="SpatialViewIds.Live"
                selected-id=""
                :trails="showPath ? trails : []"
                :recent-trails="showPath ? recentTrails : []"
                :hop-links="hopLinks"
                :pulse-at="currentSampleAt"
                :view-box="canvasViewBox"
                :pan-mode="panMode"
                :ui-scale="symbolScale"
                :label-stroke="labelStroke"
                :target-scale="1 + (zoom - 1) * 0.25"
                :show-fov="showFov"
                :fov-angle="displayFov"
                label="Recorded positions canvas"
                :cells="[]"
                :show-grid="showGrid"
                :show-room-context="true" />
            </div>
          </template>
        </SpatialViewer>

        <PositionsCoverageBar
          :segments="segments"
          :selected-id="confirmed"
          :pending-id="selected !== confirmed ? selected : ''"
          :from="overviewFrom"
          :to="overviewTo"
          :time="time"
          :loading="sessionsLoading"
          @select="selectSession" />

        <p v-if="confirmed && !pointsLoading && !points.length" class="nh-muted">
          No recorded positions in this session.
        </p>

        <PositionsTransport
          v-if="activeSession"
          :session="activeSession"
          :time="time"
          :playing="playing"
          :speed="speed"
          :show-path="showPath"
          :points="points"
          :distance-mm="distanceMm"
          :total-distance-mm="totalDistanceMm"
          :idle-for-ms="idleForMs"
          :last-sample-at="currentSampleAt"
          :has-points="points.length > 0"
          @seek="seekTo"
          @toggle="playback.toggle"
          @next="seekTo"
          @update:speed="speed = $event"
          @update:show-path="showPath = $event" />
      </div>

      <aside v-if="railOpen" class="nh-positions-rail" aria-label="Recordings">
        <div class="nh-rail-head">
          <span class="nh-rail-title">Recordings</span>
          <UiButton
            size="sm"
            variant="ghost"
            icon="chevronRight"
            label="Hide recordings"
            @click="railOpen = false" />
        </div>
        <PositionsSessionRail
          :sessions="sessionsNewestFirst"
          :selected-id="confirmed"
          :pending-id="selected !== confirmed ? selected : ''"
          :span-ms="to - from"
          :loading="sessionsLoading || pointsLoading"
          @select="selectSession" />
      </aside>
    </div>

    <div v-else class="nh-positions-main">
      <SpatialViewer
        v-model:zoom="zoom"
        v-model:show-grid="showGrid"
        v-model:show-fov="showFov"
        :base-view-box="baseViewBox"
        :focus="zoom > 1 ? zoomAnchor : null"
        :fov-angle="displayFov"
        :reset-key="resetKey">
        <template #default="{ viewBox: canvasViewBox, panMode }">
          <div ref="viewer" class="nh-viewer">
            <SpatialCanvas
              :workspace="workspace"
              :tracks="[]"
              :occupancy="noOccupancy"
              :scale="SPATIAL_DEFAULT_SCALE_MM"
              :orientation="SPATIAL_DEFAULT_ORIENTATION"
              :online="true"
              :mode="SpatialViewIds.Heatmap"
              selected-id=""
              :trails="[]"
              :view-box="canvasViewBox"
              :pan-mode="panMode"
              :ui-scale="symbolScale"
              :label-stroke="labelStroke"
              :show-fov="showFov"
              :fov-angle="displayFov"
              label="Recorded heatmap canvas"
              :cells="heat"
              :show-grid="showGrid"
              :show-room-context="true" />
          </div>
        </template>
      </SpatialViewer>
      <p v-if="heatmapError" class="nh-alert is-danger" role="alert">{{ heatmapError }}</p>
      <p v-else-if="!cellCount && !heatmapLoading" class="nh-muted">
        No recorded positions in this range.
      </p>
    </div>

    <UiDialog :open="sheetOpen" title="Recordings" drag-dismiss @close="sheetOpen = false">
      <PositionsSessionRail
        :sessions="sessionsNewestFirst"
        :selected-id="confirmed"
        :pending-id="selected !== confirmed ? selected : ''"
        :span-ms="to - from"
        :loading="sessionsLoading || pointsLoading"
        @select="selectSession" />
    </UiDialog>
  </section>
</template>

<style scoped>
  .nh-positions-toolbar {
    display: flex;
    flex-wrap: wrap;
    gap: 0.5rem;
    align-items: center;
    justify-content: space-between;
    margin-block: 0.5rem;
  }
  .nh-positions-actions {
    display: flex;
    flex-wrap: wrap;
    gap: 0.4rem;
    align-items: center;
  }
  .nh-positions-layout {
    display: grid;
    grid-template-columns: minmax(0, 1fr);
    gap: 0.75rem;
    align-items: start;
  }
  .nh-positions-main {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
    min-width: 0;
  }
  .nh-viewer {
    width: 100%;
  }
  /* Wide viewer with a fixed height so the timeline and transport stay reachable. */
  .nh-viewer :deep(.nh-spatial-canvas) {
    aspect-ratio: auto;
    height: clamp(16rem, 40vh, 24rem);
    max-height: none;
  }
  .nh-positions-rail {
    display: none;
  }
  .nh-rail-head {
    display: flex;
    flex: none;
    align-items: center;
    justify-content: space-between;
    gap: 0.5rem;
    min-height: 2.5rem;
    padding: 0.35rem 0.65rem;
    border-bottom: 1px solid var(--nh-border);
  }
  .nh-rail-title {
    font-size: 0.85rem;
    font-weight: 600;
  }
  .nh-recordings-show {
    display: none;
  }
  .nh-positions-layout.is-rail-collapsed {
    grid-template-columns: minmax(0, 1fr);
  }
  @media (min-width: 900px) {
    .nh-positions-layout {
      grid-template-columns: minmax(0, 1fr) minmax(15rem, 18rem);
    }
    .nh-positions-rail {
      display: flex;
      flex-direction: column;
      position: sticky;
      top: 0.5rem;
      max-height: calc(100dvh - 2rem);
      overflow: hidden;
      border: 1px solid var(--nh-border);
      border-radius: var(--nh-radius-m);
      background: var(--nh-surface);
    }
    .nh-positions-rail > .nh-session-rail {
      min-height: 0;
      padding: 0.5rem 0.65rem 0.65rem;
      overflow: auto;
    }
    .nh-recordings-open {
      display: none;
    }
    .nh-recordings-show {
      display: inline-flex;
    }
  }
</style>
