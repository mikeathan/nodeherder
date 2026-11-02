import type { FuelPrice, FuelStation } from '@/types/integrations/fuel';

/** Return the cheapest supplied grade without inventing a cross-grade price. */
export function cheapestFuelPrice(station: FuelStation): FuelPrice | null {
  return station.prices.reduce<FuelPrice | null>(
    (cheapest, price) => (cheapest == null || price.pence_per_litre < cheapest.pence_per_litre ? price : cheapest),
    null
  );
}

/** Present integer pence per litre in the familiar UK forecourt format. */
export function fuelPriceValue(price: FuelPrice): string {
  return `${price.pence_per_litre}p/l`;
}
