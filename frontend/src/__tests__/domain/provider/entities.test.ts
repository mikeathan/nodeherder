import { describe, expect, test } from '@jest/globals';
import {
  PROVIDER_ENTITY_PAGE_SIZE,
  providerPageCount,
  providerPageSlice,
  searchEntityOrder,
  stableEntityOrder,
} from '@/domain/provider/entities';
import { expose } from '../../helpers/devices';

describe('provider entity inventory', () => {
  test('orders by display label and reuses an unchanged ordering snapshot', () => {
    const entities = [
      expose('z', { attributes: { label: 'Alpha' } as never }),
      expose('a', { attributes: { label: 'Beta' } as never }),
    ];
    const first = stableEntityOrder(undefined, entities);
    expect(first.map((item) => item.name)).toEqual(['z', 'a']);
    expect(stableEntityOrder(first, entities)).toBe(first);
  });

  test('reorders when a discovered label changes', () => {
    const entities = [expose('z', { attributes: { label: 'Alpha' } as never }), expose('a', { attributes: { label: 'Beta' } as never })];
    const first = stableEntityOrder(undefined, entities);
    entities[0].attributes = { label: 'Zulu' } as never;
    expect(stableEntityOrder(first, entities).map((item) => item.name)).toEqual(['a', 'z']);
  });

  test('search resolves current values through the stable order', () => {
    const entities = [expose('one', { attributes: { label: 'One' } as never }), expose('two', { attributes: { label: 'Two' } as never })];
    const order = stableEntityOrder(undefined, entities);
    entities[1].data = 7;
    expect(searchEntityOrder(order, entities, 'two')).toEqual([entities[1]]);
    expect(searchEntityOrder(order, entities, '')[1].data).toBe(7);
  });

  test('pagination uses the documented page size', () => {
    expect(PROVIDER_ENTITY_PAGE_SIZE).toBe(8);
    expect(providerPageCount(0)).toBe(1);
    expect(providerPageCount(9)).toBe(2);
    expect(providerPageSlice([1, 2, 3, 4, 5, 6, 7, 8, 9], 2)).toEqual([9]);
  });
});
