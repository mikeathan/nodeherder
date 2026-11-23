import { onBeforeUnmount, ref, watch, type Ref } from 'vue';

/**
 * Measures an element and keeps its size current.
 *
 * The recorded canvas derives its view box from the pixel aspect, so the size has to be observed
 * rather than read once: the recordings rail collapsing, the view switching and the window
 * resizing all change it. Observation is re-attached when `element` is replaced, because a
 * `v-if` swap creates a new node, and it is disconnected on unmount.
 *
 * @param element The element to measure, or null while it is not rendered.
 * @returns Reactive width/height in pixels, both 0 until the first measurement.
 */
export function useElementSize(element: Ref<HTMLElement | null>) {
  const width = ref(0);
  const height = ref(0);
  let observer: ResizeObserver | null = null;

  function measure(): void {
    const node = element.value;
    if (!node) return;
    const rect = node.getBoundingClientRect();
    width.value = rect.width;
    height.value = rect.height;
  }

  watch(
    element,
    (node) => {
      observer?.disconnect();
      observer = null;
      if (!node) return;
      observer = new ResizeObserver((entries) => {
        const box = entries[0]?.contentRect;
        if (!box) return;
        width.value = box.width;
        height.value = box.height;
      });
      observer.observe(node);
      measure();
    },
    { flush: 'post' }
  );

  onBeforeUnmount(() => {
    observer?.disconnect();
    observer = null;
  });

  return { width, height, measure };
}
