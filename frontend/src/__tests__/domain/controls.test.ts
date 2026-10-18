import { nativeNumber, numericStep, sliderValue, isMired } from '@/domain/controls';
import { expose } from '../helpers/devices';

test('native bounds, decimal steps and invalid numbers', () => {
  const e = expose('sensitivity', { attributes: { min: 0, max: 10, step: 0.1 } });
  expect(nativeNumber(e, 2.56)).toBe(2.6);
  expect(nativeNumber(e, -1)).toBe(0);
  expect(nativeNumber(e, 15)).toBe(10);
  expect(nativeNumber(e, NaN)).toBeNull();
  expect(numericStep(expose('x', { attributes: { value_step: 0.5 } }))).toBe(0.5);
  expect(numericStep(expose('x', { attributes: { step: -1 } }))).toBe(1);
  expect(numericStep(expose('x', { attributes: null }))).toBe(1);
  expect(nativeNumber(expose('x', { attributes: null }), 2.2)).toBe(2);
});
test('mired temperature reverses visual direction and Kelvin stays native', () => {
  const e = expose('color_temp', { unit: 'mired', attributes: { min: 153, max: 500 } });
  expect(isMired(e)).toBe(true);
  expect(sliderValue(e, 153)).toBe(500);
  expect(sliderValue(e, sliderValue(e, 370))).toBe(370);
  expect(sliderValue({ ...e, unit: 'K' }, 370)).toBe(370);
  expect(sliderValue(expose('x'), 5)).toBe(5);
});

test('invalid bounds disable edits and chart conversions do not round the data', () => {
  const { invalidRange, displayNumber } = require('@/domain/controls');
  expect(invalidRange(expose('x', { attributes: { min: 10, max: 0 } }))).toBe(true);
  expect(invalidRange(expose('x', { attributes: null }))).toBe(false);
  expect(displayNumber(expose('brightness', { attributes: { max: 254 } }), 184)).toBeCloseTo(72.44094);
  expect(displayNumber(expose('color_temp', { unit: 'mired' }), 370)).toBeCloseTo(2702.7027);
  expect(displayNumber(expose('color_temp', { unit: 'K' }), 2700)).toBe(2700);
  expect(displayNumber(expose('temperature', { unit: '°C' }), 9.732)).toBe(9.732);
});

test('native bounds accept metadata aliases and reject contradictory or absent limits', () => {
  const { exposeRange } = require('@/domain/controls');
  expect(exposeRange(expose('x', { attributes: { value_min: -10, value_max: 10 } }))).toEqual({ min: -10, max: 10 });
  expect(exposeRange(expose('x', { attributes: { min: 10, max: 10 } }))).toBeNull();
  expect(exposeRange(expose('x', { attributes: null }))).toBeNull();
});
