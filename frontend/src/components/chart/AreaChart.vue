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
const temperatureData = ref<MetricsData[]>([
    { timestamp: '2024-07-17T09:00:00', value: 15.1 },
    { timestamp: '2024-07-17T12:00:00', value: 15.9 },
    { timestamp: '2024-07-17T14:00:00', value: 16.2 },
    { timestamp: '2024-07-17T18:00:00', value: 16.5 },
    { timestamp: '2024-07-17T20:00:00', value: 17.1 },
    { timestamp: '2024-07-17T20:05:00', value: 15.1 },
    { timestamp: '2024-07-17T20:10:00', value: 15.0 },
    { timestamp: '2024-07-17T20:24:00', value: 14.5 },
])

const chartDataTest = computed(() => {
    return convertToTimelineRangebarData(presenceData.value);
});

type MetricsData = {
    timestamp: string;
    value: number;
};

type AreaChartEntry = {
    name: string;
    data: [{
        x: string;
        y: number;
    }]
};


function convert(
    data: MetricsData[],
): AreaChartEntry[] {

    const transformedData: AreaChartEntry[] = timestamps.map((timestamp, index) => (
        data: {
            x: timestamp,
            y: yValues[index],
        }))


    return transformedData;
}


const chartOptions = {
    chart: {
        height: 450,
        type: 'area',
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
    stroke: {
        curve: 'smooth'
    },
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
