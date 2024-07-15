<script setup lang="ts">
import { ref, onBeforeMount } from 'vue';
import type { PropType, Ref } from 'vue';
import { Line, Bar, HorizontalBar } from 'vue-chartjs';
import { Chart as ChartJS, Title, Tooltip, Legend, LineElement, PointElement, LinearScale, CategoryScale, BarElement } from 'chart.js';

import 'chartjs-adapter-date-fns';


var data = {
    datasets: [{
        label: 'My Dataset',
        data: [
            { x: '2024-07-10T00:00:00', y: 20 },
            { x: '2024-07-12T12:30:00', y: 35 },
            { x: '2024-07-14T18:00:00', y: 10 },
        ],
        backgroundColor: 'rgba(75, 192, 192, 0.2)',
        borderColor: 'rgba(75, 192, 192, 1)',
    }],

};

const datasetoptions = {
    plugins: {
        legend: {
            position: 'right',
        },
    }
    , scales: {
        xAxes: [{
            type: 'time',
            time: {
                unit: 'day',
                tooltipFormat: 'YYYY-MM-DD HH:mm:ss' // Optional: format for hover tooltip
            }
        }],
        yAxes: [{
            stacked: true // Optional: stack bars on top of each other
        }]
    }
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
