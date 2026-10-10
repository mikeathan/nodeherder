/* Reactive CSS media query (layout decisions that CSS alone cannot make, e.g. drag handles). */
import { onScopeDispose, ref } from 'vue';

export const PHONE_QUERY = '(max-width: 760px)';
export const NAV_DRAWER_QUERY = '(max-width: 960px)';

export function useMediaQuery(query: string) {
  const media = typeof window !== 'undefined' && window.matchMedia ? window.matchMedia(query) : null;
  const matches = ref(media?.matches ?? false);
  const onChange = (e: MediaQueryListEvent) => (matches.value = e.matches);
  media?.addEventListener?.('change', onChange);
  onScopeDispose(() => media?.removeEventListener?.('change', onChange));
  return matches;
}
