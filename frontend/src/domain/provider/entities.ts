/*
 * Pure ordering, search and pagination for provider entity sections.
 * Components render the result; live readings cannot change inventory order.
 */
import type { Expose } from '@/types/device';
import { exposeDisplayLabel } from '../exposes';

/** Rows shown per provider entity page. */
export const PROVIDER_ENTITY_PAGE_SIZE = 8;

/** One stable ordering entry, independent of live readings. */
export type EntityOrderItem = { name: string; label: string };

/**
 * Reuses the previous ordering snapshot when names and labels are unchanged,
 * so a live reading update does not trigger a new sort.
 */
export function stableEntityOrder(previous: readonly EntityOrderItem[] | undefined, entities: readonly Expose[]): EntityOrderItem[] {
  const next = entities.map((entity) => ({ name: entity.name, label: exposeDisplayLabel(entity) }));
  const unchanged =
    previous?.length === next.length &&
    next.every((entity, index) => entity.name === previous[index].name && entity.label === previous[index].label);
  return unchanged ? (previous as EntityOrderItem[]) : next.sort((a, b) => a.label.localeCompare(b.label) || a.name.localeCompare(b.name));
}

/** Resolves the stable order against current entity values, then applies the search text. */
export function searchEntityOrder(
  order: readonly EntityOrderItem[],
  entities: readonly Expose[],
  query: string
): Expose[] {
  const byName = new Map(entities.map((entity) => [entity.name, entity]));
  const search = query.trim().toLocaleLowerCase();
  return order
    .map((item) => byName.get(item.name))
    .filter((entity): entity is Expose => !!entity)
    .filter((entity) => !search || exposeDisplayLabel(entity).toLocaleLowerCase().includes(search));
}

/** Number of pages required for the filtered rows. */
export function providerPageCount(itemCount: number, pageSize = PROVIDER_ENTITY_PAGE_SIZE): number {
  return Math.max(1, Math.ceil(itemCount / pageSize));
}

/** Slices one page; an out-of-range page yields an empty list. */
export function providerPageSlice<T>(items: readonly T[], page: number, pageSize = PROVIDER_ENTITY_PAGE_SIZE): T[] {
  return items.slice((page - 1) * pageSize, page * pageSize);
}
