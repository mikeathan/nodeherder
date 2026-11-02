<script setup lang="ts">
  import { computed } from 'vue';
  import UiIcon from '@/components/ui/UiIcon.vue';
  import { weatherIcon } from '@/components/ui/icons';
  import { weatherDayCard } from '@/integrations/weather/day-card';
  import type { WeatherDay } from '@/types/integrations/weather';

  const props = defineProps<{ day: WeatherDay; selected: boolean }>();
  defineEmits<{ select: [] }>();
  const model = computed(() => weatherDayCard(props.day));
</script>

<template>
  <button
    type="button"
    class="weather-day-card"
    :class="{ 'is-selected': selected }"
    :aria-pressed="selected"
    @click="$emit('select')">
    <span class="weather-day-card-head">
      <strong>{{ model.title }}</strong>
      <span class="weather-day-card-icon" aria-hidden="true">{{ weatherIcon(model.condition) }}</span>
    </span>
    <span v-if="model.primary" class="weather-day-card-primary">{{ model.primary }}</span>
    <span class="weather-day-card-condition">{{ model.conditionLabel }}</span>
    <span v-if="model.metrics.length" class="weather-day-card-metrics">
      <span v-for="metric in model.metrics" :key="metric.id" class="weather-day-card-metric">
        <UiIcon :name="metric.icon" />
        {{ metric.label }} {{ metric.value }}
      </span>
    </span>
  </button>
</template>

<style scoped>
  .weather-day-card {
    display: flex;
    flex-direction: column;
    gap: 0.35rem;
    flex: 1 0 11rem;
    min-width: 0;
    padding: 0.9rem 1rem;
    border: var(--nh-border-w) solid transparent;
    border-radius: var(--nh-radius-m);
    background: var(--nh-surface-2);
    color: inherit;
    font: inherit;
    text-align: left;
    cursor: pointer;
    scroll-snap-align: start;
  }
  .weather-day-card:hover {
    border-color: var(--nh-border-strong);
  }
  .weather-day-card:focus-visible {
    outline: 3px solid var(--nh-accent);
    outline-offset: 2px;
  }
  .weather-day-card.is-selected {
    border-color: var(--nh-accent);
    box-shadow: 0 0 0 1px var(--nh-accent);
  }
  .weather-day-card-head {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 0.5rem;
  }
  .weather-day-card-head strong {
    font-size: 0.9rem;
  }
  .weather-day-card-icon {
    font-size: 1.7rem;
    line-height: 1;
  }
  .weather-day-card-primary {
    font-size: 1.15rem;
    font-weight: 600;
  }
  .weather-day-card-condition {
    color: var(--nh-text-muted);
    font-size: 0.82rem;
  }
  .weather-day-card-metrics {
    display: flex;
    flex-wrap: wrap;
    gap: 0.35rem 0.6rem;
    margin-top: 0.3rem;
  }
  .weather-day-card-metric {
    display: inline-flex;
    align-items: center;
    gap: 0.25rem;
    color: var(--nh-text-muted);
    font-size: 0.75rem;
  }
  .weather-day-card-metric :deep(.nh-ic) {
    width: 0.95rem;
    height: 0.95rem;
  }
  @media (max-width: 700px) {
    .weather-day-card {
      flex: none;
    }
  }
</style>
