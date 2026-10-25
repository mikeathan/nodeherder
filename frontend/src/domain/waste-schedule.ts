import { wasteBinTitle } from './waste-bin';
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

const checkedLabel = new Intl.DateTimeFormat('en-GB', {
  day: 'numeric',
  month: 'short',
  hour: '2-digit',
  minute: '2-digit',
});
const timestampLabel = new Intl.DateTimeFormat('en-GB', { dateStyle: 'medium', timeStyle: 'short' });
/** Build compact display data once per local projection read; source descriptions stay intact. */
export function presentWasteSchedule(schedule: WasteSchedule) {
  const today = calendarDate.format(new Date(schedule.evaluated_at));
  // Collection values are calendar dates, not instants. Midday UTC avoids date rollover.
  const label = (date: string) => collectionLabel.format(new Date(`${date}T12:00:00Z`));
  return {
    ...schedule,
    statusLabel:
      schedule.freshness === 'stale'
        ? 'Needs refreshing'
        : schedule.freshness === 'fresh'
          ? 'Saved schedule'
          : 'Freshness unknown',
    checkedLabel: checkedLabel.format(new Date(schedule.last_checked_at)),
    checkedDetail: timestampLabel.format(new Date(schedule.last_checked_at)),
    sourceDetail: schedule.source_retrieved_at
      ? timestampLabel.format(new Date(schedule.source_retrieved_at))
      : 'Not supplied',
    cards: schedule.services.map((service) => {
      const upcoming = service.dates.filter((date) => date >= today);
      return {
        ...service,
        title: wasteBinTitle(service.bin_colour) ?? service.name,
        binLabel: wasteBinTitle(service.bin_colour) ?? service.bin_colour,
        coverageLabel:
          service.coverage.state === 'limited'
            ? 'Limited dates'
            : service.coverage.state === 'next_only'
              ? 'Next date only'
              : 'Coverage unknown',
        nextDate: upcoming[0] ? label(upcoming[0]) : null,
        laterDates: upcoming.slice(1, 5).map(label),
      };
    }),
  };
}
