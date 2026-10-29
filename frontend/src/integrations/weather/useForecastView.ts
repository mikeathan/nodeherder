import { computed, ref, watch } from 'vue';
import { calendarDateInZone, createZonedTimeFormatter, formatInstant, formatShortCalendarDate } from '@/utils/date.utils';
import type { WeatherForecast } from '@/types/weather.type';

const DAYS_PER_PAGE = 5;

/** Project an accepted forecast into upcoming periods and five-day pages.
 * The getter retains reactive ownership in the caller. Scope-owned watchers reset or clamp
 * pagination on resource changes; the projection performs no I/O and starts no timers.
 */
export function useForecastView(forecast: () => WeatherForecast) {
  const page = ref(0);
  const today = computed(() => calendarDateInZone(forecast().evaluated_at, forecast().timezone));
  const daily = computed(() => forecast().daily.filter((day) => day.date >= today.value));
  const hourly = computed(() => {
    const now = Date.parse(forecast().evaluated_at);
    return forecast()
      .hourly.filter((point) => Date.parse(point.time) > now)
      .slice(0, 24);
  });
  const pages = computed(() => Math.ceil(daily.value.length / DAYS_PER_PAGE));
  const firstDay = computed(() => page.value * DAYS_PER_PAGE + 1);
  const lastDay = computed(() => Math.min(firstDay.value + DAYS_PER_PAGE - 1, daily.value.length));
  watch(
    () => forecast().resource_id,
    () => {
      page.value = 0;
    }
  );
  watch(pages, (count) => {
    page.value = Math.min(page.value, Math.max(0, count - 1));
  });
  const hourFormatter = computed(
    () => createZonedTimeFormatter(forecast().timezone)
  );
  /** Format one forecast instant using the selected location's timezone. */
  function hour(value: string) {
    return formatInstant(value, hourFormatter.value);
  }
  /** Format a date-only summary without applying the browser's local timezone. */
  function date(value: string) {
    return formatShortCalendarDate(value);
  }
  /** Report whether the day index belongs to the currently displayed mobile page. */
  function outsidePage(index: number) {
    return index < firstDay.value - 1 || index >= lastDay.value;
  }
  return { page, daily, hourly, pages, firstDay, lastDay, hour, date, outsidePage };
}
