<script setup lang="ts">
  import { computed, PropType } from 'vue';
  import VueApexCharts from 'vue3-apexcharts';
  import { NumericDataPoint } from '@/types/metrics.type';
  import { getFormattedSensorValueByName, getSensorName } from '@/modules/formatters/sensor-formatter';
  import { normalizeNumericData } from '@/utils/chart.utils';

  const props = defineProps({
    data: {
      type: Array as PropType<NumericDataPoint[]>,
      required: true,
    },
    exposeName: {
      type: String,
      required: true,
    },
    height: {
      type: Number,
      default: 150,
    },
    color: {
      type: String,
      default: undefined,
    },
    unit: {
      type: String,
      default: '',
    },
    maxValue: {
      type: Number,
      default: undefined,
    },
  });

  const normalizedData = computed(() => normalizeNumericData(props.data ?? []));
  const lastValue = computed(() => {
    const points = normalizedData.value;
    if (!points.length) return 0;
    return points[points.length - 1]?.y ?? 0;
  });

  const isPowerUnit = computed(() => {
    const unit = props.unit.toLowerCase();
    return unit === 'w' || unit === 'kw';
  });

  const gaugeMax = computed(() => {
    if (Number.isFinite(props.maxValue)) return props.maxValue as number;
    const unit = props.unit.toLowerCase();
    if (unit === 'w') return 1000;
    if (unit === 'kw') return 5;
    if (unit.includes('kwh')) return 10;
    if (unit.includes('wh')) return 1000;
    return 100;
  });

  const gaugePercent = computed(() => {
    const max = gaugeMax.value;
    if (!Number.isFinite(max) || max <= 0) return 0;
    return Math.min(1, Math.max(0, lastValue.value / max));
  });

  const needleRotation = computed(() => -90 + gaugePercent.value * 180);

  const gaugeSeries = computed(() => [1, 1, 1, 1]);

  const label = computed(() => {
    if (isPowerUnit.value) return 'Current Electricity Usage';
    return getSensorName(props.exposeName);
  });

  const formattedValue = computed(() => {
    const raw = getFormattedSensorValueByName(props.exposeName, lastValue.value, props.unit);
    if (!props.unit) return raw;
    return raw.replace(/(\d)([a-zA-Z])/g, '$1 $2');
  });

  const chartOptions = computed(() => ({
    chart: {
      type: 'donut',
      background: 'transparent',
      sparkline: { enabled: true },
    },
    stroke: { width: 0 },
    dataLabels: { enabled: false },
    plotOptions: {
      pie: {
        startAngle: -90,
        endAngle: 90,
        offsetY: 10,
        donut: {
          size: '72%',
        },
      },
    },
    colors: ['#2ecc71', '#7ed957', '#f4c430', '#e74c3c'],
    tooltip: { enabled: false },
    legend: { show: false },
  }));
</script>

<template>
  <div class="mini-energy-chart">
    <div class="energy-gauge" :style="{ '--gauge-height': `${height}px` }">
      <VueApexCharts :height="height" :options="chartOptions" :series="gaugeSeries" />
      <div class="gauge-needle" :style="{ transform: `translateX(-50%) rotate(${needleRotation}deg)` }"></div>
      <div class="gauge-center">
        <div class="gauge-value">{{ formattedValue }}</div>
        <div class="gauge-label">{{ label }}</div>
      </div>
    </div>
  </div>
</template>

<style scoped>
  .mini-energy-chart {
    width: 100%;
    padding: 0.75rem 0.75rem 0.5rem;
    background: linear-gradient(145deg, #1f1f1f, #151515);
    border-radius: 12px;
  }

  .energy-gauge {
    position: relative;
    height: var(--gauge-height);
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .gauge-needle {
    position: absolute;
    left: 50%;
    top: 50%;
    width: 2px;
    height: 30px;
    background: #f8fafc;
    transform-origin: bottom center;
    border-radius: 999px;
    box-shadow: 0 0 6px rgba(248, 250, 252, 0.6);
    z-index: 3;
  }

  .gauge-needle::after {
    content: '';
    position: absolute;
    bottom: -4px;
    left: 50%;
    width: 8px;
    height: 8px;
    background: #f8fafc;
    border-radius: 50%;
    transform: translateX(-50%);
    box-shadow: 0 0 6px rgba(248, 250, 252, 0.7);
  }

  .gauge-center {
    position: absolute;
    left: 50%;
    top: 64%;
    transform: translate(-50%, -50%);
    text-align: center;
    color: #f8fafc;
    z-index: 4;
  }

  .gauge-value {
    font-size: 26px;
    font-weight: 600;
    letter-spacing: 0.5px;
  }

  .gauge-label {
    font-size: 11px;
    color: #cbd5e1;
    margin-top: 2px;
  }

  :deep(.apexcharts-canvas) {
    margin: 0 auto;
    position: relative;
    z-index: 1;
  }

  :deep(.apexcharts-svg) {
    overflow: visible;
  }
</style>
