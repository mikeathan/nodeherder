import { homePickerChoices, homePickerOptions, selectedHomeCards } from '@/domain/home-picker';
import { homeSourceKey } from '@/domain/home-layout';
import { device, expose } from '../helpers/devices';
import { ExposeCategories } from '@/types/device.type';
import type { HomeCard } from '@/types/home.type';
import type { HomeSourcePresentation } from '@/domain/home-cards';

const lamp = device('lamp', { friendly_name: 'Kitchen light' }, [
  expose('state'),
  expose('temperature'),
  expose('calibration', { category: ExposeCategories.Config }),
]);
const existing: HomeCard = { id: 'accepted', kind: 'device', device_id: 'lamp', expose: 'state', panel: false };

test('picker excludes configuration and disables only sources already shown in the destination', () => {
  const choices = homePickerChoices('device:lamp', [lamp], new Map(), [existing]);
  expect(choices.map((item) => item.reference.kind === 'device' && item.reference.expose)).toEqual([
    'state',
    'temperature',
  ]);
  expect(choices.map((item) => item.alreadyShown)).toEqual([true, false]);
  expect(homePickerOptions([lamp])[1]).toEqual({ value: 'device:lamp', label: 'Kitchen light' });
  expect(homePickerChoices('device:missing', [lamp], new Map(), [])).toEqual([]);
  const cards = selectedHomeCards(
    choices,
    choices.map((item) => item.key)
  );
  expect(cards).toHaveLength(1);
  expect(cards[0]).toMatchObject({ kind: 'device', expose: 'temperature', panel: true });
  expect(cards[0].id).toBeTruthy();
  expect(existing.panel).toBe(false);
});

test('summary and individual service remain distinct choices with independent placement IDs', () => {
  const item: HomeSourcePresentation = {
    reference: { kind: 'waste_summary', instance_id: 'bins', resource_id: 'calendar' },
    title: 'Bins',
    value: 'Schedule',
    origin: 'Bins',
    status: 'Saved schedule',
    checked: '',
    colour: null,
    warning: false,
    rows: [],
  };
  const sources = new Map([[homeSourceKey(item.reference), item]]);
  const choices = homePickerChoices('integrations', [], sources, []);
  expect(choices[0].label).toBe('Bins · complete schedule');
  const first = selectedHomeCards(choices, [choices[0].key])[0];
  const second = selectedHomeCards(choices, [choices[0].key])[0];
  expect(first.id).not.toBe(second.id);
  expect(homeSourceKey(first)).toBe(homeSourceKey(second));
  expect(item.reference).not.toHaveProperty('id');
  expect(item.reference).not.toHaveProperty('panel');
});
