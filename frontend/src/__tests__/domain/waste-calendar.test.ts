import { collectionCalendarDate, compareCollectionDates, presentCollectionDates } from '@/domain/waste-calendar';
import { WASTE_SCHEDULE_TIMEZONE } from '@/config/integrations';

test.each([
  ['2026-10-24T22:30:00Z', '2026-10-24'],
  ['2026-10-24T23:30:00Z', '2026-10-25'],
  ['2026-10-26T00:30:00Z', '2026-10-26'],
])('source-calendar day for %s is %s across the BST/GMT boundary', (instant, expected) => {
  expect(collectionCalendarDate(instant, WASTE_SCHEDULE_TIMEZONE)).toBe(expected);
});

test('date-only labels retain their calendar day and later dates are bounded without mutation', () => {
  const dates = [
    '2026-12-08',
    '2026-10-27',
    '2026-11-24',
    '2026-11-03',
    '2026-12-01',
    '2026-11-17',
    '2026-11-10',
    '2026-10-20',
  ];
  const original = [...dates];
  const labels = presentCollectionDates(dates, '2026-10-26');
  expect(labels).toMatchObject({
    nextDateISO: '2026-10-27',
    nextDate: 'Tuesday 27 October',
    nextDateShort: 'Tue 27 Oct',
  });
  expect(labels.laterDates).toEqual([
    'Tuesday 3 November',
    'Tuesday 10 November',
    'Tuesday 17 November',
    'Tuesday 24 November',
  ]);
  expect(dates).toEqual(original);
});

test('missing or exhausted dates keep nullable meanings and stable comparison ties', () => {
  expect(presentCollectionDates(['2026-10-25'], '2026-10-26')).toEqual({
    nextDateISO: null,
    nextDate: null,
    nextDateShort: null,
    laterDates: [],
  });
  const dates = [null, '2027-01-01', '2026-11-03', '2026-10-27', null, '2026-11-03'];
  expect([...dates].sort(compareCollectionDates)).toEqual([
    '2026-10-27',
    '2026-11-03',
    '2026-11-03',
    '2027-01-01',
    null,
    null,
  ]);
  expect(compareCollectionDates(null, null)).toBe(0);
  expect(compareCollectionDates('2026-11-03', '2026-11-03')).toBe(0);
});
