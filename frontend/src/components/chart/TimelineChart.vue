<script setup lang="ts">
import { ref, onBeforeMount } from 'vue';
import type { PropType, Ref } from 'vue';
import VueApexCharts from 'vue3-apexcharts';

const presenceData = ref([
  // Replace with your actual presence data
  { timestamp: '2024-07-17T09:00:00', value: 1 }, // On at 9:00
  { timestamp: '2024-07-17T12:00:00', value: 0 }, // Off at 12:00
  { timestamp: '2024-07-17T14:00:00', value: 1 }, // On at 14:00
  { timestamp: '2024-07-17T18:00:00', value: 0 }, // Off at 18:00
  { timestamp: '2024-07-17T20:00:00', value: 1 }, // On at 20:00
  { timestamp: '2024-07-17T20:05:00', value: 0 }, // Off at 20:05
  { timestamp: '2024-07-17T20:10:00', value: 1 }, // On at 20:10
  { timestamp: '2024-07-17T20:24:00', value: 0 }, // Off at 20:24
]);

function convertToApexTimelineRangebarData(
  data: { timestamp: Date; value: number }[],
): any[] {
  const apexData: any[] = [];

  // Group data by presence (value)
  const presenceGroups = data.reduce((acc, curr) => {
    const presence = curr.value;
    acc[presence] = acc[presence] || [];
    acc[presence].push(curr.timestamp);
    return acc;
  }, {});

  // Convert timestamps to epoch milliseconds for ApexCharts
  for (const presence in presenceGroups) {
    const timestamps = presenceGroups[presence].map(
      (timestamp) => timestamp.getTime(),
    );
    apexData.push({
      x: presence === '1' ? 'Present' : 'Absent', // Set labels based on presence value
      y: timestamps,
    });
  }

  return apexData;
}

const series = [
  {
    name: 'ON',
    data: [
      {
        x: 'Presence',
        y: [
          new Date('2024-07-17T09:00:00').getTime(),
          new Date('2024-07-17T12:00:00').getTime(),
        ],
      },
    ],
  },
  {
    name: 'OFF',
    data: [
      {
        x: 'Presence',
        y: [
          new Date('2024-07-17T12:00:00').getTime(),
          new Date('2024-07-17T14:00:00').getTime(),
        ],
      },
    ],
  },
];

const chartOptions = {
  chart: {
    height: 450,
    type: 'rangeBar',
  },
  plotOptions: {
    bar: {
      horizontal: true,
      barHeight: '20%',
      rangeBarGroupRows: true,
    },
    fill: {
      type: 'solid',
    },
    xaxis: {
      type: 'datetime',
    },
    legend: {
      position: 'right',
    },
  },
  xaxis: {
    type: 'datetime',
  },
  stroke: {
    width: 1,
  },
  fill: {
    type: 'solid',
    opacity: 0.6,
  },
  legend: {
    position: 'top',
    horizontalAlign: 'left',
  },
};

onBeforeMount(() => {});
</script>

<template>
  <div class="chart-container">
    <VueApexCharts
      width="800"
      height="400"
      :options="chartOptions"
      :series="series">
    </VueApexCharts>
  </div>
</template>

<!-- 
// Bar
// Line -->
