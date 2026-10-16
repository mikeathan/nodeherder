import { readonly, ref } from 'vue';
import { isDefaultView } from '@/domain/default-view';
import { safeStorage } from '@/utils/storage';
import { RouteName } from '@/types/router';

export const DEFAULT_VIEW_KEY = 'nodeherder_default_view';
const stored = safeStorage.getItem(DEFAULT_VIEW_KEY);
const preferred = ref<RouteName | null>(isDefaultView(stored) ? stored : null);

/** One browser-local presentation preference, also usable when storage is blocked. */
export function useDefaultView() {
  function toggle(name: RouteName) {
    if (!isDefaultView(name)) return;
    preferred.value = preferred.value === name ? null : name;
    if (preferred.value === null) safeStorage.removeItem(DEFAULT_VIEW_KEY);
    else safeStorage.setItem(DEFAULT_VIEW_KEY, preferred.value);
  }
  return { preferred: readonly(preferred), toggle };
}
