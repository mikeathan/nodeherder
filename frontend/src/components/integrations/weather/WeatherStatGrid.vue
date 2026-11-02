<script setup lang="ts">
  import UiIcon from '@/components/ui/UiIcon.vue';
  import type { WeatherStat } from '@/integrations/weather/stats';

  defineProps<{ stats: WeatherStat[] }>();
</script>

<template>
  <div v-if="stats.length" class="weather-stats">
    <div v-for="stat in stats" :key="stat.id">
      <span class="weather-stat-label">
        <UiIcon v-if="stat.icon" :name="stat.icon" />
        {{ stat.label }}
      </span>
      <strong>{{ stat.value }}</strong>
    </div>
  </div>
</template>

<style scoped>
  .weather-stats {
    display: grid;
    grid-template-columns: repeat(4, 1fr);
    gap: 0.75rem;
    margin: 1.25rem 0;
  }
  .weather-stats > div {
    min-width: 0;
    padding: 1rem;
    border-radius: var(--nh-radius-m);
    background: var(--nh-surface-2);
  }
  .weather-stat-label {
    display: flex;
    align-items: center;
    gap: 0.4rem;
    color: var(--nh-text-muted);
    font-size: 0.8rem;
  }
  .weather-stat-label :deep(.nh-ic) {
    width: 1.1rem;
    height: 1.1rem;
  }
  .weather-stats strong {
    display: block;
    margin-top: 0.5rem;
  }
  @media (max-width: 700px) {
    .weather-stats {
      grid-template-columns: repeat(2, 1fr);
    }
  }
</style>
