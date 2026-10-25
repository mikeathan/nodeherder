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
