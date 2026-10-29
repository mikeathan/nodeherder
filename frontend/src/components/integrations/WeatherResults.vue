<script setup lang="ts">
  import { computed, ref } from 'vue';
  import { useForecastView } from '@/integrations/weather/useForecastView';
  import { weatherHeadline } from '@/integrations/weather/headline';
  import type { WeatherForecast } from '@/types/weather.type';
  import { weatherLabel } from '@/domain/weather';
  import { formatQuantity } from '@/utils/units';
  import { weatherIcon } from '@/components/ui/icons';
  import { formatLocalDateTime } from '@/utils/date.utils';
  import UiButton from '@/components/ui/UiButton.vue';
  const props = defineProps<{ forecast: WeatherForecast }>();
  const hours = ref<HTMLElement>();
  const days = ref<HTMLElement>();
  const headline = computed(() => weatherHeadline(props.forecast));
  const { page, daily, hourly, pages, firstDay, lastDay, hour, date, outsidePage } = useForecastView(
    () => props.forecast
  );
  function scroll(target: HTMLElement | undefined, direction: number) {
    target?.scrollBy({ left: direction * target.clientWidth, behavior: 'smooth' });
  }
</script>
<template>
  <div class="weather-results">
    <section class="weather-panel" aria-label="Weather summary">
      <p class="weather-muted">
        {{ forecast.location }} · {{ headline?.label ?? 'Forecast' }} · {{ forecast.timezone }}
      </p>
      <div class="weather-current">
        <strong>{{
          headline?.temperature != null ? formatQuantity(headline.temperature, 'temperature') : '—'
        }}</strong
        ><span aria-hidden="true">{{ weatherIcon(headline?.condition) }}</span>
      </div>
      <h2>{{ headline ? weatherLabel(headline.condition) : 'No saved forecast' }}</h2>
      <p v-if="forecast.current" class="weather-muted">
        Feels like {{ formatQuantity(forecast.current.feels_like, 'temperature') }}
      </p>
      <div v-if="forecast.current" class="weather-stats">
        <div>
          Rain chance<strong>{{ formatQuantity(forecast.current.precipitation_probability, 'percent') }}</strong>
        </div>
        <div>
          Humidity<strong>{{ formatQuantity(forecast.current.humidity, 'percent') }}</strong>
        </div>
        <div>
          Wind<strong>{{ formatQuantity(forecast.current.wind, 'speed') }}</strong>
        </div>
        <div>
          Wind gusts<strong>{{ formatQuantity(forecast.current.gust, 'speed') }}</strong>
        </div>
      </div>
      <p v-else class="weather-muted" role="status">
        {{
          headline
            ? `No current observation. Showing ${headline.label.toLowerCase()}.`
            : 'No saved forecast. Refresh to retrieve one.'
        }}
      </p>
      <p class="weather-muted">Last refreshed {{ formatLocalDateTime(forecast.captured_at) }}</p>
      <p v-if="forecast.freshness === 'stale'" role="status">This forecast is out of date.</p>
    </section>
    <section class="weather-panel">
      <header>
        <h2>Next 24 hours</h2>
        <div>
          <UiButton variant="ghost" aria-label="Earlier hours" @click="scroll(hours, -1)">←</UiButton
          ><UiButton variant="ghost" aria-label="Later hours" @click="scroll(hours, 1)">→</UiButton>
        </div>
      </header>
      <div ref="hours" class="weather-hours" role="region" aria-label="Hourly forecast" tabindex="0">
        <article v-for="item in hourly" :key="item.time">
          <span>{{ hour(item.time) }}</span
          ><span aria-hidden="true">{{ weatherIcon(item.condition) }}</span
          ><strong>{{ formatQuantity(item.temperature, 'temperature', { short: true }) }}</strong
          ><small>{{ formatQuantity(item.precipitation_probability, 'percent') }} rain</small>
        </article>
      </div>
      <p v-if="!hourly.length" class="weather-muted">
        No upcoming hourly forecast is available. Refresh for newer information.
      </p>
    </section>
    <section class="weather-panel">
      <header>
        <h2>{{ daily.length }}-day forecast</h2>
        <div class="desktop-controls">
          <UiButton variant="ghost" aria-label="Earlier days" @click="scroll(days, -1)">←</UiButton
          ><UiButton variant="ghost" aria-label="Later days" @click="scroll(days, 1)">→</UiButton>
        </div>
      </header>
      <div ref="days" class="weather-days" role="region" aria-label="Daily forecast" tabindex="0">
        <article v-for="(item, index) in daily" :key="item.date" :class="{ 'outside-page': outsidePage(index) }">
          <strong>{{ date(item.date) }}</strong
          ><span class="day-icon" aria-hidden="true">{{ weatherIcon(item.condition) }}</span
          ><b
            >{{ formatQuantity(item.high, 'temperature', { short: true }) }} /
            {{ formatQuantity(item.low, 'temperature', { short: true }) }}</b
          >
          <p>{{ weatherLabel(item.condition) }}</p>
          <p>Rain chance {{ formatQuantity(item.precipitation_probability, 'percent') }}</p>
          <p>Wind {{ formatQuantity(item.wind, 'speed') }}</p>
        </article>
      </div>
      <div v-if="pages > 1" class="weather-pagination">
        <UiButton :disabled="page === 0" @click="page--">← Previous</UiButton
        ><span role="status">Days {{ firstDay }}–{{ lastDay }} of {{ daily.length }}</span
        ><UiButton :disabled="page >= pages - 1" @click="page++">Next →</UiButton>
      </div>
      <p v-if="!daily.length" class="weather-muted">No upcoming daily forecast is available.</p>
    </section>
    <p class="weather-muted">{{ forecast.attribution }}</p>
  </div>
</template>
<style scoped>
  .weather-results {
    display: grid;
    gap: 1rem;
  }
  .weather-panel {
    min-width: 0;
    border: var(--nh-border-w) solid var(--nh-border);
    border-radius: var(--nh-radius-l);
    padding: 1.25rem;
    background: var(--nh-surface);
  }
  h2 {
    margin: 0;
    font-size: 1.1rem;
  }
  header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: 1rem;
    margin-bottom: 1rem;
  }
  header > div {
    display: flex;
    gap: 0.5rem;
  }
  .weather-current {
    display: flex;
    align-items: center;
    justify-content: space-between;
    font-size: 3rem;
    margin: 0.75rem 0;
  }
  .weather-stats {
    display: grid;
    grid-template-columns: repeat(4, 1fr);
    gap: 0.75rem;
    margin: 1.25rem 0;
  }
  .weather-stats > div,
  .weather-hours article,
  .weather-days article {
    background: var(--nh-surface-2);
    border-radius: var(--nh-radius-m);
    padding: 1rem;
  }
  .weather-stats strong {
    display: block;
    margin-top: 0.5rem;
  }
  .weather-muted,
  small {
    color: var(--nh-text-muted);
    font-size: 0.8rem;
  }
  .weather-hours,
  .weather-days {
    display: flex;
    gap: 0.75rem;
    overflow-x: auto;
    scroll-snap-type: x mandatory;
    padding-bottom: 0.5rem;
  }
  .weather-hours article {
    flex: 1 0 7rem;
    display: grid;
    gap: 0.75rem;
    text-align: center;
    scroll-snap-align: start;
  }
  .weather-days article {
    flex: 1 0 11rem;
    scroll-snap-align: start;
  }
  .day-icon {
    display: block;
    font-size: 2rem;
    margin: 0.75rem 0;
  }
  .weather-days p {
    font-size: 0.85rem;
    color: var(--nh-text-muted);
    margin: 0.5rem 0;
  }
  .weather-pagination {
    display: none;
  }
  @media (max-width: 700px) {
    .weather-panel {
      padding: 1rem;
    }
    .weather-stats {
      grid-template-columns: repeat(2, 1fr);
    }
    .weather-hours {
      gap: 0.5rem;
    }
    .weather-hours article {
      flex: 0 0 calc((100% - 1rem) / 3);
      min-width: 0;
      padding: 0.65rem;
    }
    .weather-days {
      display: grid;
      overflow: visible;
    }
    .weather-days article {
      display: grid;
      grid-template-columns: 1fr 2rem auto;
      gap: 0.3rem 0.5rem;
      padding: 0.75rem;
    }
    .weather-days article.outside-page {
      display: none;
    }
    .day-icon {
      margin: 0;
      font-size: 1.5rem;
    }
    .weather-days p {
      margin: 0;
    }
    .weather-days p:nth-of-type(2) {
      grid-column: 2/4;
      text-align: right;
    }
    .weather-days p:nth-of-type(3) {
      grid-column: 1/-1;
    }
    .desktop-controls {
      display: none;
    }
    header .desktop-controls {
      display: none;
    }
    .weather-pagination {
      display: flex;
      justify-content: space-between;
      align-items: center;
      gap: 0.5rem;
      font-size: 0.75rem;
      margin-top: 0.75rem;
    }
  }
</style>
