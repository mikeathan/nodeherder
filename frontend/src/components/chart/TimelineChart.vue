<script setup lang="ts">
import { ref, onBeforeMount } from 'vue';
import type { PropType, Ref } from 'vue';
import 'chartjs-adapter-date-fns';
import VueApexCharts from 'vue3-apexcharts'

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

const options = ref({
    chart: {
        type: 'bar',
        height: 350,
        stacked: false, // Set to false for horizontal bars
    },
    plotOptions: {
        bar: {
            horizontal: true, // Make bars horizontal
            dataLabels: {
                total: {
                    enabled: true,
                    offsetX: 0,
                    style: {
                        fontSize: '14px',
                        fontWeight: 'bold',
                    },
                },
            },
        },
    },
    xaxis: {
        categories: ['Product A', 'Product B', 'Product C'],
        title: {
            text: 'Products',
        },
    },
    yaxis: {
        title: {
            text: 'Sales',
        },
    },
    tooltip: {
        y: {
            formatter: function (val) {
                return val + ' units';
            },
        },
    },
    fill: {
        opacity: 1,
    },
    legend: {
        position: 'top',
        horizontalAlign: 'left',
    },
});

const presenceData = ref([
    { timestamp: '2024-07-17T09:00:00', value: 1 },
    { timestamp: '2024-07-17T12:00:00', value: 0 },
    { timestamp: '2024-07-17T15:00:00', value: 1 },
]);


const series = ref([
    {
        name: 'Presence',
        data: presenceData.value.map((item) => item.value),
    },
]);


const chartOptions = {
    chart: {
        type: 'bar',
        height: '350', // Adjust chart height as needed
        stacked: true, // Set to true for stacked bars
        toolbar: {
            show: false, // Optional: Hide chart toolbar
        },
    },
    plotOptions: {
        bar: {
            horizontal: true, // Make bars horizontal
            dataLabels: {
                total: {
                    enabled: true,
                    offsetX: 0,
                    style: {
                        fontSize: '14px',
                        fontWeight: 'bold',
                    },
                },
            },
        },
    },
    xaxis: {
        categories: presenceData.value.map((item) => item.timestamp),
        title: {
            text: 'Timestamps',
        },
    },
    yaxis: {
        title: {
            text: 'Presence Status',
        },
        labels: {
            formatter: function (val: number) {
                return val === 1 ? 'On' : 'Off';
            },
        },
    },

    colors: ['#38a47d', '#dc3545'], // Customize colors
    fill: {
        opacity: 0.8,
    },
    legend: {
        show: false, // Optional: Hide legend
    },
    dataLabels: {
        enabled: false, // Optional: Disable data labels
    },
};

onBeforeMount(() => {

});
</script>

<template>
    <div class="chart-container">
        <VueApexCharts width="800" height="400" :options="chartOptions" :series="series"></VueApexCharts>
    </div>
</template>

<!-- 
// Bar
// Line -->
