<script setup lang="ts">
  import { computed, PropType } from 'vue';
  import VueApexCharts from 'vue3-apexcharts';
  import { NumericDataPoint } from '@/types/metrics.type';
  import { getExposeColor } from '@/contracts/chart';
  import { downsampleNumericData, getNumericStats, normalizeNumericData } from '@/utils/chart.utils';

  const props = defineProps({
    data: {
      type: Array as PropType<NumericDataPoint[]>,
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
    color: {
      type: String,
      default: undefined,
    },
    unit: {
      type: String,
      default: '',
    },
  });

  const normalizedData = computed(() => normalizeNumericData(props.data ?? []));
  const downsampledData = computed(() => downsampleNumericData(normalizedData.value, 60));

  const chartData = computed(() => [
    {
      name: props.exposeName,
      data: downsampledData.value.map((point) => ({
        x: point.x,
        y: point.y,
      })),
    },
  ]);

  const lineColor = computed(() => {
    if (props.color) return props.color;
    return getExposeColor(props.exposeName);
  });

  const stats = computed(() => {
    const numericStats = getNumericStats(normalizedData.value);
    if (!numericStats) {
      return { min: 0, max: 0, last: 0 };
    }

    const points = normalizedData.value;
    const last = points.length ? points[points.length - 1]?.y ?? 0 : 0;
    return { min: numericStats.min, max: numericStats.max, last };
  });

  const chartOptions = computed(() => ({
    chart: {
      type: 'line',
      toolbar: { show: false },
      zoom: { enabled: false },
      background: 'transparent',
      sparkline: { enabled: true },
    },
    dataLabels: { enabled: false },
    stroke: {
      curve: 'smooth',
      width: 2,
    },
    colors: [lineColor.value],
    grid: {
      borderColor: '#333',
      strokeDashArray: 3,
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
      axisBorder: { show: false },
      axisTicks: { show: false },
    },
    yaxis: {
      labels: {
        style: {
          colors: '#999',
          fontSize: '10px',
        },
        formatter: (value: number) => value.toFixed(1) + (props.unit ? ` ${props.unit}` : ''),
      },
    },
    tooltip: {
      enabled: true,
      theme: 'dark',
      x: { format: 'dd MMM HH:mm' },
      y: {
        formatter: (value: number) => value.toFixed(1) + (props.unit ? ` ${props.unit}` : ''),
      },
    },
    legend: { show: false },
  }));
</script>

<template>
  <div class="mini-instant-chart">
    <div class="chart-stats">
      <div class="stat">
        <span class="stat-label">Min</span>
        <span class="stat-value">{{ stats.min.toFixed(1) }}{{ unit }}</span>
      </div>
      <div class="stat">
        <span class="stat-label">Now</span>
        <span class="stat-value">{{ stats.last.toFixed(1) }}{{ unit }}</span>
      </div>
      <div class="stat">
        <span class="stat-label">Max</span>
        <span class="stat-value">{{ stats.max.toFixed(1) }}{{ unit }}</span>
      </div>
    </div>
    <VueApexCharts :height="height" :options="chartOptions" :series="chartData" />
  </div>
</template>

<style scoped>
  .mini-instant-chart {
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
</style>
