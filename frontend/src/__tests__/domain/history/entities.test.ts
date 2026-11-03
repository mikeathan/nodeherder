import { historyEntityOptions } from '@/domain/history/entities';
import { device, expose } from '../../helpers/devices';

test('history options use human labels while preserving stable expose keys', () => {
  const value = device('radar', {}, [
    expose('00000000-0e4537ac', { description: 'Target 1 X' }),
    expose('00000000-15d100e3', { type: 'binary' as never, attributes: { label: 'Occupancy' } as never }),
    expose('state', { type: 'enum' as never, attributes: { label: 'Status' } as never }),
  ]);
  expect(historyEntityOptions(value)).toEqual([
    { value: 'state', label: 'Status', group: 'Switches & lights' },
    { value: '00000000-15d100e3', label: 'Occupancy', group: 'Other' },
    { value: '00000000-0e4537ac', label: 'Target 1 X', group: 'Other' },
  ]);
});

test('history options exclude non-history expose types', () => {
  const value = device('sensor', {}, [
    expose('temperature'),
    expose('command', { type: 'button' as never }),
  ]);
  expect(historyEntityOptions(value)).toEqual([{ value: 'temperature', label: 'Temperature', group: 'Temperature' }]);
});

test('history options leave out config and write-only values and group the rest by meaning', () => {
  const value = device('epl', {}, [
    expose('00000000-0000000a', { attributes: { device_class: 'illuminance', label: 'Illuminance' } as never }),
    expose('00000000-0000000b', { type: 'binary' as never, attributes: { device_class: 'motion', label: 'Motion' } as never }),
    expose('00000000-0000000c', { category: 'config' as never, attributes: { label: 'Zone 1 start' } as never }),
    expose('00000000-0000000d', { access_mode: 'write' as never, attributes: { label: 'Restart' } as never }),
    expose('uptime', { category: 'diagnostic' as never }),
    expose('presence', { type: 'binary' as never }),
  ]);
  expect(historyEntityOptions(value).map((o) => [o.group, o.label])).toEqual([
    ['Presence & motion', 'Motion'],
    ['Presence & motion', 'Presence'],
    ['Light level', 'Illuminance'],
    ['Diagnostics', 'Uptime'],
  ]);
});
