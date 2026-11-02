<script setup lang="ts">
  import { computed } from 'vue';
  import UiIcon from '@/components/ui/UiIcon.vue';
  import type { IconName } from '@/components/ui/icons';
  import {
    airQualityCategoryLabel,
    airQualityCategoryTone,
    airQualityIndexFamilyLabel,
    currentAirQualityPoint,
    pollenReadings,
    upcomingAirQualityHours,
  } from '@/domain/integrations/air-quality';
  import type { AirQualityForecast } from '@/types/integrations/air-quality';
  import { createZonedTimeFormatter, formatInstant, formatLocalDateTime } from '@/utils/date.utils';
  import { formatQuantity } from '@/utils/units';

  const props = defineProps<{ forecast: AirQualityForecast }>();
  const current = computed(() => currentAirQualityPoint(props.forecast));
  const categoryLabel = computed(() => airQualityCategoryLabel(props.forecast.category));
  const tone = computed(() => airQualityCategoryTone(props.forecast.category));
  const familyLabel = computed(() => airQualityIndexFamilyLabel(props.forecast.index_family));

  const pollutants = computed(() => {
    const point = current.value;
    if (!point) return [];
    const values: { id: string; label: string; value: number | null; icon: IconName }[] = [
      { id: 'pm2_5', label: 'PM2.5', value: point.pm2_5, icon: 'weatherParticle' },
      { id: 'pm10', label: 'PM10', value: point.pm10, icon: 'weatherParticle' },
      { id: 'ozone', label: 'Ozone', value: point.ozone, icon: 'weatherAir' },
      { id: 'nitrogen_dioxide', label: 'NO₂', value: point.nitrogen_dioxide, icon: 'weatherAir' },
      { id: 'sulphur_dioxide', label: 'SO₂', value: point.sulphur_dioxide, icon: 'weatherAir' },
      { id: 'carbon_monoxide', label: 'CO', value: point.carbon_monoxide, icon: 'weatherAir' },
    ];
    return values.flatMap((item) =>
      item.value == null ? [] : [{ ...item, value: formatQuantity(item.value, 'pollutant') }]
    );
  });

  const pollens = computed(() =>
    pollenReadings(props.forecast.pollen).map((item) => ({
      ...item,
      value: formatQuantity(item.value, 'pollen'),
    }))
  );

  const hourly = computed(() => {
    const formatter = createZonedTimeFormatter(props.forecast.timezone);
    return upcomingAirQualityHours(props.forecast).map((point) => ({
      id: point.time,
      time: formatInstant(point.time, formatter),
      pm25: point.pm2_5 == null ? null : formatQuantity(point.pm2_5, 'pollutant'),
      pm10: point.pm10 == null ? null : formatQuantity(point.pm10, 'pollutant'),
    }));
  });
</script>

<template>
  <div class="aq-results">
    <section class="aq-panel" aria-label="Air quality summary">
      <p class="aq-muted">{{ forecast.location }} · {{ forecast.timezone }}</p>
      <div class="aq-summary">
        <div>
          <span class="aq-category" :data-tone="tone">{{ categoryLabel }}</span>
          <strong>{{ forecast.index == null ? '—' : formatQuantity(forecast.index, 'aqi') }}</strong>
          <small>{{ familyLabel }}</small>
        </div>
        <UiIcon name="weatherAir" class="aq-badge" />
      </div>
      <div v-if="pollutants.length" class="aq-grid">
        <div v-for="item in pollutants" :key="item.id">
          <span class="aq-stat-label"><UiIcon :name="item.icon" />{{ item.label }}</span>
          <strong>{{ item.value }}</strong>
        </div>
      </div>
      <p v-else class="aq-muted">No pollutant values were returned.</p>
      <p class="aq-muted">Captured {{ formatLocalDateTime(forecast.captured_at) }}</p>
      <p class="aq-muted">Last checked {{ formatLocalDateTime(forecast.last_checked_at) }}</p>
      <p v-if="forecast.freshness === 'stale'" role="status">This air-quality data is out of date.</p>
    </section>

    <section v-if="pollens.length" class="aq-panel" aria-label="Pollen">
      <h2>Pollen</h2>
      <div class="aq-grid">
        <div v-for="item in pollens" :key="item.key">
          <span class="aq-stat-label"><UiIcon name="weatherPollen" />{{ item.label }}</span>
          <strong>{{ item.value }}</strong>
        </div>
      </div>
    </section>

    <section class="aq-panel">
      <header>
        <h2>Next 24 hours</h2>
      </header>
      <div class="aq-hours" role="region" aria-label="Hourly air quality" tabindex="0">
        <article v-for="item in hourly" :key="item.id">
          <span>{{ item.time }}</span>
          <UiIcon name="weatherAir" class="aq-hour-icon" />
          <small v-if="item.pm25">PM2.5 {{ item.pm25 }}</small>
          <small v-if="item.pm10">PM10 {{ item.pm10 }}</small>
        </article>
      </div>
      <p v-if="!hourly.length" class="aq-muted">No upcoming hourly air quality is available.</p>
    </section>

    <p class="aq-muted">{{ forecast.attribution }}</p>
  </div>
</template>

<style scoped>
  .aq-results {
    display: grid;
    gap: 1rem;
  }
  .aq-panel {
    min-width: 0;
    padding: 1.25rem;
    border: var(--nh-border-w) solid var(--nh-border);
    border-radius: var(--nh-radius-l);
    background: var(--nh-surface);
  }
  .aq-panel h2 {
    margin: 0;
    font-size: 1.1rem;
  }
  .aq-panel header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 1rem;
  }
  .aq-muted {
    margin: 0.5rem 0 0;
    color: var(--nh-text-muted);
    font-size: 0.8rem;
  }
  .aq-summary {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 1rem;
    margin: 0.75rem 0 1.25rem;
  }
  .aq-summary > div {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
  }
  .aq-summary strong {
    font-size: 2.5rem;
    line-height: 1;
  }
  .aq-summary small {
    color: var(--nh-text-muted);
  }
  .aq-badge {
    width: 2.5rem;
    height: 2.5rem;
    color: var(--nh-accent);
  }
  .aq-category {
    font-size: 0.72rem;
    font-weight: 600;
    letter-spacing: 0.04em;
    text-transform: uppercase;
    color: var(--nh-text-muted);
  }
  .aq-category[data-tone='good'] {
    color: var(--nh-ok);
  }
  .aq-category[data-tone='fair'] {
    color: var(--nh-info);
  }
  .aq-category[data-tone='moderate'] {
    color: var(--nh-warn);
  }
  .aq-category[data-tone='poor'] {
    color: var(--nh-danger);
  }
  .aq-grid {
    display: grid;
    grid-template-columns: repeat(4, 1fr);
    gap: 0.75rem;
  }
  .aq-grid > div {
    min-width: 0;
    padding: 0.85rem;
    border-radius: var(--nh-radius-m);
    background: var(--nh-surface-2);
  }
  .aq-stat-label {
    display: flex;
    align-items: center;
    gap: 0.4rem;
    color: var(--nh-text-muted);
    font-size: 0.8rem;
  }
  .aq-stat-label :deep(.nh-ic) {
    width: 1.1rem;
    height: 1.1rem;
  }
  .aq-grid strong {
    display: block;
    margin-top: 0.45rem;
  }
  .aq-hours {
    display: flex;
    gap: 0.75rem;
    overflow-x: auto;
    scroll-snap-type: x mandatory;
    padding-bottom: 0.5rem;
  }
  .aq-hours article {
    flex: 1 0 7.5rem;
    display: grid;
    gap: 0.55rem;
    padding: 0.85rem;
    border-radius: var(--nh-radius-m);
    background: var(--nh-surface-2);
    text-align: center;
    scroll-snap-align: start;
  }
  .aq-hours article > strong {
    font-size: 1.15rem;
  }
  .aq-hours article > small {
    color: var(--nh-text-muted);
    font-size: 0.72rem;
  }
  .aq-hour-icon {
    width: 1.3rem;
    height: 1.3rem;
    margin: 0 auto;
    color: var(--nh-text-muted);
  }
  @media (max-width: 700px) {
    .aq-panel {
      padding: 1rem;
    }
    .aq-grid {
      grid-template-columns: repeat(2, 1fr);
    }
    .aq-hours article {
      flex: 0 0 calc((100% - 1rem) / 3);
      padding: 0.65rem;
    }
  }
</style>
