<script setup lang="ts">
  /*
   * Interactive room canvas: draws the outline, sensor cone, overlays, heat
   * cells, trails and targets, and translates pointer gestures into millimetre
   * edits. Projection/snapping math lives in domain/spatial/canvas.ts.
   */
  import { computed, ref } from 'vue';
  import type { SpatialPoint, SpatialWorkspace } from '@/types/spatial';
  import type { SpatialTarget } from '@/domain/spatial';
  import type { HeatCell } from '@/domain/spatial/analytics';
  import {
    CANVAS_HANDLE_RADIUS_RATIO,
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
  import { screenPoint, sensorPoint, zoneIsOccupied } from '@/domain/spatial/workspace';
  import { radarGridOffsets, radarViewBox, rotatePoint } from '@/domain/spatial/radar';
  import { SpatialViewIds, isSpatialEditView, type SpatialViewId } from '@/domain/spatial/view';

  const props = defineProps<{
    workspace: SpatialWorkspace;
    targets: readonly SpatialTarget[];
    scale: number;
    orientation: number;
    online: boolean;
    mode: SpatialViewId;
    selectedId: string;
    trails: { id: number; targets: SpatialTarget[] }[];
    cells: HeatCell[];
    showGrid: boolean;
    showRoomContext: boolean;
  }>();
  const emit = defineEmits<{
    select: [id: string];
    gestureStart: [index: number | null];
    gestureEnd: [];
    move: [delta: SpatialPoint];
    vertex: [index: number, point: SpatialPoint];
  }>();

  const svg = ref<SVGSVGElement>();
  const offsets = computed(() => radarGridOffsets(props.scale));
  const selected = computed(() => props.workspace.zones.find((zone) => zone.id === props.selectedId));
  const sensor = computed(() => screenPoint(props.workspace.sensor, props.orientation));
  const fov = computed(() => fovPolygon(props.scale, props.workspace.sensor, props.orientation));
  const editable = computed(() => isSpatialEditView(props.mode));
  const heatmap = computed(() => props.mode === SpatialViewIds.Heatmap);
  // Heatmap keeps only room context (outline, furniture, sensor) by default;
  // zones, targets, trails and handles would obscure the activity cells.
  const contextVisible = computed(() => !heatmap.value || props.showRoomContext);
  const objects = computed(() => (contextVisible.value ? props.workspace.objects : []));
  const zones = computed(() => (heatmap.value ? [] : props.workspace.zones));
  const targets = computed(() => (heatmap.value ? [] : props.targets));
  const trails = computed(() => (heatmap.value ? [] : props.trails));

  function projected(point: SpatialPoint) {
    return sensorProjectedPoint(point, props.workspace.sensor, props.orientation);
  }
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
    :viewBox="radarViewBox(scale)"
    role="img"
    aria-label="Live room canvas with sensor, zones and target positions. Numeric editing controls and target list follow."
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
    <polygon v-if="contextVisible && !heatmap && fov" :points="fov" class="nh-fov" />
    <g
      v-for="item in objects"
      :key="item.id"
      @pointerdown="mode === SpatialViewIds.Room && start($event, item.id, null, true)">
      <polygon :points="objectPoints(item)" :class="['nh-object', item.kind, { 'is-selected': selectedId === item.id, 'is-context': heatmap }]" />
      <text
        :x="screenPoint(item, orientation).x"
        :y="screenPoint(item, orientation).y + scale * CANVAS_OBJECT_LABEL_OFFSET_RATIO"
        :font-size="scale * CANVAS_OBJECT_FONT_RATIO">
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
    <g v-for="zone in zones" :key="zone.id" @pointerdown="mode === SpatialViewIds.Zones && start($event, zone.id)">
      <polygon
        :points="zonePoints(zone.points)"
        :class="[
          'nh-zone',
          zone.kind,
          {
            'is-selected': selectedId === zone.id,
            'is-occupied': zoneIsOccupied(zone, workspace.zones, targets, online),
          },
        ]" />
      <text
        :x="projected(zone.points[0]).x"
        :y="projected(zone.points[0]).y - scale * CANVAS_ZONE_LABEL_OFFSET_RATIO"
        :font-size="scale * CANVAS_ZONE_FONT_RATIO">
        {{ zone.name }} · {{ zone.kind }}
      </text>
    </g>
    <polyline v-for="(trail, index) in trails" :key="index" :points="trailPoints(trail.targets)" class="nh-trail" />
    <g v-for="target in targets" :key="target.id" :transform="`translate(${projected(target).x} ${projected(target).y})`">
      <circle :r="scale * CANVAS_TARGET_RADIUS_RATIO" class="nh-target" />
      <text :x="scale * CANVAS_TARGET_LABEL_X_RATIO" :y="scale * CANVAS_TARGET_LABEL_Y_RATIO" :font-size="scale * CANVAS_TARGET_FONT_RATIO">
        {{ target.id }}
      </text>
    </g>
    <template v-if="contextVisible">
      <circle :cx="sensor.x" :cy="sensor.y" :r="scale * CANVAS_SENSOR_RADIUS_RATIO" class="nh-sensor" />
      <text
        :x="sensor.x + scale * CANVAS_SENSOR_LABEL_X_RATIO"
        :y="sensor.y + scale * CANVAS_SENSOR_LABEL_Y_RATIO"
        :font-size="scale * CANVAS_SENSOR_FONT_RATIO">
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
  .nh-sensor {
    fill: var(--nh-text);
  }
  text {
    fill: var(--nh-text);
    paint-order: stroke;
    stroke: var(--nh-surface);
    stroke-width: 8;
    pointer-events: none;
  }
  .nh-trail {
    fill: none;
    stroke: var(--nh-accent);
    stroke-opacity: 0.35;
    stroke-width: 3;
    pointer-events: none;
  }
  .nh-handle {
    fill: var(--nh-surface);
    stroke: var(--nh-accent);
    stroke-width: 15;
    cursor: crosshair;
  }
</style>
