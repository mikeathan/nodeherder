import { serviceHomeURL } from '@/domain/integrations/service-links';

test('map provider API hosts to human-facing pages', () => {
  expect(serviceHomeURL('https://geocoding-api.open-meteo.com')).toBe(
    'https://open-meteo.com/en/docs/geocoding-api'
  );
  expect(serviceHomeURL('https://api.postcodes.io')).toBe('https://postcodes.io/');
  expect(serviceHomeURL('https://api.open-meteo.com')).toBe('https://open-meteo.com/');
  expect(serviceHomeURL('https://unknown.example/v1')).toBe('https://unknown.example/');
  expect(serviceHomeURL('')).toBe('');
});
