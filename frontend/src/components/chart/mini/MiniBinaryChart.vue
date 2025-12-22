<script setup lang="ts">
  import { computed, PropType } from 'vue';
  import VueApexCharts from 'vue3-apexcharts';
  import { BinaryDataPoint } from '@/types/metrics.type';
  import { getExposeBinaryColour } from '@/contracts/chart';
  import {
    getBinaryRanges,
    getBinaryStats,
    toBinaryRangeBarData,
    renderRangeTooltip,
    resolveBinaryLabel,
  } from '@/utils/chart.utils';

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
      default: 150,
    },
  });

  const colors = computed(() => {
    const color = getExposeBinaryColour(props.exposeName);
    return [color.on, color.off];
  });

  // Compact range bar to keep binary states readable at small sizes.
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
  const stats = computed(() => {
    const result = getBinaryStats(props.data ?? []);
    return {
      onCount: result.onCount,
      offCount: result.offCount,
      onPercentage: result.onPercentage.toFixed(0),
    };
  });

  const chartOptions = computed(() => ({
    chart: {
      type: 'rangeBar',
      toolbar: { show: false },
      background: 'transparent',
    },
    plotOptions: {
      bar: {
        horizontal: true,
        barHeight: '85%',
        borderRadius: 8,
      },
    },
    stroke: {
      width: 2,
      colors: ['rgba(15, 23, 42, 0.9)'],
    },
    fill: {
      opacity: 1,
    },
    colors: [colors.value[0], colors.value[1]],
    grid: {
      borderColor: '#333',
      strokeDashArray: 3,
      yaxis: {
        lines: { show: true },
      },
      padding: {
        left: 10,
        right: 10,
      },
    },
    xaxis: {
      type: 'datetime',
      labels: {
        style: {
          colors: '#999',
          fontSize: '10px',
        },
        datetimeFormatter: {
          hour: 'HH:mm',
          minute: 'HH:mm',
        },
      },
      axisBorder: {
        show: false,
      },
      axisTicks: {
        show: false,
      },
    },
    yaxis: {
      show: false,
    },
    tooltip: {
      enabled: true,
      theme: 'dark',
      followCursor: true,
      custom: ({ w, seriesIndex, dataPointIndex }: { w: any; seriesIndex: number; dataPointIndex: number }) => {
        const d = w.config.series[seriesIndex].data[dataPointIndex];
        const start: number = Array.isArray(d.y) ? d.y[0] : d.y?.from ?? d.y ?? 0;
        const end: number = Array.isArray(d.y) ? d.y[1] : d.y?.to ?? d.y ?? 0;
        const stateValue: string = d.stateValue ?? 'false';
        const label = resolveBinaryLabel(props.exposeName, stateValue);
        return renderRangeTooltip(props.exposeName, label, start, end, d.fillColor);
      },
    },
    legend: {
      show: false,
    },
    dataLabels: { enabled: false },
  }));
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
    <VueApexCharts :height="height" :options="chartOptions" :series="chartData" />
  </div>
</template>

<style scoped>
  .mini-binary-chart {
    width: 100%;
    padding: 0.5rem;
    background: rgba(0, 0, 0, 0.1);
    border-radius: 8px;
  }

  .chart-stats {
    display: flex;
    justify-content: space-around;
    margin-bottom: 0.5rem;
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
    color: #999;
    text-transform: uppercase;
  }

.stat-value {
  font-size: 14px;
  font-weight: 600;
  color: #fff;
}

:deep(.apexcharts-rangebar-area) {
  transition: filter 0.15s ease;
}

:deep(.apexcharts-rangebar-area:hover) {
  filter: drop-shadow(0 0 6px rgba(226, 232, 240, 0.45));
}

:deep(.apexcharts-tooltip) {
  transform: translateY(-52px);
}
</style>
