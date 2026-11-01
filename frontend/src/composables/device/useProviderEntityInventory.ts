import { computed, ref, watch } from 'vue';
import type { Expose } from '@/types/device';
import {
  PROVIDER_ENTITY_PAGE_SIZE,
  providerPageCount,
  providerPageSlice,
  searchEntityOrder,
  stableEntityOrder,
  type EntityOrderItem,
} from '@/domain/provider/entities';

/** Owns search text and pagination for one provider entity section. */
export function useProviderEntityInventory(entities: () => readonly Expose[], deviceId: () => string) {
  const query = ref('');
  const page = ref(1);
  const order = computed<EntityOrderItem[]>((previous) => stableEntityOrder(previous, entities()));
  const filtered = computed(() => searchEntityOrder(order.value, entities(), query.value));
  const pages = computed(() => providerPageCount(filtered.value.length));
  const visible = computed(() => providerPageSlice(filtered.value, page.value));

  watch([deviceId, query], () => { page.value = 1; });
  watch(pages, (count) => { page.value = Math.min(page.value, count); });

  return { query, page, pages, filtered, visible, pageSize: PROVIDER_ENTITY_PAGE_SIZE };
}
