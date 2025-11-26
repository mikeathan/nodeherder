<script setup lang="ts">
  import { computed, PropType } from 'vue';
  import VueApexCharts from 'vue3-apexcharts';
  import { BinaryDataPoint } from '@/types/metrics.type';
  import { getExposeBinaryColour } from '@/contracts/chart';

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

  // Transform binary data into step line (Home Assistant style)
  const chartData = computed(() => {
    if (!props.data || props.data.length === 0) return [];

    const points: Array<{ x: number; y: number }> = [];

    props.data.forEach((point) => {
      const value =
        point.value.toLowerCase() === 'true' || point.value.toLowerCase() === 'on' || point.value === '1' ? 1 : 0;

      points.push({
        x: point.timestamp,
        y: value,
      });
    });

    // Add current time with last value
    if (points.length > 0) {
      points.push({
        x: Date.now(),
        y: points[points.length - 1].y,
      });
    }

    return [
      {
        name: props.exposeName,
        data: points,
      },
    ];
  });
  const stats = computed(() => {
    if (!props.data || props.data.length === 0) {
      return { onCount: 0, offCount: 0, onPercentage: 0 };
    }

    let onDuration = 0;
    let offDuration = 0;

    for (let i = 0; i < props.data.length - 1; i++) {
      const current = props.data[i];
      const next = props.data[i + 1];
      const duration = next.timestamp - current.timestamp;

      if (current.value.toLowerCase() === 'on' || current.value.toLowerCase() === 'true') {
        onDuration += duration;
      } else {
        offDuration += duration;
      }
    }

    // Handle last data point
    if (props.data.length > 0) {
      const last = props.data[props.data.length - 1];
      const duration = Date.now() - last.timestamp;

      if (last.value.toLowerCase() === 'on' || last.value.toLowerCase() === 'true') {
        onDuration += duration;
      } else {
        offDuration += duration;
      }
    }

    const total = onDuration + offDuration;
    const onPercentage = total > 0 ? (onDuration / total) * 100 : 0;

    return {
      onCount: props.data.filter((d) => d.value.toLowerCase() === 'on' || d.value.toLowerCase() === 'true').length,
      offCount: props.data.filter((d) => d.value.toLowerCase() === 'off' || d.value.toLowerCase() === 'false').length,
      onPercentage: onPercentage.toFixed(0),
    };
  });

  const chartOptions = computed(() => ({
    chart: {
      type: 'area',
      toolbar: {
        show: false,
      },
      background: 'transparent',
    },
    stroke: {
      curve: 'stepline',
      width: 2.5,
    },
    fill: {
      type: 'gradient',
      gradient: {
        shadeIntensity: 1,
        opacityFrom: 0.6,
        opacityTo: 0.1,
        stops: [0, 100],
      },
    },
    colors: [colors.value[0]],
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
      min: 0,
      max: 1,
      show: false,
    },
    tooltip: {
      enabled: true,
      theme: 'dark',
      x: {
        format: 'dd MMM HH:mm',
      },
      y: {
        formatter: (value: number) => {
          return value === 1 ? 'On' : 'Off';
        },
      },
    },
    legend: {
      show: false,
    },
    markers: {
      size: 0,
    },
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
</style>
