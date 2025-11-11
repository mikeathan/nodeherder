<script setup lang="ts">
  import { ref, watch, toRaw, computed } from 'vue';
  import type { PropType, Ref } from 'vue';
  import BaseChart from '../BaseChart.vue';
  import { TimelineChartEntry } from '@/types/chart.type';
  import { DeviceExposeBinaryMetrics, DeviceExposeMetrics } from '@/types/metrics.type';
  import { ColorValue } from '@/types/color.type';
  import { getExposeBinaryColour } from '@/contracts/chart';
  import { ExposeTypes } from '@/types/device.type';

  const props = defineProps({
    chartData: {
      type: Object as PropType<DeviceExposeBinaryMetrics[]>,
      default: null,
    },
  });

  const timelineData = ref<TimelineChartEntry[]>([]);

  const timelineColours = computed(() => {
    return props.chartData.flatMap((item) => {
      if (item.type === ExposeTypes.Binary) {
        const color = getExposeBinaryColour(item.name);
        return [color.on, color.off];
      }
      return []; // TODO: handle Enum types
    });
  });

  watch(
    () => props.chartData,
    () => {
      if (props.chartData !== null) {
        timelineData.value = transformedChartData(props.chartData);
      }
    },
    { immediate: true }
  );

  function transformedChartData(exposeMetrics: DeviceExposeBinaryMetrics[]): TimelineChartEntry[] {
    const transformedData: TimelineChartEntry[] = [];

    exposeMetrics.forEach((item) => {
      item.data.forEach((point) => {
        const entry = {
          name: point.x,
          data: [
            {
              x: item.name,
              y: point.y,
            },
          ],
        };

        transformedData.push(entry);
      });
    });

    return transformedData;
  }

  const chartOptions = {
    chart: {
      type: 'rangeBar',
      background: '#fff',
      toolbar: {
        show: false,
      },
      zoom: {
        enabled: true,
        type: 'x',
      },
    },
    tooltip: {
      x: {
        format: 'dd MMM yyyy HH:mm:ss',
        formatter: function (value: number) {
          const date = new Date(value);
          const now = new Date();
          const diffMs = now.getTime() - date.getTime();
          const diffMins = Math.floor(diffMs / 60000);
          const diffHours = Math.floor(diffMs / 3600000);
          const diffDays = Math.floor(diffMs / 86400000);

          const timeStr = date.toLocaleTimeString('en-GB', {
            hour: '2-digit',
            minute: '2-digit',
            second: '2-digit',
          });

          if (diffMins < 60) {
            return `${diffMins} min ago (${timeStr})`;
          } else if (diffHours < 24) {
            return `${diffHours}h ago (${timeStr})`;
          } else if (diffDays === 1) {
            return `Yesterday ${timeStr}`;
          } else if (diffDays < 7) {
            return `${diffDays} days ago (${timeStr})`;
          }

          const dateStr = date.toLocaleDateString('en-GB', {
            day: '2-digit',
            month: 'short',
            year: 'numeric',
          });
          return `${dateStr} ${timeStr}`;
        },
      },
      y: {
        formatter: function (value: any, opts: any) {
          if (Array.isArray(value) && value.length === 2) {
            const start = new Date(value[0]);
            const end = new Date(value[1]);
            const durationMs = end.getTime() - start.getTime();
            const durationMins = Math.floor(durationMs / 60000);
            const durationHours = Math.floor(durationMins / 60);
            const remainingMins = durationMins % 60;

            if (durationHours > 0) {
              return `Duration: ${durationHours}h ${remainingMins}m`;
            }
            return `Duration: ${durationMins}m`;
          }
          return value;
        },
      },
    },
    legend: {
      show: false,
    },
    plotOptions: {
      bar: {
        horizontal: true,
        barHeight: '50%',
        rangeBarGroupRows: true,
      },
      fill: {
        type: 'solid',
      },
      xaxis: {
        type: 'datetime',
      },
    },
    colors: timelineColours.value,
    xaxis: {
      type: 'datetime',
      labels: {
        datetimeUTC: false,
        datetimeFormatter: {
          year: 'yyyy',
          month: "MMM 'yy",
          day: 'dd MMM',
          hour: 'HH:mm',
          minute: 'HH:mm',
        },
        rotate: 0,
        hideOverlappingLabels: true,
        style: {
          fontSize: '11px',
        },
      },
    },
    stroke: {
      width: 1,
    },
    fill: {
      type: 'solid',
      opacity: 0.6,
    },
    responsive: [
      {
        breakpoint: undefined, // Matches all screens
        options: {
          chart: {
            width: '100%', // Set chart width to 100% for all screens
          },
        },
      },
    ],
  };
</script>

<template>
  <div class="timeline-chart">
    <BaseChart height="200" :data="toRaw(timelineData)" :options="chartOptions" />
  </div>
</template>
