import { formatDistance, formatQuantity, formatWithUnit } from '@/utils/units';

test('temperature uses its canonical unit and a compact short form', () => {
  expect(formatQuantity(14, 'temperature')).toBe('14°C');
  expect(formatQuantity(14, 'temperature', { short: true })).toBe('14°');
  expect(formatQuantity(14.6, 'temperature')).toBe('15°C');
});

test('percent and speed keep their documented spacing', () => {
  expect(formatQuantity(20, 'percent')).toBe('20%');
  expect(formatQuantity(20, 'percent', { spaced: true })).toBe('20 %');
  expect(formatQuantity(18, 'speed')).toBe('18 km/h');
});

test('other quantities use one owner for unit and precision', () => {
  expect(formatQuantity(2.54, 'precipitation')).toBe('2.5 mm');
  expect(formatQuantity(1013, 'pressure')).toBe('1013 hPa');
  expect(formatQuantity(11, 'pollutant')).toBe('11 µg/m³');
  expect(formatQuantity(3.26, 'uv')).toBe('3.3');
});

test('distance converts metres to kilometres above 1000', () => {
  expect(formatDistance(850)).toBe('850 m');
  expect(formatDistance(1240)).toBe('1.2 km');
});

test('missing and non-finite values use an explicit placeholder, never zero', () => {
  expect(formatQuantity(null, 'temperature')).toBe('—');
  expect(formatQuantity(undefined, 'speed')).toBe('—');
  expect(formatQuantity(Number.NaN, 'percent')).toBe('—');
  expect(formatQuantity(null, 'temperature', { missing: 'No data' })).toBe('No data');
  expect(formatWithUnit(null, 'W')).toBe('—');
});

test('dynamic units control their own spacing', () => {
  expect(formatWithUnit(19.84, '°C', { spaced: true, digits: 1 })).toBe('19.8 °C');
  expect(formatWithUnit(20, '%', { spaced: true })).toBe('20 %');
  expect(formatWithUnit(20, '%')).toBe('20%');
  expect(formatWithUnit(2700, 'K', { spaced: true })).toBe('2700 K');
  expect(formatWithUnit(0, 'W', { spaced: true, digits: 2 })).toBe('0 W');
});
