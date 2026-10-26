import type { WasteSchedule } from '@/types/integrations.type';

/** Date-only collection labels; null means that no upcoming date was supplied. */
export interface CollectionDatePresentation {
  /** Canonical YYYY-MM-DD key for ordering, independent of translated labels. */
  readonly nextDateISO: string | null;
  readonly nextDate: string | null;
  readonly nextDateShort: string | null;
  /** At most four later dates, formatted for the expanded results view. */
  readonly laterDates: string[];
}

const MAX_LATER_DATES = 4;
const longDateFormat = new Intl.DateTimeFormat('en-GB', {
  weekday: 'long',
  day: 'numeric',
  month: 'long',
  timeZone: 'UTC',
});
const shortDateFormat = new Intl.DateTimeFormat('en-GB', {
  weekday: 'short',
  day: 'numeric',
  month: 'short',
  timeZone: 'UTC',
});

// Ingress limits this cache to the zones supported by the versioned waste contract.
const calendarFormats = new Map<WasteSchedule['timezone'], Intl.DateTimeFormat>();

/**
 * Resolve a UTC evaluation instant to its calendar date in the resource's declared zone.
 * @param evaluatedAt Validated operational timestamp, not a collection date.
 * @param timezone Source calendar from the resource, independent of browser/process settings.
 * @returns Canonical YYYY-MM-DD for filtering and comparing collection dates.
 */
export function collectionCalendarDate(evaluatedAt: string, timezone: WasteSchedule['timezone']): string {
  let formatter = calendarFormats.get(timezone);
  if (!formatter) {
    formatter = new Intl.DateTimeFormat('en-CA', {
      timeZone: timezone,
      calendar: 'gregory',
      numberingSystem: 'latn',
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
    });
    calendarFormats.set(timezone, formatter);
  }
  // Build ISO explicitly: locale punctuation/order is not a reliable comparison key.
  const parts = Object.fromEntries(
    formatter.formatToParts(new Date(evaluatedAt)).map((part) => [part.type, part.value])
  );
  return `${parts.year}-${parts.month}-${parts.day}`;
}

/**
 * Compare canonical collection dates with undated services last.
 * @returns A sort comparison; equal/missing dates return zero to retain relative source order.
 */
export function compareCollectionDates(left: string | null, right: string | null): number {
  if (left === right) return 0;
  if (left === null) return 1;
  if (right === null) return -1;
  return left < right ? -1 : 1;
}

/** Format a date without converting its calendar day into the viewer's timezone. */
function collectionLabel(date: string, formatter: Intl.DateTimeFormat): string {
  // UTC is only a formatting carrier for an already-resolved date, not its source timezone.
  return formatter.format(new Date(`${date}T12:00:00Z`));
}

/**
 * Present the next and later collection dates without changing the saved/provider array.
 * @param dates Validated YYYY-MM-DD values; defensive sorting also supports cached/imported data.
 * @param today Calendar date resolved in the resource timezone by collectionCalendarDate.
 * @returns Nullable next-date labels and a bounded list; never infers a recurring schedule.
 */
export function presentCollectionDates(dates: readonly string[], today: string): CollectionDatePresentation {
  const upcoming = dates.filter((date) => date >= today).sort();
  const nextDateISO = upcoming[0] ?? null;
  if (nextDateISO === null) {
    return { nextDateISO: null, nextDate: null, nextDateShort: null, laterDates: [] };
  }
  return {
    nextDateISO,
    nextDate: collectionLabel(nextDateISO, longDateFormat),
    nextDateShort: collectionLabel(nextDateISO, shortDateFormat),
    laterDates: upcoming.slice(1, MAX_LATER_DATES + 1).map((date) => collectionLabel(date, longDateFormat)),
  };
}
