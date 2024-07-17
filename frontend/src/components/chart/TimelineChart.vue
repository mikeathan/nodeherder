<script setup lang="ts">
import { ref, onBeforeMount } from 'vue';
import type { PropType, Ref } from 'vue';
import 'chartjs-adapter-date-fns';
import VueApexCharts from 'vue3-apexcharts'

const presenceData = ref([
    // Replace with your actual presence data
    { timestamp: '2024-07-17T09:00:00', value: 1 }, // On at 9:00
    { timestamp: '2024-07-17T12:00:00', value: 0 }, // Off at 12:00
    { timestamp: '2024-07-17T14:00:00', value: 1 }, // On at 14:00
    { timestamp: '2024-07-17T18:00:00', value: 0 }, // Off at 18:00

]);


const series = {
    data: [
        {
            x: 'Code',
            y: [
                new Date('2019-03-02').getTime(),
                new Date('2019-03-04').getTime()
            ]
        },
        {
            x: 'Test',
            y: [
                new Date('2019-03-04').getTime(),
                new Date('2019-03-08').getTime()
            ]
        },
        {
            x: 'Validation',
            y: [
                new Date('2019-03-08').getTime(),
                new Date('2019-03-12').getTime()
            ]
        },
        {
            x: 'Deployment',
            y: [
                new Date('2019-03-12').getTime(),
                new Date('2019-03-18').getTime()
            ]
        }
    ]
}




const chartOptions = {
    chart: {
        height: 350,
        type: 'rangeBar'
    },
    plotOptions: {
        bar: {
            horizontal: true
        }
    },
    xaxis: {
        type: 'datetime'
    }
};


onBeforeMount(() => {

});
</script>

<template>
    <div class="chart-container">
        <VueApexCharts width="800" height="400" :options="chartOptions" :series="series">
        </VueApexCharts>
    </div>
</template>

<!-- 
// Bar
// Line -->
