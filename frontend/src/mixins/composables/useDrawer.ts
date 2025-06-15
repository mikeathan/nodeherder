// composables/useDrawer.ts
import { computed, watch, onMounted } from 'vue';
import { useWindowSize } from './useWindowsSize';

export function useDrawer(isExpanded: () => boolean, emitWidth: (width: number) => void) {
  const { isMobile } = useWindowSize(() => emitDrawerWidth());

  const drawerWidth = computed(() => {
    if (isMobile.value) return 0;
    return isExpanded() ? 250 : 60;
  });

  const emitDrawerWidth = () => emitWidth(drawerWidth.value);

  onMounted(() => emitDrawerWidth());
  watch([isExpanded, isMobile], emitDrawerWidth);

  return {
    isMobile,
    drawerWidth,
  };
}
