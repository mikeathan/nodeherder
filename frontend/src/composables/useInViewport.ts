import { onBeforeUnmount, onMounted, ref, type Ref } from 'vue';

/**
 * Tracks whether an element is near the viewport. Defaults to visible when the
 * IntersectionObserver API is unavailable, so a feature never disappears silently.
 */
export function useInViewport(target: Ref<HTMLElement | null>, rootMargin = '120px') {
  const visible = ref(true);
  let observer: IntersectionObserver | null = null;
  onMounted(() => {
    if (typeof IntersectionObserver === 'undefined' || !target.value) return;
    observer = new IntersectionObserver(([entry]) => {
      visible.value = entry.isIntersecting;
    }, { rootMargin });
    observer.observe(target.value);
  });
  onBeforeUnmount(() => observer?.disconnect());
  return visible;
}
