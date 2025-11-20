<script setup lang="ts">
  import { computed, PropType } from 'vue';
  import VueApexCharts from 'vue3-apexcharts';
  import { NumericDataPoint } from '@/types/metrics.type';
  import { getExposeColor } from '@/contracts/chart';

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

  // Downsample for cleaner visualization
  const downsampledData = computed(() => {
    if (!props.data || props.data.length === 0) return [];

    // Target ~50-60 points for detailed view
    const targetPoints = 50;
    const step = Math.max(1, Math.floor(props.data.length / targetPoints));

    const sampled: any[] = [];
    for (let i = 0; i < props.data.length; i += step) {
      sampled.push(props.data[i]);
    }

    // Always include last point
    if (sampled[sampled.length - 1] !== props.data[props.data.length - 1]) {
      sampled.push(props.data[props.data.length - 1]);
    }

    return sampled;
  });

  const chartData = computed(() => {
    return [
      {
        name: props.exposeName,
        data: downsampledData.value.map((point) => ({
          x: point.x,
          y: point.y,
        })),
      },
    ];
  });

  const lineColor = computed(() => {
    if (props.color) return props.color;
    return getExposeColor(props.exposeName);
  });

  const stats = computed(() => {
    if (!props.data || props.data.length === 0) {
      return { min: 0, max: 0, avg: 0 };
    }

    const values = props.data.map((p) => p.y);
    const min = Math.min(...values);
    const max = Math.max(...values);
    const avg = values.reduce((sum, val) => sum + val, 0) / values.length;

    return { min, max, avg };
  });

  const chartOptions = computed(() => ({
    chart: {
      type: 'area',
      toolbar: {
        show: false,
      },
      zoom: {
        enabled: false,
      },
      background: 'transparent',
    },
    dataLabels: {
      enabled: false,
    },
    stroke: {
      curve: 'smooth',
      width: 2.5,
    },
    fill: {
      type: 'gradient',
      gradient: {
        shadeIntensity: 1,
        opacityFrom: 0.5,
        opacityTo: 0.1,
        stops: [0, 90, 100],
      },
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
      axisBorder: {
        show: false,
      },
      axisTicks: {
        show: false,
      },
    },
    yaxis: {
      labels: {
        style: {
          colors: '#999',
          fontSize: '10px',
        },
        formatter: (value: number) => {
          return value.toFixed(1) + (props.unit ? ` ${props.unit}` : '');
        },
      },
    },
    tooltip: {
      enabled: true,
      theme: 'dark',
      x: {
        format: 'dd MMM HH:mm',
      },
      y: {
        formatter: (value: number) => {
          return value.toFixed(1) + (props.unit ? ` ${props.unit}` : '');
        },
      },
    },
    legend: {
      show: false,
    },
  }));
</script>

<template>
  <div class="mini-numeric-chart">
    <div class="chart-stats">
      <div class="stat">
        <span class="stat-label">Min</span>
        <span class="stat-value">{{ stats.min.toFixed(1) }}{{ unit }}</span>
      </div>
      <div class="stat">
        <span class="stat-label">Avg</span>
        <span class="stat-value">{{ stats.avg.toFixed(1) }}{{ unit }}</span>
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
  .mini-numeric-chart {
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
