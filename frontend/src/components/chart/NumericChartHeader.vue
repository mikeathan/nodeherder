<script setup lang="ts">
  import type { PropType } from 'vue';
  import type { NumericChartHeaderData } from '@/types/metrics.type';


  defineProps({
    chart: {
      type: Object as PropType<NumericChartHeaderData>,
      required: true,
    },
    formatValue: {
      type: Function as PropType<(name: string, value: number) => string>,
      required: true,
    },
    formatDelta: {
      type: Function as PropType<(name: string, delta: number, deltaPct: number) => string>,
      required: true,
    },
  });
</script>

<template>
  <div class="chart-header">
    <div class="chart-title">
      <span class="chart-dot" :style="{ backgroundColor: chart.color }"></span>
      <span class="chart-name">{{ chart.label }}</span>
      <span v-if="chart.unit" class="chart-unit">{{ chart.unit }}</span>
    </div>
    <div v-if="chart.stats" class="chart-stats grid w-full">
      <div class="col-6 md:col-2">
        <div class="stat">
          <span class="stat-label">Last</span>
          <span class="stat-value">{{ formatValue(chart.name, chart.stats.last) }}</span>
        </div>
      </div>
      <div class="col-6 md:col-2">
        <div class="stat">
          <span class="stat-label">Min</span>
          <span class="stat-value">{{ formatValue(chart.name, chart.stats.min) }}</span>
        </div>
      </div>
      <div class="col-6 md:col-2">
        <div class="stat">
          <span class="stat-label">Avg</span>
          <span class="stat-value">{{ formatValue(chart.name, chart.stats.avg) }}</span>
        </div>
      </div>
      <div class="col-6 md:col-2">
        <div class="stat">
          <span class="stat-label">Max</span>
          <span class="stat-value">{{ formatValue(chart.name, chart.stats.max) }}</span>
        </div>
      </div>
      <div class="col-12 md:col-2">
        <div class="stat delta" :class="{ up: chart.stats.delta > 0, down: chart.stats.delta < 0 }">
          <span class="stat-label">Δ</span>
          <span class="stat-value">{{ formatDelta(chart.name, chart.stats.delta, chart.stats.deltaPct) }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
  .chart-header {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: 16px;
    padding: 6px 0 10px;
    flex-wrap: wrap;
  }

  .chart-title {
    display: flex;
    align-items: center;
    gap: 8px;
    font-weight: 600;
    color: #e5e7eb;
  }

  .chart-dot {
    width: 10px;
    height: 10px;
    border-radius: 999px;
    display: inline-block;
    box-shadow: 0 0 12px rgba(0, 0, 0, 0.35);
  }

  .chart-name {
    font-size: 1rem;
  }

  .chart-unit {
    font-size: 0.8rem;
    color: #9aa4b2;
  }

  .stat {
    display: flex;
    flex-direction: column;
    gap: 2px;
    padding: 6px 10px;
    border-radius: 8px;
    background: rgba(15, 23, 42, 0.35);
    min-width: 90px;
  }

  .stat-label {
    font-size: 0.65rem;
    letter-spacing: 0.08em;
    text-transform: uppercase;
    color: #8a94a6;
  }

  .stat-value {
    font-size: 0.9rem;
    color: #f8fafc;
    font-weight: 600;
  }

  .stat.delta {
    border: 1px solid transparent;
  }

  .stat.delta.up {
    border-color: rgba(34, 197, 94, 0.4);
    color: #bbf7d0;
  }

  .stat.delta.down {
    border-color: rgba(248, 113, 113, 0.4);
    color: #fecaca;
  }

  @media (max-width: 640px) {
    .chart-stats {
      display: grid;
      grid-template-columns: repeat(5, minmax(0, 1fr));
      gap: 6px;
    }

    .stat {
      padding: 6px 6px;
      min-width: 0;
    }

    .stat-label {
      font-size: 0.55rem;
      letter-spacing: 0.06em;
    }

    .stat-value {
      font-size: 0.8rem;
    }
  }
</style>
