import { cheapestFuelPrice, fuelPriceValue } from '@/domain/integrations/fuel';
import type { FuelStation } from '@/types/integrations/fuel';

const station: FuelStation = {
  station_id: 'station',
  name: 'Shell',
  brand: 'Shell',
  address: '',
  postcode: '',
  latitude: null,
  longitude: null,
  distance_metres: 1200,
  prices: [
    { grade: 'e10', label: 'E10', pence_per_litre: 143, updated_at: null },
    { grade: 'b7', label: 'B7', pence_per_litre: 150, updated_at: null },
  ],
};

test('select and format the cheapest supplied grade', () => {
  const cheapest = cheapestFuelPrice(station);
  expect(cheapest?.grade).toBe('e10');
  expect(cheapest ? fuelPriceValue(cheapest) : '').toBe('143p/l');
  expect(cheapestFuelPrice({ ...station, prices: [] })).toBeNull();
});
