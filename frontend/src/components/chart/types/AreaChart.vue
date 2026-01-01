<script setup lang="ts">
  import { computed } from 'vue';
  import type { PropType } from 'vue';
  import BaseChart from '../BaseChart.vue';
  import type { DeviceExposeNumericMetrics } from '@/types/metrics.type';
  import { getExposeColor, resolveChartOptions } from '@/contracts/chart';
  import { ChartTypes } from '@/types/chart.type';
  import { getFormattedSensorValueByName } from '@/modules/formatters/sensor-formatter';
  import {
    downsampleNumericData,
    formatNumericXAxisLabel,
    getNumericDomain,
    normalizeNumericData,
  } from '@/utils/chart.utils';
  import { parseTimestamp } from '@/utils/date.utils';

  const props = defineProps({
    chartData: {
      type: Object as PropType<DeviceExposeNumericMetrics>,
      required: true,
    },
  });

  const normalizedData = computed(() => normalizeNumericData(props.chartData?.data ?? []));
  const sampledData = computed(() => downsampleNumericData(normalizedData.value, 350));

  const series = computed<ApexAxisChartSeries>(() => {
    if (!sampledData.value.length) return [];

    return [
      {
        name: props.chartData.name,
        data: sampledData.value.map((point) => ({
          x: point.x,
          y: point.y,
        })),
      },
    ];
  });

  const options = computed(() => {
    const color = getExposeColor(props.chartData?.name ?? '');
    const dataPoints = normalizedData.value;
    const showMarkers = dataPoints.length <= 2;
    const domain = getNumericDomain(dataPoints);
    const firstPoint = dataPoints[0]?.x;
    const lastPoint = dataPoints[dataPoints.length - 1]?.x;
    const fromMs = parseTimestamp(props.chartData?.from, Number.isFinite(firstPoint) ? firstPoint : Date.now());
    const toMs = parseTimestamp(props.chartData?.to, Number.isFinite(lastPoint) ? lastPoint : fromMs);
    const rangeMs = Number.isFinite(fromMs) && Number.isFinite(toMs) ? Math.max(0, toMs - fromMs) : 0;
    const dayMs = 24 * 60 * 60 * 1000;
    const showDayLabels = rangeMs > dayMs;
    const rangeDays = showDayLabels ? Math.max(1, Math.round(rangeMs / dayMs)) : 0;
    const tickAmount = showDayLabels ? Math.min(6, rangeDays + 1) : 6;
    const axisMin = showDayLabels ? fromMs : undefined;
    const axisMax = showDayLabels ? toMs : undefined;
    const formatXAxis = (value: string | number, _timestamp: number, opts: { date?: number }) =>
      formatNumericXAxisLabel(value, showDayLabels, opts);

    return resolveChartOptions(ChartTypes.AreaChart, {
      stroke: { colors: [color], width: 2.2 },
      colors: [color],
      markers: {
        size: showMarkers ? 4 : 0,
        hover: { size: showMarkers ? 6 : 4 },
      },
      grid: {
        borderColor: 'rgba(255,255,255,0.08)',
        strokeDashArray: 4,
        padding: { left: 6, right: 6, top: 8, bottom: 0 },
      },
      xaxis: {
        tickAmount,
        min: axisMin,
        max: axisMax,
        labels: {
          style: { colors: '#b8c1cc' },
          formatter: formatXAxis,
        },
      },
      yaxis: {
        min: domain.min,
        max: domain.max,
        labels: {
          style: { colors: '#b8c1cc' },
          formatter: (value: number) => {
            return value !== null ? value.toFixed(1) : '';
          },
        },
      },
      fill: {
        gradient: {
          opacityFrom: 0.35,
          opacityTo: 0.03,
          stops: [0, 60, 100],
        },
      },
      legend: { show: false },
      tooltip: {
        x: { format: 'dd MMM HH:mm' },
        y: {
          formatter: (value: number) => getFormattedSensorValueByName(props.chartData?.name ?? '', value),
        },
      },
    });
  });
  const containerStyle = {
    height: 'clamp(200px, 30vw, 260px)',
  };
</script>

<template>
  <div v-if="series.length > 0" class="w-full" :style="containerStyle">
    <BaseChart :height="'100%'" :width="'100%'" :options="options" :data="series" />
  </div>
</template>
