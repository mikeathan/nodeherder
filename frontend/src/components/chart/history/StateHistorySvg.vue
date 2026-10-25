<script setup lang="ts">
  import { computed } from 'vue';
  import { stateTransitions, type StateHistoryBin, type StateSegment } from '@/domain/history/state';
  import {
    stateChartX,
    stateLaneY,
    type StateAxisTick,
    type StateChartLayout,
  } from './stateLayout';

  /** Presentational SVG for the binary history chart; all interaction lives in the parent. */
  const props = defineProps<{
    layout: StateChartLayout;
    segments: StateSegment[];
    from: number;
    to: number;
    bins: StateHistoryBin[];
    dense: boolean;
    ticks: StateAxisTick[];
    activeLabel: string;
    inactiveLabel: string;
    patternId: string;
    crosshairX: number | null;
  }>();

  const known = computed(() => props.segments.filter((segment) => segment.active !== null));
  const inactive = computed(() => props.segments.filter((segment) => segment.active === false));
  const unknown = computed(() => props.segments.filter((segment) => segment.active === null));
  const transitions = computed(() => stateTransitions(props.segments));

  const x = (timestamp: number) => stateChartX(timestamp, props.from, props.to, props.layout);
  const y = (active: boolean) => stateLaneY(active, props.layout);
  const binWidth = computed(() => (props.bins.length ? props.layout.plotWidth / props.bins.length : props.layout.plotWidth));
</script>

<template>
  <svg
    class="state-chart-svg"
    aria-hidden="true"
    :viewBox="`0 0 ${layout.width} ${layout.height}`"
    preserveAspectRatio="none">
    <defs>
      <pattern :id="patternId" width="6" height="6" patternTransform="rotate(45)" patternUnits="userSpaceOnUse">
        <rect width="6" height="6" fill="var(--nh-surface-2)" />
        <line x1="0" y1="0" x2="0" y2="6" stroke="var(--nh-border-strong)" stroke-width="2" />
      </pattern>
    </defs>
    <g class="state-grid">
      <template v-for="tick in ticks" :key="tick.timestamp">
        <line :x1="tick.x" :x2="tick.x" :y1="layout.plotTop" :y2="layout.plotBottom" />
        <text
          :x="tick.x"
          :y="layout.height - 8"
          :text-anchor="tick.anchor"
          :font-size="layout.narrow ? 10 : 10.5">{{ tick.label }}</text>
      </template>
      <line
        :x1="layout.padLeft"
        :x2="layout.width - layout.padRight"
        :y1="layout.activeY"
        :y2="layout.activeY"
        data-state-lane="active"
        stroke-dasharray="3 4" />
      <line
        :x1="layout.padLeft"
        :x2="layout.width - layout.padRight"
        :y1="layout.inactiveY"
        :y2="layout.inactiveY"
        data-state-lane="inactive"
        stroke-dasharray="3 4" />
    </g>
    <g class="state-lane-labels">
      <circle cx="7" :cy="layout.activeY" r="3.5" fill="var(--state-active)" />
      <text x="15" :y="layout.activeY + 3.5" :font-size="layout.fontSize">{{ activeLabel }}</text>
      <circle cx="7" :cy="layout.inactiveY" r="3.5" fill="var(--state-inactive)" />
      <text x="15" :y="layout.inactiveY + 3.5" :font-size="layout.fontSize">{{ inactiveLabel }}</text>
    </g>
    <g v-if="!dense" class="state-sparse-marks">
      <rect
        v-for="(segment, index) in inactive"
        :key="`inactive-fill-${index}`"
        :x="x(segment.from)"
        :y="layout.inactiveY"
        :width="Math.max(0.5, x(segment.to) - x(segment.from))"
        :height="layout.plotBottom - layout.inactiveY"
        fill="var(--state-inactive)"
        fill-opacity="0.14" />
      <line
        v-for="(segment, index) in known"
        :key="`state-run-${index}`"
        :x1="x(segment.from)"
        :x2="x(segment.to)"
        :y1="y(segment.active!)"
        :y2="y(segment.active!)"
        :stroke="segment.active ? 'var(--state-active)' : 'var(--state-inactive)'"
        stroke-width="2.5" />
      <line
        v-for="(change, index) in transitions"
        :key="`state-change-${index}`"
        :x1="x(change.from)"
        :x2="x(change.from)"
        :y1="y(change.before)"
        :y2="y(change.after)"
        :stroke="change.after ? 'var(--state-active)' : 'var(--state-inactive)'"
        stroke-width="2.5" />
      <circle
        v-for="(change, index) in transitions"
        :key="`state-marker-${index}`"
        :cx="x(change.from)"
        :cy="y(change.after)"
        :r="layout.narrow ? 2.5 : 3"
        :fill="change.after ? 'var(--state-active)' : 'var(--state-inactive)'"
        stroke="var(--nh-surface)"
        stroke-width="1.5" />
      <g v-for="(segment, index) in unknown" :key="`unknown-${index}`">
        <rect
          :x="x(segment.from)"
          :y="layout.plotTop"
          :width="Math.max(1, x(segment.to) - x(segment.from))"
          :height="layout.plotHeight"
          :fill="`url(#${patternId})`"
          data-unknown-band="true" />
        <text
          v-if="x(segment.to) - x(segment.from) > 64"
          :x="(x(segment.from) + x(segment.to)) / 2"
          :y="layout.plotTop + 13"
          text-anchor="middle"
          font-size="10">No data</text>
      </g>
    </g>
    <g v-else class="state-aggregated-marks">
      <line
        :x1="layout.padLeft"
        :x2="layout.width - layout.padRight"
        :y1="layout.inactiveY"
        :y2="layout.inactiveY"
        stroke="var(--state-inactive)"
        stroke-width="1.5" />
      <template v-for="(bin, index) in bins" :key="`bin-${index}`">
        <rect
          v-if="bin.percent !== null && bin.percent > 0"
          :x="layout.padLeft + index * binWidth"
          :y="layout.inactiveY - (bin.percent / 100) * (layout.inactiveY - layout.activeY)"
          :width="binWidth + 0.35"
          :height="(bin.percent / 100) * (layout.inactiveY - layout.activeY)"
          fill="var(--state-active)"
          :fill-opacity="0.25 + Math.min(1, bin.known / (bin.to - bin.from)) * 0.65"
          data-state-bin="true" />
        <rect
          v-if="bin.known < bin.to - bin.from - 1"
          :x="layout.padLeft + index * binWidth"
          :y="layout.plotTop"
          :width="binWidth + 0.35"
          :height="layout.plotHeight"
          :fill="`url(#${patternId})`"
          :fill-opacity="Math.max(0.12, 1 - bin.known / (bin.to - bin.from))"
          data-unknown-band="true" />
      </template>
    </g>
    <line
      v-if="crosshairX !== null"
      :x1="crosshairX"
      :x2="crosshairX"
      :y1="layout.plotTop"
      :y2="layout.plotBottom"
      stroke="var(--nh-text-muted)"
      stroke-width="1"
      stroke-dasharray="2 3"
      opacity="0.8" />
  </svg>
</template>

<style scoped>
  .state-chart-svg {
    display: block;
    width: 100%;
    height: 100%;
    overflow: hidden;
  }
  .state-grid line {
    stroke: var(--nh-border);
    stroke-opacity: 0.7;
  }
  .state-grid text,
  .state-lane-labels text,
  .state-sparse-marks text {
    fill: var(--nh-text-muted);
    font-family: var(--nh-font-ui);
  }
  .state-lane-labels text {
    fill: var(--nh-text);
    font-weight: 600;
  }
</style>
