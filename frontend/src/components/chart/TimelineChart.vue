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
    return convertToApexTimelineRangebarData(presenceData.value);
});

type PresenceData = {
    timestamp: string; // Timestamp string
    value: 0 | 1;
};


function convertToApexTimelineRangebarData(
    data: PresenceData[],
) {

    const apexData: any[] = [];

    const groupedByPresence: Record<number, number[]> = data.reduce((acc: any, curr) => {
        const presence: number = curr.value as number;
        acc[presence] = acc[presence] || [];
        acc[presence].push(new Date(curr.timestamp).getTime(),);
        return acc;
    }, {} as Record<number, number[]>);



    for (const presence in groupedByPresence) {
        const timestamps: number[] = groupedByPresence[presence];
        if (timestamps.length !== 2) {
            console.warn(
                `Presence value ${presence} has ${timestamps.length} timestamps, expected 2 for rangebar chart. Skipping this presence.`
            );
            continue;
        }

        apexData.push({
            x: presence,
            y: timestamps,
        });
    }
    // // Convert timestamps to epoch milliseconds for ApexCharts
    // for (const presence in presenceGroups) {
    //     const timestamps = presenceGroups[presence].map(
    //         (timestamp: Date) => new Date(timestamp).getTime(),
    //     );
    //     apexData.push({
    //         x: presence === '1' ? 'Present' : 'Absent', // Set labels based on presence value
    //         y: timestamps,
    //     });
    // }

    // return groupedByPresence;
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

onBeforeMount(() => {

});
</script>

<template>

    {{ chartDataTest }}
    <div class="chart-container">
        <VueApexCharts width="800" height="400" :options="chartOptions" :series="series">
        </VueApexCharts>
    </div>
</template>

<!-- 
// Bar
// Line -->
