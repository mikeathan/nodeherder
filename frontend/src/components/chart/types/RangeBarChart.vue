<script setup lang="ts">
import { computed } from 'vue';
import BaseChart from '../BaseChart.vue';
import { BinaryDataPoint } from '@/types/metrics.type';

const props = defineProps({
  chartData: { type: Object, required: true },
});

function eventsToRanges(events: BinaryDataPoint[]) {
  if (!events.length) return [];

  const ranges = [];
  let prev = events[0];

  for (let i = 1; i < events.length; i++) {
    const curr = events[i];
    ranges.push({
      x: prev.value === 'true' ? 'On' : 'Off',
      y: [prev.timestamp, curr.timestamp],
      fillColor: prev.value === 'true' ? '#4ade80' : '#94a3b8',
    });
    prev = curr;
  }

  ranges.push({
    x: prev.value === 'true' ? 'On' : 'Off',
    y: [prev.timestamp, props.chartData.to],
    fillColor: prev.value === 'true' ? '#4ade80' : '#94a3b8',
  });

  return ranges;
}

const series = computed(() => [
  { name: props.chartData.name, data: eventsToRanges(props.chartData.data) }
]);

const options = computed(() => ({
  chart: {
    type: 'rangeBar',
    background: 'transparent',
    toolbar: { show: false },
  },
  plotOptions: {
    bar: {
      horizontal: true,
      barHeight: '80%',
      rangeBarGroupRows: true,
    },
  },
  xaxis: {
    type: 'datetime',
  },
  yaxis: {
    labels: {
      style: { colors: '#ccc' }
    }
  },
  tooltip: {
    x: { format: "dd MMM HH:mm" }
  },
}));
</script>

<template>
  <BaseChart :options="options" :data="series" width="100%" height="250" />
</template>

