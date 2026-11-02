import {
  airQualityCategoryLabel,
  airQualityCategoryTone,
  airQualityIndexFamilyLabel,
  currentAirQualityPoint,
  dominantPollen,
  pollenReadings,
  upcomingAirQualityHours,
} from '@/domain/integrations/air-quality';
import type { AirQualityForecast, AirQualityPollen } from '@/types/integrations/air-quality';

const pollen: AirQualityPollen = {
  alder: 1,
  birch: null,
  grass: 4,
  mugwort: 0,
  olive: null,
  ragweed: null,
};

test('present category, tone and index family labels', () => {
  expect(airQualityCategoryLabel('unhealthy_sensitive')).toBe('Unhealthy for sensitive groups');
  expect(airQualityCategoryLabel('nonsense')).toBe('Unknown');
  expect(airQualityCategoryTone('good')).toBe('good');
  expect(airQualityCategoryTone('very_unhealthy')).toBe('poor');
  expect(airQualityIndexFamilyLabel('us')).toBe('US AQI');
});

test('list supplied pollen species and choose the highest reading', () => {
  expect(pollenReadings(pollen).map((item) => item.key)).toEqual(['alder', 'grass', 'mugwort']);
  expect(dominantPollen(pollen)).toMatchObject({ key: 'grass', label: 'Grass', value: 4 });
  expect(dominantPollen({ ...pollen, alder: null, grass: null, mugwort: null })).toBeNull();
});

test('current and upcoming hours ignore elapsed provider points', () => {
  const forecast = {
    current: null,
    evaluated_at: '2026-11-02T17:00:00Z',
    hourly: [
      { time: '2026-11-02T15:00:00Z', pm2_5: 1 },
      { time: '2026-11-02T16:00:00Z', pm2_5: 2 },
      { time: '2026-11-02T17:00:00Z', pm2_5: 3 },
      { time: '2026-11-02T18:00:00Z', pm2_5: 4 },
      { time: '2026-11-02T19:00:00Z', pm2_5: 5 },
    ],
  } as AirQualityForecast;
  expect(currentAirQualityPoint(forecast)).toMatchObject({ pm2_5: 3 });
  expect(upcomingAirQualityHours(forecast).map((point) => point.pm2_5)).toEqual([4, 5]);
});
