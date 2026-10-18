import { onBeforeUnmount, ref } from 'vue';

/** Reactive media query for presentation-only layout choices; the listener owns its cleanup. */
export function useMediaQuery(query: string) {
  const media =
    typeof window !== 'undefined' && typeof window.matchMedia === 'function' ? window.matchMedia(query) : null;
  const matches = ref(media?.matches ?? false);
  const update = (event?: MediaQueryListEvent) => {
    matches.value = event ? event.matches : (media?.matches ?? false);
  };
  media?.addEventListener('change', update);
  onBeforeUnmount(() => media?.removeEventListener('change', update));
  return matches;
}
