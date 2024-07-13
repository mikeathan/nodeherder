<script setup lang="ts">
import { ref, onBeforeMount } from 'vue';
import type { PropType, Ref } from 'vue';
import { Line } from 'vue-chartjs';
import { Chart as ChartJS, Title, Tooltip, Legend, LineElement, PointElement, LinearScale, TimeScale } from 'chart.js';
import 'chartjs-adapter-date-fns';


import moment from 'moment';
import 'chartjs-adapter-moment';


const options = {
    responsive: true,
    maintainAspectRatio: true,
    scales: {
        x: {
            type: 'time',
            time: {
                unit: 'day',
            },
            title: {
                display: true,
                text: 'Date',
            },
        },
        y: {
            title: {
                display: true,
                text: 'Value',
            },
        },
    },
    plugins: {
        legend: {
            display: true,
        },
        title: {
            display: true,
            text: 'Timeline Chart',
        },
    },
}

onBeforeMount(() => {
    ChartJS.register(Title, Tooltip, Legend, LineElement, PointElement, LinearScale, TimeScale);

});
</script>

<template>
    <Line :options="timelineOptions" :data="props.chartData" />
</template>

<!-- 
// Bar
// Line -->
