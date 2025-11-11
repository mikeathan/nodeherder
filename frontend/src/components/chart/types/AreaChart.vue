<script setup lang="ts">
  import { ref, watch } from 'vue';
  import type { PropType, Ref } from 'vue';
  import BaseChart from '../BaseChart.vue';
  import { AreaChartEntry } from '@/types/chart.type';
  import { DeviceExposeNumericMetrics } from '@/types/metrics.type';
  import { getExposeColor } from '@/contracts/chart';

  const props = defineProps({
    chartData: {
      type: Object as PropType<DeviceExposeNumericMetrics>,
      default: null,
    },
  });

  const chartData = ref<AreaChartEntry[]>([]);

  watch(
    () => props.chartData,
    () => {
      if (props.chartData !== null) {
        chartData.value = transformedChartData(props.chartData);
      }
    },
    { immediate: true }
  );

  function transformedChartData(chartData: DeviceExposeNumericMetrics): AreaChartEntry[] {
    return [
      {
        name: chartData.name,
        color: getExposeColor(chartData.name),
        data: chartData.data.map((point) => ({
          x: point.x,
          y: point.y,
        })),
      },
    ] as AreaChartEntry[];
  }

  const chartOptions = {
    chart: {
      type: 'area',
      background: '#fff',
      toolbar: {
        autoselected: 'pan',
        theme: 'dark',
        show: false,
      },
      zoom: {
        enabled: true,
        type: 'x',
        autoScaleYaxis: true,
      },
    },
    fill: {
      type: 'gradient',
      gradient: {
        shadeIntensity: 1,
        inverseColors: false,
        opacityFrom: 0.6,
        opacityTo: 0,
        stops: [0, 100],
      },
    },
    dataLabels: {
      enabled: false,
    },
    legend: {
      showForSingleSeries: true,
      position: 'top',
    },
    stroke: {
      curve: 'smooth',
      width: 2,
    },
    xaxis: {
      type: 'datetime',
      labels: {
        datetimeUTC: false,
        datetimeFormatter: {
          year: 'yyyy',
          month: "MMM 'yy",
          day: 'dd MMM',
          hour: 'HH:mm',
          minute: 'HH:mm',
        },
        rotate: 0,
        rotateAlways: false,
        hideOverlappingLabels: true,
        trim: false,
        style: {
          fontSize: '11px',
        },
      },
      tooltip: {
        enabled: false,
      },
    },
    yaxis: {
      decimalsInFloat: 1,
      labels: {
        formatter: (value: number) => {
          return value !== null ? value.toFixed(1) : '';
        },
      },
    },
    tooltip: {
      x: {
        format: 'dd MMM yyyy HH:mm:ss',
        formatter: function (value: number) {
          const date = new Date(value);
          const now = new Date();
          const diffMs = now.getTime() - date.getTime();
          const diffMins = Math.floor(diffMs / 60000);
          const diffHours = Math.floor(diffMs / 3600000);
          const diffDays = Math.floor(diffMs / 86400000);

          // Format based on how old the data is
          const timeStr = date.toLocaleTimeString('en-GB', {
            hour: '2-digit',
            minute: '2-digit',
            second: '2-digit',
          });

          if (diffMins < 60) {
            return `${diffMins} min ago (${timeStr})`;
          } else if (diffHours < 24) {
            return `${diffHours}h ago (${timeStr})`;
          } else if (diffDays === 1) {
            return `Yesterday ${timeStr}`;
          } else if (diffDays < 7) {
            return `${diffDays} days ago (${timeStr})`;
          }

          const dateStr = date.toLocaleDateString('en-GB', {
            day: '2-digit',
            month: 'short',
            year: 'numeric',
          });
          return `${dateStr} ${timeStr}`;
        },
      },
      y: {
        formatter: (value: number) => {
          return value !== null ? value.toFixed(2) : '';
        },
      },
    },
    responsive: [
      {
        breakpoint: undefined,
        options: {
          chart: {
            width: '100%',
          },
        },
      },
    ],
  };
</script>

<template>
  <div class="area-chart">
    <BaseChart height="200" :data="chartData" :options="chartOptions" />
  </div>
</template>
