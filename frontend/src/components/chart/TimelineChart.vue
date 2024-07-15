<script setup lang="ts">
import { ref, onBeforeMount } from 'vue';
import type { PropType, Ref } from 'vue';
import { Line, Bar } from 'vue-chartjs';
import { Chart as ChartJS, Title, Tooltip, Legend, LineElement, PointElement, LinearScale, CategoryScale, BarElement } from 'chart.js';

import 'chartjs-adapter-date-fns';


var data = {
    labels: ['Data 1', 'Data 2', 'Data 3'],
    datasets: [
        {
            label: 'My Dataset',
            data: [[-10, 5], [2, 8], [12, 15]], // Start and end values for floating bars

            backgroundColor: 'rgba(255, 99, 132, 0.2)',
            borderColor: 'rgba(255, 99, 132, 1)',
            borderWidth: 1,
        },
    ],

};

const datasetoptions = {
    indexAxis: 'y', // Set y-axis as the main axis
    scales: {
        x: {
            stacked: false, // Disable stacking for floating bars
            beginAtZero: true, // Ensure x-axis starts at zero
        },
    },
    plugins: {
        tooltip: {
            // Optional: customize tooltip position
            yAlign: 'bottom',
        },
    },
};

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
    ChartJS.register(Title, Tooltip, Legend, LineElement, PointElement, LinearScale, BarElement, CategoryScale);

});
</script>

<template>
    <Bar :options="datasetoptions" :data="data" />
</template>

<!-- 
// Bar
// Line -->
