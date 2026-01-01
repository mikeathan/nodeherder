<script setup lang="ts">
  import { computed, PropType } from 'vue';
  import { DeviceExposeNumericMetrics } from '@/types/metrics.type';
  import AreaChart from '../chart/types/AreaChart.vue';
  import NumericChartHeader from './NumericChartHeader.vue';
  import { getExposeColor } from '@/contracts/chart';
  import { getFormattedSensorValueByName, getSensorName, getSensorUnit } from '@/modules/formatters/sensor-formatter';
  import { getNumericStats, normalizeNumericData } from '@/utils/chart.utils';

  const props = defineProps({
    chartData: {
      type: Object as PropType<DeviceExposeNumericMetrics[]>,
      required: true,
    },
  });

  const MAX_STATS_POINTS = 2000;

  const charts = computed(() =>
    props.chartData
      .map((entry) => {
        const normalized = normalizeNumericData(entry.data ?? []);
        const stats = normalized.length <= MAX_STATS_POINTS ? getNumericStats(normalized) : null;

        return {
          ...entry,
          label: getSensorName(entry.name),
          unit: getSensorUnit(entry.name),
          color: getExposeColor(entry.name),
          stats,
        };
      })
      .sort((a, b) => a.label.localeCompare(b.label))
  );

  const formatValue = (name: string, value: number) => getFormattedSensorValueByName(name, value);
  const formatDelta = (name: string, delta: number, deltaPct: number) => {
    const sign = delta > 0 ? '+' : '';
    const base = `${sign}${formatValue(name, delta)}`;
    if (!Number.isFinite(deltaPct) || deltaPct === 0) {
      return base;
    }
    const pct = `${deltaPct > 0 ? '+' : ''}${deltaPct.toFixed(1)}%`;
    return `${base} (${pct})`;
  };
</script>

<template>
  <div v-for="chart in charts" :key="chart.name" class="numeric-chart">
    <DateRangeDisplay v-if="chart.from && chart.to" :from="chart.from" :to="chart.to" />
    <NumericChartHeader :chart="chart" :format-value="formatValue" :format-delta="formatDelta" />
    <AreaChart :chartData="chart"></AreaChart>
  </div>
</template>

<style scoped>
  .numeric-chart {
    padding: 8px 0 18px;
  }
</style>
