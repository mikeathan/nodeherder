<script setup lang="ts">
import { computed } from 'vue';
import BaseChart from '../BaseChart.vue';
import { BinaryDataPoint } from '@/types/metrics.type';

const props = defineProps({
  chartData: { type: Object, required: true },
});


function normalize(events: BinaryDataPoint[], from: number, to: number) {
  if (!events.length) {
    return [{ value: 'false', start: from, end: to }];
  }

  const out = [];

  // First segment: chart start → first event
  out.push({
    value: events[0].value,
    start: from,
    end: events[0].timestamp,
  });

  // Middle segments
  for (let i = 1; i < events.length; i++) {
    const prev = events[i - 1];
    const curr = events[i];

    out.push({
      value: prev.value,
      start: prev.timestamp,
      end: curr.timestamp,
    });
  }

  // Last segment: last event → chart end
  const last = events[events.length - 1];
  out.push({
    value: last.value,
    start: last.timestamp,
    end: to,
  });

  return out;
}


function mergeFlickers(
  ranges: Array<{ value: string; start: number; end: number }>,
  minDurationMs = 20_000 // 20 seconds
) {
  if (ranges.length <= 1) return ranges;

  const merged = [];
  let prev = ranges[0];

  for (let i = 1; i < ranges.length; i++) {
    const curr = ranges[i];
    const duration = prev.end - prev.start;

    const sameState = prev.value === curr.value;
    const tooShort = duration < minDurationMs;

    if (sameState || tooShort) {
      // Merge into one segment
      prev = {
        value: curr.value,
        start: prev.start,
        end: curr.end,
      };
    } else {
      merged.push(prev);
      prev = curr;
    }
  }

  merged.push(prev);
  return merged;
}


function toApex(ranges: Array<{ value: string; start: number; end: number }>) {
  return ranges.map(r => ({
    x: r.value === 'true' ? 'On' : 'Off',
    y: [r.start, r.end],
    fillColor: r.value === 'true' ? '#4ade80' : '#94a3b8',
  }));
}


const series = computed(() => {
  const { data, from, to } = props.chartData;

  const ranges = normalize(data, from, to);
  const merged = mergeFlickers(ranges, 15_000); // REMOVE flickers < 15s
  const apex = toApex(merged);

  return [
    {
      name: props.chartData.name,
      data: apex,
    },
  ];
});


const options = computed(() => ({
  chart: {
    type: 'rangeBar',
    background: 'transparent',
    toolbar: { show: false },
    zoom: { enabled: false }, // Disable all zoom
  },

  plotOptions: {
    bar: {
      horizontal: true,
      barHeight: '70%',
      borderRadius: 6,
      rangeBarGroupRows: true,
    },
  },

  xaxis: {
    type: 'datetime',
    labels: {
      style: { colors: '#ccc', fontSize: '11px' },
    },
  },

  yaxis: {
    labels: {
      style: { colors: '#ccc', fontSize: '12px' },
      formatter: (value: string) => value, // Show "On" / "Off"
    },
  },

  tooltip: {
    theme: 'dark',
    x: { format: 'dd MMM HH:mm' },
  },

  grid: { borderColor: 'rgba(255,255,255,0.15)' },
  legend: { show: false },
}));
</script>

<template>
  <BaseChart :options="options" :data="series" width="100%" height="250" />
</template>
