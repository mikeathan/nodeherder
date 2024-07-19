<script setup lang="ts">
import { ref, onBeforeMount, computed } from 'vue';
import type { PropType, Ref } from 'vue';
import VueApexCharts from 'vue3-apexcharts';


const props = defineProps({
    chartData: {
        type: Object as PropType<any>, // ChartData<'bar'
        default: null,
    },
    options: {
        type: Object,
        default: null,
    },
});
const presenceData = ref<PresenceData[]>([
    // Replace with your actual presence data
    { timestamp: '2024-07-17T09:00:00', value: 1 }, // On at 9:00
    { timestamp: '2024-07-17T12:00:00', value: 0 }, // Off at 12:00
    { timestamp: '2024-07-17T14:00:00', value: 1 }, // On at 14:00
    { timestamp: '2024-07-17T18:00:00', value: 0 }, // Off at 18:00
    { timestamp: '2024-07-17T20:00:00', value: 1 }, // On at 20:00
    { timestamp: '2024-07-17T20:05:00', value: 0 }, // Off at 20:05
    { timestamp: '2024-07-17T20:10:00', value: 1 }, // On at 20:10
    { timestamp: '2024-07-17T20:24:00', value: 0 }, // Off at 20:24
])

const chartDataTest = computed(() => {
    return convertToTimelineRangebarData(presenceData.value);
});

type PresenceData = {
    timestamp: string; // Timestamp string
    value: 0 | 1;
};

type TimelineChartEntry = {
    name: string;
    data: [{
        x: string;
        y: number[]
    }]
};


function convertToTimelineRangebarData(
    data: PresenceData[],
): TimelineChartEntry[] {
    const transformedData: TimelineChartEntry[] = [];

    for (let j = 0; j < data.length; j++) {
        const currentValue = data[j].value;
        const currentTimestamp = data[j].timestamp;

        // if we dont have next timestamp
        //d efault to now as its still in that state
        const nextTimestamp = (j + 1 >= data.length) ?
            new Date().getTime() :
            new Date(data[j + 1].timestamp).getTime();

        transformedData.push({
            name: currentValue === 0 ? "Present" : "Absent",
            data: [{
                x: 'Presence',
                y: [new Date(currentTimestamp).getTime(), new Date(nextTimestamp).getTime()],
            }]
        });
    }

    return transformedData;
}

// const series: TimelineChartEntry[] = [
//     {
//         name: 'ON',
//         data: [
//             {
//                 x: 'Presence',
//                 y: [
//                     new Date('2024-07-17T09:00:00').getTime(),
//                     new Date('2024-07-17T12:00:00').getTime(),
//                 ],
//             },
//         ],
//     },
//     {
//         name: 'OFF',
//         data: [
//             {
//                 x: 'Presence',
//                 y: [
//                     new Date('2024-07-17T12:00:00').getTime(),
//                     new Date('2024-07-17T14:00:00').getTime(),
//                 ],
//             },
//         ],
//     },
// ];

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

onBeforeMount(() => {

});
</script>

<template>
    <div class="chart-container">
        <VueApexCharts width="800" height="400" :options="chartOptions" :series="chartDataTest">
        </VueApexCharts>
    </div>
</template>
