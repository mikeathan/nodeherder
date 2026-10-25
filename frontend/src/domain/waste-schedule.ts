import type { WasteSchedule } from '@/types/integrations.type';

const calendarDate = new Intl.DateTimeFormat('en-CA', {
  timeZone: 'Europe/London',
  year: 'numeric',
  month: '2-digit',
  day: '2-digit',
});
const collectionLabel = new Intl.DateTimeFormat('en-GB', {
  weekday: 'long',
  day: 'numeric',
  month: 'long',
  timeZone: 'UTC',
});

/** Build cards once per local projection read, using the backend's evaluation time. */
export function presentWasteSchedule(schedule: WasteSchedule) {
  const today = calendarDate.format(new Date(schedule.evaluated_at));
  // Collection values are calendar dates, not instants. Midday UTC avoids date rollover.
  const label = (date: string) => collectionLabel.format(new Date(`${date}T12:00:00Z`));
  return {
    ...schedule,
    cards: schedule.services.map((service) => {
      const upcoming = service.dates.filter((date) => date >= today);
      return {
        ...service,
        nextDate: upcoming[0] ? label(upcoming[0]) : null,
        laterDates: upcoming.slice(1, 5).map(label),
      };
    }),
  };
}
