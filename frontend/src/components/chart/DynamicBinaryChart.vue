<script setup lang="ts">
  import { computed, ref, onMounted, onUnmounted, type PropType } from 'vue';
  // NOTE: onMounted/onUnmounted kept for the outside-click dismiss listener
  import type { DeviceExposeBinaryMetrics, BinaryBucket } from '@/types/metrics.type';
  import { getExposeBinaryColour } from '@/contracts/chart';
  import {
    getBinaryRanges,
    getBinaryStats,
    renderRangeTooltip,
    renderHeatmapTooltip,
    resolveBinaryLabel,
    bucketBinaryEvents,
    BINARY_NOISE_THRESHOLD,
    DENSITY_COLORS,
    getDensityTimeLabels,
    computeTimelineSegments,
  } from '@/utils/chart.utils';
  import { parseTimestamp } from '@/utils/date.utils';
  import { formatDuration } from '@/utils/date.utils';
  import { getSensorName } from '@/modules/formatters/sensor-formatter';

  type ChartMode = 'timeline' | 'heatmap';
  const DENSITY_BUCKET_MINUTES = 15;

  const props = defineProps({
    chartData: {
      type: Object as PropType<DeviceExposeBinaryMetrics>,
      required: true,
    },
    chartType: {
      type: String as PropType<ChartMode>,
      default: undefined,
    },
  });

  // ── Derived state ──

  const range = computed(() => {
    const now = Date.now();
    return {
      from: parseTimestamp(props.chartData?.from, now - 86400000),
      to: parseTimestamp(props.chartData?.to, now),
    };
  });

  const colors = computed(() => getExposeBinaryColour(props.chartData?.name ?? ''));
  const label = computed(() => getSensorName(props.chartData?.name ?? 'State'));
  const activeLabel = computed(() => resolveBinaryLabel(props.chartData?.name ?? '', true));

  const stats = computed(() => getBinaryStats(props.chartData?.data ?? [], range.value.to));

  const activeDurationStr = computed(() => formatDuration(stats.value.onDurationMs));
  const totalTriggers = computed(() => stats.value.onCount + stats.value.offCount);

  // ── Smart switch logic ──

  const resolvedMode = computed<ChartMode>(() => {
    if (props.chartType) return props.chartType;
    const eventCount = props.chartData?.data?.length ?? 0;
    return eventCount >= BINARY_NOISE_THRESHOLD ? 'heatmap' : 'timeline';
  });

  // ── Timeline (State A — low noise) ──

  const mergedRanges = computed(() => {
    if (!props.chartData?.data) return [];
    // No flicker merging — the MIN_WIDTH_PCT algorithm handles visual compaction
    return getBinaryRanges(props.chartData.data, range.value.from, range.value.to);
  });

  const timelineSegments = computed(() => {
    const ranges = mergedRanges.value;
    const totalMs = range.value.to - range.value.from;

    return computeTimelineSegments(
      ranges,
      totalMs,
      2.5, // MIN_WIDTH_PCT
      {
        colorOn: colors.value.on,
        colorOff: colors.value.off,
        exposeName: props.chartData?.name ?? '',
      }
    );
  });

  // ── Heatmap (State B — high noise) ──

  const heatmapBuckets = computed<BinaryBucket[]>(() => {
    if (!props.chartData?.data) return [];
    return bucketBinaryEvents(props.chartData.data, range.value.from, range.value.to, DENSITY_BUCKET_MINUTES);
  });

  const densitySegments = computed(() => {
    return heatmapBuckets.value.map((b) => {
      let color: string = DENSITY_COLORS.idle;
      if (b.count >= 8) color = DENSITY_COLORS.high;
      else if (b.count >= 4) color = DENSITY_COLORS.medium;
      else if (b.count >= 1) color = DENSITY_COLORS.low;
      return { color, count: b.count, start: b.start, bucket: b };
    });
  });

  const densityTimeLabels = computed(() => getDensityTimeLabels(heatmapBuckets.value));

  // ── Tooltip state ──

  const tooltipHtml = ref('');
  const tooltipVisible = ref(false);
  const tooltipPos = ref({ x: 0, y: 0 });

  type TooltipSeg = {
    bucket?: { start: number; end: number; count: number; activeMs: number; intensity: number };
    start?: number;
    end?: number;
    stateLabel?: string;
    color?: string;
  };

  function onCellClick(event: MouseEvent, seg: TooltipSeg) {
    const html = seg.bucket
      ? renderHeatmapTooltip(seg.bucket as BinaryBucket, props.chartData?.name ?? 'State')
      : renderRangeTooltip(
          props.chartData?.name ?? 'State',
          seg.stateLabel ?? '',
          seg.start ?? 0,
          seg.end ?? 0,
          seg.color ?? colors.value.on
        );

    // Toggle off if tapping the same cell
    if (tooltipVisible.value && tooltipHtml.value === html) {
      tooltipVisible.value = false;
      return;
    }

    tooltipHtml.value = html;
    tooltipVisible.value = true;
    tooltipPos.value = { x: event.clientX, y: event.clientY };
  }

  // Dismiss tooltip when clicking outside the strip
  function onOutsideClick(e: Event) {
    const target = e.target as HTMLElement;
    if (!target.closest('.density-cell')) {
      tooltipVisible.value = false;
    }
  }
  onMounted(() => document.addEventListener('click', onOutsideClick));
  onUnmounted(() => document.removeEventListener('click', onOutsideClick));
</script>

<template>
  <div v-if="chartData?.data?.length" class="binary-detail-chart">
    <!-- Stats header -->
    <div class="chart-header">
      <div class="chart-title">
        <span class="chart-dot" :style="{ backgroundColor: colors.on }"></span>
        <span class="chart-name">{{ label }}</span>
        <span class="chart-mode-badge">{{ resolvedMode === 'heatmap' ? 'Density' : 'Timeline' }}</span>
      </div>
      <div class="binary-stats">
        <div class="stat">
          <span class="stat-label">{{ activeLabel }}</span>
          <span class="stat-value">{{ stats.onPercentage.toFixed(0) }}%</span>
        </div>
        <div class="stat">
          <span class="stat-label">Time Active</span>
          <span class="stat-value">{{ activeDurationStr }}</span>
        </div>
        <div class="stat">
          <span class="stat-label">Triggers</span>
          <span class="stat-value">{{ totalTriggers }}</span>
        </div>
      </div>
    </div>

    <!-- Chart body — switches based on resolvedMode -->
    <div class="chart-body">
      <!-- State A: Timeline (Custom Strip) -->
      <div v-if="resolvedMode === 'timeline'" class="density-strip-container">
        <!-- Re-use density classes for the flex container to match layout perfectly -->
        <div class="density-strip timeline-strip" :style="{ gap: timelineSegments.length > 100 ? '0px' : '1px' }">
          <div
            v-for="(seg, i) in timelineSegments"
            :key="i"
            class="density-cell"
            :style="{ width: seg.width, backgroundColor: seg.color }"
            @click="onCellClick($event, seg)" />
        </div>
        <!-- Align to density strip time labels to guarantee X-Axis alignment -->
        <div class="density-time-labels">
          <span
            v-for="(lbl, i) in densityTimeLabels"
            :key="i"
            class="density-time-label"
            :style="{ left: lbl.offset + '%' }">
            {{ lbl.text }}
          </span>
        </div>
      </div>

      <!-- State B: Heatmap (Density Strip) -->
      <div v-else class="density-strip-container">
        <div class="density-strip" :style="{ height: '160px', gap: densitySegments.length > 100 ? '0px' : '1px' }">
          <div
            v-for="(seg, i) in densitySegments"
            :key="i"
            class="density-cell"
            :style="{ backgroundColor: seg.color }"
            @click="onCellClick($event, seg)" />
        </div>
        <div class="density-time-labels">
          <span
            v-for="(lbl, i) in densityTimeLabels"
            :key="i"
            class="density-time-label"
            :style="{ left: lbl.offset + '%' }">
            {{ lbl.text }}
          </span>
        </div>
      </div>

      <!-- Floating tooltip (Global to both modes) -->
      <Teleport to="body">
        <div
          v-if="tooltipVisible"
          class="density-tooltip"
          :style="{ top: tooltipPos.y - 8 + 'px', left: tooltipPos.x + 'px' }"
          v-html="tooltipHtml" />
      </Teleport>
    </div>
  </div>
</template>

<style scoped>
  :deep(.apexcharts-yaxis text),
  :deep(.apexcharts-yaxis-label) {
    display: none;
  }

  :deep(.apexcharts-rangebar-area) {
    transition: opacity 0.15s ease;
  }

  :deep(.apexcharts-rangebar-area:hover) {
    opacity: 0.85;
  }

  :deep(.apexcharts-tooltip) {
    transform: translateY(-46px);
  }

  .binary-detail-chart {
    padding: 8px 0 18px;
  }

  .chart-body {
    position: relative;
  }

  .chart-header {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: 16px;
    padding: 6px 0 10px;
    flex-wrap: wrap;
  }

  .chart-title {
    display: flex;
    align-items: center;
    gap: 8px;
    font-weight: 600;
    color: var(--p-surface-0, #e5e7eb);
  }

  .chart-dot {
    width: 10px;
    height: 10px;
    border-radius: 999px;
    display: inline-block;
    box-shadow: 0 0 12px rgba(0, 0, 0, 0.35);
  }

  .chart-name {
    font-size: 1rem;
  }

  .chart-mode-badge {
    font-size: 0.6rem;
    font-weight: 600;
    letter-spacing: 0.06em;
    text-transform: uppercase;
    padding: 2px 6px;
    border-radius: 4px;
    background: rgba(148, 163, 184, 0.15);
    color: var(--p-surface-400, #94a3b8);
  }

  .binary-stats {
    display: flex;
    gap: 10px;
  }

  .stat {
    display: flex;
    flex-direction: column;
    gap: 2px;
    padding: 6px 10px;
    border-radius: 8px;
    background: rgba(15, 23, 42, 0.35);
    min-width: 88px;
  }

  .stat-label {
    font-size: 0.65rem;
    letter-spacing: 0.08em;
    text-transform: uppercase;
    color: var(--p-surface-400, #8a94a6);
  }

  .stat-value {
    font-size: 0.9rem;
    color: var(--p-surface-0, #f8fafc);
    font-weight: 600;
  }

  @media (max-width: 640px) {
    .binary-stats {
      display: grid;
      grid-template-columns: repeat(3, minmax(0, 1fr));
      gap: 6px;
    }

    .stat {
      padding: 6px 6px;
      min-width: 0;
    }

    .stat-label {
      font-size: 0.55rem;
      letter-spacing: 0.06em;
    }

    .stat-value {
      font-size: 0.8rem;
    }
  }

  /* ── Light theme overrides ── */
  :root[data-p-theme='light'] .binary-detail-chart .chart-title,
  .p-light .binary-detail-chart .chart-title {
    color: var(--p-surface-700, #334155);
  }

  :root[data-p-theme='light'] .binary-detail-chart .stat,
  .p-light .binary-detail-chart .stat {
    background: rgba(0, 0, 0, 0.05);
  }

  :root[data-p-theme='light'] .binary-detail-chart .stat-value,
  .p-light .binary-detail-chart .stat-value {
    color: var(--p-surface-700, #334155);
  }

  /* ── Density strip styles ── */

  .density-strip-container {
    position: relative;
    padding-bottom: 24px;
    padding-left: 6px;
    padding-right: 6px;
  }

  .density-strip {
    display: flex;
    gap: 1px;
    border-radius: 4px;
    overflow: hidden;
    width: 100%;
  }

  .density-cell {
    /* Set min-width to 0 so flex basis can shrink cells below 1px on small screens 
       when there are many buckets (e.g. 30 days) */
    min-width: 0;
    height: 100%;
    transition: opacity 0.15s ease;
    cursor: default;
  }

  /* Make sure Heatmap mode cells divide equally */
  .density-strip:not(.timeline-strip) .density-cell {
    flex: 1;
  }

  .timeline-strip {
    /* For timeline we want exact width percentages, not flex 1 equal boxes */
    display: flex;
    flex-direction: row;
    height: auto;
    padding: 18px 0 10px;
  }

  /* Make sure the timeline bar has a sleek inner height */
  .timeline-strip .density-cell {
    height: 60px;
    align-self: center;
    /* min-width removed: handled dynamically via JS percentages to prevent mobile flex overflow */
  }

  .density-cell:hover {
    opacity: 0.8;
  }

  .density-time-labels {
    position: relative;
    height: 18px;
    margin-top: 8px;
  }

  .density-time-label {
    position: absolute;
    transform: translateX(-50%);
    font-size: 11px;
    color: var(--p-surface-400, #94a3b8);
    white-space: nowrap;
  }
</style>

<style>
  /* Tooltip must be unscoped since it's teleported to body */
  .density-tooltip {
    position: fixed;
    transform: translate(-50%, -100%);
    z-index: 9999;
    pointer-events: none;
    filter: drop-shadow(0 2px 8px rgba(0, 0, 0, 0.4));
  }
</style>
