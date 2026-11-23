<script setup lang="ts">
  import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue';
  import { copyWorkspace, screenPoint } from '@/domain/spatial/workspace';
  import {
    SPATIAL_DEFAULT_ORIENTATION,
    SPATIAL_DEFAULT_SCALE_MM,
    SPATIAL_ILLUSTRATIVE_FOV,
  } from '@/domain/spatial';
  import { sensorProjectedPoint } from '@/domain/spatial/canvas';
  import { fittedPointsBounds, positionViewMetrics, zoomAnchorPoint } from '@/domain/spatial/positions';
  import { SpatialViewIds } from '@/domain/spatial/view';
  import type { ZoneOccupancy } from '@/domain/spatial/occupancy';
  import { type HeatCell } from '@/domain/spatial/analytics';
  import { useDeviceSettings } from '@/composables/useDeviceSettings';
  import { usePositionsHistory, POSITIONS_PRESETS } from '@/composables/positions/usePositionsHistory';
  import { usePositionsPlayback } from '@/composables/positions/usePositionsPlayback';
  import { useElementSize } from '@/composables/positions/useElementSize';
  import { emptyRecordingLabel } from '@/domain/spatial/positions-labels';
  import {
    POSITIONS_HOP_MM,
    POSITIONS_IDLE_GAP_MS,
    POSITIONS_LAST_KNOWN_MS,
    idleSpans,
    playedDistanceMm,
    sessionSegments,
    trailGroups,
  } from '@/domain/spatial/playback';
  import {
    latestSampleAt,
    playbackHopLinks,
    playbackTrail,
    playbackTracks,
    recordedEnvelope,
  } from '@/domain/spatial/positions-playback';
  import SpatialCanvas from './SpatialCanvas.vue';
  import SpatialViewer from './SpatialViewer.vue';
  import PositionsCoverageBar from './PositionsCoverageBar.vue';
  import PositionsSessionRail from './PositionsSessionRail.vue';
  import PositionsTransport from './PositionsTransport.vue';
  import UiButton from '@/components/ui/UiButton.vue';
  import UiDialog from '@/components/ui/UiDialog.vue';

  const props = defineProps<{ deviceId: string }>();
  const { settings } = useDeviceSettings(() => props.deviceId);
  const workspace = computed(() => copyWorkspace(settings.value.spatial?.workspace));
  const noOccupancy: ReadonlyMap<string, ZoneOccupancy> = new Map();

  const history = usePositionsHistory(() => props.deviceId);
  // Destructured to top-level bindings: Vue unwraps refs in the template only for setup-level
  // bindings, not for properties reached through an object returned by a composable.
  const {
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
    loadHeatmap,
    refresh,
    choosePreset,
  } = history;

  const playbackState = usePositionsPlayback({
    sessions,
    confirmed,
    load: (sessionId) => loadPoints(sessionId),
    supersededAt,
  });
  const {
    time,
    playing,
    speed,
    reachedEnd,
    awaitingCrossing,
    pendingGap,
    skippedGap,
    neighbours,
    recordings,
    playback,
  } = playbackState;

  const view = ref<'player' | 'heatmap'>('player');
  const showPath = ref(true);
  const showGrid = ref(true);
  const sheetOpen = ref(false);
  const railOpen = ref(true);
  const zoom = ref(1);
  const zoomAnchor = ref<{ x: number; y: number } | null>(null);
  const showFov = ref(false);
  const showDistances = ref(false);
  const viewer = ref<HTMLElement | null>(null);
  const { width: viewerWidth, height: viewerHeight, measure: measureViewer } = useElementSize(viewer);

  const configuredFov = computed(() => workspace.value.sensor.fov);
  const displayFov = computed(() =>
    configuredFov.value > 0 ? configuredFov.value : SPATIAL_ILLUSTRATIVE_FOV
  );
  const activeSession = computed(() =>
    sessions.value.find((session) => session.id === confirmed.value)
  );
  const sessionsNewestFirst = computed(() => [...sessions.value].sort((a, b) => b.from - a.from));
  const overview = computed(() => recordedEnvelope(sessions.value, from.value, to.value));
  const segments = computed(() =>
    sessionSegments(sessions.value, overview.value.from, overview.value.to)
  );
  const cellCount = computed(() => cells.value.length);

  /** Heat cells with their intensity normalised for the canvas. */
  const heat = computed<HeatCell[]>(() => {
    const max = Math.max(1, ...cells.value.map((cell) => cell.count));
    return cells.value.map((cell) => ({ ...cell, intensity: cell.count / max }));
  });

  /** Confirmed points projected into screen millimetres, with their timestamps. */
  const projectedPoints = computed(() =>
    points.value.map((point) => ({
      at: point.at,
      ...sensorProjectedPoint(point, workspace.value.sensor, SPATIAL_DEFAULT_ORIENTATION),
    }))
  );
  const shownHeatCells = computed(() => (view.value === 'heatmap' ? cells.value : []));

  /** Fitted view box for the canvas; zoom shrinks it around the anchor. */
  const fittedBounds = computed(() =>
    fittedPointsBounds({
      workspace: workspace.value,
      points: projectedPoints.value,
      heatCells: shownHeatCells.value,
      orientation: SPATIAL_DEFAULT_ORIENTATION,
    })
  );
  const viewMetrics = computed(() =>
    positionViewMetrics({
      bounds: fittedBounds.value,
      center: screenPoint(workspace.value.sensor, SPATIAL_DEFAULT_ORIENTATION),
      viewerWidth: viewerWidth.value,
      viewerHeight: viewerHeight.value,
      zoom: zoom.value,
    })
  );
  const baseViewBox = computed(() => viewMetrics.value.viewBox);
  const symbolScale = computed(() => viewMetrics.value.symbolScale);
  const labelStroke = computed(() => viewMetrics.value.labelStroke);
  /** Session/preset/view switches clear the viewer's pan. */
  const resetKey = computed(() => `${confirmed.value}:${preset.value}:${view.value}`);

  /**
   * Re-anchors the zoom on what is currently on screen. Called when the level changes or the
   * operator moves the playhead, never on a playback tick.
   */
  function captureZoomAnchor(): void {
    zoomAnchor.value =
      zoom.value > 1
        ? zoomAnchorPoint({
            heatCells: cells.value,
            projectedPoints: projectedPoints.value,
            at: time.value,
            sensor: workspace.value.sensor,
            orientation: SPATIAL_DEFAULT_ORIENTATION,
            heatmap: view.value === 'heatmap',
          })
        : null;
  }

  /** Explicit seeks re-anchor the zoom; playback ticks leave the view fixed. */
  function seekTo(value: number): void {
    playback.seek(value);
    captureZoomAnchor();
  }

  const groupedTrails = computed(() => trailGroups(points.value, time.value));
  const trails = computed(() =>
    groupedTrails.value
      .filter((group) => group.points.length)
      .map((group) => playbackTrail(group.slot, group.points))
  );
  const recentTrails = computed(() =>
    groupedTrails.value
      .filter((group) => group.recent.length > 1)
      .map((group) => playbackTrail(group.slot, group.recent))
  );
  const currentSampleAt = computed(() => latestSampleAt(points.value, time.value));

  /**
   * Spans with no recorded sample, including the head and tail of the recording. The transport
   * draws the same spans, so the idle notice and the stripes cannot disagree.
   */
  const idleGaps = computed(() =>
    activeSession.value
      ? idleSpans(points.value, activeSession.value.from, activeSession.value.to)
      : []
  );
  const activeIdleGap = computed(
    () => idleGaps.value.find((gap) => time.value >= gap.from && time.value < gap.to) ?? null
  );
  const tracks = computed(() =>
    playbackTracks(points.value, time.value, activeIdleGap.value !== null, POSITIONS_LAST_KNOWN_MS)
  );
  const hopLinks = computed(() =>
    playbackHopLinks(points.value, time.value, POSITIONS_IDLE_GAP_MS, POSITIONS_HOP_MM)
  );
  const idleForMs = computed(() => {
    const gap = activeIdleGap.value;
    return gap && currentSampleAt.value ? Math.max(0, time.value - gap.from) : 0;
  });
  const distanceMm = computed(() => playedDistanceMm(points.value, time.value));
  const totalDistanceMm = computed(() => playedDistanceMm(points.value, Number.POSITIVE_INFINITY));

  const hasRecordings = computed(() => recordings.value.length > 1);
  /** What an empty-looking recording means, given what its own summary claimed. */
  const emptyRecordingText = computed(() =>
    emptyRecordingLabel(activeSession.value?.points ?? 0)
  );
  /** Requested but not yet confirmed; shown as pending, never as current. */
  const pendingId = computed(() => (selected.value !== confirmed.value ? selected.value : ''));

  /** Selects a recording without starting playback. */
  function selectSession(id: string): void {
    if (!id) return;
    selected.value = id;
    sheetOpen.value = false;
    void playbackState.select(id);
  }

  /** Plays one recording from a recorded moment inside it. */
  function playFrom(payload: { id: string; at: number }): void {
    if (!payload.id) return;
    selected.value = payload.id;
    sheetOpen.value = false;
    void playbackState.play(payload.id, payload.at);
  }

  /** Starts the whole filtered history at its earliest recording. */
  function playAll(): void {
    const first = recordings.value[0];
    if (!first) return;
    selected.value = first.id;
    sheetOpen.value = false;
    void playbackState.play(first.id, first.from);
  }

  /** Moves to the adjacent recording, keeping the play/pause state. */
  function goToRecording(id: string | null): void {
    if (!id) return;
    selected.value = id;
    sheetOpen.value = false;
    void playbackState.goTo(id);
  }

  /**
   * Chapter-style shortcuts, matching the players this view is modelled on. They require Shift so
   * they never fight the scrubber's own arrow keys, and they do nothing in the heatmap view.
   */
  function onKeydown(event: KeyboardEvent): void {
    if (view.value !== 'player' || !event.shiftKey) return;
    // A window-level shortcut must not steal a chord from an input, a modifier chord, or a
    // dialog sitting on top of the player.
    if (event.altKey || event.ctrlKey || event.metaKey) return;
    if (sheetOpen.value) return;
    const target = event.target as HTMLElement | null;
    if (target && (target.isContentEditable || /^(INPUT|TEXTAREA|SELECT)$/.test(target.tagName))) return;
    if (event.key === 'ArrowRight') {
      event.preventDefault();
      goToRecording(neighbours.value.next);
    } else if (event.key === 'ArrowLeft') {
      event.preventDefault();
      goToRecording(neighbours.value.previous);
    }
  }
  onMounted(() => window.addEventListener('keydown', onKeydown));
  onBeforeUnmount(() => window.removeEventListener('keydown', onKeydown));

  watch(zoom, captureZoomAnchor);
  // A recording change while zoomed must re-anchor: the baseline did this after every load, and
  // without it the zoom keeps the previous recording's target. Post-flush so the new playhead is
  // already placed when the anchor is read.
  watch(confirmed, () => captureZoomAnchor(), { flush: 'post' });

  function chooseView(value: 'player' | 'heatmap'): void {
    /*
     * The transport is unmounted in the heatmap, so a running clock would keep crossing recording
     * boundaries and issuing points requests for a surface nobody can see. Leaving the player is a
     * pause, not a hidden play.
     */
    if (value === 'heatmap') playback.pause();
    view.value = value;
    captureZoomAnchor();
    if (value === 'heatmap') void loadHeatmap();
  }

  // The filter or device changed: stop the clock and drop a now-meaningless zoom anchor. The
  // history composable owns re-reading the data, so this only resets playback.
  watch([preset, () => props.deviceId], () => {
    playback.pause();
    zoomAnchor.value = null;
  });

  // Layout changes can precede the observer callback; measure after the DOM update.
  watch([railOpen, view], () => measureViewer(), { flush: 'post' });
</script>

<template>
  <section
    class="spatial-positions"
    aria-label="Recorded positions"
    :aria-busy="
      sessionsLoading ||
      pointsLoading ||
      (view === 'heatmap' && heatmapLoading)
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
            v-for="option in POSITIONS_PRESETS"
            :key="option"
            size="sm"
            :variant="preset === option ? 'primary' : 'ghost'"
            :pressed="preset === option"
            @click="choosePreset(option)">
            {{ option }}
          </UiButton>
        </div>
        <!--
          One action for the whole filtered  The viewer opens on the newest recording
          (spec 057 AC-04), so without this the operator would have to page backwards before a
          continuous play could mean anything.
        -->
        <UiButton
          v-if="view === 'player' && hasRecordings"
          class="nh-play-all"
          size="sm"
          icon="run"
          @click="playAll">
          Play all recordings
        </UiButton>
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
        <UiButton
          size="sm"
          variant="ghost"
          :loading="sessionsLoading"
          @click="refresh">
          Refresh
        </UiButton>
      </div>
    </div>

    <p v-if="sessionsError" class="nh-alert is-danger" role="alert">
      {{ sessionsError }}
    </p>
    <p v-else-if="pointsError" class="nh-alert is-danger" role="alert">
      {{ pointsError }}
    </p>
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
          v-model:show-distances="showDistances"
          distance-option
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
                :show-distances="showDistances"
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
          :pending-id="pendingId"
          :from="overview.from"
          :to="overview.to"
          :loading="sessionsLoading"
          @play-from="playFrom" />

        <!--
          A recording can be listed from its summary while its detail is empty (the summary counts
          points inside the filter, the detail returns them). Saying only "no positions" leaves the
          operator thinking the recording is broken, so state both numbers when they disagree.
        -->
        <p v-if="confirmed && !pointsLoading && !points.length" class="nh-muted">
          {{ emptyRecordingText }}
        </p>

        <PositionsTransport
          v-if="activeSession"
          :session="activeSession"
          :neighbours="neighbours"
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
          :reached-end="reachedEnd"
          :awaiting-crossing="awaitingCrossing"
          :pending-gap="pendingGap"
          :skipped-gap="skippedGap"
          @seek="seekTo"
          @toggle="playback.toggle"
          @next="seekTo"
          @previous-recording="goToRecording(neighbours.previous)"
          @next-recording="goToRecording(neighbours.next)"
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
          :pending-id="pendingId"
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
      <p v-if="heatmapError" class="nh-alert is-danger" role="alert">
        {{ heatmapError }}
      </p>
      <p v-else-if="!cellCount && !heatmapLoading" class="nh-muted">
        No recorded positions in this range.
      </p>
    </div>

    <UiDialog :open="sheetOpen" title="Recordings" drag-dismiss @close="sheetOpen = false">
      <PositionsSessionRail
        :sessions="sessionsNewestFirst"
        :selected-id="confirmed"
        :pending-id="pendingId"
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
    font-size: 0.72rem;
    font-weight: 600;
    letter-spacing: 0.04em;
    text-transform: uppercase;
    color: var(--nh-text-muted);
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
