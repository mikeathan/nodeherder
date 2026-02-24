<script setup lang="ts">
  import { computed, ref, onMounted, onUnmounted, PropType } from 'vue';
  import VueApexCharts from 'vue3-apexcharts';
  import { BinaryDataPoint } from '@/types/metrics.type';
  import { getExposeBinaryColour, resolveChartOptions } from '@/contracts/chart';
  import { ChartTypes } from '@/types/chart.type';
  import {
    getBinaryRanges,
    getBinaryStats,
    toBinaryRangeBarData,
    renderRangeTooltip,
    resolveBinaryLabel,
    bucketBinaryEvents,
    BINARY_NOISE_THRESHOLD,
    DENSITY_COLORS,
    getDensityTimeLabels,
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

  const colors = computed(() => {
    const color = getExposeBinaryColour(props.exposeName);
    return [color.on, color.off];
  });

  const isNoisy = computed(() => (props.data?.length ?? 0) >= BINARY_NOISE_THRESHOLD);

  const stats = computed(() => {
    const result = getBinaryStats(props.data ?? []);
    return {
      onCount: result.onCount,
      offCount: result.offCount,
      onPercentage: result.onPercentage.toFixed(0),
    };
  });

  // ── Density mode (noisy data) ──

  const densityRange = computed(() => {
    if (!props.data?.length) return { from: 0, to: 0 };
    const timestamps = props.data.filter((p) => Number.isFinite(p.timestamp)).map((p) => p.timestamp);
    if (timestamps.length === 0) return { from: 0, to: 0 };
    return { from: Math.min(...timestamps), to: Math.max(...timestamps) };
  });

  const densityBuckets = computed(() => {
    if (!isNoisy.value || !props.data?.length) return [];
    const { from, to } = densityRange.value;
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

  // ── Tooltip state ──

  const tooltipHtml = ref('');
  const tooltipVisible = ref(false);
  const tooltipPos = ref({ x: 0, y: 0 });

  function onCellClick(
    event: MouseEvent,
    seg: { bucket: { start: number; end: number; count: number; activeMs: number; intensity: number } }
  ) {
    const label = resolveBinaryLabel(props.exposeName, true);
    const html = renderRangeTooltip(
      props.exposeName,
      `${label} (${seg.bucket.count} triggers)`,
      seg.bucket.start,
      seg.bucket.end,
      colors.value[0]
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

  // ── Timeline mode (clean data) ──

  const chartData = computed(() => {
    if (!props.data || props.data.length === 0) return [];

    const now = Date.now();
    const from = props.data.reduce((min, point) => {
      if (!Number.isFinite(point.timestamp)) return min;
      return Math.min(min, point.timestamp);
    }, Number.POSITIVE_INFINITY);

    if (!Number.isFinite(from)) return [];

    const ranges = getBinaryRanges(props.data, from, now);
    const apexDataRaw = toBinaryRangeBarData(ranges, colors.value[0], colors.value[1]);
    const apexData = apexDataRaw.map((d) => ({
      ...d,
      x: props.exposeName,
      stateValue: d.x === 'On' ? 'true' : 'false',
    }));

    return [
      {
        name: props.exposeName,
        data: apexData,
      },
    ];
  });

  const chartOptions = computed(() => {
    return resolveChartOptions(ChartTypes.BinaryChart, {
      chart: {
        sparkline: { enabled: false },
      },
      plotOptions: {
        bar: {
          barHeight: '100%',
          borderRadius: 0,
        },
      },
      stroke: { width: 0 },
      fill: { opacity: 1 },
      colors: [colors.value[0], colors.value[1]],
      grid: {
        show: false,
        padding: { left: 0, right: 0, top: -20, bottom: 0 },
      },
      xaxis: {
        labels: {
          style: { colors: 'var(--p-surface-400, #94a3b8)', fontSize: '10px' },
          datetimeFormatter: { hour: 'HH:mm', minute: 'HH:mm' },
        },
        axisBorder: { show: false },
        axisTicks: { show: false },
      },
      yaxis: { show: false },
      tooltip: {
        enabled: true,
        theme: 'dark',
        followCursor: true,
        custom: ({ w, seriesIndex, dataPointIndex }: { w: any; seriesIndex: number; dataPointIndex: number }) => {
          const d = w.config.series[seriesIndex].data[dataPointIndex];
          const start: number = Array.isArray(d.y) ? d.y[0] : (d.y?.from ?? d.y ?? 0);
          const end: number = Array.isArray(d.y) ? d.y[1] : (d.y?.to ?? d.y ?? 0);
          const stateValue: string = d.stateValue ?? 'false';
          const label = resolveBinaryLabel(props.exposeName, stateValue);
          return renderRangeTooltip(props.exposeName, label, start, end, d.fillColor);
        },
      },
      legend: { show: false },
      dataLabels: { enabled: false },
    });
  });
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
      <!-- Floating tooltip -->
      <Teleport to="body">
        <div
          v-if="tooltipVisible"
          class="density-tooltip"
          :style="{ top: tooltipPos.y - 8 + 'px', left: tooltipPos.x + 'px' }"
          v-html="tooltipHtml" />
      </Teleport>
    </div>

    <!-- Timeline for clean data -->
    <VueApexCharts v-else :height="height" :options="chartOptions" :series="chartData" />
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
    min-width: 1px;
    transition: opacity 0.15s ease;
    cursor: default;
  }

  .density-cell:hover {
    opacity: 0.8;
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

  /* ── Timeline styles ── */
  :deep(.apexcharts-rangebar-area) {
    transition: opacity 0.15s ease;
  }

  :deep(.apexcharts-rangebar-area:hover) {
    opacity: 0.85;
  }

  :deep(.apexcharts-tooltip) {
    transform: translateY(-40px);
  }

  :deep(.apexcharts-plot-area) {
    overflow: visible;
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
