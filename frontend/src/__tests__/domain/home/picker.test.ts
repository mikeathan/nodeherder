import {
  filterHomeChoices,
  HOME_FILTER_ALL,
  HOME_FILTER_MAIN,
  homePickerChoices,
  homePickerFilters,
  homePickerOptions,
  selectedHomeCards,
  stableHomePickerFilters,
  stableHomePickerOptions,
} from '@/domain/home/picker';
import { homeSourceKey } from '@/domain/home/layout';
import { device, expose } from '../../helpers/devices';
import { ExposeCategories } from '@/types/device.type';
import type { HomeCard } from '@/types/home';
import type { HomeSourcePresentation } from '@/domain/home/cards';

const lamp = device('lamp', { friendly_name: 'Kitchen light' }, [
  expose('state', { data: true }),
  expose('temperature', { data: 21.5 }),
  expose('calibration', { category: ExposeCategories.Config, data: 1 }),
  expose('no_reading_yet'),
  expose('identify', { access_mode: 'write' as never, data: 'x' }),
]);
const existing: HomeCard = { id: 'accepted', kind: 'device', device_id: 'lamp', expose: 'state', panel: false };

test('picker excludes configuration, write-only and unreported values, and disables only sources already shown in the destination', () => {
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
    pickerLabel: 'Bins · complete schedule',
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

test('opaque provider entity IDs show the discovered label, not the raw ID', () => {
  const presence = device('epl', { friendly_name: 'Everything Presence Lite' }, [
    expose('00000000-0e4537ac', { data: 120, attributes: { label: 'Illuminance' } as never }),
    expose('00000000-15d100e3', { data: true, description: 'Occupancy' }),
  ]);
  const labels = homePickerChoices('device:epl', [presence], new Map(), []).map((item) => item.label);
  expect(labels).toEqual(['Illuminance', 'Occupancy']);
});

test('device values are grouped by meaning for ESPHome classes and Zigbee names alike', () => {
  const mixed = device('mixed', { friendly_name: 'Mixed' }, [
    expose('00000000-0000000a', { data: 90, attributes: { device_class: 'illuminance', label: 'Illuminance' } as never }),
    expose('presence', { data: true, type: 'binary' as never }),
    expose('00000000-0000000e', { data: 3, attributes: { label: 'Zone 3 count' } as never }),
    expose('uptime', { data: 5, category: 'diagnostic' as never }),
  ]);
  const all = homePickerChoices('device:mixed', [mixed], new Map(), []);
  expect(all.map((item) => item.group)).toEqual(['light', 'motion', 'other', 'diagnostic']);
  expect(filterHomeChoices(all, HOME_FILTER_MAIN).map((item) => item.label)).toEqual(['Illuminance', 'Presence']);
  expect(filterHomeChoices(all, 'motion').map((item) => item.label)).toEqual(['Presence']);
  expect(filterHomeChoices(all, HOME_FILTER_ALL)).toHaveLength(4);
  expect(homePickerFilters(all).map((item) => item.label)).toEqual([
    'Sensors and controls',
    'Presence & motion',
    'Light level',
    'Diagnostics',
    'Other',
    'Everything',
  ]);
});

test('source selectors keep their identity across reports that only change readings', () => {
  const garden = device('garden', { friendly_name: 'Garden sensor' }, [expose('temperature', { data: 8 })]);
  const first = stableHomePickerOptions(undefined, [lamp, garden]);
  const reported = [
    { ...lamp, exposes: { ...lamp.exposes, temperature: { ...lamp.exposes.temperature, data: 22.5 } } },
    { ...garden, exposes: { ...garden.exposes, temperature: { ...garden.exposes.temperature, data: 9 } } },
  ];
  expect(stableHomePickerOptions(first, reported)).toBe(first);
  expect(stableHomePickerOptions(first, [garden, lamp])).not.toBe(first);
  expect(stableHomePickerOptions(first, [{ ...lamp, friendly_name: 'Hall light' }, garden])).not.toBe(first);
  expect(
    stableHomePickerOptions(first, [lamp, garden, device('hall', { friendly_name: 'Hall sensor' }, [])])
  ).not.toBe(first);
});

test('value filters keep their identity while the available groups stay the same', () => {
  const choices = homePickerChoices('device:lamp', [lamp], new Map(), []);
  const first = stableHomePickerFilters(undefined, choices);
  const warmer = [
    { ...lamp, exposes: { ...lamp.exposes, temperature: { ...lamp.exposes.temperature, data: 24 } } },
  ];
  expect(stableHomePickerFilters(first, homePickerChoices('device:lamp', warmer, new Map(), []))).toBe(first);
  expect(stableHomePickerFilters(first, [])).not.toBe(first);
});

test('a spatial-capable device offers one whole-device presence choice', () => {
  const radar = device('radar', { provider: 'esphome', connection_type: 'esphome' }, [
    expose('target_x', { unit: 'mm', data: 500 }),
    expose('target_y', { unit: 'mm', data: 1500 }),
    expose('target_active', { type: 'binary', data: true }),
  ]);
  const [presence] = homePickerChoices('device:radar', [radar], new Map(), []);
  expect(presence.reference).toEqual({ kind: 'presence', device_id: 'radar' });
  expect(presence.label).toBe('Presence');
  expect(homePickerChoices('device:lamp', [lamp], new Map(), []).some((item) => item.reference.kind === 'presence')).toBe(false);
});

test('a presence overview is offered globally when a spatial sensor exists', () => {
  const radar = device('radar', { provider: 'esphome', connection_type: 'esphome' }, [
    expose('target_x', { unit: 'mm', data: 500 }),
    expose('target_y', { unit: 'mm', data: 1500 }),
    expose('target_active', { type: 'binary', data: true }),
  ]);
  expect(homePickerOptions([radar])).toContainEqual({ value: 'presence', label: 'Presence overview' });
  const [overview] = homePickerChoices('presence', [radar], new Map(), []);
  expect(overview.reference).toEqual({ kind: 'presence_overview' });
  expect(homePickerChoices('presence', [lamp], new Map(), [])).toEqual([]);
});
