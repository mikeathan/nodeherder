<script setup lang="ts">
import { ref, computed } from 'vue';
import type { PropType, Ref } from 'vue';
import BaseChart from './BaseChart.vue';
import { AreaChartEntry } from '@/types/chart.type';

const props = defineProps({
    chartData: {
        type: Object as PropType<any>,
        default: null,
    },
});

const NEWTemperatureData = {
    name: 'Temperature',
    timestamp: [
        '2024-07-17T09:00:00',
        '2024-07-17T12:00:00',
        '2024-07-17T14:00:00',
        '2024-07-17T18:00:00',
        '2024-07-17T20:00:00',
        '2024-07-17T20:05:00',
        '2024-07-17T20:10:00',
        '2024-07-17T20:24:00',
    ],
    value: [15.1, 15.9, 16.2, 16.5, 17.1, 15.1, 15.0, 14.5],
};

const transformedChartData = computed(() => {
    return [
        {
            name: NEWTemperatureData.name,
            data: NEWTemperatureData.timestamp.map((timestamp, index) => ({
                x: timestamp,
                y: NEWTemperatureData.value[index],
            })),
        },
    ] as AreaChartEntry[];
});

const chartOptions = {
    chart: {
        type: 'area',
        background: '#fff',
        toolbar: {
            show: false,
        },
    },
    dataLabels: {
        enabled: false,
    },

    stroke: {
        curve: 'smooth',
    },
    xaxis: {
        type: 'datetime',
        labels: {
            datetimeFormatter: {
                year: 'yyyy',
                month: "MMM 'yy",
                day: 'dd MMM',
                hour: 'HH:mm',
            },
        },
    },
    Tooltip: {
        x: {
            format: 'dd/MMM/yy HH:mm:ss ',
        },
    },
    responsive: [
        {
            breakpoint: undefined,
            options: {
                chart: {
                    width: '100%',
                },
            },
        },
    ],
};

</script>

<template>
    <div class="area-chart">
        <BaseChart :data="transformedChartData" :options="chartOptions" />
    </div>
</template>
