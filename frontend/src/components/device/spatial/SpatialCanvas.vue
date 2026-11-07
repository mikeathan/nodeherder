<script setup lang="ts">
  /*
   * Interactive room canvas: draws the outline, sensor cone, overlays, heat
   * cells, trails and targets, and translates pointer gestures into millimetre
   * edits. Projection/snapping math lives in domain/spatial/canvas.ts.
   */
  import { computed, ref } from 'vue';
  import type { SpatialPoint, SpatialWorkspace } from '@/types/spatial';
  import type { SpatialDisplayTrack } from '@/domain/spatial/tracks';
  import type { SpatialTarget } from '@/domain/spatial';
  import type { HeatCell } from '@/domain/spatial/analytics';
  import type { ZoneOccupancy } from '@/domain/spatial/occupancy';
  import {
    CANVAS_HANDLE_RADIUS_RATIO,
    CANVAS_HEADING_MIN_SPEED_MM_S,
    CANVAS_HEADING_SECONDS,
    CANVAS_OBJECT_FONT_RATIO,
    CANVAS_OBJECT_LABEL_OFFSET_RATIO,
    CANVAS_SENSOR_FONT_RATIO,
    CANVAS_SENSOR_LABEL_X_RATIO,
    CANVAS_SENSOR_LABEL_Y_RATIO,
    CANVAS_SENSOR_RADIUS_RATIO,
    CANVAS_TARGET_FONT_RATIO,
    CANVAS_TARGET_LABEL_X_RATIO,
    CANVAS_TARGET_LABEL_Y_RATIO,
    CANVAS_TARGET_RADIUS_RATIO,
    CANVAS_ZONE_FONT_RATIO,
    CANVAS_ZONE_LABEL_OFFSET_RATIO,
    HEAT_CELL_OPACITY,
    fovPolygon,
    heatCellColor,
    heatCellPolygon,
    objectPolygon,
    roomPolygon,
    sensorPolygon,
    sensorProjectedPoint,
    snapDelta,
    snapPoint,
  } from '@/domain/spatial/canvas';
  import { screenPoint, sensorPoint } from '@/domain/spatial/workspace';
  import { radarGridOffsets, radarViewBox, rotatePoint } from '@/domain/spatial/radar';
  import { SpatialViewIds, isSpatialEditView, type SpatialViewId } from '@/domain/spatial/view';
  import { ICONS } from '@/components/ui/icons';

  const props = defineProps<{
    workspace: SpatialWorkspace;
    /** Display tracks with stable identity; the caller owns filtering and hold state. */
    tracks: readonly SpatialDisplayTrack[];
    /** Hysteresis-tracked zone occupancy; never a firmware claim. */
    occupancy: ReadonlyMap<string, ZoneOccupancy>;
    scale: number;
    orientation: number;
    online: boolean;
    mode: SpatialViewId;
    selectedId: string;
    trails: { id: number; targets: SpatialTarget[] }[];
    cells: HeatCell[];
    showGrid: boolean;
    showRoomContext: boolean;
    /** Trailing path segment drawn brighter than `trails`; omitted by live surfaces. */
    recentTrails?: { id: number; targets: SpatialTarget[] }[];
    /** Connectors between the previous and current sample, explained as a jump. */
    hopLinks?: { from: SpatialPoint; to: SpatialPoint }[];
    /** Timestamp of the current sample; a new value restarts the pulse ring. */
    pulseAt?: number;
    /** Divides labels and markers so a zoomed viewBox keeps them readable. */
    uiScale?: number;
    /** Text halo width in view units; the zoomed Positions surface scales it with uiScale. */
    labelStroke?: number;
    /** Hide the illustrative field-of-view cone without changing the saved angle. */
    showFov?: boolean;
    /** Display-only FOV angle, used when the device has no configured angle. */
    fovAngle?: number;
    /** Panning suspends editing so the drag cannot also move a zone or handle. */
    panMode?: boolean;
    /** Explicit view box for auto-fitted surfaces; the panel range selector omits it. */
    viewBox?: string;
    /** Draw unfiltered source positions instead of filtered ones (evidence view). */
    showRaw?: boolean;
    label?: string;
    /** Preview surfaces hide per-slot labels because a slot is not an identity. */
    showTargetLabels?: boolean;
    /** Preview dots need to stay legible in a small fitted view. */
    targetScale?: number;
  }>();
  const emit = defineEmits<{
    select: [id: string];
    gestureStart: [index: number | null];
    gestureEnd: [];
    move: [delta: SpatialPoint];
    vertex: [index: number, point: SpatialPoint];
  }>();

  const svg = ref<SVGSVGElement>();
  const displayViewBox = computed(() => props.viewBox ?? radarViewBox(props.scale));
  const offsets = computed(() => radarGridOffsets(props.scale));
  const selected = computed(() => props.workspace.zones.find((zone) => zone.id === props.selectedId));
  // Positions zoom shrinks the viewBox; dividing symbols keeps labels and dots legible.
  const symbolScale = computed(() => props.uiScale ?? 1);
  const labelStrokeWidth = computed(() => props.labelStroke ?? 8);
  // Device marker: fixed on-screen size, with the presence glyph inside a filled disc.
  const deviceRadius = computed(
    () => (props.scale * CANVAS_SENSOR_RADIUS_RATIO) / symbolScale.value
  );
  const deviceIcon = ICONS.presence;
  const sensor = computed(() => screenPoint(props.workspace.sensor, props.orientation));
  /** Clamps the device to the visible edge so zooming never makes it vanish silently. */
  const devicePoint = computed(() => {
    const [x, y, width, height] = displayViewBox.value.split(' ').map(Number);
    const margin = Math.min(width, height) * 0.06;
    const point = {
      x: Math.min(Math.max(sensor.value.x, x + margin), x + width - margin),
      y: Math.min(Math.max(sensor.value.y, y + margin), y + height - margin),
    };
    return { ...point, offscreen: point.x !== sensor.value.x || point.y !== sensor.value.y };
  });
  const fov = computed(() =>
    fovPolygon(
      props.scale,
      props.fovAngle === undefined
        ? props.workspace.sensor
        : { ...props.workspace.sensor, fov: props.fovAngle },
      props.orientation
    )
  );
  const editable = computed(() => isSpatialEditView(props.mode) && !props.panMode);
  const heatmap = computed(() => props.mode === SpatialViewIds.Heatmap);
  const canvasLabel = computed(
    () => props.label ?? 'Live room canvas with sensor, zones and target positions. Numeric editing controls and target list follow.'
  );
  // Heatmap keeps only room context (outline, furniture, sensor) by default;
  // zones, targets, trails and handles would obscure the activity cells.
  const contextVisible = computed(() => !heatmap.value || props.showRoomContext);
  const objects = computed(() => (contextVisible.value ? props.workspace.objects : []));
  const zones = computed(() => (heatmap.value ? [] : props.workspace.zones));
  const tracks = computed(() => (heatmap.value ? [] : props.tracks));
  const trails = computed(() => (heatmap.value ? [] : props.trails));
  // Optional Position-player cues; live surfaces pass none, so nothing changes for them.
  const recentTrails = computed(() => (heatmap.value ? [] : props.recentTrails ?? []));
  const hopLinks = computed(() =>
    heatmap.value
      ? []
      : (props.hopLinks ?? []).map((link) => ({
          from: projected(link.from),
          to: projected(link.to),
        }))
  );

  function projected(point: SpatialPoint) {
    return sensorProjectedPoint(point, props.workspace.sensor, props.orientation);
  }
  /** Source position for a track in the mode being shown. */
  function trackPoint(track: SpatialDisplayTrack): SpatialPoint {
    return props.showRaw ? { x: track.rawX, y: track.rawY } : { x: track.x, y: track.y };
  }
  /**
   * Per-frame render model. Projection and the heading vector are computed once per track so
   * the template stays free of repeated geometry work.
   */
  const renderedTracks = computed(() =>
    tracks.value.map((track) => {
      const origin = trackPoint(track);
      const point = projected(origin);
      let heading: SpatialPoint | null = null;
      if (track.state === 'live' && Math.hypot(track.vx, track.vy) >= CANVAS_HEADING_MIN_SPEED_MM_S) {
        const ahead = projected({
          x: origin.x + track.vx * CANVAS_HEADING_SECONDS,
          y: origin.y + track.vy * CANVAS_HEADING_SECONDS,
        });
        heading = { x: ahead.x - point.x, y: ahead.y - point.y };
      }
      return { key: track.id, label: track.label, state: track.state, colorIndex: track.colorIndex, at: track.at, point, heading };
    })
  );
  function zonePoints(points: readonly SpatialPoint[]) {
    return sensorPolygon(points, props.workspace.sensor, props.orientation);
  }
  function trailPoints(points: readonly SpatialPoint[]) {
    return zonePoints(points);
  }
  function heatPoints(cell: HeatCell) {
    return heatCellPolygon(cell, props.workspace.sensor, props.orientation);
  }
  function objectPoints(object: SpatialWorkspace['objects'][number]) {
    return objectPolygon(object, props.orientation);
  }
  function outlinePoints(points: readonly SpatialPoint[]) {
    return roomPolygon(points, props.orientation);
  }

  let drag: { point: SpatialPoint; vertex: number | null; room: boolean } | null = null;

  /** Converts a pointer event into sensor or room millimetres through the SVG matrix. */
  function pointerPoint(event: PointerEvent, room = false): SpatialPoint {
    const matrix = svg.value?.getScreenCTM();
    if (!matrix) return { x: 0, y: 0 };
    const point = new DOMPoint(event.clientX, event.clientY).matrixTransform(matrix.inverse());
    if (room) return rotatePoint(point.x, -point.y, -props.orientation);
    return sensorPoint(point, props.workspace.sensor, props.orientation);
  }

  function start(event: PointerEvent, id: string, vertex: number | null = null, room = false) {
    if (!editable.value || event.button !== 0) return;
    emit('select', id);
    emit('gestureStart', vertex);
    drag = { point: pointerPoint(event, room), vertex, room };
    svg.value?.setPointerCapture(event.pointerId);
    event.preventDefault();
  }

  function end() {
    drag = null;
    emit('gestureEnd');
  }

  function move(event: PointerEvent) {
    if (!drag) return;
    const point = snapPoint(pointerPoint(event, drag.room));
    if (drag.vertex !== null) emit('vertex', drag.vertex, point);
    else emit('move', snapDelta(drag.point, point));
    drag.point = point;
  }
</script>
<template>
  <svg
    ref="svg"
    class="nh-spatial-canvas"
    :class="{ 'is-editable': editable }"
    :style="{ '--nh-label-stroke': labelStrokeWidth }"
    :viewBox="displayViewBox"
    role="img"
    :aria-label="canvasLabel"
    @pointermove="move"
    @pointerup="end"
    @pointercancel="end"
    @lostpointercapture="end">
    <g v-if="showGrid" class="nh-grid">
      <template v-for="offset in offsets" :key="offset">
        <line :x1="-scale" :x2="scale" :y1="offset" :y2="offset" />
        <line :x1="offset" :x2="offset" :y1="-scale" :y2="scale" />
      </template>
    </g>
    <polygon v-if="contextVisible && workspace.outline.length" :points="outlinePoints(workspace.outline)" class="nh-outline" />
    <g
      v-for="item in objects"
      :key="item.id"
      @pointerdown="mode === SpatialViewIds.Room && start($event, item.id, null, true)">
      <polygon :points="objectPoints(item)" :class="['nh-object', item.kind, { 'is-selected': selectedId === item.id, 'is-context': heatmap }]" />
      <text
        :x="screenPoint(item, orientation).x"
        :y="screenPoint(item, orientation).y + scale * CANVAS_OBJECT_LABEL_OFFSET_RATIO / symbolScale"
        :font-size="scale * CANVAS_OBJECT_FONT_RATIO / symbolScale">
        {{ item.name }}
      </text>
    </g>
    <polygon
      v-for="cell in cells"
      :key="`${cell.x}:${cell.y}`"
      :points="heatPoints(cell)"
      :fill="heatCellColor(cell.intensity)"
      :opacity="HEAT_CELL_OPACITY">
      <title>{{ cell.count }} target-second samples</title>
    </polygon>
    <!-- Above heat cells so the illustrative cone is visible on the heatmap as well. -->
    <polygon v-if="contextVisible && showFov !== false && fov" :points="fov" class="nh-fov" />
    <g v-for="zone in zones" :key="zone.id" @pointerdown="mode === SpatialViewIds.Zones && start($event, zone.id)">
      <polygon
        :points="zonePoints(zone.points)"
        :class="[
          'nh-zone',
          zone.kind,
          {
            'is-selected': selectedId === zone.id,
            'is-occupied': occupancy.get(zone.id)?.occupied ?? false,
          },
        ]" />
      <text
        :x="projected(zone.points[0]).x"
        :y="projected(zone.points[0]).y - scale * CANVAS_ZONE_LABEL_OFFSET_RATIO / symbolScale"
        :font-size="scale * CANVAS_ZONE_FONT_RATIO / symbolScale">
        {{ zone.name }} · {{ zone.kind }}
      </text>
    </g>
    <polyline v-for="(trail, index) in trails" :key="index" :points="trailPoints(trail.targets)" class="nh-trail" />
    <polyline v-for="(trail, index) in recentTrails" :key="`recent-${index}`" :points="trailPoints(trail.targets)" class="nh-trail is-recent" />
    <line
      v-for="(link, index) in hopLinks"
      :key="`hop-${index}`"
      :x1="link.from.x"
      :y1="link.from.y"
      :x2="link.to.x"
      :y2="link.to.y"
      class="nh-hop" />
    <g v-for="track in renderedTracks" :key="track.key" :transform="`translate(${track.point.x} ${track.point.y})`">
      <circle
        v-if="pulseAt && track.at === pulseAt"
        :key="`pulse-${pulseAt}`"
        :r="scale * CANVAS_TARGET_RADIUS_RATIO * (targetScale ?? 1) / symbolScale"
        class="nh-pulse" />
      <line v-if="track.heading" :x1="0" :y1="0" :x2="track.heading.x" :y2="track.heading.y" class="nh-heading" />
      <circle
        :r="scale * CANVAS_TARGET_RADIUS_RATIO * (targetScale ?? 1) / symbolScale"
        :class="['nh-target', `is-${track.state}`, `is-color-${track.colorIndex}`, { 'is-raw': showRaw }]" />
      <text v-if="showTargetLabels !== false" :x="scale * CANVAS_TARGET_LABEL_X_RATIO / symbolScale" :y="scale * CANVAS_TARGET_LABEL_Y_RATIO / symbolScale" :font-size="scale * CANVAS_TARGET_FONT_RATIO / symbolScale">
        {{ track.label }}
      </text>
    </g>
    <template v-if="contextVisible">
      <g
        :transform="`translate(${devicePoint.x} ${devicePoint.y})`"
        :class="['nh-device', { 'is-offscreen': devicePoint.offscreen }]">
        <title>{{ devicePoint.offscreen ? 'Sensor (off-screen)' : 'Sensor' }}</title>
        <circle :r="deviceRadius" class="nh-device-bg" />
        <g :transform="`scale(${deviceRadius / 15}) translate(-12 -12)`">
          <path :d="deviceIcon" class="nh-device-icon" />
        </g>
      </g>
      <text
        v-if="!devicePoint.offscreen"
        :x="devicePoint.x + scale * CANVAS_SENSOR_LABEL_X_RATIO / symbolScale"
        :y="devicePoint.y + scale * CANVAS_SENSOR_LABEL_Y_RATIO / symbolScale"
        :font-size="scale * CANVAS_SENSOR_FONT_RATIO / symbolScale">
        Sensor
      </text>
    </template>
    <g v-if="mode === SpatialViewIds.Zones && selected">
      <circle
        v-for="(corner, index) in selected.points"
        :key="index"
        :cx="projected(corner).x"
        :cy="projected(corner).y"
        :r="scale * CANVAS_HANDLE_RADIUS_RATIO"
        class="nh-handle"
        @pointerdown.stop="start($event, selected.id, index)">
        <title>Corner {{ index + 1 }}; also editable in numeric fields</title>
      </circle>
    </g>
  </svg>
</template>
<style scoped>
  .nh-spatial-canvas {
    display: block;
    width: 100%;
    aspect-ratio: 1;
    max-height: 70vh;
    background: var(--nh-surface-2);
    border: 1px solid var(--nh-border);
    border-radius: var(--nh-radius-m);
  }
  .is-editable {
    touch-action: none;
  }
  polygon,
  polyline,
  line {
    vector-effect: non-scaling-stroke;
  }
  .nh-grid {
    stroke: var(--nh-border);
    stroke-width: 1;
  }
  .nh-outline {
    fill: var(--nh-surface);
    fill-opacity: 0.45;
    stroke: var(--nh-text-muted);
    stroke-width: 3;
  }
  .nh-fov {
    fill: var(--nh-accent);
    fill-opacity: 0.08;
    stroke: var(--nh-accent);
    stroke-dasharray: 5 5;
  }
  .nh-object {
    fill: var(--nh-border-strong);
    fill-opacity: 0.35;
    stroke: var(--nh-text-muted);
    stroke-width: 2;
  }
  .door {
    stroke-dasharray: 5 3;
  }
  .wall {
    fill: var(--nh-text-muted);
  }
  .nh-object.is-context {
    fill-opacity: 0.08;
    stroke-width: 1.5;
  }
  .nh-zone {
    fill: var(--nh-spatial-detection);
    stroke: var(--nh-spatial-detection);
    fill-opacity: 0.12;
    stroke-width: 2;
  }
  .exclusion {
    fill: var(--nh-spatial-exclusion);
    stroke: var(--nh-spatial-exclusion);
    stroke-dasharray: 5 3;
  }
  .entry {
    fill: var(--nh-spatial-entry);
    stroke: var(--nh-spatial-entry);
    stroke-dasharray: 8 3;
  }
  .is-occupied {
    fill-opacity: 0.3;
    stroke-width: 4;
  }
  .is-selected {
    stroke: var(--nh-text);
    stroke-width: 4;
  }
  .nh-target {
    fill: var(--nh-accent);
  }
  .nh-target.is-color-1 {
    fill: var(--nh-spatial-detection);
  }
  .nh-target.is-color-2 {
    fill: var(--nh-spatial-entry);
  }
  .nh-target.is-color-3 {
    fill: var(--nh-spatial-exclusion);
  }
  .nh-target.is-held {
    fill-opacity: 0.45;
    stroke: var(--nh-text-muted);
    stroke-dasharray: 4 3;
    stroke-width: 6;
  }
  .nh-target.is-idle {
    fill-opacity: 0.55;
  }
  .nh-target.is-predicted {
    fill-opacity: 0.6;
    stroke: var(--nh-text);
    stroke-width: 6;
  }
  .nh-target.is-raw {
    fill-opacity: 0.35;
    stroke: var(--nh-text-muted);
    stroke-width: 4;
    stroke-dasharray: 2 4;
  }
  .nh-heading {
    stroke: var(--nh-text);
    stroke-opacity: 0.55;
    stroke-width: 6;
    stroke-linecap: round;
    pointer-events: none;
  }
  .nh-device {
    pointer-events: none;
  }
  .nh-device-bg {
    fill: var(--nh-text);
  }
  .nh-device-icon {
    fill: var(--nh-surface);
  }
  .nh-device.is-offscreen .nh-device-bg {
    fill: var(--nh-surface-2);
  }
  .nh-device.is-offscreen .nh-device-icon {
    fill: var(--nh-text-muted);
  }
  text {
    fill: var(--nh-text);
    paint-order: stroke;
    stroke: var(--nh-surface);
    stroke-width: var(--nh-label-stroke, 8);
    pointer-events: none;
  }
  .nh-trail {
    fill: none;
    stroke: var(--nh-accent);
    stroke-opacity: 0.35;
    stroke-width: 3;
    pointer-events: none;
  }
  .nh-trail.is-recent {
    stroke-opacity: 0.95;
    stroke-width: 6;
  }
  .nh-hop {
    stroke: var(--nh-text-muted);
    stroke-width: 5;
    stroke-dasharray: 14 10;
    stroke-opacity: 0.75;
    pointer-events: none;
  }
  .nh-pulse {
    fill: none;
    stroke: var(--nh-accent);
    stroke-width: 4;
    pointer-events: none;
    transform-box: fill-box;
    transform-origin: center;
    animation: nh-target-pulse 900ms ease-out;
  }
  @keyframes nh-target-pulse {
    from {
      opacity: 0.9;
      transform: scale(1);
    }
    to {
      opacity: 0;
      transform: scale(2.6);
    }
  }
  @media (prefers-reduced-motion: reduce) {
    .nh-pulse {
      animation: none;
      opacity: 0.5;
    }
  }
  .nh-handle {
    fill: var(--nh-surface);
    stroke: var(--nh-accent);
    stroke-width: 15;
    cursor: crosshair;
  }
</style>
