import { readFileSync } from 'node:fs';
import { resolve } from 'node:path';
import { parseWasteSchedule } from '@/contracts/integrations';
import { presentWasteSchedule } from '@/domain/waste-schedule';

const examples = JSON.parse(
  readFileSync(
    resolve(__dirname, '../../../../specs/032-external-integrations/contracts/waste_schedule.v1.examples.json'),
    'utf8'
  )
);

test('upcoming cards use the UK date at evaluation, not the browser timezone or a last-check timestamp', () => {
  const schedule = parseWasteSchedule(examples.exact);
  schedule.evaluated_at = '2026-10-24T23:30:00Z'; // Still British Summer Time: local date is the 25th.
  schedule.services[0].dates = ['2026-10-24', '2026-10-25', '2026-10-26'];
  const card = presentWasteSchedule(schedule).cards[0];
  expect(card.nextDate).toBe('Sunday 25 October');
  expect(card.laterDates).toEqual(['Monday 26 October']);
  expect(schedule.services[0].dates).toHaveLength(3);
});

test('unknown or exhausted dates stay empty instead of repeating a prior calendar pattern', () => {
  const schedule = parseWasteSchedule(examples.exact);
  schedule.evaluated_at = '2026-10-26T00:00:00Z';
  schedule.services[0].dates = ['2026-10-25'];
  expect(presentWasteSchedule(schedule).cards[0].nextDate).toBeNull();
  expect(presentWasteSchedule(schedule).cards[0].laterDates).toEqual([]);
});

test('concise collection titles retain full provider descriptions and bin labels separately', () => {
  const schedule = parseWasteSchedule(examples.exact);
  schedule.services[0].name = 'Brown Bin - Glass bottles, jars, tins, cans, foil, plastic and cartons';
  schedule.services[0].bin_colour = 'brown';
  schedule.services[0].waste_type = 'glass';
  const card = presentWasteSchedule(schedule).cards[0];
  expect(card.title).toBe('Brown bin');
  expect(card.binLabel).toBe('Brown bin');
  expect(card.name).toContain('Glass bottles, jars');
  expect(schedule.services[0].name).toContain('Brown Bin -');
});

test('unknown bin colours and categories preserve literal provider names', () => {
  const schedule = parseWasteSchedule(examples.exact);
  schedule.services[0].name = '[Mixed (subscription)';
  schedule.services[0].bin_colour = 'unlisted';
  schedule.services[0].waste_type = 'unknown';
  const card = presentWasteSchedule(schedule).cards[0];
  expect(card.title).toBe('[Mixed (subscription)');
  expect(card.name).toBe('[Mixed (subscription)');
});

test('collection cards use the earliest real upcoming date across months and keep undated services last', () => {
  const schedule = parseWasteSchedule(structuredClone(examples.exact));
  schedule.evaluated_at = '2026-10-26T12:00:00Z';
  const original = schedule.services[0];
  schedule.services = [
    { ...original, service_id: 'blue', dates: ['2026-11-17', '2026-12-15'] },
    { ...original, service_id: 'missing', date_state: 'unavailable', dates: [] },
    { ...original, service_id: 'black', dates: ['2026-12-22', '2026-10-27', '2026-10-20'] },
    { ...original, service_id: 'brown', dates: ['2026-11-03'] },
    { ...original, service_id: 'same-day', dates: ['2026-11-03'] },
    { ...original, service_id: 'empty', date_state: 'confirmed_empty', dates: [] },
    { ...original, service_id: 'exhausted', dates: ['2026-10-25'] },
    { ...original, service_id: 'next-year', dates: ['2027-01-01'] },
  ];
  const before = structuredClone(schedule);
  const cards = presentWasteSchedule(schedule).cards;
  expect(cards.map((card) => card.service_id)).toEqual([
    'black',
    'brown',
    'same-day',
    'blue',
    'next-year',
    'missing',
    'empty',
    'exhausted',
  ]);
  expect(cards[0].nextDate).toBe('Tuesday 27 October');
  expect(cards[0].laterDates).toEqual(['Tuesday 22 December']);
  expect(schedule).toEqual(before);
});
