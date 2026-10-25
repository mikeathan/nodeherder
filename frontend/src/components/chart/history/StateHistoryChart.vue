<script setup lang="ts">
  import { computed, onBeforeUnmount, onMounted, ref, useId, watch } from 'vue';
  import type { Expose } from '@/types/device';
  import type { StateSegment } from '@/domain/history/state';
  import { RENDER_LIMIT } from '@/domain/history/types';
  import { binaryOnValue, binaryOffValue, exposeLabel, formatExposeValue } from '@/domain/exposes';
  import {
    formatHistoryCoverage,
    formatHistoryMoment,
    formatHistoryPreciseDuration,
    formatHistorySharePercent,
  } from '@/domain/history/format';
  import {
    clipStateSegments,
    stateHistoryBins,
    type StateHistoryBin,
  } from '@/domain/history/state';
  import {
    isDenseStateChart,
    stateChartLayout,
    stateChartTicks,
    stateChartX,
    type StateAxisTick,
    type StateChartLayout,
  } from './stateLayout';
  import StateHistorySvg from './StateHistorySvg.vue';

  const props = defineProps<{
    segments: StateSegment[];
    expose: Expose;
    from: number;
    to: number;
    viewFrom: number;
    viewTo: number;
  }>();
  const emit = defineEmits<{ 'plot-width': [value: number] }>();

  const activeLabel = computed(() => formatExposeValue(props.expose, binaryOnValue(props.expose)));
  const inactiveLabel = computed(() => formatExposeValue(props.expose, binaryOffValue(props.expose)));
  const chartName = computed(() => `${exposeLabel(props.expose.name)} history`);

  const host = ref<HTMLDivElement | null>(null);
  const tooltip = ref<HTMLDivElement | null>(null);
  const crosshairX = ref<number | null>(null);
  const width = ref(0);
  const instanceId = useId().replace(/[^a-zA-Z0-9_-]/g, '');
  const titleId = `state-history-title-${instanceId}`;
  const descriptionId = `state-history-description-${instanceId}`;
  const tooltipId = `state-history-tooltip-${instanceId}`;
  const patternId = `state-history-unknown-${instanceId}`;

  const measureContext = typeof document === 'undefined' ? null : document.createElement('canvas').getContext('2d');
  function textWidth(text: string, fontSize: number): number {
    if (!measureContext || typeof document === 'undefined') return text.length * fontSize * 0.62;
    measureContext.font = `600 ${fontSize}px ${getComputedStyle(document.body).fontFamily}`;
    return measureContext.measureText(text).width;
  }

  const layout = computed<StateChartLayout>(() => {
    // Before the first measurement, lay out against a phone-sized plot instead of 1px.
    const measured = width.value || 320;
    const fontSize = measured < 560 ? 10 : 11;
    const labelWidth = Math.max(textWidth(activeLabel.value, fontSize), textWidth(inactiveLabel.value, fontSize));
    return stateChartLayout(measured, labelWidth);
  });
  const visible = computed(() => clipStateSegments(props.segments, props.viewFrom, props.viewTo));
  const dense = computed(() => isDenseStateChart(visible.value.length, layout.value.plotWidth));
  const binCount = computed(() => Math.max(1, Math.min(RENDER_LIMIT, Math.floor(layout.value.plotWidth))));
  const bins = computed((): StateHistoryBin[] =>
    dense.value ? stateHistoryBins(visible.value, props.viewFrom, props.viewTo, binCount.value) : []
  );
  const ticks = computed<StateAxisTick[]>(() =>
    stateChartTicks({
      from: props.viewFrom,
      to: props.viewTo,
      width: layout.value.width,
      padLeft: layout.value.padLeft,
      padRight: layout.value.padRight,
      plotWidth: layout.value.plotWidth,
      fontSize: layout.value.fontSize,
      measure: textWidth,
    })
  );

  let observer: ResizeObserver | null = null;
  let resizeFallback: (() => void) | null = null;
  let publishedPlotWidth = -1;
  function publishWidth() {
    const next = host.value?.clientWidth ?? 0;
    if (!next) return;
    const rounded = Math.round(next);
    if (rounded !== width.value) width.value = rounded;
    const plot = layout.value.plotWidth;
    if (plot === publishedPlotWidth) return;
    publishedPlotWidth = plot;
    emit('plot-width', plot);
  }
  function stopObserving() {
    observer?.disconnect();
    observer = null;
    if (resizeFallback) window.removeEventListener('resize', resizeFallback);
    resizeFallback = null;
  }
  function observe() {
    stopObserving();
    if (!host.value) {
      width.value = 0;
      return;
    }
    publishWidth();
    if (typeof ResizeObserver !== 'undefined') {
      observer = new ResizeObserver(publishWidth);
      observer.observe(host.value);
    } else {
      resizeFallback = publishWidth;
      window.addEventListener('resize', resizeFallback);
    }
  }
  onMounted(observe);
  onBeforeUnmount(stopObserving);

  const cursor = { segment: 0, bin: 0 };
  let pinned = false;
  let frame = 0;
  let pending: PointerEvent | null = null;

  function bounded(value: number, min: number, max: number) {
    return Math.min(Math.max(value, min), max);
  }
  function hideInspection() {
    pinned = false;
    tooltip.value?.classList.remove('is-visible');
    crosshairX.value = null;
  }
  function showInspection(text: string, atX: number, announce: boolean) {
    const node = tooltip.value;
    const root = host.value;
    if (!node || !root) return;
    node.textContent = text;
    node.setAttribute('aria-live', announce ? 'polite' : 'off');
    node.classList.add('is-visible');
    const tipWidth = node.offsetWidth || 220;
    const tipHeight = node.offsetHeight || 54;
    node.style.left = `${bounded(atX + 10, 4, Math.max(4, root.clientWidth - tipWidth - 4))}px`;
    node.style.top = `${bounded(layout.value.plotTop + 8 - tipHeight, 4, Math.max(4, layout.value.height - tipHeight - 4))}px`;
    crosshairX.value = atX;
  }
  function inspectSegment(index: number, announce = false) {
    if (!visible.value.length) return;
    cursor.segment = bounded(index, 0, visible.value.length - 1);
    const segment = visible.value[cursor.segment];
    const text = segment.active === null
      ? `Unknown · ${formatHistoryMoment(segment.from)} – ${formatHistoryMoment(segment.to)} · No recorded data`
      : `${segment.label} · ${formatHistoryMoment(segment.from)} – ${formatHistoryMoment(segment.to)} · ` +
        `${formatHistoryPreciseDuration(segment.to - segment.from)}`;
    const mid = segment.from + (segment.to - segment.from) / 2;
    showInspection(text, stateChartX(mid, props.viewFrom, props.viewTo, layout.value), announce);
  }
  function inspectBin(index: number, announce = false) {
    if (!bins.value.length) return;
    cursor.bin = bounded(index, 0, bins.value.length - 1);
    const bin = bins.value[cursor.bin];
    const share = formatHistorySharePercent(bin.percent);
    const coverage = formatHistoryCoverage(bin.known, bin.to - bin.from);
    const detail = share === null
      ? 'No recorded known time'
      : `${bin.transitions} transition${bin.transitions === 1 ? '' : 's'} · ` +
        `${formatHistoryPreciseDuration(bin.active)} active · ${coverage} known`;
    const label = share === null ? 'No data' : `${share} ${activeLabel.value}`;
    const atX = layout.value.padLeft + ((cursor.bin + 0.5) / bins.value.length) * layout.value.plotWidth;
    showInspection(
      `${label} · ${formatHistoryMoment(bin.from)} – ${formatHistoryMoment(bin.to)} · ${detail}`,
      atX,
      announce
    );
  }
  function segmentIndexAt(timestamp: number): number {
    const items = visible.value;
    let low = 0;
    let high = items.length;
    while (low < high) {
      const middle = (low + high) >>> 1;
      if (items[middle].to <= timestamp) low = middle + 1;
      else high = middle;
    }
    return Math.min(low, Math.max(0, items.length - 1));
  }
  function inspectAt(clientX: number) {
    const root = host.value;
    if (!root) return;
    const rect = root.getBoundingClientRect();
    const ratio = bounded((clientX - rect.left - layout.value.padLeft) / layout.value.plotWidth, 0, 1);
    if (dense.value) inspectBin(Math.floor(ratio * bins.value.length));
    else inspectSegment(segmentIndexAt(props.viewFrom + ratio * (props.viewTo - props.viewFrom)));
  }
  /** Pointer work is coalesced to one inspection per animation frame. */
  function queueInspection(event: PointerEvent, pin: boolean) {
    pinned = pin;
    pending = event;
    if (frame) return;
    frame = requestAnimationFrame(() => {
      frame = 0;
      const next = pending;
      pending = null;
      if (next) inspectAt(next.clientX);
    });
  }
  function onPointerMove(event: PointerEvent) {
    if (event.pointerType !== 'mouse' && event.buttons === 0) return;
    if (event.pointerType === 'mouse') pinned = false;
    queueInspection(event, false);
  }
  function onPointerDown(event: PointerEvent) {
    if (event.pointerType === 'mouse' && event.button !== 0) return;
    queueInspection(event, event.pointerType !== 'mouse');
  }
  function onPointerLeave() {
    if (!pinned) hideInspection();
  }
  function onFocus() {
    if (dense.value) inspectBin(cursor.bin, true);
    else inspectSegment(cursor.segment, true);
  }
  function onKeydown(event: KeyboardEvent) {
    if (event.key === 'Escape') {
      event.preventDefault();
      hideInspection();
      return;
    }
    if (event.key !== 'ArrowLeft' && event.key !== 'ArrowRight') return;
    event.preventDefault();
    const step = event.key === 'ArrowRight' ? 1 : -1;
    if (dense.value) inspectBin(cursor.bin + step, true);
    else inspectSegment(cursor.segment + step, true);
  }
  watch([() => props.from, () => props.to, () => props.viewFrom, () => props.viewTo], () => {
    cursor.segment = 0;
    cursor.bin = 0;
    hideInspection();
  });
  onBeforeUnmount(() => {
    if (frame) cancelAnimationFrame(frame);
    frame = 0;
    pending = null;
  });
</script>

<template>
  <div
    ref="host"
    class="state-chart"
    role="group"
    :aria-labelledby="titleId"
    :aria-describedby="descriptionId"
    :style="{ '--state-chart-height': `${layout.height}px` }"
    :data-state-mode="dense ? 'aggregated' : 'steps'"
    :data-bin-count="dense ? bins.length : 0"
    :data-view-from="viewFrom"
    :data-view-to="viewTo"
    :data-layout-width="layout.width"
    :data-plot-width="layout.plotWidth"
    :data-plot-top="layout.plotTop"
    :data-plot-bottom="layout.plotBottom">
    <span :id="titleId" class="sr-only">{{ chartName }}</span>
    <span :id="descriptionId" class="sr-only">
      {{ chartName }}. Use left and right arrow keys to inspect the visible range; Escape clears the readout.
    </span>
    <StateHistorySvg
      :layout="layout"
      :segments="visible"
      :bins="bins"
      :dense="dense"
      :ticks="ticks"
      :active-label="activeLabel"
      :inactive-label="inactiveLabel"
      :pattern-id="patternId"
      :crosshair-x="crosshairX"
      :from="viewFrom"
      :to="viewTo" />
    <div
      class="state-inspect-target"
      role="group"
      tabindex="0"
      :aria-label="`Inspect ${chartName}. Use left and right arrow keys to move between states or time bins; Escape clears the readout.`"
      :aria-describedby="tooltipId"
      aria-keyshortcuts="ArrowLeft ArrowRight Escape"
      :style="{
        left: `${layout.padLeft}px`,
        top: `${layout.plotTop}px`,
        width: `${layout.plotWidth}px`,
        height: `${layout.plotHeight}px`,
      }"
      @pointermove="onPointerMove"
      @pointerdown="onPointerDown"
      @pointerleave="onPointerLeave"
      @pointercancel="hideInspection"
      @focus="onFocus"
      @blur="hideInspection"
      @keydown="onKeydown" />
    <div ref="tooltip" :id="tooltipId" class="state-tooltip" role="status" aria-live="off" />
  </div>
</template>

<style scoped>
  .state-chart {
    position: relative;
    width: 100%;
    min-width: 0;
    height: var(--state-chart-height);
    overflow: hidden;
  }
  .state-inspect-target {
    position: absolute;
    z-index: 1;
    cursor: crosshair;
    background: transparent;
    outline: none;
    touch-action: auto;
  }
  .state-inspect-target:focus {
    outline: 2px solid var(--nh-focus);
    outline-offset: 1px;
  }
  .state-tooltip {
    position: absolute;
    z-index: 2;
    visibility: hidden;
    pointer-events: none;
    max-width: min(260px, calc(100% - 8px));
    padding: 0.45rem 0.6rem;
    border: 1px solid var(--nh-border-strong);
    border-radius: var(--nh-radius-s);
    background: var(--nh-surface-2);
    color: var(--nh-text);
    font-size: 0.75rem;
    line-height: 1.35;
    box-shadow: var(--nh-shadow-2);
    overflow-wrap: anywhere;
  }
  .state-tooltip.is-visible {
    visibility: visible;
  }
</style>
