import { wasteBinTitle } from './waste-bin';
import { collectionCalendarDate, compareCollectionDates, presentCollectionDates } from './waste-calendar';
import type { CollectionDatePresentation } from './waste-calendar';
import type { WasteSchedule, WasteService } from '@/types/integrations.type';

// Preserve the existing comparator export while keeping calendar behavior in one module.
export { compareCollectionDates } from './waste-calendar';

/** A display card retaining the provider's original fields alongside concise UI labels. */
export interface WasteServiceCard extends WasteService, CollectionDatePresentation {
  readonly title: string;
  readonly binLabel: string | null;
  readonly coverageLabel: string;
}

/** Validated resource plus presentation-only metadata; it is never written back to storage. */
export interface WasteSchedulePresentation extends WasteSchedule {
  readonly statusLabel: string;
  readonly checkedLabel: string;
  readonly checkedDetail: string;
  readonly sourceDetail: string;
  readonly cards: WasteServiceCard[];
}

const freshnessLabels: Readonly<Record<WasteSchedule['freshness'], string>> = {
  fresh: 'Saved schedule',
  stale: 'Needs refreshing',
  unknown: 'Freshness unknown',
};
const coverageLabels: Readonly<Record<string, string>> = {
  limited: 'Limited dates',
  next_only: 'Next date only',
  unknown: 'Coverage unknown',
};
const missingDateLabels: Readonly<Record<WasteService['date_state'], string>> = {
  known: 'No upcoming date',
  confirmed_empty: 'None scheduled',
  unavailable: 'Dates unavailable',
};

/** Show a date only when supplied; keep confirmed-empty, unknown and exhausted horizons distinct. */
export function wasteServiceDateLabel(card: WasteServiceCard): string {
  return card.nextDateShort ?? missingDateLabels[card.date_state];
}

// Operational timestamps retain the existing browser-local display behavior. Collection dates
// use the resource calendar instead; a future General timezone setting can target timestamps.
const checkedTimeFormat = new Intl.DateTimeFormat('en-GB', {
  day: 'numeric',
  month: 'short',
  hour: '2-digit',
  minute: '2-digit',
});
const timestampFormat = new Intl.DateTimeFormat('en-GB', { dateStyle: 'medium', timeStyle: 'short' });

/** Keep missing source verification distinct from a locally successful check. */
function sourceVerificationLabel(retrievedAt: string | null): string {
  if (retrievedAt === null) return 'Not supplied';
  return timestampFormat.format(new Date(retrievedAt));
}

/** Map one service without changing its identity, provider description or date evidence. */
function presentService(service: WasteService, today: string): WasteServiceCard {
  const colourName = wasteBinTitle(service.bin_colour);
  return {
    ...service,
    ...presentCollectionDates(service.dates, today),
    title: colourName ?? service.name,
    binLabel: colourName ?? service.bin_colour,
    coverageLabel: coverageLabels[service.coverage.state] ?? coverageLabels.unknown,
  };
}

/** Build new cards in chronological order; equal/undated cards keep their relative order. */
function presentServiceCards(services: readonly WasteService[], today: string): WasteServiceCard[] {
  return services
    .map((service) => presentService(service, today))
    .sort((left, right) => compareCollectionDates(left.nextDateISO, right.nextDateISO));
}

/**
 * Compose the Results/summary view model from a validated schedule.
 * @param schedule Resource containing UTC check times and source-local collection dates.
 * @returns New presentation data; saved arrays, identities and unknown values remain intact.
 */
export function presentWasteSchedule(schedule: WasteSchedule): WasteSchedulePresentation {
  const today = collectionCalendarDate(schedule.evaluated_at, schedule.timezone);
  const checkedAt = new Date(schedule.last_checked_at);
  return {
    ...schedule,
    statusLabel: freshnessLabels[schedule.freshness],
    checkedLabel: checkedTimeFormat.format(checkedAt),
    checkedDetail: timestampFormat.format(checkedAt),
    sourceDetail: sourceVerificationLabel(schedule.source_retrieved_at),
    cards: presentServiceCards(schedule.services, today),
  };
}
