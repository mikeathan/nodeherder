<script setup lang="ts">
import { ref, onBeforeMount, computed } from 'vue';
import type { PropType, Ref } from 'vue';
import VueApexCharts from 'vue3-apexcharts';


const props = defineProps({
    chartData: {
        type: Object as PropType<any>,
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
        y: number[];
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
        // default to now as its still in that state

        // TODO:
        // maybe get the range ofthe query and use that for the next timestamp
        // that something to be done on the server side 
        const nextTimestamp = (j + 1 >= data.length) ?
            new Date().getTime() /* TEMPORARY */ :
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


const chartOptions = {
    chart: {
        type: 'rangeBar',
        background: '#fff',
        toolbar: {
            show: false
        }
    },
    tooltip: {
        x: {
            format: "dd/MMM/yy HH:mm:ss ",
        }
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
    },
    colors: ['#FF4560', '#00E396',],
    xaxis: {
        type: 'datetime',
        labels: {
            datetimeFormatter: {
                year: 'yyyy',
                month: 'MMM \'yy',
                day: 'dd MMM',
                hour: 'HH:mm'
            }
        }
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
    responsive: [
        {
            breakpoint: undefined, // Matches all screens
            options: {
                chart: {
                    width: '100%' // Set chart width to 100% for all screens
                }
            }
        }
    ]
};

onBeforeMount(() => {

});
</script>

<template>
    <div class="chart-container">
        <VueApexCharts :options="chartOptions" :series="chartDataTest">
        </VueApexCharts>
    </div>
</template>
