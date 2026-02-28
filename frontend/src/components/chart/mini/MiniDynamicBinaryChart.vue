<script setup lang="ts">
  import { computed, ref, onMounted, onUnmounted, PropType } from 'vue';
  import { BinaryDataPoint } from '@/types/metrics.type';
  import { getExposeBinaryColour } from '@/contracts/chart';
  import {
    getBinaryRanges,
    getBinaryStats,
    renderRangeTooltip,
    resolveBinaryLabel,
    bucketBinaryEvents,
    BINARY_NOISE_THRESHOLD,
    DENSITY_COLORS,
    getDensityTimeLabels,
    computeTimelineSegments,
  } from '@/utils/chart.utils';

  const DENSITY_BUCKET_MINUTES = 5;

  const props = defineProps({
    data: {
      type: Array as PropType<BinaryDataPoint[]>,
      required: true,
    },
    exposeName: {
      type: String,
      required: true,
    },
    height: {
      type: Number,
      default: 80,
    },
  });

  const colors = computed(() => getExposeBinaryColour(props.exposeName));

  const isNoisy = computed(() => (props.data?.length ?? 0) >= BINARY_NOISE_THRESHOLD);

  // ── Range — computed from data timestamps ──

  const range = computed(() => {
    if (!props.data?.length) return { from: 0, to: 0 };
    const now = Date.now();
    const timestamps = props.data.filter((p) => Number.isFinite(p.timestamp)).map((p) => p.timestamp);
    if (timestamps.length === 0) return { from: 0, to: 0 };
    return { from: Math.min(...timestamps), to: now };
  });

  const stats = computed(() => {
    const result = getBinaryStats(props.data ?? [], range.value.to);
    return {
      onCount: result.onCount,
      offCount: result.offCount,
      onPercentage: result.onPercentage.toFixed(0),
    };
  });

  // ── Density mode (noisy data) ──

  const densityBuckets = computed(() => {
    if (!isNoisy.value || !props.data?.length) return [];
    const { from, to } = range.value;
    if (from >= to) return [];
    return bucketBinaryEvents(props.data, from, to, DENSITY_BUCKET_MINUTES);
  });

  const densitySegments = computed(() => {
    return densityBuckets.value.map((b) => {
      let color: string = DENSITY_COLORS.idle;
      if (b.count >= 8) color = DENSITY_COLORS.high;
      else if (b.count >= 4) color = DENSITY_COLORS.medium;
      else if (b.count >= 1) color = DENSITY_COLORS.low;
      return { color, count: b.count, start: b.start, bucket: b };
    });
  });

  const densityTimeLabels = computed(() => getDensityTimeLabels(densityBuckets.value));

  // ── Timeline mode (clean data) — custom strip with MIN_WIDTH_PCT ──

  const timelineSegments = computed(() => {
    if (!props.data?.length) return [];
    const { from, to } = range.value;
    const totalMs = to - from;
    if (totalMs <= 0) return [];

    // Filter out zero-width ranges that occur when from === first event timestamp
    const ranges = getBinaryRanges(props.data, from, to).filter((r) => r.end > r.start);

    return computeTimelineSegments(
      ranges,
      totalMs,
      3, // MIN_WIDTH_PCT
      {
        colorOn: colors.value.on,
        colorOff: colors.value.off,
        exposeName: props.exposeName,
      }
    );
  });

  const timelineTimeLabels = computed(() => {
    if (!props.data?.length) return [];
    const { from, to } = range.value;
    if (from >= to) return [];
    const ranges = getBinaryRanges(props.data, from, to).filter((r) => r.end > r.start);
    // Reuse density time label logic by creating a simple bucket array
    const buckets = ranges.map((r) => ({ start: r.start, end: r.end, count: 0, activeMs: 0, intensity: 0 }));
    return getDensityTimeLabels(buckets);
  });

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
      ? renderRangeTooltip(
          props.exposeName,
          `${resolveBinaryLabel(props.exposeName, true)} (${seg.bucket.count} triggers)`,
          seg.bucket.start,
          seg.bucket.end,
          colors.value.on
        )
      : renderRangeTooltip(
          props.exposeName,
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
    if (tooltipVisible.value) {
      const target = e.target as HTMLElement;
      if (!target.closest('.density-cell')) {
        tooltipVisible.value = false;
      }
    }
  }
  onMounted(() => document.addEventListener('click', onOutsideClick));
  onUnmounted(() => document.removeEventListener('click', onOutsideClick));
</script>

<template>
  <div class="mini-binary-chart">
    <div class="chart-stats">
      <div class="stat">
        <span class="stat-label">On Time</span>
        <span class="stat-value">{{ stats.onPercentage }}%</span>
      </div>
      <div class="stat">
        <span class="stat-label">Changes</span>
        <span class="stat-value">{{ stats.onCount + stats.offCount }}</span>
      </div>
    </div>

    <!-- Density strip for noisy data -->
    <div v-if="isNoisy" class="density-strip-container">
      <div class="density-strip" :style="{ height: height + 'px' }">
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

    <!-- Timeline strip for clean data -->
    <div v-else class="density-strip-container">
      <div class="density-strip timeline-strip" :style="{ height: height + 'px' }">
        <div
          v-for="(seg, i) in timelineSegments"
          :key="i"
          class="density-cell"
          :style="{ width: seg.width, backgroundColor: seg.color }"
          @click="onCellClick($event, seg)" />
      </div>
      <div class="density-time-labels">
        <span
          v-for="(lbl, i) in timelineTimeLabels"
          :key="i"
          class="density-time-label"
          :style="{ left: lbl.offset + '%' }">
          {{ lbl.text }}
        </span>
      </div>
    </div>

    <!-- Floating tooltip (shared by both modes) -->
    <Teleport to="body">
      <div
        v-if="tooltipVisible"
        class="density-tooltip"
        :style="{ top: tooltipPos.y - 8 + 'px', left: tooltipPos.x + 'px' }"
        v-html="tooltipHtml" />
    </Teleport>
  </div>
</template>

<style scoped>
  .mini-binary-chart {
    width: 100%;
    padding: 0.5rem;
    background: transparent;
    border-radius: 8px;
  }

  .chart-stats {
    display: flex;
    justify-content: space-around;
    margin-bottom: 0.25rem;
    gap: 0.5rem;
  }

  .stat {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 0.1rem;
  }

  .stat-label {
    font-size: 10px;
    color: var(--p-surface-400, #94a3b8);
    text-transform: uppercase;
  }

  .stat-value {
    font-size: 14px;
    font-weight: 600;
    color: var(--p-surface-0, #f1f5f9);
  }

  /* ── Density strip styles ── */

  .density-strip-container {
    position: relative;
    padding-bottom: 18px;
  }

  .density-strip {
    display: flex;
    gap: 1px;
    border-radius: 4px;
    overflow: hidden;
  }

  .density-cell {
    flex: 1;
    min-width: 0;
    transition: opacity 0.15s ease;
    cursor: default;
  }

  .density-cell:hover {
    opacity: 0.8;
  }

  /* ── Timeline strip overrides ── */

  .timeline-strip {
    display: flex;
    flex-direction: row;
  }

  .timeline-strip .density-cell {
    flex: none; /* Use explicit width percentages, not flex: 1 */
    min-width: 4px; /* Fallback minimum for clickability */
  }

  .density-time-labels {
    position: relative;
    height: 18px;
    margin-top: 4px;
  }

  .density-time-label {
    position: absolute;
    transform: translateX(-50%);
    font-size: 10px;
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
