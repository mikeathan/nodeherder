<script setup lang="ts">
import { ref, onBeforeMount } from 'vue';
import type { PropType, Ref } from 'vue';
import 'chartjs-adapter-date-fns';
import Highcharts from 'highcharts';

var data = {
  datasets: [
    {
      label: 'My Dataset',
      data: [
        { x: '2024-07-10T00:00:00', y: 20 },
        { x: '2024-07-12T12:30:00', y: 35 },
        { x: '2024-07-14T18:00:00', y: 10 },
      ],
      backgroundColor: 'rgba(75, 192, 192, 0.2)',
      borderColor: 'rgba(75, 192, 192, 1)',
    },
  ],
};
const chartOptions = ref({
  series: [
    {
      data: [1, 2, 3],
    },
  ],
});

const hcInstance = Highcharts;

onBeforeMount(() => {});
</script>

<template>
  <template>
    <div id="app">
      <highcharts
        :highcharts="hcInstance"
        :options="chartOptions"></highcharts>
    </div>
  </template>
</template>

<!-- 
// Bar
// Line -->
