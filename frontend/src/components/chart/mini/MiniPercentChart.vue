<script setup lang="ts">
  import { computed, PropType } from 'vue';
  import VueApexCharts from 'vue3-apexcharts';
  import { NumericDataPoint } from '@/types/metrics.type';
  import { getExposeColor } from '@/contracts/chart';
  import { getFormattedSensorValueByName, getSensorName } from '@/modules/formatters/sensor-formatter';
  import { downsampleNumericData, normalizeNumericData } from '@/utils/chart.utils';

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
  const downsampledData = computed(() => downsampleNumericData(normalizedData.value, 50));
  const lastValue = computed(() => {
    const points = normalizedData.value;
    if (!points.length) return 0;
    return points[points.length - 1]?.y ?? 0;
  });

  const percentValue = computed(() => Math.max(0, Math.min(100, lastValue.value)));

  const barColor = computed(() => {
    if (props.color) return props.color;
    return getExposeColor(props.exposeName);
  });

  const label = computed(() => getSensorName(props.exposeName));
  const formattedValue = computed(() => {
    const raw = getFormattedSensorValueByName(props.exposeName, percentValue.value, props.unit || '%');
    return raw.replace(/(\d)([a-zA-Z%])/g, '$1 $2');
  });

  const chartData = computed(() => [
    {
      name: props.exposeName,
      data: downsampledData.value.map((point) => ({
        x: point.x,
        y: point.y,
      })),
    },
  ]);

  const chartHeight = computed(() => Math.max(60, Math.min(110, Math.round(props.height * 0.55))));

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
    colors: [barColor.value],
    tooltip: {
      enabled: true,
      theme: 'dark',
      x: { format: 'dd MMM HH:mm' },
      y: {
        formatter: (value: number) => value.toFixed(0) + (props.unit ? ` ${props.unit}` : '%'),
      },
    },
    legend: { show: false },
  }));
</script>

<template>
  <div class="mini-percent-chart">
    <div class="percent-header">
      <div class="percent-label">{{ label }}</div>
      <div class="percent-value">{{ formattedValue }}</div>
    </div>
    <VueApexCharts :height="chartHeight" :options="chartOptions" :series="chartData" />
    <div class="percent-bar">
      <div class="percent-fill" :style="{ width: `${percentValue}%`, backgroundColor: barColor }"></div>
    </div>
  </div>
</template>

<style scoped>
  .mini-percent-chart {
    width: 100%;
    padding: 0.6rem 0.75rem 0.7rem;
    background: rgba(0, 0, 0, 0.1);
    border-radius: 8px;
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }

  .percent-header {
    display: flex;
    align-items: baseline;
    justify-content: space-between;
    gap: 0.5rem;
  }

  .percent-label {
    font-size: 12px;
    color: #cbd5e1;
  }

  .percent-value {
    font-size: 16px;
    font-weight: 600;
    color: #fff;
  }

  .percent-bar {
    width: 100%;
    height: 10px;
    background: rgba(148, 163, 184, 0.25);
    border-radius: 999px;
    overflow: hidden;
    margin-top: 0.2rem;
  }

  .percent-fill {
    height: 100%;
    border-radius: 999px;
    box-shadow: 0 0 8px rgba(255, 255, 255, 0.12);
    transition: width 0.2s ease;
  }
</style>
