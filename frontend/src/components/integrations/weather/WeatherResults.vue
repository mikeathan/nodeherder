<script setup lang="ts">
  import { computed, nextTick, ref, watch } from 'vue';
  import { DAYS_PER_PAGE, useForecastView } from '@/integrations/weather/useForecastView';
  import { weatherHeadline } from '@/integrations/weather/headline';
  import type { WeatherForecast } from '@/types/integrations/weather';
  import { hoursForDate, weatherLabel } from '@/domain/integrations/weather';
  import { weatherDayDetail, weatherHourCards, weatherHourDetail } from '@/integrations/weather/day-detail';
  import { currentWeatherStats } from '@/integrations/weather/stats';
  import { formatQuantity } from '@/utils/units';
  import { weatherIcon } from '@/components/ui/icons';
  import { calendarDateInZone, formatLocalDateTime } from '@/utils/date.utils';
  import UiButton from '@/components/ui/UiButton.vue';
  import UiPager from '@/components/ui/UiPager.vue';
  import WeatherDayCard from './WeatherDayCard.vue';
  import WeatherStatGrid from './WeatherStatGrid.vue';
  const props = defineProps<{ forecast: WeatherForecast }>();
  const hourlyPanel = ref<HTMLElement>();
  const hourDetailPanel = ref<HTMLElement>();
  const selectedDate = ref<string | null>(null);
  const selectedHour = ref<string | null>(null);
  const headline = computed(() => weatherHeadline(props.forecast));
  const { page, daily, visibleDays, hourly, pages, firstDay, lastDay } = useForecastView(
    () => props.forecast
  );
  /** Resolve the selected day from the current forecast so a refresh cannot show a stale copy. */
  const selectedDay = computed(() =>
    selectedDate.value ? (daily.value.find((item) => item.date === selectedDate.value) ?? null) : null
  );
  /** Selected-day summary and hourly cards, or null when the default 24-hour view applies. */
  const selectedDetail = computed(() =>
    selectedDay.value ? weatherDayDetail(props.forecast, selectedDay.value) : null
  );
  const currentStats = computed(() => currentWeatherStats(props.forecast));
  const visibleHourPoints = computed(() =>
    selectedDay.value ? hoursForDate(props.forecast, selectedDay.value.date) : hourly.value
  );
  const activeHourCards = computed(() => weatherHourCards(props.forecast, visibleHourPoints.value));
  const selectedHourPoint = computed(() =>
    selectedHour.value ? (visibleHourPoints.value.find((point) => point.time === selectedHour.value) ?? null) : null
  );
  const selectedHourDetail = computed(() =>
    selectedHourPoint.value ? weatherHourDetail(props.forecast, selectedHourPoint.value) : null
  );
  const today = computed(() => calendarDateInZone(props.forecast.evaluated_at, props.forecast.timezone));
  const showingToday = computed(() => selectedDate.value === today.value);
  const selectedIndex = computed(() =>
    selectedDate.value ? daily.value.findIndex((day) => day.date === selectedDate.value) : -1
  );
  const canPreviousDay = computed(() => selectedIndex.value > 0);
  const canNextDay = computed(
    () => selectedIndex.value >= 0 && selectedIndex.value < daily.value.length - 1
  );
  const panelTitle = computed(() =>
    selectedDetail.value ? (showingToday.value ? 'Today' : selectedDetail.value.title) : 'Next 24 hours'
  );
  // A replaced resource or a day that no longer exists returns to the default hourly view.
  watch(
    () => props.forecast.resource_id,
    () => {
      selectedDate.value = null;
    }
  );
  watch(daily, (days) => {
    if (selectedDate.value && !days.some((day) => day.date === selectedDate.value)) selectedDate.value = null;
  });
  watch(selectedDate, () => {
    selectedHour.value = null;
  });
  watch(visibleHourPoints, (points) => {
    if (selectedHour.value && !points.some((point) => point.time === selectedHour.value)) selectedHour.value = null;
  });
  function moveSelectedDay(offset: number) {
    const index = selectedIndex.value + offset;
    const day = daily.value[index];
    if (!day) return;
    selectedDate.value = day.date;
    page.value = Math.floor(index / DAYS_PER_PAGE);
  }
  function scrollBehavior(): ScrollBehavior {
    return typeof matchMedia !== 'undefined' && matchMedia('(prefers-reduced-motion: reduce)').matches
      ? 'auto'
      : 'smooth';
  }
  function selectDay(date: string) {
    selectedDate.value = selectedDate.value === date ? null : date;
    if (selectedDate.value)
      nextTick(() => hourlyPanel.value?.scrollIntoView({ behavior: scrollBehavior(), block: 'start' }));
  }
  function selectHour(time: string) {
    selectedHour.value = selectedHour.value === time ? null : time;
    if (selectedHour.value)
      nextTick(() => hourDetailPanel.value?.scrollIntoView({ behavior: scrollBehavior(), block: 'nearest' }));
  }
</script>
<template>
  <div class="weather-results">
    <section v-if="!selectedDetail" class="weather-panel" aria-label="Weather summary">
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
      <WeatherStatGrid :stats="currentStats" />
      <p v-if="!forecast.current && headline" class="weather-muted" role="status">
        No current observation. Showing {{ headline.label.toLowerCase() }}.
      </p>
      <p v-if="!headline" class="weather-muted" role="status">
        No saved forecast. Refresh to retrieve one.
      </p>
      <p class="weather-muted">Last refreshed {{ formatLocalDateTime(forecast.captured_at) }}</p>
      <p v-if="forecast.freshness === 'stale'" role="status">This forecast is out of date.</p>
    </section>
    <section
      ref="hourlyPanel"
      class="weather-panel"
      :aria-label="selectedDetail ? `Weather for ${panelTitle}` : 'Next 24 hours'">
      <header>
        <h2>{{ panelTitle }}</h2>
        <div v-if="selectedDetail" class="weather-day-nav">
          <UiButton v-if="!showingToday" variant="ghost" @click="selectedDate = null">Today</UiButton>
          <UiButton
            variant="ghost"
            aria-label="Previous day"
            title="Previous day"
            :disabled="!canPreviousDay"
            @click="moveSelectedDay(-1)"
            >←</UiButton
          >
          <UiButton
            variant="ghost"
            aria-label="Next day"
            title="Next day"
            :disabled="!canNextDay"
            @click="moveSelectedDay(1)"
            >→</UiButton
          >
        </div>
      </header>
      <template v-if="selectedDetail">
        <p class="weather-muted">{{ forecast.location }} · {{ forecast.timezone }}</p>
        <div v-if="selectedDetail.primary" class="weather-current">
          <strong>{{ selectedDetail.primary }}</strong>
          <span aria-hidden="true">{{ weatherIcon(selectedDetail.condition) }}</span>
        </div>
        <h2>{{ selectedDetail.conditionLabel }}</h2>
        <WeatherStatGrid :stats="selectedDetail.stats" />
        <p class="weather-muted">Last refreshed {{ formatLocalDateTime(forecast.captured_at) }}</p>
        <p v-if="forecast.freshness === 'stale'" role="status">This forecast is out of date.</p>
      </template>
      <div
        class="weather-hours"
        role="region"
        :aria-label="selectedDetail ? `Hourly forecast for ${selectedDetail.title}` : 'Hourly forecast'"
        tabindex="0">
        <button
          v-for="item in activeHourCards"
          :key="item.id"
          type="button"
          class="weather-hour"
          :class="{ 'is-selected': selectedHour === item.id }"
          :aria-pressed="selectedHour === item.id"
          @click="selectHour(item.id)">
          <span>{{ item.time }}</span
          ><span aria-hidden="true">{{ weatherIcon(item.condition) }}</span
          ><strong v-if="item.temperature">{{ item.temperature }}</strong
          ><small v-if="item.precipitation">{{ item.precipitation }} rain</small
          ><small v-if="item.uv">{{ item.uv }}</small>
        </button>
      </div>
      <div v-if="selectedHourDetail" ref="hourDetailPanel" class="weather-hour-detail">
        <header>
          <h3>{{ selectedHourDetail.title }} · {{ selectedHourDetail.conditionLabel }}</h3>
          <div>
            <span class="weather-hour-detail-icon" aria-hidden="true">{{
              weatherIcon(selectedHourDetail.condition)
            }}</span>
            <strong v-if="selectedHourDetail.temperature">{{ selectedHourDetail.temperature }}</strong>
            <UiButton variant="ghost" @click="selectedHour = null">Hide</UiButton>
          </div>
        </header>
        <WeatherStatGrid :stats="selectedHourDetail.stats" />
      </div>
      <p v-if="selectedDetail && !selectedDetail.hours.length" class="weather-muted" role="status">
        No hourly detail is available for this day.
        <template v-if="selectedDetail.hourlyHorizonHours">
          Hourly data covers the next {{ selectedDetail.hourlyHorizonHours }} hours.
        </template>
      </p>
      <p v-if="!selectedDetail && !hourly.length" class="weather-muted">
        No upcoming hourly forecast is available. Refresh for newer information.
      </p>
    </section>
    <section class="weather-panel">
      <header>
        <h2>{{ daily.length }}-day forecast</h2>
      </header>
      <div class="weather-days" role="region" aria-label="Daily forecast">
        <WeatherDayCard
          v-for="day in visibleDays"
          :key="day.date"
          :day="day"
          :selected="selectedDate === day.date"
          @select="selectDay(day.date)" />
      </div>
      <UiPager
        v-if="pages > 1"
        :page="page + 1"
        :page-count="pages"
        :status="`Days ${firstDay}–${lastDay} of ${daily.length}`"
        label="Daily forecast pages"
        announce-status
        @update:page="page = $event - 1" />
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
  .weather-hour {
    appearance: none;
    border: var(--nh-border-w) solid transparent;
    border-radius: var(--nh-radius-m);
    background: var(--nh-surface-2);
    color: inherit;
    font: inherit;
    padding: 1rem;
    cursor: pointer;
  }
  .weather-hour:hover {
    border-color: var(--nh-border-strong);
  }
  .weather-hour:focus-visible {
    outline: 3px solid var(--nh-accent);
    outline-offset: 2px;
  }
  .weather-hour.is-selected {
    border-color: var(--nh-accent);
    box-shadow: 0 0 0 1px var(--nh-accent);
  }
  .weather-muted,
  small {
    color: var(--nh-text-muted);
    font-size: 0.8rem;
  }
  .weather-hours {
    display: flex;
    gap: 0.75rem;
    overflow-x: auto;
    scroll-snap-type: x mandatory;
    padding-bottom: 0.5rem;
  }
  .weather-days {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(min(100%, 10rem), 1fr));
    gap: 0.75rem;
  }
  .weather-hour {
    flex: 1 0 7rem;
    display: grid;
    gap: 0.75rem;
    text-align: center;
    scroll-snap-align: start;
  }
  .weather-hour-detail {
    margin-top: 0.75rem;
    padding-top: 0.75rem;
    border-top: var(--nh-border-w) solid var(--nh-border);
  }
  .weather-hour-detail header {
    margin-bottom: 0;
  }
  .weather-hour-detail h3 {
    margin: 0;
    font-size: 1rem;
  }
  .weather-hour-detail-icon {
    font-size: 1.35rem;
    line-height: 1;
  }
  .weather-hour-detail :deep(.weather-stats) {
    margin: 0.75rem 0 0;
  }
  @media (max-width: 700px) {
    .weather-panel {
      padding: 1rem;
    }
    .weather-hours {
      gap: 0.5rem;
    }
    .weather-hour {
      flex: 0 0 calc((100% - 1rem) / 3);
      min-width: 0;
      padding: 0.65rem;
    }
  }
</style>
