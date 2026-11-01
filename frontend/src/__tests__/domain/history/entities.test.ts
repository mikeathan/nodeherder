import { historyEntityOptions } from '@/domain/history/entities';
import { device, expose } from '../../helpers/devices';

test('history options use human labels while preserving stable expose keys', () => {
  const value = device('radar', {}, [
    expose('00000000-0e4537ac', { description: 'Target 1 X' }),
    expose('00000000-15d100e3', { type: 'binary' as never, attributes: { label: 'Occupancy' } as never }),
    expose('state', { type: 'enum' as never, attributes: { label: 'Status' } as never }),
  ]);
  expect(historyEntityOptions(value)).toEqual([
    { value: '00000000-15d100e3', label: 'Occupancy' },
    { value: 'state', label: 'Status' },
    { value: '00000000-0e4537ac', label: 'Target 1 X' },
  ]);
});

test('history options exclude non-history expose types', () => {
  const value = device('sensor', {}, [
    expose('temperature'),
    expose('command', { type: 'button' as never }),
  ]);
  expect(historyEntityOptions(value)).toEqual([{ value: 'temperature', label: 'Temperature' }]);
});
