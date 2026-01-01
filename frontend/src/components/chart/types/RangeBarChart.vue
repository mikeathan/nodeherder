<script setup lang="ts">
  import { computed } from 'vue';
  import type { PropType } from 'vue';
  import BaseChart from '../BaseChart.vue';
  import type { DeviceExposeBinaryMetrics } from '@/types/metrics.type';
  import { getExposeBinaryColour, resolveChartOptions } from '@/contracts/chart';
  import { ChartTypes } from '@/types/chart.type';
  import {
    getBinaryRanges,
    getBinaryStats,
    toBinaryRangeBarData,
    renderRangeTooltip,
    resolveBinaryLabel,
  } from '@/utils/chart.utils';
  import { parseTimestamp } from '@/utils/date.utils';
  import { getSensorName } from '@/modules/formatters/sensor-formatter';

  const props = defineProps({
    chartData: {
      type: Object as PropType<DeviceExposeBinaryMetrics>,
      required: true,
    },
  });

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

  const chartOffsetX = -52;

  const mergedRanges = computed(() => {
    if (!props.chartData?.data) return [];
    const rangeMs = Math.max(0, range.value.to - range.value.from);
    const flickerThreshold = Math.max(15_000, Math.round(rangeMs / 1200));
    return getBinaryRanges(props.chartData.data, range.value.from, range.value.to, flickerThreshold);
  });

  const series = computed<ApexAxisChartSeries>(() => {
    if (!props.chartData?.data) return [];

    const { name } = props.chartData;
    const ranges = mergedRanges.value;
    if (ranges.length === 0) return [];
    const apexDataRaw = toBinaryRangeBarData(ranges, colors.value.on, colors.value.off);
    const apexData = apexDataRaw.map((d) => ({
      ...d,
      x: name ?? 'State',
      stateLabel: resolveBinaryLabel(name ?? '', d.x === 'On' ? 'true' : 'false'),
    }));

    return [
      {
        name: name ?? 'State',
        data: apexData,
      },
    ];
  });


  const options = computed(() => {
    return resolveChartOptions(ChartTypes.BinaryChart, {
      chart: {
        offsetX: chartOffsetX,
      },
      plotOptions: {
        bar: {
          barHeight: '28%',
          borderRadius: 4,
          dataLabels: {
            position: 'center',
            hideOverflowingLabels: true,
          },
        },
      },
      stroke: {
        width: 1.5,
        colors: ['rgba(15, 23, 42, 0.75)'],
      },
      fill: {
        opacity: 1,
      },
      xaxis: {
        labels: {
          style: { colors: '#a3abb8', fontSize: '11px' },
        },
        axisBorder: { show: false },
        axisTicks: { show: false },
      },
      yaxis: {
        show: false,
        labels: { show: false },
        minWidth: 0,
        maxWidth: 0,
        axisBorder: { show: false },
        axisTicks: { show: false },
      },
      grid: {
        borderColor: 'rgba(148, 163, 184, 0.12)',
        strokeDashArray: 2,
        padding: {
          left: 0,
          right: 0,
        },
      },
      dataLabels: { enabled: false },
      tooltip: {
        theme: 'dark',
        x: { format: 'dd MMM HH:mm' },
        followCursor: true,
        custom: ({ w, seriesIndex, dataPointIndex }: { w: any; seriesIndex: number; dataPointIndex: number }) => {
          const d = w.config.series[seriesIndex].data[dataPointIndex];
          const start: number = Array.isArray(d.y) ? d.y[0] : d.y?.from ?? d.y ?? 0;
          const end: number = Array.isArray(d.y) ? d.y[1] : d.y?.to ?? d.y ?? 0;
          const label: string = d.stateLabel ?? (typeof d.x === 'string' ? d.x : d.x?.toString?.() ?? '');
          return renderRangeTooltip(props.chartData.name ?? 'State', label, start, end, d.fillColor);
        },
      },
    });
  });
</script>

<template>
  <div v-if="series.length > 0" class="binary-chart">
    <div class="chart-header">
      <div class="chart-title">
        <span class="chart-dot" :style="{ backgroundColor: colors.on }"></span>
        <span class="chart-name">{{ label }}</span>
      </div>
      <div class="binary-stats">
        <div class="stat">
          <span class="stat-label">{{ activeLabel }}</span>
          <span class="stat-value">{{ stats.onPercentage.toFixed(0) }}%</span>
        </div>
        <div class="stat">
          <span class="stat-label">Changes</span>
          <span class="stat-value">{{ stats.onCount + stats.offCount }}</span>
        </div>
      </div>
    </div>
    <div class="chart-body">
      <BaseChart :options="options" :data="series" :width="'100%'" :height="180" />
    </div>
  </div>
</template>

<style scoped>
  :deep(.apexcharts-yaxis text),
  :deep(.apexcharts-yaxis-label) {
    display: none;
  }

  :deep(.apexcharts-rangebar-area) {
    transition: filter 0.15s ease;
  }

  :deep(.apexcharts-rangebar-area:hover) {
    filter: drop-shadow(0 0 6px rgba(226, 232, 240, 0.45));
  }

  :deep(.apexcharts-tooltip) {
    transform: translateY(-58px);
  }

  .binary-chart {
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
    color: #e5e7eb;
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
    color: #8a94a6;
  }

  .stat-value {
    font-size: 0.9rem;
    color: #f8fafc;
    font-weight: 600;
  }

  @media (max-width: 640px) {
    .binary-stats {
      display: grid;
      grid-template-columns: repeat(2, minmax(0, 1fr));
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
</style>
