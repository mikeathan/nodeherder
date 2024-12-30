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
        chartData.value = transformedChartData(
          props.chartData
        );
      }
    },
    { immediate: true }
  );

  function transformedChartData(
    chartData: DeviceExposeNumericMetrics
  ): AreaChartEntry[] {
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
    },
    xaxis: {
      type: 'datetime',
      labels: {
        datetimeFormatter: {
          year: 'yyyy',
          month: "MMM 'yy",
          day: 'dd MMM',
          hour: 'HH:mm',
        },
      },
    },
    Tooltip: {
      x: {
        format: 'dd/MMM/yy HH:mm:ss ',
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
    <BaseChart
      height="200"
      :data="chartData"
      :options="chartOptions" />
  </div>
</template>
