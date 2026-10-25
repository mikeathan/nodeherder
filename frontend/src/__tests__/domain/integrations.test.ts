import { fixedValue, valueSource } from '@/domain/integrations';
test('preserves explicit scalar types rather than interpreting input text as expressions', () => {
  expect(fixedValue('123', 'string')).toBe('123');
  expect(fixedValue('123', 'number')).toBe(123);
  expect(fixedValue('false', 'boolean')).toBe(false);
  expect(valueSource({ ref: 'input.postcode' })).toBe('input.postcode');
  expect(valueSource('input.postcode')).toBe('');
});
