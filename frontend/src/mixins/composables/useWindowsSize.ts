import { ref, onMounted, onUnmounted, computed } from 'vue';

export function useWindowSize(onResizeCallback?: () => void) {
  const windowWidth = ref(window.innerWidth);

  const onResize = () => {
    windowWidth.value = window.innerWidth;
    onResizeCallback?.();
  };

  onMounted(() => window.addEventListener('resize', onResize));
  onUnmounted(() => window.removeEventListener('resize', onResize));

  return {
    windowWidth,
    isMobile: computed(() => windowWidth.value < 768),
  };
}